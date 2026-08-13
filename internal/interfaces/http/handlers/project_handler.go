package handlers

import (
	"net/http"
	"strconv"

	"github.com/DmitriiCherkasow/synergyconnect.git/internal/application"
	"github.com/DmitriiCherkasow/synergyconnect.git/internal/domain"
	"github.com/DmitriiCherkasow/synergyconnect.git/internal/interfaces/http/dto"
	"github.com/DmitriiCherkasow/synergyconnect.git/internal/interfaces/http/middleware"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ProjectHandler — обработчик для проектов
type ProjectHandler struct {
	projectService *application.ProjectService
}

// NewProjectHandler создает новый обработчик
func NewProjectHandler(projectService *application.ProjectService) *ProjectHandler {
	return &ProjectHandler{
		projectService: projectService,
	}
}

// getUserID извлекает ID пользователя из контекста
func (h *ProjectHandler) getUserID(c *gin.Context) (uuid.UUID, error) {
	userIDStr := middleware.GetUserIDFromContext(c)
	if userIDStr == "" {
		return uuid.Nil, nil
	}
	return uuid.Parse(userIDStr)
}

// CreateProject — создание проекта
func (h *ProjectHandler) CreateProject(c *gin.Context) {
	userID, err := h.getUserID(c)
	if err != nil || userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req dto.CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := req.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	project, err := h.projectService.CreateProject(c.Request.Context(), userID, application.CreateProjectRequest{
		Title:       req.Title,
		Description: req.Description,
		MaxTeamSize: req.MaxTeamSize,
		IsPublic:    req.IsPublic,
		Tags:        req.Tags,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, dto.ToProjectResponse(project))
}

// GetProject — получение проекта по ID
func (h *ProjectHandler) GetProject(c *gin.Context) {
	userID, err := h.getUserID(c)
	if err != nil || userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project id"})
		return
	}

	project, err := h.projectService.GetProjectByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if project == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}

	// Проверяем, может ли пользователь видеть проект
	if !project.IsPublic {
		isMember := project.IsMember(userID)
		isOwner := project.IsOwner(userID)
		if !isMember && !isOwner {
			c.JSON(http.StatusForbidden, gin.H{"error": "you don't have access to this project"})
			return
		}
	}

	// Собираем ответ
	response := dto.ProjectDetailResponse{
		ProjectResponse: dto.ToProjectResponse(project),
		IsMember:        project.IsMember(userID),
		IsOwner:         project.IsOwner(userID),
		CanManage:       project.CanManage(userID),
	}

	// Добавляем участников
	members, err := h.projectService.GetMembers(c.Request.Context(), id)
	if err == nil {
		response.Members = make([]dto.ProjectMemberResponse, len(members))
		for i, member := range members {
			resp := dto.ToProjectMemberResponse(&member)
			// Добавляем данные пользователя
			if member.User.ID != uuid.Nil {
				resp.User = dto.UserResponse{
					ID:         member.User.ID.String(),
					Email:      member.User.Email,
					Role:       string(member.User.Role),
					FirstName:  member.User.FirstName,
					LastName:   member.User.LastName,
					AvatarURL:  member.User.AvatarURL,
					IsVerified: member.User.IsVerified,
				}
			}
			response.Members[i] = resp
		}
	}

	c.JSON(http.StatusOK, response)
}

// UpdateProject — обновление проекта
func (h *ProjectHandler) UpdateProject(c *gin.Context) {
	userID, err := h.getUserID(c)
	if err != nil || userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project id"})
		return
	}

	var req dto.UpdateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := req.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	project, err := h.projectService.UpdateProject(c.Request.Context(), id, userID, application.UpdateProjectRequest{
		Title:       req.Title,
		Description: req.Description,
		Status:      (*domain.ProjectStatus)(req.Status),
		MaxTeamSize: req.MaxTeamSize,
		IsPublic:    req.IsPublic,
		Tags:        req.Tags,
	})
	if err != nil {
		if err == domain.ErrProjectNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if err == domain.ErrForbidden {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.ToProjectResponse(project))
}

// DeleteProject — удаление проекта
func (h *ProjectHandler) DeleteProject(c *gin.Context) {
	userID, err := h.getUserID(c)
	if err != nil || userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project id"})
		return
	}

	err = h.projectService.DeleteProject(c.Request.Context(), id, userID)
	if err != nil {
		if err == domain.ErrProjectNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if err == domain.ErrForbidden {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// ListProjects — список проектов
func (h *ProjectHandler) ListProjects(c *gin.Context) {
	userID, err := h.getUserID(c)
	if err != nil || userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	req := application.ListProjectsRequest{
		Limit:  20,
		Offset: 0,
	}

	if status := c.Query("status"); status != "" {
		s := domain.ProjectStatus(status)
		req.Status = &s
	}

	if ownerID := c.Query("owner_id"); ownerID != "" {
		id, err := uuid.Parse(ownerID)
		if err == nil {
			req.OwnerID = &id
		}
	}

	if memberID := c.Query("member_id"); memberID != "" {
		id, err := uuid.Parse(memberID)
		if err == nil {
			req.MemberID = &id
		}
	}

	req.Tag = c.Query("tag")
	req.Search = c.Query("search")

	if limit := c.Query("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil && l > 0 {
			req.Limit = l
		}
	}
	if offset := c.Query("offset"); offset != "" {
		if o, err := strconv.Atoi(offset); err == nil && o >= 0 {
			req.Offset = o
		}
	}

	projects, total, err := h.projectService.ListProjects(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := dto.ProjectListResponse{
		Projects: make([]dto.ProjectResponse, len(projects)),
		Total:    total,
	}

	for i, project := range projects {
		response.Projects[i] = dto.ToProjectResponse(&project)
	}

	c.JSON(http.StatusOK, response)
}

// GetProjectMembers — список участников проекта
func (h *ProjectHandler) GetProjectMembers(c *gin.Context) {
	userID, err := h.getUserID(c)
	if err != nil || userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project id"})
		return
	}

	members, err := h.projectService.GetMembers(c.Request.Context(), id)
	if err != nil {
		if err == domain.ErrProjectNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := make([]dto.ProjectMemberResponse, len(members))
	for i, member := range members {
		resp := dto.ToProjectMemberResponse(&member)
		if member.User.ID != uuid.Nil {
			resp.User = dto.UserResponse{
				ID:         member.User.ID.String(),
				Email:      member.User.Email,
				Role:       string(member.User.Role),
				FirstName:  member.User.FirstName,
				LastName:   member.User.LastName,
				AvatarURL:  member.User.AvatarURL,
				IsVerified: member.User.IsVerified,
			}
		}
		response[i] = resp
	}

	c.JSON(http.StatusOK, response)
}

// AddMember — добавление участника в проект
func (h *ProjectHandler) AddMember(c *gin.Context) {
	userID, err := h.getUserID(c)
	if err != nil || userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	idParam := c.Param("id")
	projectID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project id"})
		return
	}

	var req dto.AddMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := req.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	memberID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id format"})
		return
	}

	role := domain.ProjectMemberRoleMember
	if req.Role != "" {
		role = domain.ProjectMemberRole(req.Role)
	}

	err = h.projectService.AddMember(c.Request.Context(), projectID, memberID, userID, role)
	if err != nil {
		if err == domain.ErrProjectNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if err == domain.ErrForbidden {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		if err == domain.ErrAlreadyMember {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		if err == domain.ErrProjectFull {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "member added successfully"})
}

// RemoveMember — удаление участника из проекта
func (h *ProjectHandler) RemoveMember(c *gin.Context) {
	userID, err := h.getUserID(c)
	if err != nil || userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project id"})
		return
	}

	memberID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	err = h.projectService.RemoveMember(c.Request.Context(), projectID, memberID, userID)
	if err != nil {
		if err == domain.ErrProjectNotFound || err == domain.ErrNotMember {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if err == domain.ErrForbidden {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		if err == domain.ErrCannotRemoveOwner {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// CreateApplication — создание заявки на вступление
func (h *ProjectHandler) CreateApplication(c *gin.Context) {
	userID, err := h.getUserID(c)
	if err != nil || userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project id"})
		return
	}

	var req dto.CreateApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.projectService.CreateApplication(c.Request.Context(), projectID, userID, req.Message)
	if err != nil {
		if err == domain.ErrProjectNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if err == domain.ErrApplicationExists {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		if err == domain.ErrAlreadyMember {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "application created successfully"})
}

// UpdateApplication — изменение статуса заявки
func (h *ProjectHandler) UpdateApplication(c *gin.Context) {
	userID, err := h.getUserID(c)
	if err != nil || userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	applicationID, err := uuid.Parse(c.Param("applicationId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid application id"})
		return
	}

	var req dto.UpdateApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := req.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	status := domain.ProjectApplicationStatus(req.Status)
	err = h.projectService.UpdateApplicationStatus(c.Request.Context(), applicationID, userID, status)
	if err != nil {
		if err == domain.ErrApplicationNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if err == domain.ErrForbidden {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "application status updated"})
}