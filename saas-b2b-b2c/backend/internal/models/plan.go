package models

import (
	"time"

	"github.com/google/uuid"
)

type Plan struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name      string    `json:"name" gorm:"not null"`
	Price     float64   `json:"price"` // GORM сам сконвертирует numeric -> float64
	MaxSalons int       `json:"max_salons"`
	MaxUsers  int       `json:"max_users"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Plan) TableName() string {
	return "plans"
}
