package rest

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"laps/internal/domain"
)

// @Summary Получить отзыв по ID
// @Description Возвращает информацию об отзыве по указанному ID
// @Tags Отзывы
// @Accept json
// @Produce json
// @Param id path int true "ID отзыва"
// @Success 200 {object} domain.Review "Данные отзыва"
// @Failure 400 {object} errorResponseBody "Неверный формат ID"
// @Failure 404 {object} errorResponseBody "Отзыв не найден"
// @Failure 500 {object} errorResponseBody "Внутренняя ошибка сервера"
// @Router /reviews/{id} [get]
func (h *Handler) getReviewByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		h.logger.Warn("неверный формат ID", zap.Error(err))
		badRequestResponse(c, "неверный формат ID")
		return
	}

	review, err := h.services.Review.GetByID(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("ошибка получения отзыва", zap.Error(err), zap.Int64("id", id))
		notFoundResponse(c, "отзыв не найден")
		return
	}

	successResponse(c, http.StatusOK, review)
}

// @Summary Создать отзыв
// @Description Создает новый отзыв о специалисте
// @Tags Отзывы
// @Accept json
// @Produce json
// @Param input body domain.CreateReviewDTO true "Данные отзыва, включая рейтинги по различным критериям"
// @Success 201 {object} map[string]interface{} "ID созданного отзыва"
// @Failure 400 {object} errorResponseBody "Ошибка валидации"
// @Failure 401 {object} errorResponseBody "Не авторизован"
// @Failure 403 {object} errorResponseBody "Доступ запрещен"
// @Failure 404 {object} errorResponseBody "Специалист не найден"
// @Failure 500 {object} errorResponseBody "Внутренняя ошибка сервера"
// @Security ApiKeyAuth
// @Router /reviews [post]
func (h *Handler) createReview(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		unauthorizedResponse(c)
		return
	}

	userRole, err := getUserRole(c)
	if err != nil {
		unauthorizedResponse(c)
		return
	}

	if userRole != domain.UserRoleClient && userRole != domain.UserRoleAdmin {
		forbiddenResponse(c)
		return
	}

	var req domain.CreateReviewDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("неверный формат данных", zap.Error(err))
		badRequestResponse(c, "неверный формат данных")
		return
	}

	h.logger.Info("Получены данные для создания отзыва",
		zap.Int64("specialist_id", req.SpecialistID),
		zap.Int64("appointment_id", req.AppointmentID),
		zap.Int("rating", req.Rating),
		zap.Bool("is_recommended", req.IsRecommended),
		zap.Any("service_rating", req.ServiceRating),
		zap.Any("meeting_efficiency", req.MeetingEfficiency),
		zap.Any("professionalism", req.Professionalism),
		zap.Any("price_quality", req.PriceQuality),
		zap.Any("cleanliness", req.Cleanliness),
		zap.Any("attentiveness", req.Attentiveness),
		zap.Any("specialist_experience", req.SpecialistExperience),
		zap.Any("grammar", req.Grammar))

	id, err := h.services.Review.Create(c.Request.Context(), userID, req)
	if err != nil {
		h.logger.Error("ошибка при создании отзыва", zap.Error(err))
		errorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	createdResponse(c, map[string]interface{}{
		"id": id,
	})
}

// @Summary Удалить отзыв
// @Description Удаляет отзыв (только автор или администратор)
// @Tags Отзывы
// @Accept json
// @Produce json
// @Param id path int true "ID отзыва"
// @Success 204 {object} nil "Отзыв успешно удален"
// @Failure 400 {object} errorResponseBody "Неверный формат ID"
// @Failure 401 {object} errorResponseBody "Не авторизован"
// @Failure 403 {object} errorResponseBody "Доступ запрещен"
// @Failure 404 {object} errorResponseBody "Отзыв не найден"
// @Failure 500 {object} errorResponseBody "Внутренняя ошибка сервера"
// @Security ApiKeyAuth
// @Router /reviews/{id} [delete]
func (h *Handler) deleteReview(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		h.logger.Warn("ошибка получения ID пользователя", zap.Error(err))
		unauthorizedResponse(c)
		return
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		h.logger.Warn("неверный формат ID", zap.Error(err))
		badRequestResponse(c, "неверный формат ID")
		return
	}

	review, err := h.services.Review.GetByID(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("ошибка получения отзыва", zap.Error(err), zap.Int64("id", id))
		notFoundResponse(c, "отзыв не найден")
		return
	}

	userRole, _ := getUserRole(c)
	if review.ClientID != userID && userRole != domain.UserRoleAdmin {
		h.logger.Warn("попытка несанкционированного доступа", zap.Int64("userID", userID))
		forbiddenResponse(c)
		return
	}

	err = h.services.Review.Delete(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("ошибка удаления отзыва", zap.Error(err))
		internalServerErrorResponse(c)
		return
	}

	noContentResponse(c)
}

// @Summary Добавить ответ на отзыв
// @Description Добавляет ответ специалиста на отзыв (только специалист, о котором отзыв)
// @Tags Отзывы
// @Accept json
// @Produce json
// @Param id path int true "ID отзыва"
// @Param input body domain.CreateReplyDTO true "Текст ответа"
// @Success 201 {object} map[string]interface{} "ID созданного ответа"
// @Failure 400 {object} errorResponseBody "Ошибка валидации"
// @Failure 401 {object} errorResponseBody "Не авторизован"
// @Failure 403 {object} errorResponseBody "Доступ запрещен"
// @Failure 404 {object} errorResponseBody "Отзыв не найден"
// @Failure 500 {object} errorResponseBody "Внутренняя ошибка сервера"
// @Security ApiKeyAuth
// @Router /reviews/{id}/replies [post]
func (h *Handler) createReviewReply(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		h.logger.Warn("ошибка получения ID пользователя", zap.Error(err))
		unauthorizedResponse(c)
		return
	}

	reviewID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		h.logger.Warn("неверный формат ID отзыва", zap.Error(err))
		badRequestResponse(c, "неверный формат ID отзыва")
		return
	}

	var req domain.CreateReplyDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("неверный формат данных", zap.Error(err))
		badRequestResponse(c, "неверный формат данных")
		return
	}

	id, err := h.services.Review.CreateReply(c.Request.Context(), userID, reviewID, req)
	if err != nil {
		h.logger.Error("ошибка создания ответа на отзыв", zap.Error(err))
		badRequestResponse(c, err.Error())
		return
	}

	createdResponse(c, gin.H{"id": id})
}

// @Summary Удалить ответ на отзыв
// @Description Удаляет ответ на отзыв (только администратор)
// @Tags Отзывы
// @Accept json
// @Produce json
// @Param replyId path int true "ID ответа"
// @Success 204 {object} nil "Ответ успешно удален"
// @Failure 400 {object} errorResponseBody "Неверный формат ID"
// @Failure 401 {object} errorResponseBody "Не авторизован"
// @Failure 403 {object} errorResponseBody "Доступ запрещен"
// @Failure 500 {object} errorResponseBody "Внутренняя ошибка сервера"
// @Security ApiKeyAuth
// @Router /reviews/replies/{replyId} [delete]
func (h *Handler) deleteReviewReply(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		h.logger.Warn("ошибка получения ID пользователя", zap.Error(err))
		unauthorizedResponse(c)
		return
	}

	replyID, err := strconv.ParseInt(c.Param("replyId"), 10, 64)
	if err != nil {
		h.logger.Warn("неверный формат ID ответа", zap.Error(err))
		badRequestResponse(c, "неверный формат ID ответа")
		return
	}

	userRole, _ := getUserRole(c)
	if userRole != domain.UserRoleAdmin {
		// Здесь нужна дополнительная проверка, является ли пользователь автором ответа
		// Для этого потребуется получить ответ из БД, но такого метода нет в интерфейсе
		// Поэтому для простоты разрешим удаление только админам
		h.logger.Warn("попытка несанкционированного доступа", zap.Int64("userID", userID))
		forbiddenResponse(c)
		return
	}

	err = h.services.Review.DeleteReply(c.Request.Context(), replyID)
	if err != nil {
		h.logger.Error("ошибка удаления ответа на отзыв", zap.Error(err))
		internalServerErrorResponse(c)
		return
	}

	noContentResponse(c)
}

// @Summary Получить список отзывов
// @Description Возвращает список отзывов с возможностью фильтрации и пагинацией
// @Tags Отзывы
// @Accept json
// @Produce json
// @Param specialist_id query int true "ID специалиста"
// @Param client_id query int false "ID клиента"
// @Param min_rating query int false "Минимальный рейтинг"
// @Param max_rating query int false "Максимальный рейтинг"
// @Param limit query int false "Лимит записей на странице (по умолчанию 10)"
// @Param offset query int false "Смещение (по умолчанию 0)"
// @Success 200 {object} paginatedResponse "Список отзывов с пагинацией"
// @Failure 400 {object} errorResponseBody "Ошибка валидации параметров"
// @Failure 500 {object} errorResponseBody "Внутренняя ошибка сервера"
// @Router /reviews [get]
func (h *Handler) getReviews(c *gin.Context) {
	filter := domain.ReviewFilter{
		Limit:  10,
		Offset: 0,
	}

	specialistIDStr := c.Query("specialist_id")
	if specialistIDStr == "" {
		h.logger.Warn("отсутствует обязательный параметр specialist_id")
		badRequestResponse(c, "отсутствует обязательный параметр specialist_id")
		return
	}

	specialistID, err := strconv.ParseInt(specialistIDStr, 10, 64)
	if err != nil {
		h.logger.Warn("неверный формат ID специалиста", zap.Error(err))
		badRequestResponse(c, "неверный формат ID специалиста")
		return
	}
	filter.SpecialistID = &specialistID

	if clientIDStr := c.Query("client_id"); clientIDStr != "" {
		clientID, err := strconv.ParseInt(clientIDStr, 10, 64)
		if err == nil {
			filter.ClientID = &clientID
		}
	}

	if minRatingStr := c.Query("min_rating"); minRatingStr != "" {
		minRating, err := strconv.Atoi(minRatingStr)
		if err == nil {
			filter.MinRating = &minRating
		}
	}

	if maxRatingStr := c.Query("max_rating"); maxRatingStr != "" {
		maxRating, err := strconv.Atoi(maxRatingStr)
		if err == nil {
			filter.MaxRating = &maxRating
		}
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err == nil && limit > 0 {
			filter.Limit = limit
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err == nil && offset >= 0 {
			filter.Offset = offset
		}
	}

	reviews, total, err := h.services.Review.List(c.Request.Context(), filter)
	if err != nil {
		h.logger.Error("ошибка при получении отзывов", zap.Error(err))
		errorResponse(c, http.StatusInternalServerError, "ошибка при получении отзывов")
		return
	}

	page := filter.Offset/filter.Limit + 1
	paginatedSuccessResponse(c, reviews, total, page, filter.Limit)
}

// @Summary Получить ответы на отзыв
// @Description Возвращает список ответов на конкретный отзыв
// @Tags Отзывы
// @Accept json
// @Produce json
// @Param id path int true "ID отзыва"
// @Success 200 {array} domain.Reply "Список ответов на отзыв"
// @Failure 400 {object} errorResponseBody "Неверный формат ID отзыва"
// @Failure 404 {object} errorResponseBody "Отзыв не найден"
// @Failure 500 {object} errorResponseBody "Внутренняя ошибка сервера"
// @Router /reviews/{id}/replies [get]
func (h *Handler) getReviewReplies(c *gin.Context) {
	reviewID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		h.logger.Warn("неверный формат ID отзыва", zap.Error(err))
		badRequestResponse(c, "неверный формат ID отзыва")
		return
	}

	// Проверяем существование отзыва
	_, err = h.services.Review.GetByID(c.Request.Context(), reviewID)
	if err != nil {
		h.logger.Error("ошибка получения отзыва", zap.Error(err), zap.Int64("reviewID", reviewID))
		notFoundResponse(c, "отзыв не найден")
		return
	}

	replies, err := h.services.Review.GetRepliesByReviewID(c.Request.Context(), reviewID)
	if err != nil {
		h.logger.Error("ошибка получения ответов на отзыв", zap.Error(err), zap.Int64("reviewID", reviewID))
		errorResponse(c, http.StatusInternalServerError, "ошибка при получении ответов на отзыв")
		return
	}

	successResponse(c, http.StatusOK, replies)
}
