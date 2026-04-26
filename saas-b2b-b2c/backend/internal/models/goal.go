package models

import (
	"time"

	"github.com/google/uuid"
)

type Goal struct {
	ID           uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	AssignerID   uuid.UUID  `json:"assigner_id" gorm:"type:uuid;index"`         // кто создал цель
	AssigneeID   uuid.UUID  `json:"assignee_id" gorm:"type:uuid;index"`         // кому цель
	Role         string     `json:"role"`                                       // роль получателя (franchise_manager, dealer, dealer_manager, salon_manager)
	TenantID     *uuid.UUID `json:"tenant_id,omitempty" gorm:"type:uuid;index"` // если нужен (можно оставить nil)
	SalesPlan    float64    `json:"sales_plan"`
	LeadsPlan    int        `json:"leads_plan"`
	CallsPlan    int        `json:"calls_plan"`
	MeetingsPlan int        `json:"meetings_plan"`
	TargetDate   time.Time  `json:"target_date"`
	CreatedAt    time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}
