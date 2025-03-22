package domain

import (
	"time"
)

type User struct {
	ID           int64     `json:"id"`
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	MiddleName   string    `json:"middle_name,omitempty"`
	Email        string    `json:"email"`
	Phone        string    `json:"phone"`
	PasswordHash string    `json:"-"`
	Role         UserRole  `json:"role"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type UserRole string

const (
	UserRoleClient     UserRole = "client"
	UserRoleSpecialist UserRole = "specialist"
	UserRoleAdmin      UserRole = "admin"
)

type CreateUserDTO struct {
	FirstName  string   `json:"first_name" binding:"required"`
	LastName   string   `json:"last_name" binding:"required"`
	MiddleName string   `json:"middle_name"`
	Email      string   `json:"email" binding:"required,email"`
	Phone      string   `json:"phone" binding:"required"`
	Password   string   `json:"password" binding:"required,min=6"`
	Role       UserRole `json:"role" binding:"required,oneof=client specialist"`
}

type UpdateUserDTO struct {
	FirstName  *string `json:"first_name"`
	LastName   *string `json:"last_name"`
	MiddleName *string `json:"middle_name"`
	Email      *string `json:"email" binding:"omitempty,email"`
	Phone      *string `json:"phone"`
	IsActive   *bool   `json:"is_active"`
}

type AuthUserDTO struct {
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type PasswordUpdateDTO struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}
