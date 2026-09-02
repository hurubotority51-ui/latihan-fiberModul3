package main

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber/v2"
)

func main() {
	var err error

	db, err = connectDB()
	if err != nil {
		panic(err)
	}
	defer db.Close(context.Background())

	app := fiber.New()

	app.Get("/api/v1/students", listStudents)
	app.Get("/api/v1/students/:id", getStudent)
	app.Post("/api/v1/students", createStudent)

	fmt.Println("Server is running at http://localhost:3000")

	if err := app.Listen(":3000"); err != nil {
		panic(err)
	}
}