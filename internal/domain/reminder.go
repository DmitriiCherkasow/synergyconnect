package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// ReminderRecurrence — тип повторения напоминания
type ReminderRecurrence string

const (
	RecurrenceOneTime ReminderRecurrence = "one_time"
	RecurrenceDaily   ReminderRecurrence = "daily"
	RecurrenceWeekly  ReminderRecurrence = "weekly"
	RecurrenceCustom  ReminderRecurrence = "custom"
)

// Reminder — модель напоминания для стикера
type Reminder struct {
	ID                 uuid.UUID           `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	StickerID          uuid.UUID           `json:"sticker_id" gorm:"type:uuid;not null"`
	UserID             uuid.UUID           `json:"user_id" gorm:"type:uuid;not null"`
	RemindAt           time.Time           `json:"remind_at" gorm:"not null"`
	Recurrence         ReminderRecurrence  `json:"recurrence" gorm:"default:'one_time'"`
	RecurrenceInterval *int                `json:"recurrence_interval,omitempty"`
	WarningMinutes     pq.Int64Array       `json:"warning_minutes" gorm:"type:integer[];default:'{}'"`
	IsSent             bool                `json:"is_sent" gorm:"default:false"`
	SentAt             *time.Time          `json:"sent_at,omitempty"`
	CreatedAt          time.Time           `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt          time.Time           `json:"updated_at" gorm:"autoUpdateTime"`

	// Связи
	Sticker Sticker `json:"sticker" gorm:"foreignKey:StickerID"`
	User    User    `json:"user" gorm:"foreignKey:UserID"`
}

// ShouldSend проверяет, нужно ли отправлять напоминание
func (r *Reminder) ShouldSend() bool {
	if r.IsSent {
		return false
	}
	return time.Now().After(r.RemindAt) || time.Now().Equal(r.RemindAt)
}

// CreateNextReminder создает следующее напоминание на основе recurrence
func (r *Reminder) CreateNextReminder() *Reminder {
	if r.Recurrence == RecurrenceOneTime {
		return nil
	}

	nextRemindAt := r.RemindAt
	switch r.Recurrence {
	case RecurrenceDaily:
		nextRemindAt = nextRemindAt.Add(24 * time.Hour)
	case RecurrenceWeekly:
		nextRemindAt = nextRemindAt.Add(7 * 24 * time.Hour)
	case RecurrenceCustom:
		if r.RecurrenceInterval != nil {
			nextRemindAt = nextRemindAt.Add(time.Duration(*r.RecurrenceInterval) * time.Minute)
		}
	default:
		return nil
	}

	return &Reminder{
		StickerID:          r.StickerID,
		UserID:             r.UserID,
		RemindAt:           nextRemindAt,
		Recurrence:         r.Recurrence,
		RecurrenceInterval: r.RecurrenceInterval,
		WarningMinutes:     r.WarningMinutes,
		IsSent:             false,
	}
}