package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"laps/internal/domain"
)

type ChatRepository interface {
	// Chat Sessions
	CreateChatSession(ctx context.Context, dto domain.CreateChatSessionDTO, clientID, specialistID int64) (*domain.ChatSession, error)
	GetChatSessionByID(ctx context.Context, id int64) (*domain.ChatSession, error)
	GetChatSessionByAppointmentID(ctx context.Context, appointmentID int64) (*domain.ChatSession, error)
	GetChatSessionsByUserID(ctx context.Context, userID int64, filter domain.ChatFilter) ([]domain.ChatSession, error)
	UpdateChatSessionStatus(ctx context.Context, id int64, status string) error
	EndChatSession(ctx context.Context, id int64) error

	// Chat Messages
	CreateMessage(ctx context.Context, dto domain.CreateChatMessageDTO, senderID int64) (*domain.ChatMessage, error)
	GetMessageByID(ctx context.Context, id int64) (*domain.ChatMessage, error)
	GetMessagesBySessionID(ctx context.Context, sessionID int64, filter domain.MessageFilter) ([]domain.ChatMessage, error)
	UpdateMessage(ctx context.Context, id int64, dto domain.UpdateChatMessageDTO) error
	MarkMessageAsRead(ctx context.Context, id int64, userID int64) error
	MarkAllMessagesAsRead(ctx context.Context, sessionID int64, userID int64) error
	DeleteMessage(ctx context.Context, id int64) error

	// Video Call Sessions
	CreateVideoCallSession(ctx context.Context, dto domain.CreateVideoCallSessionDTO, initiatorID int64) (*domain.VideoCallSession, error)
	GetVideoCallSessionByID(ctx context.Context, id string) (*domain.VideoCallSession, error)
	GetVideoCallSessionsByChatID(ctx context.Context, chatSessionID int64) ([]domain.VideoCallSession, error)
	UpdateVideoCallSession(ctx context.Context, id string, dto domain.UpdateVideoCallSessionDTO) error
	EndVideoCallSession(ctx context.Context, id string, endReason string, quality *string) error

	// Chat Participants
	AddParticipant(ctx context.Context, sessionID, userID int64, role string) error
	RemoveParticipant(ctx context.Context, sessionID, userID int64) error
	GetParticipantsBySessionID(ctx context.Context, sessionID int64) ([]domain.ChatParticipant, error)
}

type ChatRepo struct {
	db *pgxpool.Pool
}

func NewChatRepository(db *pgxpool.Pool) ChatRepository {
	return &ChatRepo{
		db: db,
	}
}

// Chat Sessions Implementation

func (r *ChatRepo) CreateChatSession(ctx context.Context, dto domain.CreateChatSessionDTO, clientID, specialistID int64) (*domain.ChatSession, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `
		INSERT INTO chat_sessions (appointment_id, client_id, specialist_id, status, started_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $6)
		RETURNING id, appointment_id, client_id, specialist_id, status, started_at, ended_at, created_at, updated_at
	`

	now := time.Now()
	var session domain.ChatSession

	err = tx.QueryRow(ctx, query,
		dto.AppointmentID,
		clientID,
		specialistID,
		domain.ChatSessionStatusActive,
		now,
		now,
	).Scan(
		&session.ID,
		&session.AppointmentID,
		&session.ClientID,
		&session.SpecialistID,
		&session.Status,
		&session.StartedAt,
		&session.EndedAt,
		&session.CreatedAt,
		&session.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create chat session: %w", err)
	}

	// Add participants
	_, err = tx.Exec(ctx, `
		INSERT INTO chat_participants (session_id, user_id, role, joined_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $5), ($1, $6, $7, $4, $5, $5)
	`, session.ID, clientID, domain.ParticipantRoleClient, now, now, specialistID, domain.ParticipantRoleSpecialist)

	if err != nil {
		return nil, fmt.Errorf("failed to add participants: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &session, nil
}

func (r *ChatRepo) GetChatSessionByID(ctx context.Context, id int64) (*domain.ChatSession, error) {
	query := `
		SELECT cs.id, cs.appointment_id, cs.client_id, cs.specialist_id, cs.status, 
		       cs.started_at, cs.ended_at, cs.created_at, cs.updated_at,
		       u1.first_name, u1.last_name, u1.email,
		       u2.first_name, u2.last_name, u2.email,
		       cm.id, cm.content, cm.message_type, cm.created_at
		FROM chat_sessions cs
		LEFT JOIN users u1 ON cs.client_id = u1.id
		LEFT JOIN users u2 ON cs.specialist_id = u2.id
		LEFT JOIN chat_messages cm ON cs.id = cm.session_id 
		    AND cm.id = (SELECT MAX(id) FROM chat_messages WHERE session_id = cs.id)
		WHERE cs.id = $1
	`

	var session domain.ChatSession
	var client domain.User
	var specialist domain.User
	var lastMessage domain.ChatMessage
	var lastMessageID *int64
	var lastMessageContent, lastMessageType *string
	var lastMessageCreatedAt *time.Time

	err := r.db.QueryRow(ctx, query, id).Scan(
		&session.ID,
		&session.AppointmentID,
		&session.ClientID,
		&session.SpecialistID,
		&session.Status,
		&session.StartedAt,
		&session.EndedAt,
		&session.CreatedAt,
		&session.UpdatedAt,
		&client.FirstName,
		&client.LastName,
		&client.Email,
		&specialist.FirstName,
		&specialist.LastName,
		&specialist.Email,
		&lastMessageID,
		&lastMessageContent,
		&lastMessageType,
		&lastMessageCreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get chat session: %w", err)
	}

	client.ID = session.ClientID
	specialist.ID = session.SpecialistID
	session.Client = &client
	session.Specialist = &specialist

	if lastMessageID != nil {
		lastMessage.ID = *lastMessageID
		lastMessage.Content = *lastMessageContent
		lastMessage.MessageType = *lastMessageType
		lastMessage.CreatedAt = *lastMessageCreatedAt
		session.LastMessage = &lastMessage
	}

	return &session, nil
}

func (r *ChatRepo) GetChatSessionByAppointmentID(ctx context.Context, appointmentID int64) (*domain.ChatSession, error) {
	query := `
		SELECT id, appointment_id, client_id, specialist_id, status, 
		       started_at, ended_at, created_at, updated_at
		FROM chat_sessions
		WHERE appointment_id = $1
	`

	var session domain.ChatSession
	err := r.db.QueryRow(ctx, query, appointmentID).Scan(
		&session.ID,
		&session.AppointmentID,
		&session.ClientID,
		&session.SpecialistID,
		&session.Status,
		&session.StartedAt,
		&session.EndedAt,
		&session.CreatedAt,
		&session.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get chat session: %w", err)
	}

	return &session, nil
}

func (r *ChatRepo) GetChatSessionsByUserID(ctx context.Context, userID int64, filter domain.ChatFilter) ([]domain.ChatSession, error) {
	conditions := []string{"(cs.client_id = $1 OR cs.specialist_id = $1)"}
	args := []interface{}{userID}
	argCount := 2

	if filter.Status != nil {
		conditions = append(conditions, fmt.Sprintf("cs.status = $%d", argCount))
		args = append(args, *filter.Status)
		argCount++
	}

	if filter.StartDate != nil {
		conditions = append(conditions, fmt.Sprintf("cs.created_at >= $%d", argCount))
		args = append(args, *filter.StartDate)
		argCount++
	}

	if filter.EndDate != nil {
		conditions = append(conditions, fmt.Sprintf("cs.created_at <= $%d", argCount))
		args = append(args, *filter.EndDate)
		argCount++
	}

	whereClause := "WHERE " + strings.Join(conditions, " AND ")
	args = append(args, filter.Limit, filter.Offset)

	query := fmt.Sprintf(`
		SELECT cs.id, cs.appointment_id, cs.client_id, cs.specialist_id, cs.status,
		       cs.started_at, cs.ended_at, cs.created_at, cs.updated_at,
		       u1.first_name, u1.last_name, u1.email,
		       u2.first_name, u2.last_name, u2.email
		FROM chat_sessions cs
		LEFT JOIN users u1 ON cs.client_id = u1.id
		LEFT JOIN users u2 ON cs.specialist_id = u2.id
		%s
		ORDER BY cs.updated_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argCount, argCount+1)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query chat sessions: %w", err)
	}
	defer rows.Close()

	var sessions []domain.ChatSession
	for rows.Next() {
		var session domain.ChatSession
		var client domain.User
		var specialist domain.User

		err := rows.Scan(
			&session.ID,
			&session.AppointmentID,
			&session.ClientID,
			&session.SpecialistID,
			&session.Status,
			&session.StartedAt,
			&session.EndedAt,
			&session.CreatedAt,
			&session.UpdatedAt,
			&client.FirstName,
			&client.LastName,
			&client.Email,
			&specialist.FirstName,
			&specialist.LastName,
			&specialist.Email,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan chat session: %w", err)
		}

		client.ID = session.ClientID
		specialist.ID = session.SpecialistID
		session.Client = &client
		session.Specialist = &specialist

		sessions = append(sessions, session)
	}

	return sessions, nil
}

func (r *ChatRepo) UpdateChatSessionStatus(ctx context.Context, id int64, status string) error {
	query := `UPDATE chat_sessions SET status = $1, updated_at = $2 WHERE id = $3`
	_, err := r.db.Exec(ctx, query, status, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update chat session status: %w", err)
	}
	return nil
}

func (r *ChatRepo) EndChatSession(ctx context.Context, id int64) error {
	query := `UPDATE chat_sessions SET status = $1, ended_at = $2, updated_at = $2 WHERE id = $3`
	now := time.Now()
	_, err := r.db.Exec(ctx, query, domain.ChatSessionStatusEnded, now, id)
	if err != nil {
		return fmt.Errorf("failed to end chat session: %w", err)
	}
	return nil
}

// Chat Messages Implementation

func (r *ChatRepo) CreateMessage(ctx context.Context, dto domain.CreateChatMessageDTO, senderID int64) (*domain.ChatMessage, error) {
	query := `
		INSERT INTO chat_messages (session_id, sender_id, message_type, content, file_url, file_name, file_size, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $8)
		RETURNING id, session_id, sender_id, message_type, content, file_url, file_name, file_size, is_read, read_at, created_at, updated_at
	`

	now := time.Now()
	var message domain.ChatMessage

	err := r.db.QueryRow(ctx, query,
		dto.SessionID,
		senderID,
		dto.MessageType,
		dto.Content,
		dto.FileURL,
		dto.FileName,
		dto.FileSize,
		now,
	).Scan(
		&message.ID,
		&message.SessionID,
		&message.SenderID,
		&message.MessageType,
		&message.Content,
		&message.FileURL,
		&message.FileName,
		&message.FileSize,
		&message.IsRead,
		&message.ReadAt,
		&message.CreatedAt,
		&message.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	return &message, nil
}

func (r *ChatRepo) GetMessageByID(ctx context.Context, id int64) (*domain.ChatMessage, error) {
	query := `
		SELECT cm.id, cm.session_id, cm.sender_id, cm.message_type, cm.content, 
		       cm.file_url, cm.file_name, cm.file_size, cm.is_read, cm.read_at, 
		       cm.created_at, cm.updated_at,
		       u.first_name, u.last_name, u.email
		FROM chat_messages cm
		LEFT JOIN users u ON cm.sender_id = u.id
		WHERE cm.id = $1
	`

	var message domain.ChatMessage
	var sender domain.User

	err := r.db.QueryRow(ctx, query, id).Scan(
		&message.ID,
		&message.SessionID,
		&message.SenderID,
		&message.MessageType,
		&message.Content,
		&message.FileURL,
		&message.FileName,
		&message.FileSize,
		&message.IsRead,
		&message.ReadAt,
		&message.CreatedAt,
		&message.UpdatedAt,
		&sender.FirstName,
		&sender.LastName,
		&sender.Email,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get message: %w", err)
	}

	sender.ID = message.SenderID
	message.Sender = &sender

	return &message, nil
}

func (r *ChatRepo) GetMessagesBySessionID(ctx context.Context, sessionID int64, filter domain.MessageFilter) ([]domain.ChatMessage, error) {
	conditions := []string{"cm.session_id = $1"}
	args := []interface{}{sessionID}
	argCount := 2

	if filter.SenderID != nil {
		conditions = append(conditions, fmt.Sprintf("cm.sender_id = $%d", argCount))
		args = append(args, *filter.SenderID)
		argCount++
	}

	if filter.MessageType != nil {
		conditions = append(conditions, fmt.Sprintf("cm.message_type = $%d", argCount))
		args = append(args, *filter.MessageType)
		argCount++
	}

	if filter.IsRead != nil {
		conditions = append(conditions, fmt.Sprintf("cm.is_read = $%d", argCount))
		args = append(args, *filter.IsRead)
		argCount++
	}

	if filter.StartDate != nil {
		conditions = append(conditions, fmt.Sprintf("cm.created_at >= $%d", argCount))
		args = append(args, *filter.StartDate)
		argCount++
	}

	if filter.EndDate != nil {
		conditions = append(conditions, fmt.Sprintf("cm.created_at <= $%d", argCount))
		args = append(args, *filter.EndDate)
		argCount++
	}

	whereClause := "WHERE " + strings.Join(conditions, " AND ")
	args = append(args, filter.Limit, filter.Offset)

	query := fmt.Sprintf(`
		SELECT cm.id, cm.session_id, cm.sender_id, cm.message_type, cm.content,
		       cm.file_url, cm.file_name, cm.file_size, cm.is_read, cm.read_at,
		       cm.created_at, cm.updated_at,
		       u.first_name, u.last_name, u.email
		FROM chat_messages cm
		LEFT JOIN users u ON cm.sender_id = u.id
		%s
		ORDER BY cm.created_at ASC
		LIMIT $%d OFFSET $%d
	`, whereClause, argCount, argCount+1)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query messages: %w", err)
	}
	defer rows.Close()

	var messages []domain.ChatMessage
	for rows.Next() {
		var message domain.ChatMessage
		var sender domain.User

		err := rows.Scan(
			&message.ID,
			&message.SessionID,
			&message.SenderID,
			&message.MessageType,
			&message.Content,
			&message.FileURL,
			&message.FileName,
			&message.FileSize,
			&message.IsRead,
			&message.ReadAt,
			&message.CreatedAt,
			&message.UpdatedAt,
			&sender.FirstName,
			&sender.LastName,
			&sender.Email,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}

		sender.ID = message.SenderID
		message.Sender = &sender
		messages = append(messages, message)
	}

	return messages, nil
}

func (r *ChatRepo) UpdateMessage(ctx context.Context, id int64, dto domain.UpdateChatMessageDTO) error {
	setParts := []string{}
	args := []interface{}{}
	argCount := 1

	if dto.Content != nil {
		setParts = append(setParts, fmt.Sprintf("content = $%d", argCount))
		args = append(args, *dto.Content)
		argCount++
	}

	if dto.IsRead != nil {
		setParts = append(setParts, fmt.Sprintf("is_read = $%d", argCount))
		args = append(args, *dto.IsRead)
		argCount++

		if *dto.IsRead {
			setParts = append(setParts, fmt.Sprintf("read_at = $%d", argCount))
			args = append(args, time.Now())
			argCount++
		}
	}

	if len(setParts) == 0 {
		return fmt.Errorf("no fields to update")
	}

	setParts = append(setParts, fmt.Sprintf("updated_at = $%d", argCount))
	args = append(args, time.Now())
	argCount++

	args = append(args, id)

	query := fmt.Sprintf("UPDATE chat_messages SET %s WHERE id = $%d", strings.Join(setParts, ", "), argCount)
	_, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update message: %w", err)
	}

	return nil
}

func (r *ChatRepo) MarkMessageAsRead(ctx context.Context, id int64, userID int64) error {
	query := `
		UPDATE chat_messages 
		SET is_read = true, read_at = $1, updated_at = $1 
		WHERE id = $2 AND sender_id != $3
	`
	_, err := r.db.Exec(ctx, query, time.Now(), id, userID)
	if err != nil {
		return fmt.Errorf("failed to mark message as read: %w", err)
	}
	return nil
}

func (r *ChatRepo) MarkAllMessagesAsRead(ctx context.Context, sessionID int64, userID int64) error {
	query := `
		UPDATE chat_messages 
		SET is_read = true, read_at = $1, updated_at = $1 
		WHERE session_id = $2 AND sender_id != $3 AND is_read = false
	`
	_, err := r.db.Exec(ctx, query, time.Now(), sessionID, userID)
	if err != nil {
		return fmt.Errorf("failed to mark all messages as read: %w", err)
	}
	return nil
}

func (r *ChatRepo) DeleteMessage(ctx context.Context, id int64) error {
	query := `DELETE FROM chat_messages WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete message: %w", err)
	}
	return nil
}

// Video Call Sessions Implementation

func (r *ChatRepo) CreateVideoCallSession(ctx context.Context, dto domain.CreateVideoCallSessionDTO, initiatorID int64) (*domain.VideoCallSession, error) {
	// Generate UUID for call session
	id := generateUUID()
	
	query := `
		INSERT INTO video_call_sessions (id, chat_session_id, initiator_id, call_type, status, started_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $7)
		RETURNING id, chat_session_id, initiator_id, call_type, status, started_at, connected_at, ended_at, duration, end_reason, quality, created_at, updated_at
	`

	now := time.Now()
	var session domain.VideoCallSession

	err := r.db.QueryRow(ctx, query,
		id,
		dto.ChatSessionID,
		initiatorID,
		dto.CallType,
		domain.VideoCallStatusWaiting,
		now,
		now,
	).Scan(
		&session.ID,
		&session.ChatSessionID,
		&session.InitiatorID,
		&session.CallType,
		&session.Status,
		&session.StartedAt,
		&session.ConnectedAt,
		&session.EndedAt,
		&session.Duration,
		&session.EndReason,
		&session.Quality,
		&session.CreatedAt,
		&session.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create video call session: %w", err)
	}

	return &session, nil
}

func (r *ChatRepo) GetVideoCallSessionByID(ctx context.Context, id string) (*domain.VideoCallSession, error) {
	query := `
		SELECT id, chat_session_id, initiator_id, call_type, status, started_at, 
		       connected_at, ended_at, duration, end_reason, quality, created_at, updated_at
		FROM video_call_sessions
		WHERE id = $1
	`

	var session domain.VideoCallSession
	err := r.db.QueryRow(ctx, query, id).Scan(
		&session.ID,
		&session.ChatSessionID,
		&session.InitiatorID,
		&session.CallType,
		&session.Status,
		&session.StartedAt,
		&session.ConnectedAt,
		&session.EndedAt,
		&session.Duration,
		&session.EndReason,
		&session.Quality,
		&session.CreatedAt,
		&session.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get video call session: %w", err)
	}

	return &session, nil
}

func (r *ChatRepo) GetVideoCallSessionsByChatID(ctx context.Context, chatSessionID int64) ([]domain.VideoCallSession, error) {
	query := `
		SELECT id, chat_session_id, initiator_id, call_type, status, started_at,
		       connected_at, ended_at, duration, end_reason, quality, created_at, updated_at
		FROM video_call_sessions
		WHERE chat_session_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, chatSessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to query video call sessions: %w", err)
	}
	defer rows.Close()

	var sessions []domain.VideoCallSession
	for rows.Next() {
		var session domain.VideoCallSession
		err := rows.Scan(
			&session.ID,
			&session.ChatSessionID,
			&session.InitiatorID,
			&session.CallType,
			&session.Status,
			&session.StartedAt,
			&session.ConnectedAt,
			&session.EndedAt,
			&session.Duration,
			&session.EndReason,
			&session.Quality,
			&session.CreatedAt,
			&session.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan video call session: %w", err)
		}

		sessions = append(sessions, session)
	}

	return sessions, nil
}

func (r *ChatRepo) UpdateVideoCallSession(ctx context.Context, id string, dto domain.UpdateVideoCallSessionDTO) error {
	setParts := []string{}
	args := []interface{}{}
	argCount := 1

	if dto.Status != nil {
		setParts = append(setParts, fmt.Sprintf("status = $%d", argCount))
		args = append(args, *dto.Status)
		argCount++
	}

	if dto.ConnectedAt != nil {
		setParts = append(setParts, fmt.Sprintf("connected_at = $%d", argCount))
		args = append(args, *dto.ConnectedAt)
		argCount++
	}

	if dto.EndedAt != nil {
		setParts = append(setParts, fmt.Sprintf("ended_at = $%d", argCount))
		args = append(args, *dto.EndedAt)
		argCount++
	}

	if dto.Duration != nil {
		setParts = append(setParts, fmt.Sprintf("duration = $%d", argCount))
		args = append(args, *dto.Duration)
		argCount++
	}

	if dto.EndReason != nil {
		setParts = append(setParts, fmt.Sprintf("end_reason = $%d", argCount))
		args = append(args, *dto.EndReason)
		argCount++
	}

	if dto.Quality != nil {
		setParts = append(setParts, fmt.Sprintf("quality = $%d", argCount))
		args = append(args, *dto.Quality)
		argCount++
	}

	if len(setParts) == 0 {
		return fmt.Errorf("no fields to update")
	}

	setParts = append(setParts, fmt.Sprintf("updated_at = $%d", argCount))
	args = append(args, time.Now())
	argCount++

	args = append(args, id)

	query := fmt.Sprintf("UPDATE video_call_sessions SET %s WHERE id = $%d", strings.Join(setParts, ", "), argCount)
	_, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update video call session: %w", err)
	}

	return nil
}

func (r *ChatRepo) EndVideoCallSession(ctx context.Context, id string, endReason string, quality *string) error {
	// Get the session to calculate duration
	session, err := r.GetVideoCallSessionByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get video call session: %w", err)
	}

	if session == nil {
		return fmt.Errorf("video call session not found")
	}

	now := time.Now()
	var duration int64

	if session.ConnectedAt != nil {
		duration = int64(now.Sub(*session.ConnectedAt).Seconds())
	} else {
		duration = int64(now.Sub(session.StartedAt).Seconds())
	}

	query := `
		UPDATE video_call_sessions 
		SET status = $1, ended_at = $2, duration = $3, end_reason = $4, quality = $5, updated_at = $2
		WHERE id = $6
	`

	_, err = r.db.Exec(ctx, query, domain.VideoCallStatusEnded, now, duration, endReason, quality, id)
	if err != nil {
		return fmt.Errorf("failed to end video call session: %w", err)
	}

	return nil
}

// Chat Participants Implementation

func (r *ChatRepo) AddParticipant(ctx context.Context, sessionID, userID int64, role string) error {
	query := `
		INSERT INTO chat_participants (session_id, user_id, role, joined_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $5)
		ON CONFLICT (session_id, user_id) DO UPDATE SET
		is_active = true, left_at = NULL, updated_at = EXCLUDED.updated_at
	`

	now := time.Now()
	_, err := r.db.Exec(ctx, query, sessionID, userID, role, now, now)
	if err != nil {
		return fmt.Errorf("failed to add participant: %w", err)
	}

	return nil
}

func (r *ChatRepo) RemoveParticipant(ctx context.Context, sessionID, userID int64) error {
	query := `
		UPDATE chat_participants 
		SET is_active = false, left_at = $1, updated_at = $1
		WHERE session_id = $2 AND user_id = $3
	`

	now := time.Now()
	_, err := r.db.Exec(ctx, query, now, sessionID, userID)
	if err != nil {
		return fmt.Errorf("failed to remove participant: %w", err)
	}

	return nil
}

func (r *ChatRepo) GetParticipantsBySessionID(ctx context.Context, sessionID int64) ([]domain.ChatParticipant, error) {
	query := `
		SELECT cp.id, cp.session_id, cp.user_id, cp.role, cp.joined_at, cp.left_at, 
		       cp.is_active, cp.created_at, cp.updated_at,
		       u.first_name, u.last_name, u.email
		FROM chat_participants cp
		LEFT JOIN users u ON cp.user_id = u.id
		WHERE cp.session_id = $1
		ORDER BY cp.joined_at ASC
	`

	rows, err := r.db.Query(ctx, query, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to query participants: %w", err)
	}
	defer rows.Close()

	var participants []domain.ChatParticipant
	for rows.Next() {
		var participant domain.ChatParticipant
		var user domain.User

		err := rows.Scan(
			&participant.ID,
			&participant.SessionID,
			&participant.UserID,
			&participant.Role,
			&participant.JoinedAt,
			&participant.LeftAt,
			&participant.IsActive,
			&participant.CreatedAt,
			&participant.UpdatedAt,
			&user.FirstName,
			&user.LastName,
			&user.Email,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan participant: %w", err)
		}

		user.ID = participant.UserID
		participant.User = &user
		participants = append(participants, participant)
	}

	return participants, nil
}

// Helper function to generate UUID (simplified)
func generateUUID() string {
	// In production, use a proper UUID library like github.com/google/uuid
	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), time.Now().Unix())
} 