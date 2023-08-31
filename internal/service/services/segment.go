package services

import (
	"context"

	"github.com/realPointer/segments/internal/repo"
)

type SegmentService struct {
	segmentRepo repo.Segment
}

func NewSegmentService(segmentRepo repo.Segment) *SegmentService {
	return &SegmentService{segmentRepo: segmentRepo}
}

func (s *SegmentService) CreateSegment(ctx context.Context, name string) error {
	return s.segmentRepo.CreateSegment(ctx, name)
}

func (s *SegmentService) CreateSegmentAuto(ctx context.Context, name string, percentage float64) error {
	return s.segmentRepo.CreateSegmentAuto(ctx, name, percentage)
}

func (s *SegmentService) DeleteSegment(ctx context.Context, name string) error {
	return s.segmentRepo.DeleteSegment(ctx, name)
}

func (s *SegmentService) GetSegments(ctx context.Context) ([]string, error) {
	return s.segmentRepo.GetSegments(ctx)
}
