package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"laps/internal/domain"
)

type AuthRepo struct {
	db *pgxpool.Pool
}

func NewAuthRepository(db *pgxpool.Pool) *AuthRepo {
	return &AuthRepo{
		db: db,
	}
}

func (r *AuthRepo) CreateSession(ctx context.Context, session domain.Session) error {
	query := `
		INSERT INTO sessions (id, user_id, refresh_token, user_agent, ip, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.db.Exec(ctx, query,
		session.ID,
		session.UserID,
		session.RefreshToken,
		session.UserAgent,
		session.IP,
		session.ExpiresAt,
		session.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("ошибка создания сессии: %w", err)
	}

	return nil
}

func (r *AuthRepo) GetSessionByRefreshToken(ctx context.Context, refreshToken string) (*domain.Session, error) {
	query := `
		SELECT id, user_id, refresh_token, user_agent, ip, expires_at, created_at
		FROM sessions
		WHERE refresh_token = $1
	`

	var session domain.Session
	err := r.db.QueryRow(ctx, query, refreshToken).Scan(
		&session.ID,
		&session.UserID,
		&session.RefreshToken,
		&session.UserAgent,
		&session.IP,
		&session.ExpiresAt,
		&session.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("сессия не найдена")
		}
		return nil, fmt.Errorf("ошибка получения сессии: %w", err)
	}

	return &session, nil
}

func (r *AuthRepo) DeleteSession(ctx context.Context, id string) error {
	query := `DELETE FROM sessions WHERE id = $1`

	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("ошибка удаления сессии: %w", err)
	}

	return nil
}

func (r *AuthRepo) DeleteSessionsByUserID(ctx context.Context, userID int64) error {
	query := `DELETE FROM sessions WHERE user_id = $1`

	_, err := r.db.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("ошибка удаления сессий пользователя: %w", err)
	}

	return nil
}
