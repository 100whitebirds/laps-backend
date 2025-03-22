package rest

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"laps/internal/domain"
)

// @Summary Создать пользователя
// @Description Создает нового пользователя (только для администраторов)
// @Tags Пользователи
// @Accept json
// @Produce json
// @Param input body domain.CreateUserDTO true "Данные пользователя"
// @Success 201 {object} map[string]interface{} "ID созданного пользователя"
// @Failure 400 {object} errorResponseBody "Ошибка валидации"
// @Failure 401 {object} errorResponseBody "Не авторизован"
// @Failure 403 {object} errorResponseBody "Доступ запрещен"
// @Failure 500 {object} errorResponseBody "Внутренняя ошибка сервера"
// @Security ApiKeyAuth
// @Router /users [post]
func (h *Handler) createUser(c *gin.Context) {
	var req domain.CreateUserDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("неверный формат данных", zap.Error(err))
		badRequestResponse(c, "неверный формат данных")
		return
	}

	id, err := h.services.User.Create(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("ошибка при создании пользователя", zap.Error(err))
		errorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	createdResponse(c, map[string]interface{}{
		"id": id,
	})
}

// @Summary Получить пользователя по ID
// @Description Возвращает информацию о пользователе по указанному ID
// @Tags Пользователи
// @Accept json
// @Produce json
// @Param id path int true "ID пользователя"
// @Success 200 {object} domain.User "Данные пользователя"
// @Failure 400 {object} errorResponseBody "Неверный формат ID"
// @Failure 401 {object} errorResponseBody "Не авторизован"
// @Failure 403 {object} errorResponseBody "Доступ запрещен"
// @Failure 404 {object} errorResponseBody "Пользователь не найден"
// @Security ApiKeyAuth
// @Router /users/{id} [get]
func (h *Handler) getUserByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		badRequestResponse(c, "неверный формат ID")
		return
	}

	currentUserID, err := getUserID(c)
	if err != nil {
		unauthorizedResponse(c)
		return
	}

	userRole, err := getUserRole(c)
	if err != nil {
		unauthorizedResponse(c)
		return
	}

	if currentUserID != id && userRole != domain.UserRoleAdmin {
		forbiddenResponse(c)
		return
	}

	user, err := h.services.User.GetByID(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("ошибка при получении пользователя", zap.Error(err))
		notFoundResponse(c, "пользователь не найден")
		return
	}

	successResponse(c, http.StatusOK, user)
}

// @Summary Обновить пользователя
// @Description Обновляет данные пользователя
// @Tags Пользователи
// @Accept json
// @Produce json
// @Param id path int true "ID пользователя"
// @Param input body domain.UpdateUserDTO true "Новые данные пользователя"
// @Success 204 {object} nil "Пользователь успешно обновлен"
// @Failure 400 {object} errorResponseBody "Ошибка валидации"
// @Failure 401 {object} errorResponseBody "Не авторизован"
// @Failure 403 {object} errorResponseBody "Доступ запрещен"
// @Failure 500 {object} errorResponseBody "Внутренняя ошибка сервера"
// @Security ApiKeyAuth
// @Router /users/{id} [put]
func (h *Handler) updateUser(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		badRequestResponse(c, "неверный формат ID")
		return
	}

	currentUserID, err := getUserID(c)
	if err != nil {
		unauthorizedResponse(c)
		return
	}

	userRole, err := getUserRole(c)
	if err != nil {
		unauthorizedResponse(c)
		return
	}

	if currentUserID != id && userRole != domain.UserRoleAdmin {
		forbiddenResponse(c)
		return
	}

	var req domain.UpdateUserDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("неверный формат данных", zap.Error(err))
		badRequestResponse(c, "неверный формат данных")
		return
	}

	err = h.services.User.Update(c.Request.Context(), id, req)
	if err != nil {
		h.logger.Error("ошибка при обновлении пользователя", zap.Error(err))
		errorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	noContentResponse(c)
}

// @Summary Обновить пароль пользователя
// @Description Обновляет пароль пользователя (только сам пользователь может обновить свой пароль)
// @Tags Пользователи
// @Accept json
// @Produce json
// @Param id path int true "ID пользователя"
// @Param input body domain.PasswordUpdateDTO true "Данные для обновления пароля"
// @Success 204 {object} nil "Пароль успешно обновлен"
// @Failure 400 {object} errorResponseBody "Ошибка валидации"
// @Failure 401 {object} errorResponseBody "Не авторизован"
// @Failure 403 {object} errorResponseBody "Доступ запрещен"
// @Failure 500 {object} errorResponseBody "Внутренняя ошибка сервера"
// @Security ApiKeyAuth
// @Router /users/{id}/password [put]
func (h *Handler) updatePassword(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		badRequestResponse(c, "неверный формат ID")
		return
	}

	currentUserID, err := getUserID(c)
	if err != nil {
		unauthorizedResponse(c)
		return
	}

	if currentUserID != id {
		forbiddenResponse(c)
		return
	}

	var req domain.PasswordUpdateDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("неверный формат данных", zap.Error(err))
		badRequestResponse(c, "неверный формат данных")
		return
	}

	err = h.services.User.UpdatePassword(c.Request.Context(), id, req)
	if err != nil {
		h.logger.Error("ошибка при обновлении пароля", zap.Error(err))
		errorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	noContentResponse(c)
}

// @Summary Удалить пользователя
// @Description Удаляет пользователя по ID (только для администраторов)
// @Tags Пользователи
// @Accept json
// @Produce json
// @Param id path int true "ID пользователя"
// @Success 204 {object} nil "Пользователь успешно удален"
// @Failure 400 {object} errorResponseBody "Неверный формат ID"
// @Failure 401 {object} errorResponseBody "Не авторизован"
// @Failure 403 {object} errorResponseBody "Доступ запрещен"
// @Failure 500 {object} errorResponseBody "Внутренняя ошибка сервера"
// @Security ApiKeyAuth
// @Router /users/{id} [delete]
func (h *Handler) deleteUser(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		badRequestResponse(c, "неверный формат ID")
		return
	}

	err = h.services.User.Delete(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("ошибка при удалении пользователя", zap.Error(err))
		errorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	noContentResponse(c)
}

// @Summary Получить список пользователей
// @Description Возвращает список пользователей с пагинацией (только для администраторов)
// @Tags Пользователи
// @Accept json
// @Produce json
// @Param limit query int false "Лимит записей на странице (по умолчанию 20)"
// @Param offset query int false "Смещение (по умолчанию 0)"
// @Success 200 {array} domain.User "Список пользователей"
// @Failure 401 {object} errorResponseBody "Не авторизован"
// @Failure 403 {object} errorResponseBody "Доступ запрещен"
// @Failure 500 {object} errorResponseBody "Внутренняя ошибка сервера"
// @Security ApiKeyAuth
// @Router /users [get]
func (h *Handler) getUsers(c *gin.Context) {
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if err != nil || limit < 0 {
		limit = 20
	}

	offset, err := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if err != nil || offset < 0 {
		offset = 0
	}

	users, err := h.services.User.List(c.Request.Context(), limit, offset)
	if err != nil {
		h.logger.Error("ошибка при получении списка пользователей", zap.Error(err))
		errorResponse(c, http.StatusInternalServerError, "ошибка при получении списка пользователей")
		return
	}

	successResponse(c, http.StatusOK, users)
}

// @Summary Получить текущего пользователя
// @Description Возвращает информацию о текущем авторизованном пользователе
// @Tags Пользователи
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} domain.User "Данные пользователя"
// @Failure 401 {object} errorResponseBody "Не авторизован"
// @Failure 500 {object} errorResponseBody "Внутренняя ошибка сервера"
// @Router /users/me [get]
func (h *Handler) getCurrentUser(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		unauthorizedResponse(c)
		return
	}

	user, err := h.services.User.GetByID(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("ошибка при получении текущего пользователя", zap.Error(err))
		internalServerErrorResponse(c)
		return
	}

	successResponse(c, http.StatusOK, user)
}
