package models

import (
	"time"

	"github.com/google/uuid"
)

// Dealer представляет бизнес-сущность дилера (владельца салонов)
type Dealer struct {
	ID       uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TenantID uuid.UUID `json:"tenant_id" gorm:"type:uuid;not null"` // Принадлежность к Франшизе
	UserID   uuid.UUID `json:"user_id" gorm:"type:uuid;not null"`   // Связь с аккаунтом пользователя (владельца)

	Name   string `json:"name" gorm:"not null"` // Название компании дилера
	Phone  string `json:"phone"`
	Email  string `json:"email"`
	Status string `json:"status" gorm:"default:'active'"` // active, blocked

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Dealer) TableName() string {
	return "dealers"
}
