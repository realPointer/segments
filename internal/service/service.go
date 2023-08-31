package service

import (
	"context"

	"github.com/realPointer/segments/internal/entity"
	"github.com/realPointer/segments/internal/repo"
	"github.com/realPointer/segments/internal/service/services"
	webapi "github.com/realPointer/segments/internal/ydisk"
)

type User interface {
	CreateUser(ctx context.Context, userId int) error
	GetUserSegments(ctx context.Context, userId int) ([]string, error)
	AddOrRemoveUserSegments(ctx context.Context, userId int, addSegments []entity.AddSegment, removeSegments []string) error
	GetUserOperations(ctx context.Context, userId int) ([]string, error)
	GetUserOperationsByMonth(ctx context.Context, userId int, yearMonth string) ([]string, error)
	UploadAndReturnDownloadURL(ctx context.Context, name string, data []string) (string, error)
}

type Segment interface {
	CreateSegment(ctx context.Context, name string) error
	CreateSegmentAuto(ctx context.Context, name string, percentage float64) error
	DeleteSegment(ctx context.Context, name string) error
	GetSegments(ctx context.Context) ([]string, error)
}

type Scheduler interface {
	DeleteExpiredRows(ctx context.Context) (int, error)
}

type Services struct {
	User
	Segment
	Scheduler
}

type ServicesDependencies struct {
	Repos      *repo.Repositories
	YandexDisk webapi.Disk
}

func NewServices(deps ServicesDependencies) *Services {
	return &Services{
		User:      services.NewUserService(deps.Repos.User, deps.YandexDisk),
		Segment:   services.NewSegmentService(deps.Repos.Segment),
		Scheduler: services.NewSheduler(deps.Repos.Expired),
	}
}
