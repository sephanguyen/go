package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	mock_repositories "github.com/manabie-com/backend/mock/eureka/repositories"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"

	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestStudentAdHocFactory_SetStudent(t *testing.T) {
	t.Parallel()
	factory := &StudentAdHocFactory{}

	testCases := []TestCase{
		{
			name: "student with invalid grade",
			req: &bpb.StudentProfile{
				Country: cpb.Country_COUNTRY_VN,
				Grade:   "invalid-grade",
			},
			expectedErr: fmt.Errorf("i18n.ConvertStringGradeToInt: %w", status.Error(codes.InvalidArgument, "cannot find grade in map")),
		},
		{
			name: "happy case",
			req: &bpb.StudentProfile{
				Country: cpb.Country_COUNTRY_VN,
				Grade:   "Lớp 1",
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			err := factory.SetStudent(testCase.req.(*bpb.StudentProfile))
			assert.Equal(t, testCase.expectedErr, err)
		})
	}
}

func TestStudentAdHocFactory_CreateAdHocBookContent(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	t.Parallel()
	mockDB := &mock_database.Ext{}
	mockBookRepo := new(mock_repositories.MockBookRepo)
	mockChapterRepo := new(mock_repositories.MockChapterRepo)
	mockBookChapterRepo := new(mock_repositories.MockBookChapterRepo)
	mockTopicRepo := new(mock_repositories.MockTopicRepo)
	factory := &StudentAdHocFactory{
		DB:              mockDB,
		BookRepo:        mockBookRepo,
		ChapterRepo:     mockChapterRepo,
		BookChapterRepo: mockBookChapterRepo,
		TopicRepo:       mockTopicRepo,
	}
	validStudent := &bpb.StudentProfile{
		Id:      "student-id",
		Name:    "Alice",
		Country: cpb.Country_COUNTRY_VN,
		Grade:   "Lớp 1",
		School:  &bpb.School{},
	}
	validInput := CreateBookContentInput{
		BookName:    "",
		ChapterName: "",
		TopicName:   "",
	}
	testCases := []TestCase{
		{
			name:        "error missing student",
			setup:       func(ctx context.Context) {},
			req:         validInput,
			expectedErr: fmt.Errorf("student profile must not be nil"),
		},
		{
			name: "error upsert book",
			req:  validInput,
			setup: func(ctx context.Context) {
				factory.Student = validStudent
				mockBookRepo.On("Upsert", mock.Anything, mockDB, mock.Anything).Once().Return(pgx.ErrTxClosed)
			},
			expectedErr: fmt.Errorf("s.BookRepo.Upsert: %w", pgx.ErrTxClosed),
		},
		{
			name: "error upsert chapter",
			setup: func(ctx context.Context) {
				factory.Student = validStudent
				mockBookRepo.On("Upsert", mock.Anything, mockDB, mock.Anything).Once().Return(nil)
				mockChapterRepo.On("Upsert", mock.Anything, mock.Anything, mock.AnythingOfType("[]*entities.Chapter")).Once().Return(pgx.ErrTxClosed)
			},
			req:         validInput,
			expectedErr: fmt.Errorf("s.ChapterRepo.Upsert: %w", pgx.ErrTxClosed),
		},
		{
			name: "error upsert book_chapter",
			setup: func(ctx context.Context) {
				factory.Student = validStudent
				mockBookRepo.On("Upsert", mock.Anything, mockDB, mock.Anything).Once().Return(nil)
				mockChapterRepo.On("Upsert", mock.Anything, mock.Anything, mock.AnythingOfType("[]*entities.Chapter")).Once().Return(nil)
				mockBookChapterRepo.On("Upsert", mock.Anything, mock.Anything, mock.AnythingOfType("[]*entities.BookChapter")).Once().Return(pgx.ErrTxClosed)
			},
			req:         validInput,
			expectedErr: fmt.Errorf("s.BookChapterRepo.Upsert: %w", pgx.ErrTxClosed),
		},
		{
			name: "error upsert topic",
			setup: func(ctx context.Context) {
				factory.Student = validStudent
				mockBookRepo.On("Upsert", mock.Anything, mockDB, mock.Anything).Once().Return(nil)
				mockChapterRepo.On("Upsert", mock.Anything, mock.Anything, mock.AnythingOfType("[]*entities.Chapter")).Once().Return(nil)
				mockBookChapterRepo.On("Upsert", mock.Anything, mock.Anything, mock.AnythingOfType("[]*entities.BookChapter")).Once().Return(nil)
				mockTopicRepo.On("BulkImport", mock.Anything, mock.Anything, mock.AnythingOfType("[]*entities.Topic")).Once().Return(pgx.ErrTxClosed)
			},
			req:         validInput,
			expectedErr: fmt.Errorf("s.TopicRepo.BulkImport: %w", pgx.ErrTxClosed),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			result, err := factory.CreateAdHocBookContent(ctx, testCase.req.(CreateBookContentInput))
			assert.Equal(t, testCase.expectedErr, err)
			if testCase.expectedErr == nil {
				assert.NotNil(t, result)
				assert.IsType(t, mock.AnythingOfType("*entities.Book"), result.Book)
				assert.IsType(t, mock.AnythingOfType("*entities.Chapter"), result.Chapter)
				assert.IsType(t, mock.AnythingOfType("*entities.Topic"), result.Topic)

			}
		})
	}
}
