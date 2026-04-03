package models

import (
	"time"

	"github.com/google/uuid"
)

// Task представляет задачу или чек-лист
// В вашей SQL миграции это таблица 'tasks'
type Task struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Title       string    `json:"title" gorm:"not null"`
	Description string    `json:"description"`

	// Иерархия и ответственные (как в вашем SQL файле)
	AssignedTo uuid.UUID  `json:"assigned_to" gorm:"type:uuid;not null"` // Кому поставлена задача (Менеджер)
	CreatedBy  uuid.UUID  `json:"created_by" gorm:"type:uuid;not null"`  // Кто поставил (Дилер/Франчайзер)
	SalonID    *uuid.UUID `json:"salon_id" gorm:"type:uuid"`             // Касается какого салона

	// Статусы: pending, in_progress, review, completed, rejected
	Status  string     `json:"status" gorm:"default:'pending'"`
	DueDate *time.Time `json:"due_date"`

	CreatedAt time.Time `json:"created_at"`
}

func (Task) TableName() string {
	return "tasks"
}
