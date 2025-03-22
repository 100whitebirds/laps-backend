package rest

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"laps/internal/domain"
)

func (h *Handler) addSpecialistSpecialization(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		h.logger.Warn("ошибка получения ID пользователя", zap.Error(err))
		unauthorizedResponse(c)
		return
	}

	specialistID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		h.logger.Warn("неверный формат ID специалиста", zap.Error(err))
		badRequestResponse(c, "неверный формат ID специалиста")
		return
	}

	specializationID, err := strconv.ParseInt(c.Param("specId"), 10, 64)
	if err != nil {
		h.logger.Warn("неверный формат ID специализации", zap.Error(err))
		badRequestResponse(c, "неверный формат ID специализации")
		return
	}

	userRole, _ := getUserRole(c)
	specialist, err := h.services.Specialist.GetByID(c.Request.Context(), specialistID)
	if err != nil {
		h.logger.Error("ошибка получения данных специалиста", zap.Error(err))
		notFoundResponse(c, "специалист не найден")
		return
	}

	if specialist.UserID != userID && userRole != domain.UserRoleAdmin {
		h.logger.Warn("попытка несанкционированного доступа",
			zap.Int64("userID", userID),
			zap.Int64("specialistID", specialistID))
		forbiddenResponse(c)
		return
	}

	err = h.services.Specialist.AddSpecialization(c.Request.Context(), specialistID, specializationID)
	if err != nil {
		h.logger.Error("ошибка добавления специализации", zap.Error(err))
		badRequestResponse(c, "ошибка добавления специализации")
		return
	}

	messageResponse(c, http.StatusOK, "специализация успешно добавлена")
}

func (h *Handler) removeSpecialistSpecialization(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		h.logger.Warn("ошибка получения ID пользователя", zap.Error(err))
		unauthorizedResponse(c)
		return
	}

	specialistID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		h.logger.Warn("неверный формат ID специалиста", zap.Error(err))
		badRequestResponse(c, "неверный формат ID специалиста")
		return
	}

	specializationID, err := strconv.ParseInt(c.Param("specId"), 10, 64)
	if err != nil {
		h.logger.Warn("неверный формат ID специализации", zap.Error(err))
		badRequestResponse(c, "неверный формат ID специализации")
		return
	}

	userRole, _ := getUserRole(c)
	specialist, err := h.services.Specialist.GetByID(c.Request.Context(), specialistID)
	if err != nil {
		h.logger.Error("ошибка получения данных специалиста", zap.Error(err))
		notFoundResponse(c, "специалист не найден")
		return
	}

	if specialist.UserID != userID && userRole != domain.UserRoleAdmin {
		h.logger.Warn("попытка несанкционированного доступа",
			zap.Int64("userID", userID),
			zap.Int64("specialistID", specialistID))
		forbiddenResponse(c)
		return
	}

	err = h.services.Specialist.RemoveSpecialization(c.Request.Context(), specialistID, specializationID)
	if err != nil {
		h.logger.Error("ошибка удаления специализации", zap.Error(err))
		badRequestResponse(c, "ошибка удаления специализации")
		return
	}

	noContentResponse(c)
}
