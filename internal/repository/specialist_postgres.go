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

type SpecialistRepo struct {
	db *pgxpool.Pool
}

func NewSpecialistRepository(db *pgxpool.Pool) *SpecialistRepo {
	return &SpecialistRepo{
		db: db,
	}
}

func (r *SpecialistRepo) Create(ctx context.Context, userID int64, dto domain.CreateSpecialistDTO) (int64, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("ошибка начала транзакции: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `
		INSERT INTO specialists (
			user_id, 
			type, 
			specialization, 
			experience, 
			description, 
			experience_years, 
			association_member, 
			primary_consult_price, 
			secondary_consult_price,
			profile_photo_url, 
			created_at, 
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $11)
		RETURNING id
	`

	now := time.Now()
	var id int64
	err = tx.QueryRow(ctx, query,
		userID,
		dto.Type,
		dto.Specialization,
		dto.Experience,
		dto.Description,
		dto.ExperienceYears,
		dto.AssociationMember,
		dto.PrimaryConsultPrice,
		dto.SecondaryConsultPrice,
		"",
		now,
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("ошибка создания специалиста: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("ошибка при коммите транзакции: %w", err)
	}

	return id, nil
}

func (r *SpecialistRepo) GetByID(ctx context.Context, id int64) (*domain.Specialist, error) {
	query := `
		SELECT s.id, s.user_id, s.type, s.specialization, s.experience, s.description, 
		       s.experience_years, s.association_member, s.rating, s.reviews_count, 
		       s.recommendation_rate, s.primary_consult_price, s.secondary_consult_price, 
		       s.is_verified, s.profile_photo_url, s.created_at, s.updated_at,
			   u.id, u.email, u.phone, u.first_name, u.last_name, u.middle_name, u.role, u.created_at, u.updated_at
		FROM specialists s
		JOIN users u ON s.user_id = u.id
		WHERE s.id = $1
	`

	var specialist domain.Specialist
	var user domain.User

	err := r.db.QueryRow(ctx, query, id).Scan(
		&specialist.ID,
		&specialist.UserID,
		&specialist.Type,
		&specialist.Specialization,
		&specialist.Experience,
		&specialist.Description,
		&specialist.ExperienceYears,
		&specialist.AssociationMember,
		&specialist.Rating,
		&specialist.ReviewsCount,
		&specialist.RecommendationRate,
		&specialist.PrimaryConsultPrice,
		&specialist.SecondaryConsultPrice,
		&specialist.IsVerified,
		&specialist.ProfilePhotoURL,
		&specialist.CreatedAt,
		&specialist.UpdatedAt,
		&user.ID,
		&user.Email,
		&user.Phone,
		&user.FirstName,
		&user.LastName,
		&user.MiddleName,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("специалист с id %d не найден", id)
		}
		return nil, fmt.Errorf("ошибка получения специалиста: %w", err)
	}

	specialist.User = user

	specialist.Education, err = r.GetEducationBySpecialistID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения образования: %w", err)
	}

	specialist.WorkExperience, err = r.GetWorkExperienceBySpecialistID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения опыта работы: %w", err)
	}

	return &specialist, nil
}

func (r *SpecialistRepo) GetByUserID(ctx context.Context, userID int64) (*domain.Specialist, error) {
	query := `
		SELECT id FROM specialists WHERE user_id = $1
	`

	var specialistID int64
	err := r.db.QueryRow(ctx, query, userID).Scan(&specialistID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("специалист с user_id %d не найден", userID)
		}
		return nil, fmt.Errorf("ошибка получения ID специалиста: %w", err)
	}

	return r.GetByID(ctx, specialistID)
}

func (r *SpecialistRepo) Update(ctx context.Context, id int64, dto domain.UpdateSpecialistDTO) error {
	query := `
		UPDATE specialists
		SET type = COALESCE($1, type),
		    specialization = COALESCE($2, specialization),
		    experience = COALESCE($3, experience),
		    description = COALESCE($4, description),
		    experience_years = COALESCE($5, experience_years),
		    association_member = COALESCE($6, association_member),
		    primary_consult_price = COALESCE($7, primary_consult_price),
		    secondary_consult_price = COALESCE($8, secondary_consult_price),
		    updated_at = $9
		WHERE id = $10
	`

	_, err := r.db.Exec(ctx, query,
		dto.Type,
		dto.Specialization,
		dto.Experience,
		dto.Description,
		dto.ExperienceYears,
		dto.AssociationMember,
		dto.PrimaryConsultPrice,
		dto.SecondaryConsultPrice,
		time.Now(),
		id,
	)

	if err != nil {
		return fmt.Errorf("ошибка обновления специалиста: %w", err)
	}

	return nil
}

func (r *SpecialistRepo) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM specialists WHERE id = $1`

	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("ошибка удаления специалиста: %w", err)
	}

	return nil
}

func (r *SpecialistRepo) List(ctx context.Context, specialistType *domain.SpecialistType, limit, offset int) ([]domain.Specialist, error) {
	var whereClause string
	var args []interface{}
	var argPos int = 1

	if specialistType != nil {
		whereClause = fmt.Sprintf("WHERE s.type = $%d", argPos)
		args = append(args, *specialistType)
		argPos++
	}

	query := fmt.Sprintf(`
		SELECT s.id, s.user_id, s.type, s.specialization, s.experience, s.description, 
		       s.experience_years, s.association_member, s.rating, s.reviews_count, 
		       s.recommendation_rate, s.primary_consult_price, s.secondary_consult_price, 
		       s.is_verified, s.created_at, s.updated_at,
			   u.id, u.email, u.phone, u.first_name, u.last_name, u.middle_name, u.role, u.created_at, u.updated_at
		FROM specialists s
		JOIN users u ON s.user_id = u.id
		%s
		ORDER BY s.rating DESC, s.reviews_count DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argPos, argPos+1)

	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения списка специалистов: %w", err)
	}
	defer rows.Close()

	specialists := make([]domain.Specialist, 0)
	for rows.Next() {
		var specialist domain.Specialist
		var user domain.User

		if err := rows.Scan(
			&specialist.ID,
			&specialist.UserID,
			&specialist.Type,
			&specialist.Specialization,
			&specialist.Experience,
			&specialist.Description,
			&specialist.ExperienceYears,
			&specialist.AssociationMember,
			&specialist.Rating,
			&specialist.ReviewsCount,
			&specialist.RecommendationRate,
			&specialist.PrimaryConsultPrice,
			&specialist.SecondaryConsultPrice,
			&specialist.IsVerified,
			&specialist.CreatedAt,
			&specialist.UpdatedAt,
			&user.ID,
			&user.Email,
			&user.Phone,
			&user.FirstName,
			&user.LastName,
			&user.MiddleName,
			&user.Role,
			&user.CreatedAt,
			&user.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("ошибка сканирования строки специалиста: %w", err)
		}

		specialist.User = user
		specialists = append(specialists, specialist)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при итерации по строкам: %w", err)
	}

	for i, spec := range specialists {
		education, err := r.GetEducationBySpecialistID(ctx, spec.ID)
		if err != nil {
			return nil, fmt.Errorf("ошибка получения образования: %w", err)
		}
		specialists[i].Education = education

		workExperience, err := r.GetWorkExperienceBySpecialistID(ctx, spec.ID)
		if err != nil {
			return nil, fmt.Errorf("ошибка получения опыта работы: %w", err)
		}
		specialists[i].WorkExperience = workExperience
	}

	return specialists, nil
}

func (r *SpecialistRepo) AddEducation(ctx context.Context, specialistID int64, education domain.EducationDTO) (int64, error) {
	query := `
		INSERT INTO education (
			specialist_id, institution, specialization, degree, graduation_year, 
			created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $6)
		RETURNING id
	`

	now := time.Now()
	var id int64
	err := r.db.QueryRow(ctx, query,
		specialistID,
		education.Institution,
		education.Specialization,
		education.Degree,
		education.GraduationYear,
		now,
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("ошибка добавления образования: %w", err)
	}

	return id, nil
}

func (r *SpecialistRepo) UpdateEducation(ctx context.Context, id int64, education domain.EducationDTO) error {
	query := `
		UPDATE education
		SET institution = $1,
		    specialization = $2,
		    degree = $3,
		    graduation_year = $4,
		    updated_at = $5
		WHERE id = $6
	`

	_, err := r.db.Exec(ctx, query,
		education.Institution,
		education.Specialization,
		education.Degree,
		education.GraduationYear,
		time.Now(),
		id,
	)

	if err != nil {
		return fmt.Errorf("ошибка обновления образования: %w", err)
	}

	return nil
}

func (r *SpecialistRepo) DeleteEducation(ctx context.Context, id int64) error {
	query := `DELETE FROM education WHERE id = $1`

	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("ошибка удаления образования: %w", err)
	}

	return nil
}

func (r *SpecialistRepo) GetEducationBySpecialistID(ctx context.Context, specialistID int64) ([]domain.Education, error) {
	query := `
		SELECT id, specialist_id, institution, specialization, degree, graduation_year, 
		       created_at, updated_at
		FROM education
		WHERE specialist_id = $1
		ORDER BY graduation_year DESC
	`

	rows, err := r.db.Query(ctx, query, specialistID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения образования: %w", err)
	}
	defer rows.Close()

	education := make([]domain.Education, 0)
	for rows.Next() {
		var edu domain.Education
		if err := rows.Scan(
			&edu.ID,
			&edu.SpecialistID,
			&edu.Institution,
			&edu.Specialization,
			&edu.Degree,
			&edu.GraduationYear,
			&edu.CreatedAt,
			&edu.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("ошибка при сканировании строки образования: %w", err)
		}
		education = append(education, edu)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при итерации по строкам: %w", err)
	}

	return education, nil
}

func (r *SpecialistRepo) GetEducationByID(ctx context.Context, id int64) (*domain.Education, error) {
	query := `
		SELECT id, specialist_id, institution, specialization, degree, graduation_year, 
		       created_at, updated_at
		FROM education
		WHERE id = $1
		LIMIT 1
	`

	var edu domain.Education
	err := r.db.QueryRow(ctx, query, id).Scan(
		&edu.ID,
		&edu.SpecialistID,
		&edu.Institution,
		&edu.Specialization,
		&edu.Degree,
		&edu.GraduationYear,
		&edu.CreatedAt,
		&edu.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("образование с ID %d не найдено", id)
		}
		return nil, fmt.Errorf("ошибка получения образования: %w", err)
	}

	return &edu, nil
}

func (r *SpecialistRepo) AddWorkExperience(ctx context.Context, specialistID int64, workExperience domain.WorkExperienceDTO) (int64, error) {
	query := `
		INSERT INTO work_experience (specialist_id, company, position, start_year, end_year, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $7)
		RETURNING id
	`

	now := time.Now()
	var id int64
	err := r.db.QueryRow(ctx, query,
		specialistID,
		workExperience.Company,
		workExperience.Position,
		workExperience.StartYear,
		workExperience.EndYear,
		workExperience.Description,
		now,
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("ошибка добавления опыта работы: %w", err)
	}

	return id, nil
}

func (r *SpecialistRepo) UpdateWorkExperience(ctx context.Context, id int64, workExperience domain.WorkExperienceDTO) error {
	query := `
		UPDATE work_experience
		SET company = $1,
		    position = $2,
		    start_year = $3,
		    end_year = $4,
		    description = $5,
		    updated_at = $6
		WHERE id = $7
	`

	_, err := r.db.Exec(ctx, query,
		workExperience.Company,
		workExperience.Position,
		workExperience.StartYear,
		workExperience.EndYear,
		workExperience.Description,
		time.Now(),
		id,
	)

	if err != nil {
		return fmt.Errorf("ошибка обновления опыта работы: %w", err)
	}

	return nil
}

func (r *SpecialistRepo) DeleteWorkExperience(ctx context.Context, id int64) error {
	query := `DELETE FROM work_experience WHERE id = $1`

	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("ошибка удаления опыта работы: %w", err)
	}

	return nil
}

func (r *SpecialistRepo) GetWorkExperienceBySpecialistID(ctx context.Context, specialistID int64) ([]domain.WorkPlace, error) {
	query := `
		SELECT id, specialist_id, company, position, start_year, end_year, description, created_at, updated_at
		FROM work_experience
		WHERE specialist_id = $1
		ORDER BY end_year DESC NULLS FIRST, start_year DESC
	`

	rows, err := r.db.Query(ctx, query, specialistID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения опыта работы: %w", err)
	}
	defer rows.Close()

	workExperience := make([]domain.WorkPlace, 0)
	for rows.Next() {
		var work domain.WorkPlace
		if err := rows.Scan(
			&work.ID,
			&work.SpecialistID,
			&work.Company,
			&work.Position,
			&work.StartYear,
			&work.EndYear,
			&work.Description,
			&work.CreatedAt,
			&work.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("ошибка сканирования строки опыта работы: %w", err)
		}
		workExperience = append(workExperience, work)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при итерации по строкам: %w", err)
	}

	return workExperience, nil
}

func (r *SpecialistRepo) GetWorkExperienceByID(ctx context.Context, id int64) (*domain.WorkPlace, error) {
	query := `
		SELECT id, specialist_id, company, position, start_year, end_year, description, created_at, updated_at
		FROM work_experience
		WHERE id = $1
		LIMIT 1
	`

	var work domain.WorkPlace
	err := r.db.QueryRow(ctx, query, id).Scan(
		&work.ID,
		&work.SpecialistID,
		&work.Company,
		&work.Position,
		&work.StartYear,
		&work.EndYear,
		&work.Description,
		&work.CreatedAt,
		&work.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("опыт работы с ID %d не найден", id)
		}
		return nil, fmt.Errorf("ошибка получения опыта работы: %w", err)
	}

	return &work, nil
}

func (r *SpecialistRepo) AddSpecialization(ctx context.Context, specialistID, specializationID int64) error {
	query := `
		INSERT INTO specialist_specializations (specialist_id, specialization_id, created_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (specialist_id, specialization_id) DO NOTHING
	`

	_, err := r.db.Exec(ctx, query, specialistID, specializationID, time.Now())
	if err != nil {
		return fmt.Errorf("ошибка добавления специализации: %w", err)
	}

	return nil
}

func (r *SpecialistRepo) RemoveSpecialization(ctx context.Context, specialistID, specializationID int64) error {
	query := `
		DELETE FROM specialist_specializations
		WHERE specialist_id = $1 AND specialization_id = $2
	`

	_, err := r.db.Exec(ctx, query, specialistID, specializationID)
	if err != nil {
		return fmt.Errorf("ошибка удаления специализации: %w", err)
	}

	return nil
}

func (r *SpecialistRepo) GetSpecializationsBySpecialistID(ctx context.Context, specialistID int64) ([]domain.Specialization, error) {
	query := `
		SELECT s.id, s.name, s.description, s.type, s.is_active, s.created_at, s.updated_at
		FROM specializations s
		JOIN specialist_specializations ss ON s.id = ss.specialization_id
		WHERE ss.specialist_id = $1
		ORDER BY s.name
	`

	rows, err := r.db.Query(ctx, query, specialistID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения специализаций: %w", err)
	}
	defer rows.Close()

	specializations := make([]domain.Specialization, 0)
	for rows.Next() {
		var spec domain.Specialization
		if err := rows.Scan(
			&spec.ID,
			&spec.Name,
			&spec.Description,
			&spec.Type,
			&spec.IsActive,
			&spec.CreatedAt,
			&spec.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("ошибка сканирования строки специализации: %w", err)
		}
		specializations = append(specializations, spec)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при итерации по строкам: %w", err)
	}

	return specializations, nil
}

func (r *SpecialistRepo) UpdateProfilePhoto(ctx context.Context, id int64, photoURL string) error {
	query := `
		UPDATE specialists
		SET profile_photo_url = $1,
		    updated_at = $2
		WHERE id = $3
	`

	_, err := r.db.Exec(ctx, query, photoURL, time.Now(), id)
	if err != nil {
		return fmt.Errorf("ошибка обновления фотографии профиля: %w", err)
	}

	return nil
}
