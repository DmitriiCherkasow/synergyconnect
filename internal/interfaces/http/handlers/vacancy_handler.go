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

// VacancyHandler — обработчик для вакансий
type VacancyHandler struct {
	vacancyService *application.VacancyService
}

// NewVacancyHandler создает новый обработчик
func NewVacancyHandler(vacancyService *application.VacancyService) *VacancyHandler {
	return &VacancyHandler{
		vacancyService: vacancyService,
	}
}

// getUserID извлекает ID пользователя из контекста
func (h *VacancyHandler) getUserID(c *gin.Context) (uuid.UUID, error) {
	userIDStr := middleware.GetUserIDFromContext(c)
	if userIDStr == "" {
		return uuid.Nil, nil
	}
	return uuid.Parse(userIDStr)
}

// CreateVacancy — создание вакансии
// @Summary Создание вакансии
// @Tags vacancies
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.CreateVacancyRequest true "Данные для создания вакансии"
// @Success 201 {object} dto.VacancyResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /vacancies [post]
func (h *VacancyHandler) CreateVacancy(c *gin.Context) {
	userID, err := h.getUserID(c)
	if err != nil || userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req dto.CreateVacancyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := req.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	vacancy, err := h.vacancyService.CreateVacancy(c.Request.Context(), userID, application.CreateVacancyRequest{
		Title:           req.Title,
		Company:         req.Company,
		Description:     req.Description,
		Requirements:    req.Requirements,
		SalaryMin:       req.SalaryMin,
		SalaryMax:       req.SalaryMax,
		Currency:        req.Currency,
		Location:        req.Location,
		IsRemote:        req.IsRemote,
		EmploymentType:  domain.EmploymentType(req.EmploymentType),
		ExperienceLevel: domain.ExperienceLevel(req.ExperienceLevel),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, dto.ToVacancyResponse(vacancy))
}

// GetVacancy — получение вакансии по ID
func (h *VacancyHandler) GetVacancy(c *gin.Context) {
	userID, err := h.getUserID(c)
	if err != nil || userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid vacancy id"})
		return
	}

	// Увеличиваем счетчик просмотров
	_ = h.vacancyService.IncrementViews(c.Request.Context(), id)

	vacancy, err := h.vacancyService.GetVacancyByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if vacancy == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "vacancy not found"})
		return
	}

	response := dto.VacancyDetailResponse{
		VacancyResponse: dto.ToVacancyResponse(vacancy),
		CanManage:       vacancy.EmployerID == userID,
	}

	// Проверяем, откликался ли пользователь
	for _, resp := range vacancy.Responses {
		if resp.UserID == userID {
			response.HasResponded = true
			break
		}
	}

	// Если пользователь - работодатель, показываем отклики
	if response.CanManage {
		responses, err := h.vacancyService.GetResponsesByVacancy(c.Request.Context(), id, userID)
		if err == nil {
			response.Responses = make([]dto.VacancyResponseItem, len(responses))
			for i, r := range responses {
				item := dto.ToVacancyResponseItem(&r)
				if r.User.ID != uuid.Nil {
					item.User = dto.UserResponse{
						ID:         r.User.ID.String(),
						Email:      r.User.Email,
						Role:       string(r.User.Role),
						FirstName:  r.User.FirstName,
						LastName:   r.User.LastName,
						AvatarURL:  r.User.AvatarURL,
						IsVerified: r.User.IsVerified,
					}
				}
				response.Responses[i] = item
			}
		}
	}

	c.JSON(http.StatusOK, response)
}

// UpdateVacancy — обновление вакансии
func (h *VacancyHandler) UpdateVacancy(c *gin.Context) {
	userID, err := h.getUserID(c)
	if err != nil || userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid vacancy id"})
		return
	}

	var req dto.UpdateVacancyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := req.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updateReq := application.UpdateVacancyRequest{
		Title:        req.Title,
		Company:      req.Company,
		Description:  req.Description,
		Requirements: req.Requirements,
		SalaryMin:    req.SalaryMin,
		SalaryMax:    req.SalaryMax,
		Currency:     req.Currency,
		Location:     req.Location,
		IsRemote:     req.IsRemote,
	}

	if req.EmploymentType != nil {
		et := domain.EmploymentType(*req.EmploymentType)
		updateReq.EmploymentType = &et
	}
	if req.ExperienceLevel != nil {
		el := domain.ExperienceLevel(*req.ExperienceLevel)
		updateReq.ExperienceLevel = &el
	}
	if req.Status != nil {
		s := domain.VacancyStatus(*req.Status)
		updateReq.Status = &s
	}

	vacancy, err := h.vacancyService.UpdateVacancy(c.Request.Context(), id, userID, updateReq)
	if err != nil {
		if err == domain.ErrVacancyNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if err == domain.ErrForbidden {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.ToVacancyResponse(vacancy))
}

// DeleteVacancy — удаление вакансии
func (h *VacancyHandler) DeleteVacancy(c *gin.Context) {
	userID, err := h.getUserID(c)
	if err != nil || userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid vacancy id"})
		return
	}

	err = h.vacancyService.DeleteVacancy(c.Request.Context(), id, userID)
	if err != nil {
		if err == domain.ErrVacancyNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if err == domain.ErrForbidden {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// ListVacancies — список вакансий
func (h *VacancyHandler) ListVacancies(c *gin.Context) {
	userID, err := h.getUserID(c)
	if err != nil || userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	req := application.ListVacanciesRequest{
		Limit:  20,
		Offset: 0,
	}

	if company := c.Query("company"); company != "" {
		req.Company = &company
	}
	if location := c.Query("location"); location != "" {
		req.Location = &location
	}
	if isRemote := c.Query("is_remote"); isRemote != "" {
		remote := isRemote == "true"
		req.IsRemote = &remote
	}
	if employmentType := c.Query("employment_type"); employmentType != "" {
		et := domain.EmploymentType(employmentType)
		req.EmploymentType = &et
	}
	if experienceLevel := c.Query("experience_level"); experienceLevel != "" {
		el := domain.ExperienceLevel(experienceLevel)
		req.ExperienceLevel = &el
	}
	if status := c.Query("status"); status != "" {
		s := domain.VacancyStatus(status)
		req.Status = &s
	}
	req.Search = c.Query("search")

	if limit := c.Query("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil && l > 0 {
			req.Limit = l
		}
	}
	if offset := c.Query("offset"); offset != "" {
		if o, err := strconv.Atoi(offset); err == nil && o >= 0 {
			req.Offset = o
		}
	}

	vacancies, total, err := h.vacancyService.ListVacancies(c.Request.Context(), req)
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

// SearchVacancies — поиск вакансий
func (h *VacancyHandler) SearchVacancies(c *gin.Context) {
	userID, err := h.getUserID(c)
	if err != nil || userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "search query is required"})
		return
	}

	limit := 20
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	offset := 0
	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	vacancies, total, err := h.vacancyService.SearchVacancies(c.Request.Context(), query, limit, offset)
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

// CreateResponse — отклик на вакансию
func (h *VacancyHandler) CreateResponse(c *gin.Context) {
	userID, err := h.getUserID(c)
	if err != nil || userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	vacancyID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid vacancy id"})
		return
	}

	var req dto.CreateResponseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.vacancyService.CreateResponse(c.Request.Context(), vacancyID, userID, application.CreateResponseRequest{
		CoverLetter: req.CoverLetter,
	})
	if err != nil {
		if err == domain.ErrVacancyNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if err == domain.ErrResponseExists {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		if err == domain.ErrVacancyClosed {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "response created successfully"})
}

// UpdateResponse — изменение статуса отклика
func (h *VacancyHandler) UpdateResponse(c *gin.Context) {
	userID, err := h.getUserID(c)
	if err != nil || userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	responseID, err := uuid.Parse(c.Param("responseId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid response id"})
		return
	}

	var req dto.UpdateResponseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := req.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	status := domain.VacancyResponseStatus(req.Status)
	err = h.vacancyService.UpdateResponseStatus(c.Request.Context(), responseID, userID, status)
	if err != nil {
		if err == domain.ErrResponseNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if err == domain.ErrVacancyNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if err == domain.ErrForbidden {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "response status updated"})
}

// GetMyResponses — мои отклики
func (h *VacancyHandler) GetMyResponses(c *gin.Context) {
	userID, err := h.getUserID(c)
	if err != nil || userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	responses, err := h.vacancyService.GetUserResponses(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	result := make([]dto.VacancyResponseItem, len(responses))
	for i, r := range responses {
		item := dto.ToVacancyResponseItem(&r)
		if r.Vacancy.ID != uuid.Nil {
			item.VacancyID = r.Vacancy.ID.String()
			item.VacancyTitle = r.Vacancy.Title
		}
		if r.User.ID != uuid.Nil {
			item.User = dto.UserResponse{
				ID:         r.User.ID.String(),
				Email:      r.User.Email,
				Role:       string(r.User.Role),
				FirstName:  r.User.FirstName,
				LastName:   r.User.LastName,
				AvatarURL:  r.User.AvatarURL,
				IsVerified: r.User.IsVerified,
			}
		}
		result[i] = item
	}

	c.JSON(http.StatusOK, result)
}