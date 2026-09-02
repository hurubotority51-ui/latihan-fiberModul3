package main

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

func main() {
	var err error

	db, err = connectDB()
	if err != nil {
		panic(err)
	}
	defer db.Close()
	studentRepo = NewPostgresStudentRepository(db)

	app := fiber.New()

	app.Get("/api/v1/students", listStudents)
	app.Get("/api/v1/students/:id", getStudent)
	app.Post("/api/v1/students", createStudent)
	app.Put("/api/v1/students/:id", updateStudent)
	app.Delete("/api/v1/students/:id", deleteStudent)

	fmt.Println("Server is running at http://localhost:3000")

	if err := app.Listen(":3000"); err != nil {
		panic(err)
	}
}
