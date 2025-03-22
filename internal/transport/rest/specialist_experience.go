package rest

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"laps/internal/domain"
)

func (h *Handler) addSpecialistWorkExperience(c *gin.Context) {
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

	var req domain.WorkExperienceDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("неверный формат данных", zap.Error(err))
		badRequestResponse(c, "неверный формат данных")
		return
	}

	id, err := h.services.Specialist.AddWorkExperience(c.Request.Context(), specialistID, req)
	if err != nil {
		h.logger.Error("ошибка добавления опыта работы", zap.Error(err))
		internalServerErrorResponse(c)
		return
	}

	createdResponse(c, gin.H{"id": id})
}

func (h *Handler) updateSpecialistWorkExperience(c *gin.Context) {
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

	experienceID, err := strconv.ParseInt(c.Param("expId"), 10, 64)
	if err != nil {
		h.logger.Warn("неверный формат ID опыта работы", zap.Error(err))
		badRequestResponse(c, "неверный формат ID опыта работы")
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

	var req domain.WorkExperienceDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("неверный формат данных", zap.Error(err))
		badRequestResponse(c, "неверный формат данных")
		return
	}

	err = h.services.Specialist.UpdateWorkExperience(c.Request.Context(), experienceID, req)
	if err != nil {
		h.logger.Error("ошибка обновления опыта работы", zap.Error(err))
		notFoundResponse(c, "опыт работы не найден или ошибка обновления")
		return
	}

	messageResponse(c, http.StatusOK, "опыт работы успешно обновлен")
}

func (h *Handler) deleteSpecialistWorkExperience(c *gin.Context) {
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

	experienceID, err := strconv.ParseInt(c.Param("expId"), 10, 64)
	if err != nil {
		h.logger.Warn("неверный формат ID опыта работы", zap.Error(err))
		badRequestResponse(c, "неверный формат ID опыта работы")
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

	err = h.services.Specialist.DeleteWorkExperience(c.Request.Context(), experienceID)
	if err != nil {
		h.logger.Error("ошибка удаления опыта работы", zap.Error(err))
		notFoundResponse(c, "опыт работы не найден или ошибка удаления")
		return
	}

	noContentResponse(c)
}
