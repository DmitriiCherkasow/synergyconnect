package dto

import (
	"errors"
	"strings"
	"time"

	"github.com/DmitriiCherkasow/synergyconnect.git/internal/domain"
)

// CreateVacancyRequest — запрос на создание вакансии
type CreateVacancyRequest struct {
	Title           string `json:"title" binding:"required"`
	Company         string `json:"company" binding:"required"`
	Description     string `json:"description"`
	Requirements    string `json:"requirements"`
	SalaryMin       *int   `json:"salary_min,omitempty"`
	SalaryMax       *int   `json:"salary_max,omitempty"`
	Currency        string `json:"currency"`
	Location        string `json:"location"`
	IsRemote        bool   `json:"is_remote"`
	EmploymentType  string `json:"employment_type"`
	ExperienceLevel string `json:"experience_level"`
}

func (r *CreateVacancyRequest) Validate() error {
	if strings.TrimSpace(r.Title) == "" {
		return errors.New("title is required")
	}
	if len(r.Title) > 255 {
		return errors.New("title must be less than 255 characters")
	}
	if strings.TrimSpace(r.Company) == "" {
		return errors.New("company is required")
	}
	if r.SalaryMin != nil && r.SalaryMax != nil && *r.SalaryMin > *r.SalaryMax {
		return errors.New("salary_min cannot be greater than salary_max")
	}
	return nil
}

// UpdateVacancyRequest — запрос на обновление вакансии
type UpdateVacancyRequest struct {
	Title           *string `json:"title,omitempty"`
	Company         *string `json:"company,omitempty"`
	Description     *string `json:"description,omitempty"`
	Requirements    *string `json:"requirements,omitempty"`
	SalaryMin       *int    `json:"salary_min,omitempty"`
	SalaryMax       *int    `json:"salary_max,omitempty"`
	Currency        *string `json:"currency,omitempty"`
	Location        *string `json:"location,omitempty"`
	IsRemote        *bool   `json:"is_remote,omitempty"`
	EmploymentType  *string `json:"employment_type,omitempty"`
	ExperienceLevel *string `json:"experience_level,omitempty"`
	Status          *string `json:"status,omitempty"`
}

func (r *UpdateVacancyRequest) Validate() error {
	if r.Title != nil && strings.TrimSpace(*r.Title) == "" {
		return errors.New("title cannot be empty")
	}
	if r.Company != nil && strings.TrimSpace(*r.Company) == "" {
		return errors.New("company cannot be empty")
	}
	if r.SalaryMin != nil && r.SalaryMax != nil && *r.SalaryMin > *r.SalaryMax {
		return errors.New("salary_min cannot be greater than salary_max")
	}
	if r.Status != nil {
		switch domain.VacancyStatus(*r.Status) {
		case domain.VacancyStatusActive, domain.VacancyStatusClosed, domain.VacancyStatusArchived:
			// valid
		default:
			return errors.New("invalid status")
		}
	}
	return nil
}

// VacancyResponse — ответ с данными вакансии
type VacancyResponse struct {
	ID              string    `json:"id"`
	Title           string    `json:"title"`
	Company         string    `json:"company"`
	Description     string    `json:"description"`
	Requirements    string    `json:"requirements"`
	SalaryMin       *int      `json:"salary_min,omitempty"`
	SalaryMax       *int      `json:"salary_max,omitempty"`
	Currency        string    `json:"currency"`
	Location        string    `json:"location"`
	IsRemote        bool      `json:"is_remote"`
	EmploymentType  string    `json:"employment_type"`
	ExperienceLevel string    `json:"experience_level"`
	EmployerID      string    `json:"employer_id"`
	Employer        UserResponse `json:"employer,omitempty"`
	Status          string    `json:"status"`
	ViewsCount      int       `json:"views_count"`
	ResponsesCount  int       `json:"responses_count"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// VacancyDetailResponse — детальный ответ с откликами
type VacancyDetailResponse struct {
	VacancyResponse
	Responses []VacancyResponseItem `json:"responses,omitempty"`
	CanManage bool                  `json:"can_manage"`
	HasResponded bool               `json:"has_responded"`
}

// VacancyResponseItem — отклик на вакансию
type VacancyResponseItem struct {
	ID          string       `json:"id"`
	UserID      string       `json:"user_id"`
	User        UserResponse `json:"user"`
	VacancyID   string       `json:"vacancy_id"`
	VacancyTitle string      `json:"vacancy_title"`
	CoverLetter string       `json:"cover_letter"`
	Status      string       `json:"status"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

// VacancyListResponse — список вакансий
type VacancyListResponse struct {
	Vacancies []VacancyResponse `json:"vacancies"`
	Total     int64             `json:"total"`
}

// CreateResponseRequest — запрос на отклик
type CreateResponseRequest struct {
	CoverLetter string `json:"cover_letter"`
}

// UpdateResponseRequest — запрос на изменение статуса отклика
type UpdateResponseRequest struct {
	Status string `json:"status" binding:"required"`
}

func (r *UpdateResponseRequest) Validate() error {
	switch domain.VacancyResponseStatus(r.Status) {
	case domain.VacancyResponseStatusReviewed, domain.VacancyResponseStatusAccepted, domain.VacancyResponseStatusRejected:
		return nil
	default:
		return errors.New("invalid status, must be reviewed, accepted or rejected")
	}
}

// ToVacancyResponse конвертирует доменную модель в DTO
func ToVacancyResponse(vacancy *domain.Vacancy) VacancyResponse {
	return VacancyResponse{
		ID:              vacancy.ID.String(),
		Title:           vacancy.Title,
		Company:         vacancy.Company,
		Description:     vacancy.Description,
		Requirements:    vacancy.Requirements,
		SalaryMin:       vacancy.SalaryMin,
		SalaryMax:       vacancy.SalaryMax,
		Currency:        vacancy.Currency,
		Location:        vacancy.Location,
		IsRemote:        vacancy.IsRemote,
		EmploymentType:  string(vacancy.EmploymentType),
		ExperienceLevel: string(vacancy.ExperienceLevel),
		EmployerID:      vacancy.EmployerID.String(),
		Status:          string(vacancy.Status),
		ViewsCount:      vacancy.ViewsCount,
		ResponsesCount:  len(vacancy.Responses),
		CreatedAt:       vacancy.CreatedAt,
		UpdatedAt:       vacancy.UpdatedAt,
	}
}

// ToVacancyResponseItem конвертирует доменную модель в DTO
func ToVacancyResponseItem(response *domain.VacancyResponse) VacancyResponseItem {
	return VacancyResponseItem{
		ID:          response.ID.String(),
		UserID:      response.UserID.String(),
		CoverLetter: response.CoverLetter,
		Status:      string(response.Status),
		CreatedAt:   response.CreatedAt,
		UpdatedAt:   response.UpdatedAt,
	}
}