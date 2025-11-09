package v1

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/mahendrakalkura/saas-blueprint/internal/auth"
	"github.com/mahendrakalkura/saas-blueprint/internal/models"
	"github.com/mahendrakalkura/saas-blueprint/internal/repository"
	"github.com/mahendrakalkura/saas-blueprint/internal/storage"
	"github.com/rs/zerolog/log"
)

type FileHandler struct {
	fileRepo       *repository.FileRepository
	userRepo       *repository.UserRepository
	storageService *storage.Service
}

func NewFileHandler(fileRepo *repository.FileRepository, userRepo *repository.UserRepository, storageService *storage.Service) *FileHandler {
	return &FileHandler{
		fileRepo:       fileRepo,
		userRepo:       userRepo,
		storageService: storageService,
	}
}

type FileUploadResponse struct {
	ID        string    `json:"id"`
	Key       string    `json:"key"`
	Filename  string    `json:"filename"`
	MimeType  string    `json:"mime_type"`
	Size      int64     `json:"size"`
	URL       string    `json:"url"`
	CreatedAt time.Time `json:"created_at"`
}

// UploadFile handles file uploads
func (h *FileHandler) UploadFile(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := auth.GetUserFromContext(r)
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Parse multipart form (max 32MB)
	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		respondError(w, http.StatusBadRequest, "File too large or invalid multipart form")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		respondError(w, http.StatusBadRequest, "File is required")
		return
	}
	defer file.Close()

	// Validate file size (max 10MB)
	if header.Size > 10<<20 {
		respondError(w, http.StatusBadRequest, "File size exceeds 10MB limit")
		return
	}

	// Get mime type
	mimeType := header.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	// Upload to storage
	result, err := h.storageService.Upload(r.Context(), "uploads/"+header.Filename, file, header.Size, mimeType)
	if err != nil {
		log.Error().Err(err).Msg("Failed to upload file to storage")
		respondError(w, http.StatusInternalServerError, "Failed to upload file")
		return
	}

	// Save file metadata to database
	fileModel := &models.File{
		UserID:   userCtx.UserID,
		Key:      result.Key,
		Filename: header.Filename,
		MimeType: result.MimeType,
		Size:     result.Size,
	}

	if err := h.fileRepo.Create(r.Context(), fileModel); err != nil {
		log.Error().Err(err).Msg("Failed to save file metadata")
		// Try to delete from storage
		_ = h.storageService.Delete(r.Context(), result.Key)
		respondError(w, http.StatusInternalServerError, "Failed to save file metadata")
		return
	}

	// Generate presigned URL for download (valid for 1 hour)
	downloadURL, err := h.storageService.GetPresignedURL(r.Context(), result.Key, 1*time.Hour)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate presigned URL")
		downloadURL = ""
	}

	respondJSON(w, http.StatusCreated, FileUploadResponse{
		ID:        fileModel.ID,
		Key:       fileModel.Key,
		Filename:  fileModel.Filename,
		MimeType:  fileModel.MimeType,
		Size:      fileModel.Size,
		URL:       downloadURL,
		CreatedAt: fileModel.CreatedAt,
	})
}

// UploadAvatar handles avatar uploads
func (h *FileHandler) UploadAvatar(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := auth.GetUserFromContext(r)
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Parse multipart form (max 5MB)
	err := r.ParseMultipartForm(5 << 20)
	if err != nil {
		respondError(w, http.StatusBadRequest, "File too large or invalid multipart form")
		return
	}

	file, header, err := r.FormFile("avatar")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Avatar file is required")
		return
	}
	defer file.Close()

	// Validate file size (max 2MB for avatars)
	if header.Size > 2<<20 {
		respondError(w, http.StatusBadRequest, "Avatar size exceeds 2MB limit")
		return
	}

	// Get mime type
	mimeType := header.Header.Get("Content-Type")

	// Validate image type
	if mimeType != "image/jpeg" && mimeType != "image/png" && mimeType != "image/gif" && mimeType != "image/webp" {
		respondError(w, http.StatusBadRequest, "Avatar must be an image (JPEG, PNG, GIF, or WebP)")
		return
	}

	// Upload to storage
	result, err := h.storageService.UploadAvatar(r.Context(), userCtx.UserID, file, header.Size, mimeType)
	if err != nil {
		log.Error().Err(err).Msg("Failed to upload avatar to storage")
		respondError(w, http.StatusInternalServerError, "Failed to upload avatar")
		return
	}

	// Save file metadata to database
	fileModel := &models.File{
		UserID:   userCtx.UserID,
		Key:      result.Key,
		Filename: header.Filename,
		MimeType: result.MimeType,
		Size:     result.Size,
		Metadata: models.JSONMap{
			"type": "avatar",
		},
	}

	if err := h.fileRepo.Create(r.Context(), fileModel); err != nil {
		log.Error().Err(err).Msg("Failed to save avatar metadata")
		_ = h.storageService.Delete(r.Context(), result.Key)
		respondError(w, http.StatusInternalServerError, "Failed to save avatar metadata")
		return
	}

	// Update user avatar URL
	user, err := h.userRepo.GetByID(r.Context(), userCtx.UserID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user")
	} else {
		// Generate presigned URL (valid for 7 days)
		avatarURL, err := h.storageService.GetPresignedURL(r.Context(), result.Key, 7*24*time.Hour)
		if err == nil {
			user.AvatarURL = &avatarURL
			_ = h.userRepo.Update(r.Context(), user)
		}
	}

	// Generate presigned URL for immediate use (valid for 1 hour)
	downloadURL, err := h.storageService.GetPresignedURL(r.Context(), result.Key, 1*time.Hour)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate presigned URL")
		downloadURL = ""
	}

	respondJSON(w, http.StatusCreated, FileUploadResponse{
		ID:        fileModel.ID,
		Key:       fileModel.Key,
		Filename:  fileModel.Filename,
		MimeType:  fileModel.MimeType,
		Size:      fileModel.Size,
		URL:       downloadURL,
		CreatedAt: fileModel.CreatedAt,
	})
}

// GetFile returns file metadata and download URL
func (h *FileHandler) GetFile(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := auth.GetUserFromContext(r)
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	fileID := chi.URLParam(r, "id")
	if fileID == "" {
		respondError(w, http.StatusBadRequest, "File ID is required")
		return
	}

	file, err := h.fileRepo.GetByID(r.Context(), fileID)
	if err != nil {
		respondError(w, http.StatusNotFound, "File not found")
		return
	}

	// Check if user owns the file
	if file.UserID != userCtx.UserID {
		respondError(w, http.StatusForbidden, "Access denied")
		return
	}

	// Generate presigned URL (valid for 1 hour)
	downloadURL, err := h.storageService.GetPresignedURL(r.Context(), file.Key, 1*time.Hour)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate presigned URL")
		respondError(w, http.StatusInternalServerError, "Failed to generate download URL")
		return
	}

	respondJSON(w, http.StatusOK, FileUploadResponse{
		ID:        file.ID,
		Key:       file.Key,
		Filename:  file.Filename,
		MimeType:  file.MimeType,
		Size:      file.Size,
		URL:       downloadURL,
		CreatedAt: file.CreatedAt,
	})
}

// ListFiles returns paginated list of user's files
func (h *FileHandler) ListFiles(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := auth.GetUserFromContext(r)
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get pagination parameters
	page := 1
	pageSize := 20

	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	if ps := r.URL.Query().Get("page_size"); ps != "" {
		if parsed, err := strconv.Atoi(ps); err == nil && parsed > 0 && parsed <= 100 {
			pageSize = parsed
		}
	}

	offset := (page - 1) * pageSize

	files, err := h.fileRepo.ListByUserID(r.Context(), userCtx.UserID, pageSize, offset)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list files")
		respondError(w, http.StatusInternalServerError, "Failed to list files")
		return
	}

	totalCount, err := h.fileRepo.CountByUserID(r.Context(), userCtx.UserID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to count files")
		totalCount = 0
	}

	totalPages := int(totalCount) / pageSize
	if int(totalCount)%pageSize > 0 {
		totalPages++
	}

	// Generate presigned URLs for all files
	fileResponses := make([]FileUploadResponse, 0, len(files))
	for _, file := range files {
		downloadURL, _ := h.storageService.GetPresignedURL(r.Context(), file.Key, 1*time.Hour)

		fileResponses = append(fileResponses, FileUploadResponse{
			ID:        file.ID,
			Key:       file.Key,
			Filename:  file.Filename,
			MimeType:  file.MimeType,
			Size:      file.Size,
			URL:       downloadURL,
			CreatedAt: file.CreatedAt,
		})
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"data":        fileResponses,
		"page":        page,
		"page_size":   pageSize,
		"total_items": totalCount,
		"total_pages": totalPages,
	})
}

// DeleteFile deletes a file
func (h *FileHandler) DeleteFile(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := auth.GetUserFromContext(r)
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	fileID := chi.URLParam(r, "id")
	if fileID == "" {
		respondError(w, http.StatusBadRequest, "File ID is required")
		return
	}

	file, err := h.fileRepo.GetByID(r.Context(), fileID)
	if err != nil {
		respondError(w, http.StatusNotFound, "File not found")
		return
	}

	// Check if user owns the file
	if file.UserID != userCtx.UserID {
		respondError(w, http.StatusForbidden, "Access denied")
		return
	}

	// Delete from database (soft delete)
	if err := h.fileRepo.Delete(r.Context(), fileID); err != nil {
		log.Error().Err(err).Msg("Failed to delete file from database")
		respondError(w, http.StatusInternalServerError, "Failed to delete file")
		return
	}

	// Delete from storage (async, don't fail request if it fails)
	go func() {
		if err := h.storageService.Delete(r.Context(), file.Key); err != nil {
			log.Error().Err(err).Str("key", file.Key).Msg("Failed to delete file from storage")
		}
	}()

	respondJSON(w, http.StatusOK, map[string]string{"message": "File deleted successfully"})
}
