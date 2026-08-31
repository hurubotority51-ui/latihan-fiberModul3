package main

import (
	"sort"
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

	hasil := []Student{}

	// Filter data
	for _, s := range students {

		if q.IsActive != nil && s.IsActive != *q.IsActive {
			continue
		}

		if q.Search != "" && !searchMatch(s, q.Search) {
			continue
		}

		hasil = append(hasil, s)
	}

	// Sorting
	sort.SliceStable(hasil, func(i, j int) bool {
		var smaller bool

		switch q.Sort {

		case "nim":
			smaller = hasil[i].NIM < hasil[j].NIM

		case "name":
			smaller = hasil[i].Name < hasil[j].Name

		case "grade":
			smaller = hasil[i].Grade < hasil[j].Grade

		case "created_at":
			smaller = hasil[i].CreatedAt.Before(
				hasil[j].CreatedAt,
			)

		default:
			smaller = hasil[i].ID < hasil[j].ID
		}

		if q.Order == "desc" {
			return !smaller
		}

		return smaller
	})

	// Pagination
	total := len(hasil)

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
		hasil[mulai:akhir],
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
