package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/DmitriiCherkasow/synergyconnect.git/internal/application"
	"github.com/DmitriiCherkasow/synergyconnect.git/internal/domain"
	"github.com/DmitriiCherkasow/synergyconnect.git/internal/interfaces/http/dto"
	"github.com/DmitriiCherkasow/synergyconnect.git/internal/interfaces/http/middleware"
)

// ReminderHandler — обработчик для напоминаний
type ReminderHandler struct {
	reminderService *application.ReminderService
}

// NewReminderHandler создает новый обработчик
func NewReminderHandler(reminderService *application.ReminderService) *ReminderHandler {
	return &ReminderHandler{
		reminderService: reminderService,
	}
}

// CreateReminder — создание напоминания для стикера
// @Summary Создать напоминание
// @Tags reminders
// @Accept json
// @Produce json
// @Param stickerId path string true "ID стикера"
// @Param request body dto.CreateReminderRequest true "Данные напоминания"
// @Success 201 {object} dto.ReminderResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /api/v1/stickers/{stickerId}/reminders [post]
func (h *ReminderHandler) CreateReminder(c *gin.Context) {
	stickerID, err := uuid.Parse(c.Param("stickerId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid sticker id"})
		return
	}

	var req dto.CreateReminderRequest
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

	// Проверяем, что время не в прошлом
	if req.RemindAt.Before(time.Now()) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "reminder time cannot be in the past"})
		return
	}

	// Преобразуем recurrence
	recurrence := domain.RecurrenceOneTime
	if req.Recurrence != "" {
		recurrence = domain.ReminderRecurrence(req.Recurrence)
	}

	// Проверяем интервал для custom
	if recurrence == domain.RecurrenceCustom && req.RecurrenceInterval == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "recurrence_interval is required for custom recurrence"})
		return
	}

	reminder, err := h.reminderService.CreateReminder(c.Request.Context(), application.CreateReminderRequest{
		StickerID:          stickerID,
		UserID:             userID,
		RemindAt:           req.RemindAt,
		Recurrence:         recurrence,
		RecurrenceInterval: req.RecurrenceInterval,
		WarningMinutes:     req.WarningMinutes,
	})
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

	// Получаем стикер для заголовка
	sticker, _ := h.reminderService.GetSticker(c.Request.Context(), stickerID)
	stickerTitle := ""
	if sticker != nil {
		stickerTitle = sticker.Title
	}

	c.JSON(http.StatusCreated, dto.ToReminderResponse(reminder, stickerTitle))
}

// GetUserReminders — получение всех напоминаний пользователя
// @Summary Получить все напоминания пользователя
// @Tags reminders
// @Produce json
// @Success 200 {array} dto.ReminderResponse
// @Router /api/v1/reminders [get]
func (h *ReminderHandler) GetUserReminders(c *gin.Context) {
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

	reminders, err := h.reminderService.GetUserReminders(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var responses []dto.ReminderResponse
	for _, reminder := range reminders {
		// Получаем стикер для заголовка
		sticker, _ := h.reminderService.GetSticker(c.Request.Context(), reminder.StickerID)
		stickerTitle := ""
		if sticker != nil {
			stickerTitle = sticker.Title
		}
		responses = append(responses, dto.ToReminderResponse(&reminder, stickerTitle))
	}

	c.JSON(http.StatusOK, responses)
}

// DeleteReminder — удаление напоминания
// @Summary Удалить напоминание
// @Tags reminders
// @Param id path string true "ID напоминания"
// @Success 204 "No Content"
// @Failure 404 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Router /api/v1/reminders/{id} [delete]
func (h *ReminderHandler) DeleteReminder(c *gin.Context) {
	reminderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid reminder id"})
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

	err = h.reminderService.DeleteReminder(c.Request.Context(), reminderID, userID)
	if err != nil {
		if err.Error() == "reminder not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "reminder not found"})
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

// SnoozeReminder — откладывание напоминания
// @Summary Отложить напоминание
// @Tags reminders
// @Accept json
// @Produce json
// @Param id path string true "ID напоминания"
// @Param request body dto.SnoozeReminderRequest true "Данные для откладывания"
// @Success 200 {object} dto.ReminderResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /api/v1/reminders/{id}/snooze [patch]
func (h *ReminderHandler) SnoozeReminder(c *gin.Context) {
	reminderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid reminder id"})
		return
	}

	var req dto.SnoozeReminderRequest
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

	reminder, err := h.reminderService.SnoozeReminder(c.Request.Context(), reminderID, userID, req.Minutes)
	if err != nil {
		if err.Error() == "reminder not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "reminder not found"})
			return
		}
		if err.Error() == "access denied" {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}
		if err.Error() == "cannot snooze a reminder that has already been sent" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Получаем стикер для заголовка
	sticker, _ := h.reminderService.GetSticker(c.Request.Context(), reminder.StickerID)
	stickerTitle := ""
	if sticker != nil {
		stickerTitle = sticker.Title
	}

	c.JSON(http.StatusOK, dto.ToReminderResponse(reminder, stickerTitle))
}