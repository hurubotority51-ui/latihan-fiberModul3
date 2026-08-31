package main

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

func main() {
	app := fiber.New()

	app.Get("/api/v1/students", listStudents)
	app.Get("/api/v1/students/:id", getStudent)
	app.Post("/api/v1/students", createStudent)

	fmt.Println("Server is running at http://localhost:3000")

	err := app.Listen(":3000")
	if err != nil {
		fmt.Println(err)
	}
}