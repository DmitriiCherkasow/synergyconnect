package database

import (
    "context"
    "errors"
    "time"

    "github.com/DmitriiCherkasow/synergyconnect.git/internal/domain"
    "github.com/google/uuid"
    "gorm.io/gorm"
)

type deviceRepository struct {
    db *gorm.DB
}

func NewDeviceRepository(db *gorm.DB) domain.DeviceRepository {
    return &deviceRepository{db: db}
}

// Create implements domain.DeviceRepository
func (r *deviceRepository) Create(ctx context.Context, device *domain.UserDevice) error {
    return r.db.WithContext(ctx).Create(device).Error
}

// GetByID implements domain.DeviceRepository
func (r *deviceRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.UserDevice, error) {
    var device domain.UserDevice
    err := r.db.WithContext(ctx).Where("id = ?", id).First(&device).Error
    
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, nil
        }
        return nil, err
    }
    return &device, nil
}

// GetByUserID implements domain.DeviceRepository
func (r *deviceRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]domain.UserDevice, error) {
    var devices []domain.UserDevice
    err := r.db.WithContext(ctx).
        Where("user_id = ?", userID).
        Order("last_used_at DESC, created_at DESC").
        Find(&devices).Error
    return devices, err
}

// GetByUserAndName implements domain.DeviceRepository
func (r *deviceRepository) GetByUserAndName(ctx context.Context, userID uuid.UUID, deviceName string) (*domain.UserDevice, error) {
    var device domain.UserDevice
    err := r.db.WithContext(ctx).
        Where("user_id = ? AND device_name = ?", userID, deviceName).
        First(&device).Error
    
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, nil
        }
        return nil, err
    }
    return &device, nil
}

// UpdateLastUsed implements domain.DeviceRepository
func (r *deviceRepository) UpdateLastUsed(ctx context.Context, id uuid.UUID) error {
    now := time.Now()
    return r.db.WithContext(ctx).
        Model(&domain.UserDevice{}).
        Where("id = ?", id).
        Update("last_used_at", now).Error
}

// Delete implements domain.DeviceRepository
func (r *deviceRepository) Delete(ctx context.Context, id uuid.UUID) error {
    return r.db.WithContext(ctx).Delete(&domain.UserDevice{}, "id = ?", id).Error
}

// DeleteByUserID implements domain.DeviceRepository
func (r *deviceRepository) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
    return r.db.WithContext(ctx).Delete(&domain.UserDevice{}, "user_id = ?", userID).Error
}

// CountByUserID implements domain.DeviceRepository
func (r *deviceRepository) CountByUserID(ctx context.Context, userID uuid.UUID) (int64, error) {
    var count int64
    err := r.db.WithContext(ctx).
        Model(&domain.UserDevice{}).
        Where("user_id = ?", userID).
        Count(&count).Error
    return count, err
}