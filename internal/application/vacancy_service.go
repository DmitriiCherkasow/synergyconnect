package application

import (
	"context"
	"errors"
	"strings"

	"github.com/DmitriiCherkasow/synergyconnect.git/internal/domain"
	"github.com/google/uuid"
)

// VacancyService — сервис для работы с вакансиями
type VacancyService struct {
	vacancyRepo domain.VacancyRepository
}

// NewVacancyService создает новый сервис вакансий
func NewVacancyService(vacancyRepo domain.VacancyRepository) *VacancyService {
	return &VacancyService{
		vacancyRepo: vacancyRepo,
	}
}

// CreateVacancyRequest — данные для создания вакансии
type CreateVacancyRequest struct {
	Title           string
	Company         string
	Description     string
	Requirements    string
	SalaryMin       *int
	SalaryMax       *int
	Currency        string
	Location        string
	IsRemote        bool
	EmploymentType  domain.EmploymentType
	ExperienceLevel domain.ExperienceLevel
}

// CreateVacancy создает новую вакансию
func (s *VacancyService) CreateVacancy(ctx context.Context, employerID uuid.UUID, req CreateVacancyRequest) (*domain.Vacancy, error) {
	// Валидация
	if strings.TrimSpace(req.Title) == "" {
		return nil, errors.New("title is required")
	}
	if len(req.Title) > 255 {
		return nil, errors.New("title must be less than 255 characters")
	}
	if strings.TrimSpace(req.Company) == "" {
		return nil, errors.New("company is required")
	}
	if len(req.Company) > 255 {
		return nil, errors.New("company must be less than 255 characters")
	}

	vacancy := &domain.Vacancy{
		ID:              uuid.New(),
		Title:           req.Title,
		Company:         req.Company,
		Description:     req.Description,
		Requirements:    req.Requirements,
		SalaryMin:       req.SalaryMin,
		SalaryMax:       req.SalaryMax,
		Currency:        req.Currency,
		Location:        req.Location,
		IsRemote:        req.IsRemote,
		EmploymentType:  req.EmploymentType,
		ExperienceLevel: req.ExperienceLevel,
		EmployerID:      employerID,
		Status:          domain.VacancyStatusActive,
		ViewsCount:      0,
	}

	if err := s.vacancyRepo.Create(ctx, vacancy); err != nil {
		return nil, err
	}

	return vacancy, nil
}

// GetVacancyByID возвращает вакансию по ID
func (s *VacancyService) GetVacancyByID(ctx context.Context, id uuid.UUID) (*domain.Vacancy, error) {
	return s.vacancyRepo.GetByID(ctx, id)
}

// UpdateVacancyRequest — данные для обновления вакансии
type UpdateVacancyRequest struct {
	Title           *string
	Company         *string
	Description     *string
	Requirements    *string
	SalaryMin       *int
	SalaryMax       *int
	Currency        *string
	Location        *string
	IsRemote        *bool
	EmploymentType  *domain.EmploymentType
	ExperienceLevel *domain.ExperienceLevel
	Status          *domain.VacancyStatus
}

// UpdateVacancy обновляет вакансию
func (s *VacancyService) UpdateVacancy(ctx context.Context, vacancyID, employerID uuid.UUID, req UpdateVacancyRequest) (*domain.Vacancy, error) {
	vacancy, err := s.vacancyRepo.GetByID(ctx, vacancyID)
	if err != nil {
		return nil, err
	}
	if vacancy == nil {
		return nil, domain.ErrVacancyNotFound
	}

	// Проверяем права (только работодатель может редактировать)
	if vacancy.EmployerID != employerID {
		return nil, domain.ErrForbidden
	}

	if req.Title != nil {
		if strings.TrimSpace(*req.Title) == "" {
			return nil, errors.New("title cannot be empty")
		}
		vacancy.Title = *req.Title
	}
	if req.Company != nil {
		if strings.TrimSpace(*req.Company) == "" {
			return nil, errors.New("company cannot be empty")
		}
		vacancy.Company = *req.Company
	}
	if req.Description != nil {
		vacancy.Description = *req.Description
	}
	if req.Requirements != nil {
		vacancy.Requirements = *req.Requirements
	}
	if req.SalaryMin != nil {
		vacancy.SalaryMin = req.SalaryMin
	}
	if req.SalaryMax != nil {
		vacancy.SalaryMax = req.SalaryMax
	}
	if req.Currency != nil {
		vacancy.Currency = *req.Currency
	}
	if req.Location != nil {
		vacancy.Location = *req.Location
	}
	if req.IsRemote != nil {
		vacancy.IsRemote = *req.IsRemote
	}
	if req.EmploymentType != nil {
		vacancy.EmploymentType = *req.EmploymentType
	}
	if req.ExperienceLevel != nil {
		vacancy.ExperienceLevel = *req.ExperienceLevel
	}
	if req.Status != nil {
		vacancy.Status = *req.Status
	}

	if err := s.vacancyRepo.Update(ctx, vacancy); err != nil {
		return nil, err
	}

	return vacancy, nil
}

// DeleteVacancy удаляет вакансию
func (s *VacancyService) DeleteVacancy(ctx context.Context, vacancyID, employerID uuid.UUID) error {
	vacancy, err := s.vacancyRepo.GetByID(ctx, vacancyID)
	if err != nil {
		return err
	}
	if vacancy == nil {
		return domain.ErrVacancyNotFound
	}

	// Только работодатель может удалить вакансию
	if vacancy.EmployerID != employerID {
		return domain.ErrForbidden
	}

	return s.vacancyRepo.Delete(ctx, vacancyID)
}

// ListVacanciesRequest — фильтры для списка вакансий
type ListVacanciesRequest struct {
	Company         *string
	Location        *string
	IsRemote        *bool
	EmploymentType  *domain.EmploymentType
	ExperienceLevel *domain.ExperienceLevel
	SalaryMin       *int
	SalaryMax       *int
	Status          *domain.VacancyStatus
	Search          string
	Limit           int
	Offset          int
}

// ListVacancies возвращает список вакансий с фильтрацией
func (s *VacancyService) ListVacancies(ctx context.Context, req ListVacanciesRequest) ([]domain.Vacancy, int64, error) {
	if req.Limit <= 0 {
		req.Limit = 20
	}
	if req.Limit > 100 {
		req.Limit = 100
	}

	filter := domain.VacancyFilter{
		Company:         req.Company,
		Location:        req.Location,
		IsRemote:        req.IsRemote,
		EmploymentType:  req.EmploymentType,
		ExperienceLevel: req.ExperienceLevel,
		SalaryMin:       req.SalaryMin,
		SalaryMax:       req.SalaryMax,
		Status:          req.Status,
		Search:          req.Search,
		Limit:           req.Limit,
		Offset:          req.Offset,
	}

	return s.vacancyRepo.List(ctx, filter)
}

// SearchVacancies выполняет полнотекстовый поиск вакансий
func (s *VacancyService) SearchVacancies(ctx context.Context, query string, limit, offset int) ([]domain.Vacancy, int64, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return s.vacancyRepo.Search(ctx, query, limit, offset)
}

// IncrementViews увеличивает счетчик просмотров
func (s *VacancyService) IncrementViews(ctx context.Context, vacancyID uuid.UUID) error {
	return s.vacancyRepo.IncrementViews(ctx, vacancyID)
}

// CreateResponseRequest — данные для отклика на вакансию
type CreateResponseRequest struct {
	CoverLetter string
}

// CreateResponse создает отклик на вакансию
func (s *VacancyService) CreateResponse(ctx context.Context, vacancyID, userID uuid.UUID, req CreateResponseRequest) error {
	response := &domain.VacancyResponse{
		ID:          uuid.New(),
		VacancyID:   vacancyID,
		UserID:      userID,
		CoverLetter: req.CoverLetter,
		Status:      domain.VacancyResponseStatusPending,
	}

	return s.vacancyRepo.CreateResponse(ctx, response)
}

// UpdateResponseStatus изменяет статус отклика
func (s *VacancyService) UpdateResponseStatus(ctx context.Context, responseID uuid.UUID, employerID uuid.UUID, status domain.VacancyResponseStatus) error {
	response, err := s.vacancyRepo.GetResponse(ctx, responseID)
	if err != nil {
		return err
	}
	if response == nil {
		return domain.ErrResponseNotFound
	}

	// Проверяем, что пользователь является работодателем
	vacancy, err := s.vacancyRepo.GetByID(ctx, response.VacancyID)
	if err != nil {
		return err
	}
	if vacancy == nil {
		return domain.ErrVacancyNotFound
	}
	if vacancy.EmployerID != employerID {
		return domain.ErrForbidden
	}

	response.Status = status
	return s.vacancyRepo.UpdateResponse(ctx, response)
}

// GetResponsesByVacancy возвращает отклики на вакансию
func (s *VacancyService) GetResponsesByVacancy(ctx context.Context, vacancyID, employerID uuid.UUID) ([]domain.VacancyResponse, error) {
	vacancy, err := s.vacancyRepo.GetByID(ctx, vacancyID)
	if err != nil {
		return nil, err
	}
	if vacancy == nil {
		return nil, domain.ErrVacancyNotFound
	}
	if vacancy.EmployerID != employerID {
		return nil, domain.ErrForbidden
	}

	return s.vacancyRepo.GetResponsesByVacancy(ctx, vacancyID)
}

// GetUserResponses возвращает отклики пользователя
func (s *VacancyService) GetUserResponses(ctx context.Context, userID uuid.UUID) ([]domain.VacancyResponse, error) {
	return s.vacancyRepo.GetUserResponses(ctx, userID)
}