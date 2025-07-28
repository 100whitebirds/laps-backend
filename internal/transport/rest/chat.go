package rest

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"laps/internal/domain"
)

// Chat Sessions Endpoints

func (h *Handler) createChatSession(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		errorResponse(c, http.StatusUnauthorized, "user not authenticated")
		return
	}

	var dto domain.CreateChatSessionDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		errorResponse(c, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	session, err := h.services.Chat.CreateChatSession(c.Request.Context(), dto, userID.(int64))
	if err != nil {
		h.logger.Error("Failed to create chat session: " + err.Error())
		errorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	successResponse(c, http.StatusCreated, session)
}

func (h *Handler) getChatSession(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		errorResponse(c, http.StatusUnauthorized, "user not authenticated")
		return
	}

	sessionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "invalid session ID")
		return
	}

	session, err := h.services.Chat.GetChatSessionByID(c.Request.Context(), sessionID, userID.(int64))
	if err != nil {
		if err.Error() == "access denied: user not participant in chat session" {
			errorResponse(c, http.StatusForbidden, err.Error())
			return
		}
		h.logger.Error("Failed to get chat session: " + err.Error())
		errorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	if session == nil {
		errorResponse(c, http.StatusNotFound, "chat session not found")
		return
	}

	successResponse(c, http.StatusOK, session)
}

func (h *Handler) getChatSessionByAppointment(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		errorResponse(c, http.StatusUnauthorized, "user not authenticated")
		return
	}

	appointmentID, err := strconv.ParseInt(c.Param("appointment_id"), 10, 64)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "invalid appointment ID")
		return
	}

	session, err := h.services.Chat.GetChatSessionByAppointmentID(c.Request.Context(), appointmentID, userID.(int64))
	if err != nil {
		if err.Error() == "access denied: user not participant in appointment" {
			errorResponse(c, http.StatusForbidden, err.Error())
			return
		}
		h.logger.Error("Failed to get chat session by appointment: " + err.Error())
		errorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	if session == nil {
		errorResponse(c, http.StatusNotFound, "chat session not found for this appointment")
		return
	}

	successResponse(c, http.StatusOK, session)
}

func (h *Handler) getUserChatSessions(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		errorResponse(c, http.StatusUnauthorized, "user not authenticated")
		return
	}

	filter := domain.ChatFilter{
		Limit:  20,
		Offset: 0,
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			filter.Limit = limit
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			filter.Offset = offset
		}
	}

	if status := c.Query("status"); status != "" {
		filter.Status = &status
	}

	sessions, err := h.services.Chat.GetUserChatSessions(c.Request.Context(), userID.(int64), filter)
	if err != nil {
		h.logger.Error("Failed to get user chat sessions: " + err.Error())
		errorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	successResponse(c, http.StatusOK, sessions)
}

func (h *Handler) endChatSession(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		errorResponse(c, http.StatusUnauthorized, "user not authenticated")
		return
	}

	sessionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "invalid session ID")
		return
	}

	err = h.services.Chat.EndChatSession(c.Request.Context(), sessionID, userID.(int64))
	if err != nil {
		if err.Error() == "access denied: user not participant in chat session" {
			errorResponse(c, http.StatusForbidden, err.Error())
			return
		}
		h.logger.Error("Failed to end chat session: " + err.Error())
		errorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	successResponse(c, http.StatusOK, "Chat session ended successfully")
}

// Chat Messages Endpoints

func (h *Handler) sendMessage(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		errorResponse(c, http.StatusUnauthorized, "user not authenticated")
		return
	}

	var dto domain.CreateChatMessageDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		errorResponse(c, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	message, err := h.services.Chat.SendMessage(c.Request.Context(), dto, userID.(int64))
	if err != nil {
		if err.Error() == "access denied: user not participant in chat session" {
			errorResponse(c, http.StatusForbidden, err.Error())
			return
		}
		h.logger.Error("Failed to send message: " + err.Error())
		errorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	successResponse(c, http.StatusCreated, message)
}

func (h *Handler) getMessages(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		errorResponse(c, http.StatusUnauthorized, "user not authenticated")
		return
	}

	sessionID, err := strconv.ParseInt(c.Param("session_id"), 10, 64)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "invalid session ID")
		return
	}

	filter := domain.MessageFilter{
		Limit:  50,
		Offset: 0,
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			filter.Limit = limit
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			filter.Offset = offset
		}
	}

	if messageType := c.Query("message_type"); messageType != "" {
		filter.MessageType = &messageType
	}

	messages, err := h.services.Chat.GetMessages(c.Request.Context(), sessionID, filter, userID.(int64))
	if err != nil {
		if err.Error() == "access denied: user not participant in chat session" {
			errorResponse(c, http.StatusForbidden, err.Error())
			return
		}
		h.logger.Error("Failed to get messages: " + err.Error())
		errorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	successResponse(c, http.StatusOK, messages)
}

func (h *Handler) markMessageAsRead(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		errorResponse(c, http.StatusUnauthorized, "user not authenticated")
		return
	}

	messageID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "invalid message ID")
		return
	}

	err = h.services.Chat.MarkMessageAsRead(c.Request.Context(), messageID, userID.(int64))
	if err != nil {
		if err.Error() == "access denied: user cannot access this message" {
			errorResponse(c, http.StatusForbidden, err.Error())
			return
		}
		h.logger.Error("Failed to mark message as read: " + err.Error())
		errorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	successResponse(c, http.StatusOK, "Message marked as read")
}

func (h *Handler) startVideoCall(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		errorResponse(c, http.StatusUnauthorized, "user not authenticated")
		return
	}

	var dto domain.CreateVideoCallSessionDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		errorResponse(c, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	videoCall, err := h.services.Chat.StartVideoCall(c.Request.Context(), dto, userID.(int64))
	if err != nil {
		if err.Error() == "access denied: user not participant in chat session" {
			errorResponse(c, http.StatusForbidden, err.Error())
			return
		}
		h.logger.Error("Failed to start video call: " + err.Error())
		errorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	successResponse(c, http.StatusCreated, videoCall)
}

func (h *Handler) getVideoCallSession(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		errorResponse(c, http.StatusUnauthorized, "user not authenticated")
		return
	}

	callID := c.Param("id")
	if callID == "" {
		errorResponse(c, http.StatusBadRequest, "invalid call ID")
		return
	}

	videoCall, err := h.services.Chat.GetVideoCallSession(c.Request.Context(), callID, userID.(int64))
	if err != nil {
		if err.Error() == "access denied: user cannot access this video call" {
			errorResponse(c, http.StatusForbidden, err.Error())
			return
		}
		h.logger.Error("Failed to get video call session: " + err.Error())
		errorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	if videoCall == nil {
		errorResponse(c, http.StatusNotFound, "video call session not found")
		return
	}

	successResponse(c, http.StatusOK, videoCall)
} 