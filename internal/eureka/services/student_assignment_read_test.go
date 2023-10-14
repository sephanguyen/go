package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	mock_repositories "github.com/manabie-com/backend/mock/eureka/repositories"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestListSubmissions(t *testing.T) {
	t.Parallel()
	submissionRepo := &mock_repositories.MockStudentSubmissionRepo{}
	studentRepo := &mock_repositories.MockStudentRepo{}
	studyPlanRepo := &mock_repositories.MockStudyPlanRepo{}
	studyPlanItemRepo := &mock_repositories.MockStudyPlanItemRepo{}
	mockDB := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	svc := &StudentAssignmentReaderService{
		DB:                mockDB,
		SubmissionRepo:    submissionRepo,
		StudentRepo:       studentRepo,
		StudyPlanRepo:     studyPlanRepo,
		StudyPlanItemRepo: studyPlanItemRepo,
	}

	validReq := &pb.ListSubmissionsRequest{
		ClassIds: []string{"class-id-1,class-id-2"},
		CourseId: wrapperspb.String("course-id-1"),
		Start:    timestamppb.Now(),
		End:      timestamppb.Now(),
		Statuses: []pb.SubmissionStatus{
			pb.SubmissionStatus_SUBMISSION_STATUS_IN_PROGRESS,
			pb.SubmissionStatus_SUBMISSION_STATUS_MARKED,
		},
		Paging: &cpb.Paging{},
	}

	testCases := []TestCase{
		{
			name:        "error no rows find student by class IDs",
			req:         validReq,
			expectedErr: pgx.ErrNoRows,
			setup: func(ctx context.Context) {
				studentIDs := database.TextArray([]string{"student-id-1", "student-id-2"})
				studentRepo.On("FindStudentsByClassIDs", ctx, mockDB, mock.Anything).Once().Return(
					&studentIDs,
					pgx.ErrNoRows,
				)
			},
		},
		{
			name: "error no rows find student by course",
			req: &pb.ListSubmissionsRequest{
				CourseId: wrapperspb.String("course-id-1"),
				Start:    timestamppb.Now(),
				End:      timestamppb.Now(),
				Statuses: []pb.SubmissionStatus{
					pb.SubmissionStatus_SUBMISSION_STATUS_IN_PROGRESS,
					pb.SubmissionStatus_SUBMISSION_STATUS_MARKED,
				},
				Paging: &cpb.Paging{},
			},
			expectedErr: pgx.ErrNoRows,
			setup: func(ctx context.Context) {
				studentIDs := database.TextArray([]string{"student-id-1", "student-id-2"})
				studentRepo.On("FindStudentsByCourseID", ctx, mockDB, mock.Anything).Once().Return(
					&studentIDs,
					pgx.ErrNoRows,
				)
			},
		},
		{
			name:        "error list submission",
			req:         validReq,
			expectedErr: pgx.ErrNoRows,
			setup: func(ctx context.Context) {
				mockDB.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
				tx.On("Rollback", mock.Anything).Return(nil)
				studentIDs := database.TextArray([]string{"student-id-1", "student-id-2"})
				studentRepo.On("FindStudentsByClassIDs", ctx, mockDB, mock.Anything).Once().Return(
					&studentIDs,
					nil,
				)
				submission := entities.StudentSubmissions{}
				submissionRepo.On("List", ctx, tx, mock.Anything).Once().Return(
					submission, pgx.ErrNoRows,
				)
			},
		},
		{
			name:        "happy case",
			req:         validReq,
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockDB.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
				tx.On("Commit", mock.Anything).Return(nil)
				now := time.Now()
				next10m := time.Now().Add(10 * time.Minute)
				next20m := time.Now().Add(20 * time.Minute)

				studyPlanItem1 := &entities.StudyPlanItem{}
				studyPlanItem1.ID.Set("spii1")
				studyPlanItem1.StudyPlanID.Set("spi1")
				studyPlanItem1.StartDate.Set(now)
				studyPlanItem1.EndDate.Set(next10m)

				studyPlanItem2 := &entities.StudyPlanItem{}
				studyPlanItem2.ID.Set("spii2")
				studyPlanItem2.StudyPlanID.Set("spi1")
				studyPlanItem2.StartDate.Set(next10m)
				studyPlanItem2.EndDate.Set(next10m)

				studyPlanItem3 := &entities.StudyPlanItem{}
				studyPlanItem3.ID.Set("spii3")
				studyPlanItem3.StudyPlanID.Set("spi2")
				studyPlanItem3.StartDate.Set(next10m)
				studyPlanItem3.EndDate.Set(next20m)

				studyPlan1 := &entities.StudyPlan{}
				studyPlan1.ID.Set("spi1")

				studyPlan2 := &entities.StudyPlan{}
				studyPlan2.ID.Set("spi2")

				studentIDs := database.TextArray([]string{"student-id-1", "student-id-2"})
				studentRepo.On("FindStudentsByClassIDs", ctx, mockDB, mock.Anything).Once().Return(
					&studentIDs,
					nil,
				)
				submission := entities.StudentSubmissions{
					{
						AssignmentID:    database.Text("assignment-id"),
						ID:              database.Text("id"),
						StudentID:       database.Text("student-id"),
						StudyPlanItemID: database.Text("spii1"),
					},
					{
						AssignmentID:    database.Text("assignment-id-2"),
						ID:              database.Text("id2"),
						StudentID:       database.Text("student-id"),
						StudyPlanItemID: database.Text("spii2"),
					},
					{
						AssignmentID:    database.Text("assignment-id-3"),
						ID:              database.Text("id3"),
						StudentID:       database.Text("student-id"),
						StudyPlanItemID: database.Text("spii3"),
					},
				}
				submissionRepo.On("List", ctx, tx, mock.Anything).Once().Return(
					submission, nil,
				)

				studyPlanItemRepo.On("FindByIDs",
					ctx,
					tx,
					database.TextArray([]string{
						studyPlanItem1.ID.String,
						studyPlanItem2.ID.String,
						studyPlanItem3.ID.String,
					}),
				).Once().Return([]*entities.StudyPlanItem{studyPlanItem1, studyPlanItem2, studyPlanItem3}, nil)

				studyPlanRepo.On("FindByIDs",
					ctx,
					tx,
					database.TextArray([]string{
						studyPlan1.ID.String,
						studyPlan2.ID.String,
					}),
				).Once().Return([]*entities.StudyPlan{studyPlan1, studyPlan2}, nil)
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		_, err := svc.ListSubmissions(ctx, testCase.req.(*pb.ListSubmissionsRequest))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}

func TestRetrieveSubmissions(t *testing.T) {
	t.Parallel()
	submissionRepo := &mock_repositories.MockStudentSubmissionRepo{}
	studentRepo := &mock_repositories.MockStudentRepo{}
	studyPlanRepo := &mock_repositories.MockStudyPlanRepo{}
	studyPlanItemRepo := &mock_repositories.MockStudyPlanItemRepo{}
	mockDB := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	svc := &StudentAssignmentReaderService{
		DB:                mockDB,
		SubmissionRepo:    submissionRepo,
		StudentRepo:       studentRepo,
		StudyPlanRepo:     studyPlanRepo,
		StudyPlanItemRepo: studyPlanItemRepo,
	}

	studyPlanIDs := []string{"study-plan-item-id", "study-plan-item-id-2"}
	validReq := &pb.RetrieveSubmissionsRequest{
		StudyPlanItemIds: studyPlanIDs,
	}

	testCases := []TestCase{
		{
			name:        "error no rows retrieve submission",
			req:         validReq,
			expectedErr: pgx.ErrNoRows,
			setup: func(ctx context.Context) {
				mockDB.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
				tx.On("Rollback", mock.Anything).Return(nil)
				submissionRepo.On("RetrieveByStudyPlanItemIDs", ctx, tx, database.TextArray(studyPlanIDs)).Once().Return(entities.StudentSubmissions{}, pgx.ErrNoRows)
			},
		},
		{
			name:        "happy case",
			req:         validReq,
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockDB.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
				tx.On("Commit", mock.Anything).Return(nil)
				now := time.Now()
				next10m := time.Now().Add(10 * time.Minute)
				next20m := time.Now().Add(20 * time.Minute)

				studyPlanItem1 := &entities.StudyPlanItem{}
				studyPlanItem1.ID.Set("spii1")
				studyPlanItem1.StudyPlanID.Set("spi1")
				studyPlanItem1.StartDate.Set(now)
				studyPlanItem1.EndDate.Set(next10m)

				studyPlanItem2 := &entities.StudyPlanItem{}
				studyPlanItem2.ID.Set("spii2")
				studyPlanItem2.StudyPlanID.Set("spi1")
				studyPlanItem2.StartDate.Set(next10m)
				studyPlanItem2.EndDate.Set(next10m)

				studyPlanItem3 := &entities.StudyPlanItem{}
				studyPlanItem3.ID.Set("spii3")
				studyPlanItem3.StudyPlanID.Set("spi2")
				studyPlanItem3.StartDate.Set(next10m)
				studyPlanItem3.EndDate.Set(next20m)

				studyPlan1 := &entities.StudyPlan{}
				studyPlan1.ID.Set("spi1")

				studyPlan2 := &entities.StudyPlan{}
				studyPlan2.ID.Set("spi2")

				submissionRepo.On("RetrieveByStudyPlanItemIDs", ctx, tx, database.TextArray(studyPlanIDs)).Once().Return(entities.StudentSubmissions{
					{
						ID:              database.Text("submission-id"),
						StudyPlanItemID: database.Text("spii1"),
					},
					{
						ID:              database.Text("retrieve-id-2"),
						StudyPlanItemID: database.Text("spii2"),
					},
					{
						ID:              database.Text("retrieve-id-3"),
						StudyPlanItemID: database.Text("spii3"),
					},
				}, nil)

				studyPlanItemRepo.On("FindByIDs",
					ctx,
					tx,
					database.TextArray([]string{
						studyPlanItem1.ID.String,
						studyPlanItem2.ID.String,
						studyPlanItem3.ID.String,
					}),
				).Once().Return([]*entities.StudyPlanItem{studyPlanItem1, studyPlanItem2, studyPlanItem3}, nil)

				studyPlanRepo.On("FindByIDs",
					ctx,
					tx,
					database.TextArray([]string{
						studyPlan1.ID.String,
						studyPlan2.ID.String,
					}),
				).Once().Return([]*entities.StudyPlan{studyPlan1, studyPlan2}, nil)
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		_, err := svc.RetrieveSubmissions(ctx, testCase.req.(*pb.RetrieveSubmissionsRequest))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}

func TestStudentAssignmentReaderService_RetrieveSubmissionGrades(t *testing.T) {
	t.Parallel()
	gradeRepo := &mock_repositories.MockStudentSubmissionGradeRepo{}
	mockDB := &mock_database.Ext{}
	svc := &StudentAssignmentReaderService{
		DB:        mockDB,
		GradeRepo: gradeRepo,
	}

	submissionGradeIDs := []string{"submission-grade-id-1", "submission-grade-id-2"}
	validReq := &pb.RetrieveSubmissionGradesRequest{
		SubmissionGradeIds: submissionGradeIDs,
	}

	e := &entities.StudentSubmissionGrade{
		Grade:               database.Numeric(12),
		ID:                  database.Text("id"),
		StudentSubmissionID: database.Text("student-submission-id"),
		GraderComment:       database.Text("grader-comment"),
		GradeContent:        database.JSONB("grade-content"),
	}
	es := entities.StudentSubmissionGrades{e}

	testCases := []TestCase{
		{
			name:        "error no rows retrieve submission grades",
			req:         validReq,
			expectedErr: pgx.ErrNoRows,
			setup: func(ctx context.Context) {
				gradeRepo.On("RetrieveByIDs", ctx, mockDB, database.TextArray(submissionGradeIDs)).Once().
					Return(es, pgx.ErrNoRows)
			},
		},
		{
			name:        "happy case",
			req:         validReq,
			expectedErr: nil,
			setup: func(ctx context.Context) {
				gradeRepo.On("RetrieveByIDs", ctx, mockDB, database.TextArray(submissionGradeIDs)).Once().Return(es, nil)
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		_, err := svc.RetrieveSubmissionGrades(ctx, testCase.req.(*pb.RetrieveSubmissionGradesRequest))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}

func TestListSubmissionsV2(t *testing.T) {
	t.Parallel()
	submissionRepo := &mock_repositories.MockStudentSubmissionRepo{}
	studentRepo := &mock_repositories.MockStudentRepo{}
	studyPlanRepo := &mock_repositories.MockStudyPlanRepo{}
	studyPlanItemRepo := &mock_repositories.MockStudyPlanItemRepo{}
	mockDB := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	svc := &StudentAssignmentReaderService{
		DB:                mockDB,
		SubmissionRepo:    submissionRepo,
		StudentRepo:       studentRepo,
		StudyPlanRepo:     studyPlanRepo,
		StudyPlanItemRepo: studyPlanItemRepo,
	}

	validReq := &pb.ListSubmissionsV2Request{
		ClassIds: []string{"class-id-1,class-id-2"},
		CourseId: wrapperspb.String("course-id-1"),
		Start:    timestamppb.Now(),
		End:      timestamppb.Now(),
		Statuses: []pb.SubmissionStatus{
			pb.SubmissionStatus_SUBMISSION_STATUS_IN_PROGRESS,
			pb.SubmissionStatus_SUBMISSION_STATUS_MARKED,
		},
		Paging: &cpb.Paging{},
	}

	testCases := []TestCase{
		{
			name:        "error list submission",
			req:         validReq,
			expectedErr: status.Errorf(codes.Internal, fmt.Errorf("database.ExecInTx: %w", pgx.ErrNoRows).Error()),
			setup: func(ctx context.Context) {
				mockDB.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
				tx.On("Rollback", mock.Anything).Return(nil)
				studentIDs := database.TextArray([]string{"student-id-1", "student-id-2"})
				studentRepo.On("FindStudentsByClassIDs", ctx, mockDB, mock.Anything).Once().Return(
					&studentIDs,
					nil,
				)
				submission := entities.StudentSubmissions{}
				submissionRepo.On("ListV2", ctx, tx, mock.Anything).Once().Return(
					submission, pgx.ErrNoRows,
				)
			},
		},
		{
			name:        "happy case",
			req:         validReq,
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockDB.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
				tx.On("Commit", mock.Anything).Return(nil)
				now := time.Now()
				next10m := time.Now().Add(10 * time.Minute)
				next20m := time.Now().Add(20 * time.Minute)

				studyPlanItem1 := &entities.StudyPlanItem{}
				studyPlanItem1.ID.Set("spii1")
				studyPlanItem1.StudyPlanID.Set("spi1")
				studyPlanItem1.StartDate.Set(now)
				studyPlanItem1.EndDate.Set(next10m)

				studyPlanItem2 := &entities.StudyPlanItem{}
				studyPlanItem2.ID.Set("spii2")
				studyPlanItem2.StudyPlanID.Set("spi1")
				studyPlanItem2.StartDate.Set(next10m)
				studyPlanItem2.EndDate.Set(next10m)

				studyPlanItem3 := &entities.StudyPlanItem{}
				studyPlanItem3.ID.Set("spii3")
				studyPlanItem3.StudyPlanID.Set("spi2")
				studyPlanItem3.StartDate.Set(next10m)
				studyPlanItem3.EndDate.Set(next20m)

				studyPlan1 := &entities.StudyPlan{}
				studyPlan1.ID.Set("spi1")

				studyPlan2 := &entities.StudyPlan{}
				studyPlan2.ID.Set("spi2")

				studentIDs := database.TextArray([]string{"student-id-1", "student-id-2"})
				studentRepo.On("FindStudentsByClassIDs", ctx, mockDB, mock.Anything).Once().Return(
					&studentIDs,
					nil,
				)
				submission := entities.StudentSubmissions{
					{
						AssignmentID:    database.Text("assignment-id"),
						ID:              database.Text("id"),
						StudentID:       database.Text("student-id"),
						StudyPlanItemID: database.Text("spii1"),
					},
					{
						AssignmentID:    database.Text("assignment-id-2"),
						ID:              database.Text("id2"),
						StudentID:       database.Text("student-id"),
						StudyPlanItemID: database.Text("spii2"),
					},
					{
						AssignmentID:    database.Text("assignment-id-3"),
						ID:              database.Text("id3"),
						StudentID:       database.Text("student-id"),
						StudyPlanItemID: database.Text("spii3"),
					},
				}
				submissionRepo.On("ListV2", ctx, tx, mock.Anything).Once().Return(
					submission, nil,
				)

				studyPlanItemRepo.On("FindByIDs",
					ctx,
					tx,
					database.TextArray([]string{
						studyPlanItem1.ID.String,
						studyPlanItem2.ID.String,
						studyPlanItem3.ID.String,
					}),
				).Once().Return([]*entities.StudyPlanItem{studyPlanItem1, studyPlanItem2, studyPlanItem3}, nil)

				studyPlanRepo.On("FindByIDs",
					ctx,
					tx,
					database.TextArray([]string{
						studyPlan1.ID.String,
						studyPlan2.ID.String,
					}),
				).Once().Return([]*entities.StudyPlan{studyPlan1, studyPlan2}, nil)
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		_, err := svc.ListSubmissionsV2(ctx, testCase.req.(*pb.ListSubmissionsV2Request))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}
