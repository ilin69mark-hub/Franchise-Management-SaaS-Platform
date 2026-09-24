package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Plan struct {
	ID        uuid.UUID       `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name      string          `json:"name" gorm:"size:255;not null"`
	Price     decimal.Decimal `json:"price" gorm:"type:numeric(12,2);not null;default:0"`
	MaxSalons int             `json:"max_salons" gorm:"not null"`
	MaxUsers  int             `json:"max_users" gorm:"not null"`
	CreatedAt time.Time       `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time       `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt *time.Time      `json:"-" gorm:"index"` // soft‑delete, не будет в JSON‑ответе
}

func (Plan) TableName() string { return "plans" }
