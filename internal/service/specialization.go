package service

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"

	"laps/internal/domain"
	"laps/internal/repository"
)

type SpecializationServiceImpl struct {
	repo   repository.SpecializationRepository
	logger *zap.Logger
}

func NewSpecializationService(repo repository.SpecializationRepository, logger *zap.Logger) *SpecializationServiceImpl {
	return &SpecializationServiceImpl{
		repo:   repo,
		logger: logger,
	}
}

func (s *SpecializationServiceImpl) Create(ctx context.Context, dto domain.CreateSpecializationDTO) (int64, error) {
	id, err := s.repo.Create(ctx, dto)
	if err != nil {
		s.logger.Error("ошибка создания специализации", zap.Error(err))
		return 0, errors.New("ошибка при создании специализации")
	}

	return id, nil
}

func (s *SpecializationServiceImpl) GetByID(ctx context.Context, id int64) (*domain.Specialization, error) {
	specialization, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("ошибка получения специализации", zap.Int64("id", id), zap.Error(err))
		return nil, errors.New("специализация не найдена")
	}

	return specialization, nil
}

func (s *SpecializationServiceImpl) Update(ctx context.Context, id int64, dto domain.UpdateSpecializationDTO) error {
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("специализация для обновления не найдена", zap.Int64("id", id), zap.Error(err))
		return errors.New("специализация не найдена")
	}

	err = s.repo.Update(ctx, id, dto)
	if err != nil {
		s.logger.Error("ошибка обновления специализации", zap.Int64("id", id), zap.Error(err))
		return errors.New("ошибка при обновлении специализации")
	}

	return nil
}

func (s *SpecializationServiceImpl) Delete(ctx context.Context, id int64) error {
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("специализация для удаления не найдена", zap.Int64("id", id), zap.Error(err))
		return errors.New("специализация не найдена")
	}

	err = s.repo.Delete(ctx, id)
	if err != nil {
		s.logger.Error("ошибка удаления специализации", zap.Int64("id", id), zap.Error(err))
		return errors.New("ошибка при удалении специализации")
	}

	return nil
}

func (s *SpecializationServiceImpl) List(ctx context.Context, filter domain.SpecializationFilter) ([]domain.Specialization, int, error) {
	total, err := s.repo.CountByFilter(ctx, filter)
	if err != nil {
		s.logger.Error("ошибка подсчета специализаций", zap.Error(err))
		return nil, 0, fmt.Errorf("ошибка при получении списка специализаций: %w", err)
	}

	specializations, err := s.repo.List(ctx, filter)
	if err != nil {
		s.logger.Error("ошибка получения списка специализаций", zap.Error(err))
		return nil, 0, fmt.Errorf("ошибка при получении списка специализаций: %w", err)
	}

	return specializations, total, nil
}
