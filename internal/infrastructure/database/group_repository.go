package database

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/DmitriiCherkasow/synergyconnect.git/internal/domain"
)

// GroupRepository — репозиторий для работы с группами
type GroupRepository struct {
	db *gorm.DB
}

// NewGroupRepository создает новый репозиторий
func NewGroupRepository(db *gorm.DB) *GroupRepository {
	return &GroupRepository{db: db}
}

// Create создает новую группу
func (r *GroupRepository) Create(ctx context.Context, group *domain.Group) error {
	return r.db.WithContext(ctx).Create(group).Error
}

// FindByID ищет группу по ID
func (r *GroupRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Group, error) {
	var group domain.Group
	err := r.db.WithContext(ctx).First(&group, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &group, err
}

// FindBySlug ищет группу по slug
func (r *GroupRepository) FindBySlug(ctx context.Context, slug string) (*domain.Group, error) {
	var group domain.Group
	err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&group).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &group, err
}

// GetTree возвращает дерево групп
func (r *GroupRepository) GetTree(ctx context.Context) ([]domain.Group, error) {
	var groups []domain.Group
	err := r.db.WithContext(ctx).
		Where("parent_id IS NULL").
		Preload("Children").
		Order("name ASC").
		Find(&groups).Error
	return groups, err
}

// GetChildren возвращает дочерние группы
func (r *GroupRepository) GetChildren(ctx context.Context, parentID uuid.UUID) ([]domain.Group, error) {
	var children []domain.Group
	err := r.db.WithContext(ctx).
		Where("parent_id = ?", parentID).
		Order("name ASC").
		Find(&children).Error
	return children, err
}

// Update обновляет группу
func (r *GroupRepository) Update(ctx context.Context, group *domain.Group) error {
	return r.db.WithContext(ctx).Save(group).Error
}

// Delete удаляет группу
func (r *GroupRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.Group{}, "id = ?", id).Error
}