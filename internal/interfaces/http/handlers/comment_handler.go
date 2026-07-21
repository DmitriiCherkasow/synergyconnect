package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/DmitriiCherkasow/synergyconnect.git/internal/application"
	"github.com/DmitriiCherkasow/synergyconnect.git/internal/domain"
	"github.com/DmitriiCherkasow/synergyconnect.git/internal/interfaces/http/dto"
	"github.com/DmitriiCherkasow/synergyconnect.git/internal/interfaces/http/middleware"
)

// CommentHandler — обработчик для комментариев
type CommentHandler struct {
	postService *application.PostService
}

// NewCommentHandler создает новый обработчик
func NewCommentHandler(postService *application.PostService) *CommentHandler {
	return &CommentHandler{
		postService: postService,
	}
}

// AddComment — добавление комментария к посту
// @Summary Добавить комментарий
// @Tags comments
// @Accept json
// @Produce json
// @Param postId path string true "ID поста"
// @Param request body object true "Данные комментария"
// @Success 201 {object} dto.CommentResponse
// @Failure 400 {object} map[string]interface{}
// @Router /api/v1/posts/{postId}/comments [post]
func (h *CommentHandler) AddComment(c *gin.Context) {
	postID, err := uuid.Parse(c.Param("postId"))
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

	var req struct {
		Content string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	comment, err := h.postService.AddComment(c.Request.Context(), postID, userID, req.Content, nil)
	if err != nil {
		if err.Error() == "post not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "post not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, dto.CommentResponse{
		ID:        comment.ID.String(),
		Author:    dto.ToUserResponse(&domain.User{ID: userID}), // TODO: Загрузить пользователя
		Content:   comment.Content,
		CreatedAt: comment.CreatedAt,
		UpdatedAt: comment.UpdatedAt,
	})
}

// DeleteComment — удаление комментария
// @Summary Удалить комментарий
// @Tags comments
// @Param id path string true "ID комментария"
// @Success 204 "No Content"
// @Failure 404 {object} map[string]interface{}
// @Router /api/v1/comments/{id} [delete]
func (h *CommentHandler) DeleteComment(c *gin.Context) {
	commentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid comment id"})
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

	err = h.postService.DeleteComment(c.Request.Context(), commentID, userID, userRole)
	if err != nil {
		if err.Error() == "comment not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "comment not found"})
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