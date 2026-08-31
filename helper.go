package main

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

func ok(c *fiber.Ctx, message string, data any) error {
	return c.Status(fiber.StatusOK).JSON(WebResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func okList(c *fiber.Ctx, message string, data any, meta Meta) error {
	return c.Status(fiber.StatusOK).JSON(WebResponse{
		Success: true,
		Message: message,
		Data:    data,
		Meta:    &meta,
	})
}

func created(c *fiber.Ctx, message string, data any) error {
	return c.Status(fiber.StatusCreated).JSON(WebResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func noContent(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}

func fail(c *fiber.Ctx, status int, message string, errors any) error {
	return c.Status(status).JSON(WebResponse{
		Success: false,
		Message: message,
		Errors:  errors,
	})
}

func failValidation(c *fiber.Ctx, errors any) error {
	return fail(
		c,
		fiber.StatusBadRequest,
		"Validation failed",
		errors,
	)
}

var allowedSort = map[string]bool{
	"id":         true,
	"nim":        true,
	"name":       true,
	"grade":      true,
	"created_at": true,
}

func parseListQuery(c *fiber.Ctx) ListQuery {
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)

	if page < 1 {
		page = 1
	}

	if limit < 1 {
		limit = 10
	}

	if limit > 100 {
		limit = 100
	}

	search := strings.TrimSpace(c.Query("search"))
	sort := strings.ToLower(c.Query("sort", "id"))
	order := strings.ToLower(c.Query("order", "asc"))

	var isActive *bool

	if raw := c.Context().QueryArgs().Peek("is_active"); len(raw) > 0 {
		value := strings.ToLower(string(raw))

		if value == "true" {
			v := true
			isActive = &v
		} else if value == "false" {
			v := false
			isActive = &v
		}
	}

	if !allowedSort[sort] {
		sort = "id"
	}

	if order != "asc" && order != "desc" {
		order = "asc"
	}

	return ListQuery{
		Page:     page,
		Limit:    limit,
		Search:   search,
		Sort:     sort,
		Order:    order,
		IsActive: isActive,
	}
}
