package service

import (
	"context"
	"database/sql"
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/muazwzxv/kafka-consumer-worker/internal/dto/response"
)

func (s *UserServiceImpl) FetchUser(ctx context.Context, uuid string) (*response.UserDetailResponse, error) {
	user, err := s.repo.GetByUUID(ctx, uuid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Warnw("user not found",
				"uuid", uuid)
			return nil, response.BuildError(fiber.StatusNotFound, response.NotFound)
		}
		log.Errorw("error querying db",
			"uuid", uuid,
			"err", err.Error())
		return nil, response.BuildError(fiber.StatusInternalServerError, response.Internal)
	}

	return &response.UserDetailResponse{
		UserResponse: s.entityToResponse(user),
	}, nil
}
