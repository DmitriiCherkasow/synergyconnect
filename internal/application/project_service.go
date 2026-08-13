package application

import (
	"context"
	"errors"
	"strings"

	"github.com/DmitriiCherkasow/synergyconnect.git/internal/domain"
	"github.com/google/uuid"
)

// ProjectService — сервис для работы с проектами
type ProjectService struct {
	projectRepo     domain.ProjectRepository
	//notificationSvc *NotificationService // добавим позже
}

// NewProjectService создает новый сервис проектов
func NewProjectService(projectRepo domain.ProjectRepository) *ProjectService {
	return &ProjectService{
		projectRepo: projectRepo,
	}
}

// CreateProjectRequest — данные для создания проекта
type CreateProjectRequest struct {
	Title       string
	Description string
	MaxTeamSize int
	IsPublic    bool
	Tags        []string
}

// CreateProject создает новый проект
func (s *ProjectService) CreateProject(ctx context.Context, ownerID uuid.UUID, req CreateProjectRequest) (*domain.Project, error) {
	if strings.TrimSpace(req.Title) == "" {
		return nil, errors.New("title is required")
	}
	if len(req.Title) > 255 {
		return nil, errors.New("title must be less than 255 characters")
	}
	if req.MaxTeamSize < 1 {
		req.MaxTeamSize = 5
	}
	if req.MaxTeamSize > 100 {
		return nil, errors.New("max team size cannot exceed 100")
	}

	project := &domain.Project{
		ID:          uuid.New(),
		Title:       req.Title,
		Description: req.Description,
		Status:      domain.ProjectStatusOpen,
		OwnerID:     ownerID,
		MaxTeamSize: req.MaxTeamSize,
		IsPublic:    req.IsPublic,
		Tags:        req.Tags,
	}

	if err := s.projectRepo.Create(ctx, project); err != nil {
		return nil, err
	}

	// Добавляем владельца как участника
	member := &domain.ProjectMember{
		ProjectID: project.ID,
		UserID:    ownerID,
		Role:      domain.ProjectMemberRoleOwner,
	}
	if err := s.projectRepo.AddMember(ctx, member); err != nil {
		return nil, err
	}

	return project, nil
}

// GetProjectByID возвращает проект по ID
func (s *ProjectService) GetProjectByID(ctx context.Context, id uuid.UUID) (*domain.Project, error) {
	return s.projectRepo.GetByID(ctx, id)
}

// UpdateProjectRequest — данные для обновления проекта
type UpdateProjectRequest struct {
	Title       *string
	Description *string
	Status      *domain.ProjectStatus
	MaxTeamSize *int
	IsPublic    *bool
	Tags        []string
}

// UpdateProject обновляет проект
func (s *ProjectService) UpdateProject(ctx context.Context, projectID uuid.UUID, userID uuid.UUID, req UpdateProjectRequest) (*domain.Project, error) {
	project, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return nil, err
	}
	if project == nil {
		return nil, domain.ErrProjectNotFound
	}

	// Проверяем права (только владелец или админ может редактировать)
	if !project.CanManage(userID) {
		return nil, domain.ErrForbidden
	}

	if req.Title != nil {
		if *req.Title == "" {
			return nil, errors.New("title cannot be empty")
		}
		project.Title = *req.Title
	}
	if req.Description != nil {
		project.Description = *req.Description
	}
	if req.Status != nil {
		project.Status = *req.Status
	}
	if req.MaxTeamSize != nil {
		if *req.MaxTeamSize < 1 {
			return nil, errors.New("max team size must be at least 1")
		}
		if *req.MaxTeamSize > 100 {
			return nil, errors.New("max team size cannot exceed 100")
		}
		project.MaxTeamSize = *req.MaxTeamSize
	}
	if req.IsPublic != nil {
		project.IsPublic = *req.IsPublic
	}
	if req.Tags != nil {
		project.Tags = req.Tags
	}

	if err := s.projectRepo.Update(ctx, project); err != nil {
		return nil, err
	}

	return project, nil
}

// DeleteProject удаляет проект
func (s *ProjectService) DeleteProject(ctx context.Context, projectID uuid.UUID, userID uuid.UUID) error {
	project, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return err
	}
	if project == nil {
		return domain.ErrProjectNotFound
	}

	// Только владелец может удалить проект
	if !project.IsOwner(userID) {
		return domain.ErrForbidden
	}

	return s.projectRepo.Delete(ctx, projectID)
}

// ListProjectsRequest — фильтры для списка проектов
type ListProjectsRequest struct {
	Status   *domain.ProjectStatus
	OwnerID  *uuid.UUID
	MemberID *uuid.UUID
	Tag      string
	Search   string
	Limit    int
	Offset   int
}

// ListProjects возвращает список проектов с фильтрацией
func (s *ProjectService) ListProjects(ctx context.Context, req ListProjectsRequest) ([]domain.Project, int64, error) {
	if req.Limit <= 0 {
		req.Limit = 20
	}
	if req.Limit > 100 {
		req.Limit = 100
	}

	filter := domain.ProjectFilter{
		Status:   req.Status,
		OwnerID:  req.OwnerID,
		MemberID: req.MemberID,
		Tag:      req.Tag,
		Search:   req.Search,
		Limit:    req.Limit,
		Offset:   req.Offset,
	}

	return s.projectRepo.List(ctx, filter)
}

// AddMember добавляет участника в проект
func (s *ProjectService) AddMember(ctx context.Context, projectID uuid.UUID, userID uuid.UUID, actorID uuid.UUID, role domain.ProjectMemberRole) error {
	project, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return err
	}
	if project == nil {
		return domain.ErrProjectNotFound
	}

	// Проверяем права (владелец или админ может добавлять)
	if !project.CanManage(actorID) {
		return domain.ErrForbidden
	}

	member := &domain.ProjectMember{
		ProjectID: projectID,
		UserID:    userID,
		Role:      role,
	}

	return s.projectRepo.AddMember(ctx, member)
}

// RemoveMember удаляет участника из проекта
func (s *ProjectService) RemoveMember(ctx context.Context, projectID uuid.UUID, userID uuid.UUID, actorID uuid.UUID) error {
	project, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return err
	}
	if project == nil {
		return domain.ErrProjectNotFound
	}

	// Проверяем права
	if !project.CanManage(actorID) && actorID != userID {
		return domain.ErrForbidden
	}

	return s.projectRepo.RemoveMember(ctx, projectID, userID)
}

// GetMembers возвращает список участников проекта
func (s *ProjectService) GetMembers(ctx context.Context, projectID uuid.UUID) ([]domain.ProjectMember, error) {
	_, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return nil, err
	}

	return s.projectRepo.GetMembers(ctx, projectID)
}

// CreateApplication создает заявку на вступление в проект
func (s *ProjectService) CreateApplication(ctx context.Context, projectID uuid.UUID, userID uuid.UUID, message string) error {
	project, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return err
	}
	if project == nil {
		return domain.ErrProjectNotFound
	}

	// Проверяем, не закрыт ли проект
	if project.Status == domain.ProjectStatusCompleted || project.Status == domain.ProjectStatusArchived {
		return errors.New("project is closed for applications")
	}

	application := &domain.ProjectApplication{
		ID:        uuid.New(),
		ProjectID: projectID,
		UserID:    userID,
		Message:   message,
		Status:    domain.ProjectApplicationStatusPending,
	}

	return s.projectRepo.CreateApplication(ctx, application)
}

// UpdateApplicationStatus изменяет статус заявки
func (s *ProjectService) UpdateApplicationStatus(ctx context.Context, applicationID uuid.UUID, actorID uuid.UUID, status domain.ProjectApplicationStatus) error {
	application, err := s.projectRepo.GetApplication(ctx, applicationID)
	if err != nil {
		return err
	}
	if application == nil {
		return domain.ErrApplicationNotFound
	}

	project, err := s.projectRepo.GetByID(ctx, application.ProjectID)
	if err != nil {
		return err
	}
	if project == nil {
		return domain.ErrProjectNotFound
	}

	// Проверяем права (владелец или админ может менять статус)
	if !project.CanManage(actorID) {
		return domain.ErrForbidden
	}

	application.Status = status
	return s.projectRepo.UpdateApplication(ctx, application)
}

// GetUserApplications возвращает заявки пользователя
func (s *ProjectService) GetUserApplications(ctx context.Context, userID uuid.UUID) ([]domain.ProjectApplication, error) {
	return s.projectRepo.GetUserApplications(ctx, userID)
}