package database

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/DmitriiCherkasow/synergyconnect.git/internal/domain"
)

// SubscriptionRepository — репозиторий для работы с подписками
type SubscriptionRepository struct {
	db *gorm.DB
}

// NewSubscriptionRepository создает новый репозиторий
func NewSubscriptionRepository(db *gorm.DB) *SubscriptionRepository {
	return &SubscriptionRepository{db: db}
}

// Create создает подписку
func (r *SubscriptionRepository) Create(ctx context.Context, sub *domain.Subscription) error {
	return r.db.WithContext(ctx).Create(sub).Error
}

// Delete удаляет подписку
func (r *SubscriptionRepository) Delete(ctx context.Context, subscriberID, targetID uuid.UUID, subType domain.SubscriptionType) error {
	query := r.db.WithContext(ctx).
		Where("subscriber_id = ? AND type = ?", subscriberID, subType)

	if subType == domain.SubscriptionUser {
		query = query.Where("target_user_id = ?", targetID)
	} else {
		query = query.Where("target_group_id = ?", targetID)
	}

	return query.Delete(&domain.Subscription{}).Error
}

// FindBySubscriber возвращает все подписки пользователя
func (r *SubscriptionRepository) FindBySubscriber(ctx context.Context, subscriberID uuid.UUID) ([]domain.Subscription, error) {
	var subs []domain.Subscription
	err := r.db.WithContext(ctx).
		Where("subscriber_id = ?", subscriberID).
		Preload("TargetUser").
		Preload("TargetGroup").
		Find(&subs).Error
	return subs, err
}

// Exists проверяет, существует ли подписка
func (r *SubscriptionRepository) Exists(ctx context.Context, subscriberID, targetID uuid.UUID, subType domain.SubscriptionType) (bool, error) {
	var count int64
	query := r.db.WithContext(ctx).
		Model(&domain.Subscription{}).
		Where("subscriber_id = ? AND type = ?", subscriberID, subType)

	if subType == domain.SubscriptionUser {
		query = query.Where("target_user_id = ?", targetID)
	} else {
		query = query.Where("target_group_id = ?", targetID)
	}

	err := query.Count(&count).Error
	return count > 0, err
}