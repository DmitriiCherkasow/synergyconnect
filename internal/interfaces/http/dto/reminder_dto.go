package dto

import (
	"time"

	"github.com/DmitriiCherkasow/synergyconnect.git/internal/domain"
)

// CreateReminderRequest — запрос на создание напоминания
type CreateReminderRequest struct {
	RemindAt           time.Time `json:"remind_at" binding:"required"`
	Recurrence         string    `json:"recurrence" binding:"omitempty,oneof=one_time daily weekly custom"`
	RecurrenceInterval *int      `json:"recurrence_interval,omitempty"` // в минутах, для custom
	WarningMinutes     []int     `json:"warning_minutes,omitempty"`     // например [15, 60, 1440]
}

// UpdateReminderRequest — запрос на обновление напоминания
type UpdateReminderRequest struct {
	RemindAt   *time.Time `json:"remind_at"`
	Recurrence *string    `json:"recurrence" binding:"omitempty,oneof=one_time daily weekly custom"`
}

// SnoozeReminderRequest — запрос на откладывание напоминания
type SnoozeReminderRequest struct {
	Minutes int `json:"minutes" binding:"required,min=1,max=1440"` // от 1 минуты до 24 часов
}

// ReminderResponse — ответ с данными напоминания
type ReminderResponse struct {
	ID                 string    `json:"id"`
	StickerID          string    `json:"sticker_id"`
	StickerTitle       string    `json:"sticker_title"`
	UserID             string    `json:"user_id"`
	RemindAt           time.Time `json:"remind_at"`
	Recurrence         string    `json:"recurrence"`
	RecurrenceInterval *int      `json:"recurrence_interval,omitempty"`
	WarningMinutes     []int     `json:"warning_minutes,omitempty"`
	IsSent             bool      `json:"is_sent"`
	SentAt             *time.Time `json:"sent_at,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// ToReminderResponse преобразует доменную модель в ответ
func ToReminderResponse(reminder *domain.Reminder, stickerTitle string) ReminderResponse {
	return ReminderResponse{
		ID:                 reminder.ID.String(),
		StickerID:          reminder.StickerID.String(),
		StickerTitle:       stickerTitle,
		UserID:             reminder.UserID.String(),
		RemindAt:           reminder.RemindAt,
		Recurrence:         string(reminder.Recurrence),
		RecurrenceInterval: reminder.RecurrenceInterval,
		WarningMinutes:     reminder.WarningMinutes,
		IsSent:             reminder.IsSent,
		SentAt:             reminder.SentAt,
		CreatedAt:          reminder.CreatedAt,
		UpdatedAt:          reminder.UpdatedAt,
	}
}