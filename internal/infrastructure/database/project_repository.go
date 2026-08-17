package database

import (
	"context"
	"errors"
	"strings"

	"github.com/DmitriiCherkasow/synergyconnect.git/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type projectRepository struct {
	db *gorm.DB
}

func NewProjectRepository(db *gorm.DB) domain.ProjectRepository {
	return &projectRepository{db: db}
}

// Create implements domain.ProjectRepository
func (r *projectRepository) Create(ctx context.Context, project *domain.Project) error {
	return r.db.WithContext(ctx).Create(project).Error
}

// GetByID implements domain.ProjectRepository
func (r *projectRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Project, error) {
	var project domain.Project
	err := r.db.WithContext(ctx).
		Preload("Owner").
		Preload("Members.User").
		Preload("Applications").
		Where("id = ?", id).
		First(&project).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrProjectNotFound
		}
		return nil, err
	}
	return &project, nil
}

// Update implements domain.ProjectRepository
func (r *projectRepository) Update(ctx context.Context, project *domain.Project) error {
	return r.db.WithContext(ctx).Save(project).Error
}

// Delete implements domain.ProjectRepository
func (r *projectRepository) Delete(ctx context.Context, id uuid.UUID) error {
	// Начинаем транзакцию для каскадного удаления
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Удаляем участников
		if err := tx.Where("project_id = ?", id).Delete(&domain.ProjectMember{}).Error; err != nil {
			return err
		}
		// Удаляем заявки
		if err := tx.Where("project_id = ?", id).Delete(&domain.ProjectApplication{}).Error; err != nil {
			return err
		}
		// Удаляем сам проект
		return tx.Delete(&domain.Project{}, "id = ?", id).Error
	})
}

// List implements domain.ProjectRepository
func (r *projectRepository) List(ctx context.Context, filter domain.ProjectFilter) ([]domain.Project, int64, error) {
	var projects []domain.Project
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.Project{})

	// Применяем фильтры
	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}

	if filter.OwnerID != nil {
		query = query.Where("owner_id = ?", *filter.OwnerID)
	}

	if filter.MemberID != nil {
		// Подзапрос для поиска проектов, в которых пользователь является участником
		subQuery := r.db.Table("project_members").
			Select("project_id").
			Where("user_id = ?", *filter.MemberID)
		query = query.Where("id IN (?)", subQuery)
	}

	if filter.Tag != "" {
		query = query.Where("? = ANY(tags)", filter.Tag)
	}

	if filter.Search != "" {
		searchTerm := "%" + strings.ToLower(filter.Search) + "%"
		query = query.Where("LOWER(title) LIKE ? OR LOWER(description) LIKE ?", searchTerm, searchTerm)
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

	// Загружаем связанные данные
	err := query.
		Preload("Owner").
		Preload("Members.User").
		Order("created_at DESC").
		Find(&projects).Error

	return projects, total, err
}

// AddMember implements domain.ProjectRepository
func (r *projectRepository) AddMember(ctx context.Context, member *domain.ProjectMember) error {
	// Проверяем, не превышен ли лимит участников
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&domain.ProjectMember{}).
		Where("project_id = ?", member.ProjectID).
		Count(&count).Error; err != nil {
		return err
	}

	var project domain.Project
	if err := r.db.WithContext(ctx).
		Select("max_team_size").
		Where("id = ?", member.ProjectID).
		First(&project).Error; err != nil {
		return err
	}

	if int(count) >= project.MaxTeamSize {
		return domain.ErrProjectFull
	}

	// Проверяем, не является ли пользователь уже участником
	var existingMember domain.ProjectMember
	err := r.db.WithContext(ctx).
		Where("project_id = ? AND user_id = ?", member.ProjectID, member.UserID).
		First(&existingMember).Error

	if err == nil {
		return domain.ErrAlreadyMember
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	return r.db.WithContext(ctx).Create(member).Error
}

// RemoveMember implements domain.ProjectRepository
func (r *projectRepository) RemoveMember(ctx context.Context, projectID, userID uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Проверяем, не является ли пользователь владельцем
		var project domain.Project
		if err := tx.Select("owner_id").Where("id = ?", projectID).First(&project).Error; err != nil {
			return err
		}

		if project.OwnerID == userID {
			return domain.ErrCannotRemoveOwner
		}

		// Удаляем участника
		result := tx.Where("project_id = ? AND user_id = ?", projectID, userID).
			Delete(&domain.ProjectMember{})

		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return domain.ErrNotMember
		}

		return nil
	})
}

// GetMembers implements domain.ProjectRepository
func (r *projectRepository) GetMembers(ctx context.Context, projectID uuid.UUID) ([]domain.ProjectMember, error) {
	var members []domain.ProjectMember
	err := r.db.WithContext(ctx).
		Preload("User").
		Where("project_id = ?", projectID).
		Order("joined_at ASC").
		Find(&members).Error
	return members, err
}

// GetMemberRole implements domain.ProjectRepository
func (r *projectRepository) GetMemberRole(ctx context.Context, projectID, userID uuid.UUID) (*domain.ProjectMemberRole, error) {
	var member domain.ProjectMember
	err := r.db.WithContext(ctx).
		Select("role").
		Where("project_id = ? AND user_id = ?", projectID, userID).
		First(&member).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotMember
		}
		return nil, err
	}
	return &member.Role, nil
}

// IsMember implements domain.ProjectRepository
func (r *projectRepository) IsMember(ctx context.Context, projectID, userID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.ProjectMember{}).
		Where("project_id = ? AND user_id = ?", projectID, userID).
		Count(&count).Error
	return count > 0, err
}

// CreateApplication implements domain.ProjectRepository
func (r *projectRepository) CreateApplication(ctx context.Context, application *domain.ProjectApplication) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Проверяем, не подавал ли пользователь уже заявку
		var existingApp domain.ProjectApplication
		err := tx.Where("project_id = ? AND user_id = ?", application.ProjectID, application.UserID).
			First(&existingApp).Error

		if err == nil {
			return domain.ErrApplicationExists
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		// Проверяем, не является ли пользователь уже участником
		var count int64
		if err := tx.Model(&domain.ProjectMember{}).
			Where("project_id = ? AND user_id = ?", application.ProjectID, application.UserID).
			Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return domain.ErrAlreadyMember
		}

		return tx.Create(application).Error
	})
}

// UpdateApplication implements domain.ProjectRepository
func (r *projectRepository) UpdateApplication(ctx context.Context, application *domain.ProjectApplication) error {
	return r.db.WithContext(ctx).Save(application).Error
}

// GetApplication implements domain.ProjectRepository
func (r *projectRepository) GetApplication(ctx context.Context, id uuid.UUID) (*domain.ProjectApplication, error) {
	var application domain.ProjectApplication
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Project").
		Where("id = ?", id).
		First(&application).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrApplicationNotFound
		}
		return nil, err
	}
	return &application, nil
}

// GetApplicationsByProject implements domain.ProjectRepository
func (r *projectRepository) GetApplicationsByProject(ctx context.Context, projectID uuid.UUID) ([]domain.ProjectApplication, error) {
	var applications []domain.ProjectApplication
	err := r.db.WithContext(ctx).
		Preload("User").
		Where("project_id = ?", projectID).
		Order("created_at DESC").
		Find(&applications).Error
	return applications, err
}

// GetApplicationsByUser implements domain.ProjectRepository
func (r *projectRepository) GetApplicationsByUser(ctx context.Context, userID uuid.UUID) ([]domain.ProjectApplication, error) {
	var applications []domain.ProjectApplication
	err := r.db.WithContext(ctx).
		Preload("Project").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&applications).Error
	return applications, err
}

// GetUserApplications implements domain.ProjectRepository
func (r *projectRepository) GetUserApplications(ctx context.Context, userID uuid.UUID) ([]domain.ProjectApplication, error) {
	return r.GetApplicationsByUser(ctx, userID)
}

// GetPendingApplications implements domain.ProjectRepository
func (r *projectRepository) GetPendingApplications(ctx context.Context, projectID uuid.UUID) ([]domain.ProjectApplication, error) {
	var applications []domain.ProjectApplication
	err := r.db.WithContext(ctx).
		Preload("User").
		Where("project_id = ? AND status = ?", projectID, domain.ProjectApplicationStatusPending).
		Order("created_at ASC").
		Find(&applications).Error
	return applications, err
}

// Count implements domain.ProjectRepository
func (r *projectRepository) Count(ctx context.Context) (int64, error) {
    var count int64
    err := r.db.WithContext(ctx).Model(&domain.Project{}).Count(&count).Error
    return count, err
}