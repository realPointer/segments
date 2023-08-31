package postgresdb

import (
	"context"
	"fmt"
	"time"

	"github.com/realPointer/segments/internal/entity"
	"github.com/realPointer/segments/pkg/postgres"
)

type UserRepo struct {
	*postgres.Postgres
}

func NewUserRepo(pg *postgres.Postgres) *UserRepo {
	return &UserRepo{pg}
}

func (r *UserRepo) CreateUser(ctx context.Context, userId int) error {
	sql, args, _ := r.Builder.
		Insert("users").
		Columns("id").
		Values(userId).
		ToSql()

	_, err := r.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("UserRepo.CreateUser - r.Pool.Exec: %v", err)
	}

	return nil
}

func (r *UserRepo) GetUserSegments(ctx context.Context, userId int) ([]string, error) {
	sql, args, _ := r.Builder.
		Select("us.segment_name").
		From("user_segments as us").
		Where("us.user_id = $1", userId).
		ToSql()

	rows, err := r.Pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("UserRepo.GetUserSegments - r.Pool.Query: %v", err)
	}
	defer rows.Close()

	var segments []string
	for rows.Next() {
		var segment string
		err := rows.Scan(&segment)
		if err != nil {
			return nil, fmt.Errorf("UserRepo.GetUserSegments - rows.Scan: %v", err)
		}

		segments = append(segments, segment)
	}

	return segments, nil
}

func (r *UserRepo) AddOrRemoveUserSegments(ctx context.Context, userId int, addSegments []entity.AddSegment, removeSegments []string) error {
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("UserRepo.AddOrRemoveUserSegments - r.Pool.Begin: %v", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	sql, args, _ := r.Builder.
		Select("id").
		From("users").
		Where("id = $1", userId).
		ToSql()

	var userCheckID int
	err = tx.QueryRow(ctx, sql, args...).Scan(&userCheckID)
	if err != nil {
		return fmt.Errorf("UserRepo.AddOrRemoveUserSegments - tx.QueryRow: %v", err)
	}

	for _, segment := range addSegments {
		sql, args, _ = r.Builder.
			Select("name").
			From("segments").
			Where("name = $1", segment.Name).
			ToSql()

		var segmentCheckName string
		err = tx.QueryRow(ctx, sql, args...).Scan(&segmentCheckName)
		if err != nil {
			return fmt.Errorf("UserRepo.AddOrRemoveUserSegments - tx.QueryRow2: %v", err)
		}

		if segment.Expire == "" {
			sql, args, _ = r.Builder.
				Insert("user_segments").
				Columns("user_id", "segment_name").
				Values(userId, segment.Name).
				ToSql()

			_, err := tx.Exec(ctx, sql, args...)
			if err != nil {
				return fmt.Errorf("UserRepo.AddOrRemoveUserSegments - tx.Exec1: %v", err)
			}
		} else {

			expire, err := time.ParseDuration(segment.Expire)
			if err != nil {
				return fmt.Errorf("UserRepo.AddOrRemoveUserSegments - time.ParseDuration: %v", err)
			}

			sql, args, _ = r.Builder.
				Insert("user_segments").
				Columns("user_id", "segment_name", "expire").
				Values(userId, segment.Name, time.Now().Add(expire)).
				ToSql()

			_, err = tx.Exec(ctx, sql, args...)
			if err != nil {
				return fmt.Errorf("UserRepo.AddOrRemoveUserSegments - tx.Exec1: %v", err)
			}
		}

		sql, args, _ = r.Builder.
			Insert("user_segments_log").
			Columns("user_id", "segment_name", "operation").
			Values(userId, segment.Name, "add").
			ToSql()

		_, err = tx.Exec(ctx, sql, args...)
		if err != nil {
			return fmt.Errorf("UserRepo.AddOrRemoveUserSegments - tx.Exec2: %v", err)
		}
	}

	for _, segment := range removeSegments {
		sql, args, _ = r.Builder.
			Select("name").
			From("segments").
			Where("name = $1", segment).
			ToSql()

		var segmentCheckName string
		err = tx.QueryRow(ctx, sql, args...).Scan(&segmentCheckName)
		if err != nil {
			return fmt.Errorf("UserRepo.AddOrRemoveUserSegments - tx.QueryRow2: %v", err)
		}

		sql, args, _ = r.Builder.
			Delete("user_segments").
			Where("user_id = $1", userId).
			Where("segment_name = $2", segment).
			ToSql()

		_, err := tx.Exec(ctx, sql, args...)
		if err != nil {
			return fmt.Errorf("UserRepo.AddOrRemoveUserSegments - tx.Exec3: %v", err)
		}

		sql, args, _ = r.Builder.
			Insert("user_segments_log").
			Columns("user_id", "segment_name", "operation").
			Values(userId, segment, "delete").
			ToSql()

		_, err = tx.Exec(ctx, sql, args...)
		if err != nil {
			return fmt.Errorf("UserRepo.AddOrRemoveUserSegments - tx.Exec4: %v", err)
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("UserRepo.AddOrRemoveUserSegments - tx.Commit: %v", err)
	}

	return nil
}

func (r *UserRepo) GetUserOperations(ctx context.Context, userId int) ([]string, error) {
	sql, args, _ := r.Builder.
		Select("user_id", "segment_name", "operation", "operation_time").
		From("user_segments_log").
		Where("user_id = $1", userId).
		ToSql()

	rows, err := r.Pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("UserRepo.GetUserOperations - r.Pool.Query: %v", err)
	}
	defer rows.Close()

	var operations []string
	for rows.Next() {
		var userID int
		var segmentName, operation string
		var operationTime time.Time
		err := rows.Scan(&userID, &segmentName, &operation, &operationTime)
		if err != nil {
			return nil, fmt.Errorf("UserRepo.GetUserOperations - rows.Scan: %v", err)
		}

		operations = append(operations, fmt.Sprintf("(%d, %s, %s, %s)", userID, segmentName, operation, operationTime))
	}

	return operations, nil
}

func (r *UserRepo) GetUserOperationsByMonth(ctx context.Context, userId int, yearMonth string) ([]string, error) {
	startDate, err := time.Parse("2006-01", yearMonth)
	if err != nil {
		return nil, fmt.Errorf("UserRepo.GetUserOperationsByMonth - time.Parse: %v", err)
	}
	endDate := startDate.AddDate(0, 1, 0).Add(-time.Nanosecond)

	sql, args, _ := r.Builder.
		Select("user_id", "segment_name", "operation", "operation_time").
		From("user_segments_log").
		Where("user_id = $1", userId).
		Where("operation_time >= $2", startDate).
		Where("operation_time <= $3", endDate).
		ToSql()

	rows, err := r.Pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("UserRepo.GetUserOperationsByMonth - r.Pool.Query: %v", err)
	}
	defer rows.Close()

	var operations []string
	for rows.Next() {
		var userID int
		var segmentName, operation string
		var operationTime time.Time
		err := rows.Scan(&userID, &segmentName, &operation, &operationTime)
		if err != nil {
			return nil, fmt.Errorf("UserRepo.GetUserOperationsByMonth - rows.Scan: %v", err)
		}

		operations = append(operations, fmt.Sprintf("(%d, %s, %s, %s)", userID, segmentName, operation, operationTime))
	}

	return operations, nil
}
