package service

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"

	"laps/internal/domain"
	"laps/internal/repository"
)

type ReviewServiceImpl struct {
	repo            repository.ReviewRepository
	specialistRepo  repository.SpecialistRepository
	userRepo        repository.UserRepository
	appointmentRepo repository.AppointmentRepository
	logger          *zap.Logger
}

func NewReviewService(
	repo repository.ReviewRepository,
	specialistRepo repository.SpecialistRepository,
	userRepo repository.UserRepository,
	appointmentRepo repository.AppointmentRepository,
	logger *zap.Logger,
) *ReviewServiceImpl {
	return &ReviewServiceImpl{
		repo:            repo,
		specialistRepo:  specialistRepo,
		userRepo:        userRepo,
		appointmentRepo: appointmentRepo,
		logger:          logger,
	}
}

func (s *ReviewServiceImpl) Create(ctx context.Context, clientID int64, dto domain.CreateReviewDTO) (int64, error) {
	_, err := s.userRepo.GetByID(ctx, clientID)
	if err != nil {
		s.logger.Error("пользователь не найден при создании отзыва", zap.Int64("clientID", clientID), zap.Error(err))
		return 0, errors.New("пользователь не найден")
	}

	_, err = s.specialistRepo.GetByID(ctx, dto.SpecialistID)
	if err != nil {
		s.logger.Error("специалист не найден при создании отзыва", zap.Int64("specialistID", dto.SpecialistID), zap.Error(err))
		return 0, errors.New("специалист не найден")
	}

	// Проверяем существование приема
	appointment, err := s.appointmentRepo.GetByID(ctx, dto.AppointmentID)
	if err != nil {
		s.logger.Error("прием не найден при создании отзыва", zap.Int64("appointmentID", dto.AppointmentID), zap.Error(err))
		return 0, errors.New("прием не найден")
	}

	// Проверяем, что прием принадлежит данному клиенту и специалисту
	if appointment.ClientID != clientID || appointment.SpecialistID != dto.SpecialistID {
		s.logger.Error("попытка создать отзыв для чужого приема",
			zap.Int64("clientID", clientID),
			zap.Int64("appointmentClientID", appointment.ClientID),
			zap.Int64("specialistID", dto.SpecialistID),
			zap.Int64("appointmentSpecialistID", appointment.SpecialistID))
		return 0, errors.New("вы можете оставить отзыв только о специалисте, у которого были на приеме")
	}

	// Проверяем, что прием завершен
	if appointment.Status != domain.AppointmentStatusCompleted {
		s.logger.Error("попытка создать отзыв для незавершенного приема",
			zap.String("status", string(appointment.Status)),
			zap.Int64("appointmentID", appointment.ID))
		return 0, errors.New("вы можете оставить отзыв только после завершения приема")
	}

	// Проверяем, не оставлял ли уже пользователь отзыв для этого приема
	existingReviews, _, err := s.List(ctx, domain.ReviewFilter{
		ClientID: &clientID,
		Limit:    100,
		Offset:   0,
	})
	if err != nil {
		s.logger.Error("ошибка проверки существующих отзывов", zap.Error(err))
		return 0, errors.New("ошибка при проверке существующих отзывов")
	}

	for _, review := range existingReviews {
		if review.AppointmentID == dto.AppointmentID {
			s.logger.Error("попытка создать повторный отзыв", zap.Int64("appointmentID", dto.AppointmentID))
			return 0, errors.New("вы уже оставили отзыв для этого приема")
		}
	}

	if dto.Rating < 1 || dto.Rating > 5 {
		s.logger.Error("некорректный рейтинг", zap.Int("rating", dto.Rating))
		return 0, errors.New("рейтинг должен быть от 1 до 5")
	}

	id, err := s.repo.Create(ctx, clientID, dto)
	if err != nil {
		s.logger.Error("ошибка создания отзыва", zap.Error(err))
		return 0, errors.New("ошибка при создании отзыва")
	}

	return id, nil
}

func (s *ReviewServiceImpl) GetByID(ctx context.Context, id int64) (*domain.Review, error) {
	review, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("ошибка получения отзыва", zap.Int64("id", id), zap.Error(err))
		return nil, errors.New("отзыв не найден")
	}
	return review, nil
}

func (s *ReviewServiceImpl) Update(ctx context.Context, id int64, dto domain.UpdateReviewDTO) error {
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("отзыв для обновления не найден", zap.Int64("id", id), zap.Error(err))
		return errors.New("отзыв не найден")
	}

	if dto.Rating != nil && (*dto.Rating < 1 || *dto.Rating > 5) {
		s.logger.Error("некорректный рейтинг", zap.Int("rating", *dto.Rating))
		return errors.New("рейтинг должен быть от 1 до 5")
	}

	err = s.repo.Update(ctx, id, dto)
	if err != nil {
		s.logger.Error("ошибка обновления отзыва", zap.Int64("id", id), zap.Error(err))
		return errors.New("ошибка при обновлении отзыва")
	}

	return nil
}

func (s *ReviewServiceImpl) Delete(ctx context.Context, id int64) error {
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("отзыв не найден", zap.Int64("id", id), zap.Error(err))
		return errors.New("отзыв не найден")
	}

	err = s.repo.Delete(ctx, id)
	if err != nil {
		s.logger.Error("ошибка удаления отзыва", zap.Int64("id", id), zap.Error(err))
		return errors.New("ошибка при удалении отзыва")
	}

	return nil
}

func (s *ReviewServiceImpl) GetBySpecialistID(ctx context.Context, specialistID int64, limit, offset int) ([]domain.Review, int, error) {
	_, err := s.specialistRepo.GetByID(ctx, specialistID)
	if err != nil {
		s.logger.Error("специалист не найден при получении отзывов", zap.Int64("specialistID", specialistID), zap.Error(err))
		return nil, 0, errors.New("специалист не найден")
	}

	filter := domain.ReviewFilter{
		SpecialistID: &specialistID,
		Limit:        limit,
		Offset:       offset,
	}

	reviews, err := s.repo.List(ctx, filter)
	if err != nil {
		s.logger.Error("ошибка получения отзывов о специалисте", zap.Int64("specialistID", specialistID), zap.Error(err))
		return nil, 0, errors.New("ошибка при получении отзывов")
	}

	count, err := s.repo.CountByFilter(ctx, filter)
	if err != nil {
		s.logger.Error("ошибка получения количества отзывов", zap.Int64("specialistID", specialistID), zap.Error(err))
		return reviews, 0, nil
	}

	return reviews, count, nil
}

func (s *ReviewServiceImpl) GetByUserID(ctx context.Context, userID int64, limit, offset int) ([]domain.Review, error) {
	_, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		s.logger.Error("пользователь не найден при получении отзывов", zap.Int64("userID", userID), zap.Error(err))
		return nil, errors.New("пользователь не найден")
	}

	filter := domain.ReviewFilter{
		ClientID: &userID,
		Limit:    limit,
		Offset:   offset,
	}

	reviews, err := s.repo.List(ctx, filter)
	if err != nil {
		s.logger.Error("ошибка получения отзывов пользователя", zap.Int64("userID", userID), zap.Error(err))
		return nil, errors.New("ошибка при получении отзывов")
	}

	return reviews, nil
}

func (s *ReviewServiceImpl) List(ctx context.Context, filter domain.ReviewFilter) ([]domain.Review, int, error) {
	count, err := s.repo.CountByFilter(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("ошибка подсчета отзывов: %w", err)
	}

	reviews, err := s.repo.List(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("ошибка получения списка отзывов: %w", err)
	}

	return reviews, count, nil
}

func (s *ReviewServiceImpl) CreateReply(ctx context.Context, userID int64, reply domain.CreateReplyDTO) (int64, error) {
	_, err := s.repo.GetByID(ctx, reply.ReviewID)
	if err != nil {
		return 0, fmt.Errorf("ошибка получения отзыва: %w", err)
	}

	replyID, err := s.repo.CreateReply(ctx, userID, reply)
	if err != nil {
		return 0, fmt.Errorf("ошибка создания ответа на отзыв: %w", err)
	}

	return replyID, nil
}

func (s *ReviewServiceImpl) GetReplyByID(ctx context.Context, id int64) (*domain.Reply, error) {
	reply, err := s.repo.GetReplyByID(ctx, id)
	if err != nil {
		s.logger.Error("ошибка получения ответа на отзыв", zap.Int64("id", id), zap.Error(err))
		return nil, errors.New("ответ на отзыв не найден")
	}
	return reply, nil
}

func (s *ReviewServiceImpl) DeleteReply(ctx context.Context, replyID int64) error {
	_, err := s.repo.GetReplyByID(ctx, replyID)
	if err != nil {
		return fmt.Errorf("ошибка получения ответа: %w", err)
	}

	err = s.repo.DeleteReply(ctx, replyID)
	if err != nil {
		s.logger.Error("ошибка удаления ответа на отзыв", zap.Int64("replyID", replyID), zap.Error(err))
		return fmt.Errorf("ошибка удаления ответа на отзыв: %w", err)
	}

	return nil
}

func (s *ReviewServiceImpl) GetRepliesByReviewID(ctx context.Context, reviewID int64) ([]domain.Reply, error) {
	replies, err := s.repo.GetRepliesByReviewID(ctx, reviewID)
	if err != nil {
		s.logger.Error("ошибка получения списка ответов на отзыв", zap.Int64("reviewID", reviewID), zap.Error(err))
		return nil, errors.New("ошибка при получении списка ответов на отзыв")
	}
	return replies, nil
}
