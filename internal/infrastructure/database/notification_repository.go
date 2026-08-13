package database

import (
    "context"
    "errors"
    "time"

    "github.com/DmitriiCherkasow/synergyconnect.git/internal/domain"
    "github.com/google/uuid"
    "gorm.io/gorm"
)

type notificationRepository struct {
    db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) domain.NotificationRepository {
    return &notificationRepository{db: db}
}

// Create implements domain.NotificationRepository
func (r *notificationRepository) Create(ctx context.Context, notification *domain.Notification) error {
    return r.db.WithContext(ctx).Create(notification).Error
}

// CreateBatch implements domain.NotificationRepository
func (r *notificationRepository) CreateBatch(ctx context.Context, notifications []domain.Notification) error {
    if len(notifications) == 0 {
        return nil
    }
    return r.db.WithContext(ctx).CreateInBatches(notifications, 100).Error
}

// GetByID implements domain.NotificationRepository
func (r *notificationRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Notification, error) {
    var notification domain.Notification
    err := r.db.WithContext(ctx).
        Where("id = ?", id).
        First(&notification).Error
    
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, domain.ErrNotificationNotFound
        }
        return nil, err
    }
    return &notification, nil
}

// GetByUser implements domain.NotificationRepository
func (r *notificationRepository) GetByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.Notification, int64, error) {
    var notifications []domain.Notification
    var total int64

    query := r.db.WithContext(ctx).
        Where("user_id = ?", userID)

    // Подсчитываем общее количество
    if err := query.Model(&domain.Notification{}).Count(&total).Error; err != nil {
        return nil, 0, err
    }

    err := query.
        Order("created_at DESC").
        Limit(limit).
        Offset(offset).
        Find(&notifications).Error

    return notifications, total, err
}

// GetUnreadByUser implements domain.NotificationRepository
func (r *notificationRepository) GetUnreadByUser(ctx context.Context, userID uuid.UUID) ([]domain.Notification, error) {
    var notifications []domain.Notification
    err := r.db.WithContext(ctx).
        Where("user_id = ? AND is_read = ?", userID, false).
        Order("created_at DESC").
        Find(&notifications).Error
    return notifications, err
}

// GetUnreadCount implements domain.NotificationRepository
func (r *notificationRepository) GetUnreadCount(ctx context.Context, userID uuid.UUID) (int64, error) {
    var count int64
    err := r.db.WithContext(ctx).
        Model(&domain.Notification{}).
        Where("user_id = ? AND is_read = ?", userID, false).
        Count(&count).Error
    return count, err
}

// MarkAsRead implements domain.NotificationRepository
func (r *notificationRepository) MarkAsRead(ctx context.Context, id uuid.UUID) error {
    now := time.Now()
    return r.db.WithContext(ctx).
        Model(&domain.Notification{}).
        Where("id = ?", id).
        Updates(map[string]interface{}{
            "is_read": true,
            "read_at": now,
        }).Error
}

// MarkAllAsRead implements domain.NotificationRepository
func (r *notificationRepository) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
    now := time.Now()
    return r.db.WithContext(ctx).
        Model(&domain.Notification{}).
        Where("user_id = ? AND is_read = ?", userID, false).
        Updates(map[string]interface{}{
            "is_read": true,
            "read_at": now,
        }).Error
}

// MarkAllAsReadByType implements domain.NotificationRepository
func (r *notificationRepository) MarkAllAsReadByType(ctx context.Context, userID uuid.UUID, notificationType domain.NotificationType) error {
    now := time.Now()
    return r.db.WithContext(ctx).
        Model(&domain.Notification{}).
        Where("user_id = ? AND type = ? AND is_read = ?", userID, notificationType, false).
        Updates(map[string]interface{}{
            "is_read": true,
            "read_at": now,
        }).Error
}

// Delete implements domain.NotificationRepository
func (r *notificationRepository) Delete(ctx context.Context, id uuid.UUID) error {
    return r.db.WithContext(ctx).Delete(&domain.Notification{}, "id = ?", id).Error
}

// DeleteAllByUser implements domain.NotificationRepository
func (r *notificationRepository) DeleteAllByUser(ctx context.Context, userID uuid.UUID) error {
    return r.db.WithContext(ctx).Delete(&domain.Notification{}, "user_id = ?", userID).Error
}

// DeleteOldByUser implements domain.NotificationRepository
func (r *notificationRepository) DeleteOldByUser(ctx context.Context, userID uuid.UUID, olderThan time.Time) error {
    return r.db.WithContext(ctx).
        Where("user_id = ? AND created_at < ?", userID, olderThan).
        Delete(&domain.Notification{}).Error
}