package main

import (
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

	// Mengambil data melalui repository
	hasil, total, err := studentRepo.FindAll(q)
	if err != nil {
		return fail(
			c,
			fiber.StatusInternalServerError,
			"gagal mengambil data student",
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

	student, err := studentRepo.FindByID(id)

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
	exists, err := studentRepo.ExistsByNIM(req.NIM)

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
	student, err := studentRepo.Create(req)

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

// PUT /api/v1/students/:id
func updateStudent(c *fiber.Ctx) error {
	id, valid := paramID(c)

	if !valid {
		return fail(
			c,
			fiber.StatusBadRequest,
			"id harus berupa angka positif",
			nil,
		)
	}

	var req ReplaceStudentRequest

	if err := c.BodyParser(&req); err != nil {
		return fail(
			c,
			fiber.StatusBadRequest,
			"format JSON tidak valid",
			nil,
		)
	}

	if strings.TrimSpace(req.NIM) == "" {
		return failValidation(
			c,
			map[string]string{
				"nim": "wajib diisi",
			},
		)
	}

	if strings.TrimSpace(req.Name) == "" {
		return failValidation(
			c,
			map[string]string{
				"name": "wajib diisi",
			},
		)
	}

	student, err := studentRepo.Update(id, req)

	if err != nil {
		return fail(
			c,
			fiber.StatusInternalServerError,
			"gagal memperbarui student",
			err.Error(),
		)
	}

	return ok(
		c,
		"student berhasil diperbarui",
		student,
	)
}

// DELETE /api/v1/students/:id
func deleteStudent(c *fiber.Ctx) error {
	id, valid := paramID(c)

	if !valid {
		return fail(
			c,
			fiber.StatusBadRequest,
			"id harus berupa angka positif",
			nil,
		)
	}

	err := studentRepo.Delete(id)

	if err != nil {
		return fail(
			c,
			fiber.StatusInternalServerError,
			"gagal menghapus student",
			err.Error(),
		)
	}

	return noContent(c)
}
