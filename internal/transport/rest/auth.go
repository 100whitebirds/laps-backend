package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"laps/internal/domain"
)

// @Summary Регистрация нового пользователя
// @Description Регистрирует нового пользователя в системе
// @Tags Авторизация
// @Accept json
// @Produce json
// @Param input body domain.RegisterRequest true "Данные для регистрации"
// @Success 201 {object} domain.Tokens "Токены доступа и обновления"
// @Failure 400 {object} errorResponseBody "Ошибка валидации"
// @Failure 500 {object} errorResponseBody "Внутренняя ошибка сервера"
// @Router /auth/register [post]
func (h *Handler) register(c *gin.Context) {
	var input domain.RegisterRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		h.logger.Warn("неверный формат данных", zap.Error(err))
		badRequestResponse(c, "неверный формат данных")
		return
	}

	id, err := h.services.Auth.Register(c.Request.Context(), input)
	if err != nil {
		h.logger.Error("ошибка при регистрации", zap.Error(err))
		errorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	createdResponse(c, map[string]interface{}{
		"id": id,
	})
}

// @Summary Вход в систему
// @Description Авторизует пользователя и возвращает токены доступа
// @Tags Авторизация
// @Accept json
// @Produce json
// @Param input body domain.LoginRequest true "Данные для входа"
// @Success 200 {object} domain.Tokens "Токены доступа и обновления"
// @Failure 400 {object} errorResponseBody "Ошибка валидации"
// @Failure 401 {object} errorResponseBody "Неверные учетные данные"
// @Failure 500 {object} errorResponseBody "Внутренняя ошибка сервера"
// @Router /auth/login [post]
func (h *Handler) login(c *gin.Context) {
	var input domain.LoginRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		h.logger.Warn("неверный формат данных", zap.Error(err))
		badRequestResponse(c, "неверный формат данных")
		return
	}

	userAgent := c.Request.UserAgent()
	ip := c.ClientIP()

	tokens, err := h.services.Auth.Login(c.Request.Context(), input, userAgent, ip)
	if err != nil {
		h.logger.Error("ошибка при входе", zap.Error(err))
		errorResponse(c, http.StatusUnauthorized, err.Error())
		return
	}

	successResponse(c, http.StatusOK, tokens)
}

// @Summary Обновление токена
// @Description Обновляет токены доступа и обновления
// @Tags Авторизация
// @Accept json
// @Produce json
// @Param input body domain.RefreshTokenRequest true "Токен обновления"
// @Success 200 {object} domain.Tokens "Новые токены доступа и обновления"
// @Failure 400 {object} errorResponseBody "Ошибка валидации"
// @Failure 401 {object} errorResponseBody "Неверный токен обновления"
// @Failure 500 {object} errorResponseBody "Внутренняя ошибка сервера"
// @Router /auth/refresh [post]
func (h *Handler) refreshTokens(c *gin.Context) {
	var input domain.RefreshTokenRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		h.logger.Warn("неверный формат данных", zap.Error(err))
		badRequestResponse(c, "неверный формат данных")
		return
	}

	userAgent := c.Request.UserAgent()
	ip := c.ClientIP()

	tokens, err := h.services.Auth.RefreshTokens(c.Request.Context(), input.RefreshToken, userAgent, ip)
	if err != nil {
		h.logger.Error("ошибка при обновлении токенов", zap.Error(err))
		errorResponse(c, http.StatusUnauthorized, err.Error())
		return
	}

	successResponse(c, http.StatusOK, tokens)
}

// @Summary Выход из системы
// @Description Завершает сессию пользователя и инвалидирует токены
// @Tags Авторизация
// @Accept json
// @Produce json
// @Param input body domain.RefreshTokenRequest true "Токен обновления"
// @Success 204 {object} nil "Успешный выход"
// @Failure 400 {object} errorResponseBody "Ошибка валидации"
// @Failure 500 {object} errorResponseBody "Внутренняя ошибка сервера"
// @Router /auth/logout [post]
func (h *Handler) logout(c *gin.Context) {
	var input domain.RefreshTokenRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		h.logger.Warn("неверный формат данных", zap.Error(err))
		badRequestResponse(c, "неверный формат данных")
		return
	}

	err := h.services.Auth.Logout(c.Request.Context(), input.RefreshToken)
	if err != nil {
		h.logger.Error("ошибка при выходе", zap.Error(err))
		errorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	noContentResponse(c)
}
