package service

import (
	"context"

	"go.uber.org/zap"

	"laps/config"
	"laps/internal/domain"
	"laps/internal/repository"
	"laps/internal/storage"
)

type Deps struct {
	Repos       *repository.Repositories
	Logger      *zap.Logger
	Config      *config.Config
	FileStorage storage.FileStorage
}

type Services struct {
	User           UserService
	Auth           AuthService
	Specialist     SpecialistService
	Appointment    AppointmentService
	Review         ReviewService
	Specialization SpecializationService
	Schedule       ScheduleService
}

func NewServices(deps Deps) *Services {
	return &Services{
		User:           NewUserService(deps.Repos.User, deps.Logger),
		Auth:           NewAuthService(deps.Repos.Auth, deps.Repos.User, deps.Config.JWT, deps.Logger),
		Specialization: NewSpecializationService(deps.Repos.Specialization, deps.Logger),
		Review:         NewReviewService(deps.Repos.Review, deps.Repos.Specialist, deps.Repos.User, deps.Repos.Appointment, deps.Logger),
		Specialist:     NewSpecialistService(deps.Repos.Specialist, deps.Repos.User, deps.Repos.Specialization, deps.FileStorage, deps.Logger),
		Appointment:    NewAppointmentService(deps.Repos.Appointment, deps.Repos.Specialist, deps.Repos.User, deps.Logger),
		Schedule:       NewScheduleService(deps.Repos.Schedule, deps.Repos.Specialist, deps.Logger),
	}
}

type UserService interface {
	Create(ctx context.Context, dto domain.CreateUserDTO) (int64, error)
	GetByID(ctx context.Context, id int64) (*domain.User, error)
	Update(ctx context.Context, id int64, dto domain.UpdateUserDTO) error
	UpdatePassword(ctx context.Context, id int64, dto domain.PasswordUpdateDTO) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, limit, offset int) ([]domain.User, error)
}

type SpecialistService interface {
	Create(ctx context.Context, userID int64, dto domain.CreateSpecialistDTO) (int64, error)
	GetByID(ctx context.Context, id int64) (*domain.Specialist, error)
	GetByUserID(ctx context.Context, userID int64) (*domain.Specialist, error)
	Update(ctx context.Context, id int64, dto domain.UpdateSpecialistDTO) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, specialistType *domain.SpecialistType, limit, offset int) ([]domain.Specialist, error)

	UploadProfilePhoto(ctx context.Context, specialistID int64, photo []byte, filename string) error
	DeleteProfilePhoto(ctx context.Context, specialistID int64) error

	AddEducation(ctx context.Context, specialistID int64, dto domain.EducationDTO) (int64, error)
	UpdateEducation(ctx context.Context, id int64, dto domain.EducationDTO) error
	DeleteEducation(ctx context.Context, id int64) error
	GetEducationBySpecialistID(ctx context.Context, specialistID int64) ([]domain.Education, error)
	GetEducationByID(ctx context.Context, id int64) (*domain.Education, error)

	AddWorkExperience(ctx context.Context, specialistID int64, dto domain.WorkExperienceDTO) (int64, error)
	UpdateWorkExperience(ctx context.Context, id int64, dto domain.WorkExperienceDTO) error
	DeleteWorkExperience(ctx context.Context, id int64) error

	AddSpecialization(ctx context.Context, specialistID, specializationID int64) error
	RemoveSpecialization(ctx context.Context, specialistID, specializationID int64) error
	GetSpecializationsBySpecialistID(ctx context.Context, specialistID int64) ([]domain.Specialization, error)
}

type AppointmentService interface {
	Create(ctx context.Context, clientID int64, dto domain.CreateAppointmentDTO) (int64, error)
	GetByID(ctx context.Context, id int64) (*domain.Appointment, error)
	Update(ctx context.Context, id int64, dto domain.UpdateAppointmentDTO) error
	Cancel(ctx context.Context, id int64) error
	List(ctx context.Context, filter domain.AppointmentFilter) ([]domain.Appointment, int, error)
	GetFreeSlots(ctx context.Context, specialistID int64, date string) ([]string, error)
}

type ReviewService interface {
	Create(ctx context.Context, clientID int64, dto domain.CreateReviewDTO) (int64, error)
	GetByID(ctx context.Context, id int64) (*domain.Review, error)
	Update(ctx context.Context, id int64, dto domain.UpdateReviewDTO) error
	Delete(ctx context.Context, id int64) error
	GetBySpecialistID(ctx context.Context, specialistID int64, limit, offset int) ([]domain.Review, int, error)
	GetByUserID(ctx context.Context, userID int64, limit, offset int) ([]domain.Review, error)
	List(ctx context.Context, filter domain.ReviewFilter) ([]domain.Review, int, error)
	CreateReply(ctx context.Context, userID int64, reply domain.CreateReplyDTO) (int64, error)
	GetReplyByID(ctx context.Context, id int64) (*domain.Reply, error)
	DeleteReply(ctx context.Context, replyID int64) error
	GetRepliesByReviewID(ctx context.Context, reviewID int64) ([]domain.Reply, error)
}

type SpecializationService interface {
	Create(ctx context.Context, dto domain.CreateSpecializationDTO) (int64, error)
	GetByID(ctx context.Context, id int64) (*domain.Specialization, error)
	Update(ctx context.Context, id int64, dto domain.UpdateSpecializationDTO) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, filter domain.SpecializationFilter) ([]domain.Specialization, int, error)
}

type ScheduleService interface {
	Create(ctx context.Context, specialistID int64, dto domain.CreateScheduleDTO) (int64, error)
	GetByID(ctx context.Context, id int64) (*domain.Schedule, error)
	Update(ctx context.Context, id int64, dto domain.UpdateScheduleDTO) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, filter domain.ScheduleFilter) ([]domain.Schedule, int, error)
	GetBySpecialistAndDate(ctx context.Context, specialistID int64, date string) (*domain.Schedule, error)
	GenerateTimeSlots(ctx context.Context, specialistID int64, date string) ([]string, error)
}

type AuthService interface {
	Register(ctx context.Context, dto domain.RegisterRequest) (int64, error)
	Login(ctx context.Context, dto domain.LoginRequest, userAgent, ip string) (*domain.Tokens, error)
	RefreshTokens(ctx context.Context, refreshToken, userAgent, ip string) (*domain.Tokens, error)
	Logout(ctx context.Context, refreshToken string) error
	ParseToken(ctx context.Context, token string) (int64, domain.UserRole, error)
}
