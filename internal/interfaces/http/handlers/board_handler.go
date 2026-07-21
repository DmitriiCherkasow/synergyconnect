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

// BoardHandler — обработчик для досок
type BoardHandler struct {
	boardService   *application.BoardService
	stickerService *application.StickerService
}

// NewBoardHandler создает новый обработчик
func NewBoardHandler(
	boardService *application.BoardService,
	stickerService *application.StickerService,
) *BoardHandler {
	return &BoardHandler{
		boardService:   boardService,
		stickerService: stickerService,
	}
}

// CreateBoard — создание доски
func (h *BoardHandler) CreateBoard(c *gin.Context) {
	var req dto.CreateBoardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userIDStr := middleware.GetUserIDFromContext(c)
	if userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	colorHex := req.ColorHex
	if colorHex == "" {
		colorHex = "#3498db"
	}

	board, err := h.boardService.CreateBoard(c.Request.Context(), application.CreateBoardRequest{
		OwnerID:     userID,
		GroupID:     req.GroupID,
		Title:       req.Title,
		Description: req.Description,
		ColorHex:    colorHex,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	stickers, _ := h.stickerService.GetBoardStickers(c.Request.Context(), board.ID)
	c.JSON(http.StatusCreated, dto.ToBoardResponse(board, stickers))
}

// GetBoard — получение доски по ID
func (h *BoardHandler) GetBoard(c *gin.Context) {
	boardID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid board id"})
		return
	}

	userIDStr := middleware.GetUserIDFromContext(c)
	if userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	userRole := domain.RoleStudent
	if role, exists := c.Get("user_role"); exists {
		userRole = domain.UserRole(role.(string))
	}

	userGroupIDs := []uuid.UUID{}

	board, err := h.boardService.GetBoard(c.Request.Context(), boardID, userID, userRole, userGroupIDs)
	if err != nil {
		if err.Error() == "board not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "board not found"})
			return
		}
		if err.Error() == "access denied" {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	stickers, _ := h.stickerService.GetBoardStickers(c.Request.Context(), board.ID)
	c.JSON(http.StatusOK, dto.ToBoardResponse(board, stickers))
}

// GetUserBoards — получение всех досок пользователя
func (h *BoardHandler) GetUserBoards(c *gin.Context) {
	userIDStr := middleware.GetUserIDFromContext(c)
	if userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	userGroupIDs := []uuid.UUID{}

	boards, err := h.boardService.GetUserBoards(c.Request.Context(), userID, userGroupIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var responses []dto.BoardResponse
	for _, board := range boards {
		stickers, _ := h.stickerService.GetBoardStickers(c.Request.Context(), board.ID)
		responses = append(responses, dto.ToBoardResponse(&board, stickers))
	}

	c.JSON(http.StatusOK, responses)
}

// UpdateBoard — обновление доски
func (h *BoardHandler) UpdateBoard(c *gin.Context) {
	boardID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid board id"})
		return
	}

	var req dto.UpdateBoardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userIDStr := middleware.GetUserIDFromContext(c)
	if userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	userRole := domain.RoleStudent
	if role, exists := c.Get("user_role"); exists {
		userRole = domain.UserRole(role.(string))
	}

	board, err := h.boardService.UpdateBoard(
		c.Request.Context(),
		boardID,
		userID,
		userRole,
		req.Title,
		req.Description,
		req.ColorHex,
	)
	if err != nil {
		if err.Error() == "board not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "board not found"})
			return
		}
		if err.Error() == "access denied" {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	stickers, _ := h.stickerService.GetBoardStickers(c.Request.Context(), board.ID)
	c.JSON(http.StatusOK, dto.ToBoardResponse(board, stickers))
}

// DeleteBoard — удаление доски
func (h *BoardHandler) DeleteBoard(c *gin.Context) {
	boardID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid board id"})
		return
	}

	userIDStr := middleware.GetUserIDFromContext(c)
	if userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	userRole := domain.RoleStudent
	if role, exists := c.Get("user_role"); exists {
		userRole = domain.UserRole(role.(string))
	}

	err = h.boardService.DeleteBoard(c.Request.Context(), boardID, userID, userRole)
	if err != nil {
		if err.Error() == "board not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "board not found"})
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

// ArchiveBoard — архивирование доски
func (h *BoardHandler) ArchiveBoard(c *gin.Context) {
	boardID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid board id"})
		return
	}

	userIDStr := middleware.GetUserIDFromContext(c)
	if userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	userRole := domain.RoleStudent
	if role, exists := c.Get("user_role"); exists {
		userRole = domain.UserRole(role.(string))
	}

	board, err := h.boardService.GetBoard(c.Request.Context(), boardID, userID, userRole, []uuid.UUID{})
	if err != nil {
		if err.Error() == "board not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "board not found"})
			return
		}
		if err.Error() == "access denied" {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if !board.CanEdit(userID, userRole) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	err = h.boardService.ArchiveBoard(c.Request.Context(), boardID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "board archived successfully"})
}

// UnarchiveBoard — разархивирование доски
func (h *BoardHandler) UnarchiveBoard(c *gin.Context) {
	boardID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid board id"})
		return
	}

	userIDStr := middleware.GetUserIDFromContext(c)
	if userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	userRole := domain.RoleStudent
	if role, exists := c.Get("user_role"); exists {
		userRole = domain.UserRole(role.(string))
	}

	board, err := h.boardService.GetBoard(c.Request.Context(), boardID, userID, userRole, []uuid.UUID{})
	if err != nil {
		if err.Error() == "board not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "board not found"})
			return
		}
		if err.Error() == "access denied" {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if !board.CanEdit(userID, userRole) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	err = h.boardService.UnarchiveBoard(c.Request.Context(), boardID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "board unarchived successfully"})
}