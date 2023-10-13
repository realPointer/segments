package postgresdb

import (
	"context"
	"errors"
	"testing"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pashagolub/pgxmock/v3"
	"github.com/realPointer/segments/pkg/postgres"
	"github.com/stretchr/testify/assert"
)

func TestSegmentRepo_CreateSegment(t *testing.T) {
	type args struct {
		ctx  context.Context
		name string
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
				ctx:  context.Background(),
				name: "test_segment",
			},
			mockBehavior: func(m pgxmock.PgxPoolIface, args args) {
				m.ExpectExec("INSERT INTO segments").
					WithArgs(args.name).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
			},
			wantErr: false,
		},
		{
			name: "segment already exists",
			args: args{
				ctx:  context.Background(),
				name: "test_segment",
			},
			mockBehavior: func(m pgxmock.PgxPoolIface, args args) {
				m.ExpectExec("INSERT INTO segments").
					WithArgs(args.name).
					WillReturnError(&pgconn.PgError{
						Code: "23505",
					})
			},
			wantErr: true,
		},
		{
			name: "unexpected error",
			args: args{
				ctx:  context.Background(),
				name: "test_segment",
			},
			mockBehavior: func(m pgxmock.PgxPoolIface, args args) {
				m.ExpectExec("INSERT INTO segments").
					WithArgs(args.name).
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
			segmentRepoMock := NewSegmentRepo(postgresMock)

			err := segmentRepoMock.CreateSegment(tc.args.ctx, tc.args.name)
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

func TestSegmentRepo_CreateSegmentAuto(t *testing.T) {
	type args struct {
		ctx        context.Context
		name       string
		percentage float64
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
				ctx:        context.Background(),
				name:       "test_segment",
				percentage: 10,
			},
			mockBehavior: func(m pgxmock.PgxPoolIface, args args) {
				m.ExpectBegin()
				m.ExpectExec("INSERT INTO segments").
					WithArgs(args.name, args.percentage).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
				m.ExpectQuery("SELECT id").
					WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(1).AddRow(2))
				m.ExpectExec("INSERT INTO user_segments").
					WithArgs(1, args.name).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
				m.ExpectExec("INSERT INTO user_segments_log").
					WithArgs(1, args.name, "add").
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
				m.ExpectExec("INSERT INTO user_segments").
					WithArgs(2, args.name).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
				m.ExpectExec("INSERT INTO user_segments_log").
					WithArgs(2, args.name, "add").
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
				m.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name: "transaction error",
			args: args{
				ctx:        context.Background(),
				name:       "test_segment",
				percentage: 10,
			},
			mockBehavior: func(m pgxmock.PgxPoolIface, args args) {
				m.ExpectBegin().WillReturnError(errors.New("some error"))
				m.ExpectRollback()
			},
			wantErr: true,
		},
		{
			name: "segment already exists",
			args: args{
				ctx:        context.Background(),
				name:       "test_segment",
				percentage: 10,
			},
			mockBehavior: func(m pgxmock.PgxPoolIface, args args) {
				m.ExpectBegin()
				m.ExpectExec("INSERT INTO segments").
					WithArgs(args.name, args.percentage).
					WillReturnError(&pgconn.PgError{
						Code: "23505",
					})
				m.ExpectRollback()
			},
			wantErr: true,
		},
		{
			name: "r.Pool.Exec error",
			args: args{
				ctx:        context.Background(),
				name:       "test_segment",
				percentage: 10,
			},
			mockBehavior: func(m pgxmock.PgxPoolIface, args args) {
				m.ExpectBegin()
				m.ExpectExec("INSERT INTO segments").
					WithArgs(args.name, args.percentage).
					WillReturnError(errors.New("some error"))
				m.ExpectRollback()
			},
			wantErr: true,
		},
		{
			name: "tx.Query error",
			args: args{
				ctx:        context.Background(),
				name:       "test_segment",
				percentage: 10,
			},
			mockBehavior: func(m pgxmock.PgxPoolIface, args args) {
				m.ExpectBegin()
				m.ExpectExec("INSERT INTO segments").
					WithArgs(args.name, args.percentage).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
				m.ExpectQuery("SELECT id").
					WillReturnError(errors.New("some error"))
				m.ExpectRollback()
			},
			wantErr: true,
		},
		{
			name: "rows.Scan error",
			args: args{
				ctx:        context.Background(),
				name:       "test_segment",
				percentage: 10,
			},
			mockBehavior: func(m pgxmock.PgxPoolIface, args args) {
				m.ExpectBegin()
				m.ExpectExec("INSERT INTO segments").
					WithArgs(args.name, args.percentage).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
				m.ExpectQuery("SELECT id").
					WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(1).RowError(0, errors.New("rows.Scan error")))
				m.ExpectRollback()
			},
			wantErr: true,
		},
		{
			name: "tx.Exec1 error",
			args: args{
				ctx:        context.Background(),
				name:       "test_segment",
				percentage: 10,
			},
			mockBehavior: func(m pgxmock.PgxPoolIface, args args) {
				m.ExpectBegin()
				m.ExpectExec("INSERT INTO segments").
					WithArgs(args.name, args.percentage).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
				m.ExpectQuery("SELECT id").
					WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(1).AddRow(2))
				m.ExpectExec("INSERT INTO user_segments").
					WithArgs(1, args.name).
					WillReturnError(errors.New("some error"))
				m.ExpectRollback()
			},
			wantErr: true,
		},
		{
			name: "tx.Exec2 error",
			args: args{
				ctx:        context.Background(),
				name:       "test_segment",
				percentage: 10,
			},
			mockBehavior: func(m pgxmock.PgxPoolIface, args args) {
				m.ExpectBegin()
				m.ExpectExec("INSERT INTO segments").
					WithArgs(args.name, args.percentage).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
				m.ExpectQuery("SELECT id").
					WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(1).AddRow(2))
				m.ExpectExec("INSERT INTO user_segments").
					WithArgs(1, args.name).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
				m.ExpectExec("INSERT INTO user_segments_log").
					WithArgs(1, args.name, "add").
					WillReturnError(errors.New("some error"))
				m.ExpectRollback()
			},
			wantErr: true,
		},
		{
			name: "tx.Commit error",
			args: args{
				ctx:        context.Background(),
				name:       "test_segment",
				percentage: 10,
			},
			mockBehavior: func(m pgxmock.PgxPoolIface, args args) {
				m.ExpectBegin()
				m.ExpectExec("INSERT INTO segments").
					WithArgs(args.name, args.percentage).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
				m.ExpectQuery("SELECT id").
					WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(1).AddRow(2))
				m.ExpectExec("INSERT INTO user_segments").
					WithArgs(1, args.name).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
				m.ExpectExec("INSERT INTO user_segments_log").
					WithArgs(1, args.name, "add").
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
				m.ExpectExec("INSERT INTO user_segments").
					WithArgs(2, args.name).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
				m.ExpectExec("INSERT INTO user_segments_log").
					WithArgs(2, args.name, "add").
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
				m.ExpectCommit().WillReturnError(errors.New("some error"))
				m.ExpectRollback()
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
			segmentRepoMock := NewSegmentRepo(postgresMock)

			err := segmentRepoMock.CreateSegmentAuto(tc.args.ctx, tc.args.name, tc.args.percentage)
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

func TestSegmentRepo_countUsers(t *testing.T) {
	type args struct {
		ctx context.Context
	}

	type MockBehavior func(m pgxmock.PgxPoolIface, args args)

	testCases := []struct {
		name         string
		args         args
		mockBehavior MockBehavior
		want         int
	}{
		{
			name: "OK",
			args: args{
				ctx: context.Background(),
			},
			mockBehavior: func(m pgxmock.PgxPoolIface, args args) {
				m.ExpectQuery("SELECT count").
					WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(10))
			},
			want: 10,
		},
		{
			name: "unexpected error",
			args: args{
				ctx: context.Background(),
			},
			mockBehavior: func(m pgxmock.PgxPoolIface, args args) {
				m.ExpectQuery("SELECT count").
					WillReturnError(errors.New("some error"))
			},
			want: 0,
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
			segmentRepoMock := NewSegmentRepo(postgresMock)

			got := segmentRepoMock.countUsers(tc.args.ctx)

			assert.Equal(t, tc.want, got)
		})
	}
}

func TestSegmentRepo_DeleteSegment(t *testing.T) {
	type args struct {
		ctx  context.Context
		name string
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
				ctx:  context.Background(),
				name: "test_segment",
			},
			mockBehavior: func(m pgxmock.PgxPoolIface, args args) {
				m.ExpectBegin()
				m.ExpectQuery("SELECT us.user_id").
					WithArgs(args.name).
					WillReturnRows(pgxmock.NewRows([]string{"user_id"}).AddRow(1).AddRow(2))
				m.ExpectExec("user_segments_log").
					WithArgs(1, args.name, "delete").
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
				m.ExpectExec("user_segments_log").
					WithArgs(2, args.name, "delete").
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
				m.ExpectExec("DELETE FROM segments").
					WithArgs(args.name).
					WillReturnResult(pgxmock.NewResult("DELETE", 1))
				m.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name: "transaction error",
			args: args{
				ctx:  context.Background(),
				name: "test_segment",
			},
			mockBehavior: func(m pgxmock.PgxPoolIface, args args) {
				m.ExpectBegin().WillReturnError(errors.New("some error"))
			},
			wantErr: true,
		},
		{
			name: "segment not found",
			args: args{
				ctx:  context.Background(),
				name: "test_segment",
			},
			mockBehavior: func(m pgxmock.PgxPoolIface, args args) {
				m.ExpectBegin()
				m.ExpectQuery("SELECT us.user_id").
					WithArgs(args.name).
					WillReturnError(&pgconn.PgError{
						Code: "23505",
					})
				m.ExpectRollback()
			},
			wantErr: true,
		},
		{
			name: "r.Pool.Query error",
			args: args{
				ctx:  context.Background(),
				name: "test_segment",
			},
			mockBehavior: func(m pgxmock.PgxPoolIface, args args) {
				m.ExpectBegin()
				m.ExpectQuery("SELECT us.user_id").
					WithArgs(args.name).
					WillReturnError(errors.New("some error"))
				m.ExpectRollback()
			},
			wantErr: true,
		},
		{
			name: "rows.Scan error",
			args: args{
				ctx:  context.Background(),
				name: "test_segment",
			},
			mockBehavior: func(m pgxmock.PgxPoolIface, args args) {
				m.ExpectBegin()
				m.ExpectQuery("SELECT us.user_id").
					WithArgs(args.name).
					WillReturnRows(pgxmock.NewRows([]string{"user_id"}).AddRow(1).RowError(0, errors.New("rows.Scan error")))
				m.ExpectRollback()
			},
			wantErr: true,
		},
		{
			name: "tx.Exec1 error",
			args: args{
				ctx:  context.Background(),
				name: "test_segment",
			},
			mockBehavior: func(m pgxmock.PgxPoolIface, args args) {
				m.ExpectBegin()
				m.ExpectQuery("SELECT us.user_id").
					WithArgs(args.name).
					WillReturnRows(pgxmock.NewRows([]string{"user_id"}).AddRow(1).AddRow(2))
				m.ExpectExec("user_segments_log").
					WithArgs(1, args.name, "delete").
					WillReturnError(errors.New("some error"))
				m.ExpectRollback()
			},
			wantErr: true,
		},
		{
			name: "tx.Exec2 error",
			args: args{
				ctx:  context.Background(),
				name: "test_segment",
			},
			mockBehavior: func(m pgxmock.PgxPoolIface, args args) {
				m.ExpectBegin()
				m.ExpectQuery("SELECT us.user_id").
					WithArgs(args.name).
					WillReturnRows(pgxmock.NewRows([]string{"user_id"}).AddRow(1).AddRow(2))
				m.ExpectExec("user_segments_log").
					WithArgs(1, args.name, "delete").
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
				m.ExpectExec("user_segments_log").
					WithArgs(2, args.name, "delete").
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
				m.ExpectExec("DELETE FROM segments").
					WithArgs(args.name).
					WillReturnError(errors.New("some error"))
				m.ExpectRollback()
			},
			wantErr: true,
		},
		{
			name: "tx.Commit error",
			args: args{
				ctx:  context.Background(),
				name: "test_segment",
			},
			mockBehavior: func(m pgxmock.PgxPoolIface, args args) {
				m.ExpectBegin()
				m.ExpectQuery("SELECT us.user_id").
					WithArgs(args.name).
					WillReturnRows(pgxmock.NewRows([]string{"user_id"}).AddRow(1).AddRow(2))
				m.ExpectExec("user_segments_log").
					WithArgs(1, args.name, "delete").
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
				m.ExpectExec("user_segments_log").
					WithArgs(2, args.name, "delete").
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
				m.ExpectExec("DELETE FROM segments").
					WithArgs(args.name).
					WillReturnResult(pgxmock.NewResult("DELETE", 1))
				m.ExpectCommit().WillReturnError(errors.New("some error"))
				m.ExpectRollback()
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
			segmentRepoMock := NewSegmentRepo(postgresMock)

			err := segmentRepoMock.DeleteSegment(tc.args.ctx, tc.args.name)
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

func TestSegmentRepo_GetSegments(t *testing.T) {
	type args struct {
		ctx context.Context
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
			name: "1 segment OK",
			args: args{
				ctx: context.Background(),
			},
			mockBehavior: func(m pgxmock.PgxPoolIface, args args) {
				m.ExpectQuery("SELECT name FROM segments").
					WillReturnRows(pgxmock.NewRows([]string{"name"}).AddRow("test_segment"))
			},
			want:    []string{"test_segment"},
			wantErr: false,
		},
		{
			name: "some segments OK",
			args: args{
				ctx: context.Background(),
			},
			mockBehavior: func(m pgxmock.PgxPoolIface, args args) {
				rows := pgxmock.NewRows([]string{"name"}).
					AddRow("test_segment_1").
					AddRow("test_segment_2").
					AddRow("test_segment_3")

				m.ExpectQuery("SELECT name FROM segments").
					WillReturnRows(rows)
			},
			want:    []string{"test_segment_1", "test_segment_2", "test_segment_3"},
			wantErr: false,
		},
		{
			name: "empty result",
			args: args{
				ctx: context.Background(),
			},
			mockBehavior: func(m pgxmock.PgxPoolIface, args args) {
				m.ExpectQuery("SELECT name FROM segments").
					WillReturnRows(pgxmock.NewRows([]string{"name"}))
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "rows.Scan error",
			args: args{
				ctx: context.Background(),
			},
			mockBehavior: func(m pgxmock.PgxPoolIface, args args) {
				rows := pgxmock.NewRows([]string{"segment"}).
					AddRow("segment1").
					AddRow("segment2").
					RowError(1, errors.New("rows.Scan error"))
				m.ExpectQuery("SELECT name FROM segments").WillReturnRows(rows)
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "unexpected error",
			args: args{
				ctx: context.Background(),
			},
			mockBehavior: func(m pgxmock.PgxPoolIface, args args) {
				m.ExpectQuery("SELECT name FROM segments").
					WillReturnError(errors.New("some error"))
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
			segmentRepoMock := NewSegmentRepo(postgresMock)

			got, err := segmentRepoMock.GetSegments(tc.args.ctx)
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
