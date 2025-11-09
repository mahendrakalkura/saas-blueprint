package v1

import (
	"encoding/base64"
	"encoding/json"
	"net/http"

	"github.com/lib/pq"
	"github.com/mahendrakalkura/saas-blueprint/internal/auth"
	"github.com/mahendrakalkura/saas-blueprint/internal/mfa"
	"github.com/mahendrakalkura/saas-blueprint/internal/repository"
	"github.com/rs/zerolog/log"
)

type MFAHandler struct {
	mfaService *mfa.Service
	userRepo   *repository.UserRepository
}

func NewMFAHandler(mfaService *mfa.Service, userRepo *repository.UserRepository) *MFAHandler {
	return &MFAHandler{
		mfaService: mfaService,
		userRepo:   userRepo,
	}
}

// EnableMFA generates a new TOTP secret and QR code for the user
func (h *MFAHandler) EnableMFA(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := auth.GetUserFromContext(r)
	if !ok {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	// Get user from database
	user, err := h.userRepo.GetByID(r.Context(), userCtx.UserID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user")
		http.Error(w, `{"error":"Failed to get user"}`, http.StatusInternalServerError)
		return
	}

	// Check if MFA is already enabled
	if user.MFAEnabled {
		http.Error(w, `{"error":"MFA is already enabled"}`, http.StatusBadRequest)
		return
	}

	// Generate new TOTP secret
	qrData, err := h.mfaService.GenerateSecret(user.Email)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate TOTP secret")
		http.Error(w, `{"error":"Failed to generate MFA secret"}`, http.StatusInternalServerError)
		return
	}

	// Generate backup codes
	backupCodes, err := h.mfaService.GenerateBackupCodes(10)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate backup codes")
		http.Error(w, `{"error":"Failed to generate backup codes"}`, http.StatusInternalServerError)
		return
	}

	// Store secret temporarily (will be enabled after verification)
	// Update user with secret but don't enable MFA yet
	user.MFASecret = &qrData.Secret
	user.MFABackupCodes = backupCodes
	if err := h.userRepo.Update(r.Context(), user); err != nil {
		log.Error().Err(err).Msg("Failed to update user with MFA secret")
		http.Error(w, `{"error":"Failed to save MFA secret"}`, http.StatusInternalServerError)
		return
	}

	// Encode QR code to base64
	qrCodeBase64 := base64.StdEncoding.EncodeToString(qrData.QRCode)

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"secret":       qrData.Secret,
		"qr_code":      qrCodeBase64,
		"backup_codes": backupCodes,
	})
}

// VerifyAndActivateMFA verifies the TOTP code and enables MFA
func (h *MFAHandler) VerifyAndActivateMFA(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := auth.GetUserFromContext(r)
	if !ok {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	var req struct {
		Code string `json:"code"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Code == "" {
		http.Error(w, `{"error":"Code is required"}`, http.StatusBadRequest)
		return
	}

	// Get user from database
	user, err := h.userRepo.GetByID(r.Context(), userCtx.UserID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user")
		http.Error(w, `{"error":"Failed to get user"}`, http.StatusInternalServerError)
		return
	}

	// Check if user has a secret but MFA not yet enabled
	if user.MFASecret == nil {
		http.Error(w, `{"error":"MFA setup not initiated"}`, http.StatusBadRequest)
		return
	}

	if user.MFAEnabled {
		http.Error(w, `{"error":"MFA is already enabled"}`, http.StatusBadRequest)
		return
	}

	// Verify the code
	if !h.mfaService.VerifyCode(*user.MFASecret, req.Code) {
		http.Error(w, `{"error":"Invalid code"}`, http.StatusBadRequest)
		return
	}

	// Enable MFA
	user.MFAEnabled = true
	if err := h.userRepo.Update(r.Context(), user); err != nil {
		log.Error().Err(err).Msg("Failed to enable MFA")
		http.Error(w, `{"error":"Failed to enable MFA"}`, http.StatusInternalServerError)
		return
	}

	log.Info().Str("user_id", user.ID).Msg("MFA enabled for user")

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "MFA enabled successfully",
	})
}

// DisableMFA disables MFA for the user
func (h *MFAHandler) DisableMFA(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := auth.GetUserFromContext(r)
	if !ok {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	var req struct {
		Code string `json:"code"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Code == "" {
		http.Error(w, `{"error":"Code is required"}`, http.StatusBadRequest)
		return
	}

	// Get user from database
	user, err := h.userRepo.GetByID(r.Context(), userCtx.UserID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user")
		http.Error(w, `{"error":"Failed to get user"}`, http.StatusInternalServerError)
		return
	}

	if !user.MFAEnabled {
		http.Error(w, `{"error":"MFA is not enabled"}`, http.StatusBadRequest)
		return
	}

	// Verify the code before disabling
	if !h.mfaService.VerifyCode(*user.MFASecret, req.Code) {
		http.Error(w, `{"error":"Invalid code"}`, http.StatusBadRequest)
		return
	}

	// Disable MFA
	user.MFAEnabled = false
	user.MFASecret = nil
	user.MFABackupCodes = nil
	if err := h.userRepo.Update(r.Context(), user); err != nil {
		log.Error().Err(err).Msg("Failed to disable MFA")
		http.Error(w, `{"error":"Failed to disable MFA"}`, http.StatusInternalServerError)
		return
	}

	log.Info().Str("user_id", user.ID).Msg("MFA disabled for user")

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "MFA disabled successfully",
	})
}

// RegenerateBackupCodes generates new backup codes
func (h *MFAHandler) RegenerateBackupCodes(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := auth.GetUserFromContext(r)
	if !ok {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	var req struct {
		Code string `json:"code"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Code == "" {
		http.Error(w, `{"error":"Code is required"}`, http.StatusBadRequest)
		return
	}

	// Get user from database
	user, err := h.userRepo.GetByID(r.Context(), userCtx.UserID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user")
		http.Error(w, `{"error":"Failed to get user"}`, http.StatusInternalServerError)
		return
	}

	if !user.MFAEnabled {
		http.Error(w, `{"error":"MFA is not enabled"}`, http.StatusBadRequest)
		return
	}

	// Verify the code
	if !h.mfaService.VerifyCode(*user.MFASecret, req.Code) {
		http.Error(w, `{"error":"Invalid code"}`, http.StatusBadRequest)
		return
	}

	// Generate new backup codes
	backupCodes, err := h.mfaService.GenerateBackupCodes(10)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate backup codes")
		http.Error(w, `{"error":"Failed to generate backup codes"}`, http.StatusInternalServerError)
		return
	}

	// Update user with new backup codes
	user.MFABackupCodes = backupCodes
	if err := h.userRepo.Update(r.Context(), user); err != nil {
		log.Error().Err(err).Msg("Failed to update backup codes")
		http.Error(w, `{"error":"Failed to update backup codes"}`, http.StatusInternalServerError)
		return
	}

	log.Info().Str("user_id", user.ID).Msg("Backup codes regenerated")

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"backup_codes": backupCodes,
	})
}

// GetMFAStatus returns the MFA status for the user
func (h *MFAHandler) GetMFAStatus(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := auth.GetUserFromContext(r)
	if !ok {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	// Get user from database
	user, err := h.userRepo.GetByID(r.Context(), userCtx.UserID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user")
		http.Error(w, `{"error":"Failed to get user"}`, http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"mfa_enabled": user.MFAEnabled,
	})
}
