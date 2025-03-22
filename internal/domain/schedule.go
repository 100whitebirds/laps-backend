package domain

import (
	"time"
)

type Schedule struct {
	ID           int64     `json:"id"`
	SpecialistID int64     `json:"specialist_id"`
	Date         time.Time `json:"date"`
	StartTime    string    `json:"start_time"`
	EndTime      string    `json:"end_time"`
	SlotTime     int       `json:"slot_time"`
	ExcludeTimes []string  `json:"exclude_times"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type CreateScheduleDTO struct {
	Date         string   `json:"date" binding:"required"`
	StartTime    string   `json:"start_time" binding:"required"`
	EndTime      string   `json:"end_time" binding:"required"`
	SlotTime     int      `json:"slot_time" binding:"required"`
	ExcludeTimes []string `json:"exclude_times,omitempty"`
}

type UpdateScheduleDTO struct {
	StartTime    *string   `json:"start_time,omitempty"`
	EndTime      *string   `json:"end_time,omitempty"`
	SlotTime     *int      `json:"slot_time,omitempty"`
	ExcludeTimes *[]string `json:"exclude_times,omitempty"`
}

type ScheduleFilter struct {
	SpecialistID *int64     `json:"specialist_id"`
	StartDate    *time.Time `json:"start_date"`
	EndDate      *time.Time `json:"end_date"`
	Limit        int        `json:"limit"`
	Offset       int        `json:"offset"`
} 