package services

import (
	"context"

	"github.com/realPointer/segments/internal/repo"
)

type Scheduler struct {
	expiredStorage repo.Expired
}

func NewSheduler(expiredStorage repo.Expired) *Scheduler {
	return &Scheduler{
		expiredStorage: expiredStorage,
	}
}

func (s *Scheduler) DeleteExpiredRows(ctx context.Context) (int, error) {
	return s.expiredStorage.DeleteExpiredRows(ctx)
}
