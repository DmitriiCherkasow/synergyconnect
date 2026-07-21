package database

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/DmitriiCherkasow/synergyconnect.git/internal/domain"
)

// BoardRepository — репозиторий для работы с досками
type BoardRepository struct {
	db *gorm.DB
}

// NewBoardRepository создает новый репозиторий
func NewBoardRepository(db *gorm.DB) *BoardRepository {
	return &BoardRepository{db: db}
}

// Create создает новую доску
func (r *BoardRepository) Create(ctx context.Context, board *domain.Board) error {
	return r.db.WithContext(ctx).Create(board).Error
}

// FindByID ищет доску по ID
func (r *BoardRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Board, error) {
	var board domain.Board
	err := r.db.WithContext(ctx).
		Preload("Owner").
		Preload("Group").
		Preload("Stickers").
		First(&board, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &board, err
}

// FindByOwnerID возвращает все доски пользователя (личные)
func (r *BoardRepository) FindByOwnerID(ctx context.Context, ownerID uuid.UUID) ([]domain.Board, error) {
	var boards []domain.Board
	err := r.db.WithContext(ctx).
		Where("owner_id = ? AND group_id IS NULL AND is_archived = ?", ownerID, false).
		Preload("Owner").
		Order("created_at DESC").
		Find(&boards).Error
	return boards, err
}

// FindByGroupID возвращает все доски группы
func (r *BoardRepository) FindByGroupID(ctx context.Context, groupID uuid.UUID) ([]domain.Board, error) {
	var boards []domain.Board
	err := r.db.WithContext(ctx).
		Where("group_id = ? AND is_archived = ?", groupID, false).
		Preload("Owner").
		Preload("Group").
		Order("created_at DESC").
		Find(&boards).Error
	return boards, err
}

// Update обновляет доску
func (r *BoardRepository) Update(ctx context.Context, board *domain.Board) error {
	return r.db.WithContext(ctx).Save(board).Error
}

// Delete удаляет доску (каскадно удаляет стикеры)
func (r *BoardRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.Board{}, "id = ?", id).Error
}

// Archive архивирует доску
func (r *BoardRepository) Archive(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&domain.Board{}).
		Where("id = ?", id).
		Update("is_archived", true).Error
}

// Unarchive разархивирует доску
func (r *BoardRepository) Unarchive(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&domain.Board{}).
		Where("id = ?", id).
		Update("is_archived", false).Error
}