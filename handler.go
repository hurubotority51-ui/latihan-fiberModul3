package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

var students []Student
var nextID = 1

// Mencari index student berdasarkan ID
func findStudentIndex(id int) int {
	for i := range students {
		if students[i].ID == id {
			return i
		}
	}

	return -1
}

// Mencari student berdasarkan kata pencarian
func searchMatch(s Student, kata string) bool {
	kata = strings.ToLower(kata)

	return strings.Contains(
		strings.ToLower(s.Name),
		kata,
	)
}

// Mengambil ID dari parameter URL
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

	// Filter is_active
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

	// Menghitung total data
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

	index := findStudentIndex(id)

	if index == -1 {
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
		students[index],
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

	// Mengecek NIM duplikat
	for _, s := range students {

		if strings.EqualFold(s.NIM, req.NIM) {
			errs["nim"] = "sudah digunakan"
		}
	}

	// Jika ada error validasi
	if len(errs) > 0 {
		return failValidation(c, errs)
	}

	// Membuat data student baru
	baru := Student{
		ID:        nextID,
		NIM:       req.NIM,
		Name:      req.Name,
		Grade:     req.Grade,
		IsActive:  req.IsActive,
		CreatedAt: time.Now(),
	}

	// Menambahkan ke slice
	students = append(students, baru)

	// ID berikutnya
	nextID++

	return created(
		c,
		"student berhasil dibuat",
		baru,
	)
}
