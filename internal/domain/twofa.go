package domain

import (
    "time"
    "github.com/google/uuid"
)

type UserTwoFA struct {
    UserID        uuid.UUID   `json:"user_id" gorm:"type:uuid;primary_key"`
    User          User        `json:"user,omitempty" gorm:"foreignKey:UserID"`
    Secret        string      `json:"-" gorm:"size:255;not null"`
    Enabled       bool        `json:"enabled" gorm:"default:false"`
    RecoveryCodes []string    `json:"recovery_codes,omitempty" gorm:"type:text[]"`
    CreatedAt     time.Time   `json:"created_at"`
    UpdatedAt     time.Time   `json:"updated_at"`
}

type UserDevice struct {
    ID          uuid.UUID  `json:"id" gorm:"type:uuid;primary_key"`
    UserID      uuid.UUID  `json:"user_id" gorm:"type:uuid;not null"`
    User        User       `json:"user,omitempty" gorm:"foreignKey:UserID"`
    DeviceName  string     `json:"device_name" gorm:"size:255"`
    DeviceType  string     `json:"device_type" gorm:"size:50"`
    LastUsedAt  *time.Time `json:"last_used_at,omitempty"`
    CreatedAt   time.Time  `json:"created_at"`
}