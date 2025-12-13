// Package repository contains interfaces for data access operations.
package repository

import (
	"context"

	"github.com/muazwzxv/kafka-consumer-worker/internal/entity"
)

type UserRepository interface {
	Create(ctx context.Context, item *entity.User) error
	GetByUUID(ctx context.Context, uuid string) (*entity.User, error)
	UpdateUser(ctx context.Context, user *entity.User) error
}

type DatabaseRepository interface {
	Ping(ctx context.Context) error
	Close() error
}
