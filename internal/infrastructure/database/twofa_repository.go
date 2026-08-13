package database

import (
    "context"
    "errors"
    "time"

    "github.com/DmitriiCherkasow/synergyconnect.git/internal/domain"
    "github.com/google/uuid"
    "gorm.io/gorm"
)

type twofaRepository struct {
    db *gorm.DB
}

func NewTwoFARepository(db *gorm.DB) domain.TwoFARepository {
    return &twofaRepository{db: db}
}

// CreateOrUpdate implements domain.TwoFARepository
func (r *twofaRepository) CreateOrUpdate(ctx context.Context, twofa *domain.UserTwoFA) error {
    // Проверяем, существует ли запись
    var existing domain.UserTwoFA
    err := r.db.WithContext(ctx).Where("user_id = ?", twofa.UserID).First(&existing).Error
    
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            // Создаём новую
            return r.db.WithContext(ctx).Create(twofa).Error
        }
        return err
    }
    
    // Обновляем существующую
    twofa.CreatedAt = existing.CreatedAt
    return r.db.WithContext(ctx).Save(twofa).Error
}

// GetByUserID implements domain.TwoFARepository
func (r *twofaRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.UserTwoFA, error) {
    var twofa domain.UserTwoFA
    err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&twofa).Error
    
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, nil
        }
        return nil, err
    }
    return &twofa, nil
}

// Enable implements domain.TwoFARepository
func (r *twofaRepository) Enable(ctx context.Context, userID uuid.UUID, secret string) error {
    now := time.Now()
    return r.db.WithContext(ctx).
        Model(&domain.UserTwoFA{}).
        Where("user_id = ?", userID).
        Updates(map[string]interface{}{
            "secret":     secret,
            "enabled":    true,
            "updated_at": now,
        }).Error
}

// Disable implements domain.TwoFARepository
func (r *twofaRepository) Disable(ctx context.Context, userID uuid.UUID) error {
    now := time.Now()
    return r.db.WithContext(ctx).
        Model(&domain.UserTwoFA{}).
        Where("user_id = ?", userID).
        Updates(map[string]interface{}{
            "enabled":       false,
            "recovery_codes": []string{},
            "updated_at":    now,
        }).Error
}

// UpdateRecoveryCodes implements domain.TwoFARepository
func (r *twofaRepository) UpdateRecoveryCodes(ctx context.Context, userID uuid.UUID, codes []string) error {
    now := time.Now()
    return r.db.WithContext(ctx).
        Model(&domain.UserTwoFA{}).
        Where("user_id = ?", userID).
        Updates(map[string]interface{}{
            "recovery_codes": codes,
            "updated_at":     now,
        }).Error
}

// Delete implements domain.TwoFARepository
func (r *twofaRepository) Delete(ctx context.Context, userID uuid.UUID) error {
    return r.db.WithContext(ctx).Delete(&domain.UserTwoFA{}, "user_id = ?", userID).Error
}

// IsEnabled implements domain.TwoFARepository
func (r *twofaRepository) IsEnabled(ctx context.Context, userID uuid.UUID) (bool, error) {
    var twofa domain.UserTwoFA
    err := r.db.WithContext(ctx).Select("enabled").Where("user_id = ?", userID).First(&twofa).Error
    
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return false, nil
        }
        return false, err
    }
    return twofa.Enabled, nil
}