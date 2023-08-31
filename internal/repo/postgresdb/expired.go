package postgresdb

import (
	"context"
	"fmt"

	"github.com/realPointer/segments/pkg/postgres"
)

type ExpiredRepo struct {
	*postgres.Postgres
}

func NewExpiredRepo(pg *postgres.Postgres) *ExpiredRepo {
	return &ExpiredRepo{pg}
}

func (r *ExpiredRepo) DeleteExpiredRows(ctx context.Context) (int, error) {
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return -1, fmt.Errorf("SegmentRepo.DeleteSegment - r.Pool.Begin: %v", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	sql, args, _ := r.Builder.
		Select("user_id, segment_name").
		From("user_segments").
		Where("expire IS NOT NULL").
		Where("expire < NOW()").
		ToSql()

	rows, err := r.Pool.Query(ctx, sql, args...)
	if err != nil {
		return -1, fmt.Errorf("SegmentRepo.DeleteSegment - tx.Query: %v", err)
	}

	type userSegment struct {
		userID      int
		segmentName string
	}

	userSegments := make([]userSegment, 0)

	for rows.Next() {
		var userSegment userSegment
		err := rows.Scan(&userSegment.userID, &userSegment.segmentName)
		if err != nil {
			return -1, fmt.Errorf("SegmentRepo.DeleteSegment - rows.Scan: %v", err)
		}

		userSegments = append(userSegments, userSegment)
	}

	sql, args, _ = r.Builder.
		Delete("user_segments").
		Where("expire IS NOT NULL").
		Where("expire < NOW()").
		ToSql()

	_, err = tx.Exec(ctx, sql, args...)
	if err != nil {
		return -1, fmt.Errorf("SegmentRepo.DeleteSegment - tx.Exec: %v", err)
	}

	for _, userSegment := range userSegments {
		sql, args, _ = r.Builder.
			Insert("user_segments_log").
			Columns("user_id", "segment_name", "operation").
			Values(userSegment.userID, userSegment.segmentName, "delete").
			ToSql()

		_, err = tx.Exec(ctx, sql, args...)
		if err != nil {
			return -1, fmt.Errorf("SegmentRepo.DeleteSegment - tx.Exec: %v", err)
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return -1, fmt.Errorf("SegmentRepo.DeleteSegment - tx.Commit: %v", err)
	}

	return len(userSegments), nil
}
