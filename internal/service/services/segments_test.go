package services

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	mock_services "github.com/realPointer/segments/internal/service/mocks"
)

func TestSegmentsService_CreateSegment(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type input struct {
		ctx  context.Context
		name string
	}

	type output struct {
		err error
	}

	testCases := []struct {
		name           string
		input          input
		expectedOutput output
	}{
		{
			name: "success",
			input: input{
				ctx:  context.Background(),
				name: "segment1",
			},
			expectedOutput: output{
				err: nil,
			},
		},
		{
			name: "one letter",
			input: input{
				ctx:  context.Background(),
				name: "a",
			},
			expectedOutput: output{
				err: nil,
			},
		},
		{
			name: "one number",
			input: input{
				ctx:  context.Background(),
				name: "1",
			},
			expectedOutput: output{
				err: nil,
			},
		},
		{
			name: "only letters",
			input: input{
				ctx:  context.Background(),
				name: "segment",
			},
			expectedOutput: output{
				err: nil,
			},
		},
		{
			name: "only numbers",
			input: input{
				ctx:  context.Background(),
				name: "123",
			},
			expectedOutput: output{
				err: nil,
			},
		},
		{
			name: "contains special symbols",
			input: input{
				ctx:  context.Background(),
				name: "seg!@ment",
			},
			expectedOutput: output{
				err: nil,
			},
		},
		{
			name: "error",
			input: input{
				ctx:  context.Background(),
				name: "segment1",
			},
			expectedOutput: output{
				err: errors.New("error"),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockSegment := mock_services.NewMockSegment(ctrl)
			mockSegment.EXPECT().CreateSegment(tc.input.ctx, tc.input.name).Return(tc.expectedOutput.err)

			segmentService := NewSegmentService(mockSegment)

			err := segmentService.CreateSegment(tc.input.ctx, tc.input.name)

			assert.Equal(t, tc.expectedOutput.err, err)
		})
	}
}

func TestSegmentsService_CreateSegmentAuto(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type input struct {
		ctx        context.Context
		name       string
		percentage float64
	}

	type output struct {
		err error
	}

	testCases := []struct {
		name           string
		input          input
		expectedOutput output
	}{
		{
			name: "success",
			input: input{
				ctx:        context.Background(),
				name:       "segment1",
				percentage: 0.5,
			},
			expectedOutput: output{
				err: nil,
			},
		},
		{
			name: "error",
			input: input{
				ctx:        context.Background(),
				name:       "segment1",
				percentage: 0.5,
			},
			expectedOutput: output{
				err: errors.New("error"),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockSegment := mock_services.NewMockSegment(ctrl)
			mockSegment.EXPECT().CreateSegmentAuto(tc.input.ctx, tc.input.name, tc.input.percentage).Return(tc.expectedOutput.err)

			segmentService := NewSegmentService(mockSegment)

			err := segmentService.CreateSegmentAuto(tc.input.ctx, tc.input.name, tc.input.percentage)

			assert.Equal(t, tc.expectedOutput.err, err)
		})
	}
}

func TestSegmentsService_DeleteSegment(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type input struct {
		ctx  context.Context
		name string
	}

	type output struct {
		err error
	}

	testCases := []struct {
		name           string
		input          input
		expectedOutput output
	}{
		{
			name: "success",
			input: input{
				ctx:  context.Background(),
				name: "segment1",
			},
			expectedOutput: output{
				err: nil,
			},
		},
		{
			name: "error",
			input: input{
				ctx:  context.Background(),
				name: "segment1",
			},
			expectedOutput: output{
				err: errors.New("error"),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockSegment := mock_services.NewMockSegment(ctrl)
			mockSegment.EXPECT().DeleteSegment(tc.input.ctx, tc.input.name).Return(tc.expectedOutput.err)

			segmentService := NewSegmentService(mockSegment)

			err := segmentService.DeleteSegment(tc.input.ctx, tc.input.name)

			assert.Equal(t, tc.expectedOutput.err, err)
		})
	}
}

func TestSegmentsService_GetSegments(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type input struct {
		ctx context.Context
	}

	type output struct {
		segments []string
		err      error
	}

	testCases := []struct {
		name           string
		input          input
		expectedOutput output
	}{
		{
			name: "success",
			input: input{
				ctx: context.Background(),
			},
			expectedOutput: output{
				segments: []string{"segment1", "segment2"},
				err:      nil,
			},
		},
		{
			name: "error",
			input: input{
				ctx: context.Background(),
			},
			expectedOutput: output{
				segments: nil,
				err:      errors.New("error"),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockSegment := mock_services.NewMockSegment(ctrl)
			mockSegment.EXPECT().GetSegments(tc.input.ctx).Return(tc.expectedOutput.segments, tc.expectedOutput.err)

			segmentService := NewSegmentService(mockSegment)

			segments, err := segmentService.GetSegments(tc.input.ctx)

			assert.Equal(t, tc.expectedOutput.segments, segments)
			assert.Equal(t, tc.expectedOutput.err, err)
		})
	}
}
