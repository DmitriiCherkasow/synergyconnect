package database

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/DmitriiCherkasow/synergyconnect.git/internal/domain"
)

// ReminderRepository — репозиторий для работы с напоминаниями
type ReminderRepository struct {
	db *gorm.DB
}

// NewReminderRepository создает новый репозиторий
func NewReminderRepository(db *gorm.DB) *ReminderRepository {
	return &ReminderRepository{db: db}
}

// Create создает новое напоминание
func (r *ReminderRepository) Create(ctx context.Context, reminder *domain.Reminder) error {
	return r.db.WithContext(ctx).Create(reminder).Error
}

// FindByID ищет напоминание по ID
func (r *ReminderRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Reminder, error) {
	var reminder domain.Reminder
	err := r.db.WithContext(ctx).
		Preload("Sticker").
		Preload("User").
		First(&reminder, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &reminder, nil
}

// FindPending возвращает все неотправленные напоминания, которые должны быть отправлены
func (r *ReminderRepository) FindPending(ctx context.Context) ([]domain.Reminder, error) {
	var reminders []domain.Reminder
	err := r.db.WithContext(ctx).
		Where("is_sent = ? AND remind_at <= ?", false, time.Now()).
		Preload("Sticker").
		Preload("User").
		Find(&reminders).Error
	return reminders, err
}

// FindByUserID возвращает все напоминания пользователя
func (r *ReminderRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Reminder, error) {
	var reminders []domain.Reminder
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Preload("Sticker").
		Order("remind_at ASC").
		Find(&reminders).Error
	return reminders, err
}

// FindByStickerID возвращает все напоминания для стикера
func (r *ReminderRepository) FindByStickerID(ctx context.Context, stickerID uuid.UUID) ([]domain.Reminder, error) {
	var reminders []domain.Reminder
	err := r.db.WithContext(ctx).
		Where("sticker_id = ?", stickerID).
		Find(&reminders).Error
	return reminders, err
}

// MarkAsSent отмечает напоминание как отправленное
func (r *ReminderRepository) MarkAsSent(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&domain.Reminder{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"is_sent":  true,
			"sent_at":  now,
		}).Error
}

// Update обновляет напоминание
func (r *ReminderRepository) Update(ctx context.Context, reminder *domain.Reminder) error {
	return r.db.WithContext(ctx).Save(reminder).Error
}

// Delete удаляет напоминание
func (r *ReminderRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.Reminder{}, "id = ?", id).Error
}

// DeleteByStickerID удаляет все напоминания для стикера
func (r *ReminderRepository) DeleteByStickerID(ctx context.Context, stickerID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("sticker_id = ?", stickerID).Delete(&domain.Reminder{}).Error
}