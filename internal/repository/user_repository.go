package repository

import (
	"context"
	"database/sql"

	"github.com/muazwzxv/kafka-consumer-worker/internal/database"
	"github.com/muazwzxv/kafka-consumer-worker/internal/database/store"
	"github.com/muazwzxv/kafka-consumer-worker/internal/entity"
	"github.com/samber/do/v2"
)

type UserRepositoryImpl struct {
	queries *store.Queries
	db      store.DBTX
}

func NewUserRepository(i do.Injector) (UserRepository, error) {
	queries := do.MustInvoke[*store.Queries](i)
	db := do.MustInvoke[*database.Database](i)

	return &UserRepositoryImpl{
		queries: queries,
		db:      db.DB, // Extract *sqlx.DB from Database wrapper
	}, nil
}

func (r *UserRepositoryImpl) Create(ctx context.Context, item *entity.User) error {
	_, err := r.queries.CreateUser(ctx, r.db, store.CreateUserParams{
		Name: item.Name,
		Uuid: item.UUID,
		Description: sql.NullString{
			String: item.Description,
			Valid:  item.Description != "",
		},
		Status: string(item.Status),
	})
	if err != nil {
		return ErrDatabaseError
	}

	return nil
}

func (r *UserRepositoryImpl) GetByUUID(ctx context.Context, uuid string) (*entity.User, error) {
	row, err := r.queries.GetUserByUUID(ctx, r.db, uuid)
	if err != nil {
		return nil, err
	}

	return r.toEntity(row), nil
}

func (r *UserRepositoryImpl) UpdateUser(ctx context.Context, user *entity.User) error {
	if err := r.queries.UpdateUser(ctx, r.db, store.UpdateUserParams{
		Name: user.Name,
		Description: sql.NullString{
			String: user.Description,
			Valid:  user.Description != "",
		},
		Status: entity.UserStatusActive.String(),
		Uuid:   user.UUID,
	}); err != nil {
		return err
	}

	return nil
}

func (r *UserRepositoryImpl) toEntity(row *store.User) *entity.User {
	result := &entity.User{
		UUID:   row.Uuid,
		Name:   row.Name,
		Status: entity.UserStatus(row.Status),
	}

	if row.Description.Valid {
		result.Description = row.Description.String
	}
	if row.CreatedAt.Valid {
		result.CreatedAt = row.CreatedAt.Time
	}
	if row.UpdatedAt.Valid {
		result.UpdatedAt = row.UpdatedAt.Time
	}

	return result
}
