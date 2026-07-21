package application

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/DmitriiCherkasow/synergyconnect.git/internal/domain"
)

// ReminderService — сервис для работы с напоминаниями
type ReminderService struct {
	reminderRepo ReminderRepository
	stickerRepo  StickerRepository
}

// NewReminderService создает новый сервис
func NewReminderService(
	reminderRepo ReminderRepository,
	stickerRepo StickerRepository,
) *ReminderService {
	return &ReminderService{
		reminderRepo: reminderRepo,
		stickerRepo:  stickerRepo,
	}
}

// CreateReminderRequest — запрос на создание напоминания
type CreateReminderRequest struct {
	StickerID          uuid.UUID
	UserID             uuid.UUID
	RemindAt           time.Time
	Recurrence         domain.ReminderRecurrence
	RecurrenceInterval *int
	WarningMinutes     []int
}

// CreateReminder создает новое напоминание
func (s *ReminderService) CreateReminder(ctx context.Context, req CreateReminderRequest) (*domain.Reminder, error) {
	sticker, err := s.stickerRepo.FindByID(ctx, req.StickerID)
	if err != nil {
		return nil, err
	}
	if sticker == nil {
		return nil, errors.New("sticker not found")
	}

	if sticker.AuthorID != req.UserID {
		return nil, errors.New("access denied")
	}

	reminder := &domain.Reminder{
		StickerID:          req.StickerID,
		UserID:             req.UserID,
		RemindAt:           req.RemindAt,
		Recurrence:         req.Recurrence,
		RecurrenceInterval: req.RecurrenceInterval,
		WarningMinutes:     req.WarningMinutes,
		IsSent:             false,
	}

	if err := s.reminderRepo.Create(ctx, reminder); err != nil {
		return nil, err
	}

	return reminder, nil
}

// GetUserReminders возвращает все напоминания пользователя
func (s *ReminderService) GetUserReminders(ctx context.Context, userID uuid.UUID) ([]domain.Reminder, error) {
	return s.reminderRepo.FindByUserID(ctx, userID)
}

// DeleteReminder удаляет напоминание
func (s *ReminderService) DeleteReminder(ctx context.Context, reminderID, userID uuid.UUID) error {
	reminder, err := s.reminderRepo.FindByID(ctx, reminderID)
	if err != nil {
		return err
	}
	if reminder == nil {
		return errors.New("reminder not found")
	}

	if reminder.UserID != userID {
		return errors.New("access denied")
	}

	return s.reminderRepo.Delete(ctx, reminderID)
}

// SnoozeReminder откладывает напоминание
func (s *ReminderService) SnoozeReminder(ctx context.Context, reminderID, userID uuid.UUID, minutes int) (*domain.Reminder, error) {
	reminder, err := s.reminderRepo.FindByID(ctx, reminderID)
	if err != nil {
		return nil, err
	}
	if reminder == nil {
		return nil, errors.New("reminder not found")
	}

	if reminder.UserID != userID {
		return nil, errors.New("access denied")
	}

	if reminder.IsSent {
		return nil, errors.New("cannot snooze a reminder that has already been sent")
	}

	reminder.RemindAt = time.Now().Add(time.Duration(minutes) * time.Minute)
	reminder.IsSent = false
	reminder.SentAt = nil

	if err := s.reminderRepo.Update(ctx, reminder); err != nil {
		return nil, err
	}

	return reminder, nil
}

// GetSticker возвращает стикер по ID
func (s *ReminderService) GetSticker(ctx context.Context, stickerID uuid.UUID) (*domain.Sticker, error) {
	return s.stickerRepo.FindByID(ctx, stickerID)
}