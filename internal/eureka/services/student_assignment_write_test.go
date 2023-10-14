package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	consumer "github.com/manabie-com/backend/internal/notification/transports/nats"
	mock_repositories "github.com/manabie-com/backend/mock/eureka/repositories"
	mock_services "github.com/manabie-com/backend/mock/eureka/services"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_nats "github.com/manabie-com/backend/mock/golibs/nats"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestStudentAssignmentWriteService_SubmitAssignment(t *testing.T) {
	t.Parallel()
	submissionRepo := &mock_repositories.MockStudentSubmissionRepo{}
	latestSubmissionRepo := &mock_repositories.MockStudentLatestSubmissionRepo{}
	studyPlanItemRepo := &mock_repositories.MockStudyPlanItemRepo{}
	assignmentRepo := &mock_repositories.MockAssignmentRepo{}
	studentLearningTimeDailyRepo := &mock_repositories.MockStudentLearningTimeDailyRepo{}
	usermgmtUserReaderService := &mock_services.MockUserMgmtService{}

	mockDB := &mock_database.Ext{}
	tx := &mock_database.Tx{}

	srv := &StudentAssignmentWriteABACService{
		DB:             mockDB,
		AssignmentRepo: assignmentRepo,
		StudentAssignmentWriteService: &StudentAssignmentWriteService{
			DB:                           mockDB,
			SubmissionRepo:               submissionRepo,
			StudentLatestSubmissionRepo:  latestSubmissionRepo,
			StudyPlanItemRepo:            studyPlanItemRepo,
			StudentLearningTimeDailyRepo: studentLearningTimeDailyRepo,
			UsermgmtUserReaderService:    usermgmtUserReaderService,
		},
	}

	now := time.Now()
	studentID := "student-id"
	validReq := &pb.SubmitAssignmentRequest{
		Submission: &pb.StudentSubmission{
			SubmissionId:    "submission-id",
			AssignmentId:    "assignment-id",
			StudyPlanItemId: "study-plan-item-id",
			StudentId:       studentID,
			Note:            "note",
			SubmissionContent: []*pb.SubmissionContent{
				{
					SubmitMediaId:     "submit-media-id",
					AttachmentMediaId: "attachment-media-id",
				},
			},
			CreatedAt: timestamppb.New(now),
			UpdatedAt: timestamppb.New(now),
			Status:    1,
			CourseId:  "course-id",
			StartDate: timestamppb.New(now),
			EndDate:   timestamppb.New(now),
		},
	}

	searchBasicProfileResp := &upb.SearchBasicProfileResponse{
		Profiles: []*cpb.BasicProfile{
			{UserId: studentID, Country: cpb.Country_COUNTRY_VN},
		},
	}

	ctx := interceptors.NewIncomingContext(context.TODO())

	testCases := []TestCase{
		{
			ctx:         interceptors.ContextWithUserID(interceptors.ContextWithUserGroup(ctx, cpb.UserGroup_USER_GROUP_STUDENT.String()), "student-id-1"),
			name:        "happy case student",
			req:         validReq,
			expectedErr: nil,
			setup: func(ctx context.Context) {
				assignmentRepo.On("IsStudentAssigned", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(true, nil)
				mockDB.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
				tx.On("Commit", mock.Anything).Return(nil)
				usermgmtUserReaderService.On("SearchBasicProfile", mock.Anything, mock.Anything).Once().Return(searchBasicProfileResp, nil)
				studentLearningTimeDailyRepo.On("Retrieve", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return([]*entities.StudentLearningTimeDaily{}, nil)
				studentLearningTimeDailyRepo.On("UpsertTaskAssignment", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				submissionRepo.On("Create", ctx, tx, mock.Anything).Once().Return(nil)
				studyPlanItemRepo.On("MarkItemCompleted", ctx, tx, mock.Anything).Return(nil)
				latestSubmissionRepo.On("Upsert", ctx, tx, mock.Anything).Return(nil)
			},
		},
		{
			ctx:         interceptors.ContextWithUserID(interceptors.ContextWithUserGroup(ctx, cpb.UserGroup_USER_GROUP_TEACHER.String()), "teacher-id-1"),
			name:        "happy case teacher",
			req:         validReq,
			expectedErr: nil,
			setup: func(ctx context.Context) {
				assignmentRepo.On("IsStudentAssigned", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(true, nil)
				mockDB.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
				tx.On("Commit", mock.Anything).Return(nil)
				usermgmtUserReaderService.On("SearchBasicProfile", mock.Anything, mock.Anything).Once().Return(searchBasicProfileResp, nil)
				studentLearningTimeDailyRepo.On("Retrieve", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return([]*entities.StudentLearningTimeDaily{}, nil)
				studentLearningTimeDailyRepo.On("UpsertTaskAssignment", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				submissionRepo.On("Create", ctx, tx, mock.Anything).Once().Return(nil)
				studyPlanItemRepo.On("MarkItemCompleted", ctx, tx, mock.Anything).Return(nil)
				latestSubmissionRepo.On("Upsert", ctx, tx, mock.Anything).Return(nil)
			},
		},
		{
			name:        "ErrTxClosed case",
			ctx:         interceptors.ContextWithUserID(interceptors.ContextWithUserGroup(ctx, cpb.UserGroup_USER_GROUP_STUDENT.String()), "student-id-1"),
			req:         validReq,
			expectedErr: fmt.Errorf("database.ExecInTx: %w", pgx.ErrTxClosed),
			setup: func(ctx context.Context) {
				assignmentRepo.On("IsStudentAssigned", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(true, nil)
				mockDB.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
				tx.On("Rollback", mock.Anything).Return(nil)
				usermgmtUserReaderService.On("SearchBasicProfile", mock.Anything, mock.Anything).Once().Return(searchBasicProfileResp, nil)
				studentLearningTimeDailyRepo.On("Retrieve", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return([]*entities.StudentLearningTimeDaily{}, nil)
				studentLearningTimeDailyRepo.On("UpsertTaskAssignment", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				submissionRepo.On("Create", ctx, tx, mock.Anything).Once().Return(pgx.ErrTxClosed)
				studyPlanItemRepo.On("MarkItemCompleted", ctx, tx, mock.Anything).Return(nil)
				latestSubmissionRepo.On("Upsert", ctx, tx, mock.Anything).Return(nil)
			},
		},
		{
			name:        "error call assignmentRepo.IsStudentAssigned",
			req:         validReq,
			ctx:         interceptors.ContextWithUserID(interceptors.ContextWithUserGroup(ctx, cpb.UserGroup_USER_GROUP_STUDENT.String()), "student-id-1"),
			expectedErr: fmt.Errorf("error call assignmentRepo.IsStudentAssigned"),
			setup: func(ctx context.Context) {
				assignmentRepo.On("IsStudentAssigned", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(true, fmt.Errorf("error call assignmentRepo.IsStudentAssigned"))
			},
		},
		{
			name:        "non-assigned assignment",
			req:         validReq,
			ctx:         interceptors.ContextWithUserID(interceptors.ContextWithUserGroup(ctx, cpb.UserGroup_USER_GROUP_STUDENT.String()), "student-id-1"),
			expectedErr: status.Error(codes.PermissionDenied, "non-assigned assignment"),
			setup: func(ctx context.Context) {
				assignmentRepo.On("IsStudentAssigned", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(false, nil)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			_, err := srv.SubmitAssignment(ctx, testCase.req.(*pb.SubmitAssignmentRequest))
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, testCase.expectedErr, err)
			}
		})

	}
}

func TestStudentAssignmentWriteService_GradeStudentSubmission(t *testing.T) {
	t.Parallel()
	submissionRepo := &mock_repositories.MockStudentSubmissionRepo{}
	submissionGradeRepo := &mock_repositories.MockStudentSubmissionGradeRepo{}
	latestSubmissionRepo := &mock_repositories.MockStudentLatestSubmissionRepo{}
	studyPlanItemRepo := &mock_repositories.MockStudyPlanItemRepo{}
	mockDB := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	jsm := new(mock_nats.JetStreamManagement)
	srv := &StudentAssignmentWriteService{
		DB:                          mockDB,
		SubmissionRepo:              submissionRepo,
		SubmissionGradeRepo:         submissionGradeRepo,
		StudentLatestSubmissionRepo: latestSubmissionRepo,
		StudyPlanItemRepo:           studyPlanItemRepo,
	}

	validReq := &pb.GradeStudentSubmissionRequest{
		Grade: &pb.SubmissionGrade{
			SubmissionId: "submission-id",
			Note:         "note",
			Grade:        1,
			GradeContent: []*pb.SubmissionContent{
				{
					SubmitMediaId:     "submit-media-id",
					AttachmentMediaId: "attachment-media-id",
				},
			},
		},
		Status: 0,
	}

	testCases := []TestCase{
		{
			name:        "happy case",
			req:         validReq,
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockDB.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
				tx.On("Commit", mock.Anything).Return(nil)
				submissionGradeRepo.On("Create", ctx, tx, mock.Anything).Once().Return(nil)
				submissionRepo.On("UpdateGradeStatus", ctx, tx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				submissionRepo.On("FindBySubmissionIDs", ctx, tx, mock.Anything).Return(&entities.StudentSubmissions{}, nil)
				latestSubmissionRepo.On("BulkUpserts", ctx, tx, mock.Anything).Return(nil)
			},
		},
		{
			name:        "ErrTxClosed case",
			req:         validReq,
			expectedErr: fmt.Errorf("SubmissionGradeRepo.Create: %w", pgx.ErrTxClosed),
			setup: func(ctx context.Context) {
				mockDB.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
				tx.On("Rollback", mock.Anything).Return(nil)
				submissionGradeRepo.On("Create", ctx, tx, mock.Anything).Once().Return(pgx.ErrTxClosed)
			},
		},
		{
			name: "err SubmissionGradeRepo.RetrieveInfoByIDs",
			req:  validReq,
			setup: func(ctx context.Context) {
				mockDB.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
				tx.On("Commit", mock.Anything).Return(nil)
				submissionGradeRepo.On("Create", ctx, tx, mock.Anything).Once().Return(nil)
				submissionGradeRepo.On("RetrieveInfoByIDs", ctx, tx, mock.Anything).Once().Return(ErrSomethingWentWrong)
			},
		},
		{
			name: "err JSM.PublishContext",
			req:  validReq,
			setup: func(ctx context.Context) {
				mockDB.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
				tx.On("Commit", mock.Anything).Return(nil)
				submissionGradeRepo.On("Create", ctx, tx, mock.Anything).Once().Return(nil)
				submissionGradeRepo.On("RetrieveInfoByIDs", ctx, tx, mock.Anything).Once().Return(nil)
				jsm.On("PublishContext", mock.Anything, consumer.SubjectNotificationCreated, mock.Anything).Once().Return(nil, ErrSomethingWentWrong)

			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		_, err := srv.GradeStudentSubmission(ctx, testCase.req.(*pb.GradeStudentSubmissionRequest))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}

func TestStudentAssignmentWriteService_UpdateStudentSubmissionsStatus(t *testing.T) {
	t.Parallel()
	submissionRepo := &mock_repositories.MockStudentSubmissionRepo{}
	submissionGradeRepo := &mock_repositories.MockStudentSubmissionGradeRepo{}
	latestSubmissionRepo := &mock_repositories.MockStudentLatestSubmissionRepo{}
	studyPlanItemRepo := &mock_repositories.MockStudyPlanItemRepo{}
	mockDB := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	// jsm := new(mock_nats.JetStreamManagement)
	srv := &StudentAssignmentWriteService{
		DB:                          mockDB,
		SubmissionRepo:              submissionRepo,
		SubmissionGradeRepo:         submissionGradeRepo,
		StudentLatestSubmissionRepo: latestSubmissionRepo,
		StudyPlanItemRepo:           studyPlanItemRepo,
	}

	validReq := &pb.UpdateStudentSubmissionsStatusRequest{
		SubmissionIds: []string{"submission-id-1", "submission-id-2"},
		Status:        pb.SubmissionStatus_SUBMISSION_STATUS_IN_PROGRESS,
	}

	testCases := []TestCase{
		{
			name:        "happy case",
			req:         validReq,
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockDB.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
				tx.On("Commit", mock.Anything).Return(nil)

				submissionGradeRepo.On("FindBySubmissionIDs", ctx, tx, mock.Anything).Once().Return(&entities.StudentSubmissionGrades{}, nil)
				submissionGradeRepo.On("BulkImport", ctx, tx, mock.Anything).Once().Return(nil)
				submissionRepo.On("FindBySubmissionIDs", ctx, tx, mock.Anything).Once().Return(&entities.StudentSubmissions{}, nil)
				submissionRepo.On("BulkUpdateStatus", ctx, tx, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				latestSubmissionRepo.On("BulkUpserts", ctx, tx, mock.Anything).Once().Return(nil)
			},
		},
		{
			name:        "err submissionGradeRepo.FindBySubmissionIDs",
			req:         validReq,
			expectedErr: fmt.Errorf("database.ExecInTx: SubmissionGradeRepo.FindBySubmissionIDs: %w", ErrSomethingWentWrong),
			setup: func(ctx context.Context) {
				mockDB.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
				tx.On("Rollback", mock.Anything).Return(nil)
				submissionGradeRepo.On("FindBySubmissionIDs", ctx, tx, mock.Anything).Once().Return(nil, ErrSomethingWentWrong)
			},
		},
		{
			name:        "err submissionGradeRepo.BulkImport",
			req:         validReq,
			expectedErr: fmt.Errorf("database.ExecInTx: SubmissionGradeRepo.BulkImport: %w", ErrSomethingWentWrong),
			setup: func(ctx context.Context) {
				mockDB.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
				tx.On("Rollback", mock.Anything).Return(nil)
				submissionGradeRepo.On("FindBySubmissionIDs", ctx, tx, mock.Anything).Once().Return(&entities.StudentSubmissionGrades{}, nil)
				submissionGradeRepo.On("BulkImport", ctx, tx, mock.Anything).Once().Return(ErrSomethingWentWrong)
			},
		},
		{
			name:        "err sbmissionRepo.BulkUpdateStatus",
			req:         validReq,
			expectedErr: fmt.Errorf("database.ExecInTx: SubmissionRepo.BulkUpdateStatus: %w", ErrSomethingWentWrong),
			setup: func(ctx context.Context) {
				mockDB.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
				tx.On("Rollback", mock.Anything).Return(nil)
				submissionGradeRepo.On("FindBySubmissionIDs", ctx, tx, mock.Anything).Once().Return(&entities.StudentSubmissionGrades{}, nil)
				submissionGradeRepo.On("BulkImport", ctx, tx, mock.Anything).Once().Return(nil)
				submissionRepo.On("BulkUpdateStatus", ctx, tx, mock.Anything, mock.Anything, mock.Anything).Once().Return(ErrSomethingWentWrong)
			},
		},
		{
			name:        "err submissionRepo.FindBySubmissionIDs",
			req:         validReq,
			expectedErr: fmt.Errorf("database.ExecInTx: SubmissionRepo.FindBySubmissionIDs: %w", ErrSomethingWentWrong),
			setup: func(ctx context.Context) {
				mockDB.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
				tx.On("Rollback", mock.Anything).Return(nil)
				submissionGradeRepo.On("FindBySubmissionIDs", ctx, tx, mock.Anything).Once().Return(&entities.StudentSubmissionGrades{}, nil)
				submissionGradeRepo.On("BulkImport", ctx, tx, mock.Anything).Once().Return(nil)
				submissionRepo.On("BulkUpdateStatus", ctx, tx, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				submissionRepo.On("FindBySubmissionIDs", ctx, tx, mock.Anything).Once().Return(nil, ErrSomethingWentWrong)
			},
		},

		{
			name:        "err latestSubmissionRepo.BulkUpserts",
			req:         validReq,
			expectedErr: fmt.Errorf("database.ExecInTx: StudentLatestSubmissionRepo.BulkUpserts: %w", ErrSomethingWentWrong),
			setup: func(ctx context.Context) {
				mockDB.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
				tx.On("Rollback", mock.Anything).Return(nil)
				submissionGradeRepo.On("FindBySubmissionIDs", ctx, tx, mock.Anything).Once().Return(&entities.StudentSubmissionGrades{}, nil)
				submissionGradeRepo.On("BulkImport", ctx, tx, mock.Anything).Once().Return(nil)
				submissionRepo.On("FindBySubmissionIDs", ctx, tx, mock.Anything).Once().Return(&entities.StudentSubmissions{}, nil)
				submissionRepo.On("BulkUpdateStatus", ctx, tx, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				latestSubmissionRepo.On("BulkUpserts", ctx, tx, mock.Anything).Once().Return(ErrSomethingWentWrong)
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		_, err := srv.UpdateStudentSubmissionsStatus(ctx, testCase.req.(*pb.UpdateStudentSubmissionsStatusRequest))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}
