package models

import (
	"time"

	"github.com/google/uuid"
)

type Salon struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TenantID  uuid.UUID `json:"tenant_id" gorm:"type:uuid;not null"` // К какой сети принадлежит
	Name      string    `json:"name" gorm:"not null"`
	Address   string    `json:"address"`
	CreatedAt time.Time `json:"created_at"`

	// НОВОЕ ПОЛЕ: Связь с менеджером
	// gorm:"-" означает, что это поле не хранится в БД, оно вычисляется на лету.
	Manager *User `json:"manager" gorm:"-"`
}

func (Salon) TableName() string {
	return "salons"
}
