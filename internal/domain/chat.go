package domain

import (
	"time"
)

// ChatSession represents a chat session between a client and specialist
type ChatSession struct {
	ID            int64     `json:"id"`
	AppointmentID int64     `json:"appointment_id"`
	ClientID      int64     `json:"client_id"`
	SpecialistID  int64     `json:"specialist_id"`
	Status        string    `json:"status"` // active, ended, pending
	StartedAt     time.Time `json:"started_at"`
	EndedAt       *time.Time `json:"ended_at,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	
	// Related entities (populated by joins)
	Appointment *Appointment `json:"appointment,omitempty"`
	Client      *User        `json:"client,omitempty"`
	Specialist  *User        `json:"specialist,omitempty"`
	LastMessage *ChatMessage `json:"last_message,omitempty"`
	Messages    []ChatMessage `json:"messages,omitempty"`
}

// ChatMessage represents a message in a chat session
type ChatMessage struct {
	ID            int64     `json:"id"`
	SessionID     int64     `json:"session_id"`
	SenderID      int64     `json:"sender_id"`
	MessageType   string    `json:"message_type"` // text, image, file, system
	Content       string    `json:"content"`
	FileURL       *string   `json:"file_url,omitempty"`
	FileName      *string   `json:"file_name,omitempty"`
	FileSize      *int64    `json:"file_size,omitempty"`
	IsRead        bool      `json:"is_read"`
	ReadAt        *time.Time `json:"read_at,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	
	// Related entities
	Sender  *User        `json:"sender,omitempty"`
	Session *ChatSession `json:"session,omitempty"`
}

// ChatParticipant represents a participant in a chat session
type ChatParticipant struct {
	ID        int64     `json:"id"`
	SessionID int64     `json:"session_id"`
	UserID    int64     `json:"user_id"`
	Role      string    `json:"role"` // client, specialist
	JoinedAt  time.Time `json:"joined_at"`
	LeftAt    *time.Time `json:"left_at,omitempty"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	
	// Related entities
	User    *User        `json:"user,omitempty"`
	Session *ChatSession `json:"session,omitempty"`
}

// VideoCallSession represents a video call session within a chat
type VideoCallSession struct {
	ID               string    `json:"id"`
	ChatSessionID    int64     `json:"chat_session_id"`
	InitiatorID      int64     `json:"initiator_id"`
	CallType         string    `json:"call_type"` // audio, video
	Status           string    `json:"status"`    // waiting, ringing, active, ended, rejected
	StartedAt        time.Time `json:"started_at"`
	ConnectedAt      *time.Time `json:"connected_at,omitempty"`
	EndedAt          *time.Time `json:"ended_at,omitempty"`
	Duration         *int64    `json:"duration,omitempty"` // in seconds
	EndReason        *string   `json:"end_reason,omitempty"` // user, network, error, timeout
	Quality          *string   `json:"quality,omitempty"`    // excellent, good, poor, failed
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	
	// Related entities
	ChatSession *ChatSession `json:"chat_session,omitempty"`
	Initiator   *User        `json:"initiator,omitempty"`
}

// CreateChatSessionDTO represents data for creating a new chat session
type CreateChatSessionDTO struct {
	AppointmentID int64 `json:"appointment_id" binding:"required"`
}

// CreateChatMessageDTO represents data for creating a new chat message
type CreateChatMessageDTO struct {
	SessionID   int64  `json:"session_id" binding:"required"`
	MessageType string `json:"message_type" binding:"required,oneof=text image file system"`
	Content     string `json:"content" binding:"required"`
	FileURL     *string `json:"file_url"`
	FileName    *string `json:"file_name"`
	FileSize    *int64  `json:"file_size"`
}

// UpdateChatMessageDTO represents data for updating a chat message
type UpdateChatMessageDTO struct {
	Content *string `json:"content"`
	IsRead  *bool   `json:"is_read"`
}

// CreateVideoCallSessionDTO represents data for creating a new video call session
type CreateVideoCallSessionDTO struct {
	ChatSessionID int64  `json:"chat_session_id" binding:"required"`
	CallType      string `json:"call_type" binding:"required,oneof=audio video"`
}

// UpdateVideoCallSessionDTO represents data for updating a video call session
type UpdateVideoCallSessionDTO struct {
	Status      *string    `json:"status" binding:"omitempty,oneof=waiting ringing active ended rejected"`
	ConnectedAt *time.Time `json:"connected_at"`
	EndedAt     *time.Time `json:"ended_at"`
	Duration    *int64     `json:"duration"`
	EndReason   *string    `json:"end_reason"`
	Quality     *string    `json:"quality"`
}

// ChatFilter represents filters for chat queries
type ChatFilter struct {
	UserID       *int64  `json:"user_id"`
	Status       *string `json:"status"`
	StartDate    *time.Time `json:"start_date"`
	EndDate      *time.Time `json:"end_date"`
	Limit        int     `json:"limit"`
	Offset       int     `json:"offset"`
}

// MessageFilter represents filters for message queries  
type MessageFilter struct {
	SessionID    *int64  `json:"session_id"`
	SenderID     *int64  `json:"sender_id"`
	MessageType  *string `json:"message_type"`
	IsRead       *bool   `json:"is_read"`
	StartDate    *time.Time `json:"start_date"`
	EndDate      *time.Time `json:"end_date"`
	Limit        int     `json:"limit"`
	Offset       int     `json:"offset"`
}

// ChatSessionStatus constants
const (
	ChatSessionStatusPending = "pending"
	ChatSessionStatusActive  = "active"
	ChatSessionStatusEnded   = "ended"
)

// MessageType constants
const (
	MessageTypeText   = "text"
	MessageTypeImage  = "image"
	MessageTypeFile   = "file"
	MessageTypeSystem = "system"
)

// VideoCallStatus constants
const (
	VideoCallStatusWaiting  = "waiting"
	VideoCallStatusRinging  = "ringing"
	VideoCallStatusActive   = "active"
	VideoCallStatusEnded    = "ended"
	VideoCallStatusRejected = "rejected"
)

// ParticipantRole constants
const (
	ParticipantRoleClient     = "client"
	ParticipantRoleSpecialist = "specialist"
) 