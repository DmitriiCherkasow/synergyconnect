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

// AdminHandler — обработчик для административных функций
type AdminHandler struct {
	adminService *application.AdminService
}

// NewAdminHandler создает новый обработчик
func NewAdminHandler(adminService *application.AdminService) *AdminHandler {
	return &AdminHandler{
		adminService: adminService,
	}
}

// getUserID извлекает ID пользователя из контекста
func (h *AdminHandler) getUserID(c *gin.Context) (uuid.UUID, error) {
	userIDStr := middleware.GetUserIDFromContext(c)
	if userIDStr == "" {
		return uuid.Nil, nil
	}
	return uuid.Parse(userIDStr)
}

// GetUsers — список пользователей (админ)
// @Summary Список пользователей
// @Tags admin
// @Security BearerAuth
// @Produce json
// @Param limit query int false "Лимит" default(20)
// @Param offset query int false "Смещение" default(0)
// @Param search query string false "Поиск"
// @Param role query string false "Роль"
// @Param is_active query bool false "Активен"
// @Success 200 {object} dto.UserListResponse
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Router /admin/users [get]
func (h *AdminHandler) GetUsers(c *gin.Context) {
	actorID, err := h.getUserID(c)
	if err != nil || actorID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	search := c.Query("search")
	role := c.Query("role")
	isActive := c.Query("is_active")

	var active *bool
	if isActive != "" {
		val := isActive == "true"
		active = &val
	}

	var userRole *domain.UserRole
	if role != "" {
		r := domain.UserRole(role)
		userRole = &r
	}

	filter := domain.UserFilter{
		Search:   search,
		Role:     userRole,
		IsActive: active,
		Limit:    limit,
		Offset:   offset,
	}

	users, total, err := h.adminService.GetUsers(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := dto.UserListResponse{
		Users: make([]dto.UserResponse, len(users)),
		Total: total,
	}
	for i, user := range users {
		response.Users[i] = dto.ToUserResponse(&user)
	}

	c.JSON(http.StatusOK, response)
}

// GetUser — детали пользователя
// @Summary Детали пользователя
// @Tags admin
// @Security BearerAuth
// @Produce json
// @Param id path string true "ID пользователя"
// @Success 200 {object} dto.UserResponse
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /admin/users/{id} [get]
func (h *AdminHandler) GetUser(c *gin.Context) {
	actorID, err := h.getUserID(c)
	if err != nil || actorID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	idParam := c.Param("id")
	userID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	user, err := h.adminService.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, dto.ToUserResponse(user))
}

// BlockUser — блокировка пользователя
// @Summary Блокировка пользователя
// @Tags admin
// @Security BearerAuth
// @Param id path string true "ID пользователя"
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /admin/users/{id}/block [put]
func (h *AdminHandler) BlockUser(c *gin.Context) {
	actorID, err := h.getUserID(c)
	if err != nil || actorID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	idParam := c.Param("id")
	userID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	if err := h.adminService.BlockUser(c.Request.Context(), actorID, userID); err != nil {
		if err == domain.ErrUserNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if err == domain.ErrInsufficientPermissions {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		if err == domain.ErrCannotChangeSuperAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user blocked"})
}

// UnblockUser — разблокировка пользователя
// @Summary Разблокировка пользователя
// @Tags admin
// @Security BearerAuth
// @Param id path string true "ID пользователя"
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /admin/users/{id}/unblock [put]
func (h *AdminHandler) UnblockUser(c *gin.Context) {
	actorID, err := h.getUserID(c)
	if err != nil || actorID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	idParam := c.Param("id")
	userID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	if err := h.adminService.UnblockUser(c.Request.Context(), actorID, userID); err != nil {
		if err == domain.ErrUserNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if err == domain.ErrInsufficientPermissions {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user unblocked"})
}

// DeleteUser — удаление пользователя
// @Summary Удаление пользователя
// @Tags admin
// @Security BearerAuth
// @Param id path string true "ID пользователя"
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /admin/users/{id} [delete]
func (h *AdminHandler) DeleteUser(c *gin.Context) {
	actorID, err := h.getUserID(c)
	if err != nil || actorID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	idParam := c.Param("id")
	userID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	if err := h.adminService.DeleteUser(c.Request.Context(), actorID, userID); err != nil {
		if err == domain.ErrUserNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if err == domain.ErrInsufficientPermissions {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		if err == domain.ErrCannotDeleteSuperAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user deleted"})
}

// PromoteToAdmin — повышение до администратора
// @Summary Повышение до администратора
// @Tags admin
// @Security BearerAuth
// @Param id path string true "ID пользователя"
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /admin/admins/promote/{id} [post]
func (h *AdminHandler) PromoteToAdmin(c *gin.Context) {
	actorID, err := h.getUserID(c)
	if err != nil || actorID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	idParam := c.Param("id")
	userID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	if err := h.adminService.PromoteToAdmin(c.Request.Context(), actorID, userID); err != nil {
		if err == domain.ErrUserNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if err == domain.ErrInsufficientPermissions {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		if err == domain.ErrCannotChangeSuperAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		if err == domain.ErrUserAlreadyAdmin {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user promoted to admin"})
}

// DemoteFromAdmin — понижение администратора
// @Summary Понижение администратора
// @Tags admin
// @Security BearerAuth
// @Param id path string true "ID администратора"
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /admin/admins/demote/{id} [post]
func (h *AdminHandler) DemoteFromAdmin(c *gin.Context) {
	actorID, err := h.getUserID(c)
	if err != nil || actorID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	idParam := c.Param("id")
	userID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	if err := h.adminService.DemoteFromAdmin(c.Request.Context(), actorID, userID); err != nil {
		if err == domain.ErrUserNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if err == domain.ErrInsufficientPermissions {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		if err == domain.ErrCannotChangeSuperAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		if err == domain.ErrUserNotAdmin {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "admin demoted"})
}

// GetAdmins — список администраторов
// @Summary Список администраторов
// @Tags admin
// @Security BearerAuth
// @Produce json
// @Success 200 {array} dto.UserResponse
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Router /admin/admins [get]
func (h *AdminHandler) GetAdmins(c *gin.Context) {
	actorID, err := h.getUserID(c)
	if err != nil || actorID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	admins, err := h.adminService.GetAdmins(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := make([]dto.UserResponse, len(admins))
	for i, admin := range admins {
		response[i] = dto.ToUserResponse(&admin)
	}

	c.JSON(http.StatusOK, response)
}

// GetSuperAdmin — получение суперадмина
// @Summary Получение суперадмина
// @Tags admin
// @Security BearerAuth
// @Produce json
// @Success 200 {object} dto.UserResponse
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /admin/super-admin [get]
func (h *AdminHandler) GetSuperAdmin(c *gin.Context) {
	actorID, err := h.getUserID(c)
	if err != nil || actorID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	admin, err := h.adminService.GetSuperAdmin(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if admin == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "super admin not found"})
		return
	}

	c.JSON(http.StatusOK, dto.ToUserResponse(admin))
}

// GetPosts — список постов (админ)
// @Summary Список постов
// @Tags admin
// @Security BearerAuth
// @Produce json
// @Param limit query int false "Лимит" default(20)
// @Param offset query int false "Смещение" default(0)
// @Param search query string false "Поиск"
// @Param category query string false "Категория"
// @Success 200 {object} dto.PostListResponse
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Router /admin/posts [get]
func (h *AdminHandler) GetPosts(c *gin.Context) {
	actorID, err := h.getUserID(c)
	if err != nil || actorID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	search := c.Query("search")
	category := c.Query("category")

	var postCategory *domain.PostCategory
	if category != "" {
		cat := domain.PostCategory(category)
		postCategory = &cat
	}

	filter := domain.PostFilter{
		Category: postCategory,
		Search:   search,
		Limit:    limit,
		Offset:   offset,
	}

	posts, total, err := h.adminService.GetPosts(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := dto.PostListResponse{
		Posts: make([]dto.PostResponse, len(posts)),
		Total: total,
	}
	for i, post := range posts {
		response.Posts[i] = dto.ToPostResponse(&post, []domain.Comment{}, []domain.Tag{})
	}

	c.JSON(http.StatusOK, response)
}

// DeletePost — удаление поста (админ)
// @Summary Удаление поста
// @Tags admin
// @Security BearerAuth
// @Param id path string true "ID поста"
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /admin/posts/{id} [delete]
func (h *AdminHandler) DeletePost(c *gin.Context) {
	actorID, err := h.getUserID(c)
	if err != nil || actorID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	idParam := c.Param("id")
	postID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid post id"})
		return
	}

	if err := h.adminService.DeletePost(c.Request.Context(), actorID, postID); err != nil {
		if err == domain.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if err == domain.ErrInsufficientPermissions {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "post deleted"})
}

// GetComments — список комментариев (админ)
// @Summary Список комментариев
// @Tags admin
// @Security BearerAuth
// @Produce json
// @Param limit query int false "Лимит" default(20)
// @Param offset query int false "Смещение" default(0)
// @Param search query string false "Поиск"
// @Param post_id query string false "ID поста"
// @Success 200 {object} dto.CommentListResponse
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Router /admin/comments [get]
func (h *AdminHandler) GetComments(c *gin.Context) {
	actorID, err := h.getUserID(c)
	if err != nil || actorID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	search := c.Query("search")
	postIDStr := c.Query("post_id")

	var postID *uuid.UUID
	if postIDStr != "" {
		id, err := uuid.Parse(postIDStr)
		if err == nil {
			postID = &id
		}
	}

	filter := domain.CommentFilter{
		PostID: postID,
		Search: search,
		Limit:  limit,
		Offset: offset,
	}

	comments, total, err := h.adminService.GetComments(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := dto.CommentListResponse{
		Comments: make([]dto.CommentResponse, len(comments)),
		Total:    total,
	}
	for i, comment := range comments {
		response.Comments[i] = dto.ToCommentResponse(&comment)
	}

	c.JSON(http.StatusOK, response)
}

// DeleteComment — удаление комментария (админ)
// @Summary Удаление комментария
// @Tags admin
// @Security BearerAuth
// @Param id path string true "ID комментария"
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /admin/comments/{id} [delete]
func (h *AdminHandler) DeleteComment(c *gin.Context) {
	actorID, err := h.getUserID(c)
	if err != nil || actorID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	idParam := c.Param("id")
	commentID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid comment id"})
		return
	}

	if err := h.adminService.DeleteComment(c.Request.Context(), actorID, commentID); err != nil {
		if err == domain.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if err == domain.ErrInsufficientPermissions {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "comment deleted"})
}

// GetProjects — список проектов (админ)
// @Summary Список проектов
// @Tags admin
// @Security BearerAuth
// @Produce json
// @Param limit query int false "Лимит" default(20)
// @Param offset query int false "Смещение" default(0)
// @Param status query string false "Статус"
// @Param search query string false "Поиск"
// @Success 200 {object} dto.ProjectListResponse
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Router /admin/projects [get]
func (h *AdminHandler) GetProjects(c *gin.Context) {
	actorID, err := h.getUserID(c)
	if err != nil || actorID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	status := c.Query("status")
	search := c.Query("search")

	var projectStatus *domain.ProjectStatus
	if status != "" {
		s := domain.ProjectStatus(status)
		projectStatus = &s
	}

	filter := domain.ProjectFilter{
		Status: projectStatus,
		Search: search,
		Limit:  limit,
		Offset: offset,
	}

	projects, total, err := h.adminService.GetProjects(c.Request.Context(), filter)
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

// DeleteProject — удаление проекта (админ)
// @Summary Удаление проекта
// @Tags admin
// @Security BearerAuth
// @Param id path string true "ID проекта"
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /admin/projects/{id} [delete]
func (h *AdminHandler) DeleteProject(c *gin.Context) {
	actorID, err := h.getUserID(c)
	if err != nil || actorID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	idParam := c.Param("id")
	projectID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project id"})
		return
	}

	if err := h.adminService.DeleteProject(c.Request.Context(), actorID, projectID); err != nil {
		if err == domain.ErrProjectNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if err == domain.ErrInsufficientPermissions {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "project deleted"})
}

// GetVacancies — список вакансий (админ)
// @Summary Список вакансий
// @Tags admin
// @Security BearerAuth
// @Produce json
// @Param limit query int false "Лимит" default(20)
// @Param offset query int false "Смещение" default(0)
// @Param status query string false "Статус"
// @Param search query string false "Поиск"
// @Success 200 {object} dto.VacancyListResponse
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Router /admin/vacancies [get]
func (h *AdminHandler) GetVacancies(c *gin.Context) {
	actorID, err := h.getUserID(c)
	if err != nil || actorID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	status := c.Query("status")
	search := c.Query("search")

	var vacancyStatus *domain.VacancyStatus
	if status != "" {
		s := domain.VacancyStatus(status)
		vacancyStatus = &s
	}

	filter := domain.VacancyFilter{
		Status: vacancyStatus,
		Search: search,
		Limit:  limit,
		Offset: offset,
	}

	vacancies, total, err := h.adminService.GetVacancies(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := dto.VacancyListResponse{
		Vacancies: make([]dto.VacancyResponse, len(vacancies)),
		Total:     total,
	}
	for i, vacancy := range vacancies {
		response.Vacancies[i] = dto.ToVacancyResponse(&vacancy)
	}

	c.JSON(http.StatusOK, response)
}

// DeleteVacancy — удаление вакансии (админ)
// @Summary Удаление вакансии
// @Tags admin
// @Security BearerAuth
// @Param id path string true "ID вакансии"
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /admin/vacancies/{id} [delete]
func (h *AdminHandler) DeleteVacancy(c *gin.Context) {
	actorID, err := h.getUserID(c)
	if err != nil || actorID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	idParam := c.Param("id")
	vacancyID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid vacancy id"})
		return
	}

	if err := h.adminService.DeleteVacancy(c.Request.Context(), actorID, vacancyID); err != nil {
		if err == domain.ErrVacancyNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if err == domain.ErrInsufficientPermissions {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "vacancy deleted"})
}

// DeleteAdmin — удаление администратора (только суперадмин)
// @Summary Удаление администратора
// @Tags admin
// @Security BearerAuth
// @Param id path string true "ID администратора"
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /admin/admins/{id} [delete]
func (h *AdminHandler) DeleteAdmin(c *gin.Context) {
	actorID, err := h.getUserID(c)
	if err != nil || actorID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Проверяем, что актор — суперадмин
	actor, err := h.adminService.GetUserByID(c.Request.Context(), actorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if actor == nil || !actor.IsSuperAdmin() {
		c.JSON(http.StatusForbidden, gin.H{"error": "super admin access required"})
		return
	}

	idParam := c.Param("id")
	adminID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid admin id"})
		return
	}

	// Нельзя удалить самого себя (суперадмина)
	if actorID == adminID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot delete super admin"})
		return
	}

	admin, err := h.adminService.GetUserByID(c.Request.Context(), adminID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if admin == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "admin not found"})
		return
	}

	if !admin.IsAdmin() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user is not an admin"})
		return
	}

	// Понижаем до студента (удаляем из админов)
	if err := h.adminService.DemoteFromAdmin(c.Request.Context(), actorID, adminID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "admin removed successfully"})
}

// GetStats — статистика
// @Summary Статистика
// @Tags admin
// @Security BearerAuth
// @Produce json
// @Success 200 {object} application.Stats
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Router /admin/stats [get]
func (h *AdminHandler) GetStats(c *gin.Context) {
	actorID, err := h.getUserID(c)
	if err != nil || actorID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	stats, err := h.adminService.GetStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}