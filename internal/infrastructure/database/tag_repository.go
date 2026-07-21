package database

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/DmitriiCherkasow/synergyconnect.git/internal/domain"
)

// TagRepository — репозиторий для работы с тегами
type TagRepository struct {
	db *gorm.DB
}

// NewTagRepository создает новый репозиторий
func NewTagRepository(db *gorm.DB) *TagRepository {
	return &TagRepository{db: db}
}

// FindBySlug ищет тег по slug
func (r *TagRepository) FindBySlug(ctx context.Context, slug string) (*domain.Tag, error) {
	var tag domain.Tag
	err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&tag).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &tag, err
}

// FindOrCreate находит или создает тег
func (r *TagRepository) FindOrCreate(ctx context.Context, name string) (*domain.Tag, error) {
	slug := strings.ToLower(strings.ReplaceAll(name, " ", "-"))

	var tag domain.Tag
	err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&tag).Error
	if err == nil {
		return &tag, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// Создаем новый тег
	tag = domain.Tag{
		ID:   uuid.New(),
		Name: name,
		Slug: slug,
	}
	if err := r.db.WithContext(ctx).Create(&tag).Error; err != nil {
		return nil, err
	}
	return &tag, nil
}