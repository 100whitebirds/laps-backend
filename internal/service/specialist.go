package service

import (
	"context"
	"errors"

	"go.uber.org/zap"

	"laps/internal/domain"
	"laps/internal/repository"
)

type SpecialistServiceImpl struct {
	repo     repository.SpecialistRepository
	userRepo repository.UserRepository
	specRepo repository.SpecializationRepository
	logger   *zap.Logger
}

func NewSpecialistService(
	repo repository.SpecialistRepository,
	userRepo repository.UserRepository,
	specRepo repository.SpecializationRepository,
	logger *zap.Logger,
) *SpecialistServiceImpl {
	return &SpecialistServiceImpl{
		repo:     repo,
		userRepo: userRepo,
		specRepo: specRepo,
		logger:   logger,
	}
}

func (s *SpecialistServiceImpl) Create(ctx context.Context, userID int64, dto domain.CreateSpecialistDTO) (int64, error) {
	_, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		s.logger.Error("пользователь не найден при создании специалиста", zap.Int64("userID", userID), zap.Error(err))
		return 0, errors.New("пользователь не найден")
	}

	_, err = s.repo.GetByUserID(ctx, userID)
	if err == nil {
		s.logger.Error("пользователь уже зарегистрирован как специалист", zap.Int64("userID", userID))
		return 0, errors.New("пользователь уже зарегистрирован как специалист")
	}

	if !dto.Type.IsValid() {
		s.logger.Error("некорректный тип специалиста", zap.String("type", string(dto.Type)))
		return 0, errors.New("некорректный тип специалиста")
	}

	id, err := s.repo.Create(ctx, userID, dto)
	if err != nil {
		s.logger.Error("ошибка создания специалиста", zap.Error(err))
		return 0, errors.New("ошибка при создании специалиста")
	}

	return id, nil
}

func (s *SpecialistServiceImpl) GetByID(ctx context.Context, id int64) (*domain.Specialist, error) {
	specialist, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("ошибка получения специалиста", zap.Int64("id", id), zap.Error(err))
		return nil, errors.New("специалист не найден")
	}
	return specialist, nil
}

func (s *SpecialistServiceImpl) GetByUserID(ctx context.Context, userID int64) (*domain.Specialist, error) {
	specialist, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		s.logger.Error("ошибка получения специалиста по ID пользователя", zap.Int64("userID", userID), zap.Error(err))
		return nil, errors.New("специалист не найден")
	}
	return specialist, nil
}

func (s *SpecialistServiceImpl) Update(ctx context.Context, id int64, dto domain.UpdateSpecialistDTO) error {
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("специалист для обновления не найден", zap.Int64("id", id), zap.Error(err))
		return errors.New("специалист не найден")
	}

	if dto.Type != nil && !dto.Type.IsValid() {
		s.logger.Error("некорректный тип специалиста", zap.String("type", string(*dto.Type)))
		return errors.New("некорректный тип специалиста")
	}

	err = s.repo.Update(ctx, id, dto)
	if err != nil {
		s.logger.Error("ошибка обновления специалиста", zap.Int64("id", id), zap.Error(err))
		return errors.New("ошибка при обновлении специалиста")
	}

	return nil
}

func (s *SpecialistServiceImpl) Delete(ctx context.Context, id int64) error {
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("специалист для удаления не найден", zap.Int64("id", id), zap.Error(err))
		return errors.New("специалист не найден")
	}

	err = s.repo.Delete(ctx, id)
	if err != nil {
		s.logger.Error("ошибка удаления специалиста", zap.Int64("id", id), zap.Error(err))
		return errors.New("ошибка при удалении специалиста")
	}

	return nil
}

func (s *SpecialistServiceImpl) List(ctx context.Context, specialistType *domain.SpecialistType, limit, offset int) ([]domain.Specialist, error) {
	if specialistType != nil && !specialistType.IsValid() {
		s.logger.Error("некорректный тип специалиста", zap.String("type", string(*specialistType)))
		return nil, errors.New("некорректный тип специалиста")
	}

	specialists, err := s.repo.List(ctx, specialistType, limit, offset)
	if err != nil {
		s.logger.Error("ошибка получения списка специалистов", zap.Error(err))
		return nil, errors.New("ошибка при получении списка специалистов")
	}

	return specialists, nil
}


func (s *SpecialistServiceImpl) AddEducation(ctx context.Context, specialistID int64, dto domain.EducationDTO) (int64, error) {
	_, err := s.repo.GetByID(ctx, specialistID)
	if err != nil {
		s.logger.Error("специалист не найден при добавлении образования", zap.Int64("specialistID", specialistID), zap.Error(err))
		return 0, errors.New("специалист не найден")
	}

	id, err := s.repo.AddEducation(ctx, specialistID, dto)
	if err != nil {
		s.logger.Error("ошибка добавления образования", zap.Error(err))
		return 0, errors.New("ошибка при добавлении образования")
	}

	return id, nil
}

func (s *SpecialistServiceImpl) UpdateEducation(ctx context.Context, id int64, dto domain.EducationDTO) error {
	err := s.repo.UpdateEducation(ctx, id, dto)
	if err != nil {
		s.logger.Error("ошибка обновления образования", zap.Int64("id", id), zap.Error(err))
		return errors.New("ошибка при обновлении образования")
	}

	return nil
}

func (s *SpecialistServiceImpl) DeleteEducation(ctx context.Context, id int64) error {
	err := s.repo.DeleteEducation(ctx, id)
	if err != nil {
		s.logger.Error("ошибка удаления образования", zap.Int64("id", id), zap.Error(err))
		return errors.New("ошибка при удалении образования")
	}

	return nil
}

func (s *SpecialistServiceImpl) AddWorkExperience(ctx context.Context, specialistID int64, dto domain.WorkExperienceDTO) (int64, error) {
	_, err := s.repo.GetByID(ctx, specialistID)
	if err != nil {
		s.logger.Error("специалист не найден при добавлении опыта работы", zap.Int64("specialistID", specialistID), zap.Error(err))
		return 0, errors.New("специалист не найден")
	}

	id, err := s.repo.AddWorkExperience(ctx, specialistID, dto)
	if err != nil {
		s.logger.Error("ошибка добавления опыта работы", zap.Error(err))
		return 0, errors.New("ошибка при добавлении опыта работы")
	}

	return id, nil
}

func (s *SpecialistServiceImpl) UpdateWorkExperience(ctx context.Context, id int64, dto domain.WorkExperienceDTO) error {
	err := s.repo.UpdateWorkExperience(ctx, id, dto)
	if err != nil {
		s.logger.Error("ошибка обновления опыта работы", zap.Int64("id", id), zap.Error(err))
		return errors.New("ошибка при обновлении опыта работы")
	}

	return nil
}

func (s *SpecialistServiceImpl) DeleteWorkExperience(ctx context.Context, id int64) error {
	err := s.repo.DeleteWorkExperience(ctx, id)
	if err != nil {
		s.logger.Error("ошибка удаления опыта работы", zap.Int64("id", id), zap.Error(err))
		return errors.New("ошибка при удалении опыта работы")
	}

	return nil
}

func (s *SpecialistServiceImpl) AddSpecialization(ctx context.Context, specialistID, specializationID int64) error {
	_, err := s.repo.GetByID(ctx, specialistID)
	if err != nil {
		s.logger.Error("специалист не найден при добавлении специализации", zap.Int64("specialistID", specialistID), zap.Error(err))
		return errors.New("специалист не найден")
	}

	_, err = s.specRepo.GetByID(ctx, specializationID)
	if err != nil {
		s.logger.Error("специализация не найдена", zap.Int64("specializationID", specializationID), zap.Error(err))
		return errors.New("специализация не найдена")
	}

	err = s.repo.AddSpecialization(ctx, specialistID, specializationID)
	if err != nil {
		s.logger.Error("ошибка добавления специализации", zap.Error(err))
		return errors.New("ошибка при добавлении специализации")
	}

	return nil
}

func (s *SpecialistServiceImpl) RemoveSpecialization(ctx context.Context, specialistID, specializationID int64) error {
	_, err := s.repo.GetByID(ctx, specialistID)
	if err != nil {
		s.logger.Error("специалист не найден при удалении специализации", zap.Int64("specialistID", specialistID), zap.Error(err))
		return errors.New("специалист не найден")
	}

	err = s.repo.RemoveSpecialization(ctx, specialistID, specializationID)
	if err != nil {
		s.logger.Error("ошибка удаления специализации", zap.Error(err))
		return errors.New("ошибка при удалении специализации")
	}

	return nil
}

func (s *SpecialistServiceImpl) GetSpecializationsBySpecialistID(ctx context.Context, specialistID int64) ([]domain.Specialization, error) {
	_, err := s.repo.GetByID(ctx, specialistID)
	if err != nil {
		s.logger.Error("специалист не найден при получении специализаций", zap.Int64("specialistID", specialistID), zap.Error(err))
		return nil, errors.New("специалист не найден")
	}

	specializations, err := s.repo.GetSpecializationsBySpecialistID(ctx, specialistID)
	if err != nil {
		s.logger.Error("ошибка получения специализаций", zap.Error(err))
		return nil, errors.New("ошибка при получении специализаций")
	}

	return specializations, nil
}

func (s *SpecialistServiceImpl) GetEducationBySpecialistID(ctx context.Context, specialistID int64) ([]domain.Education, error) {
	_, err := s.repo.GetByID(ctx, specialistID)
	if err != nil {
		s.logger.Error("специалист не найден при получении образования", zap.Int64("specialistID", specialistID), zap.Error(err))
		return nil, errors.New("специалист не найден")
	}

	education, err := s.repo.GetEducationBySpecialistID(ctx, specialistID)
	if err != nil {
		s.logger.Error("ошибка при получении образования специалиста", zap.Int64("specialistID", specialistID), zap.Error(err))
		return nil, err
	}

	return education, nil
}

func (s *SpecialistServiceImpl) GetEducationByID(ctx context.Context, id int64) (*domain.Education, error) {
	education, err := s.repo.GetEducationByID(ctx, id)
	if err != nil {
		s.logger.Error("ошибка при получении образования", zap.Int64("id", id), zap.Error(err))
		return nil, err
	}

	return education, nil
}
