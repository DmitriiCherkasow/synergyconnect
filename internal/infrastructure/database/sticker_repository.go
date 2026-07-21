package database

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/DmitriiCherkasow/synergyconnect.git/internal/domain"
)

// StickerRepository — репозиторий для работы со стикерами
type StickerRepository struct {
	db *gorm.DB
}

// NewStickerRepository создает новый репозиторий
func NewStickerRepository(db *gorm.DB) *StickerRepository {
	return &StickerRepository{db: db}
}

// Create создает новый стикер
func (r *StickerRepository) Create(ctx context.Context, sticker *domain.Sticker) error {
	return r.db.WithContext(ctx).Create(sticker).Error
}

// FindByID ищет стикер по ID
func (r *StickerRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Sticker, error) {
	var sticker domain.Sticker
	err := r.db.WithContext(ctx).
		Preload("Author").
		Preload("Board").
		First(&sticker, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &sticker, err
}

// FindByBoardID возвращает все стикеры доски
func (r *StickerRepository) FindByBoardID(ctx context.Context, boardID uuid.UUID) ([]domain.Sticker, error) {
	var stickers []domain.Sticker
	err := r.db.WithContext(ctx).
		Where("board_id = ?", boardID).
		Preload("Author").
		Order("position_y ASC, position_x ASC").
		Find(&stickers).Error
	return stickers, err
}

// FindByAuthorID возвращает все стикеры автора
func (r *StickerRepository) FindByAuthorID(ctx context.Context, authorID uuid.UUID) ([]domain.Sticker, error) {
	var stickers []domain.Sticker
	err := r.db.WithContext(ctx).
		Where("author_id = ?", authorID).
		Preload("Board").
		Order("created_at DESC").
		Find(&stickers).Error
	return stickers, err
}

// FindOverdue возвращает все просроченные невыполненные стикеры
func (r *StickerRepository) FindOverdue(ctx context.Context) ([]domain.Sticker, error) {
	var stickers []domain.Sticker
	err := r.db.WithContext(ctx).
		Where("due_date IS NOT NULL AND due_date < ? AND is_completed = ?", gorm.Expr("NOW()"), false).
		Preload("Author").
		Preload("Board").
		Find(&stickers).Error
	return stickers, err
}

// Update обновляет стикер
func (r *StickerRepository) Update(ctx context.Context, sticker *domain.Sticker) error {
	return r.db.WithContext(ctx).Save(sticker).Error
}

// Delete удаляет стикер
func (r *StickerRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.Sticker{}, "id = ?", id).Error
}

// MarkComplete отмечает стикер как выполненный
func (r *StickerRepository) MarkComplete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&domain.Sticker{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"is_completed": true,
			"updated_at":   gorm.Expr("NOW()"),
		}).Error
}

// UpdatePosition обновляет позицию стикера на доске
func (r *StickerRepository) UpdatePosition(ctx context.Context, id uuid.UUID, x, y int) error {
	return r.db.WithContext(ctx).
		Model(&domain.Sticker{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"position_x": x,
			"position_y": y,
			"updated_at": gorm.Expr("NOW()"),
		}).Error
}