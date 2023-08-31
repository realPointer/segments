package services

import (
	"context"

	"github.com/realPointer/segments/internal/entity"
	"github.com/realPointer/segments/internal/repo"
	webapi "github.com/realPointer/segments/internal/ydisk"
)

type UserService struct {
	userRepo repo.User
	yDisk    webapi.Disk
}

func NewUserService(userRepo repo.User, yDisk webapi.Disk) *UserService {
	return &UserService{
		userRepo: userRepo,
		yDisk:    yDisk,
	}
}

func (s *UserService) CreateUser(ctx context.Context, userId int) error {
	return s.userRepo.CreateUser(ctx, userId)
}

func (s *UserService) GetUserSegments(ctx context.Context, userId int) ([]string, error) {
	return s.userRepo.GetUserSegments(ctx, userId)
}

func (s *UserService) AddOrRemoveUserSegments(ctx context.Context, userId int, addSegments []entity.AddSegment, removeSegments []string) error {
	return s.userRepo.AddOrRemoveUserSegments(ctx, userId, addSegments, removeSegments)
}

func (s *UserService) GetUserOperations(ctx context.Context, userId int) ([]string, error) {
	return s.userRepo.GetUserOperations(ctx, userId)
}

func (s *UserService) GetUserOperationsByMonth(ctx context.Context, userId int, yearMonth string) ([]string, error) {
	return s.userRepo.GetUserOperationsByMonth(ctx, userId, yearMonth)
}

func (s *UserService) UploadAndReturnDownloadURL(ctx context.Context, name string, data []string) (string, error) {
	return s.yDisk.UploadAndReturnDownloadURL(ctx, name, data)
}
