package models

import (
	"encoding/json"
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
	Status       string    `json:"status" gorm:"default:'active'"`

	TenantID  *uuid.UUID `json:"tenant_id" gorm:"type:uuid"`
	SalonID   *uuid.UUID `json:"salon_id" gorm:"type:uuid;index"`
	ManagedBy *uuid.UUID `json:"managed_by" gorm:"type:uuid;column:managed_by"`

	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Phone     string `json:"phone"`

	DisplayName           string `json:"display_name" gorm:"type:varchar(255)"`
	Position              string `json:"position" gorm:"type:varchar(255)"`
	Bio                   string `json:"bio" gorm:"type:text"`
	Quote                 string `json:"quote" gorm:"type:varchar(255)"`
	AvatarURL             string `json:"avatar_url" gorm:"type:varchar(500)"`
	UserStatus            string `json:"user_status" gorm:"type:varchar(50);default:'online'"`
	AvailableForQuestions bool   `json:"available_for_questions" gorm:"default:true"`
	Achievements          string `json:"achievements" gorm:"type:text"` // JSON array stored as text

	ContactsEmailVisible bool   `json:"contacts_email_visible" gorm:"default:true"`
	ContactsPhoneVisible bool   `json:"contacts_phone_visible" gorm:"default:true"`
	ContactsPhone        string `json:"contacts_phone" gorm:"type:varchar(50)"`
	ContactsTelegram     string `json:"contacts_telegram" gorm:"type:varchar(100)"`
	ContactsWhatsApp     string `json:"contacts_whatsapp" gorm:"type:varchar(50)"`
	ContactsWorkingHours  string `json:"contacts_working_hours" gorm:"type:varchar(100)"`

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

	DisplayName           *string `json:"display_name"`
	Position              *string `json:"position"`
	Bio                   *string `json:"bio"`
	Quote                 *string `json:"quote"`
	UserStatus            *string `json:"status"`
	AvailableForQuestions *bool   `json:"available_for_questions"`
	ContactsEmailVisible  *bool   `json:"contacts_email_visible"`
	ContactsPhoneVisible  *bool   `json:"contacts_phone_visible"`
	ContactsPhone         *string `json:"contacts_phone"`
	ContactsTelegram      *string `json:"contacts_telegram"`
	ContactsWhatsApp      *string `json:"contacts_whatsapp"`
	ContactsWorkingHours  *string `json:"contacts_working_hours"`
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

type UserContacts struct {
	EmailVisible   bool   `json:"email_visible"`
	PhoneVisible   bool   `json:"phone_visible"`
	Phone         string `json:"phone"`
	Telegram      string `json:"telegram"`
	WhatsApp      string `json:"whatsapp"`
	WorkingHours  string `json:"working_hours"`
}

type UserProfileResponse struct {
	ID                    string       `json:"id"`
	Email                 string       `json:"email"`
	FirstName            string       `json:"first_name"`
	LastName             string       `json:"last_name"`
	DisplayName          string       `json:"display_name"`
	Position             string       `json:"position"`
	Bio                  string       `json:"bio"`
	Quote                string       `json:"quote"`
	AvatarURL            string       `json:"avatar_url"`
	Status               string       `json:"status"`
	AvailableForQuestions bool        `json:"available_for_questions"`
	Achievements         []string     `json:"achievements"`
	Contacts             UserContacts `json:"contacts"`
}

func (u *User) ToProfileResponse() UserProfileResponse {
	var achievements []string
	if u.Achievements != "" {
		_ = json.Unmarshal([]byte(u.Achievements), &achievements)
	}

	return UserProfileResponse{
		ID:          u.ID.String(),
		Email:       u.Email,
		FirstName:   u.FirstName,
		LastName:    u.LastName,
		DisplayName: u.DisplayName,
		Position:    u.Position,
		Bio:         u.Bio,
		Quote:       u.Quote,
		AvatarURL:   u.AvatarURL,
		Status:      u.UserStatus,
		AvailableForQuestions: u.AvailableForQuestions,
		Achievements: achievements,
		Contacts: UserContacts{
			EmailVisible:  u.ContactsEmailVisible,
			PhoneVisible:  u.ContactsPhoneVisible,
			Phone:        u.ContactsPhone,
			Telegram:     u.ContactsTelegram,
			WhatsApp:     u.ContactsWhatsApp,
			WorkingHours: u.ContactsWorkingHours,
		},
	}
}
