package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/DmitriiCherkasow/synergyconnect.git/internal/application"
	"github.com/DmitriiCherkasow/synergyconnect.git/internal/domain"
	"github.com/DmitriiCherkasow/synergyconnect.git/internal/interfaces/http/dto"
	"github.com/DmitriiCherkasow/synergyconnect.git/internal/interfaces/http/middleware"
)

// PostHandler — обработчик для постов
type PostHandler struct {
	postService *application.PostService
	groupService *application.GroupService
}

// NewPostHandler создает новый обработчик
func NewPostHandler(postService *application.PostService, groupService *application.GroupService) *PostHandler {
	return &PostHandler{
		postService:  postService,
		groupService: groupService,
	}
}

// CreatePost — создание поста
// @Summary Создать пост
// @Tags posts
// @Accept json
// @Produce json
// @Param request body dto.CreatePostRequest true "Данные поста"
// @Success 201 {object} dto.PostResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/v1/posts [post]
func (h *PostHandler) CreatePost(c *gin.Context) {
	var req dto.CreatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := middleware.GetUserIDFromContext(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	authorID, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	// Преобразуем строки в типы
	category := domain.PostCategory(req.Category)
	visibility := domain.PostVisibility(req.Visibility)

	// Создаем пост
	post, err := h.postService.CreatePost(c.Request.Context(), application.CreatePostRequest{
		AuthorID:   authorID,
		GroupID:    req.GroupID,
		Title:      req.Title,
		Content:    req.Content,
		Category:   category,
		Visibility: visibility,
		TagNames:   req.Tags,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Получаем теги поста
	// TODO: Получить теги из репозитория

	c.JSON(http.StatusCreated, dto.ToPostResponse(post, nil, nil))
}

// GetPost — получение поста по ID
// @Summary Получить пост
// @Tags posts
// @Produce json
// @Param id path string true "ID поста"
// @Success 200 {object} dto.PostResponse
// @Failure 404 {object} map[string]interface{}
// @Router /api/v1/posts/{id} [get]
func (h *PostHandler) GetPost(c *gin.Context) {
	postID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid post id"})
		return
	}

	userIDStr := middleware.GetUserIDFromContext(c)
	if userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID, _ := uuid.Parse(userIDStr)

	// Получаем роль пользователя
	userRole := domain.RoleStudent
	if role, exists := c.Get("user_role"); exists {
		userRole = domain.UserRole(role.(string))
	}

	// Получаем ID групп пользователя
	userGroupIDs, err := h.groupService.GetUserGroupIDs(c.Request.Context(), userID)
	if err != nil {
		userGroupIDs = []uuid.UUID{}
	}

	post, err := h.postService.GetPost(c.Request.Context(), postID, userID, userRole, userGroupIDs)
	if err != nil {
		if err.Error() == "post not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "post not found"})
			return
		}
		if err.Error() == "access denied" {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Получаем комментарии
	comments, err := h.postService.GetComments(c.Request.Context(), postID)
	if err != nil {
		comments = []domain.Comment{}
	}

	c.JSON(http.StatusOK, dto.ToPostResponse(post, comments, nil))
}

// UpdatePost — обновление поста
// @Summary Обновить пост
// @Tags posts
// @Accept json
// @Produce json
// @Param id path string true "ID поста"
// @Param request body dto.UpdatePostRequest true "Данные для обновления"
// @Success 200 {object} dto.PostResponse
// @Failure 404 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Router /api/v1/posts/{id} [put]
func (h *PostHandler) UpdatePost(c *gin.Context) {
	postID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid post id"})
		return
	}

	userIDStr := middleware.GetUserIDFromContext(c)
	if userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID, _ := uuid.Parse(userIDStr)

	userRole := domain.RoleStudent
	if role, exists := c.Get("user_role"); exists {
		userRole = domain.UserRole(role.(string))
	}

	var req dto.UpdatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updateReq := application.UpdatePostRequest{
		PostID:   postID,
		UserID:   userID,
		UserRole: userRole,
		Title:    req.Title,
		Content:  req.Content,
	}

	if req.Category != nil {
		category := domain.PostCategory(*req.Category)
		updateReq.Category = &category
	}
	if req.Visibility != nil {
		visibility := domain.PostVisibility(*req.Visibility)
		updateReq.Visibility = &visibility
	}

	post, err := h.postService.UpdatePost(c.Request.Context(), updateReq)
	if err != nil {
		if err.Error() == "post not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "post not found"})
			return
		}
		if err.Error() == "access denied" {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.ToPostResponse(post, nil, nil))
}

// DeletePost — удаление поста
// @Summary Удалить пост
// @Tags posts
// @Param id path string true "ID поста"
// @Success 204 "No Content"
// @Failure 404 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Router /api/v1/posts/{id} [delete]
func (h *PostHandler) DeletePost(c *gin.Context) {
	postID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid post id"})
		return
	}

	userIDStr := middleware.GetUserIDFromContext(c)
	if userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID, _ := uuid.Parse(userIDStr)

	userRole := domain.RoleStudent
	if role, exists := c.Get("user_role"); exists {
		userRole = domain.UserRole(role.(string))
	}

	err = h.postService.DeletePost(c.Request.Context(), postID, userID, userRole)
	if err != nil {
		if err.Error() == "post not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "post not found"})
			return
		}
		if err.Error() == "access denied" {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// GetPublicFeed — публичная лента
// @Summary Получить публичную ленту
// @Tags posts
// @Produce json
// @Param tag query string false "Фильтр по тегу"
// @Param limit query int false "Лимит" default(20)
// @Param offset query int false "Смещение" default(0)
// @Success 200 {array} dto.PostResponse
// @Router /api/v1/posts/public [get]
func (h *PostHandler) GetPublicFeed(c *gin.Context) {
	tagSlug := c.Query("tag")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	if limit > 100 {
		limit = 100
	}

	posts, err := h.postService.GetPublicFeed(c.Request.Context(), tagSlug, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var responses []dto.PostResponse
	for _, post := range posts {
		comments, _ := h.postService.GetComments(c.Request.Context(), post.ID)
		responses = append(responses, dto.ToPostResponse(&post, comments, nil))
	}

	c.JSON(http.StatusOK, responses)
}

// GetFeed — лента подписок (требуется авторизация)
// @Summary Получить ленту подписок
// @Tags posts
// @Produce json
// @Param limit query int false "Лимит" default(20)
// @Param offset query int false "Смещение" default(0)
// @Success 200 {array} dto.PostResponse
// @Router /api/v1/posts/feed [get]
func (h *PostHandler) GetFeed(c *gin.Context) {
	userIDStr := middleware.GetUserIDFromContext(c)
	if userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID, _ := uuid.Parse(userIDStr)

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	if limit > 100 {
		limit = 100
	}

	posts, err := h.postService.GetFeedBySubscriptions(c.Request.Context(), userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var responses []dto.PostResponse
	for _, post := range posts {
		comments, _ := h.postService.GetComments(c.Request.Context(), post.ID)
		responses = append(responses, dto.ToPostResponse(&post, comments, nil))
	}

	c.JSON(http.StatusOK, responses)
}