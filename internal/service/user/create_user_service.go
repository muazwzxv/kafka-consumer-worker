package service

import (
	"context"
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/muazwzxv/kafka-consumer-worker/internal/dto/request"
	"github.com/muazwzxv/kafka-consumer-worker/internal/dto/response"
	"github.com/muazwzxv/kafka-consumer-worker/internal/entity"
	"github.com/muazwzxv/kafka-consumer-worker/internal/publisher"
	"github.com/muazwzxv/kafka-consumer-worker/internal/repository"
	"github.com/samber/do/v2"
)

type UserServiceImpl struct {
	repo                   repository.UserRepository
	userLifecyclePublisher *publisher.UserLifecyclePublisher
}

func NewUserService(i do.Injector) (UserService, error) {
	repo := do.MustInvoke[repository.UserRepository](i)
	userLifecyclePublisher := do.MustInvoke[*publisher.UserLifecyclePublisher](i)

	return &UserServiceImpl{
		repo:                   repo,
		userLifecyclePublisher: userLifecyclePublisher,
	}, nil
}

func (s *UserServiceImpl) CreateUser(ctx context.Context, req request.CreateUserRequest) (*response.UserResponse, error) {
	status := entity.UserStatusPending
	item := &entity.User{
		UUID:        uuid.New().String(),
		Name:        req.Name,
		Description: req.Description,
		Status:      status,
	}

	if err := s.repo.Create(ctx, item); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, response.BuildError(
				fiber.StatusNotFound,
				"NOT_FOUND",
			)
		}
		return nil, response.BuildError(
			fiber.StatusInternalServerError,
			"DB_ERROR",
		)
	}

	user, err := s.repo.GetByUUID(ctx, item.UUID)
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return nil, response.BuildError(
			fiber.StatusInternalServerError,
			"DB_ERROR",
		)
	}

	if publishErr := s.publish(ctx, user); publishErr != nil {
		return nil, response.BuildError(
			fiber.StatusInternalServerError,
			"INTERNAL",
		)
	}

	return s.entityToResponse(item), nil
}

func (s *UserServiceImpl) publish(ctx context.Context, user *entity.User) error {
	return s.userLifecyclePublisher.Publish(ctx, user)
}

func (s *UserServiceImpl) entityToResponse(item *entity.User) *response.UserResponse {
	return &response.UserResponse{
		UUID:        item.UUID,
		Name:        item.Name,
		Description: item.Description,
		Status:      item.Status,
		CreatedAt:   item.CreatedAt,
		UpdatedAt:   item.UpdatedAt,
		IsActive:    item.IsActive(),
	}
}
