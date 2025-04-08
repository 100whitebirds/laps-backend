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

type SpecializationRepo struct {
	db *pgxpool.Pool
}

func NewSpecializationRepository(db *pgxpool.Pool) *SpecializationRepo {
	return &SpecializationRepo{
		db: db,
	}
}

func (r *SpecializationRepo) Create(ctx context.Context, dto domain.CreateSpecializationDTO) (int64, error) {
	query := `
		INSERT INTO specializations (name, description, type, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $5)
		RETURNING id
	`

	now := time.Now()
	var id int64
	err := r.db.QueryRow(ctx, query,
		dto.Name,
		dto.Description,
		dto.Type,
		dto.IsActive,
		now,
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("ошибка создания специализации: %w", err)
	}

	return id, nil
}

func (r *SpecializationRepo) GetByID(ctx context.Context, id int64) (*domain.Specialization, error) {
	query := `
		SELECT id, name, description, type, is_active, created_at, updated_at
		FROM specializations
		WHERE id = $1
	`

	var specialization domain.Specialization
	err := r.db.QueryRow(ctx, query, id).Scan(
		&specialization.ID,
		&specialization.Name,
		&specialization.Description,
		&specialization.Type,
		&specialization.IsActive,
		&specialization.CreatedAt,
		&specialization.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("специализация с id %d не найдена", id)
		}
		return nil, fmt.Errorf("ошибка получения специализации: %w", err)
	}

	return &specialization, nil
}

func (r *SpecializationRepo) Update(ctx context.Context, id int64, dto domain.UpdateSpecializationDTO) error {
	setValues := make([]string, 0)
	args := make([]interface{}, 0)
	argID := 1

	if dto.Name != nil {
		setValues = append(setValues, fmt.Sprintf("name = $%d", argID))
		args = append(args, *dto.Name)
		argID++
	}

	if dto.Description != nil {
		setValues = append(setValues, fmt.Sprintf("description = $%d", argID))
		args = append(args, *dto.Description)
		argID++
	}

	if dto.IsActive != nil {
		setValues = append(setValues, fmt.Sprintf("is_active = $%d", argID))
		args = append(args, *dto.IsActive)
		argID++
	}

	setValues = append(setValues, fmt.Sprintf("updated_at = $%d", argID))
	args = append(args, time.Now())
	argID++

	args = append(args, id)

	query := fmt.Sprintf(`
		UPDATE specializations
		SET %s
		WHERE id = $%d
	`, strings.Join(setValues, ", "), argID)

	_, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("ошибка обновления специализации: %w", err)
	}

	return nil
}

func (r *SpecializationRepo) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM specializations WHERE id = $1`

	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("ошибка удаления специализации: %w", err)
	}

	return nil
}

func (r *SpecializationRepo) List(ctx context.Context, filter domain.SpecializationFilter) ([]domain.Specialization, error) {
	conditions := make([]string, 0)
	args := make([]interface{}, 0)
	argID := 1

	if filter.Type != nil {
		conditions = append(conditions, fmt.Sprintf("type = $%d", argID))
		args = append(args, *filter.Type)
		argID++
	}

	if filter.IsActive != nil {
		conditions = append(conditions, fmt.Sprintf("is_active = $%d", argID))
		args = append(args, *filter.IsActive)
		argID++
	}

	if filter.SearchTerm != nil {
		conditions = append(conditions, fmt.Sprintf("(name ILIKE $%d OR description ILIKE $%d)", argID, argID))
		args = append(args, "%"+*filter.SearchTerm+"%")
		argID++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	limitOffset := fmt.Sprintf("LIMIT $%d OFFSET $%d", argID, argID+1)
	args = append(args, filter.Limit, filter.Offset)
	argID += 2

	query := fmt.Sprintf(`
		SELECT id, name, description, type, is_active, created_at, updated_at
		FROM specializations
		%s
		ORDER BY name ASC
		%s
	`, whereClause, limitOffset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения списка специализаций: %w", err)
	}
	defer rows.Close()

	specializations := make([]domain.Specialization, 0)
	for rows.Next() {
		var specialization domain.Specialization
		if err := rows.Scan(
			&specialization.ID,
			&specialization.Name,
			&specialization.Description,
			&specialization.Type,
			&specialization.IsActive,
			&specialization.CreatedAt,
			&specialization.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("ошибка сканирования строки специализации: %w", err)
		}
		specializations = append(specializations, specialization)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при итерации по строкам: %w", err)
	}

	return specializations, nil
}

func (r *SpecializationRepo) CountByFilter(ctx context.Context, filter domain.SpecializationFilter) (int, error) {
	conditions := make([]string, 0)
	args := make([]interface{}, 0)
	argID := 1

	if filter.Type != nil {
		conditions = append(conditions, fmt.Sprintf("type = $%d", argID))
		args = append(args, *filter.Type)
		argID++
	}

	if filter.IsActive != nil {
		conditions = append(conditions, fmt.Sprintf("is_active = $%d", argID))
		args = append(args, *filter.IsActive)
		argID++
	}

	if filter.SearchTerm != nil {
		conditions = append(conditions, fmt.Sprintf("(name ILIKE $%d OR description ILIKE $%d)", argID, argID))
		args = append(args, "%"+*filter.SearchTerm+"%")
		argID++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	query := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM specializations
		%s
	`, whereClause)

	var count int
	err := r.db.QueryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("ошибка подсчета специализаций: %w", err)
	}

	return count, nil
}
