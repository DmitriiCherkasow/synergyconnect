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

// GetUsers — список пользователей
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

// GetStats — статистика
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