package domain

import (
	"time"
	"github.com/google/uuid"
)

// ReminderEmailStatus — статус отправки email
type ReminderEmailStatus string

const (
	EmailStatusSent    ReminderEmailStatus = "sent"
	EmailStatusFailed  ReminderEmailStatus = "failed"
	EmailStatusPending ReminderEmailStatus = "pending"
)

// ReminderEmail — история отправленных email-напоминаний
type ReminderEmail struct {
	ID          uuid.UUID            `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	ReminderID  uuid.UUID            `json:"reminder_id" gorm:"type:uuid;not null"`
	UserID      uuid.UUID            `json:"user_id" gorm:"type:uuid;not null"`
	EmailTo     string               `json:"email_to" gorm:"size:255;not null"`
	Subject     string               `json:"subject" gorm:"size:255;not null"`
	Body        string               `json:"body" gorm:"type:text"`
	Status      ReminderEmailStatus  `json:"status" gorm:"default:'pending'"`
	RetryCount  int                  `json:"retry_count" gorm:"default:0"`
	Error       string               `json:"error,omitempty" gorm:"type:text"`
	SentAt      *time.Time           `json:"sent_at,omitempty"`
	CreatedAt   time.Time            `json:"created_at" gorm:"autoCreateTime"`

	// Связи
	Reminder Reminder `json:"reminder" gorm:"foreignKey:ReminderID"`
	User     User     `json:"user" gorm:"foreignKey:UserID"`
}