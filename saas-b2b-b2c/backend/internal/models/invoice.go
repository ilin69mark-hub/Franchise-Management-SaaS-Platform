package models

import (
    "time"

    "github.com/google/uuid"
)

// Invoice - счет на оплату
type Invoice struct {
    ID          uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
    TenantID    uuid.UUID  `json:"tenant_id" gorm:"type:uuid;not null"`
    Tenant      Tenant     `json:"tenant" gorm:"foreignKey:TenantID"`
    Amount      float64    `json:"amount" gorm:"not null"`
    Description string     `json:"description"`
    Status      string     `json:"status" gorm:"default:'pending'"` // pending, paid, failed
    DueDate     time.Time  `json:"due_date"`
    PaidAt      *time.Time `json:"paid_at"`
    CreatedAt   time.Time  `json:"created_at"`
}

func (Invoice) TableName() string {
    return "invoices"
}
