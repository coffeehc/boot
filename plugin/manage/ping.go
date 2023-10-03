package manage

import (
	"github.com/gofiber/fiber/v2"
)

func ping() fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.SendString("pong")
	}
}
