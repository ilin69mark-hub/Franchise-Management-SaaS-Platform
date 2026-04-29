package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Role string

const (
	RoleSuperAdmin        Role = "super_admin"
	RoleFranchisor        Role = "franchiser"
	RoleFranchisorManager Role = "franchiser_manager"
	RoleDealer            Role = "dealer"
	RoleDealerManager     Role = "salon_manager"
)

type User struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Email        string    `json:"email" gorm:"not null"`
	PasswordHash string    `json:"-" gorm:"column:password_hash"`
	Role         Role      `json:"role" gorm:"type:varchar(50);not null"`
	Status      string    `json:"status" gorm:"default:'active'"`

	TenantID  *uuid.UUID `json:"tenant_id" gorm:"type:uuid"`
	SalonID   *uuid.UUID `json:"salon_id" gorm:"type:uuid;index"`
	ManagedBy *uuid.UUID `json:"managed_by" gorm:"type:uuid;column:managed_by"`

	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Phone     string `json:"phone"`

	// Soft Delete
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (User) TableName() string {
	return "users"
}

// === DTOs ===

type UserLoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type UserRegisterRequest struct {
	Email     string `json:"email" binding:"required"`
	Password  string `json:"password" binding:"required"`
	Role      Role   `json:"role"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Phone     string `json:"phone"`
}

type UserUpdateRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Phone     string `json:"phone"`
}

type AuthResponse struct {
	User         User   `json:"user"`
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

// === HR Module DTOs ===

type CreateEmployeeRequest struct {
	Email     string     `json:"email" binding:"required,email"`
	Password  string     `json:"password" binding:"required,min=6"`
	FirstName string     `json:"first_name" binding:"required"`
	LastName  string     `json:"last_name"`
	Phone     string     `json:"phone"`
	Role      Role       `json:"role" binding:"required"`
	ManagedBy *uuid.UUID `json:"managed_by"`
	// ДОБАВЛЕНО: Поле для указания ID сети (тенанта).
	// Обязательно для Супер-Админа при создании Франчайзера.
	TenantID *uuid.UUID `json:"tenant_id"`
}

type UpdateEmployeeRequest struct {
	FirstName string     `json:"first_name"`
	LastName  string     `json:"last_name"`
	Phone     string     `json:"phone"`
	Role      Role       `json:"role"`
	ManagedBy *uuid.UUID `json:"managed_by"`
}
