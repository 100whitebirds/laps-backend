package service

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"laps/internal/domain"
	"laps/internal/repository"
)

type UserServiceImpl struct {
	repo   repository.UserRepository
	logger *zap.Logger
}

func NewUserService(repo repository.UserRepository, logger *zap.Logger) *UserServiceImpl {
	return &UserServiceImpl{
		repo:   repo,
		logger: logger,
	}
}

func (s *UserServiceImpl) Create(ctx context.Context, dto domain.CreateUserDTO) (int64, error) {
	existingUser, err := s.repo.GetByEmail(ctx, dto.Email)
	if err == nil && existingUser != nil {
		return 0, errors.New("пользователь с таким email уже существует")
	}

	existingUser, err = s.repo.GetByPhone(ctx, dto.Phone)
	if err == nil && existingUser != nil {
		return 0, errors.New("пользователь с таким телефоном уже существует")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(dto.Password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("ошибка при хешировании пароля", zap.Error(err))
		return 0, errors.New("ошибка при создании пользователя")
	}

	dto.Password = string(hashedPassword)

	id, err := s.repo.Create(ctx, dto)
	if err != nil {
		s.logger.Error("ошибка создания пользователя", zap.Error(err))
		return 0, errors.New("ошибка при создании пользователя")
	}

	return id, nil
}

func (s *UserServiceImpl) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("ошибка получения пользователя по ID", zap.Int64("id", id), zap.Error(err))
		return nil, errors.New("пользователь не найден")
	}

	return user, nil
}

func (s *UserServiceImpl) Update(ctx context.Context, id int64, dto domain.UpdateUserDTO) error {
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("пользователь для обновления не найден", zap.Int64("id", id), zap.Error(err))
		return errors.New("пользователь не найден")
	}

	if dto.Email != nil {
		existingUser, err := s.repo.GetByEmail(ctx, *dto.Email)
		if err == nil && existingUser != nil && existingUser.ID != id {
			return errors.New("пользователь с таким email уже существует")
		}
	}

	if dto.Phone != nil {
		existingUser, err := s.repo.GetByPhone(ctx, *dto.Phone)
		if err == nil && existingUser != nil && existingUser.ID != id {
			return errors.New("пользователь с таким телефоном уже существует")
		}
	}

	err = s.repo.Update(ctx, id, dto)
	if err != nil {
		s.logger.Error("ошибка обновления пользователя", zap.Int64("id", id), zap.Error(err))
		return errors.New("ошибка при обновлении пользователя")
	}

	return nil
}

func (s *UserServiceImpl) UpdatePassword(ctx context.Context, id int64, dto domain.PasswordUpdateDTO) error {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("пользователь для обновления пароля не найден", zap.Int64("id", id), zap.Error(err))
		return errors.New("пользователь не найден")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(dto.OldPassword))
	if err != nil {
		return errors.New("неверный текущий пароль")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(dto.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("ошибка при хешировании нового пароля", zap.Error(err))
		return errors.New("ошибка при обновлении пароля")
	}

	err = s.repo.UpdatePassword(ctx, id, string(hashedPassword))
	if err != nil {
		s.logger.Error("ошибка обновления пароля", zap.Int64("id", id), zap.Error(err))
		return errors.New("ошибка при обновлении пароля")
	}

	return nil
}

func (s *UserServiceImpl) Delete(ctx context.Context, id int64) error {
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("пользователь для удаления не найден", zap.Int64("id", id), zap.Error(err))
		return errors.New("пользователь не найден")
	}

	err = s.repo.Delete(ctx, id)
	if err != nil {
		s.logger.Error("ошибка удаления пользователя", zap.Int64("id", id), zap.Error(err))
		return errors.New("ошибка при удалении пользователя")
	}

	return nil
}

func (s *UserServiceImpl) List(ctx context.Context, limit, offset int) ([]domain.User, error) {
	if limit <= 0 {
		limit = 20
	}

	if offset < 0 {
		offset = 0
	}

	users, err := s.repo.List(ctx, limit, offset)
	if err != nil {
		s.logger.Error("ошибка получения списка пользователей", zap.Error(err))
		return nil, fmt.Errorf("ошибка при получении списка пользователей: %w", err)
	}

	return users, nil
}
