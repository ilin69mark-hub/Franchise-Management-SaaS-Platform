package models

import (
	"time"

	"github.com/google/uuid"
)

// Checklist представляет задачу/чек-лист
type Checklist struct {
	ID          uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Title       string     `json:"title" gorm:"not null"`
	Description string     `json:"description"`
	UserID      uuid.UUID  `json:"user_id" gorm:"type:uuid;not null"` // Кто создал
	TenantID    *uuid.UUID `json:"tenant_id" gorm:"type:uuid"`

	AssignedTo *uuid.UUID `json:"assigned_to" gorm:"type:uuid"` // Кому назначено
	StartDate  *time.Time `json:"start_date"`
	EndDate    *time.Time `json:"end_date"`

	// Статусы и Приоритеты
	Status   string `json:"status" gorm:"default:'pending'"`  // pending, in_progress, completed, waiting
	Priority string `json:"priority" gorm:"default:'normal'"` // urgent, important, normal

	// НОВОЕ ПОЛЕ: Повторение задачи (daily, weekly, monthly, or empty)
	Recurrence string `json:"recurrence" gorm:"type:varchar(20);default:''"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Вычисляемые поля (не хранятся в БД, но возвращаются в JSON)
	IsOverdue bool `json:"is_overdue" gorm:"-"`
}

func (Checklist) TableName() string {
	return "checklists"
}

// Request DTOs
type ChecklistCreateRequest struct {
	Title       string     `json:"title" binding:"required"`
	Description string     `json:"description"`
	AssignedTo  *uuid.UUID `json:"assigned_to"`
	StartDate   *time.Time `json:"start_date"`
	EndDate     *time.Time `json:"end_date"`
	Priority    string     `json:"priority"`   // urgent, important, normal
	Recurrence  string     `json:"recurrence"` // НОВОЕ ПОЛЕ: daily, weekly, monthly
}

type ChecklistUpdateRequest struct {
	Title       string     `json:"title"`
	Description string     `json:"description"`
	AssignedTo  *uuid.UUID `json:"assigned_to"`
	StartDate   *time.Time `json:"start_date"`
	EndDate     *time.Time `json:"end_date"`
	Status      string     `json:"status"`
	Priority    string     `json:"priority"`
	Recurrence  string     `json:"recurrence"` // НОВОЕ ПОЛЕ
}
