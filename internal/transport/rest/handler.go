package rest

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"laps/config"
	"laps/internal/domain"
	"laps/internal/service"
)

type Handler struct {
	services *service.Services
	logger   *zap.Logger
	config   *config.Config
}

func NewHandler(services *service.Services, logger *zap.Logger, config *config.Config) *Handler {
	return &Handler{
		services: services,
		logger:   logger,
		config:   config,
	}
}

func (h *Handler) InitRoutes(router *gin.Engine) {
	router.Use(h.loggerMiddleware())

	router.Use(h.errorMiddleware())

	router.Use(h.corsMiddleware())

	api := router.Group("/api/v1")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", h.register)
			auth.POST("/login", h.login)
			auth.POST("/refresh", h.refreshTokens)
			auth.POST("/logout", h.logout)
		}

		users := api.Group("/users")
		users.Use(h.authMiddleware())
		{
			users.GET("/me", h.getCurrentUser)
			users.GET("/:id", h.getUserByID)
			users.PUT("/:id", h.updateUser)
			users.PUT("/:id/password", h.updatePassword)

			admin := users.Group("/")
			admin.Use(h.adminMiddleware())
			{
				admin.POST("/", h.createUser)
				admin.GET("/", h.getUsers)
				admin.DELETE("/:id", h.deleteUser)
			}
		}

		specialists := api.Group("/specialists")
		{
			specialists.GET("/", h.getSpecialists)
			specialists.GET("/:id", h.getSpecialistByID)
			specialists.GET("/:id/reviews", h.getSpecialistReviewsRedirect)
			specialists.GET("/me", h.authMiddleware(), h.getMySpecialistProfile)

			auth := specialists.Group("/", h.authMiddleware())
			{
				auth.POST("/", h.createSpecialist)
				auth.PUT("/:id", h.updateSpecialist)
				auth.DELETE("/:id", h.deleteSpecialist)

				auth.PUT("/:id/education/:eduId", h.updateSpecialistEducation)
				auth.DELETE("/:id/education/:eduId", h.deleteSpecialistEducation)

				auth.PUT("/:id/work-experience/:expId", h.updateSpecialistWorkExperience)
				auth.DELETE("/:id/work-experience/:expId", h.deleteSpecialistWorkExperience)

				auth.POST("/:id/specializations/:specId", h.addSpecialistSpecialization)
				auth.DELETE("/:id/specializations/:specId", h.removeSpecialistSpecialization)

				specialistRoutes := auth.Group("/specialist-actions")
				specialistRoutes.Use(h.specialistMiddleware())
				{
					specialistRoutes.GET("/appointments", h.getSpecialistAppointments)
				}

				auth.POST("/:id/photo", h.uploadSpecialistPhoto)
				auth.DELETE("/:id/photo", h.deleteSpecialistPhoto)
			}
		}

		h.initScheduleRoutes(api)

		appointments := api.Group("/appointments")
		{

			auth := appointments.Group("/")
			auth.Use(h.authMiddleware())
			{
				auth.POST("/", h.createAppointment)
				auth.GET("/:id", h.getAppointmentByID)
				auth.PUT("/:id", h.updateAppointment)
				auth.DELETE("/:id", h.cancelAppointment)
				auth.GET("/", h.getAppointments)
			}
		}

		reviews := api.Group("/reviews")
		{
			reviews.GET("/", h.getReviews)
			reviews.GET("/:id", h.getReviewByID)

			auth := reviews.Group("/")
			auth.Use(h.authMiddleware())
			{
				auth.POST("/", h.createReview)
				auth.DELETE("/:id", h.deleteReview)
				auth.POST("/:id/replies", h.createReviewReply)
				auth.DELETE("/replies/:replyId", h.deleteReviewReply)
			}
		}

		specializations := api.Group("/specializations")
		{
			specializations.GET("/", h.getSpecializations)
			specializations.GET("/:id", h.getSpecializationByID)

			admin := specializations.Group("/")
			admin.Use(h.authMiddleware(), h.adminMiddleware())
			{
				admin.POST("/", h.createSpecialization)
				admin.PUT("/:id", h.updateSpecialization)
				admin.DELETE("/:id", h.deleteSpecialization)
			}
		}

		education := api.Group("/education")
		{
			education.GET("/", h.getEducation)
			education.GET("/:id", h.getEducationByID)

			auth := education.Group("/")
			auth.Use(h.authMiddleware())
			{
				auth.POST("/", h.addEducation)
				auth.PUT("/:id", h.updateEducation)
				auth.DELETE("/:id", h.deleteEducation)
			}
		}

		workExperience := api.Group("/work-experience")
		{
			workExperience.GET("/", h.getWorkExperience)
			workExperience.GET("/:id", h.getWorkExperienceByID)

			auth := workExperience.Group("/")
			auth.Use(h.authMiddleware())
			{
				auth.POST("/", h.addWorkExperience)
				auth.PUT("/:id", h.updateWorkExperience)
				auth.DELETE("/:id", h.deleteWorkExperience)
			}
		}

		// REST compliant routes for specialists
		specialists.POST("/:id/work-experience", h.authMiddleware(), h.addWorkExperienceToSpecialist)
		specialists.POST("/:id/education", h.authMiddleware(), h.addEducationToSpecialist)
	}
}

func (h *Handler) initScheduleRoutes(api *gin.RouterGroup) {
	schedules := api.Group("/schedules")
	{
		schedules.GET("/free-slots", h.getFreeSlots)
		schedules.GET("/week", h.getScheduleWeek)
		schedules.GET("/", h.getSchedules)
		schedules.GET("/:id", h.getScheduleByID)

		auth := schedules.Group("/", h.authMiddleware())
		{
			specialistRoutes := auth.Group("/", h.specialistMiddleware())
			{
				specialistRoutes.POST("/", h.createSchedule)
				specialistRoutes.PUT("/", h.updateSchedule)
				specialistRoutes.DELETE("/:id", h.deleteSchedule)
			}
		}
	}
}

func (h *Handler) getSpecialistAppointments(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		unauthorizedResponse(c)
		return
	}

	specialist, err := h.services.Specialist.GetByUserID(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("ошибка при получении данных специалиста", zap.Error(err))
		notFoundResponse(c, "профиль специалиста не найден")
		return
	}

	statusStr := c.DefaultQuery("status", "")
	var status *domain.AppointmentStatus
	if statusStr != "" {
		appStatus := domain.AppointmentStatus(statusStr)
		status = &appStatus
	}

	dateFrom := c.DefaultQuery("date_from", "")
	var startDate *time.Time
	if dateFrom != "" {
		parsedDate, err := time.Parse("2006-01-02", dateFrom)
		if err == nil {
			startDate = &parsedDate
		}
	}

	dateTo := c.DefaultQuery("date_to", "")
	var endDate *time.Time
	if dateTo != "" {
		parsedDate, err := time.Parse("2006-01-02", dateTo)
		if err == nil {
			parsedDate = parsedDate.Add(24 * time.Hour).Add(-time.Second)
			endDate = &parsedDate
		}
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if err != nil || limit < 0 {
		limit = 20
	}
	offset, err := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if err != nil || offset < 0 {
		offset = 0
	}

	filter := domain.AppointmentFilter{
		SpecialistID: &specialist.ID,
		Status:       status,
		StartDate:    startDate,
		EndDate:      endDate,
		Limit:        limit,
		Offset:       offset,
	}

	appointments, total, err := h.services.Appointment.List(c.Request.Context(), filter)
	if err != nil {
		h.logger.Error("ошибка при получении записей", zap.Error(err))
		errorResponse(c, http.StatusInternalServerError, "ошибка при получении записей")
		return
	}

	page := offset/limit + 1

	paginatedSuccessResponse(c, appointments, total, page, limit)
}
