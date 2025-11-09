package v1

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/mahendrakalkura/saas-blueprint/internal/auth"
	"github.com/mahendrakalkura/saas-blueprint/internal/config"
	"github.com/mahendrakalkura/saas-blueprint/internal/mfa"
	"github.com/mahendrakalkura/saas-blueprint/internal/models"
	"github.com/mahendrakalkura/saas-blueprint/internal/repository"
	"github.com/mahendrakalkura/saas-blueprint/internal/worker"
	"github.com/rs/zerolog/log"
)

type AuthHandler struct {
	userRepo     *repository.UserRepository
	sessionRepo  *repository.SessionRepository
	workerClient *worker.Client
	cfg          *config.Config
	mfaService   *mfa.Service
}

func NewAuthHandler(userRepo *repository.UserRepository, sessionRepo *repository.SessionRepository, workerClient *worker.Client, cfg *config.Config, mfaService *mfa.Service) *AuthHandler {
	return &AuthHandler{
		userRepo:     userRepo,
		sessionRepo:  sessionRepo,
		workerClient: workerClient,
		cfg:          cfg,
		mfaService:   mfaService,
	}
}

type RegisterRequest struct {
	Email     string  `json:"email" validate:"required,email"`
	Password  string  `json:"password" validate:"required,min=8"`
	FirstName *string `json:"first_name,omitempty"`
	LastName  *string `json:"last_name,omitempty"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type VerifyEmailRequest struct {
	Token string `json:"token" validate:"required"`
}

type RequestPasswordResetRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type ResetPasswordRequest struct {
	Token    string `json:"token" validate:"required"`
	Password string `json:"password" validate:"required,min=8"`
}

type VerifyMFARequest struct {
	MFAToken string `json:"mfa_token" validate:"required"`
	Code     string `json:"code" validate:"required"`
}

type AuthResponse struct {
	AccessToken  string       `json:"access_token,omitempty"`
	RefreshToken string       `json:"refresh_token,omitempty"`
	ExpiresIn    int64        `json:"expires_in,omitempty"`
	User         *models.User `json:"user,omitempty"`
	MFARequired  bool         `json:"mfa_required,omitempty"`
	MFAToken     string       `json:"mfa_token,omitempty"`
}

// Register creates a new user account
// @Summary Register a new user
// @Description Create a new user account with email and password
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Registration details"
// @Success 201 {object} AuthResponse
// @Failure 400 {object} map[string]string "Invalid request payload"
// @Failure 409 {object} map[string]string "Email already registered"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /auth/register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Normalize email
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	// Check if user already exists
	existingUser, _ := h.userRepo.GetByEmail(r.Context(), req.Email)
	if existingUser != nil {
		respondError(w, http.StatusConflict, "Email already registered")
		return
	}

	// Hash password
	passwordHash, err := auth.HashPassword(req.Password)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to process password")
		return
	}

	// Create user
	user := &models.User{
		Email:        req.Email,
		PasswordHash: passwordHash,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
	}

	if err := h.userRepo.Create(r.Context(), user); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	// Generate email verification token
	verificationToken, err := auth.GenerateVerificationToken()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to generate verification token")
		return
	}

	expiresAt := time.Now().Add(24 * time.Hour)
	if err := h.userRepo.SetEmailVerificationToken(r.Context(), user.ID, verificationToken, expiresAt); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to set verification token")
		return
	}

	// Enqueue welcome email
	firstName := ""
	if user.FirstName != nil {
		firstName = *user.FirstName
	}
	if err := h.workerClient.EnqueueWelcomeEmail(user.Email, firstName); err != nil {
		log.Error().Err(err).Str("email", user.Email).Msg("Failed to enqueue welcome email")
		// Don't fail registration if email fails
	}

	// Enqueue verification email
	baseURL := h.cfg.Server.FrontendURL
	if baseURL == "" {
		baseURL = "http://localhost:3000"
	}
	if err := h.workerClient.EnqueueVerificationEmail(user.Email, verificationToken, baseURL); err != nil {
		log.Error().Err(err).Str("email", user.Email).Msg("Failed to enqueue verification email")
		// Don't fail registration if email fails
	}

	// Generate tokens
	accessToken, err := auth.GenerateAccessToken(user.ID, user.Email, h.cfg.JWT.Secret)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to generate access token")
		return
	}

	refreshToken, err := auth.GenerateRefreshToken()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to generate refresh token")
		return
	}

	// Create session
	session := &models.Session{
		UserID:                user.ID,
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: time.Now().Add(h.cfg.JWT.RefreshTokenDuration),
		UserAgent:             stringPtr(r.UserAgent()),
		IPAddress:             stringPtr(r.RemoteAddr),
	}

	if err := h.sessionRepo.Create(r.Context(), session); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create session")
		return
	}

	respondJSON(w, http.StatusCreated, AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(auth.AccessTokenDuration.Seconds()),
		User:         user,
	})
}

// Login authenticates a user
// @Summary Login to account
// @Description Authenticate user with email and password, returns tokens or MFA challenge
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login credentials"
// @Success 200 {object} AuthResponse "Login successful or MFA required"
// @Failure 400 {object} map[string]string "Invalid request payload"
// @Failure 401 {object} map[string]string "Invalid email or password"
// @Failure 403 {object} map[string]string "Account is deactivated"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Normalize email
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	// Get user
	user, err := h.userRepo.GetByEmail(r.Context(), req.Email)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	// Check if user is active
	if !user.IsActive {
		respondError(w, http.StatusForbidden, "Account is deactivated")
		return
	}

	// Verify password
	if err := auth.CheckPassword(req.Password, user.PasswordHash); err != nil {
		respondError(w, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	// Check if MFA is enabled
	if user.MFAEnabled {
		// Generate temporary MFA token (5 minute expiry)
		mfaToken, err := auth.GenerateMFAToken(user.ID, h.cfg.JWT.Secret)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to generate MFA token")
			return
		}

		respondJSON(w, http.StatusOK, AuthResponse{
			MFARequired: true,
			MFAToken:    mfaToken,
		})
		return
	}

	// Generate tokens
	accessToken, err := auth.GenerateAccessToken(user.ID, user.Email, h.cfg.JWT.Secret)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to generate access token")
		return
	}

	refreshToken, err := auth.GenerateRefreshToken()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to generate refresh token")
		return
	}

	// Create session
	session := &models.Session{
		UserID:                user.ID,
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: time.Now().Add(h.cfg.JWT.RefreshTokenDuration),
		UserAgent:             stringPtr(r.UserAgent()),
		IPAddress:             stringPtr(r.RemoteAddr),
	}

	if err := h.sessionRepo.Create(r.Context(), session); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create session")
		return
	}

	respondJSON(w, http.StatusOK, AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(auth.AccessTokenDuration.Seconds()),
		User:         user,
	})
}

// Refresh generates new access token from refresh token
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Get session
	session, err := h.sessionRepo.GetByRefreshToken(r.Context(), req.RefreshToken)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "Invalid refresh token")
		return
	}

	// Check if token is expired
	if time.Now().After(session.RefreshTokenExpiresAt) {
		respondError(w, http.StatusUnauthorized, "Refresh token expired")
		return
	}

	// Get user
	user, err := h.userRepo.GetByID(r.Context(), session.UserID)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "User not found")
		return
	}

	// Check if user is active
	if !user.IsActive {
		respondError(w, http.StatusForbidden, "Account is deactivated")
		return
	}

	// Revoke old refresh token
	if err := h.sessionRepo.Revoke(r.Context(), req.RefreshToken); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to revoke old token")
		return
	}

	// Generate new tokens
	accessToken, err := auth.GenerateAccessToken(user.ID, user.Email, h.cfg.JWT.Secret)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to generate access token")
		return
	}

	newRefreshToken, err := auth.GenerateRefreshToken()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to generate refresh token")
		return
	}

	// Create new session
	newSession := &models.Session{
		UserID:                user.ID,
		RefreshToken:          newRefreshToken,
		RefreshTokenExpiresAt: time.Now().Add(h.cfg.JWT.RefreshTokenDuration),
		UserAgent:             stringPtr(r.UserAgent()),
		IPAddress:             stringPtr(r.RemoteAddr),
	}

	if err := h.sessionRepo.Create(r.Context(), newSession); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create session")
		return
	}

	respondJSON(w, http.StatusOK, AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    int64(auth.AccessTokenDuration.Seconds()),
		User:         user,
	})
}

// Logout revokes the current session
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if err := h.sessionRepo.Revoke(r.Context(), req.RefreshToken); err != nil {
		// Don't return error if session not found
		respondJSON(w, http.StatusOK, map[string]string{"message": "Logged out successfully"})
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Logged out successfully"})
}

// Me returns the current authenticated user
// @Summary Get current user
// @Description Get the currently authenticated user's information
// @Tags Auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} models.User
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "User not found"
// @Router /auth/me [get]
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := auth.GetUserFromContext(r)
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	user, err := h.userRepo.GetByID(r.Context(), userCtx.UserID)
	if err != nil {
		respondError(w, http.StatusNotFound, "User not found")
		return
	}

	respondJSON(w, http.StatusOK, user)
}

// VerifyEmail verifies user email
func (h *AuthHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	var req VerifyEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if err := h.userRepo.VerifyEmail(r.Context(), req.Token); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid or expired verification token")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Email verified successfully"})
}

// RequestPasswordReset sends a password reset email
func (h *AuthHandler) RequestPasswordReset(w http.ResponseWriter, r *http.Request) {
	var req RequestPasswordResetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Normalize email
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	// Generate reset token
	resetToken, err := auth.GenerateVerificationToken()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to generate reset token")
		return
	}

	expiresAt := time.Now().Add(1 * time.Hour)

	// Set reset token (don't reveal if user exists)
	err = h.userRepo.SetPasswordResetToken(r.Context(), req.Email, resetToken, expiresAt)

	// Enqueue password reset email only if user exists (but don't reveal this)
	if err == nil {
		baseURL := h.cfg.Server.FrontendURL
		if baseURL == "" {
			baseURL = "http://localhost:3000"
		}
		if err := h.workerClient.EnqueuePasswordResetEmail(req.Email, resetToken, baseURL); err != nil {
			log.Error().Err(err).Str("email", req.Email).Msg("Failed to enqueue password reset email")
			// Don't fail the request if email fails
		}
	}

	// Always return success to prevent user enumeration
	respondJSON(w, http.StatusOK, map[string]string{
		"message": "If the email exists, a password reset link has been sent",
	})
}

// ResetPassword resets user password
func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Hash new password
	passwordHash, err := auth.HashPassword(req.Password)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to process password")
		return
	}

	if err := h.userRepo.ResetPassword(r.Context(), req.Token, passwordHash); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid or expired reset token")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Password reset successfully"})
}

// VerifyMFA verifies MFA code and completes login
// @Summary Verify MFA code
// @Description Complete login by verifying MFA TOTP code or backup code
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body VerifyMFARequest true "MFA verification details"
// @Success 200 {object} AuthResponse
// @Failure 400 {object} map[string]string "Invalid request payload"
// @Failure 401 {object} map[string]string "Invalid or expired MFA token"
// @Failure 403 {object} map[string]string "Account is deactivated"
// @Router /auth/mfa/verify [post]
func (h *AuthHandler) VerifyMFA(w http.ResponseWriter, r *http.Request) {
	var req VerifyMFARequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Validate MFA token
	claims, err := auth.ValidateMFAToken(req.MFAToken, h.cfg.JWT.Secret)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "Invalid or expired MFA token")
		return
	}

	// Get user
	user, err := h.userRepo.GetByID(r.Context(), claims.UserID)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "User not found")
		return
	}

	// Check if user is active
	if !user.IsActive {
		respondError(w, http.StatusForbidden, "Account is deactivated")
		return
	}

	// Check if MFA is still enabled
	if !user.MFAEnabled || user.MFASecret == nil {
		respondError(w, http.StatusBadRequest, "MFA is not enabled")
		return
	}

	// Try to verify as TOTP code first
	isValid := h.mfaService.VerifyCode(*user.MFASecret, req.Code)

	// If TOTP failed, try backup codes
	var backupCodeUsed bool
	if !isValid && user.MFABackupCodes != nil {
		var remainingCodes []string
		isValid, remainingCodes = h.mfaService.VerifyBackupCode(req.Code, user.MFABackupCodes)
		if isValid {
			backupCodeUsed = true
			// Update user with remaining backup codes
			user.MFABackupCodes = remainingCodes
			if err := h.userRepo.Update(r.Context(), user); err != nil {
				log.Error().Err(err).Msg("Failed to update backup codes")
				// Continue anyway since auth was successful
			}
		}
	}

	if !isValid {
		respondError(w, http.StatusUnauthorized, "Invalid code")
		return
	}

	// Generate tokens
	accessToken, err := auth.GenerateAccessToken(user.ID, user.Email, h.cfg.JWT.Secret)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to generate access token")
		return
	}

	refreshToken, err := auth.GenerateRefreshToken()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to generate refresh token")
		return
	}

	// Create session
	session := &models.Session{
		UserID:                user.ID,
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: time.Now().Add(h.cfg.JWT.RefreshTokenDuration),
		UserAgent:             stringPtr(r.UserAgent()),
		IPAddress:             stringPtr(r.RemoteAddr),
	}

	if err := h.sessionRepo.Create(r.Context(), session); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create session")
		return
	}

	if backupCodeUsed {
		log.Info().Str("user_id", user.ID).Msg("User logged in with backup code")
	}

	respondJSON(w, http.StatusOK, AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(auth.AccessTokenDuration.Seconds()),
		User:         user,
	})
}

func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}

func stringPtr(s string) *string {
	return &s
}
