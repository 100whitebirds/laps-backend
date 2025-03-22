package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"laps/internal/domain"
)

type UserRepo struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepo {
	return &UserRepo{
		db: db,
	}
}

func (r *UserRepo) Create(ctx context.Context, dto domain.CreateUserDTO) (int64, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("ошибка начала транзакции: %w", err)
	}
	defer tx.Rollback(ctx)

	var id int64
	query := `
		INSERT INTO users (first_name, last_name, middle_name, email, phone, password_hash, role, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $9)
		RETURNING id
	`

	now := time.Now()
	err = tx.QueryRow(
		ctx,
		query,
		dto.FirstName,
		dto.LastName,
		dto.MiddleName,
		dto.Email,
		dto.Phone,
		dto.Password,
		dto.Role,
		true,
		now,
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("ошибка создания пользователя: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("ошибка коммита транзакции: %w", err)
	}

	return id, nil
}

func (r *UserRepo) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	query := `
		SELECT id, first_name, last_name, middle_name, email, phone, password_hash, role, is_active, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	var user domain.User
	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.MiddleName,
		&user.Email,
		&user.Phone,
		&user.PasswordHash,
		&user.Role,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("пользователь с id %d не найден", id)
		}
		return nil, fmt.Errorf("ошибка получения пользователя: %w", err)
	}

	return &user, nil
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT id, first_name, last_name, middle_name, email, phone, password_hash, role, is_active, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	var user domain.User
	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.MiddleName,
		&user.Email,
		&user.Phone,
		&user.PasswordHash,
		&user.Role,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("пользователь с email %s не найден", email)
		}
		return nil, fmt.Errorf("ошибка получения пользователя: %w", err)
	}

	return &user, nil
}

func (r *UserRepo) GetByPhone(ctx context.Context, phone string) (*domain.User, error) {
	query := `
		SELECT id, first_name, last_name, middle_name, email, phone, password_hash, role, is_active, created_at, updated_at
		FROM users
		WHERE phone = $1
	`

	var user domain.User
	err := r.db.QueryRow(ctx, query, phone).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.MiddleName,
		&user.Email,
		&user.Phone,
		&user.PasswordHash,
		&user.Role,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("пользователь с телефоном %s не найден", phone)
		}
		return nil, fmt.Errorf("ошибка получения пользователя: %w", err)
	}

	return &user, nil
}

func (r *UserRepo) Update(ctx context.Context, id int64, dto domain.UpdateUserDTO) error {
	setValues := []string{}
	args := []interface{}{id}
	argId := 2

	if dto.FirstName != nil {
		setValues = append(setValues, fmt.Sprintf("first_name = $%d", argId))
		args = append(args, *dto.FirstName)
		argId++
	}

	if dto.LastName != nil {
		setValues = append(setValues, fmt.Sprintf("last_name = $%d", argId))
		args = append(args, *dto.LastName)
		argId++
	}

	if dto.MiddleName != nil {
		setValues = append(setValues, fmt.Sprintf("middle_name = $%d", argId))
		args = append(args, *dto.MiddleName)
		argId++
	}

	if dto.Email != nil {
		setValues = append(setValues, fmt.Sprintf("email = $%d", argId))
		args = append(args, *dto.Email)
		argId++
	}

	if dto.Phone != nil {
		setValues = append(setValues, fmt.Sprintf("phone = $%d", argId))
		args = append(args, *dto.Phone)
		argId++
	}

	if dto.IsActive != nil {
		setValues = append(setValues, fmt.Sprintf("is_active = $%d", argId))
		args = append(args, *dto.IsActive)
		argId++
	}

	setValues = append(setValues, fmt.Sprintf("updated_at = $%d", argId))
	args = append(args, time.Now())

	if len(setValues) <= 1 {
		return nil
	}

	setQuery := "UPDATE users SET " + joinWithComma(setValues) + " WHERE id = $1"

	_, err := r.db.Exec(ctx, setQuery, args...)
	if err != nil {
		return fmt.Errorf("ошибка обновления пользователя: %w", err)
	}

	return nil
}

func (r *UserRepo) UpdatePassword(ctx context.Context, id int64, passwordHash string) error {
	query := `
		UPDATE users
		SET password_hash = $1, updated_at = $2
		WHERE id = $3
	`

	_, err := r.db.Exec(ctx, query, passwordHash, time.Now(), id)
	if err != nil {
		return fmt.Errorf("ошибка обновления пароля: %w", err)
	}

	return nil
}

func (r *UserRepo) Delete(ctx context.Context, id int64) error {
	query := `
		UPDATE users
		SET is_active = false, updated_at = $1
		WHERE id = $2
	`

	_, err := r.db.Exec(ctx, query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("ошибка удаления пользователя: %w", err)
	}

	return nil
}

func (r *UserRepo) List(ctx context.Context, limit, offset int) ([]domain.User, error) {
	query := `
		SELECT id, first_name, last_name, middle_name, email, phone, password_hash, role, is_active, created_at, updated_at
		FROM users
		ORDER BY id
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса списка пользователей: %w", err)
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var user domain.User
		err := rows.Scan(
			&user.ID,
			&user.FirstName,
			&user.LastName,
			&user.MiddleName,
			&user.Email,
			&user.Phone,
			&user.PasswordHash,
			&user.Role,
			&user.IsActive,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("ошибка чтения данных пользователя: %w", err)
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка обработки результатов: %w", err)
	}

	return users, nil
}

func joinWithComma(values []string) string {
	var result string
	for i, value := range values {
		if i > 0 {
			result += ", "
		}
		result += value
	}
	return result
}
