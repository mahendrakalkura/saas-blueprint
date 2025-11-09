package v1

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/mahendrakalkura/saas-blueprint/internal/auth"
	"github.com/mahendrakalkura/saas-blueprint/internal/config"
	"github.com/mahendrakalkura/saas-blueprint/internal/models"
	"github.com/mahendrakalkura/saas-blueprint/internal/oauth"
	"github.com/mahendrakalkura/saas-blueprint/internal/repository"
	"github.com/rs/zerolog/log"
)

type OAuthHandler struct {
	oauthService  *oauth.Service
	userRepo      *repository.UserRepository
	sessionRepo   *repository.SessionRepository
	oauthRepo     *repository.OAuthProviderRepository
	cfg           *config.Config
}

func NewOAuthHandler(
	oauthService *oauth.Service,
	userRepo *repository.UserRepository,
	sessionRepo *repository.SessionRepository,
	oauthRepo *repository.OAuthProviderRepository,
	cfg *config.Config,
) *OAuthHandler {
	return &OAuthHandler{
		oauthService:  oauthService,
		userRepo:      userRepo,
		sessionRepo:   sessionRepo,
		oauthRepo:     oauthRepo,
		cfg:           cfg,
	}
}

// GoogleLogin initiates Google OAuth flow
func (h *OAuthHandler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	h.initiateOAuthFlow(w, r, models.OAuthProviderGoogle)
}

// GitHubLogin initiates GitHub OAuth flow
func (h *OAuthHandler) GitHubLogin(w http.ResponseWriter, r *http.Request) {
	h.initiateOAuthFlow(w, r, models.OAuthProviderGitHub)
}

func (h *OAuthHandler) initiateOAuthFlow(w http.ResponseWriter, r *http.Request, provider string) {
	// Generate random state for CSRF protection
	state, err := generateRandomState()
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate OAuth state")
		http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
		return
	}

	// Store state in session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		MaxAge:   600, // 10 minutes
		HttpOnly: true,
		Secure:   h.cfg.Server.Environment == "production",
		SameSite: http.SameSiteLaxMode,
	})

	// Get authorization URL
	authURL, err := h.oauthService.GetAuthURL(provider, state)
	if err != nil {
		log.Error().Err(err).Str("provider", provider).Msg("Failed to get OAuth URL")
		http.Error(w, `{"error":"Failed to initialize OAuth"}`, http.StatusInternalServerError)
		return
	}

	// Redirect to OAuth provider
	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

// GoogleCallback handles Google OAuth callback
func (h *OAuthHandler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	h.handleOAuthCallback(w, r, models.OAuthProviderGoogle)
}

// GitHubCallback handles GitHub OAuth callback
func (h *OAuthHandler) GitHubCallback(w http.ResponseWriter, r *http.Request) {
	h.handleOAuthCallback(w, r, models.OAuthProviderGitHub)
}

func (h *OAuthHandler) handleOAuthCallback(w http.ResponseWriter, r *http.Request, provider string) {
	// Verify state to prevent CSRF
	stateCookie, err := r.Cookie("oauth_state")
	if err != nil {
		log.Error().Err(err).Msg("Missing OAuth state cookie")
		http.Redirect(w, r, h.cfg.Server.AllowedOrigins[0]+"/login?error=invalid_state", http.StatusTemporaryRedirect)
		return
	}

	state := r.URL.Query().Get("state")
	if state != stateCookie.Value {
		log.Error().Msg("OAuth state mismatch")
		http.Redirect(w, r, h.cfg.Server.AllowedOrigins[0]+"/login?error=invalid_state", http.StatusTemporaryRedirect)
		return
	}

	// Clear state cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})

	// Exchange authorization code for token
	code := r.URL.Query().Get("code")
	if code == "" {
		log.Error().Msg("Missing OAuth code")
		http.Redirect(w, r, h.cfg.Server.AllowedOrigins[0]+"/login?error=missing_code", http.StatusTemporaryRedirect)
		return
	}

	token, err := h.oauthService.ExchangeCode(r.Context(), provider, code)
	if err != nil {
		log.Error().Err(err).Msg("Failed to exchange OAuth code")
		http.Redirect(w, r, h.cfg.Server.AllowedOrigins[0]+"/login?error=exchange_failed", http.StatusTemporaryRedirect)
		return
	}

	// Get user info from provider
	userInfo, err := h.oauthService.GetUserInfo(r.Context(), provider, token)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user info from OAuth provider")
		http.Redirect(w, r, h.cfg.Server.AllowedOrigins[0]+"/login?error=user_info_failed", http.StatusTemporaryRedirect)
		return
	}

	// Check if OAuth provider is already linked to a user
	existingOAuth, err := h.oauthRepo.GetByProviderAndUserID(r.Context(), provider, userInfo.ProviderUserID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to check existing OAuth provider")
		http.Redirect(w, r, h.cfg.Server.AllowedOrigins[0]+"/login?error=db_error", http.StatusTemporaryRedirect)
		return
	}

	var user *models.User

	if existingOAuth != nil {
		// User already exists with this OAuth provider
		user, err = h.userRepo.GetByID(r.Context(), existingOAuth.UserID)
		if err != nil {
			log.Error().Err(err).Msg("Failed to get user")
			http.Redirect(w, r, h.cfg.Server.AllowedOrigins[0]+"/login?error=user_not_found", http.StatusTemporaryRedirect)
			return
		}

		// Update tokens
		tokenExpiry := token.Expiry
		err = h.oauthRepo.UpdateTokens(r.Context(), existingOAuth.ID, token.AccessToken, token.RefreshToken, &tokenExpiry)
		if err != nil {
			log.Error().Err(err).Msg("Failed to update OAuth tokens")
			// Don't fail login if token update fails
		}
	} else {
		// Check if user exists with this email
		user, err = h.userRepo.GetByEmail(r.Context(), userInfo.Email)
		if err != nil {
			// User doesn't exist, create new user
			user = &models.User{
				Email:         userInfo.Email,
				Name:          userInfo.Name,
				EmailVerified: true, // OAuth emails are pre-verified
				AvatarURL:     userInfo.AvatarURL,
			}

			if err := h.userRepo.Create(r.Context(), user); err != nil {
				log.Error().Err(err).Msg("Failed to create user")
				http.Redirect(w, r, h.cfg.Server.AllowedOrigins[0]+"/login?error=create_user_failed", http.StatusTemporaryRedirect)
				return
			}
		}

		// Link OAuth provider to user
		tokenExpiry := token.Expiry
		oauthProvider := &models.OAuthProvider{
			UserID:         user.ID,
			Provider:       provider,
			ProviderUserID: userInfo.ProviderUserID,
			Email:          userInfo.Email,
			Name:           userInfo.Name,
			AvatarURL:      userInfo.AvatarURL,
			AccessToken:    token.AccessToken,
			RefreshToken:   token.RefreshToken,
			TokenExpiry:    &tokenExpiry,
		}

		if err := h.oauthRepo.Create(r.Context(), oauthProvider); err != nil {
			log.Error().Err(err).Msg("Failed to create OAuth provider")
			// Don't fail login if OAuth provider creation fails
		}
	}

	// Create session
	session, err := auth.CreateSession(r.Context(), h.sessionRepo, user.ID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create session")
		http.Redirect(w, r, h.cfg.Server.AllowedOrigins[0]+"/login?error=session_failed", http.StatusTemporaryRedirect)
		return
	}

	// Generate JWT tokens
	accessToken, err := auth.GenerateAccessToken(user.ID, h.cfg.JWT.Secret, h.cfg.JWT.AccessTokenDuration)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate access token")
		http.Redirect(w, r, h.cfg.Server.AllowedOrigins[0]+"/login?error=token_failed", http.StatusTemporaryRedirect)
		return
	}

	refreshToken, err := auth.GenerateRefreshToken()
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate refresh token")
		http.Redirect(w, r, h.cfg.Server.AllowedOrigins[0]+"/login?error=token_failed", http.StatusTemporaryRedirect)
		return
	}

	// Update session with refresh token
	if err := h.sessionRepo.UpdateRefreshToken(r.Context(), session.ID, refreshToken); err != nil {
		log.Error().Err(err).Msg("Failed to update session refresh token")
		http.Redirect(w, r, h.cfg.Server.AllowedOrigins[0]+"/login?error=session_update_failed", http.StatusTemporaryRedirect)
		return
	}

	// Encode tokens as JSON for URL
	tokensJSON, _ := json.Marshal(map[string]string{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
	tokensEncoded := base64.URLEncoding.EncodeToString(tokensJSON)

	// Redirect to frontend with tokens in URL fragment (not query params for security)
	redirectURL := h.cfg.Server.AllowedOrigins[0] + "/auth/callback?tokens=" + tokensEncoded
	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
}

func generateRandomState() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
