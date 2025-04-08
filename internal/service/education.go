package service

import (
	"context"
	"errors"

	"go.uber.org/zap"

	"laps/internal/domain"
	"laps/internal/repository"
)

type EducationServiceImpl struct {
	specialistRepo repository.SpecialistRepository
	logger         *zap.Logger
}

func NewEducationService(
	specialistRepo repository.SpecialistRepository,
	logger *zap.Logger,
) *EducationServiceImpl {
	return &EducationServiceImpl{
		specialistRepo: specialistRepo,
		logger:         logger,
	}
}

func (s *EducationServiceImpl) AddEducation(ctx context.Context, specialistID int64, dto domain.EducationDTO) (int64, error) {
	_, err := s.specialistRepo.GetByID(ctx, specialistID)
	if err != nil {
		s.logger.Error("специалист не найден при добавлении образования", zap.Int64("specialistID", specialistID), zap.Error(err))
		return 0, errors.New("специалист не найден")
	}

	id, err := s.specialistRepo.AddEducation(ctx, specialistID, dto)
	if err != nil {
		s.logger.Error("ошибка добавления образования", zap.Error(err))
		return 0, errors.New("ошибка при добавлении образования")
	}

	return id, nil
}

func (s *EducationServiceImpl) UpdateEducation(ctx context.Context, id int64, dto domain.EducationDTO) error {
	err := s.specialistRepo.UpdateEducation(ctx, id, dto)
	if err != nil {
		s.logger.Error("ошибка обновления образования", zap.Int64("id", id), zap.Error(err))
		return errors.New("ошибка при обновлении образования")
	}

	return nil
}

func (s *EducationServiceImpl) DeleteEducation(ctx context.Context, id int64) error {
	err := s.specialistRepo.DeleteEducation(ctx, id)
	if err != nil {
		s.logger.Error("ошибка удаления образования", zap.Int64("id", id), zap.Error(err))
		return errors.New("ошибка при удалении образования")
	}

	return nil
}

func (s *EducationServiceImpl) GetEducationBySpecialistID(ctx context.Context, specialistID int64) ([]domain.Education, error) {
	_, err := s.specialistRepo.GetByID(ctx, specialistID)
	if err != nil {
		s.logger.Error("специалист не найден при получении образования", zap.Int64("specialistID", specialistID), zap.Error(err))
		return nil, errors.New("специалист не найден")
	}

	education, err := s.specialistRepo.GetEducationBySpecialistID(ctx, specialistID)
	if err != nil {
		s.logger.Error("ошибка при получении образования специалиста", zap.Int64("specialistID", specialistID), zap.Error(err))
		return nil, err
	}

	return education, nil
}

func (s *EducationServiceImpl) GetEducationByID(ctx context.Context, id int64) (*domain.Education, error) {
	education, err := s.specialistRepo.GetEducationByID(ctx, id)
	if err != nil {
		s.logger.Error("ошибка при получении образования", zap.Int64("id", id), zap.Error(err))
		return nil, err
	}

	return education, nil
}
