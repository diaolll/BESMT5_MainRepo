package helper

import "github.com/gofiber/fiber/v2"

func RequestID(c *fiber.Ctx) string {
	if v, ok := c.Locals("requestid").(string); ok {
		return v
	}
	return c.Get(fiber.HeaderXRequestID)
}
