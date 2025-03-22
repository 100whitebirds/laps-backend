package rest

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"laps/internal/domain"
)

// @Summary Получить список образования специалиста
// @Description Возвращает список образования указанного специалиста
// @Tags Образование
// @Accept json
// @Produce json
// @Param specialist_id query int true "ID специалиста"
// @Success 200 {array} domain.Education "Список образования"
// @Failure 400 {object} errorResponseBody "Неверный формат ID"
// @Failure 404 {object} errorResponseBody "Специалист не найден"
// @Failure 500 {object} errorResponseBody "Внутренняя ошибка сервера"
// @Router /education [get]
func (h *Handler) getEducation(c *gin.Context) {
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

	education, err := h.services.Specialist.GetEducationBySpecialistID(c.Request.Context(), specialistID)
	if err != nil {
		h.logger.Error("ошибка при получении образования", zap.Error(err))
		errorResponse(c, http.StatusInternalServerError, "ошибка при получении образования")
		return
	}

	successResponse(c, http.StatusOK, education)
}

// @Summary Добавить образование специалисту
// @Description Добавляет новую запись об образовании для специалиста
// @Tags Образование
// @Accept json
// @Produce json
// @Param specialist_id query int true "ID специалиста"
// @Param input body domain.EducationDTO true "Данные об образовании"
// @Success 201 {object} map[string]interface{} "ID созданной записи об образовании"
// @Failure 400 {object} errorResponseBody "Ошибка валидации"
// @Failure 401 {object} errorResponseBody "Не авторизован"
// @Failure 403 {object} errorResponseBody "Доступ запрещен"
// @Failure 404 {object} errorResponseBody "Специалист не найден"
// @Failure 500 {object} errorResponseBody "Внутренняя ошибка сервера"
// @Security ApiKeyAuth
// @Router /education [post]
func (h *Handler) addEducation(c *gin.Context) {
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

	var req domain.EducationDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("неверный формат данных", zap.Error(err))
		badRequestResponse(c, "неверный формат данных")
		return
	}

	educationID, err := h.services.Specialist.AddEducation(c.Request.Context(), specialistID, req)
	if err != nil {
		h.logger.Error("ошибка при добавлении образования", zap.Error(err))
		errorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	createdResponse(c, map[string]interface{}{
		"id": educationID,
	})
}

// @Summary Получить информацию об образовании по ID
// @Description Возвращает детальную информацию об образовании по его ID
// @Tags Образование
// @Accept json
// @Produce json
// @Param id path int true "ID образования"
// @Success 200 {object} domain.Education "Данные об образовании"
// @Failure 400 {object} errorResponseBody "Неверный формат ID"
// @Failure 404 {object} errorResponseBody "Образование не найдено"
// @Failure 500 {object} errorResponseBody "Внутренняя ошибка сервера"
// @Router /education/{id} [get]
func (h *Handler) getEducationByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		badRequestResponse(c, "неверный формат ID")
		return
	}

	education, err := h.services.Specialist.GetEducationByID(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("ошибка при получении образования", zap.Error(err))
		notFoundResponse(c, "образование не найдено")
		return
	}

	successResponse(c, http.StatusOK, education)
}

// @Summary Обновить информацию об образовании
// @Description Обновляет информацию об образовании
// @Tags Образование
// @Accept json
// @Produce json
// @Param id path int true "ID образования"
// @Param input body domain.EducationDTO true "Новые данные об образовании"
// @Success 204 {object} nil "Данные успешно обновлены"
// @Failure 400 {object} errorResponseBody "Ошибка валидации"
// @Failure 401 {object} errorResponseBody "Не авторизован"
// @Failure 403 {object} errorResponseBody "Доступ запрещен"
// @Failure 404 {object} errorResponseBody "Образование не найдено"
// @Failure 500 {object} errorResponseBody "Внутренняя ошибка сервера"
// @Security ApiKeyAuth
// @Router /education/{id} [put]
func (h *Handler) updateEducation(c *gin.Context) {
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

	education, err := h.services.Specialist.GetEducationByID(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("образование не найдено", zap.Error(err))
		notFoundResponse(c, "образование не найдено")
		return
	}

	specialist, err := h.services.Specialist.GetByID(c.Request.Context(), education.SpecialistID)
	if err != nil {
		h.logger.Error("специалист не найден", zap.Error(err))
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

	var req domain.EducationDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("неверный формат данных", zap.Error(err))
		badRequestResponse(c, "неверный формат данных")
		return
	}

	err = h.services.Specialist.UpdateEducation(c.Request.Context(), id, req)
	if err != nil {
		h.logger.Error("ошибка при обновлении образования", zap.Error(err))
		errorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	noContentResponse(c)
}

// @Summary Удалить образование
// @Description Удаляет информацию об образовании
// @Tags Образование
// @Accept json
// @Produce json
// @Param id path int true "ID образования"
// @Success 204 {object} nil "Образование успешно удалено"
// @Failure 400 {object} errorResponseBody "Неверный формат ID"
// @Failure 401 {object} errorResponseBody "Не авторизован"
// @Failure 403 {object} errorResponseBody "Доступ запрещен"
// @Failure 404 {object} errorResponseBody "Образование не найдено"
// @Failure 500 {object} errorResponseBody "Внутренняя ошибка сервера"
// @Security ApiKeyAuth
// @Router /education/{id} [delete]
func (h *Handler) deleteEducation(c *gin.Context) {
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

	education, err := h.services.Specialist.GetEducationByID(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("образование не найдено", zap.Error(err))
		notFoundResponse(c, "образование не найдено")
		return
	}

	specialist, err := h.services.Specialist.GetByID(c.Request.Context(), education.SpecialistID)
	if err != nil {
		h.logger.Error("специалист не найден", zap.Error(err))
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

	err = h.services.Specialist.DeleteEducation(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("ошибка при удалении образования", zap.Error(err))
		errorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	noContentResponse(c)
}
