package models

import (
    "time"

    "github.com/google/uuid"
)

type UserLog struct {
    ID        uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
    UserID    *uuid.UUID `json:"user_id" gorm:"type:uuid;index"`
    TenantID  *uuid.UUID `json:"tenant_id" gorm:"type:uuid;index"`
    Action    string     `json:"action" gorm:"type:varchar(50)"`
    IPAddress string     `json:"ip_address"`
    UserAgent string     `json:"user_agent"`
    CreatedAt time.Time  `json:"created_at"`
}

func (UserLog) TableName() string {
    return "user_logs"
}
