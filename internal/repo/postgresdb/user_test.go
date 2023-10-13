package postgresdb

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pashagolub/pgxmock/v3"
	"github.com/realPointer/segments/internal/entity"
	"github.com/realPointer/segments/pkg/postgres"
	"github.com/stretchr/testify/assert"
)

type MockTimeProvider struct{}

func (m MockTimeProvider) Now() time.Time {
	return time.Date(2023, time.January, 1, 15, 30, 12, 345, time.UTC)
}

func TestUserRepo_CreateUser(t *testing.T) {
	type args struct {
		ctx    context.Context
		userId int
	}

	type MockBehavior func(m pgxmock.PgxPoolIface, args args)

	testCases := []struct {
		name         string
		args         args
		mockBehavior MockBehavior
		wantErr      bool
	}{
		{
			name: "OK",
			args: args{
				ctx:    context.Background(),
				userId: 1,
			},
			mockBehavior: func(m pgxmock.PgxPoolIface, args args) {
				m.ExpectExec("INSERT INTO users").
					WithArgs(args.userId).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
			},
			wantErr: false,
		},
		{
			name: "user already exists",
			args: args{
				ctx:    context.Background(),
				userId: 1,
			},
			mockBehavior: func(m pgxmock.PgxPoolIface, args args) {
				m.ExpectExec("INSERT INTO users").
					WithArgs(args.userId).
					WillReturnError(&pgconn.PgError{
						Code: "23505",
					})
			},
			wantErr: true,
		},
		{
			name: "unexpected error",
			args: args{
				ctx:    context.Background(),
				userId: 1,
			},
			mockBehavior: func(m pgxmock.PgxPoolIface, args args) {
				m.ExpectExec("INSERT INTO users").
					WithArgs(args.userId).
					WillReturnError(errors.New("some error"))
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			poolMock, _ := pgxmock.NewPool()
			defer poolMock.Close()
			tc.mockBehavior(poolMock, tc.args)

			postgresMock := &postgres.Postgres{
				Builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
				Pool:    poolMock,
			}
			userRepoMock := NewUserRepo(postgresMock, MockTimeProvider{})

			err := userRepoMock.CreateUser(tc.args.ctx, tc.args.userId)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			err = poolMock.ExpectationsWereMet()
			assert.NoError(t, err)
		})
	}
}

func TestUserRepo_GetUserSegments(t *testing.T) {
	type args struct {
		ctx    context.Context
		userId int
	}

	type MockBehavior func(m pgxmock.PgxPoolIface, args args)

	testCases := []struct {
		name         string
		args         args
		mockBehavior MockBehavior
		want         []string
		wantErr      bool
	}{
		{
			name: "OK",
			args: args{
				ctx:    context.Background(),
				userId: 1,
			},
			mockBehavior: func(m pgxmock.PgxPoolIface, args args) {
				m.ExpectQuery("SELECT us.segment_name").
					WithArgs(args.userId).
					WillReturnRows(pgxmock.NewRows([]string{"segment_name"}).AddRow("segment1"))
			},
			want:    []string{"segment1"},
			wantErr: false,
		},
		{
			name: "no segments",
			args: args{
				ctx:    context.Background(),
				userId: 1,
			},
			mockBehavior: func(m pgxmock.PgxPoolIface, args args) {
				m.ExpectQuery("SELECT us.segment_name").
					WithArgs(args.userId).
					WillReturnRows(pgxmock.NewRows([]string{"segment_name"}))
			},
			want:    []string(nil),
			wantErr: false,
		},
		{
			name: "unexpected error",
			args: args{
				ctx:    context.Background(),
				userId: 1,
			},
			mockBehavior: func(m pgxmock.PgxPoolIface, args args) {
				m.ExpectQuery("SELECT us.segment_name").
					WithArgs(args.userId).
					WillReturnError(errors.New("some error"))
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "rows.Scan error",
			args: args{
				ctx:    context.Background(),
				userId: 1,
			},
			mockBehavior: func(m pgxmock.PgxPoolIface, args args) {
				m.ExpectQuery("SELECT us.segment_name").
					WithArgs(args.userId).
					WillReturnRows(pgxmock.NewRows([]string{"segment_name"}).AddRow("segment1").RowError(0, errors.New("rows.Scan error")))
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			poolMock, _ := pgxmock.NewPool()
			defer poolMock.Close()
			tc.mockBehavior(poolMock, tc.args)

			postgresMock := &postgres.Postgres{
				Builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
				Pool:    poolMock,
			}
			userRepoMock := NewUserRepo(postgresMock, MockTimeProvider{})

			got, err := userRepoMock.GetUserSegments(tc.args.ctx, tc.args.userId)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			err = poolMock.ExpectationsWereMet()
			assert.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestUserRepo_GetUserOperations(t *testing.T) {
	type args struct {
		ctx    context.Context
		userId int
	}

	type MockBehavior func(m pgxmock.PgxPoolIface, args args)

	testCases := []struct {
		name         string
		args         args
		mockBehavior MockBehavior
		want         []string
		wantErr      bool
	}{
		{
			name: "OK",
			args: args{
				ctx:    context.Background(),
				userId: 1,
			},
			mockBehavior: func(m pgxmock.PgxPoolIface, args args) {
				operationTime := time.Unix(1672531200, 0)
				m.ExpectQuery("SELECT user_id, segment_name, operation, operation_time").
					WithArgs(args.userId).
					WillReturnRows(pgxmock.NewRows([]string{"user_id", "segment_name", "operation", "operation_time"}).AddRow(1, "segment1", "add", operationTime))
			},
			want:    []string{fmt.Sprintf("(1, segment1, add, %s)", time.Unix(1672531200, 0))},
			wantErr: false,
		},
		{
			name: "no operations",
			args: args{
				ctx:    context.Background(),
				userId: 1,
			},
			mockBehavior: func(m pgxmock.PgxPoolIface, args args) {
				m.ExpectQuery("SELECT user_id, segment_name, operation, operation_time").
					WithArgs(args.userId).
					WillReturnRows(pgxmock.NewRows([]string{"user_id", "segment_name", "operation", "operation_time"}))
			},
			want:    []string(nil),
			wantErr: false,
		},
		{
			name: "unexpected error",
			args: args{
				ctx:    context.Background(),
				userId: 1,
			},
			mockBehavior: func(m pgxmock.PgxPoolIface, args args) {
				m.ExpectQuery("SELECT user_id, segment_name, operation, operation_time").
					WithArgs(args.userId).
					WillReturnError(errors.New("some error"))
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "rows.Scan error",
			args: args{
				ctx:    context.Background(),
				userId: 1,
			},
			mockBehavior: func(m pgxmock.PgxPoolIface, args args) {
				operationTime := time.Unix(1672531200, 0)
				m.ExpectQuery("SELECT user_id, segment_name, operation, operation_time").
					WithArgs(args.userId).
					WillReturnRows(pgxmock.NewRows([]string{"user_id", "segment_name", "operation", "operation_time"}).AddRow(1, "segment1", "add", operationTime).RowError(0, errors.New("rows.Scan error")))
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			poolMock, _ := pgxmock.NewPool()
			defer poolMock.Close()
			tc.mockBehavior(poolMock, tc.args)

			postgresMock := &postgres.Postgres{
				Builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
				Pool:    poolMock,
			}
			userRepoMock := NewUserRepo(postgresMock, MockTimeProvider{})

			got, err := userRepoMock.GetUserOperations(tc.args.ctx, tc.args.userId)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			err = poolMock.ExpectationsWereMet()
			assert.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestUserRepo_GetUserOperationsByMonth(t *testing.T) {
	type args struct {
		ctx       context.Context
		userId    int
		yearMonth string
	}

	type MockBehavior func(m pgxmock.PgxPoolIface, args args)

	testCases := []struct {
		name         string
		args         args
		mockBehavior MockBehavior
		want         []string
		wantErr      bool
	}{
		{
			name: "OK",
			args: args{
				ctx:       context.Background(),
				userId:    1,
				yearMonth: "2023-01",
			},
			mockBehavior: func(m pgxmock.PgxPoolIface, args args) {
				operationTime := time.Date(2023, 1, 1, 0, 15, 23, 0, time.UTC)
				m.ExpectQuery("SELECT user_id, segment_name, operation, operation_time").
					WithArgs(args.userId, time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2023, 2, 1, 0, 0, 0, 0, time.UTC).Add(-time.Nanosecond)).
					WillReturnRows(pgxmock.NewRows([]string{"user_id", "segment_name", "operation", "operation_time"}).AddRow(1, "segment1", "add", operationTime))
			},
			want:    []string{fmt.Sprintf("(1, segment1, add, %s)", time.Date(2023, 1, 1, 0, 15, 23, 0, time.UTC))},
			wantErr: false,
		},
		{
			name: "no operations",
			args: args{
				ctx:       context.Background(),
				userId:    1,
				yearMonth: "2023-01",
			},
			mockBehavior: func(m pgxmock.PgxPoolIface, args args) {
				m.ExpectQuery("SELECT user_id, segment_name, operation, operation_time").
					WithArgs(args.userId, time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2023, 2, 1, 0, 0, 0, 0, time.UTC).Add(-time.Nanosecond)).
					WillReturnRows(pgxmock.NewRows([]string{"user_id", "segment_name", "operation", "operation_time"}))
			},
			want:    []string(nil),
			wantErr: false,
		},
		{
			name: "time.Parse error",
			args: args{
				ctx:       context.Background(),
				userId:    1,
				yearMonth: "2023-13",
			},
			mockBehavior: func(m pgxmock.PgxPoolIface, args args) {
				operationTime := time.Date(2023, 1, 1, 0, 15, 23, 0, time.UTC)
				m.ExpectQuery("SELECT user_id, segment_name, operation, operation_time").
					WithArgs(args.userId, time.Date(2023, 13, 1, 0, 0, 0, 0, time.UTC), time.Date(2023, 14, 1, 0, 0, 0, 0, time.UTC).Add(-time.Nanosecond)).
					WillReturnRows(pgxmock.NewRows([]string{"user_id", "segment_name", "operation", "operation_time"}).AddRow(1, "segment1", "add", operationTime))
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "unexpected error",
			args: args{
				ctx:       context.Background(),
				userId:    1,
				yearMonth: "2023-01",
			},
			mockBehavior: func(m pgxmock.PgxPoolIface, args args) {
				m.ExpectQuery("SELECT user_id, segment_name, operation, operation_time").
					WithArgs(args.userId, time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2023, 2, 1, 0, 0, 0, 0, time.UTC).Add(-time.Nanosecond)).
					WillReturnError(errors.New("some error"))
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "rows.Scan error",
			args: args{
				ctx:       context.Background(),
				userId:    1,
				yearMonth: "2023-01",
			},
			mockBehavior: func(m pgxmock.PgxPoolIface, args args) {
				operationTime := time.Date(2023, 1, 1, 0, 15, 23, 0, time.UTC)
				m.ExpectQuery("SELECT user_id, segment_name, operation, operation_time").
					WithArgs(args.userId, time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2023, 2, 1, 0, 0, 0, 0, time.UTC).Add(-time.Nanosecond)).
					WillReturnRows(pgxmock.NewRows([]string{"user_id", "segment_name", "operation", "operation_time"}).AddRow(1, "segment1", "add", operationTime).RowError(0, errors.New("rows.Scan error")))
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			poolMock, _ := pgxmock.NewPool()
			defer poolMock.Close()
			tc.mockBehavior(poolMock, tc.args)

			postgresMock := &postgres.Postgres{
				Builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
				Pool:    poolMock,
			}
			userRepoMock := NewUserRepo(postgresMock, MockTimeProvider{})

			got, err := userRepoMock.GetUserOperationsByMonth(tc.args.ctx, tc.args.userId, tc.args.yearMonth)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			err = poolMock.ExpectationsWereMet()
			assert.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestUserRepo_AddOrRemoveUserSegments(t *testing.T) {
	type args struct {
		ctx            context.Context
		userId         int
		addSegments    []entity.AddSegment
		removeSegments []string
	}

	type MockBehavior func(m pgxmock.PgxPoolIface, args args)

	testCases := []struct {
		name         string
		args         args
		mockBehavior MockBehavior
		wantErr      bool
	}{
		{
			name: "add 1 segment without expire time",
			args: args{
				ctx:    context.Background(),
				userId: 1,
				addSegments: []entity.AddSegment{
					{
						Name: "segment1",
					},
				},
			},
			mockBehavior: func(m pgxmock.PgxPoolIface, args args) {
				m.ExpectBegin()
				m.ExpectQuery("SELECT id").
					WithArgs(args.userId).
					WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(1))
				m.ExpectQuery("SELECT name").
					WithArgs(args.addSegments[0].Name).
					WillReturnRows(pgxmock.NewRows([]string{"name"}).AddRow(args.addSegments[0].Name))
				m.ExpectExec("INSERT INTO user_segments").
					WithArgs(args.userId, args.addSegments[0].Name).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
				m.ExpectExec("INSERT INTO user_segments_log").
					WithArgs(args.userId, args.addSegments[0].Name, "add").
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
				m.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name: "add 1 segment with expire time",
			args: args{
				ctx:    context.Background(),
				userId: 1,
				addSegments: []entity.AddSegment{
					{
						Name:   "segment1",
						Expire: "1h",
					},
				},
			},
			mockBehavior: func(m pgxmock.PgxPoolIface, args args) {
				m.ExpectBegin()
				m.ExpectQuery("SELECT id").
					WithArgs(args.userId).
					WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(1))
				m.ExpectQuery("SELECT name").
					WithArgs(args.addSegments[0].Name).
					WillReturnRows(pgxmock.NewRows([]string{"name"}).AddRow(args.addSegments[0].Name))
				expireTime := time.Date(2023, time.January, 1, 15, 30, 12, 345, time.UTC).Add(time.Hour)
				m.ExpectExec("INSERT INTO user_segments").
					WithArgs(args.userId, args.addSegments[0].Name, expireTime).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
				m.ExpectExec("INSERT INTO user_segments_log").
					WithArgs(args.userId, args.addSegments[0].Name, "add").
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
				m.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name: "remove 1 segment",
			args: args{
				ctx:    context.Background(),
				userId: 1,
				removeSegments: []string{
					"segment1",
				},
			},
			mockBehavior: func(m pgxmock.PgxPoolIface, args args) {
				m.ExpectBegin()
				m.ExpectQuery("SELECT id").
					WithArgs(args.userId).
					WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(1))
				m.ExpectQuery("SELECT name").
					WithArgs(args.removeSegments[0]).
					WillReturnRows(pgxmock.NewRows([]string{"name"}).AddRow(args.removeSegments[0]))
				m.ExpectExec("DELETE FROM user_segments").
					WithArgs(args.userId, args.removeSegments[0]).
					WillReturnResult(pgxmock.NewResult("DELETE", 1))
				m.ExpectExec("INSERT INTO user_segments_log").
					WithArgs(args.userId, args.removeSegments[0], "delete").
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
				m.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name: "add 1 segment and remove 1 segment",
			args: args{
				ctx:    context.Background(),
				userId: 1,
				addSegments: []entity.AddSegment{
					{
						Name: "segment1",
					},
				},
				removeSegments: []string{
					"segment2",
				},
			},
			mockBehavior: func(m pgxmock.PgxPoolIface, args args) {
				m.ExpectBegin()
				m.ExpectQuery("SELECT id").
					WithArgs(args.userId).
					WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(1))
				m.ExpectQuery("SELECT name").
					WithArgs(args.addSegments[0].Name).
					WillReturnRows(pgxmock.NewRows([]string{"name"}).AddRow(args.addSegments[0].Name))
				m.ExpectExec("INSERT INTO user_segments").
					WithArgs(args.userId, args.addSegments[0].Name).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
				m.ExpectExec("INSERT INTO user_segments_log").
					WithArgs(args.userId, args.addSegments[0].Name, "add").
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
				m.ExpectQuery("SELECT name").
					WithArgs(args.removeSegments[0]).
					WillReturnRows(pgxmock.NewRows([]string{"name"}).AddRow(args.removeSegments[0]))
				m.ExpectExec("DELETE FROM user_segments").
					WithArgs(args.userId, args.removeSegments[0]).
					WillReturnResult(pgxmock.NewResult("DELETE", 1))
				m.ExpectExec("INSERT INTO user_segments_log").
					WithArgs(args.userId, args.removeSegments[0], "delete").
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
				m.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name: "add 2 segments, first without expire time and second with",
			args: args{
				ctx:    context.Background(),
				userId: 1,
				addSegments: []entity.AddSegment{
					{
						Name: "segment1",
					},
					{
						Name:   "segment2",
						Expire: "1h",
					},
				},
			},
			mockBehavior: func(m pgxmock.PgxPoolIface, args args) {
				m.ExpectBegin()
				m.ExpectQuery("SELECT id").
					WithArgs(args.userId).
					WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(1))
				m.ExpectQuery("SELECT name").
					WithArgs(args.addSegments[0].Name).
					WillReturnRows(pgxmock.NewRows([]string{"name"}).AddRow(args.addSegments[0].Name))
				m.ExpectExec("INSERT INTO user_segments").
					WithArgs(args.userId, args.addSegments[0].Name).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
				m.ExpectExec("INSERT INTO user_segments_log").
					WithArgs(args.userId, args.addSegments[0].Name, "add").
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
				expireTime := time.Date(2023, time.January, 1, 15, 30, 12, 345, time.UTC).Add(time.Hour)
				m.ExpectQuery("SELECT name").
					WithArgs(args.addSegments[1].Name).
					WillReturnRows(pgxmock.NewRows([]string{"name"}).AddRow(args.addSegments[1].Name))
				m.ExpectExec("INSERT INTO user_segments").
					WithArgs(args.userId, args.addSegments[1].Name, expireTime).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
				m.ExpectExec("INSERT INTO user_segments_log").
					WithArgs(args.userId, args.addSegments[1].Name, "add").
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
				m.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name: "remove 2 segments",
			args: args{
				ctx:    context.Background(),
				userId: 1,
				removeSegments: []string{
					"segment1",
					"segment2",
				},
			},
			mockBehavior: func(m pgxmock.PgxPoolIface, args args) {
				m.ExpectBegin()
				m.ExpectQuery("SELECT id").
					WithArgs(args.userId).
					WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(1))
				m.ExpectQuery("SELECT name").
					WithArgs(args.removeSegments[0]).
					WillReturnRows(pgxmock.NewRows([]string{"name"}).AddRow(args.removeSegments[0]))
				m.ExpectExec("DELETE FROM user_segments").
					WithArgs(args.userId, args.removeSegments[0]).
					WillReturnResult(pgxmock.NewResult("DELETE", 1))
				m.ExpectExec("INSERT INTO user_segments_log").
					WithArgs(args.userId, args.removeSegments[0], "delete").
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
				m.ExpectQuery("SELECT name").
					WithArgs(args.removeSegments[1]).
					WillReturnRows(pgxmock.NewRows([]string{"name"}).AddRow(args.removeSegments[1]))
				m.ExpectExec("DELETE FROM user_segments").
					WithArgs(args.userId, args.removeSegments[1]).
					WillReturnResult(pgxmock.NewResult("DELETE", 1))
				m.ExpectExec("INSERT INTO user_segments_log").
					WithArgs(args.userId, args.removeSegments[1], "delete").
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
				m.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name: "add 1 segment and remove 1 segment with same name",
			args: args{
				ctx:    context.Background(),
				userId: 1,
				addSegments: []entity.AddSegment{
					{
						Name: "segment1",
					},
				},
				removeSegments: []string{
					"segment1",
				},
			},
			mockBehavior: func(m pgxmock.PgxPoolIface, args args) {
				m.ExpectBegin()
				m.ExpectQuery("SELECT id").
					WithArgs(args.userId).
					WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(1))
				m.ExpectQuery("SELECT name").
					WithArgs(args.addSegments[0].Name).
					WillReturnRows(pgxmock.NewRows([]string{"name"}).AddRow(args.addSegments[0].Name))
				m.ExpectExec("INSERT INTO user_segments").
					WithArgs(args.userId, args.addSegments[0].Name).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
				m.ExpectExec("INSERT INTO user_segments_log").
					WithArgs(args.userId, args.addSegments[0].Name, "add").
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
				m.ExpectQuery("SELECT name").
					WithArgs(args.removeSegments[0]).
					WillReturnRows(pgxmock.NewRows([]string{"name"}).AddRow(args.removeSegments[0]))
				m.ExpectExec("DELETE FROM user_segments").
					WithArgs(args.userId, args.removeSegments[0]).
					WillReturnResult(pgxmock.NewResult("DELETE", 1))
				m.ExpectExec("INSERT INTO user_segments_log").
					WithArgs(args.userId, args.removeSegments[0], "delete").
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
				m.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name: "r.Pool.Begin error",
			args: args{
				ctx: context.Background(),
			},
			mockBehavior: func(m pgxmock.PgxPoolIface, args args) {
				m.ExpectBegin().WillReturnError(errors.New("some error"))
				m.ExpectRollback()
			},
			wantErr: true,
		},
		{
			name: "SELECT id tx.QueryRow error",
			args: args{
				ctx:    context.Background(),
				userId: 1,
			},
			mockBehavior: func(m pgxmock.PgxPoolIface, args args) {
				m.ExpectBegin()
				m.ExpectQuery("SELECT id").
					WithArgs(args.userId).
					WillReturnError(errors.New("some error"))
				m.ExpectRollback()
			},
			wantErr: true,
		},
		{
			name: "SELECT name in addSegments tx.QueryRow2 error",
			args: args{
				ctx:    context.Background(),
				userId: 1,
				addSegments: []entity.AddSegment{
					{
						Name: "segment1",
					},
				},
			},
			mockBehavior: func(m pgxmock.PgxPoolIface, args args) {
				m.ExpectBegin()
				m.ExpectQuery("SELECT id").
					WithArgs(args.userId).
					WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(1))
				m.ExpectQuery("SELECT name").
					WithArgs(args.addSegments[0].Name).
					WillReturnError(errors.New("some error"))
				m.ExpectRollback()
			},
			wantErr: true,
		},
		{
			name: "INSERT into user_segments in addSegments without expire tx.Exec1 error",
			args: args{
				ctx:    context.Background(),
				userId: 1,
				addSegments: []entity.AddSegment{
					{
						Name: "segment1",
					},
				},
			},
			mockBehavior: func(m pgxmock.PgxPoolIface, args args) {
				segmentName := args.addSegments[0].Name
				m.ExpectBegin()
				m.ExpectQuery("SELECT id").
					WithArgs(args.userId).
					WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(1))
				m.ExpectQuery("SELECT name").
					WithArgs(segmentName).
					WillReturnRows(pgxmock.NewRows([]string{"name"}).AddRow(segmentName))
				m.ExpectExec("INSERT INTO user_segments").
					WithArgs(args.userId, segmentName).
					WillReturnError(errors.New("some error"))
				m.ExpectRollback()
			},
			wantErr: true,
		},
		{
			name: "time.ParseDuration error in addSegments with expire",
			args: args{
				ctx:    context.Background(),
				userId: 1,
				addSegments: []entity.AddSegment{
					{
						Name:   "segment1",
						Expire: "1hsss",
					},
				},
			},
			mockBehavior: func(m pgxmock.PgxPoolIface, args args) {
				segmentName := args.addSegments[0].Name
				m.ExpectBegin()
				m.ExpectQuery("SELECT id").
					WithArgs(args.userId).
					WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(1))
				m.ExpectQuery("SELECT name").
					WithArgs(segmentName).
					WillReturnRows(pgxmock.NewRows([]string{"name"}).AddRow(segmentName))
				m.ExpectRollback()
			},
			wantErr: true,
		},
		{
			name: "INSERT into user_segments in addSegments with expire tx.Exec1 error",
			args: args{
				ctx:    context.Background(),
				userId: 1,
				addSegments: []entity.AddSegment{
					{
						Name:   "segment1",
						Expire: "1h",
					},
				},
			},
			mockBehavior: func(m pgxmock.PgxPoolIface, args args) {
				segmentName := args.addSegments[0].Name
				m.ExpectBegin()
				m.ExpectQuery("SELECT id").
					WithArgs(args.userId).
					WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(1))
				m.ExpectQuery("SELECT name").
					WithArgs(segmentName).
					WillReturnRows(pgxmock.NewRows([]string{"name"}).AddRow(segmentName))
				expireTime := time.Date(2023, time.January, 1, 15, 30, 12, 345, time.UTC).Add(time.Hour)
				m.ExpectExec("INSERT INTO user_segments").
					WithArgs(args.userId, args.addSegments[0].Name, expireTime).
					WillReturnError(errors.New("some error"))
				m.ExpectRollback()
			},
			wantErr: true,
		},
		{
			name: "INSERT into user_segments_log in addSegments tx.Exec2 error",
			args: args{
				ctx:    context.Background(),
				userId: 1,
				addSegments: []entity.AddSegment{
					{
						Name: "segment1",
					},
				},
			},
			mockBehavior: func(m pgxmock.PgxPoolIface, args args) {
				m.ExpectBegin()
				m.ExpectQuery("SELECT id").
					WithArgs(args.userId).
					WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(1))
				m.ExpectQuery("SELECT name").
					WithArgs(args.addSegments[0].Name).
					WillReturnRows(pgxmock.NewRows([]string{"name"}).AddRow(args.addSegments[0].Name))
				m.ExpectExec("INSERT INTO user_segments").
					WithArgs(args.userId, args.addSegments[0].Name).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
				m.ExpectExec("INSERT INTO user_segments_log").
					WithArgs(args.userId, args.addSegments[0].Name, "add").
					WillReturnError(errors.New("some error"))
				m.ExpectRollback()
			},
			wantErr: true,
		},
		{
			name: "SELECT name in removeSegments tx.QueryRow2 error",
			args: args{
				ctx:    context.Background(),
				userId: 1,
				removeSegments: []string{
					"segment1",
				},
			},
			mockBehavior: func(m pgxmock.PgxPoolIface, args args) {
				m.ExpectBegin()
				m.ExpectQuery("SELECT id").
					WithArgs(args.userId).
					WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(1))
				m.ExpectQuery("SELECT name").
					WithArgs(args.removeSegments[0]).
					WillReturnError(errors.New("some error"))
				m.ExpectRollback()
			},
			wantErr: true,
		},
		{
			name: "DELETE user_segments in removeSegments tx.Exec3 error",
			args: args{
				ctx:    context.Background(),
				userId: 1,
				removeSegments: []string{
					"segment1",
				},
			},
			mockBehavior: func(m pgxmock.PgxPoolIface, args args) {
				m.ExpectBegin()
				m.ExpectQuery("SELECT id").
					WithArgs(args.userId).
					WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(1))
				m.ExpectQuery("SELECT name").
					WithArgs(args.removeSegments[0]).
					WillReturnRows(pgxmock.NewRows([]string{"name"}).AddRow(args.removeSegments[0]))
				m.ExpectExec("DELETE FROM user_segments").
					WithArgs(args.userId, args.removeSegments[0]).
					WillReturnError(errors.New("some error"))
				m.ExpectRollback()
			},
			wantErr: true,
		},
		{
			name: "INSERT into user_segments_log in removeSegments tx.Exec4 error",
			args: args{
				ctx:    context.Background(),
				userId: 1,
				removeSegments: []string{
					"segment1",
				},
			},
			mockBehavior: func(m pgxmock.PgxPoolIface, args args) {
				m.ExpectBegin()
				m.ExpectQuery("SELECT id").
					WithArgs(args.userId).
					WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(1))
				m.ExpectQuery("SELECT name").
					WithArgs(args.removeSegments[0]).
					WillReturnRows(pgxmock.NewRows([]string{"name"}).AddRow(args.removeSegments[0]))
				m.ExpectExec("DELETE FROM user_segments").
					WithArgs(args.userId, args.removeSegments[0]).
					WillReturnResult(pgxmock.NewResult("DELETE", 1))
				m.ExpectExec("INSERT INTO user_segments_log").
					WithArgs(args.userId, args.removeSegments[0], "delete").
					WillReturnError(errors.New("some error"))
				m.ExpectRollback()
			},
			wantErr: true,
		},
		{
			name: "tx.Commit error",
			args: args{
				ctx:    context.Background(),
				userId: 1,
			},
			mockBehavior: func(m pgxmock.PgxPoolIface, args args) {
				m.ExpectBegin()
				m.ExpectQuery("SELECT id").
					WithArgs(args.userId).
					WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(1))
				m.ExpectCommit().WillReturnError(errors.New("some error"))
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			poolMock, _ := pgxmock.NewPool()
			defer poolMock.Close()
			tc.mockBehavior(poolMock, tc.args)

			postgresMock := &postgres.Postgres{
				Builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
				Pool:    poolMock,
			}
			userRepoMock := NewUserRepo(postgresMock, MockTimeProvider{})

			err := userRepoMock.AddOrRemoveUserSegments(tc.args.ctx, tc.args.userId, tc.args.addSegments, tc.args.removeSegments)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			err = poolMock.ExpectationsWereMet()
			assert.NoError(t, err)
		})
	}
}
