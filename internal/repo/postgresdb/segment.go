package postgresdb

import (
	"context"
	"fmt"

	"github.com/realPointer/segments/pkg/postgres"
)

type SegmentRepo struct {
	*postgres.Postgres
}

func NewSegmentRepo(pg *postgres.Postgres) *SegmentRepo {
	return &SegmentRepo{pg}
}

func (r *SegmentRepo) CreateSegment(ctx context.Context, name string) error {
	sql, args, _ := r.Builder.
		Insert("segments").
		Columns("name").
		Values(name).
		ToSql()

	_, err := r.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("SegmentRepo.CreateSegment - r.Pool.Exec: %v", err)
	}

	return nil
}

func (r *SegmentRepo) CreateSegmentAuto(ctx context.Context, name string, percentage float64) error {
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("SegmentRepo.CreateSegmentAuto - r.Pool.Begin: %v", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	sql, args, _ := r.Builder.
		Insert("segments").
		Columns("name").
		Values(name).
		ToSql()

	_, err = tx.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("SegmentRepo.CreateSegmentAuto - r.Pool.Exec: %v", err)
	}

	sql, args, _ = r.Builder.
		Select("id").
		From("users").
		OrderBy("RANDOM()").
		Limit(uint64(float64(r.CountUsers(ctx)) * percentage / 100)).
		ToSql()

	rows, err := tx.Query(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("SegmentRepo.CreateSegmentAuto - tx.Query: %v", err)
	}

	var userIDs []int
	for rows.Next() {
		var userID int
		err := rows.Scan(&userID)
		if err != nil {
			return fmt.Errorf("SegmentRepo.CreateSegmentAuto - rows.Scan: %v", err)
		}

		userIDs = append(userIDs, userID)
	}

	for _, userID := range userIDs {
		sql, args, _ = r.Builder.
			Insert("user_segments").
			Columns("user_id", "segment_name").
			Values(userID, name).
			ToSql()

		_, err := tx.Exec(ctx, sql, args...)
		if err != nil {
			return fmt.Errorf("SegmentRepo.CreateSegmentAuto - tx.Exec1: %v", err)
		}

		sql, args, _ = r.Builder.
			Insert("user_segments_log").
			Columns("user_id", "segment_name", "operation").
			Values(userID, name, "add").
			ToSql()

		_, err = tx.Exec(ctx, sql, args...)
		if err != nil {
			return fmt.Errorf("SegmentRepo.CreateSegmentAuto - tx.Exec2: %v", err)
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("SegmentRepo.CreateSegmentAuto - tx.Commit: %v", err)
	}

	return nil
}

func (r *SegmentRepo) CountUsers(ctx context.Context) int {
	sql, args, _ := r.Builder.
		Select("count(*)").
		From("users").
		ToSql()

	var count int
	err := r.Pool.QueryRow(ctx, sql, args...).Scan(&count)
	if err != nil {
		return 0
	}

	return count
}

func (r *SegmentRepo) DeleteSegment(ctx context.Context, name string) error {
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("SegmentRepo.DeleteSegment - r.Pool.Begin: %v", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	sql, args, _ := r.Builder.
		Select("us.user_id").
		From("user_segments as us").
		Join("segments as s on us.segment_name = s.name").
		Where("s.name = $1", name).
		ToSql()

	rows, err := tx.Query(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("SegmentRepo.DeleteSegment - tx.Query: %v", err)
	}

	var userIDs []int
	for rows.Next() {
		var userID int
		err := rows.Scan(&userID)
		if err != nil {
			return fmt.Errorf("SegmentRepo.DeleteSegment - rows.Scan: %v", err)
		}

		userIDs = append(userIDs, userID)
	}

	for _, userID := range userIDs {
		sql, args, _ = r.Builder.
			Insert("user_segments_log").
			Columns("user_id", "segment_name", "operation").
			Values(userID, name, "delete").
			ToSql()
		_, err := tx.Exec(ctx, sql, args...)
		if err != nil {
			return fmt.Errorf("SegmentRepo.DeleteSegment - tx.Exec1: %v", err)
		}
	}

	sql, args, _ = r.Builder.
		Delete("segments").
		Where("name = $1", name).
		ToSql()

	_, err = tx.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("SegmentRepo.DeleteSegment - tx.Exec2: %v", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("SegmentRepo.DeleteSegment - tx.Commit: %v", err)
	}

	return nil
}

func (r *SegmentRepo) GetSegments(ctx context.Context) ([]string, error) {
	sql, args, _ := r.Builder.
		Select("name").
		From("segments").
		ToSql()

	rows, err := r.Pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("SegmentRepo.GetSegments - r.Pool.Query: %v", err)
	}
	defer rows.Close()

	var segments []string
	for rows.Next() {
		var segment string
		err := rows.Scan(&segment)
		if err != nil {
			return nil, fmt.Errorf("SegmentRepo.GetSegments - rows.Scan: %v", err)
		}

		segments = append(segments, segment)
	}

	return segments, nil
}
