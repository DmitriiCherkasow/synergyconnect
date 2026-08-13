package database

import (
	"context"
	"errors"
	"strings"

	"github.com/DmitriiCherkasow/synergyconnect.git/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type vacancyRepository struct {
	db *gorm.DB
}

func NewVacancyRepository(db *gorm.DB) domain.VacancyRepository {
	return &vacancyRepository{db: db}
}

// Create implements domain.VacancyRepository
func (r *vacancyRepository) Create(ctx context.Context, vacancy *domain.Vacancy) error {
	return r.db.WithContext(ctx).Create(vacancy).Error
}

// GetByID implements domain.VacancyRepository
func (r *vacancyRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Vacancy, error) {
	var vacancy domain.Vacancy
	err := r.db.WithContext(ctx).
		Preload("Employer").
		Preload("Responses.User").
		Where("id = ?", id).
		First(&vacancy).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrVacancyNotFound
		}
		return nil, err
	}
	return &vacancy, nil
}

// Update implements domain.VacancyRepository
func (r *vacancyRepository) Update(ctx context.Context, vacancy *domain.Vacancy) error {
	return r.db.WithContext(ctx).Save(vacancy).Error
}

// Delete implements domain.VacancyRepository
func (r *vacancyRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Удаляем отклики
		if err := tx.Where("vacancy_id = ?", id).Delete(&domain.VacancyResponse{}).Error; err != nil {
			return err
		}
		// Удаляем вакансию
		return tx.Delete(&domain.Vacancy{}, "id = ?", id).Error
	})
}

// List implements domain.VacancyRepository
func (r *vacancyRepository) List(ctx context.Context, filter domain.VacancyFilter) ([]domain.Vacancy, int64, error) {
	var vacancies []domain.Vacancy
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.Vacancy{})

	// Применяем фильтры
	if filter.Company != nil && *filter.Company != "" {
		query = query.Where("LOWER(company) LIKE ?", "%"+strings.ToLower(*filter.Company)+"%")
	}

	if filter.Location != nil && *filter.Location != "" {
		query = query.Where("LOWER(location) LIKE ?", "%"+strings.ToLower(*filter.Location)+"%")
	}

	if filter.IsRemote != nil {
		query = query.Where("is_remote = ?", *filter.IsRemote)
	}

	if filter.EmploymentType != nil {
		query = query.Where("employment_type = ?", *filter.EmploymentType)
	}

	if filter.ExperienceLevel != nil {
		query = query.Where("experience_level = ?", *filter.ExperienceLevel)
	}

	if filter.SalaryMin != nil {
		query = query.Where("salary_max >= ? OR salary_max IS NULL", *filter.SalaryMin)
	}

	if filter.SalaryMax != nil {
		query = query.Where("salary_min <= ? OR salary_min IS NULL", *filter.SalaryMax)
	}

	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	} else {
		// По умолчанию показываем только активные
		query = query.Where("status = ?", domain.VacancyStatusActive)
	}

	if filter.Search != "" {
		searchTerm := "%" + strings.ToLower(filter.Search) + "%"
		query = query.Where(
			"LOWER(title) LIKE ? OR LOWER(company) LIKE ? OR LOWER(description) LIKE ? OR LOWER(requirements) LIKE ?",
			searchTerm, searchTerm, searchTerm, searchTerm,
		)
	}

	// Подсчитываем общее количество
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Применяем пагинацию и сортировку
	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}
	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}

	err := query.
		Preload("Employer").
		Order("created_at DESC").
		Find(&vacancies).Error

	return vacancies, total, err
}

// IncrementViews implements domain.VacancyRepository
func (r *vacancyRepository) IncrementViews(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&domain.Vacancy{}).
		Where("id = ?", id).
		Update("views_count", gorm.Expr("views_count + 1")).Error
}

// Search implements domain.VacancyRepository (полнотекстовый поиск)
func (r *vacancyRepository) Search(ctx context.Context, query string, limit, offset int) ([]domain.Vacancy, int64, error) {
	var vacancies []domain.Vacancy
	var total int64

	// Используем полнотекстовый поиск PostgreSQL
	searchQuery := strings.Join(strings.Fields(query), " & ")

	dbQuery := r.db.WithContext(ctx).
		Where("to_tsvector('russian', title || ' ' || COALESCE(description, '') || ' ' || COALESCE(requirements, '')) @@ to_tsquery('russian', ?)", searchQuery).
		Where("status = ?", domain.VacancyStatusActive)

	if err := dbQuery.Model(&domain.Vacancy{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Строим строку сортировки
	orderClause := "ts_rank(to_tsvector('russian', title || ' ' || COALESCE(description, '') || ' ' || COALESCE(requirements, '')), to_tsquery('russian', '" + searchQuery + "')) DESC"

	err := dbQuery.
		Preload("Employer").
		Order(orderClause).
		Limit(limit).
		Offset(offset).
		Find(&vacancies).Error

	return vacancies, total, err
}

// CreateResponse implements domain.VacancyRepository
func (r *vacancyRepository) CreateResponse(ctx context.Context, response *domain.VacancyResponse) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Проверяем, не откликался ли пользователь уже
		var existingResponse domain.VacancyResponse
		err := tx.Where("vacancy_id = ? AND user_id = ?", response.VacancyID, response.UserID).
			First(&existingResponse).Error

		if err == nil {
			return domain.ErrResponseExists
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		// Проверяем, активна ли вакансия
		var vacancy domain.Vacancy
		if err := tx.Select("status").Where("id = ?", response.VacancyID).First(&vacancy).Error; err != nil {
			return err
		}
		if vacancy.Status != domain.VacancyStatusActive {
			return domain.ErrVacancyClosed
		}

		return tx.Create(response).Error
	})
}

// UpdateResponse implements domain.VacancyRepository
func (r *vacancyRepository) UpdateResponse(ctx context.Context, response *domain.VacancyResponse) error {
	return r.db.WithContext(ctx).Save(response).Error
}

// GetResponse implements domain.VacancyRepository
func (r *vacancyRepository) GetResponse(ctx context.Context, id uuid.UUID) (*domain.VacancyResponse, error) {
	var response domain.VacancyResponse
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Vacancy").
		Where("id = ?", id).
		First(&response).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrResponseNotFound
		}
		return nil, err
	}
	return &response, nil
}

// GetResponseByVacancyAndUser implements domain.VacancyRepository
func (r *vacancyRepository) GetResponseByVacancyAndUser(ctx context.Context, vacancyID, userID uuid.UUID) (*domain.VacancyResponse, error) {
	var response domain.VacancyResponse
	err := r.db.WithContext(ctx).
		Where("vacancy_id = ? AND user_id = ?", vacancyID, userID).
		First(&response).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrResponseNotFound
		}
		return nil, err
	}
	return &response, nil
}

// GetResponsesByVacancy implements domain.VacancyRepository
func (r *vacancyRepository) GetResponsesByVacancy(ctx context.Context, vacancyID uuid.UUID) ([]domain.VacancyResponse, error) {
	var responses []domain.VacancyResponse
	err := r.db.WithContext(ctx).
		Preload("User").
		Where("vacancy_id = ?", vacancyID).
		Order("created_at DESC").
		Find(&responses).Error
	return responses, err
}

// GetResponsesByUser implements domain.VacancyRepository
func (r *vacancyRepository) GetResponsesByUser(ctx context.Context, userID uuid.UUID) ([]domain.VacancyResponse, error) {
	var responses []domain.VacancyResponse
	err := r.db.WithContext(ctx).
		Preload("Vacancy").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&responses).Error
	return responses, err
}

// GetUserResponses implements domain.VacancyRepository
func (r *vacancyRepository) GetUserResponses(ctx context.Context, userID uuid.UUID) ([]domain.VacancyResponse, error) {
	return r.GetResponsesByUser(ctx, userID)
}

// GetResponsesByEmployer implements domain.VacancyRepository
func (r *vacancyRepository) GetResponsesByEmployer(ctx context.Context, employerID uuid.UUID) ([]domain.VacancyResponse, error) {
	var responses []domain.VacancyResponse
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Vacancy").
		Joins("JOIN vacancies ON vacancies.id = vacancy_responses.vacancy_id").
		Where("vacancies.employer_id = ?", employerID).
		Order("vacancy_responses.created_at DESC").
		Find(&responses).Error
	return responses, err
}

// CountResponsesByVacancy implements domain.VacancyRepository
func (r *vacancyRepository) CountResponsesByVacancy(ctx context.Context, vacancyID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.VacancyResponse{}).
		Where("vacancy_id = ?", vacancyID).
		Count(&count).Error
	return count, err
}