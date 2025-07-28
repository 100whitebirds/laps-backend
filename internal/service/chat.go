package service

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"laps/internal/domain"
	"laps/internal/repository"
)

type ChatService interface {
	// Chat Sessions
	CreateChatSession(ctx context.Context, dto domain.CreateChatSessionDTO, userID int64) (*domain.ChatSession, error)
	GetChatSessionByID(ctx context.Context, id int64, userID int64) (*domain.ChatSession, error)
	GetChatSessionByAppointmentID(ctx context.Context, appointmentID int64, userID int64) (*domain.ChatSession, error)
	GetUserChatSessions(ctx context.Context, userID int64, filter domain.ChatFilter) ([]domain.ChatSession, error)
	EndChatSession(ctx context.Context, sessionID int64, userID int64) error

	// Chat Messages
	SendMessage(ctx context.Context, dto domain.CreateChatMessageDTO, senderID int64) (*domain.ChatMessage, error)
	GetMessages(ctx context.Context, sessionID int64, filter domain.MessageFilter, userID int64) ([]domain.ChatMessage, error)
	MarkMessageAsRead(ctx context.Context, messageID int64, userID int64) error
	MarkAllMessagesAsRead(ctx context.Context, sessionID int64, userID int64) error
	UpdateMessage(ctx context.Context, messageID int64, dto domain.UpdateChatMessageDTO, userID int64) error
	DeleteMessage(ctx context.Context, messageID int64, userID int64) error

	// Video Call Sessions
	StartVideoCall(ctx context.Context, dto domain.CreateVideoCallSessionDTO, initiatorID int64) (*domain.VideoCallSession, error)
	GetVideoCallSession(ctx context.Context, callID string, userID int64) (*domain.VideoCallSession, error)
	UpdateVideoCallSession(ctx context.Context, callID string, dto domain.UpdateVideoCallSessionDTO, userID int64) error
	EndVideoCall(ctx context.Context, callID string, userID int64, endReason string, quality *string) error

	// Permissions & Access Control
	CanAccessChatSession(ctx context.Context, sessionID int64, userID int64) (bool, error)
	CanAccessMessage(ctx context.Context, messageID int64, userID int64) (bool, error)
	CanAccessVideoCall(ctx context.Context, callID string, userID int64) (bool, error)
}

type ChatServiceImpl struct {
	chatRepo        repository.ChatRepository
	appointmentRepo repository.AppointmentRepository
	userRepo        repository.UserRepository
	logger          *zap.Logger
}

func NewChatService(
	chatRepo repository.ChatRepository,
	appointmentRepo repository.AppointmentRepository,
	userRepo repository.UserRepository,
	logger *zap.Logger,
) ChatService {
	return &ChatServiceImpl{
		chatRepo:        chatRepo,
		appointmentRepo: appointmentRepo,
		userRepo:        userRepo,
		logger:          logger,
	}
}

// Chat Sessions Implementation

func (s *ChatServiceImpl) CreateChatSession(ctx context.Context, dto domain.CreateChatSessionDTO, userID int64) (*domain.ChatSession, error) {
	s.logger.Info("Creating chat session", zap.Int64("appointment_id", dto.AppointmentID), zap.Int64("user_id", userID))

	// Get appointment to validate access and get participants
	appointment, err := s.appointmentRepo.GetByID(ctx, dto.AppointmentID)
	if err != nil {
		s.logger.Error("Failed to get appointment", zap.Error(err))
		return nil, fmt.Errorf("failed to get appointment: %w", err)
	}

	if appointment == nil {
		return nil, fmt.Errorf("appointment not found")
	}

	// Validate user has access to this appointment
	if appointment.ClientID != userID && appointment.SpecialistID != userID {
		s.logger.Warn("User attempted to access unauthorized appointment", 
			zap.Int64("user_id", userID), 
			zap.Int64("appointment_id", dto.AppointmentID))
		return nil, fmt.Errorf("access denied: user not participant in appointment")
	}

	// Check if appointment is paid (for security)
	if appointment.Status != domain.AppointmentStatusPaid && appointment.Status != domain.AppointmentStatusCompleted {
		return nil, fmt.Errorf("chat session can only be created for paid appointments")
	}

	// Check if chat session already exists
	existingSession, err := s.chatRepo.GetChatSessionByAppointmentID(ctx, dto.AppointmentID)
	if err != nil {
		s.logger.Error("Failed to check existing chat session", zap.Error(err))
		return nil, fmt.Errorf("failed to check existing chat session: %w", err)
	}

	if existingSession != nil {
		s.logger.Info("Chat session already exists", zap.Int64("session_id", existingSession.ID))
		return existingSession, nil
	}

	// Create new chat session
	session, err := s.chatRepo.CreateChatSession(ctx, dto, appointment.ClientID, appointment.SpecialistID)
	if err != nil {
		s.logger.Error("Failed to create chat session", zap.Error(err))
		return nil, fmt.Errorf("failed to create chat session: %w", err)
	}

	s.logger.Info("Chat session created successfully", zap.Int64("session_id", session.ID))
	return session, nil
}

func (s *ChatServiceImpl) GetChatSessionByID(ctx context.Context, id int64, userID int64) (*domain.ChatSession, error) {
	// Check access first
	canAccess, err := s.CanAccessChatSession(ctx, id, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check access: %w", err)
	}

	if !canAccess {
		return nil, fmt.Errorf("access denied: user not participant in chat session")
	}

	session, err := s.chatRepo.GetChatSessionByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get chat session", zap.Error(err))
		return nil, fmt.Errorf("failed to get chat session: %w", err)
	}

	return session, nil
}

func (s *ChatServiceImpl) GetChatSessionByAppointmentID(ctx context.Context, appointmentID int64, userID int64) (*domain.ChatSession, error) {
	// Get appointment to validate access
	appointment, err := s.appointmentRepo.GetByID(ctx, appointmentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get appointment: %w", err)
	}

	if appointment == nil {
		return nil, fmt.Errorf("appointment not found")
	}

	// Validate user has access to this appointment
	if appointment.ClientID != userID && appointment.SpecialistID != userID {
		return nil, fmt.Errorf("access denied: user not participant in appointment")
	}

	session, err := s.chatRepo.GetChatSessionByAppointmentID(ctx, appointmentID)
	if err != nil {
		s.logger.Error("Failed to get chat session by appointment ID", zap.Error(err))
		return nil, fmt.Errorf("failed to get chat session: %w", err)
	}

	return session, nil
}

func (s *ChatServiceImpl) GetUserChatSessions(ctx context.Context, userID int64, filter domain.ChatFilter) ([]domain.ChatSession, error) {
	// Set default pagination if not provided
	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	if filter.Limit > 100 {
		filter.Limit = 100 // Max limit for performance
	}

	// Override user_id filter with the authenticated user's ID for security
	filter.UserID = &userID

	sessions, err := s.chatRepo.GetChatSessionsByUserID(ctx, userID, filter)
	if err != nil {
		s.logger.Error("Failed to get user chat sessions", zap.Error(err))
		return nil, fmt.Errorf("failed to get user chat sessions: %w", err)
	}

	return sessions, nil
}

func (s *ChatServiceImpl) EndChatSession(ctx context.Context, sessionID int64, userID int64) error {
	// Check access
	canAccess, err := s.CanAccessChatSession(ctx, sessionID, userID)
	if err != nil {
		return fmt.Errorf("failed to check access: %w", err)
	}

	if !canAccess {
		return fmt.Errorf("access denied: user not participant in chat session")
	}

	err = s.chatRepo.EndChatSession(ctx, sessionID)
	if err != nil {
		s.logger.Error("Failed to end chat session", zap.Error(err))
		return fmt.Errorf("failed to end chat session: %w", err)
	}

	s.logger.Info("Chat session ended", zap.Int64("session_id", sessionID), zap.Int64("user_id", userID))
	return nil
}

// Chat Messages Implementation

func (s *ChatServiceImpl) SendMessage(ctx context.Context, dto domain.CreateChatMessageDTO, senderID int64) (*domain.ChatMessage, error) {
	s.logger.Info("Sending message", zap.Int64("session_id", dto.SessionID), zap.Int64("sender_id", senderID))

	// Check access to the chat session
	canAccess, err := s.CanAccessChatSession(ctx, dto.SessionID, senderID)
	if err != nil {
		return nil, fmt.Errorf("failed to check access: %w", err)
	}

	if !canAccess {
		return nil, fmt.Errorf("access denied: user not participant in chat session")
	}

	// Get the chat session to check if it's active
	session, err := s.chatRepo.GetChatSessionByID(ctx, dto.SessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get chat session: %w", err)
	}

	if session == nil {
		return nil, fmt.Errorf("chat session not found")
	}

	if session.Status == domain.ChatSessionStatusEnded {
		return nil, fmt.Errorf("cannot send message to ended chat session")
	}

	// Validate message content
	if dto.Content == "" && dto.MessageType == domain.MessageTypeText {
		return nil, fmt.Errorf("text message content cannot be empty")
	}

	// Create the message
	message, err := s.chatRepo.CreateMessage(ctx, dto, senderID)
	if err != nil {
		s.logger.Error("Failed to create message", zap.Error(err))
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	s.logger.Info("Message sent successfully", zap.Int64("message_id", message.ID))
	return message, nil
}

func (s *ChatServiceImpl) GetMessages(ctx context.Context, sessionID int64, filter domain.MessageFilter, userID int64) ([]domain.ChatMessage, error) {
	// Check access to the chat session
	canAccess, err := s.CanAccessChatSession(ctx, sessionID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check access: %w", err)
	}

	if !canAccess {
		return nil, fmt.Errorf("access denied: user not participant in chat session")
	}

	// Set default pagination
	if filter.Limit <= 0 {
		filter.Limit = 50
	}
	if filter.Limit > 200 {
		filter.Limit = 200 // Max limit for performance
	}

	// Override session_id filter for security
	filter.SessionID = &sessionID

	messages, err := s.chatRepo.GetMessagesBySessionID(ctx, sessionID, filter)
	if err != nil {
		s.logger.Error("Failed to get messages", zap.Error(err))
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}

	return messages, nil
}

func (s *ChatServiceImpl) MarkMessageAsRead(ctx context.Context, messageID int64, userID int64) error {
	// Check access to the message
	canAccess, err := s.CanAccessMessage(ctx, messageID, userID)
	if err != nil {
		return fmt.Errorf("failed to check access: %w", err)
	}

	if !canAccess {
		return fmt.Errorf("access denied: user cannot access this message")
	}

	err = s.chatRepo.MarkMessageAsRead(ctx, messageID, userID)
	if err != nil {
		s.logger.Error("Failed to mark message as read", zap.Error(err))
		return fmt.Errorf("failed to mark message as read: %w", err)
	}

	return nil
}

func (s *ChatServiceImpl) MarkAllMessagesAsRead(ctx context.Context, sessionID int64, userID int64) error {
	// Check access to the chat session
	canAccess, err := s.CanAccessChatSession(ctx, sessionID, userID)
	if err != nil {
		return fmt.Errorf("failed to check access: %w", err)
	}

	if !canAccess {
		return fmt.Errorf("access denied: user not participant in chat session")
	}

	err = s.chatRepo.MarkAllMessagesAsRead(ctx, sessionID, userID)
	if err != nil {
		s.logger.Error("Failed to mark all messages as read", zap.Error(err))
		return fmt.Errorf("failed to mark all messages as read: %w", err)
	}

	return nil
}

func (s *ChatServiceImpl) UpdateMessage(ctx context.Context, messageID int64, dto domain.UpdateChatMessageDTO, userID int64) error {
	// Get the message to check ownership
	message, err := s.chatRepo.GetMessageByID(ctx, messageID)
	if err != nil {
		return fmt.Errorf("failed to get message: %w", err)
	}

	if message == nil {
		return fmt.Errorf("message not found")
	}

	// Only the sender can update the message content
	if dto.Content != nil && message.SenderID != userID {
		return fmt.Errorf("access denied: only message sender can update content")
	}

	// Check access to the message for read status updates
	if dto.IsRead != nil {
		canAccess, err := s.CanAccessMessage(ctx, messageID, userID)
		if err != nil {
			return fmt.Errorf("failed to check access: %w", err)
		}

		if !canAccess {
			return fmt.Errorf("access denied: user cannot access this message")
		}
	}

	err = s.chatRepo.UpdateMessage(ctx, messageID, dto)
	if err != nil {
		s.logger.Error("Failed to update message", zap.Error(err))
		return fmt.Errorf("failed to update message: %w", err)
	}

	return nil
}

func (s *ChatServiceImpl) DeleteMessage(ctx context.Context, messageID int64, userID int64) error {
	// Get the message to check ownership
	message, err := s.chatRepo.GetMessageByID(ctx, messageID)
	if err != nil {
		return fmt.Errorf("failed to get message: %w", err)
	}

	if message == nil {
		return fmt.Errorf("message not found")
	}

	// Only the sender can delete the message
	if message.SenderID != userID {
		return fmt.Errorf("access denied: only message sender can delete message")
	}

	err = s.chatRepo.DeleteMessage(ctx, messageID)
	if err != nil {
		s.logger.Error("Failed to delete message", zap.Error(err))
		return fmt.Errorf("failed to delete message: %w", err)
	}

	s.logger.Info("Message deleted", zap.Int64("message_id", messageID), zap.Int64("user_id", userID))
	return nil
}

// Video Call Sessions Implementation

func (s *ChatServiceImpl) StartVideoCall(ctx context.Context, dto domain.CreateVideoCallSessionDTO, initiatorID int64) (*domain.VideoCallSession, error) {
	s.logger.Info("Starting video call", zap.Int64("chat_session_id", dto.ChatSessionID), zap.Int64("initiator_id", initiatorID))

	// Check access to the chat session
	canAccess, err := s.CanAccessChatSession(ctx, dto.ChatSessionID, initiatorID)
	if err != nil {
		return nil, fmt.Errorf("failed to check access: %w", err)
	}

	if !canAccess {
		return nil, fmt.Errorf("access denied: user not participant in chat session")
	}

	// Get the chat session to validate it's active
	session, err := s.chatRepo.GetChatSessionByID(ctx, dto.ChatSessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get chat session: %w", err)
	}

	if session == nil {
		return nil, fmt.Errorf("chat session not found")
	}

	if session.Status != domain.ChatSessionStatusActive {
		return nil, fmt.Errorf("video call can only be started in active chat sessions")
	}

	// Check if there's already an active video call
	existingCalls, err := s.chatRepo.GetVideoCallSessionsByChatID(ctx, dto.ChatSessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing video calls: %w", err)
	}

	for _, call := range existingCalls {
		if call.Status == domain.VideoCallStatusWaiting || call.Status == domain.VideoCallStatusRinging || call.Status == domain.VideoCallStatusActive {
			return nil, fmt.Errorf("there is already an active video call in this chat session")
		}
	}

	// Create the video call session
	videoCall, err := s.chatRepo.CreateVideoCallSession(ctx, dto, initiatorID)
	if err != nil {
		s.logger.Error("Failed to create video call session", zap.Error(err))
		return nil, fmt.Errorf("failed to create video call session: %w", err)
	}

	// Send system message about video call started
	systemMessage := domain.CreateChatMessageDTO{
		SessionID:   dto.ChatSessionID,
		MessageType: domain.MessageTypeSystem,
		Content:     fmt.Sprintf("Video call started by %s", getInitiatorName(session, initiatorID)),
	}

	_, err = s.chatRepo.CreateMessage(ctx, systemMessage, initiatorID)
	if err != nil {
		s.logger.Warn("Failed to create system message for video call start", zap.Error(err))
	}

	s.logger.Info("Video call session created", zap.String("call_id", videoCall.ID))
	return videoCall, nil
}

func (s *ChatServiceImpl) GetVideoCallSession(ctx context.Context, callID string, userID int64) (*domain.VideoCallSession, error) {
	// Check access
	canAccess, err := s.CanAccessVideoCall(ctx, callID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check access: %w", err)
	}

	if !canAccess {
		return nil, fmt.Errorf("access denied: user cannot access this video call")
	}

	videoCall, err := s.chatRepo.GetVideoCallSessionByID(ctx, callID)
	if err != nil {
		s.logger.Error("Failed to get video call session", zap.Error(err))
		return nil, fmt.Errorf("failed to get video call session: %w", err)
	}

	return videoCall, nil
}

func (s *ChatServiceImpl) UpdateVideoCallSession(ctx context.Context, callID string, dto domain.UpdateVideoCallSessionDTO, userID int64) error {
	// Check access
	canAccess, err := s.CanAccessVideoCall(ctx, callID, userID)
	if err != nil {
		return fmt.Errorf("failed to check access: %w", err)
	}

	if !canAccess {
		return fmt.Errorf("access denied: user cannot access this video call")
	}

	err = s.chatRepo.UpdateVideoCallSession(ctx, callID, dto)
	if err != nil {
		s.logger.Error("Failed to update video call session", zap.Error(err))
		return fmt.Errorf("failed to update video call session: %w", err)
	}

	return nil
}

func (s *ChatServiceImpl) EndVideoCall(ctx context.Context, callID string, userID int64, endReason string, quality *string) error {
	// Check access
	canAccess, err := s.CanAccessVideoCall(ctx, callID, userID)
	if err != nil {
		return fmt.Errorf("failed to check access: %w", err)
	}

	if !canAccess {
		return fmt.Errorf("access denied: user cannot access this video call")
	}

	// Get the video call to get chat session info
	videoCall, err := s.chatRepo.GetVideoCallSessionByID(ctx, callID)
	if err != nil {
		return fmt.Errorf("failed to get video call session: %w", err)
	}

	if videoCall == nil {
		return fmt.Errorf("video call session not found")
	}

	err = s.chatRepo.EndVideoCallSession(ctx, callID, endReason, quality)
	if err != nil {
		s.logger.Error("Failed to end video call session", zap.Error(err))
		return fmt.Errorf("failed to end video call session: %w", err)
	}

	// Send system message about video call ended
	systemMessage := domain.CreateChatMessageDTO{
		SessionID:   videoCall.ChatSessionID,
		MessageType: domain.MessageTypeSystem,
		Content:     "Video call ended",
	}

	_, err = s.chatRepo.CreateMessage(ctx, systemMessage, userID)
	if err != nil {
		s.logger.Warn("Failed to create system message for video call end", zap.Error(err))
	}

	s.logger.Info("Video call ended", zap.String("call_id", callID), zap.Int64("user_id", userID))
	return nil
}

// Access Control Implementation

func (s *ChatServiceImpl) CanAccessChatSession(ctx context.Context, sessionID int64, userID int64) (bool, error) {
	session, err := s.chatRepo.GetChatSessionByID(ctx, sessionID)
	if err != nil {
		return false, fmt.Errorf("failed to get chat session: %w", err)
	}

	if session == nil {
		return false, nil
	}

	return session.ClientID == userID || session.SpecialistID == userID, nil
}

func (s *ChatServiceImpl) CanAccessMessage(ctx context.Context, messageID int64, userID int64) (bool, error) {
	message, err := s.chatRepo.GetMessageByID(ctx, messageID)
	if err != nil {
		return false, fmt.Errorf("failed to get message: %w", err)
	}

	if message == nil {
		return false, nil
	}

	// Check if user has access to the chat session
	return s.CanAccessChatSession(ctx, message.SessionID, userID)
}

func (s *ChatServiceImpl) CanAccessVideoCall(ctx context.Context, callID string, userID int64) (bool, error) {
	videoCall, err := s.chatRepo.GetVideoCallSessionByID(ctx, callID)
	if err != nil {
		return false, fmt.Errorf("failed to get video call session: %w", err)
	}

	if videoCall == nil {
		return false, nil
	}

	// Check if user has access to the chat session
	return s.CanAccessChatSession(ctx, videoCall.ChatSessionID, userID)
}

// Helper functions

func getInitiatorName(session *domain.ChatSession, initiatorID int64) string {
	if session.Client != nil && session.Client.ID == initiatorID {
		return session.Client.FirstName + " " + session.Client.LastName
	}
	if session.Specialist != nil && session.Specialist.ID == initiatorID {
		return session.Specialist.FirstName + " " + session.Specialist.LastName
	}
	return "User"
} 