package rest

import (
	"bytes"
	"fmt"
	"io"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func (h *Handler) updateSpecialistWorkExperience(c *gin.Context) {
	experienceID := c.Param("expId")
	h.logger.Info("перенаправление запроса на обновление опыта работы",
		zap.String("experienceID", experienceID),
		zap.String("oldPath", c.Request.URL.Path))

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		h.logger.Error("ошибка чтения тела запроса", zap.Error(err))
		badRequestResponse(c, "неверный формат данных")
		return
	}
	c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

	targetURL := fmt.Sprintf("/api/v1/work-experience/%s", experienceID)
	h.logger.Info("новый путь запроса", zap.String("targetURL", targetURL))

	c.Request.URL.Path = targetURL
	c.Request.RequestURI = targetURL

	h.updateWorkExperience(c)
}

func (h *Handler) deleteSpecialistWorkExperience(c *gin.Context) {
	experienceID := c.Param("expId")
	h.logger.Info("перенаправление запроса на удаление опыта работы",
		zap.String("experienceID", experienceID),
		zap.String("oldPath", c.Request.URL.Path))

	targetURL := fmt.Sprintf("/api/v1/work-experience/%s", experienceID)
	h.logger.Info("новый путь запроса", zap.String("targetURL", targetURL))

	c.Request.URL.Path = targetURL
	c.Request.RequestURI = targetURL

	h.deleteWorkExperience(c)
}
