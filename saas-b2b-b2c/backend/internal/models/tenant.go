package models

import (
    "time"

    "github.com/google/uuid"
    "gorm.io/gorm"
)

type Tenant struct {
    ID     uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
    Name   string     `json:"name" gorm:"not null"`
    PlanID *uuid.UUID `json:"plan_id" gorm:"type:uuid"`
    Status string     `json:"status" gorm:"default:'active'"`

    // Биллинг
    PaidUntil       *time.Time `json:"paid_until"`
    GracePeriodDays int        `json:"grace_period_days" gorm:"default:7"`

    // Soft Delete (для аналитики Churn)
    DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`

    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

func (Tenant) TableName() string {
    return "tenants"
}
