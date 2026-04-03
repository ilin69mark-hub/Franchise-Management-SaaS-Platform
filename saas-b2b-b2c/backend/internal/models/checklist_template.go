package models

import (
	"time"

	"github.com/google/uuid"
)

// ChecklistTemplate шаблон чек-листа (создает Франчайзер/Дилер)
type ChecklistTemplate struct {
	ID        uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TenantID  *uuid.UUID `json:"tenant_id" gorm:"type:uuid"`
	CreatedBy uuid.UUID  `json:"created_by" gorm:"type:uuid;not null"`

	Title       string `json:"title" gorm:"not null"`
	Description string `json:"description"`
	Type        string `json:"type" gorm:"default:'daily'"` // daily, weekly

	// ЯВНО УКАЗЫВАЕМ СВЯЗЬ
	Items []ChecklistTemplateItem `json:"items" gorm:"foreignKey:TemplateID"`

	CreatedAt time.Time `json:"created_at"`
}

func (ChecklistTemplate) TableName() string {
	return "checklist_templates"
}

// ChecklistTemplateItem пункт шаблона
type ChecklistTemplateItem struct {
	ID             uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TemplateID     uuid.UUID `json:"template_id" gorm:"type:uuid;not null"`
	Text           string    `json:"text" gorm:"not null"`
	OrderIndex     int       `json:"order_index"`
	ValidationType string    `json:"validation_type" gorm:"default:'checkbox'"` // checkbox, number, photo

	CreatedAt time.Time `json:"created_at"`
}

func (ChecklistTemplateItem) TableName() string {
	return "checklist_template_items"
}

// AssignedChecklist назначенный чек-лист менеджеру
type AssignedChecklist struct {
	ID         uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TemplateID uuid.UUID `json:"template_id" gorm:"type:uuid;not null"`
	SalonID    uuid.UUID `json:"salon_id" gorm:"type:uuid;not null"`
	AssignedTo uuid.UUID `json:"assigned_to" gorm:"type:uuid;not null"`

	DueDate *time.Time `json:"due_date"`
	Status  string     `json:"status" gorm:"default:'active'"` // active, completed

	// ЯВНО УКАЗЫВАЕМ СВЯЗЬ
	Items []ChecklistResponse `json:"items" gorm:"foreignKey:AssignedChecklistID"`

	CreatedAt time.Time `json:"created_at"`
}

func (AssignedChecklist) TableName() string {
	return "assigned_checklists"
}

// ChecklistResponse ответ/выполнение пункта
type ChecklistResponse struct {
	ID                  uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	AssignedChecklistID uuid.UUID `json:"assigned_checklist_id" gorm:"type:uuid;not null"`
	ItemID              uuid.UUID `json:"item_id" gorm:"type:uuid;not null"`

	Value     string `json:"value"`
	Completed bool   `json:"completed"`

	CreatedAt time.Time `json:"created_at"`
}

func (ChecklistResponse) TableName() string {
	return "checklist_responses"
}
