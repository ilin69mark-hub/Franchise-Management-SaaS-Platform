package models

import (
	"time"

	"github.com/google/uuid"
)

type NotificationType string

const (
	NotificationTypePayment     NotificationType = "payment"
	NotificationTypeSystem      NotificationType = "system"
	NotificationTypeMaintenance NotificationType = "maintenance"
)

type Notification struct {
	ID        uuid.UUID        `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TenantID  uuid.UUID        `json:"tenant_id" gorm:"type:uuid;index"` // Кому адресовано
	UserID    *uuid.UUID       `json:"user_id" gorm:"type:uuid;index"`   // Конкретному юзеру или всем сети?
	Type      NotificationType `json:"type"`
	Title     string           `json:"title"`
	Message   string           `json:"message"`
	IsRead    bool             `json:"is_read" gorm:"default:false"`
	Data      string           `json:"data"` // JSON строка с доп. данными (например, ID счета)
	CreatedAt time.Time        `json:"created_at"`
}

func (Notification) TableName() string {
	return "notifications"
}
