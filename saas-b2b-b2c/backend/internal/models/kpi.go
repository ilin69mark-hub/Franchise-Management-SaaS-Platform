package models

import (
	"time"

	"github.com/google/uuid"
)

// DailyGoal - планы на день
type DailyGoal struct {
	ID         uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	SalonID    *uuid.UUID `json:"salon_id" gorm:"type:uuid;index"`
	UserID     *uuid.UUID `json:"user_id" gorm:"type:uuid;index"`
	TargetDate time.Time  `json:"target_date" gorm:"type:date;not null"`

	SalesPlan    float64 `json:"sales_plan"`
	LeadsPlan    int     `json:"leads_plan"`
	CallsPlan    int     `json:"calls_plan"`
	MeetingsPlan int     `json:"meetings_plan"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (DailyGoal) TableName() string {
	return "daily_goals"
}

// ScheduleEvent - расписание/задачи
type ScheduleEvent struct {
	ID      uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	SalonID uuid.UUID `json:"salon_id" gorm:"type:uuid;not null"`
	UserID  uuid.UUID `json:"user_id" gorm:"type:uuid;not null"`

	Title       string    `json:"title" gorm:"not null"`
	Description string    `json:"description"`
	Type        string    `json:"type" gorm:"default:'task'"` // meeting, call, task
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`

	Priority string `json:"priority" gorm:"default:'normal'"`
	Status   string `json:"status" gorm:"default:'planned'"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (ScheduleEvent) TableName() string {
	return "schedule_events"
}

// --- DTOs ---

type SetGoalRequest struct {
	SalonID    string `json:"salon_id"`
	UserID     string `json:"user_id"`
	TargetDate string `json:"target_date"` // YYYY-MM-DD

	SalesPlan    float64 `json:"sales_plan"`
	LeadsPlan    int     `json:"leads_plan"`
	CallsPlan    int     `json:"calls_plan"`
	MeetingsPlan int     `json:"meetings_plan"`
}

type CreateScheduleEventRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	Type        string `json:"type"`                          // meeting, call, task
	StartTime   string `json:"start_time" binding:"required"` // ISO format
	EndTime     string `json:"end_time"`
	Priority    string `json:"priority"`
	UserID      string `json:"user_id"` // ID менеджера, если создает дилер
}

// UpdateScheduleStatusRequest - для простого обновления статуса
type UpdateScheduleStatusRequest struct {
	Status string `json:"status" binding:"required"` // completed, postponed
}

// UpdateScheduleEventRequest - ДЛЯ ПОЛНОГО ОБНОВЛЕНИЯ (Добавлено)
type UpdateScheduleEventRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	StartTime   string `json:"start_time"`
	EndTime     string `json:"end_time"`
	Status      string `json:"status"`
	Priority    string `json:"priority"`
}

// --- Response Structs ---

type KPIItem struct {
	Plan    interface{} `json:"plan"` // float64 или int
	Fact    interface{} `json:"fact"`
	Percent int         `json:"percent"`
}

type DashboardStatsResponse struct {
	Date      string  `json:"date"`
	Sales     KPIItem `json:"sales"`
	Leads     KPIItem `json:"leads"`
	Calls     KPIItem `json:"calls"`
	Meetings  KPIItem `json:"meetings"`
	Checklist struct {
		Done  int `json:"done"`
		Total int `json:"total"`
	} `json:"checklist"`
}

// KPISummary - агрегированная статистика KPI
type KPISummary struct {
	TotalSales     float64 `json:"total_sales"`
	TotalLeads     int     `json:"total_leads"`
	ConversionRate float64 `json:"conversion_rate"`
	AvgCheck       float64 `json:"avg_check"`
	TotalCalls     int     `json:"total_calls"`
	TotalMeetings  int     `json:"total_meetings"`
}
