package repo

import (
	"context"

	"github.com/realPointer/segments/internal/entity"
	"github.com/realPointer/segments/internal/repo/postgresdb"
	"github.com/realPointer/segments/pkg/postgres"
)

type User interface {
	CreateUser(ctx context.Context, userId int) error
	GetUserSegments(ctx context.Context, userId int) ([]string, error)
	AddOrRemoveUserSegments(ctx context.Context, userId int, addSegments []entity.AddSegment, removeSegments []string) error
	GetUserOperations(ctx context.Context, userId int) ([]string, error)
	GetUserOperationsByMonth(ctx context.Context, userId int, yearMonth string) ([]string, error)
}

type Segment interface {
	CreateSegment(ctx context.Context, name string) error
	CreateSegmentAuto(ctx context.Context, name string, percentage float64) error
	DeleteSegment(ctx context.Context, name string) error
	GetSegments(ctx context.Context) ([]string, error)
}

type Expired interface {
	DeleteExpiredRows(ctx context.Context) (int, error)
}

type Repositories struct {
	User
	Segment
	Expired
}

func NewRepositories(pg *postgres.Postgres) *Repositories {
	_, err := pg.Pool.Exec(context.Background(), `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY NOT NULL
	);

	CREATE TABLE IF NOT EXISTS segments (
		name VARCHAR(255) PRIMARY KEY NOT NULL,
		amount FLOAT
	);

	CREATE TABLE IF NOT EXISTS user_segments (
		user_id INTEGER NOT NULL,
		segment_name VARCHAR(255) NOT NULL,
		expire TIMESTAMP DEFAULT NULL,
		CONSTRAINT user_segments_pkey PRIMARY KEY (user_id, segment_name),
		CONSTRAINT user_segments_user_id_fkey FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
		CONSTRAINT user_segments_segment_name_fkey FOREIGN KEY (segment_name) REFERENCES segments (name) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS user_segments_log (
		user_id INTEGER NOT NULL,
		segment_name VARCHAR(255) NOT NULL,
		operation VARCHAR(20) NOT NULL,
		operation_time TIMESTAMP DEFAULT NOW()
	);`)

	if err != nil {
		panic(err)
	}

	return &Repositories{
		User:    postgresdb.NewUserRepo(pg),
		Segment: postgresdb.NewSegmentRepo(pg),
		Expired: postgresdb.NewExpiredRepo(pg),
	}
}
