package models

import (
	"time"

	"github.com/google/uuid"
)

type GoalPeriod string

const (
	PeriodDay   GoalPeriod = "day"
	PeriodWeek  GoalPeriod = "week"
	PeriodMonth GoalPeriod = "month"
	PeriodYear  GoalPeriod = "year"
	PeriodCustom GoalPeriod = "custom"
)

type Goal struct {
	ID           uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	AssignerID   uuid.UUID  `json:"assigner_id" gorm:"type:uuid;index"`
	AssigneeID   uuid.UUID  `json:"assignee_id" gorm:"type:uuid;index"`
	Role        string    `json:"role"`
	TenantID    *uuid.UUID `json:"tenant_id,omitempty" gorm:"type:uuid;index"`
	SalesPlan   float64   `json:"sales_plan"`
	LeadsPlan   int       `json:"leads_plan"`
	CallsPlan   int       `json:"calls_plan"`
	MeetingsPlan int       `json:"meetings_plan"`
	Period     GoalPeriod `json:"period" gorm:"type:varchar(20);default:'day'"`
	StartDate   time.Time `json:"start_date" gorm:"type:date"`
	EndDate     time.Time `json:"end_date" gorm:"type:date"`
	TargetDate  time.Time `json:"target_date"` // deprecated, используется для совместимости
	CreatedAt  time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt  time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}
