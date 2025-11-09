package v1

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/mahendrakalkura/saas-blueprint/internal/auth"
	"github.com/mahendrakalkura/saas-blueprint/internal/repository"
	"github.com/rs/zerolog/log"
)

type NotificationHandler struct {
	notificationRepo *repository.NotificationRepository
}

func NewNotificationHandler(notificationRepo *repository.NotificationRepository) *NotificationHandler {
	return &NotificationHandler{
		notificationRepo: notificationRepo,
	}
}

func (h *NotificationHandler) ListNotifications(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := auth.GetUserFromContext(r)
	if !ok {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	// Parse pagination parameters
	limit := 20
	offset := 0

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	notifications, err := h.notificationRepo.ListByUserID(r.Context(), userCtx.UserID, limit, offset)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list notifications")
		http.Error(w, `{"error":"Failed to fetch notifications"}`, http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"notifications": notifications,
		"limit":         limit,
		"offset":        offset,
	})
}

func (h *NotificationHandler) GetUnreadCount(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := auth.GetUserFromContext(r)
	if !ok {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	count, err := h.notificationRepo.GetUnreadCount(r.Context(), userCtx.UserID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get unread count")
		http.Error(w, `{"error":"Failed to fetch unread count"}`, http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"count": count,
	})
}

func (h *NotificationHandler) MarkAsRead(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := auth.GetUserFromContext(r)
	if !ok {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	notificationID := chi.URLParam(r, "id")
	if notificationID == "" {
		http.Error(w, `{"error":"Notification ID required"}`, http.StatusBadRequest)
		return
	}

	// Verify notification belongs to user
	notification, err := h.notificationRepo.GetByID(r.Context(), notificationID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get notification")
		http.Error(w, `{"error":"Notification not found"}`, http.StatusNotFound)
		return
	}

	if notification.UserID != userCtx.UserID {
		http.Error(w, `{"error":"Forbidden"}`, http.StatusForbidden)
		return
	}

	if err := h.notificationRepo.MarkAsRead(r.Context(), notificationID); err != nil {
		log.Error().Err(err).Msg("Failed to mark notification as read")
		http.Error(w, `{"error":"Failed to mark as read"}`, http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Notification marked as read",
	})
}

func (h *NotificationHandler) MarkAllAsRead(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := auth.GetUserFromContext(r)
	if !ok {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	if err := h.notificationRepo.MarkAllAsRead(r.Context(), userCtx.UserID); err != nil {
		log.Error().Err(err).Msg("Failed to mark all notifications as read")
		http.Error(w, `{"error":"Failed to mark all as read"}`, http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "All notifications marked as read",
	})
}

func (h *NotificationHandler) DeleteNotification(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := auth.GetUserFromContext(r)
	if !ok {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	notificationID := chi.URLParam(r, "id")
	if notificationID == "" {
		http.Error(w, `{"error":"Notification ID required"}`, http.StatusBadRequest)
		return
	}

	// Verify notification belongs to user
	notification, err := h.notificationRepo.GetByID(r.Context(), notificationID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get notification")
		http.Error(w, `{"error":"Notification not found"}`, http.StatusNotFound)
		return
	}

	if notification.UserID != userCtx.UserID {
		http.Error(w, `{"error":"Forbidden"}`, http.StatusForbidden)
		return
	}

	if err := h.notificationRepo.Delete(r.Context(), notificationID); err != nil {
		log.Error().Err(err).Msg("Failed to delete notification")
		http.Error(w, `{"error":"Failed to delete notification"}`, http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Notification deleted",
	})
}
