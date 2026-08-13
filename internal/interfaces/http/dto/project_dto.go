package dto

import (
	"errors"
	"strings"
	"time"

	"github.com/DmitriiCherkasow/synergyconnect.git/internal/domain"
	"github.com/google/uuid"
)

// CreateProjectRequest — запрос на создание проекта
type CreateProjectRequest struct {
	Title       string   `json:"title" binding:"required"`
	Description string   `json:"description"`
	MaxTeamSize int      `json:"max_team_size"`
	IsPublic    bool     `json:"is_public"`
	Tags        []string `json:"tags"`
}

func (r *CreateProjectRequest) Validate() error {
	if strings.TrimSpace(r.Title) == "" {
		return errors.New("title is required")
	}
	if len(r.Title) > 255 {
		return errors.New("title must be less than 255 characters")
	}
	if r.MaxTeamSize < 0 {
		return errors.New("max_team_size must be positive")
	}
	if r.MaxTeamSize > 100 {
		return errors.New("max_team_size cannot exceed 100")
	}
	return nil
}

// UpdateProjectRequest — запрос на обновление проекта
type UpdateProjectRequest struct {
	Title       *string   `json:"title"`
	Description *string   `json:"description"`
	Status      *string   `json:"status"`
	MaxTeamSize *int      `json:"max_team_size"`
	IsPublic    *bool     `json:"is_public"`
	Tags        []string  `json:"tags"`
}

func (r *UpdateProjectRequest) Validate() error {
	if r.Title != nil && strings.TrimSpace(*r.Title) == "" {
		return errors.New("title cannot be empty")
	}
	if r.Title != nil && len(*r.Title) > 255 {
		return errors.New("title must be less than 255 characters")
	}
	if r.MaxTeamSize != nil && *r.MaxTeamSize < 1 {
		return errors.New("max_team_size must be at least 1")
	}
	if r.MaxTeamSize != nil && *r.MaxTeamSize > 100 {
		return errors.New("max_team_size cannot exceed 100")
	}
	if r.Status != nil {
		switch domain.ProjectStatus(*r.Status) {
		case domain.ProjectStatusOpen, domain.ProjectStatusInProgress, domain.ProjectStatusCompleted, domain.ProjectStatusArchived:
			// valid
		default:
			return errors.New("invalid status")
		}
	}
	return nil
}

// ProjectResponse — ответ с данными проекта
type ProjectResponse struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	OwnerID     string    `json:"owner_id"`
	Owner       UserResponse `json:"owner,omitempty"`
	MaxTeamSize int       `json:"max_team_size"`
	IsPublic    bool      `json:"is_public"`
	Tags        []string  `json:"tags"`
	MemberCount int       `json:"member_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ProjectDetailResponse — детальный ответ с участниками и заявками
type ProjectDetailResponse struct {
	ProjectResponse
	Members      []ProjectMemberResponse      `json:"members"`
	Applications []ProjectApplicationResponse `json:"applications,omitempty"`
	IsMember     bool                         `json:"is_member"`
	IsOwner      bool                         `json:"is_owner"`
	CanManage    bool                         `json:"can_manage"`
}

// ProjectMemberResponse — ответ с данными участника
type ProjectMemberResponse struct {
	UserID   string    `json:"user_id"`
	User     UserResponse `json:"user"`
	Role     string    `json:"role"`
	JoinedAt time.Time `json:"joined_at"`
}

// ProjectApplicationResponse — ответ с данными заявки
type ProjectApplicationResponse struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	User      UserResponse `json:"user"`
	Message   string    `json:"message"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ProjectListResponse — ответ со списком проектов
type ProjectListResponse struct {
	Projects []ProjectResponse `json:"projects"`
	Total    int64             `json:"total"`
}

// AddMemberRequest — запрос на добавление участника
type AddMemberRequest struct {
	UserID string `json:"user_id" binding:"required"`
	Role   string `json:"role"`
}

func (r *AddMemberRequest) Validate() error {
	if r.UserID == "" {
		return errors.New("user_id is required")
	}
	if _, err := uuid.Parse(r.UserID); err != nil {
		return errors.New("invalid user_id format")
	}
	if r.Role != "" {
		switch domain.ProjectMemberRole(r.Role) {
		case domain.ProjectMemberRoleMember, domain.ProjectMemberRoleAdmin:
			// valid
		default:
			return errors.New("invalid role")
		}
	}
	return nil
}

// CreateApplicationRequest — запрос на создание заявки
type CreateApplicationRequest struct {
	Message string `json:"message"`
}

// UpdateApplicationRequest — запрос на изменение статуса заявки
type UpdateApplicationRequest struct {
	Status string `json:"status" binding:"required"`
}

func (r *UpdateApplicationRequest) Validate() error {
	switch domain.ProjectApplicationStatus(r.Status) {
	case domain.ProjectApplicationStatusAccepted, domain.ProjectApplicationStatusRejected:
		return nil
	default:
		return errors.New("invalid status, must be accepted or rejected")
	}
}

// ToProjectResponse конвертирует доменную модель в DTO
func ToProjectResponse(project *domain.Project) ProjectResponse {
	return ProjectResponse{
		ID:          project.ID.String(),
		Title:       project.Title,
		Description: project.Description,
		Status:      string(project.Status),
		OwnerID:     project.OwnerID.String(),
		MaxTeamSize: project.MaxTeamSize,
		IsPublic:    project.IsPublic,
		Tags:        project.Tags,
		MemberCount: len(project.Members),
		CreatedAt:   project.CreatedAt,
		UpdatedAt:   project.UpdatedAt,
	}
}

// ToProjectMemberResponse конвертирует доменную модель в DTO
func ToProjectMemberResponse(member *domain.ProjectMember) ProjectMemberResponse {
	return ProjectMemberResponse{
		UserID:   member.UserID.String(),
		Role:     string(member.Role),
		JoinedAt: member.JoinedAt,
	}
}

// ToProjectApplicationResponse конвертирует доменную модель в DTO
func ToProjectApplicationResponse(app *domain.ProjectApplication) ProjectApplicationResponse {
	return ProjectApplicationResponse{
		ID:        app.ID.String(),
		UserID:    app.UserID.String(),
		Message:   app.Message,
		Status:    string(app.Status),
		CreatedAt: app.CreatedAt,
		UpdatedAt: app.UpdatedAt,
	}
}