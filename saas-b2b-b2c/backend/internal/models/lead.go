package models

import (
    "time"

    "github.com/google/uuid"
)

// Lead представляет клиента (лид) в CRM менеджера
type Lead struct {
    ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
    SalonID   uuid.UUID `json:"salon_id" gorm:"type:uuid;not null"`
    ManagerID uuid.UUID `json:"manager_id" gorm:"type:uuid;not null"` // Менеджер, который ведет

    // Данные клиента
    FullName string `json:"full_name" gorm:"not null"`
    Phone    string `json:"phone"`
    Email    string `json:"email"`

    // Интересы
    InterestProduct string  `json:"interest_product"`
    Budget          float64 `json:"budget"`

    // Воронка: new, contact, meeting, sale, archive
    Status string `json:"status" gorm:"default:'new'"`

    // Связи
    Activities []LeadActivity `json:"activities,omitempty"`

    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

func (Lead) TableName() string {
    return "leads"
}

// LeadActivity история взаимодействий с лидом
type LeadActivity struct {
    ID          uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
    LeadID      uuid.UUID `json:"lead_id" gorm:"type:uuid;not null"`
    UserID      uuid.UUID `json:"user_id" gorm:"type:uuid;not null"`
    Type        string    `json:"type"` // call, meeting, note
    Description string    `json:"description"`

    CreatedAt time.Time `json:"created_at"`
}

func (LeadActivity) TableName() string {
    return "lead_activities"
}

// === DTOs для Лидов ===

type CreateLeadRequest struct {
    FullName        string  `json:"full_name" binding:"required,min=1,max=255"`
    Phone           string  `json:"phone" binding:"max=20"`
    Email           string  `json:"email" binding:"omitempty,email,max=255"`
    InterestProduct string  `json:"interest_product" binding:"max=255"`
    Budget          float64 `json:"budget" binding:"gte=0"`
}

type UpdateLeadStatusRequest struct {
    Status string `json:"status" binding:"required"`
}

type AddLeadActivityRequest struct {
    Type        string `json:"type" binding:"required"`
    Description string `json:"description"`
}
