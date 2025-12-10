// Package repository contains interfaces for data access operations.
package repository

import (
	"context"

	"github.com/muazwzxv/kafka-consumer-worker/internal/dto/request"
	"github.com/muazwzxv/kafka-consumer-worker/internal/entity"
)

type UserRepository interface {
	Create(ctx context.Context, item *entity.User) error

	GetByID(ctx context.Context, id int64) (*entity.User, error)
	List(ctx context.Context, req request.ListUsersRequest) ([]entity.User, int64, error)
	Count(ctx context.Context) (int64, error)
	CountByStatus(ctx context.Context, status entity.UserStatus) (int64, error)

	Update(ctx context.Context, item *entity.User) error
	UpdateStatus(ctx context.Context, id int64, status entity.UserStatus) error
	BulkUpdateStatus(ctx context.Context, ids []int64, status entity.UserStatus) (int, []int64, error)

	Delete(ctx context.Context, id int64) error

	ExistsByID(ctx context.Context, id int64) (bool, error)
}

type DatabaseRepository interface {
	Ping(ctx context.Context) error
	Close() error
}
