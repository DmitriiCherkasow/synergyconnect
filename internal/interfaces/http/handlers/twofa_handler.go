package handlers

import (
	"encoding/base64"
	"net/http"

	"github.com/DmitriiCherkasow/synergyconnect.git/internal/application"
	"github.com/DmitriiCherkasow/synergyconnect.git/internal/domain"
	"github.com/DmitriiCherkasow/synergyconnect.git/internal/interfaces/http/dto"
	"github.com/DmitriiCherkasow/synergyconnect.git/internal/interfaces/http/middleware"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// TwoFAHandler — обработчик для 2FA
type TwoFAHandler struct {
	twofaService *application.TwoFAService
}

// NewTwoFAHandler создает новый обработчик
func NewTwoFAHandler(twofaService *application.TwoFAService) *TwoFAHandler {
	return &TwoFAHandler{
		twofaService: twofaService,
	}
}

// getUserID извлекает ID пользователя из контекста
func (h *TwoFAHandler) getUserID(c *gin.Context) (uuid.UUID, error) {
	userIDStr := middleware.GetUserIDFromContext(c)
	if userIDStr == "" {
		return uuid.Nil, nil
	}
	return uuid.Parse(userIDStr)
}

// Initiate2FA — инициализация 2FA
// @Summary Инициализация 2FA
// @Description Генерирует секрет и QR-код для Google Authenticator
// @Tags 2fa
// @Security BearerAuth
// @Produce json
// @Success 200 {object} dto.Initiate2FAResponse
// @Router /2fa/initiate [post]
func (h *TwoFAHandler) Initiate2FA(c *gin.Context) {
	userID, err := h.getUserID(c)
	if err != nil || userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	resp, err := h.twofaService.Initiate2FA(c.Request.Context(), userID)
	if err != nil {
		if err == domain.ErrTwoFAAlreadyEnabled {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		if err == domain.ErrUserNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Конвертируем QR-код в base64
	qrBase64 := base64.StdEncoding.EncodeToString(resp.QRCode)

	c.JSON(http.StatusOK, dto.Initiate2FAResponse{
		Secret:        resp.Secret,
		QRCode:        qrBase64,
		RecoveryCodes: resp.RecoveryCodes,
	})
}

// Enable2FA — включение 2FA
// @Summary Включение 2FA
// @Description Подтверждает код и включает 2FA
// @Tags 2fa
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.Enable2FARequest true "Код подтверждения"
// @Success 200 {object} map[string]string
// @Router /2fa/enable [post]
func (h *TwoFAHandler) Enable2FA(c *gin.Context) {
	userID, err := h.getUserID(c)
	if err != nil || userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req dto.Enable2FARequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.twofaService.VerifyAndEnable2FA(c.Request.Context(), userID, req.Code)
	if err != nil {
		if err == domain.ErrTwoFAAlreadyEnabled {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		if err == domain.ErrInvalidCode {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid code"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "2FA enabled successfully"})
}

// Disable2FA — отключение 2FA
// @Summary Отключение 2FA
// @Description Отключает 2FA после проверки кода
// @Tags 2fa
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.Disable2FARequest true "Код подтверждения"
// @Success 200 {object} map[string]string
// @Router /2fa/disable [post]
func (h *TwoFAHandler) Disable2FA(c *gin.Context) {
	userID, err := h.getUserID(c)
	if err != nil || userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req dto.Disable2FARequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.twofaService.Disable2FA(c.Request.Context(), userID, req.Code)
	if err != nil {
		if err == domain.ErrTwoFANotEnabled {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err == domain.ErrInvalidCode {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid code"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "2FA disabled successfully"})
}

// VerifyCode — проверка кода
// @Summary Проверка TOTP кода
// @Description Проверяет правильность кода без изменения статуса
// @Tags 2fa
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.Verify2FARequest true "Код для проверки"
// @Success 200 {object} map[string]bool
// @Router /2fa/verify [post]
func (h *TwoFAHandler) VerifyCode(c *gin.Context) {
	userID, err := h.getUserID(c)
	if err != nil || userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req dto.Verify2FARequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	valid, err := h.twofaService.VerifyCode(c.Request.Context(), userID, req.Code)
	if err != nil {
		if err == domain.ErrTwoFANotEnabled {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"valid": valid})
}

// GetRecoveryCodes — получение кодов восстановления
// @Summary Получение кодов восстановления
// @Description Возвращает коды восстановления
// @Tags 2fa
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string][]string
// @Router /2fa/recovery-codes [get]
func (h *TwoFAHandler) GetRecoveryCodes(c *gin.Context) {
	userID, err := h.getUserID(c)
	if err != nil || userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	codes, err := h.twofaService.GetRecoveryCodes(c.Request.Context(), userID)
	if err != nil {
		if err == domain.ErrTwoFANotEnabled {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"recovery_codes": codes})
}

// RegenerateRecoveryCodes — перегенерация кодов восстановления
// @Summary Перегенерация кодов восстановления
// @Description Генерирует новые коды восстановления
// @Tags 2fa
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.RegenerateCodesRequest true "Код подтверждения"
// @Success 200 {object} map[string][]string
// @Router /2fa/recovery-codes/regenerate [post]
func (h *TwoFAHandler) RegenerateRecoveryCodes(c *gin.Context) {
	userID, err := h.getUserID(c)
	if err != nil || userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req dto.RegenerateCodesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	codes, err := h.twofaService.RegenerateRecoveryCodes(c.Request.Context(), userID, req.Code)
	if err != nil {
		if err == domain.ErrInvalidCode {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid code"})
			return
		}
		if err == domain.ErrTwoFANotEnabled {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"recovery_codes": codes})
}

// Status — статус 2FA
// @Summary Статус 2FA
// @Description Проверяет, включена ли 2FA для пользователя
// @Tags 2fa
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]bool
// @Router /2fa/status [get]
func (h *TwoFAHandler) Status(c *gin.Context) {
	userID, err := h.getUserID(c)
	if err != nil || userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	enabled, err := h.twofaService.IsEnabled(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"enabled": enabled})
}