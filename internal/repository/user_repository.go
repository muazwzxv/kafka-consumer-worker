package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/samber/do/v2"
	"github.com/muazwzxv/kafka-consumer-worker/internal/database"
	"github.com/muazwzxv/kafka-consumer-worker/internal/database/store"
	"github.com/muazwzxv/kafka-consumer-worker/internal/dto/request"
	"github.com/muazwzxv/kafka-consumer-worker/internal/entity"
)

type MySQLUserRepository struct {
	queries *store.Queries
	db      store.DBTX
}

func NewMySQLUserRepository(i do.Injector) (UserRepository, error) {
	queries := do.MustInvoke[*store.Queries](i)
	db := do.MustInvoke[*database.Database](i)

	return &MySQLUserRepository{
		queries: queries,
		db:      db.DB, // Extract *sqlx.DB from Database wrapper
	}, nil
}

func (r *MySQLUserRepository) Create(ctx context.Context, item *entity.User) error {
	result, err := r.queries.CreateUser(ctx, r.db, store.CreateUserParams{
		Name: item.Name,
		Description: sql.NullString{
			String: item.Description,
			Valid:  item.Description != "",
		},
		Status: string(item.Status),
	})
	if err != nil {
		return ErrDatabaseError
	}

	id, err := result.LastInsertId()
	if err != nil {
		return ErrDatabaseError
	}

	item.ID = id
	return nil
}

func (r *MySQLUserRepository) GetByID(ctx context.Context, id int64) (*entity.User, error) {
	row, err := r.queries.GetUserByID(ctx, r.db, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, ErrDatabaseError
	}

	return r.toEntity(row), nil
}

func (r *MySQLUserRepository) List(ctx context.Context, req request.ListUsersRequest) ([]entity.User, int64, error) {
	whereConditions := []string{"1=1"}
	args := make([]interface{}, 0)
	
	if req.Name != nil {
		whereConditions = append(whereConditions, "name LIKE ?")
		args = append(args, "%"+*req.Name+"%")
	}
	
	if req.Status != nil {
		whereConditions = append(whereConditions, "status = ?")
		args = append(args, *req.Status)
	}
	
	if req.FromDate != nil {
		whereConditions = append(whereConditions, "created_at >= ?")
		args = append(args, *req.FromDate)
	}
	
	if req.ToDate != nil {
		whereConditions = append(whereConditions, "created_at <= ?")
		args = append(args, *req.ToDate)
	}
	
	whereClause := strings.Join(whereConditions, " AND ")
	
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM users WHERE %s", whereClause)
	var totalCount int64
	
	row := r.db.QueryRowContext(ctx, countQuery, args...)
	if err := row.Scan(&totalCount); err != nil {
		return nil, 0, ErrDatabaseError
	}
	
	sortBy := req.SortBy
	if sortBy == "" {
		sortBy = "created_at"
	}
	sortOrder := req.SortOrder
	if sortOrder == "" {
		sortOrder = "desc"
	}
	orderBy := fmt.Sprintf("ORDER BY %s %s", sortBy, strings.ToUpper(sortOrder))
	
	queryArgs := append(args, req.GetLimit(), req.GetOffset())
	query := fmt.Sprintf(`
		SELECT id, name, description, status, created_at, updated_at
		FROM users
		WHERE %s
		%s
		LIMIT ? OFFSET ?
	`, whereClause, orderBy)
	
	rows, err := r.db.QueryContext(ctx, query, queryArgs...)
	if err != nil {
		return nil, 0, ErrDatabaseError
	}
	defer rows.Close()
	
	var items []entity.User
	for rows.Next() {
		var item entity.User
		var description sql.NullString
		var createdAt, updatedAt sql.NullTime
		
		err := rows.Scan(
			&item.ID,
			&item.Name,
			&description,
			&item.Status,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			return nil, 0, ErrDatabaseError
		}
		
		if description.Valid {
			item.Description = description.String
		}
		if createdAt.Valid {
			item.CreatedAt = createdAt.Time
		}
		if updatedAt.Valid {
			item.UpdatedAt = updatedAt.Time
		}
		
		items = append(items, item)
	}
	
	if err := rows.Err(); err != nil {
		return nil, 0, ErrDatabaseError
	}
	
	return items, totalCount, nil
}

func (r *MySQLUserRepository) Count(ctx context.Context) (int64, error) {
	count, err := r.queries.CountUsers(ctx, r.db)
	if err != nil {
		return 0, ErrDatabaseError
	}
	return count, nil
}

func (r *MySQLUserRepository) CountByStatus(ctx context.Context, status entity.UserStatus) (int64, error) {
	rows, err := r.queries.GetUsersByStatus(ctx, r.db, string(status))
	if err != nil {
		return 0, ErrDatabaseError
	}
	return int64(len(rows)), nil
}

func (r *MySQLUserRepository) Update(ctx context.Context, item *entity.User) error {
	err := r.queries.UpdateUser(ctx, r.db, store.UpdateUserParams{
		Name: item.Name,
		Description: sql.NullString{
			String: item.Description,
			Valid:  item.Description != "",
		},
		Status: string(item.Status),
		ID:     item.ID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return ErrDatabaseError
	}
	return nil
}

func (r *MySQLUserRepository) UpdateStatus(ctx context.Context, id int64, status entity.UserStatus) error {
	query := "UPDATE users SET status = ?, updated_at = NOW() WHERE id = ?"
	result, err := r.db.ExecContext(ctx, query, string(status), id)
	if err != nil {
		return ErrDatabaseError
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return ErrDatabaseError
	}
	
	if rowsAffected == 0 {
		return ErrNotFound
	}
	
	return nil
}

func (r *MySQLUserRepository) BulkUpdateStatus(ctx context.Context, ids []int64, status entity.UserStatus) (int, []int64, error) {
	if len(ids) == 0 {
		return 0, []int64{}, nil
	}
	
	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids)+1)
	args[0] = string(status)
	
	for i, id := range ids {
		placeholders[i] = "?"
		args[i+1] = id
	}
	
	query := fmt.Sprintf(`
		UPDATE users
		SET status = ?, updated_at = NOW()
		WHERE id IN (%s)
	`, strings.Join(placeholders, ","))
	
	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, ids, ErrDatabaseError
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, ids, ErrDatabaseError
	}
	
	updatedCount := int(rowsAffected)
	failedIDs := []int64{}
	
	if updatedCount < len(ids) {
		for _, id := range ids {
			exists, err := r.ExistsByID(ctx, id)
			if err != nil || !exists {
				failedIDs = append(failedIDs, id)
			}
		}
	}
	
	return updatedCount, failedIDs, nil
}

func (r *MySQLUserRepository) Delete(ctx context.Context, id int64) error {
	err := r.queries.DeleteUser(ctx, r.db, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return ErrDatabaseError
	}
	return nil
}

func (r *MySQLUserRepository) ExistsByID(ctx context.Context, id int64) (bool, error) {
	_, err := r.queries.GetUserByID(ctx, r.db, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, ErrDatabaseError
	}
	return true, nil
}

func (r *MySQLUserRepository) toEntity(row *store.User) *entity.User {
	result := &entity.User{
		ID:     row.ID,
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
