package service

import (
	"context"

	"github.com/muazwzxv/kafka-consumer-worker/internal/dto/request"
	"github.com/muazwzxv/kafka-consumer-worker/internal/dto/response"
)

type UserService interface {
	CreateUser(ctx context.Context, req request.CreateUserRequest) (*response.UserResponse, error)
	FetchUser(ctx context.Context, uuid string) (*response.UserDetailResponse, error)
}
