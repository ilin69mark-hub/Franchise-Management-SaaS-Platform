package models

import (
	"time"

	"github.com/google/uuid"
)

type Order struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	SalonID     uuid.UUID `json:"salon_id" gorm:"type:uuid;not null"`
	CreatedBy   uuid.UUID `json:"created_by" gorm:"type:uuid;not null"`
	Status      string    `json:"status" gorm:"default:'new'"` // new, waiting_payment, in_progress...
	Description string    `json:"description"`
	TotalPrice  float64   `json:"total_price"`
	CreatedAt   time.Time `json:"created_at"`
}

func (Order) TableName() string {
	return "orders"
}
