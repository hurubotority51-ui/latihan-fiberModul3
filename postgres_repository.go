package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStudentRepository struct {
	db *pgxpool.Pool
}

func NewPostgresStudentRepository(db *pgxpool.Pool) *PostgresStudentRepository {
	return &PostgresStudentRepository{
		db: db,
	}
}

func (r *PostgresStudentRepository) FindAll(q ListQuery) ([]Student, int, error) {
	offset := (q.Page - 1) * q.Limit

	query := `
		SELECT id, nim, name, grade, is_active, created_at
		FROM students
	`

	args := []any{}
	conditions := []string{}

	if q.IsActive != nil {
		args = append(args, *q.IsActive)

		conditions = append(
			conditions,
			fmt.Sprintf("is_active = $%d", len(args)),
		)
	}

	if q.Search != "" {
		args = append(args, "%"+q.Search+"%")

		conditions = append(
			conditions,
			fmt.Sprintf(
				"(nim ILIKE $%d OR name ILIKE $%d)",
				len(args),
				len(args),
			),
		)
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY " + q.Sort + " " + strings.ToUpper(q.Order)

	query += fmt.Sprintf(
		" LIMIT $%d OFFSET $%d",
		len(args)+1,
		len(args)+2,
	)

	args = append(args, q.Limit, offset)

	rows, err := r.db.Query(
		context.Background(),
		query,
		args...,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	students := []Student{}

	for rows.Next() {
		var student Student

		err := rows.Scan(
			&student.ID,
			&student.NIM,
			&student.Name,
			&student.Grade,
			&student.IsActive,
			&student.CreatedAt,
		)

		if err != nil {
			return nil, 0, err
		}

		students = append(students, student)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	countQuery := "SELECT COUNT(*) FROM students"

	if len(conditions) > 0 {
		countQuery += " WHERE " + strings.Join(conditions, " AND ")
	}

	countArgs := args[:len(args)-2]

	var total int

	err = r.db.QueryRow(
		context.Background(),
		countQuery,
		countArgs...,
	).Scan(&total)

	if err != nil {
		return nil, 0, err
	}

	return students, total, nil
}

func (r *PostgresStudentRepository) FindByID(id int) (Student, error) {
	var student Student

	err := r.db.QueryRow(
		context.Background(),
		`
		SELECT id, nim, name, grade, is_active, created_at
		FROM students
		WHERE id = $1
		`,
		id,
	).Scan(
		&student.ID,
		&student.NIM,
		&student.Name,
		&student.Grade,
		&student.IsActive,
		&student.CreatedAt,
	)

	if err != nil {
		return Student{}, err
	}

	return student, nil
}

func (r *PostgresStudentRepository) ExistsByNIM(nim string) (bool, error) {
	var exists bool

	err := r.db.QueryRow(
		context.Background(),
		`
		SELECT EXISTS(
			SELECT 1
			FROM students
			WHERE LOWER(nim) = LOWER($1)
		)
		`,
		nim,
	).Scan(&exists)

	if err != nil {
		return false, err
	}

	return exists, nil
}

func (r *PostgresStudentRepository) Create(
	req CreateStudentRequest,
) (Student, error) {
	var student Student

	err := r.db.QueryRow(
		context.Background(),
		`
		INSERT INTO students (nim, name, grade, is_active)
		VALUES ($1, $2, $3, $4)
		RETURNING id, nim, name, grade, is_active, created_at
		`,
		req.NIM,
		req.Name,
		req.Grade,
		req.IsActive,
	).Scan(
		&student.ID,
		&student.NIM,
		&student.Name,
		&student.Grade,
		&student.IsActive,
		&student.CreatedAt,
	)

	if err != nil {
		return Student{}, err
	}

	return student, nil
}

func (r *PostgresStudentRepository) Update(
	id int,
	req ReplaceStudentRequest,
) (Student, error) {
	var student Student

	err := r.db.QueryRow(
		context.Background(),
		`
		UPDATE students
		SET nim = $1,
			name = $2,
			grade = $3,
			is_active = $4
		WHERE id = $5
		RETURNING id, nim, name, grade, is_active, created_at
		`,
		req.NIM,
		req.Name,
		req.Grade,
		req.IsActive,
		id,
	).Scan(
		&student.ID,
		&student.NIM,
		&student.Name,
		&student.Grade,
		&student.IsActive,
		&student.CreatedAt,
	)

	if err != nil {
		return Student{}, err
	}

	return student, nil
}

func (r *PostgresStudentRepository) Delete(id int) error {
	tag, err := r.db.Exec(
		context.Background(),
		`
		DELETE FROM students
		WHERE id = $1
		`,
		id,
	)

	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}
