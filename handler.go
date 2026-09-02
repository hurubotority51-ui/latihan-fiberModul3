package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// Mencari ID dari parameter URL
func paramID(c *fiber.Ctx) (int, bool) {
	id, err := strconv.Atoi(c.Params("id"))

	if err != nil || id < 1 {
		return 0, false
	}

	return id, true
}

// GET /api/v1/students
func listStudents(c *fiber.Ctx) error {
	q := parseListQuery(c)

	offset := (q.Page - 1) * q.Limit

	query := `
		SELECT id, nim, name, grade, is_active, created_at
		FROM students
	`

	args := []any{}
	conditions := []string{}

	// Filter berdasarkan is_active
	if q.IsActive != nil {
		args = append(args, *q.IsActive)

		conditions = append(
			conditions,
			fmt.Sprintf("is_active = $%d", len(args)),
		)
	}

	// Search berdasarkan NIM atau nama
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

	// WHERE
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	// Sorting
	query += " ORDER BY " + q.Sort + " " + strings.ToUpper(q.Order)

	// Pagination
	query += fmt.Sprintf(
		" LIMIT $%d OFFSET $%d",
		len(args)+1,
		len(args)+2,
	)

	args = append(args, q.Limit, offset)

	// Menjalankan query
	rows, err := db.Query(
		c.Context(),
		query,
		args...,
	)

	if err != nil {
		return fail(
			c,
			fiber.StatusInternalServerError,
			"gagal mengambil data student",
			err.Error(),
		)
	}

	defer rows.Close()

	hasil := []Student{}

	// Membaca hasil query
	for rows.Next() {
		var s Student

		err := rows.Scan(
			&s.ID,
			&s.NIM,
			&s.Name,
			&s.Grade,
			&s.IsActive,
			&s.CreatedAt,
		)

		if err != nil {
			return fail(
				c,
				fiber.StatusInternalServerError,
				"gagal membaca data student",
				err.Error(),
			)
		}

		hasil = append(hasil, s)
	}

	if err := rows.Err(); err != nil {
		return fail(
			c,
			fiber.StatusInternalServerError,
			"terjadi kesalahan saat membaca data student",
			err.Error(),
		)
	}

	// Query untuk menghitung total data
	countQuery := "SELECT COUNT(*) FROM students"

	countArgs := args[:len(args)-2]

	if len(conditions) > 0 {
		countQuery += " WHERE " + strings.Join(conditions, " AND ")
	}

	var total int

	err = db.QueryRow(
		c.Context(),
		countQuery,
		countArgs...,
	).Scan(&total)

	if err != nil {
		return fail(
			c,
			fiber.StatusInternalServerError,
			"gagal menghitung total student",
			err.Error(),
		)
	}

	totalPages := 0

	if total > 0 {
		totalPages = (total + q.Limit - 1) / q.Limit
	}

	mulai := (q.Page - 1) * q.Limit

	if mulai > total {
		mulai = total
	}

	akhir := mulai + q.Limit

	if akhir > total {
		akhir = total
	}

	return okList(
		c,
		"daftar student berhasil diambil",
		hasil,
		Meta{
			Page:       q.Page,
			Limit:      q.Limit,
			Total:      total,
			TotalPages: totalPages,
		},
	)
}

// GET /api/v1/students/:id
func getStudent(c *fiber.Ctx) error {

	id, valid := paramID(c)

	if !valid {
		return fail(
			c,
			fiber.StatusBadRequest,
			"id harus berupa angka positif",
			nil,
		)
	}

	var student Student

	err := db.QueryRow(
		c.Context(),
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
		return fail(
			c,
			fiber.StatusNotFound,
			"student tidak ditemukan",
			nil,
		)
	}

	return ok(
		c,
		"student ditemukan",
		student,
	)
}

// POST /api/v1/students
func createStudent(c *fiber.Ctx) error {

	var req CreateStudentRequest

	// Membaca JSON request
	if err := c.BodyParser(&req); err != nil {
		return fail(
			c,
			fiber.StatusBadRequest,
			"body harus berupa JSON yang valid",
			nil,
		)
	}

	// Membersihkan spasi
	req.NIM = strings.TrimSpace(req.NIM)
	req.Name = strings.TrimSpace(req.Name)

	errs := map[string]string{}

	// Validasi NIM
	if req.NIM == "" {
		errs["nim"] = "wajib diisi"
	}

	// Validasi nama
	if req.Name == "" {
		errs["name"] = "wajib diisi"
	}

	// Validasi grade
	if req.Grade < 0 || req.Grade > 4 {
		errs["grade"] = "nilai harus berada antara 0 dan 4"
	}

	// Jika validasi gagal
	if len(errs) > 0 {
		return failValidation(c, errs)
	}

	// Mengecek NIM duplikat di PostgreSQL
	var exists bool

	err := db.QueryRow(
		c.Context(),
		`
		SELECT EXISTS(
			SELECT 1
			FROM students
			WHERE LOWER(nim) = LOWER($1)
		)
		`,
		req.NIM,
	).Scan(&exists)

	if err != nil {
		return fail(
			c,
			fiber.StatusInternalServerError,
			"gagal mengecek NIM",
			err.Error(),
		)
	}

	if exists {
		return failValidation(
			c,
			map[string]string{
				"nim": "sudah digunakan",
			},
		)
	}

	// Menyimpan student ke PostgreSQL
	var student Student

	err = db.QueryRow(
		c.Context(),
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
		return fail(
			c,
			fiber.StatusInternalServerError,
			"gagal membuat student",
			err.Error(),
		)
	}

	return created(
		c,
		"student berhasil dibuat",
		student,
	)
}
