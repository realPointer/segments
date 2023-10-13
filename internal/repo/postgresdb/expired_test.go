package postgresdb

import (
	"context"
	"errors"
	"testing"

	"github.com/Masterminds/squirrel"
	"github.com/pashagolub/pgxmock/v3"
	"github.com/realPointer/segments/pkg/postgres"
	"github.com/stretchr/testify/assert"
)

func TestExpiredRepo_DeleteExpiredRows(t *testing.T) {
	type args struct {
		ctx context.Context
	}

	type MockBehavior func(m pgxmock.PgxPoolIface, args args)

	testCases := []struct {
		name         string
		args         args
		mockBehavior MockBehavior
		want         int
		wantErr      bool
	}{
		{
			name: "OK",
			args: args{
				ctx: context.Background(),
			},
			mockBehavior: func(m pgxmock.PgxPoolIface, args args) {
				m.ExpectBegin()
				m.ExpectQuery("SELECT user_id, segment_name FROM user_segments WHERE expire IS NOT NULL AND expire < NOW()").
					WillReturnRows(pgxmock.NewRows([]string{"user_id", "segment_name"}).AddRow(1, "segment1"))
				m.ExpectExec("DELETE FROM user_segments WHERE expire IS NOT NULL AND expire < NOW()").
					WillReturnResult(pgxmock.NewResult("DELETE", 1))
				m.ExpectExec("INSERT INTO user_segments_log").
					WithArgs(1, "segment1", "delete").
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
				m.ExpectCommit()
			},
			want:    1,
			wantErr: false,
		},
		{
			name: "tx.Begin error",
			args: args{
				ctx: context.Background(),
			},
			mockBehavior: func(m pgxmock.PgxPoolIface, args args) {
				m.ExpectBegin().WillReturnError(errors.New("begin error"))
				m.ExpectRollback()
			},
			wantErr: true,
		},
		{
			name: "tx.Query error",
			args: args{
				ctx: context.Background(),
			},
			mockBehavior: func(m pgxmock.PgxPoolIface, args args) {
				m.ExpectBegin()
				m.ExpectQuery("SELECT user_id, segment_name FROM user_segments WHERE expire IS NOT NULL AND expire < NOW()").
					WillReturnError(errors.New("query error"))
				m.ExpectRollback()
			},
			wantErr: true,
		},
		{
			name: "rows.Scan error",
			args: args{
				ctx: context.Background(),
			},
			mockBehavior: func(m pgxmock.PgxPoolIface, args args) {
				m.ExpectBegin()
				m.ExpectQuery("SELECT user_id, segment_name FROM user_segments WHERE expire IS NOT NULL AND expire < NOW()").
					WillReturnRows(pgxmock.NewRows([]string{"user_id", "segment_name"}).AddRow(1, "segment1").RowError(0, errors.New("rows.Scan error")))
				m.ExpectRollback()
			},
			wantErr: true,
		},
		{
			name: "Delete user_segments tx.Exec error",
			args: args{
				ctx: context.Background(),
			},
			mockBehavior: func(m pgxmock.PgxPoolIface, args args) {
				m.ExpectBegin()
				m.ExpectQuery("SELECT user_id, segment_name FROM user_segments WHERE expire IS NOT NULL AND expire < NOW()").
					WillReturnRows(pgxmock.NewRows([]string{"user_id", "segment_name"}).AddRow(1, "segment1"))
				m.ExpectExec("DELETE FROM user_segments WHERE expire IS NOT NULL AND expire < NOW()").
					WillReturnError(errors.New("tx.Exec error"))
				m.ExpectRollback()
			},
			wantErr: true,
		},
		{
			name: "Insert user_segments_log tx.Exec error",
			args: args{
				ctx: context.Background(),
			},
			mockBehavior: func(m pgxmock.PgxPoolIface, args args) {
				m.ExpectBegin()
				m.ExpectQuery("SELECT user_id, segment_name FROM user_segments WHERE expire IS NOT NULL AND expire < NOW()").
					WillReturnRows(pgxmock.NewRows([]string{"user_id", "segment_name"}).AddRow(1, "segment1"))
				m.ExpectExec("DELETE FROM user_segments WHERE expire IS NOT NULL AND expire < NOW()").
					WillReturnResult(pgxmock.NewResult("DELETE", 1))
				m.ExpectExec("INSERT INTO user_segments_log").
					WithArgs(1, "segment1", "delete").
					WillReturnError(errors.New("tx.Exec error"))
				m.ExpectRollback()
			},
			wantErr: true,
		},
		{
			name: "tx.Commit error",
			args: args{
				ctx: context.Background(),
			},
			mockBehavior: func(m pgxmock.PgxPoolIface, args args) {
				m.ExpectBegin()
				m.ExpectQuery("SELECT user_id, segment_name FROM user_segments WHERE expire IS NOT NULL AND expire < NOW()").
					WillReturnRows(pgxmock.NewRows([]string{"user_id", "segment_name"}).AddRow(1, "segment1"))
				m.ExpectExec("DELETE FROM user_segments WHERE expire IS NOT NULL AND expire < NOW()").
					WillReturnResult(pgxmock.NewResult("DELETE", 1))
				m.ExpectExec("INSERT INTO user_segments_log").
					WithArgs(1, "segment1", "delete").
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
				m.ExpectCommit().WillReturnError(errors.New("commit error"))
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
			expiredRepoMock := NewExpiredRepo(postgresMock)

			got, err := expiredRepoMock.DeleteExpiredRows(tc.args.ctx)
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
