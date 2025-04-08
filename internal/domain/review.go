package domain

import (
	"time"
)

type Review struct {
	ID            int64  `json:"id"`
	ClientID      int64  `json:"client_id"`
	SpecialistID  int64  `json:"specialist_id"`
	AppointmentID int64  `json:"appointment_id"`
	Rating        int    `json:"rating"`
	Text          string `json:"text"`
	IsRecommended bool   `json:"is_recommended"`

	ServiceRating        *int `json:"service_rating"`
	MeetingEfficiency    *int `json:"meeting_efficiency"`
	Professionalism      *int `json:"professionalism"`
	PriceQuality         *int `json:"price_quality"`
	Cleanliness          *int `json:"cleanliness"`
	Attentiveness        *int `json:"attentiveness"`
	SpecialistExperience *int `json:"specialist_experience"`
	Grammar              *int `json:"grammar"`

	ReplyID   *int64    `json:"reply_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Reply struct {
	ID        int64     `json:"id"`
	ReviewID  int64     `json:"review_id"`
	UserID    int64     `json:"user_id"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateReviewDTO struct {
	SpecialistID  int64  `json:"specialist_id" binding:"required"`
	AppointmentID int64  `json:"appointment_id" binding:"required"`
	Rating        int    `json:"rating" binding:"required,min=1,max=5"`
	Text          string `json:"text" binding:"required"`
	IsRecommended bool   `json:"is_recommended"`

	ServiceRating        *int `json:"service_rating" binding:"omitempty,min=1,max=5"`
	MeetingEfficiency    *int `json:"meeting_efficiency" binding:"omitempty,min=1,max=5"`
	Professionalism      *int `json:"professionalism" binding:"omitempty,min=1,max=5"`
	PriceQuality         *int `json:"price_quality" binding:"omitempty,min=1,max=5"`
	Cleanliness          *int `json:"cleanliness" binding:"omitempty,min=1,max=5"`
	Attentiveness        *int `json:"attentiveness" binding:"omitempty,min=1,max=5"`
	SpecialistExperience *int `json:"specialist_experience" binding:"omitempty,min=1,max=5"`
	Grammar              *int `json:"grammar" binding:"omitempty,min=1,max=5"`
}

type CreateReplyDTO struct {
	ReviewID int64  `json:"review_id" binding:"required"`
	Text     string `json:"text" binding:"required"`
}

type UpdateReviewDTO struct {
	Rating *int    `json:"rating" binding:"omitempty,min=1,max=5"`
	Text   *string `json:"text" binding:"omitempty"`
}

type ReviewFilter struct {
	SpecialistID *int64 `json:"specialist_id"`
	ClientID     *int64 `json:"client_id"`
	MinRating    *int   `json:"min_rating"`
	MaxRating    *int   `json:"max_rating"`
	Limit        int    `json:"limit"`
	Offset       int    `json:"offset"`
}
