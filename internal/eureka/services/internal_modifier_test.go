package services

import (
	"context"
	"fmt"
	"testing"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	mock_repositories "github.com/manabie-com/backend/mock/eureka/repositories"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestInternalModifierService_DeleteLOStudyPlanItems(t *testing.T) {
	t.Parallel()
	LoStudyPlanItemRepo := &mock_repositories.MockLoStudyPlanItemRepo{}
	StudyPlanItemRepo := &mock_repositories.MockStudyPlanItemRepo{}
	mockDB := &mock_database.Ext{}
	//tx := &mock_database.Tx{}

	srv := &InternalModifierService{
		DB:                  mockDB,
		LoStudyPlanItemRepo: LoStudyPlanItemRepo,
		StudyPlanItemRepo:   StudyPlanItemRepo,
	}

	loids := []string{"lo-id-1", "lo-id-2"}

	validReq := &epb.DeleteLOStudyPlanItemsRequest{
		LoIds: loids,
	}

	invalidReq := &epb.DeleteLOStudyPlanItemsRequest{}

	testCases := []TestCase{
		{
			name:        "case LoIDs empty",
			req:         invalidReq,
			expectedErr: status.Error(codes.InvalidArgument, "lo_ids must not be empty"),
			setup: func(ctx context.Context) {
				//tx.On("Exec", ctx, validReq.LoIds).Once().Return(mock.Anything, nil)
				LoStudyPlanItemRepo.On("DeleteLoStudyPlanItemsAndStudyPlanItemByLoIDs", ctx, mock.Anything, mock.Anything).Once().Return(nil)
			},
		},
		{
			name:        "happy case",
			req:         validReq,
			expectedErr: nil,
			setup: func(ctx context.Context) {
				//tx.On("Exec", ctx, validReq.LoIds).Once().Return(mock.Anything, nil)
				LoStudyPlanItemRepo.On("DeleteLoStudyPlanItemsAndStudyPlanItemByLoIDs", ctx, mock.Anything, mock.Anything).Once().Return(nil)
			},
		},
	}
	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			ctx := context.Background()
			testCase.setup(ctx)
			_, err := srv.DeleteLOStudyPlanItems(ctx, testCase.req.(*epb.DeleteLOStudyPlanItemsRequest))
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr, err)
			}
		})

	}
}

func TestInternalModifierService_UpsertAdHocIndividualStudyPlan(t *testing.T) {
	t.Parallel()
	LoStudyPlanItemRepo := &mock_repositories.MockLoStudyPlanItemRepo{}
	StudyPlanItemRepo := &mock_repositories.MockStudyPlanItemRepo{}
	StudyPlanRepo := &mock_repositories.MockStudyPlanRepo{}
	CourseBookRepo := &mock_repositories.MockCourseBookRepo{}
	StudentStudyPlanRepo := &mock_repositories.MockStudentStudyPlanRepo{}
	BookRepo := &mock_repositories.MockBookRepo{}
	AssignmentRepo := &mock_repositories.MockAssignmentRepo{}
	learningObjectRepo := &mock_repositories.MockLearningObjectiveRepo{}
	AssignmentStudyPlanItemRepo := &mock_repositories.MockAssignmentStudyPlanItemRepo{}
	mockDB := &mock_database.Ext{}
	tx := &mock_database.Tx{}

	srv := &InternalModifierService{
		DB:                          mockDB,
		AssignmentRepo:              AssignmentRepo,
		AssignmentStudyPlanItemRepo: AssignmentStudyPlanItemRepo,
		LoStudyPlanItemRepo:         LoStudyPlanItemRepo,
		StudyPlanItemRepo:           StudyPlanItemRepo,
		StudyPlanRepo:               StudyPlanRepo,
		StudentStudyPlanRepo:        StudentStudyPlanRepo,
		CourseBookRepo:              CourseBookRepo,
		LearningObjectiveRepo:       learningObjectRepo,
		BookRepo:                    BookRepo,
	}

	testCases := []TestCase{
		{
			name: "missing book id",
			req: &epb.UpsertAdHocIndividualStudyPlanRequest{
				BookId: "",
			},
			expectedErr: status.Errorf(codes.InvalidArgument, "req must have book id"),
			setup:       func(ctx context.Context) {},
		},
		{
			name: "missing course id",
			req: &epb.UpsertAdHocIndividualStudyPlanRequest{
				BookId:   "book-id",
				CourseId: "",
			},
			expectedErr: status.Errorf(codes.InvalidArgument, "req must have course id"),
			setup:       func(ctx context.Context) {},
		},
		{
			name: "missing student id",
			req: &epb.UpsertAdHocIndividualStudyPlanRequest{
				BookId:    "book-id",
				CourseId:  "course-id",
				StudentId: "",
			},
			expectedErr: status.Errorf(codes.InvalidArgument, "req must have student id"),
			setup:       func(ctx context.Context) {},
		},
		{
			name: "unable to retrieve course book by course id and book id",
			req: &epb.UpsertAdHocIndividualStudyPlanRequest{
				BookId:    "book-id",
				CourseId:  "course-id",
				StudentId: "student-id",
			},
			expectedErr: status.Errorf(codes.Internal, fmt.Errorf("unable to retrieve course book by course id and book id: %w", pgx.ErrTxClosed).Error()),
			setup: func(ctx context.Context) {
				CourseBookRepo.On("FindByCourseIDAndBookID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, pgx.ErrTxClosed)
			},
		},
		{
			name: "book must have type adhoc",
			req: &epb.UpsertAdHocIndividualStudyPlanRequest{
				BookId:    "book-id",
				CourseId:  "course-id",
				StudentId: "student-id",
			},
			expectedErr: status.Errorf(codes.Internal, "book must have type adhoc"),
			setup: func(ctx context.Context) {
				CourseBookRepo.On("FindByCourseIDAndBookID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, nil)
				BookRepo.On("FindByID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(&entities.Book{
					ID:       database.Text("book-id"),
					BookType: database.Text(cpb.BookType_BOOK_TYPE_GENERAL.String()),
				}, nil)
			},
		},
		{
			name: "err StudyPlanRepo.FindByID",
			req: &epb.UpsertAdHocIndividualStudyPlanRequest{
				StudyPlanId: &wrapperspb.StringValue{
					Value: "studyPlan-id",
				},
				BookId:    "book-id",
				CourseId:  "course-id",
				StudentId: "student-id",
			},
			expectedErr: status.Errorf(codes.Internal, fmt.Errorf("studyPlanRepo.FindByID: %w", pgx.ErrTxClosed).Error()),
			setup: func(ctx context.Context) {
				CourseBookRepo.On("FindByCourseIDAndBookID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, nil)
				BookRepo.On("FindByID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(&entities.Book{
					ID:       database.Text("book-id"),
					BookType: database.Text(cpb.BookType_BOOK_TYPE_ADHOC.String()),
				}, nil)
				StudyPlanRepo.On("FindByID", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, pgx.ErrTxClosed)
			},
		},
		{
			name: "err StudyPlanRepo.BulkUpsert",
			req: &epb.UpsertAdHocIndividualStudyPlanRequest{
				StudyPlanId: &wrapperspb.StringValue{
					Value: "studyPlan-id",
				},
				BookId:    "book-id",
				CourseId:  "course-id",
				StudentId: "student-id",
			},
			expectedErr: status.Errorf(codes.Internal, fmt.Errorf("studyPlanRepo.BulkUpsert: %w", pgx.ErrTxClosed).Error()),
			setup: func(ctx context.Context) {
				CourseBookRepo.On("FindByCourseIDAndBookID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, nil)
				BookRepo.On("FindByID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(&entities.Book{
					ID:       database.Text("book-id"),
					BookType: database.Text(cpb.BookType_BOOK_TYPE_ADHOC.String()),
				}, nil)
				StudyPlanRepo.On("FindByID", mock.Anything, mock.Anything, mock.Anything).Once().Return(&entities.StudyPlan{}, nil)
				StudyPlanRepo.On("BulkUpsert", mock.Anything, mock.Anything, mock.Anything).Once().Return(pgx.ErrTxClosed)
			},
		},
		{
			name: "err StudentStudyPlanRepo.BulkUpsert",
			req: &epb.UpsertAdHocIndividualStudyPlanRequest{
				BookId:    "book-id",
				CourseId:  "course-id",
				StudentId: "student-id",
			},
			expectedErr: status.Errorf(codes.Internal, fmt.Errorf("studentStudyPlan.BulkUpsert: %w", pgx.ErrTxClosed).Error()),
			setup: func(ctx context.Context) {
				learningObjectRepo.On("RetrieveLearningObjectivesByTopicIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, nil)
				CourseBookRepo.On("FindByCourseIDAndBookID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, nil)
				BookRepo.On("FindByID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(&entities.Book{
					ID:       database.Text("book-id"),
					BookType: database.Text(cpb.BookType_BOOK_TYPE_ADHOC.String()),
				}, nil)
				StudyPlanRepo.On("BulkUpsert", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				mockDB.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
				tx.On("Rollback", mock.Anything).Return(nil)
				StudentStudyPlanRepo.On("BulkUpsert", mock.Anything, mock.Anything, mock.Anything).Once().Return(pgx.ErrTxClosed)
			},
		},
		{
			name: "happy case",
			req: &epb.UpsertAdHocIndividualStudyPlanRequest{
				BookId:    "book-id",
				CourseId:  "course-id",
				StudentId: "student-id",
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				CourseBookRepo.On("FindByCourseIDAndBookID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, nil)
				BookRepo.On("FindByID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(&entities.Book{
					ID:       database.Text("book-id"),
					BookType: database.Text(cpb.BookType_BOOK_TYPE_ADHOC.String()),
				}, nil)
				StudyPlanRepo.On("BulkUpsert", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				mockDB.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
				tx.On("Commit", mock.Anything).Return(nil)
				StudentStudyPlanRepo.On("BulkUpsert", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				BookRepo.On("RetrieveBookTreeByBookID", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, nil)
				AssignmentRepo.On("RetrieveAssignmentsByTopicIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, nil)
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			ctx := context.Background()
			testCase.setup(ctx)
			_, err := srv.UpsertAdHocIndividualStudyPlan(ctx, testCase.req.(*epb.UpsertAdHocIndividualStudyPlanRequest))
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr, err)
			}
		})

	}
}
