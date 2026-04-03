package handlers

import (
    "net/http"
    "franchise-saas-backend/internal/services"

    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
)

type NotificationHandler struct {
    service *services.NotificationService
}

func NewNotificationHandler(service *services.NotificationService) *NotificationHandler {
    return &NotificationHandler{service: service}
}

func (h *NotificationHandler) GetMyNotifications(c *gin.Context) {
    // Получаем tenantID
    val, exists := c.Get("tenantID")
    
    // Если ключа нет или он пуст (например, super_admin), возвращаем пустой список
    if !exists || val == "" {
        c.JSON(http.StatusOK, []interface{}{}) // Пустой массив, чтобы фронтенд не падал
        return
    }

    // Безопасное приведение типов
    tenantIDStr, ok := val.(string)
    if !ok {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid tenant ID format"})
        return
    }

    tenantID, err := uuid.Parse(tenantIDStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid UUID"})
        return
    }

    notifications, err := h.service.GetNotifications(c.Request.Context(), tenantID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, notifications)
}

func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
    id := c.Param("id")
    uid, _ := uuid.Parse(id)

    if err := h.service.MarkAsRead(c.Request.Context(), uid); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"message": "Marked as read"})
}

func (h *NotificationHandler) MarkAllAsRead(c *gin.Context) {
    val, exists := c.Get("tenantID")
    if !exists || val == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Tenant ID required"})
        return
    }

    tenantIDStr, ok := val.(string)
    if !ok {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid tenant ID"})
        return
    }

    tenantID, _ := uuid.Parse(tenantIDStr)

    if err := h.service.MarkAllAsRead(c.Request.Context(), tenantID); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"message": "All marked as read"})
}
