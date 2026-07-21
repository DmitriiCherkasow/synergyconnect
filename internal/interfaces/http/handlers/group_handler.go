package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/DmitriiCherkasow/synergyconnect.git/internal/application"
	"github.com/DmitriiCherkasow/synergyconnect.git/internal/interfaces/http/dto"
	"github.com/DmitriiCherkasow/synergyconnect.git/internal/interfaces/http/middleware"
)

// GroupHandler — обработчик для групп
type GroupHandler struct {
	groupService *application.GroupService
}

// NewGroupHandler создает новый обработчик
func NewGroupHandler(groupService *application.GroupService) *GroupHandler {
	return &GroupHandler{
		groupService: groupService,
	}
}

// GetGroupTree — получение дерева групп
// @Summary Получить дерево групп
// @Tags groups
// @Produce json
// @Success 200 {object} []dto.GroupResponse
// @Router /api/v1/groups/tree [get]
func (h *GroupHandler) GetGroupTree(c *gin.Context) {
	tree, err := h.groupService.GetGroupTree(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var responses []dto.GroupResponse
	for _, group := range tree {
		responses = append(responses, dto.ToGroupResponse(&group))
	}

	c.JSON(http.StatusOK, responses)
}

// SubscribeToGroup — подписка на группу
// @Summary Подписаться на группу
// @Tags groups
// @Param id path string true "ID группы"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/groups/{id}/subscribe [post]
func (h *GroupHandler) SubscribeToGroup(c *gin.Context) {
	groupID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group id"})
		return
	}

	userIDStr := middleware.GetUserIDFromContext(c)
	if userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID, _ := uuid.Parse(userIDStr)

	err = h.groupService.SubscribeToGroup(c.Request.Context(), userID, groupID)
	if err != nil {
		if err.Error() == "group not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "group not found"})
			return
		}
		if err.Error() == "already subscribed to this group" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "subscribed successfully"})
}

// UnsubscribeFromGroup — отписка от группы
// @Summary Отписаться от группы
// @Tags groups
// @Param id path string true "ID группы"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/groups/{id}/unsubscribe [delete]
func (h *GroupHandler) UnsubscribeFromGroup(c *gin.Context) {
	groupID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group id"})
		return
	}

	userIDStr := middleware.GetUserIDFromContext(c)
	if userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID, _ := uuid.Parse(userIDStr)

	err = h.groupService.UnsubscribeFromGroup(c.Request.Context(), userID, groupID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "unsubscribed successfully"})
}