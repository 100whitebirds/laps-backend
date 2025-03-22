package rest

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"laps/internal/domain"
)

// @Summary Создать расписание
// @Description Создает новое расписание для специалиста
// @Tags Расписание
// @Accept json
// @Produce json
// @Param input body domain.CreateScheduleDTO true "Данные для создания расписания"
// @Success 201 {object} map[string]interface{} "ID созданного расписания"
// @Failure 400 {object} errorResponseBody "Ошибка валидации данных"
// @Failure 401 {object} errorResponseBody "Не авторизован"
// @Failure 403 {object} errorResponseBody "Доступ запрещен"
// @Failure 500 {object} errorResponseBody "Внутренняя ошибка сервера"
// @Security ApiKeyAuth
// @Router /schedules [post]
func (h *Handler) createSchedule(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		unauthorizedResponse(c)
		return
	}

	specialist, err := h.services.Specialist.GetByUserID(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("ошибка при получении данных специалиста", zap.Error(err))
		notFoundResponse(c, "профиль специалиста не найден")
		return
	}

	var req domain.CreateScheduleDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("неверный формат данных", zap.Error(err))
		badRequestResponse(c, "неверный формат данных")
		return
	}

	_, err = time.Parse("2006-01-02", req.Date)
	if err != nil {
		badRequestResponse(c, "неверный формат даты, ожидается YYYY-MM-DD")
		return
	}

	_, err = time.Parse("15:04", req.StartTime)
	if err != nil {
		badRequestResponse(c, "неверный формат времени начала, ожидается HH:MM")
		return
	}

	_, err = time.Parse("15:04", req.EndTime)
	if err != nil {
		badRequestResponse(c, "неверный формат времени окончания, ожидается HH:MM")
		return
	}

	if req.SlotTime < 10 || req.SlotTime > 120 {
		badRequestResponse(c, "длительность слота должна быть от 10 до 120 минут")
		return
	}

	scheduleID, err := h.services.Schedule.Create(c.Request.Context(), specialist.ID, req)
	if err != nil {
		h.logger.Error("ошибка создания расписания", zap.Error(err))
		errorResponse(c, http.StatusInternalServerError, "ошибка создания расписания")
		return
	}

	createdResponse(c, gin.H{"id": scheduleID})
}

// @Summary Получить расписание по ID
// @Description Получает расписание по ID
// @Tags Расписание
// @Produce json
// @Param id path int true "ID расписания"
// @Success 200 {object} domain.Schedule "Расписание"
// @Failure 400 {object} errorResponseBody "Ошибка валидации данных"
// @Failure 401 {object} errorResponseBody "Не авторизован"
// @Failure 404 {object} errorResponseBody "Расписание не найдено"
// @Failure 500 {object} errorResponseBody "Внутренняя ошибка сервера"
// @Security ApiKeyAuth
// @Router /schedules/{id} [get]
func (h *Handler) getScheduleByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		badRequestResponse(c, "неверный формат ID")
		return
	}

	schedule, err := h.services.Schedule.GetByID(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("ошибка получения расписания", zap.Error(err))
		errorResponse(c, http.StatusInternalServerError, "ошибка получения расписания")
		return
	}

	if schedule == nil {
		notFoundResponse(c, "расписание не найдено")
		return
	}

	successResponse(c, http.StatusOK, schedule)
}

// @Summary Обновить расписание
// @Description Обновляет существующее расписание
// @Tags Расписание
// @Accept json
// @Produce json
// @Param id path int true "ID расписания"
// @Param input body domain.UpdateScheduleDTO true "Данные для обновления расписания"
// @Success 200 {object} messageResponseType "Сообщение об успешном обновлении"
// @Failure 400 {object} errorResponseBody "Ошибка валидации данных"
// @Failure 401 {object} errorResponseBody "Не авторизован"
// @Failure 403 {object} errorResponseBody "Доступ запрещен"
// @Failure 404 {object} errorResponseBody "Расписание не найдено"
// @Failure 500 {object} errorResponseBody "Внутренняя ошибка сервера"
// @Security ApiKeyAuth
// @Router /schedules/{id} [put]
func (h *Handler) updateSchedule(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		badRequestResponse(c, "неверный формат ID")
		return
	}

	var req domain.UpdateScheduleDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequestResponse(c, "неверный формат данных")
		return
	}

	userID, err := getUserID(c)
	if err != nil {
		unauthorizedResponse(c)
		return
	}

	specialist, err := h.services.Specialist.GetByUserID(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("ошибка при получении данных специалиста", zap.Error(err))
		notFoundResponse(c, "профиль специалиста не найден")
		return
	}

	schedule, err := h.services.Schedule.GetByID(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("ошибка получения расписания", zap.Error(err))
		errorResponse(c, http.StatusInternalServerError, "ошибка получения расписания")
		return
	}

	if schedule == nil {
		notFoundResponse(c, "расписание не найдено")
		return
	}

	if schedule.SpecialistID != specialist.ID {
		forbiddenResponse(c, "нет доступа к данному расписанию")
		return
	}

	err = h.services.Schedule.Update(c.Request.Context(), id, req)
	if err != nil {
		h.logger.Error("ошибка обновления расписания", zap.Error(err))
		errorResponse(c, http.StatusInternalServerError, "ошибка обновления расписания")
		return
	}

	messageResponse(c, http.StatusOK, "расписание успешно обновлено")
}

// @Summary Удалить расписание
// @Description Удаляет существующее расписание
// @Tags Расписание
// @Produce json
// @Param id path int true "ID расписания"
// @Success 200 {object} messageResponseType "Сообщение об успешном удалении"
// @Failure 400 {object} errorResponseBody "Ошибка валидации данных"
// @Failure 401 {object} errorResponseBody "Не авторизован"
// @Failure 403 {object} errorResponseBody "Доступ запрещен"
// @Failure 404 {object} errorResponseBody "Расписание не найдено"
// @Failure 500 {object} errorResponseBody "Внутренняя ошибка сервера"
// @Security ApiKeyAuth
// @Router /schedules/{id} [delete]
func (h *Handler) deleteSchedule(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		badRequestResponse(c, "неверный формат ID")
		return
	}

	userID, err := getUserID(c)
	if err != nil {
		unauthorizedResponse(c)
		return
	}

	specialist, err := h.services.Specialist.GetByUserID(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("ошибка при получении данных специалиста", zap.Error(err))
		notFoundResponse(c, "профиль специалиста не найден")
		return
	}

	schedule, err := h.services.Schedule.GetByID(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("ошибка получения расписания", zap.Error(err))
		errorResponse(c, http.StatusInternalServerError, "ошибка получения расписания")
		return
	}

	if schedule == nil {
		notFoundResponse(c, "расписание не найдено")
		return
	}

	if schedule.SpecialistID != specialist.ID {
		forbiddenResponse(c, "нет доступа к данному расписанию")
		return
	}

	err = h.services.Schedule.Delete(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("ошибка удаления расписания", zap.Error(err))
		errorResponse(c, http.StatusInternalServerError, "ошибка удаления расписания")
		return
	}

	messageResponse(c, http.StatusOK, "расписание успешно удалено")
}

// @Summary Получить список расписаний
// @Description Возвращает список расписаний с поддержкой фильтрации
// @Tags Расписание
// @Produce json
// @Param specialist_id query int false "ID специалиста"
// @Param date_from query string false "Начальная дата (YYYY-MM-DD)"
// @Param date_to query string false "Конечная дата (YYYY-MM-DD)"
// @Param limit query int false "Лимит (по умолчанию 20)"
// @Param offset query int false "Смещение (по умолчанию 0)"
// @Success 200 {object} paginatedResponse "Список расписаний с пагинацией"
// @Failure 400 {object} errorResponseBody "Ошибка валидации данных"
// @Failure 401 {object} errorResponseBody "Не авторизован"
// @Failure 500 {object} errorResponseBody "Внутренняя ошибка сервера"
// @Security ApiKeyAuth
// @Router /schedules [get]
func (h *Handler) getSchedules(c *gin.Context) {
	specialistIDStr := c.DefaultQuery("specialist_id", "")
	var specialistID *int64
	if specialistIDStr != "" {
		id, err := strconv.ParseInt(specialistIDStr, 10, 64)
		if err == nil {
			specialistID = &id
		}
	}

	dateFrom := c.DefaultQuery("date_from", "")
	var startDate *time.Time
	if dateFrom != "" {
		parsedDate, err := time.Parse("2006-01-02", dateFrom)
		if err == nil {
			startDate = &parsedDate
		}
	}

	dateTo := c.DefaultQuery("date_to", "")
	var endDate *time.Time
	if dateTo != "" {
		parsedDate, err := time.Parse("2006-01-02", dateTo)
		if err == nil {
			parsedDate = parsedDate.Add(24 * time.Hour).Add(-time.Second)
			endDate = &parsedDate
		}
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if err != nil || limit < 0 {
		limit = 20
	}

	offset, err := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if err != nil || offset < 0 {
		offset = 0
	}

	filter := domain.ScheduleFilter{
		SpecialistID: specialistID,
		StartDate:    startDate,
		EndDate:      endDate,
		Limit:        limit,
		Offset:       offset,
	}

	schedules, total, err := h.services.Schedule.List(c.Request.Context(), filter)
	if err != nil {
		h.logger.Error("ошибка получения списка расписаний", zap.Error(err))
		errorResponse(c, http.StatusInternalServerError, "ошибка получения списка расписаний")
		return
	}

	page := offset/limit + 1

	paginatedSuccessResponse(c, schedules, total, page, limit)
}

// @Summary Получить свободные слоты специалиста
// @Description Возвращает список свободных временных слотов на выбранную дату
// @Tags Расписание
// @Produce json
// @Param specialist_id query int true "ID специалиста"
// @Param date query string true "Дата (YYYY-MM-DD)"
// @Success 200 {object} map[string]interface{} "Список свободных слотов"
// @Failure 400 {object} errorResponseBody "Ошибка валидации данных"
// @Failure 500 {object} errorResponseBody "Внутренняя ошибка сервера"
// @Router /schedules/free-slots [get]
func (h *Handler) getFreeSlots(c *gin.Context) {
	specialistIDStr := c.Query("specialist_id")
	date := c.Query("date")

	if specialistIDStr == "" || date == "" {
		badRequestResponse(c, "необходимо указать ID специалиста и дату")
		return
	}

	specialistID, err := strconv.ParseInt(specialistIDStr, 10, 64)
	if err != nil {
		badRequestResponse(c, "неверный формат ID специалиста")
		return
	}

	_, err = time.Parse("2006-01-02", date)
	if err != nil {
		badRequestResponse(c, "неверный формат даты, ожидается YYYY-MM-DD")
		return
	}

	slots, err := h.services.Schedule.GenerateTimeSlots(c.Request.Context(), specialistID, date)
	if err != nil {
		h.logger.Error("ошибка получения свободных слотов", zap.Error(err))
		errorResponse(c, http.StatusInternalServerError, "ошибка получения свободных слотов")
		return
	}

	// Фильтруем занятые слоты (можно использовать сервис записей на прием)
	// TODO: добавить фильтрацию занятых слотов через AppointmentService

	successResponse(c, http.StatusOK, gin.H{
		"specialist_id": specialistID,
		"date":          date,
		"free_slots":    slots,
	})
}
