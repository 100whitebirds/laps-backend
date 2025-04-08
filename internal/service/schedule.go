package service

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"go.uber.org/zap"

	"laps/internal/domain"
	"laps/internal/repository"
)

type ScheduleServiceImpl struct {
	repo           repository.ScheduleRepository
	specialistRepo repository.SpecialistRepository
	logger         *zap.Logger
}

func NewScheduleService(
	repo repository.ScheduleRepository,
	specialistRepo repository.SpecialistRepository,
	logger *zap.Logger,
) *ScheduleServiceImpl {
	return &ScheduleServiceImpl{
		repo:           repo,
		specialistRepo: specialistRepo,
		logger:         logger,
	}
}

func (s *ScheduleServiceImpl) Create(ctx context.Context, specialistID int64, dto domain.CreateScheduleDTO) (int64, error) {
	_, err := s.specialistRepo.GetByID(ctx, specialistID)
	if err != nil {
		s.logger.Error("ошибка при получении специалиста", zap.Error(err))
		return 0, errors.New("специалист не найден")
	}

	if dto.SlotTime < 10 || dto.SlotTime > 120 {
		s.logger.Error("недопустимая длительность слота", zap.Int("slot_time", dto.SlotTime))
		return 0, errors.New("длительность слота должна быть от 10 до 120 минут")
	}

	now := time.Now()
	startDate := now.AddDate(0, 0, -int(now.Weekday())+1)
	var lastID int64

	for i := 0; i < 7; i++ {
		currentDate := startDate.AddDate(0, 0, i)
		var daySchedule *domain.DaySchedule

		switch i {
		case 0:
			daySchedule = dto.WeekSchedule.Monday
		case 1:
			daySchedule = dto.WeekSchedule.Tuesday
		case 2:
			daySchedule = dto.WeekSchedule.Wednesday
		case 3:
			daySchedule = dto.WeekSchedule.Thursday
		case 4:
			daySchedule = dto.WeekSchedule.Friday
		case 5:
			daySchedule = dto.WeekSchedule.Saturday
		case 6:
			daySchedule = dto.WeekSchedule.Sunday
		}

		if daySchedule != nil && len(daySchedule.WorkTime) > 0 {
			for _, slot := range daySchedule.WorkTime {
				_, err = time.Parse("15:04", slot.StartTime)
				if err != nil {
					s.logger.Error("неверный формат времени начала", zap.Error(err))
					return 0, errors.New("неверный формат времени начала")
				}

				_, err = time.Parse("15:04", slot.EndTime)
				if err != nil {
					s.logger.Error("неверный формат времени окончания", zap.Error(err))
					return 0, errors.New("неверный формат времени окончания")
				}

				schedule := domain.Schedule{
					SpecialistID: specialistID,
					Date:         currentDate,
					StartTime:    slot.StartTime,
					EndTime:      slot.EndTime,
					SlotTime:     dto.SlotTime,
					CreatedAt:    time.Now(),
					UpdatedAt:    time.Now(),
				}

				id, err := s.repo.Create(ctx, schedule)
				if err != nil {
					s.logger.Error("ошибка создания расписания", zap.Error(err))
					return 0, fmt.Errorf("ошибка создания расписания: %w", err)
				}
				lastID = id
			}
		}
	}

	return lastID, nil
}

func (s *ScheduleServiceImpl) GetByID(ctx context.Context, id int64) (*domain.Schedule, error) {
	schedule, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("ошибка получения расписания", zap.Error(err))
		return nil, fmt.Errorf("ошибка получения расписания: %w", err)
	}
	return schedule, nil
}

func (s *ScheduleServiceImpl) Update(ctx context.Context, specialistID int64, dto domain.UpdateScheduleDTO) error {
	now := time.Now()
	startDate := now.AddDate(0, 0, -int(now.Weekday())+1)
	endDate := startDate.AddDate(0, 0, 6)

	filter := domain.ScheduleFilter{
		SpecialistID: &specialistID,
		StartDate:    &startDate,
		EndDate:      &endDate,
		Limit:        100,
		Offset:       0,
	}

	schedules, _, err := s.repo.List(ctx, filter)
	if err != nil {
		s.logger.Error("ошибка получения расписаний", zap.Error(err))
		return fmt.Errorf("ошибка получения расписаний: %w", err)
	}

	for _, schedule := range schedules {
		err = s.repo.Delete(ctx, schedule.ID)
		if err != nil {
			s.logger.Error("ошибка удаления расписания", zap.Error(err))
			return fmt.Errorf("ошибка удаления расписания: %w", err)
		}
	}

	slotTime := 30
	if dto.SlotTime != nil {
		slotTime = *dto.SlotTime
	}

	if slotTime < 10 || slotTime > 120 {
		s.logger.Error("недопустимая длительность слота", zap.Int("slot_time", slotTime))
		return errors.New("длительность слота должна быть от 10 до 120 минут")
	}

	for i := 0; i < 7; i++ {
		currentDate := startDate.AddDate(0, 0, i)
		var daySchedule *domain.DaySchedule

		switch i {
		case 0:
			daySchedule = dto.WeekSchedule.Monday
		case 1:
			daySchedule = dto.WeekSchedule.Tuesday
		case 2:
			daySchedule = dto.WeekSchedule.Wednesday
		case 3:
			daySchedule = dto.WeekSchedule.Thursday
		case 4:
			daySchedule = dto.WeekSchedule.Friday
		case 5:
			daySchedule = dto.WeekSchedule.Saturday
		case 6:
			daySchedule = dto.WeekSchedule.Sunday
		}

		if daySchedule != nil && len(daySchedule.WorkTime) > 0 {
			for _, slot := range daySchedule.WorkTime {
				_, err = time.Parse("15:04", slot.StartTime)
				if err != nil {
					s.logger.Error("неверный формат времени начала", zap.Error(err))
					return errors.New("неверный формат времени начала")
				}

				_, err = time.Parse("15:04", slot.EndTime)
				if err != nil {
					s.logger.Error("неверный формат времени окончания", zap.Error(err))
					return errors.New("неверный формат времени окончания")
				}

				schedule := domain.Schedule{
					SpecialistID: specialistID,
					Date:         currentDate,
					StartTime:    slot.StartTime,
					EndTime:      slot.EndTime,
					SlotTime:     slotTime,
					CreatedAt:    time.Now(),
					UpdatedAt:    time.Now(),
				}

				_, err := s.repo.Create(ctx, schedule)
				if err != nil {
					s.logger.Error("ошибка создания расписания", zap.Error(err))
					return fmt.Errorf("ошибка создания расписания: %w", err)
				}
			}
		}
	}

	return nil
}

func (s *ScheduleServiceImpl) Delete(ctx context.Context, id int64) error {
	err := s.repo.Delete(ctx, id)
	if err != nil {
		s.logger.Error("ошибка удаления расписания", zap.Error(err))
		return fmt.Errorf("ошибка удаления расписания: %w", err)
	}
	return nil
}

func (s *ScheduleServiceImpl) List(ctx context.Context, filter domain.ScheduleFilter) ([]domain.Schedule, int, error) {
	schedules, total, err := s.repo.List(ctx, filter)
	if err != nil {
		s.logger.Error("ошибка получения списка расписаний", zap.Error(err))
		return nil, 0, fmt.Errorf("ошибка получения списка расписаний: %w", err)
	}
	return schedules, total, nil
}

func (s *ScheduleServiceImpl) GetBySpecialistAndDate(ctx context.Context, specialistID int64, dateStr string) (*domain.Schedule, error) {
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		s.logger.Error("неверный формат даты", zap.Error(err))
		return nil, errors.New("неверный формат даты")
	}

	schedule, err := s.repo.GetBySpecialistAndDate(ctx, specialistID, date)
	if err != nil {
		s.logger.Error("ошибка получения расписания", zap.Error(err))
		return nil, fmt.Errorf("ошибка получения расписания: %w", err)
	}

	return schedule, nil
}

func (s *ScheduleServiceImpl) GenerateTimeSlots(ctx context.Context, specialistID int64, dateStr string) ([]string, error) {
	schedule, err := s.GetBySpecialistAndDate(ctx, specialistID, dateStr)
	if err != nil {
		return nil, err
	}

	if schedule == nil {
		return []string{}, nil
	}

	startTime, _ := time.Parse("15:04", schedule.StartTime)
	endTime, _ := time.Parse("15:04", schedule.EndTime)

	excludedSlots := make(map[string]bool)
	for _, excludeTime := range schedule.ExcludeTimes {
		excludedSlots[excludeTime] = true
	}

	var slots []string
	currentTime := startTime
	duration := time.Duration(schedule.SlotTime) * time.Minute

	for currentTime.Before(endTime) {
		timeStr := currentTime.Format("15:04")

		if !excludedSlots[timeStr] {
			slots = append(slots, timeStr)
		}

		currentTime = currentTime.Add(duration)
	}

	sort.Strings(slots)

	return slots, nil
}

func (s *ScheduleServiceImpl) GetWeekSchedule(ctx context.Context, specialistID int64, startDate time.Time) (*domain.WeekSchedule, int, error) {
	endDate := startDate.AddDate(0, 0, 6)

	filter := domain.ScheduleFilter{
		SpecialistID: &specialistID,
		StartDate:    &startDate,
		EndDate:      &endDate,
		Limit:        100,
		Offset:       0,
	}

	schedules, _, err := s.repo.List(ctx, filter)
	if err != nil {
		s.logger.Error("ошибка получения расписаний", zap.Error(err))
		return nil, 0, fmt.Errorf("ошибка получения расписаний: %w", err)
	}

	weekSchedule := domain.WeekSchedule{}
	var slotTime int

	schedulesByDay := make(map[int][]domain.Schedule)
	for _, schedule := range schedules {
		dayOfWeek := int(schedule.Date.Weekday())
		if dayOfWeek == 0 {
			dayOfWeek = 7
		}
		schedulesByDay[dayOfWeek] = append(schedulesByDay[dayOfWeek], schedule)
		slotTime = schedule.SlotTime
	}

	for day, daySchedules := range schedulesByDay {
		workTimeSlots := make([]domain.WorkTimeSlot, 0, len(daySchedules))
		for _, schedule := range daySchedules {
			workTimeSlots = append(workTimeSlots, domain.WorkTimeSlot{
				StartTime: schedule.StartTime,
				EndTime:   schedule.EndTime,
			})
		}

		daySchedule := &domain.DaySchedule{
			WorkTime: workTimeSlots,
		}

		switch day {
		case 1:
			weekSchedule.Monday = daySchedule
		case 2:
			weekSchedule.Tuesday = daySchedule
		case 3:
			weekSchedule.Wednesday = daySchedule
		case 4:
			weekSchedule.Thursday = daySchedule
		case 5:
			weekSchedule.Friday = daySchedule
		case 6:
			weekSchedule.Saturday = daySchedule
		case 7:
			weekSchedule.Sunday = daySchedule
		}
	}

	return &weekSchedule, slotTime, nil
}
