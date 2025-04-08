package service

import (
	"context"
	"errors"

	"go.uber.org/zap"

	"laps/internal/domain"
	"laps/internal/repository"
)

type WorkExperienceServiceImpl struct {
	specialistRepo repository.SpecialistRepository
	logger         *zap.Logger
}

func NewWorkExperienceService(
	specialistRepo repository.SpecialistRepository,
	logger *zap.Logger,
) *WorkExperienceServiceImpl {
	return &WorkExperienceServiceImpl{
		specialistRepo: specialistRepo,
		logger:         logger,
	}
}

func (s *WorkExperienceServiceImpl) AddWorkExperience(ctx context.Context, specialistID int64, dto domain.WorkExperienceDTO) (int64, error) {
	_, err := s.specialistRepo.GetByID(ctx, specialistID)
	if err != nil {
		s.logger.Error("специалист не найден при добавлении опыта работы", zap.Int64("specialistID", specialistID), zap.Error(err))
		return 0, errors.New("специалист не найден")
	}

	id, err := s.specialistRepo.AddWorkExperience(ctx, specialistID, dto)
	if err != nil {
		s.logger.Error("ошибка добавления опыта работы", zap.Error(err))
		return 0, errors.New("ошибка при добавлении опыта работы")
	}

	return id, nil
}

func (s *WorkExperienceServiceImpl) UpdateWorkExperience(ctx context.Context, id int64, dto domain.WorkExperienceDTO) error {
	err := s.specialistRepo.UpdateWorkExperience(ctx, id, dto)
	if err != nil {
		s.logger.Error("ошибка обновления опыта работы", zap.Int64("id", id), zap.Error(err))
		return errors.New("ошибка при обновлении опыта работы")
	}

	return nil
}

func (s *WorkExperienceServiceImpl) DeleteWorkExperience(ctx context.Context, id int64) error {
	err := s.specialistRepo.DeleteWorkExperience(ctx, id)
	if err != nil {
		s.logger.Error("ошибка удаления опыта работы", zap.Int64("id", id), zap.Error(err))
		return errors.New("ошибка при удалении опыта работы")
	}

	return nil
}

func (s *WorkExperienceServiceImpl) GetWorkExperienceByID(ctx context.Context, id int64) (*domain.WorkPlace, error) {
	workplace, err := s.specialistRepo.GetWorkExperienceByID(ctx, id)
	if err != nil {
		s.logger.Error("ошибка получения опыта работы", zap.Int64("id", id), zap.Error(err))
		return nil, errors.New("опыт работы не найден")
	}

	return workplace, nil
}

func (s *WorkExperienceServiceImpl) GetWorkExperienceBySpecialistID(ctx context.Context, specialistID int64) ([]domain.WorkPlace, error) {
	_, err := s.specialistRepo.GetByID(ctx, specialistID)
	if err != nil {
		s.logger.Error("специалист не найден при получении опыта работы", zap.Int64("specialistID", specialistID), zap.Error(err))
		return nil, errors.New("специалист не найден")
	}

	workExperience, err := s.specialistRepo.GetWorkExperienceBySpecialistID(ctx, specialistID)
	if err != nil {
		s.logger.Error("ошибка при получении опыта работы специалиста", zap.Int64("specialistID", specialistID), zap.Error(err))
		return nil, err
	}

	return workExperience, nil
}
