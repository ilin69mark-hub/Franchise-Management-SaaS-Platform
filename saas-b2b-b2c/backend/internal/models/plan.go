package models

import (
	"time"

	"github.com/google/uuid"
)

type Plan struct {
	ID        uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name      string     `json:"name" gorm:"size:255;not null"`
	Price     float64    `json:"price" gorm:"type:numeric;not null"`
	MaxSalons int        `json:"max_salons" gorm:"not null"`
	MaxUsers  int        `json:"max_users" gorm:"not null"`
	CreatedAt time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt *time.Time `json:"-" gorm:"index"` // soft‑delete, не будет в JSON‑ответе
}

func (Plan) TableName() string { return "plans" }
