package rest

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"laps/internal/domain"
)

// @Summary Получить список опыта работы специалиста
// @Description Возвращает список опыта работы указанного специалиста
// @Tags Опыт работы
// @Accept json
// @Produce json
// @Param specialist_id query int true "ID специалиста"
// @Success 200 {array} domain.WorkPlace "Список опыта работы"
// @Failure 400 {object} errorResponseBody "Неверный формат ID"
// @Failure 404 {object} errorResponseBody "Специалист не найден"
// @Failure 500 {object} errorResponseBody "Внутренняя ошибка сервера"
// @Router /work-experience [get]
func (h *Handler) getWorkExperience(c *gin.Context) {
	specialistIDStr := c.DefaultQuery("specialist_id", "")
	if specialistIDStr == "" {
		badRequestResponse(c, "не указан ID специалиста")
		return
	}

	specialistID, err := strconv.ParseInt(specialistIDStr, 10, 64)
	if err != nil {
		badRequestResponse(c, "неверный формат ID специалиста")
		return
	}

	_, err = h.services.Specialist.GetByID(c.Request.Context(), specialistID)
	if err != nil {
		h.logger.Error("специалист не найден", zap.Int64("id", specialistID), zap.Error(err))
		notFoundResponse(c, "специалист не найден")
		return
	}

	workExperience, err := h.services.WorkExperience.GetWorkExperienceBySpecialistID(c.Request.Context(), specialistID)
	if err != nil {
		h.logger.Error("ошибка при получении опыта работы", zap.Error(err))
		errorResponse(c, http.StatusInternalServerError, "ошибка при получении опыта работы")
		return
	}

	successResponse(c, http.StatusOK, workExperience)
}

// @Summary Добавить опыт работы специалисту
// @Description Добавляет новую запись об опыте работы для специалиста
// @Tags Опыт работы
// @Accept json
// @Produce json
// @Param specialist_id query int true "ID специалиста"
// @Param input body domain.WorkExperienceDTO true "Данные об опыте работы"
// @Success 201 {object} map[string]interface{} "ID созданной записи об опыте работы"
// @Failure 400 {object} errorResponseBody "Ошибка валидации"
// @Failure 401 {object} errorResponseBody "Не авторизован"
// @Failure 403 {object} errorResponseBody "Доступ запрещен"
// @Failure 404 {object} errorResponseBody "Специалист не найден"
// @Failure 500 {object} errorResponseBody "Внутренняя ошибка сервера"
// @Security ApiKeyAuth
// @Router /work-experience [post]
func (h *Handler) addWorkExperience(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		unauthorizedResponse(c)
		return
	}

	specialistIDStr := c.DefaultQuery("specialist_id", "")
	if specialistIDStr == "" {
		badRequestResponse(c, "не указан ID специалиста")
		return
	}

	specialistID, err := strconv.ParseInt(specialistIDStr, 10, 64)
	if err != nil {
		badRequestResponse(c, "неверный формат ID специалиста")
		return
	}

	specialist, err := h.services.Specialist.GetByID(c.Request.Context(), specialistID)
	if err != nil {
		h.logger.Error("специалист не найден", zap.Int64("id", specialistID), zap.Error(err))
		notFoundResponse(c, "специалист не найден")
		return
	}

	userRole, err := getUserRole(c)
	if err != nil {
		unauthorizedResponse(c)
		return
	}

	if specialist.UserID != userID && userRole != domain.UserRoleAdmin {
		forbiddenResponse(c)
		return
	}

	var req domain.WorkExperienceDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("неверный формат данных", zap.Error(err))
		badRequestResponse(c, "неверный формат данных")
		return
	}

	workExperienceID, err := h.services.WorkExperience.AddWorkExperience(c.Request.Context(), specialistID, req)
	if err != nil {
		h.logger.Error("ошибка при добавлении опыта работы", zap.Error(err))
		errorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	createdResponse(c, map[string]interface{}{
		"id": workExperienceID,
	})
}

// @Summary Добавить опыт работы специалисту по ID
// @Description Добавляет новую запись об опыте работы для специалиста
// @Tags Опыт работы
// @Accept json
// @Produce json
// @Param id path int true "ID специалиста"
// @Param input body domain.WorkExperienceDTO true "Данные об опыте работы"
// @Success 201 {object} map[string]interface{} "ID созданной записи об опыте работы"
// @Failure 400 {object} errorResponseBody "Ошибка валидации"
// @Failure 401 {object} errorResponseBody "Не авторизован"
// @Failure 403 {object} errorResponseBody "Доступ запрещен"
// @Failure 404 {object} errorResponseBody "Специалист не найден"
// @Failure 500 {object} errorResponseBody "Внутренняя ошибка сервера"
// @Security ApiKeyAuth
// @Router /specialists/{id}/work-experience [post]
func (h *Handler) addWorkExperienceToSpecialist(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		unauthorizedResponse(c)
		return
	}

	specialistID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		badRequestResponse(c, "неверный формат ID специалиста")
		return
	}

	specialist, err := h.services.Specialist.GetByID(c.Request.Context(), specialistID)
	if err != nil {
		h.logger.Error("специалист не найден", zap.Int64("id", specialistID), zap.Error(err))
		notFoundResponse(c, "специалист не найден")
		return
	}

	userRole, err := getUserRole(c)
	if err != nil {
		unauthorizedResponse(c)
		return
	}

	if specialist.UserID != userID && userRole != domain.UserRoleAdmin {
		forbiddenResponse(c)
		return
	}

	var req domain.WorkExperienceDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("неверный формат данных", zap.Error(err))
		badRequestResponse(c, "неверный формат данных")
		return
	}

	workExperienceID, err := h.services.WorkExperience.AddWorkExperience(c.Request.Context(), specialistID, req)
	if err != nil {
		h.logger.Error("ошибка при добавлении опыта работы", zap.Error(err))
		errorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	createdResponse(c, map[string]interface{}{
		"id": workExperienceID,
	})
}

// @Summary Получить информацию об опыте работы по ID
// @Description Возвращает детальную информацию об опыте работы по его ID
// @Tags Опыт работы
// @Accept json
// @Produce json
// @Param id path int true "ID опыта работы"
// @Success 200 {object} domain.WorkPlace "Данные об опыте работы"
// @Failure 400 {object} errorResponseBody "Неверный формат ID"
// @Failure 404 {object} errorResponseBody "Опыт работы не найден"
// @Failure 500 {object} errorResponseBody "Внутренняя ошибка сервера"
// @Router /work-experience/{id} [get]
func (h *Handler) getWorkExperienceByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		badRequestResponse(c, "неверный формат ID")
		return
	}

	workExperience, err := h.services.WorkExperience.GetWorkExperienceByID(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("ошибка при получении опыта работы", zap.Error(err))
		notFoundResponse(c, "опыт работы не найден")
		return
	}

	successResponse(c, http.StatusOK, workExperience)
}

// @Summary Обновить опыт работы
// @Description Обновляет информацию об опыте работы
// @Tags Опыт работы
// @Accept json
// @Produce json
// @Param id path int true "ID опыта работы"
// @Param input body domain.WorkExperienceDTO true "Новые данные об опыте работы"
// @Success 200 {object} messageResponseType "Опыт работы успешно обновлен"
// @Failure 400 {object} errorResponseBody "Ошибка валидации"
// @Failure 401 {object} errorResponseBody "Не авторизован"
// @Failure 403 {object} errorResponseBody "Доступ запрещен"
// @Failure 404 {object} errorResponseBody "Опыт работы не найден"
// @Failure 500 {object} errorResponseBody "Внутренняя ошибка сервера"
// @Security ApiKeyAuth
// @Router /work-experience/{id} [put]
func (h *Handler) updateWorkExperience(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		unauthorizedResponse(c)
		return
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		badRequestResponse(c, "неверный формат ID")
		return
	}

	workExperience, err := h.services.WorkExperience.GetWorkExperienceByID(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("опыт работы не найден", zap.Int64("id", id), zap.Error(err))
		notFoundResponse(c, "опыт работы не найден")
		return
	}

	specialist, err := h.services.Specialist.GetByID(c.Request.Context(), workExperience.SpecialistID)
	if err != nil {
		h.logger.Error("специалист не найден", zap.Int64("id", workExperience.SpecialistID), zap.Error(err))
		notFoundResponse(c, "специалист не найден")
		return
	}

	userRole, err := getUserRole(c)
	if err != nil {
		unauthorizedResponse(c)
		return
	}

	if specialist.UserID != userID && userRole != domain.UserRoleAdmin {
		forbiddenResponse(c)
		return
	}

	var req domain.WorkExperienceDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("неверный формат данных", zap.Error(err))
		badRequestResponse(c, "неверный формат данных")
		return
	}

	err = h.services.WorkExperience.UpdateWorkExperience(c.Request.Context(), id, req)
	if err != nil {
		h.logger.Error("ошибка при обновлении опыта работы", zap.Error(err))
		errorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	messageResponse(c, http.StatusOK, "опыт работы успешно обновлен")
}

// @Summary Удалить опыт работы
// @Description Удаляет запись об опыте работы
// @Tags Опыт работы
// @Accept json
// @Produce json
// @Param id path int true "ID опыта работы"
// @Success 204 {object} nil "Опыт работы успешно удален"
// @Failure 400 {object} errorResponseBody "Ошибка валидации"
// @Failure 401 {object} errorResponseBody "Не авторизован"
// @Failure 403 {object} errorResponseBody "Доступ запрещен"
// @Failure 404 {object} errorResponseBody "Опыт работы не найден"
// @Failure 500 {object} errorResponseBody "Внутренняя ошибка сервера"
// @Security ApiKeyAuth
// @Router /work-experience/{id} [delete]
func (h *Handler) deleteWorkExperience(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		unauthorizedResponse(c)
		return
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		badRequestResponse(c, "неверный формат ID")
		return
	}

	workExperience, err := h.services.WorkExperience.GetWorkExperienceByID(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("опыт работы не найден", zap.Int64("id", id), zap.Error(err))
		notFoundResponse(c, "опыт работы не найден")
		return
	}

	specialist, err := h.services.Specialist.GetByID(c.Request.Context(), workExperience.SpecialistID)
	if err != nil {
		h.logger.Error("специалист не найден", zap.Int64("id", workExperience.SpecialistID), zap.Error(err))
		notFoundResponse(c, "специалист не найден")
		return
	}

	userRole, err := getUserRole(c)
	if err != nil {
		unauthorizedResponse(c)
		return
	}

	if specialist.UserID != userID && userRole != domain.UserRoleAdmin {
		forbiddenResponse(c)
		return
	}

	err = h.services.WorkExperience.DeleteWorkExperience(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("ошибка при удалении опыта работы", zap.Error(err))
		errorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	noContentResponse(c)
}
