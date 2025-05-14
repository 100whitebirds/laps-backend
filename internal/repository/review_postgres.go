package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"laps/internal/domain"
)

type ReviewRepo struct {
	db *pgxpool.Pool
}

func NewReviewRepository(db *pgxpool.Pool) *ReviewRepo {
	return &ReviewRepo{
		db: db,
	}
}

func (r *ReviewRepo) Create(ctx context.Context, clientID int64, review domain.CreateReviewDTO) (int64, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("ошибка начала транзакции: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `
		INSERT INTO reviews (client_id, specialist_id, appointment_id, rating, text, is_recommended, 
		                     service_rating, meeting_efficiency, professionalism, price_quality, 
		                     cleanliness, attentiveness, specialist_experience, grammar, 
		                     created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $15)
		RETURNING id
	`

	now := time.Now()
	var id int64

	err = tx.QueryRow(ctx, query,
		clientID,
		review.SpecialistID,
		review.AppointmentID,
		review.Rating,
		review.Text,
		review.IsRecommended,
		review.ServiceRating,
		review.MeetingEfficiency,
		review.Professionalism,
		review.PriceQuality,
		review.Cleanliness,
		review.Attentiveness,
		review.SpecialistExperience,
		review.Grammar,
		now,
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("ошибка создания отзыва: %w", err)
	}

	updateRatingQuery := `
		UPDATE specialists
		SET rating = (
			SELECT AVG(rating) FROM reviews WHERE specialist_id = $1
		),
		reviews_count = (
			SELECT COUNT(*) FROM reviews WHERE specialist_id = $1
		)
		WHERE id = $1
	`

	_, err = tx.Exec(ctx, updateRatingQuery, review.SpecialistID)
	if err != nil {
		return 0, fmt.Errorf("ошибка обновления рейтинга специалиста: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("ошибка при коммите транзакции: %w", err)
	}

	return id, nil
}

func (r *ReviewRepo) GetByID(ctx context.Context, id int64) (*domain.Review, error) {
	query := `
		SELECT r.id, r.client_id, r.specialist_id, r.appointment_id, r.rating, r.text, r.is_recommended,
		       r.service_rating, r.meeting_efficiency, r.professionalism, r.price_quality,
		       r.cleanliness, r.attentiveness, r.specialist_experience, r.grammar,
		       r.created_at, r.updated_at,
		       u.first_name, u.last_name
		FROM reviews r
		JOIN users u ON r.client_id = u.id
		WHERE r.id = $1
	`

	var review domain.Review
	var userName, userLastName string

	err := r.db.QueryRow(ctx, query, id).Scan(
		&review.ID,
		&review.ClientID,
		&review.SpecialistID,
		&review.AppointmentID,
		&review.Rating,
		&review.Text,
		&review.IsRecommended,
		&review.ServiceRating,
		&review.MeetingEfficiency,
		&review.Professionalism,
		&review.PriceQuality,
		&review.Cleanliness,
		&review.Attentiveness,
		&review.SpecialistExperience,
		&review.Grammar,
		&review.CreatedAt,
		&review.UpdatedAt,
		&userName,
		&userLastName,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("отзыв с id %d не найден", id)
		}
		return nil, fmt.Errorf("ошибка получения отзыва: %w", err)
	}

	return &review, nil
}

func (r *ReviewRepo) Update(ctx context.Context, id int64, dto domain.UpdateReviewDTO) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("ошибка начала транзакции: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `
		UPDATE reviews SET
	`

	args := make([]interface{}, 0)
	argCount := 1

	setStatements := []string{}

	if dto.Rating != nil {
		setStatements = append(setStatements, fmt.Sprintf("rating = $%d", argCount))
		args = append(args, *dto.Rating)
		argCount++
	}

	if dto.Text != nil {
		setStatements = append(setStatements, fmt.Sprintf("text = $%d", argCount))
		args = append(args, *dto.Text)
		argCount++
	}

	if len(setStatements) == 0 {
		return nil
	}

	setStatements = append(setStatements, fmt.Sprintf("updated_at = $%d", argCount))
	args = append(args, time.Now())
	argCount++

	query += strings.Join(setStatements, ", ")
	query += fmt.Sprintf(" WHERE id = $%d", argCount)
	args = append(args, id)

	_, err = tx.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("ошибка обновления отзыва: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("ошибка при коммите транзакции: %w", err)
	}

	return nil
}

func (r *ReviewRepo) Delete(ctx context.Context, id int64) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("ошибка начала транзакции: %w", err)
	}
	defer tx.Rollback(ctx)

	var specialistID int64
	getSpecialistQuery := `SELECT specialist_id FROM reviews WHERE id = $1`
	err = tx.QueryRow(ctx, getSpecialistQuery, id).Scan(&specialistID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("отзыв с id %d не найден", id)
		}
		return fmt.Errorf("ошибка получения ID специалиста: %w", err)
	}

	deleteQuery := `DELETE FROM reviews WHERE id = $1`
	_, err = tx.Exec(ctx, deleteQuery, id)
	if err != nil {
		return fmt.Errorf("ошибка удаления отзыва: %w", err)
	}

	updateRatingQuery := `
		UPDATE specialists
		SET rating = (
			SELECT COALESCE(AVG(rating), 0) FROM reviews WHERE specialist_id = $1
		),
		reviews_count = (
			SELECT COUNT(*) FROM reviews WHERE specialist_id = $1
		)
		WHERE id = $1
	`

	_, err = tx.Exec(ctx, updateRatingQuery, specialistID)
	if err != nil {
		return fmt.Errorf("ошибка обновления рейтинга специалиста: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("ошибка при коммите транзакции: %w", err)
	}

	return nil
}

func (r *ReviewRepo) GetBySpecialistID(ctx context.Context, specialistID int64, limit, offset int) ([]domain.Review, error) {
	query := `
		SELECT r.id, r.client_id, r.specialist_id, r.appointment_id, r.rating, r.text, r.is_recommended,
		       r.service_rating, r.meeting_efficiency, r.professionalism, r.price_quality,
		       r.cleanliness, r.attentiveness, r.specialist_experience, r.grammar,
		       r.created_at, r.updated_at,
		       u.first_name, u.last_name
		FROM reviews r
		JOIN users u ON r.client_id = u.id
		WHERE r.specialist_id = $1
		ORDER BY r.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, specialistID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения отзывов о специалисте: %w", err)
	}
	defer rows.Close()

	reviews := make([]domain.Review, 0)
	for rows.Next() {
		var review domain.Review
		var userName, userLastName string

		if err := rows.Scan(
			&review.ID,
			&review.ClientID,
			&review.SpecialistID,
			&review.AppointmentID,
			&review.Rating,
			&review.Text,
			&review.IsRecommended,
			&review.ServiceRating,
			&review.MeetingEfficiency,
			&review.Professionalism,
			&review.PriceQuality,
			&review.Cleanliness,
			&review.Attentiveness,
			&review.SpecialistExperience,
			&review.Grammar,
			&review.CreatedAt,
			&review.UpdatedAt,
			&userName,
			&userLastName,
		); err != nil {
			return nil, fmt.Errorf("ошибка сканирования строки отзыва: %w", err)
		}

		reviews = append(reviews, review)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при итерации по строкам: %w", err)
	}

	return reviews, nil
}

func (r *ReviewRepo) GetByUserID(ctx context.Context, userID int64, limit, offset int) ([]domain.Review, error) {
	query := `
		SELECT r.id, r.client_id, r.specialist_id, r.appointment_id, r.rating, r.text, r.is_recommended,
		       r.service_rating, r.meeting_efficiency, r.professionalism, r.price_quality,
		       r.cleanliness, r.attentiveness, r.specialist_experience, r.grammar,
		       r.created_at, r.updated_at,
		       u.first_name, u.last_name
		FROM reviews r
		JOIN users u ON r.client_id = u.id
		WHERE r.client_id = $1
		ORDER BY r.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения отзывов пользователя: %w", err)
	}
	defer rows.Close()

	reviews := make([]domain.Review, 0)
	for rows.Next() {
		var review domain.Review
		var userName, userLastName string

		if err := rows.Scan(
			&review.ID,
			&review.ClientID,
			&review.SpecialistID,
			&review.AppointmentID,
			&review.Rating,
			&review.Text,
			&review.IsRecommended,
			&review.ServiceRating,
			&review.MeetingEfficiency,
			&review.Professionalism,
			&review.PriceQuality,
			&review.Cleanliness,
			&review.Attentiveness,
			&review.SpecialistExperience,
			&review.Grammar,
			&review.CreatedAt,
			&review.UpdatedAt,
			&userName,
			&userLastName,
		); err != nil {
			return nil, fmt.Errorf("ошибка сканирования строки отзыва: %w", err)
		}

		reviews = append(reviews, review)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при итерации по строкам: %w", err)
	}

	return reviews, nil
}

func (r *ReviewRepo) CountBySpecialistID(ctx context.Context, specialistID int64) (int, error) {
	query := `SELECT COUNT(*) FROM reviews WHERE specialist_id = $1`

	var count int
	err := r.db.QueryRow(ctx, query, specialistID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("ошибка подсчета отзывов о специалисте: %w", err)
	}

	return count, nil
}

func (r *ReviewRepo) CountByFilter(ctx context.Context, filter domain.ReviewFilter) (int, error) {
	var conditions []string
	var args []interface{}
	argCount := 1

	if filter.SpecialistID != nil {
		conditions = append(conditions, fmt.Sprintf("specialist_id = $%d", argCount))
		args = append(args, *filter.SpecialistID)
		argCount++
	}

	if filter.ClientID != nil {
		conditions = append(conditions, fmt.Sprintf("client_id = $%d", argCount))
		args = append(args, *filter.ClientID)
		argCount++
	}

	if filter.MinRating != nil {
		conditions = append(conditions, fmt.Sprintf("rating >= $%d", argCount))
		args = append(args, *filter.MinRating)
		argCount++
	}

	if filter.MaxRating != nil {
		conditions = append(conditions, fmt.Sprintf("rating <= $%d", argCount))
		args = append(args, *filter.MaxRating)
		argCount++
	}

	query := "SELECT COUNT(*) FROM reviews"

	if len(conditions) > 0 {
		query += " WHERE " + stringJoin(conditions, " AND ")
	}

	var count int
	err := r.db.QueryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("ошибка получения количества отзывов: %w", err)
	}

	return count, nil
}

func stringJoin(elems []string, sep string) string {
	switch len(elems) {
	case 0:
		return ""
	case 1:
		return elems[0]
	}
	n := len(sep) * (len(elems) - 1)
	for i := 0; i < len(elems); i++ {
		n += len(elems[i])
	}

	var b strings.Builder
	b.Grow(n)
	b.WriteString(elems[0])
	for _, s := range elems[1:] {
		b.WriteString(sep)
		b.WriteString(s)
	}
	return b.String()
}

func (r *ReviewRepo) List(ctx context.Context, filter domain.ReviewFilter) ([]domain.Review, error) {
	var conditions []string
	var args []interface{}
	argCount := 1

	if filter.SpecialistID != nil {
		conditions = append(conditions, fmt.Sprintf("r.specialist_id = $%d", argCount))
		args = append(args, *filter.SpecialistID)
		argCount++
	}

	if filter.ClientID != nil {
		conditions = append(conditions, fmt.Sprintf("r.client_id = $%d", argCount))
		args = append(args, *filter.ClientID)
		argCount++
	}

	if filter.MinRating != nil {
		conditions = append(conditions, fmt.Sprintf("r.rating >= $%d", argCount))
		args = append(args, *filter.MinRating)
		argCount++
	}

	if filter.MaxRating != nil {
		conditions = append(conditions, fmt.Sprintf("r.rating <= $%d", argCount))
		args = append(args, *filter.MaxRating)
		argCount++
	}

	baseQuery := `
		SELECT r.id, r.client_id, r.specialist_id, r.appointment_id, r.rating, r.text, r.is_recommended,
		       r.service_rating, r.meeting_efficiency, r.professionalism, r.price_quality,
		       r.cleanliness, r.attentiveness, r.specialist_experience, r.grammar,
		       r.created_at, r.updated_at,
		       u.first_name, u.last_name
		FROM reviews r
		JOIN users u ON r.client_id = u.id
	`

	query := baseQuery
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY r.created_at DESC"
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argCount, argCount+1)
	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса: %w", err)
	}
	defer rows.Close()

	reviews := make([]domain.Review, 0)
	for rows.Next() {
		var review domain.Review
		var userName, userLastName string

		if err := rows.Scan(
			&review.ID,
			&review.ClientID,
			&review.SpecialistID,
			&review.AppointmentID,
			&review.Rating,
			&review.Text,
			&review.IsRecommended,
			&review.ServiceRating,
			&review.MeetingEfficiency,
			&review.Professionalism,
			&review.PriceQuality,
			&review.Cleanliness,
			&review.Attentiveness,
			&review.SpecialistExperience,
			&review.Grammar,
			&review.CreatedAt,
			&review.UpdatedAt,
			&userName,
			&userLastName,
		); err != nil {
			return nil, fmt.Errorf("ошибка сканирования строки отзыва: %w", err)
		}

		reviews = append(reviews, review)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при итерации по строкам: %w", err)
	}

	return reviews, nil
}

func (r *ReviewRepo) CreateReply(ctx context.Context, userID int64, reviewID int64, reply domain.CreateReplyDTO) (int64, error) {
	query := `
		INSERT INTO review_replies (review_id, user_id, text, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $4)
		RETURNING id
	`

	now := time.Now()
	var id int64
	err := r.db.QueryRow(ctx, query,
		reviewID,
		userID,
		reply.Text,
		now,
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("ошибка создания ответа на отзыв: %w", err)
	}

	updateReviewQuery := `
		UPDATE reviews
		SET reply_id = $1
		WHERE id = $2
	`

	_, err = r.db.Exec(ctx, updateReviewQuery, id, reviewID)
	if err != nil {
		return 0, fmt.Errorf("ошибка обновления отзыва с ID ответа: %w", err)
	}

	return id, nil
}

func (r *ReviewRepo) GetReplyByID(ctx context.Context, id int64) (*domain.Reply, error) {
	query := `
		SELECT id, review_id, user_id, text, created_at, updated_at
		FROM review_replies
		WHERE id = $1
	`

	var reply domain.Reply
	err := r.db.QueryRow(ctx, query, id).Scan(
		&reply.ID,
		&reply.ReviewID,
		&reply.UserID,
		&reply.Text,
		&reply.CreatedAt,
		&reply.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("ответ на отзыв с id %d не найден", id)
		}
		return nil, fmt.Errorf("ошибка получения ответа на отзыв: %w", err)
	}

	return &reply, nil
}

func (r *ReviewRepo) DeleteReply(ctx context.Context, id int64) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("ошибка начала транзакции: %w", err)
	}
	defer tx.Rollback(ctx)

	var reviewID int64
	getReviewIDQuery := `SELECT review_id FROM review_replies WHERE id = $1`
	err = tx.QueryRow(ctx, getReviewIDQuery, id).Scan(&reviewID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("ответ на отзыв с id %d не найден", id)
		}
		return fmt.Errorf("ошибка получения ID отзыва: %w", err)
	}

	updateReviewQuery := `
		UPDATE reviews
		SET reply_id = NULL
		WHERE id = $1
	`
	_, err = tx.Exec(ctx, updateReviewQuery, reviewID)
	if err != nil {
		return fmt.Errorf("ошибка обновления отзыва: %w", err)
	}

	deleteQuery := `DELETE FROM review_replies WHERE id = $1`
	_, err = tx.Exec(ctx, deleteQuery, id)
	if err != nil {
		return fmt.Errorf("ошибка удаления ответа на отзыв: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("ошибка при коммите транзакции: %w", err)
	}

	return nil
}

func (r *ReviewRepo) GetRepliesByReviewID(ctx context.Context, reviewID int64) ([]domain.Reply, error) {
	query := `
		SELECT id, review_id, user_id, text, created_at, updated_at
		FROM review_replies
		WHERE review_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.db.Query(ctx, query, reviewID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения ответов на отзыв: %w", err)
	}
	defer rows.Close()

	replies := make([]domain.Reply, 0)
	for rows.Next() {
		var reply domain.Reply
		if err := rows.Scan(
			&reply.ID,
			&reply.ReviewID,
			&reply.UserID,
			&reply.Text,
			&reply.CreatedAt,
			&reply.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("ошибка сканирования строки ответа: %w", err)
		}

		replies = append(replies, reply)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при итерации по строкам: %w", err)
	}

	return replies, nil
}
