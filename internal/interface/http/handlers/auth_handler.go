package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/DmitriiCherkasow/synergyconnect/internal/application"
	"github.com/DmitriiCherkasow/synergyconnect/internal/interfaces/http/dto"
)

// AuthHandler — обработчик для эндпоинтов аутентификации
type AuthHandler struct {
	authService *application.AuthService
}

// NewAuthHandler создает новый обработчик
func NewAuthHandler(authService *application.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Register — обработчик регистрации
// @Summary Регистрация нового пользователя
// @Description Создает нового пользователя с ролью "student"
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequest true "Данные для регистрации"
// @Success 201 {object} dto.AuthResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 409 {object} map[string]interface{}
// @Router /api/v1/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Валидация
	if err := req.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Регистрация
	user, err := h.authService.Register(c.Request.Context(), application.RegisterRequest{
		Email:     req.Email,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	})
	if err != nil {
		if err.Error() == "user with this email already exists" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Генерируем токены для нового пользователя
	loginResp, err := h.authService.Login(c.Request.Context(), application.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate tokens"})
		return
	}

	c.JSON(http.StatusCreated, dto.AuthResponse{
		AccessToken:  loginResp.TokenPair.AccessToken,
		RefreshToken: loginResp.TokenPair.RefreshToken,
		ExpiresIn:    loginResp.TokenPair.ExpiresIn,
		User: dto.UserResponse{
			ID:         user.ID.String(),
			Email:      user.Email,
			Role:       string(user.Role),
			FirstName:  user.FirstName,
			LastName:   user.LastName,
			AvatarURL:  user.AvatarURL,
			IsVerified: user.IsVerified,
		},
	})
}

// Login — обработчик входа
// @Summary Вход пользователя
// @Description Аутентификация и выдача JWT-токенов
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "Данные для входа"
// @Success 200 {object} dto.AuthResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Валидация
	if err := req.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Аутентификация
	resp, err := h.authService.Login(c.Request.Context(), application.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	c.JSON(http.StatusOK, dto.AuthResponse{
		AccessToken:  resp.TokenPair.AccessToken,
		RefreshToken: resp.TokenPair.RefreshToken,
		ExpiresIn:    resp.TokenPair.ExpiresIn,
		User: dto.UserResponse{
			ID:         resp.User.ID.String(),
			Email:      resp.User.Email,
			Role:       string(resp.User.Role),
			FirstName:  resp.User.FirstName,
			LastName:   resp.User.LastName,
			AvatarURL:  resp.User.AvatarURL,
			IsVerified: resp.User.IsVerified,
		},
	})
}

// RefreshToken — обновление access-токена
// @Summary Обновление access-токена
// @Description Использует refresh-токен для получения новой пары токенов
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.RefreshTokenRequest true "Refresh token"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/v1/auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.RefreshToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "refresh token is required"})
		return
	}

	tokenPair, err := h.authService.RefreshTokens(c.Request.Context(), req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  tokenPair.AccessToken,
		"refresh_token": tokenPair.RefreshToken,
		"expires_in":    tokenPair.ExpiresIn,
	})
}