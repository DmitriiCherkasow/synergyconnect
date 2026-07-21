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

// StickerHandler — обработчик для стикеров
type StickerHandler struct {
	stickerService *application.StickerService
}

// NewStickerHandler создает новый обработчик
func NewStickerHandler(stickerService *application.StickerService) *StickerHandler {
	return &StickerHandler{
		stickerService: stickerService,
	}
}

// CreateSticker — создание стикера
// @Summary Создать стикер
// @Tags stickers
// @Accept json
// @Produce json
// @Param boardId path string true "ID доски"
// @Param request body dto.CreateStickerRequest true "Данные стикера"
// @Success 201 {object} dto.StickerResponse
// @Failure 400 {object} map[string]interface{}
// @Router /api/v1/boards/{boardId}/stickers [post]
func (h *StickerHandler) CreateSticker(c *gin.Context) {
	boardID, err := uuid.Parse(c.Param("boardId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid board id"})
		return
	}

	var req dto.CreateStickerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userIDStr := middleware.GetUserIDFromContext(c)
	if userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID, _ := uuid.Parse(userIDStr)

	// Устанавливаем значения по умолчанию
	color := domain.StickerColor(req.Color)
	if color == "" {
		color = domain.ColorYellow
	}

	priority := domain.StickerPriority(req.Priority)
	if priority == "" {
		priority = domain.PriorityMedium
	}

	width := req.Width
	if width == 0 {
		width = 200
	}
	height := req.Height
	if height == 0 {
		height = 150
	}

	sticker, err := h.stickerService.CreateSticker(c.Request.Context(), application.CreateStickerRequest{
		BoardID:   boardID,
		AuthorID:  userID,
		Title:     req.Title,
		Content:   req.Content,
		Color:     color,
		Priority:  priority,
		PositionX: req.PositionX,
		PositionY: req.PositionY,
		Width:     width,
		Height:    height,
		DueDate:   req.DueDate,
	})
	if err != nil {
		if err.Error() == "board not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "board not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, dto.ToStickerResponse(sticker))
}

// GetSticker — получение стикера
// @Summary Получить стикер
// @Tags stickers
// @Produce json
// @Param id path string true "ID стикера"
// @Success 200 {object} dto.StickerResponse
// @Failure 404 {object} map[string]interface{}
// @Router /api/v1/stickers/{id} [get]
func (h *StickerHandler) GetSticker(c *gin.Context) {
	stickerID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid sticker id"})
		return
	}

	sticker, err := h.stickerService.GetSticker(c.Request.Context(), stickerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if sticker == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "sticker not found"})
		return
	}

	c.JSON(http.StatusOK, dto.ToStickerResponse(sticker))
}

// UpdateSticker — обновление стикера
// @Summary Обновить стикер
// @Tags stickers
// @Accept json
// @Produce json
// @Param id path string true "ID стикера"
// @Param request body dto.UpdateStickerRequest true "Данные для обновления"
// @Success 200 {object} dto.StickerResponse
// @Failure 404 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Router /api/v1/stickers/{id} [put]
func (h *StickerHandler) UpdateSticker(c *gin.Context) {
	stickerID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid sticker id"})
		return
	}

	var req dto.UpdateStickerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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

	updateReq := application.UpdateStickerRequest{
		StickerID: stickerID,
		UserID:    userID,
		UserRole:  userRole,
		Title:     req.Title,
		Content:   req.Content,
		Width:     req.Width,
		Height:    req.Height,
		DueDate:   req.DueDate,
	}

	if req.Color != nil {
		color := domain.StickerColor(*req.Color)
		updateReq.Color = &color
	}
	if req.Priority != nil {
		priority := domain.StickerPriority(*req.Priority)
		updateReq.Priority = &priority
	}

	sticker, err := h.stickerService.UpdateSticker(c.Request.Context(), updateReq)
	if err != nil {
		if err.Error() == "sticker not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "sticker not found"})
			return
		}
		if err.Error() == "access denied" {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.ToStickerResponse(sticker))
}

// DeleteSticker — удаление стикера
// @Summary Удалить стикер
// @Tags stickers
// @Param id path string true "ID стикера"
// @Success 204 "No Content"
// @Failure 404 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Router /api/v1/stickers/{id} [delete]
func (h *StickerHandler) DeleteSticker(c *gin.Context) {
	stickerID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid sticker id"})
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

	err = h.stickerService.DeleteSticker(c.Request.Context(), stickerID, userID, userRole)
	if err != nil {
		if err.Error() == "sticker not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "sticker not found"})
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

// ToggleComplete — переключение статуса выполнения стикера
// @Summary Переключить статус выполнения
// @Tags stickers
// @Param id path string true "ID стикера"
// @Success 200 {object} dto.StickerResponse
// @Failure 404 {object} map[string]interface{}
// @Router /api/v1/stickers/{id}/toggle-complete [patch]
func (h *StickerHandler) ToggleComplete(c *gin.Context) {
	stickerID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid sticker id"})
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

	err = h.stickerService.ToggleComplete(c.Request.Context(), stickerID, userID, userRole)
	if err != nil {
		if err.Error() == "sticker not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "sticker not found"})
			return
		}
		if err.Error() == "access denied" {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	sticker, _ := h.stickerService.GetSticker(c.Request.Context(), stickerID)
	c.JSON(http.StatusOK, dto.ToStickerResponse(sticker))
}

// UpdatePosition — обновление позиции стикера
// @Summary Обновить позицию стикера
// @Tags stickers
// @Accept json
// @Produce json
// @Param id path string true "ID стикера"
// @Param request body dto.UpdateStickerPositionRequest true "Новая позиция"
// @Success 200 {object} dto.StickerResponse
// @Failure 404 {object} map[string]interface{}
// @Router /api/v1/stickers/{id}/position [patch]
func (h *StickerHandler) UpdatePosition(c *gin.Context) {
	stickerID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid sticker id"})
		return
	}

	var req dto.UpdateStickerPositionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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

	err = h.stickerService.UpdatePosition(c.Request.Context(), stickerID, userID, userRole, req.PositionX, req.PositionY)
	if err != nil {
		if err.Error() == "sticker not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "sticker not found"})
			return
		}
		if err.Error() == "access denied" {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	sticker, _ := h.stickerService.GetSticker(c.Request.Context(), stickerID)
	c.JSON(http.StatusOK, dto.ToStickerResponse(sticker))
}