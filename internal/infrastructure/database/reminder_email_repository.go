package database

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/DmitriiCherkasow/synergyconnect.git/internal/domain"
)

// ReminderEmailRepository — репозиторий для работы с email-логами
type ReminderEmailRepository struct {
	db *gorm.DB
}

// NewReminderEmailRepository создает новый репозиторий
func NewReminderEmailRepository(db *gorm.DB) *ReminderEmailRepository {
	return &ReminderEmailRepository{db: db}
}

// Create создает запись об отправленном email
func (r *ReminderEmailRepository) Create(ctx context.Context, log *domain.ReminderEmail) error {
	return r.db.WithContext(ctx).Create(log).Error
}

// MarkAsSent отмечает email как отправленный
func (r *ReminderEmailRepository) MarkAsSent(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&domain.ReminderEmail{}).
		Where("id = ?", id).
		Update("status", domain.EmailStatusSent).Error
}

// MarkAsFailed отмечает email как неудачный
func (r *ReminderEmailRepository) MarkAsFailed(ctx context.Context, id uuid.UUID, errMsg string) error {
	return r.db.WithContext(ctx).
		Model(&domain.ReminderEmail{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status": domain.EmailStatusFailed,
			"error":  errMsg,
		}).Error
}

// FindByReminderID возвращает все email-логи для напоминания
func (r *ReminderEmailRepository) FindByReminderID(ctx context.Context, reminderID uuid.UUID) ([]domain.ReminderEmail, error) {
	var logs []domain.ReminderEmail
	err := r.db.WithContext(ctx).
		Where("reminder_id = ?", reminderID).
		Order("created_at DESC").
		Find(&logs).Error
	return logs, err
}