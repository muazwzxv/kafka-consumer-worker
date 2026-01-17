package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/muazwzxv/kafka-consumer-worker/internal/dto/response"
)

func (h *UserHandler) GetUser(c *fiber.Ctx) error {
	logger := log.WithContext(c.UserContext())

	userUuid := c.Params("uuid")
	if userUuid == "" {
		logger.Warnw("uuid provided is empty",
			"path", c.Path(),
			"ip", c.IP())
		return response.HandleError(c, response.BuildError(
			fiber.StatusBadRequest,
			response.BadRequest,
		))
	}

	logger.Infow("fetching user",
		"uuid", userUuid)

	userResp, err := h.service.FetchUser(c.UserContext(), userUuid)
	if err != nil {
		return response.HandleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(userResp)
}
