package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/timeutil"
	mock_repositories "github.com/manabie-com/backend/mock/eureka/repositories"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_nats "github.com/manabie-com/backend/mock/golibs/nats"
	pb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/puddle"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type TestCase struct {
	ctx          context.Context
	name         string
	req          interface{}
	expectedResp interface{}
	expectedErr  error
	setup        func(ctx context.Context)
}

func TestScheduleStudyPlan(t *testing.T) {
	loStudyPlanItemRepo := &mock_repositories.MockLoStudyPlanItemRepo{}
	assignmentStudyPlanItemRepo := &mock_repositories.MockAssignmentStudyPlanItemRepo{}
	mockDB := &mock_database.Ext{}
	mockTxer := &mock_database.Tx{}

	AssignmentModifierService := &AssignmentModifierService{
		DB:                          mockDB,
		AssignmentStudyPlanItemRepo: assignmentStudyPlanItemRepo,
		LoStudyPlanItemRepo:         loStudyPlanItemRepo,
	}

	now := time.Now()
	timeutil.Now = func() time.Time {
		return now
	}
	req := &pb.ScheduleStudyPlanRequest{
		Schedule: []*pb.ScheduleStudyPlan{
			{
				StudyPlanItemId: "study-plan-id-1",
				Item: &pb.ScheduleStudyPlan_AssignmentId{
					AssignmentId: "assignment-id",
				},
			},
			{
				StudyPlanItemId: "study-plan-id-2",
				Item: &pb.ScheduleStudyPlan_AssignmentId{
					AssignmentId: "assignment-id-2",
				},
			},
			{
				StudyPlanItemId: "study-plan-id-3",
				Item: &pb.ScheduleStudyPlan_LoId{
					LoId: "lo-id",
				},
			},
			{
				StudyPlanItemId: "study-plan-id-4",
				Item: &pb.ScheduleStudyPlan_LoId{
					LoId: "lo-id-2",
				},
			},
		},
	}
	scheduleAssignment, _ := toAssignmentStudyPlanItems("study-plan-id-1", "assignment-id")
	scheduleAssignment2, _ := toAssignmentStudyPlanItems("study-plan-id-2", "assignment-id-2")

	scheduleLo, _ := toLoStudyPlanItems("study-plan-id-3", "lo-id")
	scheduleLo2, _ := toLoStudyPlanItems("study-plan-id-4", "lo-id-2")

	var nowtime pgtype.Timestamptz
	_ = nowtime.Set(now)
	testCases := []TestCase{
		{
			name:         "err create schedule assignment",
			req:          req,
			expectedResp: nil,
			expectedErr:  status.Error(codes.Unknown, pgx.ErrNoRows.Error()),
			setup: func(ctx context.Context) {
				mockDB.On("Begin", mock.Anything, mock.Anything).Return(mockTxer, nil)
				mockTxer.On("Rollback", mock.Anything).Return(nil)

				assignmentStudyPlanItemRepo.On("BulkInsert", mock.Anything, mock.Anything, []*entities.AssignmentStudyPlanItem{
					scheduleAssignment, scheduleAssignment2,
				}).Once().Return(pgx.ErrTxClosed)
			},
		},
		{
			name:         "err create schedule learning objective",
			req:          req,
			expectedResp: nil,
			expectedErr:  status.Error(codes.Unknown, pgx.ErrNoRows.Error()),
			setup: func(ctx context.Context) {
				mockDB.On("Begin", mock.Anything, mock.Anything).Return(mockTxer, nil)
				mockTxer.On("Rollback", mock.Anything).Return(nil)

				assignmentStudyPlanItemRepo.On("BulkInsert", mock.Anything, mock.Anything, []*entities.AssignmentStudyPlanItem{
					scheduleAssignment, scheduleAssignment2,
				}).Once().Return(nil)
				loStudyPlanItemRepo.On("BulkInsert", mock.Anything, mock.Anything, []*entities.LoStudyPlanItem{
					scheduleLo, scheduleLo2,
				}).Once().Return(pgx.ErrTxClosed)
			},
		},
		{
			name:         "happy case",
			req:          req,
			expectedResp: &pb.ScheduleStudyPlanResponse{},
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				mockDB.On("Begin", mock.Anything, mock.Anything).Return(mockTxer, nil)
				mockTxer.On("Commit", mock.Anything).Return(nil)

				assignmentStudyPlanItemRepo.On("BulkInsert", mock.Anything, mock.Anything, []*entities.AssignmentStudyPlanItem{
					scheduleAssignment, scheduleAssignment2,
				}).Once().Return(nil)
				loStudyPlanItemRepo.On("BulkInsert", mock.Anything, mock.Anything, []*entities.LoStudyPlanItem{
					scheduleLo, scheduleLo2,
				}).Once().Return(nil)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			ctx := context.Background()
			testCase.setup(ctx)
			rsp, err := AssignmentModifierService.ScheduleStudyPlan(ctx, testCase.req.(*pb.ScheduleStudyPlanRequest))
			assert.Equal(t, status.Code(testCase.expectedErr), status.Code(err))
			if testCase.expectedErr != nil {
				assert.Nil(t, rsp, "expecting nil response")
			}
		})
	}
}

func TestUpsertAssignment(t *testing.T) {
}

func TestUpsertAssignmentData(t *testing.T) {
	t.Parallel()
	assignmentRepo := &mock_repositories.MockAssignmentRepo{}
	topicsAssignmentsRepo := &mock_repositories.MockTopicsAssignmentsRepo{}

	mockDB := &mock_database.Ext{}
	mockTxer := &mock_database.Tx{}

	jsm := new(mock_nats.JetStreamManagement)

	AssignmentModifierService := &AssignmentModifierService{
		AssignmentRepo:        assignmentRepo,
		TopicsAssignmentsRepo: topicsAssignmentsRepo,
		DB:                    mockDB,

		JSM: jsm,
	}
	validReq := &pb.UpsertAssignmentsDataRequest{
		Assignments: []*pb.Assignment{
			{
				AssignmentId:     idutil.ULIDNow(),
				Name:             "assignment-name-1",
				AssignmentStatus: pb.AssignmentStatus_ASSIGNMENT_STATUS_ACTIVE,
				Attachments:      []string{"media-id-1", "media-id-2"},
				Content: &pb.AssignmentContent{
					TopicId: "topic-id-1",
					LoId:    []string{"lo-id-1", "lo-id-2"},
				},
			},
			{
				AssignmentId:     idutil.ULIDNow(),
				Name:             "assignment-name-2",
				AssignmentStatus: pb.AssignmentStatus_ASSIGNMENT_STATUS_ACTIVE,
				Attachments:      []string{"media-id-1"},
				Content: &pb.AssignmentContent{
					TopicId: "topic-id-2",
					LoId:    []string{"lo-id-1", "lo-id-2"},
				},
			},
		},
	}
	testCases := []TestCase{
		{
			name:         "err bulk upsert assignment",
			req:          validReq,
			expectedResp: nil,
			expectedErr:  status.Error(codes.Internal, fmt.Errorf("unable to bulk upsert assignment: %w", puddle.ErrNotAvailable).Error()),
			setup: func(ctx context.Context) {
				mockDB.On("Begin", mock.Anything, mock.Anything).Return(mockTxer, nil)
				mockTxer.On("Rollback", mock.Anything).Return(nil)
				assignmentRepo.On("BulkUpsert", mock.Anything, mock.Anything, mock.Anything).Once().Return(puddle.ErrNotAvailable)
			},
		},
		{
			name:        "err bulk upsert topic assignment",
			req:         validReq,
			expectedErr: status.Error(codes.Internal, fmt.Errorf("unable to bulk upsert topic assignment: %w", puddle.ErrNotAvailable).Error()),
			setup: func(ctx context.Context) {
				mockDB.On("Begin", mock.Anything, mock.Anything).Return(mockTxer, nil)
				mockTxer.On("Rollback", mock.Anything).Return(nil)
				assignmentRepo.On("BulkUpsert", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				topicsAssignmentsRepo.On("BulkUpsert", mock.Anything, mock.Anything, mock.Anything).Once().Return(puddle.ErrNotAvailable)
			},
		},
		{
			name:        "err push message to NATS",
			req:         validReq,
			expectedErr: status.Errorf(codes.Internal, fmt.Errorf("s.JSM.PublishContext: subject: %q, something went wrong", constants.SubjectAssignmentsCreated).Error()),
			setup: func(ctx context.Context) {
				mockDB.On("Begin", mock.Anything, mock.Anything).Return(mockTxer, nil)
				mockTxer.On("Commit", mock.Anything).Return(nil)
				assignmentRepo.On("BulkUpsert", mock.Anything, mock.Anything, mock.Anything).Times(2).Return(nil)
				topicsAssignmentsRepo.On("BulkUpsert", mock.Anything, mock.Anything, mock.Anything).Times(2).Return(nil)
				jsm.On("PublishContext", ctx, mock.Anything, mock.Anything).Once().Return(nil, fmt.Errorf("something went wrong"))
			},
		},
		{
			name: "happy case",
			req:  validReq,
			setup: func(ctx context.Context) {
				mockDB.On("Begin", mock.Anything, mock.Anything).Return(mockTxer, nil)
				mockTxer.On("Commit", mock.Anything).Return(nil)
				assignmentRepo.On("BulkUpsert", mock.Anything, mock.Anything, mock.Anything).Times(2).Return(nil)
				topicsAssignmentsRepo.On("BulkUpsert", mock.Anything, mock.Anything, mock.Anything).Times(2).Return(nil)
				jsm.On("PublishContext", ctx, constants.SubjectAssignmentsCreated, mock.Anything).Once().Return(&nats.PubAck{}, nil)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			ctx := context.Background()
			m := map[string]string{
				"pkg":     "com.manabie.liz",
				"version": "1.0.0",
				"token":   "token",
			}
			md := metadata.New(m)
			ctx = metadata.NewIncomingContext(ctx, md)
			testCase.setup(ctx)
			rsp, err := AssignmentModifierService.UpsertAssignmentsData(ctx, testCase.req.(*pb.UpsertAssignmentsDataRequest))
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
				assert.Nil(t, rsp, "expecting nil response")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUpsertStudyPlanItem(t *testing.T) {
	t.Parallel()
	studyPlanItemRepo := &mock_repositories.MockStudyPlanItemRepo{}
	studyPlanRepo := &mock_repositories.MockStudyPlanRepo{}
	mockDB := &mock_database.Ext{}
	mockTxer := &mock_database.Tx{}

	AssignmentModifierService := &AssignmentModifierService{
		DB:                mockDB,
		StudyPlanItemRepo: studyPlanItemRepo,
		StudyPlanRepo:     studyPlanRepo,
	}
	nowTime := timestamppb.Now()

	validReq := &pb.UpsertStudyPlanItemRequest{
		StudyPlanItems: []*pb.StudyPlanItem{
			{
				StudyPlanId:     "study-plan-1",
				StudyPlanItemId: "study-plan-item-id-1",
				AvailableFrom:   nowTime,
				AvailableTo:     nowTime,
				EndDate:         nowTime,
				StartDate:       nowTime,
			},
			{
				StudyPlanId:     "study-plan-1",
				StudyPlanItemId: "study-plan-item-id-2",
				AvailableFrom:   nowTime,
				AvailableTo:     nowTime,
				EndDate:         nowTime,
				StartDate:       nowTime,
			},
			{
				StudyPlanId:     "study-plan-1",
				StudyPlanItemId: "study-plan-item-id-3",
				AvailableFrom:   nowTime,
				AvailableTo:     nowTime,
				EndDate:         nowTime,
				StartDate:       nowTime,
			},
		},
	}
	testCases := []TestCase{
		{
			name:         "err create study plan",
			req:          validReq,
			expectedResp: nil,
			expectedErr:  status.Error(codes.Unknown, puddle.ErrNotAvailable.Error()),
			setup: func(ctx context.Context) {
				mockDB.On("Begin", mock.Anything, mock.Anything).Return(mockTxer, nil)
				mockTxer.On("Rollback", mock.Anything).Return(nil)

				studyPlanRepo.On("FindByIDs", mock.Anything, mockTxer, mock.Anything).Once().Return(nil, nil)
				studyPlanItemRepo.On("BulkInsert", mock.Anything, mock.Anything, mock.Anything).Once().Return(puddle.ErrNotAvailable)
			},
		},
		{
			name:         "err update copied from item",
			req:          validReq,
			expectedResp: nil,
			expectedErr:  status.Error(codes.Unknown, puddle.ErrNotAvailable.Error()),
			setup: func(ctx context.Context) {
				mockDB.On("Begin", mock.Anything, mock.Anything).Return(mockTxer, nil)
				mockTxer.On("Rollback", mock.Anything).Return(nil)

				studyPlanRepo.On("FindByIDs", mock.Anything, mockTxer, mock.Anything).Once().Return(nil, nil)
				studyPlanItemRepo.On("BulkInsert", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				studyPlanItemRepo.On("UpdateWithCopiedFromItem", mock.Anything, mock.Anything, mock.Anything).Once().Return(puddle.ErrNotAvailable)
			},
		},
		{
			name: "success",
			req:  validReq,
			expectedResp: &pb.UpsertStudyPlanItemResponse{
				StudyPlanItemIds: []string{"study-plan-item-id-1", "study-plan-item-id-2", "study-plan-item-id-3"},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockDB.On("Begin", mock.Anything, mock.Anything).Return(mockTxer, nil)
				mockTxer.On("Commit", mock.Anything).Return(nil)

				studyPlanRepo.On("FindByIDs", mock.Anything, mockTxer, mock.Anything).Once().Return(nil, nil)
				studyPlanItemRepo.On("BulkInsert", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				studyPlanItemRepo.On("UpdateWithCopiedFromItem", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				studyPlanRepo.On("BulkUpdateBook", mock.Anything, mockTxer, mock.Anything).Once().Return(nil, nil)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			ctx := context.Background()
			testCase.setup(ctx)
			rsp, err := AssignmentModifierService.UpsertStudyPlanItem(ctx, testCase.req.(*pb.UpsertStudyPlanItemRequest))
			assert.Equal(t, status.Code(testCase.expectedErr), status.Code(err))
			if testCase.expectedErr != nil {
				assert.Nil(t, rsp, "expecting nil response")
			}
		})
	}
}

func TestAssignStudyPlanCourseLevel(t *testing.T) {
	t.Parallel()
	studyPlanItemRepo := &mock_repositories.MockStudyPlanItemRepo{}
	studyPlanRepo := &mock_repositories.MockStudyPlanRepo{}
	studentStudyPlanRepo := &mock_repositories.MockStudentStudyPlanRepo{}
	courseStudyPlanRepo := &mock_repositories.MockCourseStudyPlanRepo{}
	assignmentStudyPlanItemRepo := &mock_repositories.MockAssignmentStudyPlanItemRepo{}
	loStudyPlanItemRepo := &mock_repositories.MockLoStudyPlanItemRepo{}
	studentRepo := &mock_repositories.MockStudentRepo{}

	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	assignmentModifierService := &AssignmentModifierService{
		DB:                          db,
		StudyPlanItemRepo:           studyPlanItemRepo,
		StudyPlanRepo:               studyPlanRepo,
		StudentStudyPlanRepo:        studentStudyPlanRepo,
		CourseStudyPlanRepo:         courseStudyPlanRepo,
		AssignmentStudyPlanItemRepo: assignmentStudyPlanItemRepo,
		LoStudyPlanItemRepo:         loStudyPlanItemRepo,
		StudentRepo:                 studentRepo,
	}

	courseID := database.Text("course-id")
	// nowTime := types.TimestampNow()
	validReq := &pb.AssignStudyPlanRequest{
		StudyPlanId: "study-plan-id",
		Data: &pb.AssignStudyPlanRequest_CourseId{
			CourseId: "course-id",
		},
	}
	oStudyPlanIDs := []string{"study-plan-id", "study-plan-id"}
	createdStudyPlanIDs := []string{"student-plan-id-student-1", "study-plan-id-student-2"}

	testCases := []TestCase{
		{
			name:         "err upsert course",
			req:          validReq,
			expectedResp: nil,
			expectedErr:  fmt.Errorf("UpsertCourseStudyPlan: s.CourseStudyPlanRepo.BulkUpsert: %w", puddle.ErrNotAvailable),
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)
				courseStudyPlanRepo.On("BulkUpsert", mock.Anything, mock.Anything, mock.Anything).Once().Return(puddle.ErrNotAvailable)
			},
		},
		{
			name:         "err find course study plan",
			req:          validReq,
			expectedResp: nil,
			expectedErr:  fmt.Errorf("HandleAssignCourseStudyPlan: s.StudentRepo.FindStudentsByCourseID: %w", pgx.ErrNoRows),
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)
				courseStudyPlanRepo.On("BulkUpsert", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				studentRepo.On("FindStudentsByCourseID", mock.Anything, mock.Anything, courseID).Once().Return(nil, pgx.ErrNoRows)
			},
		},
		{
			name:         "err upsert study plan",
			req:          validReq,
			expectedResp: nil,
			expectedErr:  fmt.Errorf("HandleAssignCourseStudyPlan: s.StudentStudyPlan.BulkUpsert: %w", pgx.ErrNoRows),
			setup: func(ctx context.Context) {
				ids := database.TextArray([]string{
					"student-id-1",
					"student-id-2",
				})
				studentRepo.On("FindStudentsByCourseID", mock.Anything, mock.Anything, courseID).Once().
					Return(&ids, nil)
				db.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)
				studyPlanRepo.On("BulkCopy", mock.Anything, mock.Anything, database.TextArray(oStudyPlanIDs)).Once().Return(oStudyPlanIDs, createdStudyPlanIDs, nil)
				courseStudyPlanRepo.On("BulkUpsert", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				studentStudyPlanRepo.On("BulkUpsert", mock.Anything, mock.Anything, mock.Anything).Once().Return(pgx.ErrNoRows)
			},
		},
		{
			name:         "err copy study plan item",
			req:          validReq,
			expectedResp: nil,
			expectedErr:  fmt.Errorf("HandleAssignCourseStudyPlan: s.StudyPlanItemRepo.BulkCopy: %w", pgx.ErrNoRows),
			setup: func(ctx context.Context) {
				ids := database.TextArray([]string{
					"student-id-1",
					"student-id-2",
				})
				studentRepo.On("FindStudentsByCourseID", mock.Anything, mock.Anything, courseID).Once().
					Return(&ids, nil)
				db.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)
				studyPlanRepo.On("BulkCopy", mock.Anything, mock.Anything, database.TextArray(oStudyPlanIDs)).Once().Return(oStudyPlanIDs, createdStudyPlanIDs, nil)
				courseStudyPlanRepo.On("BulkUpsert", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				studentStudyPlanRepo.On("BulkUpsert", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				studyPlanItemRepo.On("BulkCopy", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(pgx.ErrNoRows)
			},
		},
		{
			name:         "err copy assignment from study plan",
			req:          validReq,
			expectedResp: nil,
			expectedErr:  fmt.Errorf("HandleAssignCourseStudyPlan: s.AssignmentStudyPlanItemRepo.CopyFromStudyPlan: %w", pgx.ErrNoRows),
			setup: func(ctx context.Context) {
				ids := database.TextArray([]string{
					"student-id-1",
					"student-id-2",
				})

				studentRepo.On("FindStudentsByCourseID", mock.Anything, mock.Anything, courseID).Once().
					Return(&ids, nil)
				db.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)
				studyPlanRepo.On("BulkCopy", mock.Anything, mock.Anything, database.TextArray(oStudyPlanIDs)).Once().Return(oStudyPlanIDs, createdStudyPlanIDs, nil)
				courseStudyPlanRepo.On("BulkUpsert", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				studentStudyPlanRepo.On("BulkUpsert", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				studyPlanItemRepo.On("BulkCopy", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				assignmentStudyPlanItemRepo.On("CopyFromStudyPlan", mock.Anything, mock.Anything, mock.Anything).Once().Return(pgx.ErrNoRows)
			},
		},
		{
			name:         "err copy lo from study plan",
			req:          validReq,
			expectedResp: nil,
			expectedErr:  fmt.Errorf("HandleAssignCourseStudyPlan: s.LoStudyPlanItemRepo.CopyFromStudyPlan: %w", pgx.ErrNoRows),
			setup: func(ctx context.Context) {
				ids := database.TextArray([]string{
					"student-id-1",
					"student-id-2",
				})
				studentRepo.On("FindStudentsByCourseID", mock.Anything, mock.Anything, courseID).Once().
					Return(&ids, nil)
				db.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)
				studyPlanRepo.On("BulkCopy", mock.Anything, mock.Anything, database.TextArray(oStudyPlanIDs)).Once().Return(oStudyPlanIDs, createdStudyPlanIDs, nil)
				courseStudyPlanRepo.On("BulkUpsert", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				studentStudyPlanRepo.On("BulkUpsert", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				studyPlanItemRepo.On("BulkCopy", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				assignmentStudyPlanItemRepo.On("CopyFromStudyPlan", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				loStudyPlanItemRepo.On("CopyFromStudyPlan", mock.Anything, mock.Anything, mock.Anything).Once().Return(pgx.ErrNoRows)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			ctx := context.Background()
			testCase.setup(ctx)
			rsp, err := assignmentModifierService.AssignStudyPlan(ctx, testCase.req.(*pb.AssignStudyPlanRequest))
			if testCase.expectedErr != nil {
				assert.Equal(t, fmt.Errorf("AssignStudyPlan: %w", testCase.expectedErr).Error(), err.Error())
			}
			if testCase.expectedResp == nil {
				assert.Nil(t, testCase.expectedResp, rsp)
			} else {
				assert.Equal(t, testCase.expectedResp, rsp)
			}
		})
	}
}

func TestAssignStudyPlanStudentLevel(t *testing.T) {
	t.Parallel()
	studentStudyPlanRepo := &mock_repositories.MockStudentStudyPlanRepo{}
	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}

	AssignmentModifierService := &AssignmentModifierService{
		DB:                   db,
		StudentStudyPlanRepo: studentStudyPlanRepo,
	}

	validReq := &pb.AssignStudyPlanRequest{
		StudyPlanId: "study-plan-id",
		Data: &pb.AssignStudyPlanRequest_StudentId{
			StudentId: "student-id",
		},
	}

	testCases := []TestCase{
		{
			name:         "happy case",
			req:          validReq,
			expectedResp: &pb.AssignStudyPlanResponse{},
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)

				studentStudyPlanRepo.On("BulkUpsert", ctx, tx, mock.Anything).Once().Return(nil)
			},
		},
		{
			name:         "err assign student study plan",
			req:          validReq,
			expectedResp: nil,
			expectedErr:  fmt.Errorf("StudentStudyPlan.BulkUpsert: %w", pgx.ErrNoRows),
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)
				studentStudyPlanRepo.On("BulkUpsert", ctx, tx, mock.Anything).Once().Return(pgx.ErrNoRows)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			ctx := context.Background()
			testCase.setup(ctx)
			rsp, err := AssignmentModifierService.AssignStudyPlan(ctx, testCase.req.(*pb.AssignStudyPlanRequest))
			if testCase.expectedErr != nil {
				expectErr := fmt.Errorf("handleStudentStudyPlan: %w", testCase.expectedErr)
				assert.Equal(t, fmt.Errorf("AssignStudyPlan: %w", expectErr), err)
			}
			if testCase.expectedResp == nil {
				assert.Nil(t, testCase.expectedResp, rsp)
			} else {
				assert.Equal(t, testCase.expectedResp.(*pb.AssignStudyPlanResponse), rsp)
			}
		})
	}
}

func TestSoftDeleteAssignment(t *testing.T) {
	t.Parallel()
	assignmentRepo := &mock_repositories.MockAssignmentRepo{}
	topicsAssignmentsRepo := &mock_repositories.MockTopicsAssignmentsRepo{}
	assignmentStudyPlanItemRepo := &mock_repositories.MockAssignmentStudyPlanItemRepo{}
	studyPlanItemRepo := &mock_repositories.MockStudyPlanItemRepo{}

	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	AssignmentModifierService := &AssignmentModifierService{
		DB:                          db,
		AssignmentRepo:              assignmentRepo,
		TopicsAssignmentsRepo:       topicsAssignmentsRepo,
		AssignmentStudyPlanItemRepo: assignmentStudyPlanItemRepo,
		StudyPlanItemRepo:           studyPlanItemRepo,
	}
	validReq := &pb.DeleteAssignmentsRequest{
		AssignmentIds: []string{"assignment-id-1"},
	}

	testCases := []TestCase{
		{
			name:         "happy case",
			req:          validReq,
			expectedResp: &pb.DeleteAssignmentsResponse{},
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
				assignmentRepo.On("SoftDelete", ctx, tx, mock.Anything).Once().Return(nil)
				topicsAssignmentsRepo.On("SoftDeleteByAssignmentIDs", ctx, tx, mock.Anything).Once().Return(nil)
				assignmentStudyPlanItemRepo.On("SoftDeleteByAssigmentIDs", ctx, tx, mock.Anything).Once().Return(database.TextArray([]string{"assignment_id"}), nil)
				studyPlanItemRepo.On("SoftDeleteByStudyPlanItemIDs", ctx, tx, mock.Anything).Once().Return(nil)
			},
		},
		{
			name:         "err delete assignment",
			req:          validReq,
			expectedResp: nil,
			expectedErr:  status.Errorf(codes.Internal, fmt.Errorf("s.AssignmentRepo.SoftDelete: %w", pgx.ErrNoRows).Error()),
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)
				assignmentRepo.On("SoftDelete", ctx, tx, mock.Anything).Once().Return(pgx.ErrNoRows)
			},
		},
		{
			name:         "err delete topic assignment",
			req:          validReq,
			expectedResp: nil,
			expectedErr:  status.Errorf(codes.Internal, fmt.Errorf("TopicsAssignmentsRepo.SoftDeleteByAssignmentIDs: %w", pgx.ErrNoRows).Error()),
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)
				assignmentRepo.On("SoftDelete", ctx, tx, mock.Anything).Once().Return(nil)
				topicsAssignmentsRepo.On("SoftDeleteByAssignmentIDs", ctx, tx, mock.Anything).Once().Return(pgx.ErrNoRows)
			},
		},
		{
			name:         "err assignment studyplan item",
			req:          validReq,
			expectedResp: nil,
			expectedErr:  status.Errorf(codes.Internal, fmt.Errorf("AssignmentStudyPlanItemRepo.SoftDeleteByAssigmentIDs: %w", pgx.ErrNoRows).Error()),
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)
				assignmentRepo.On("SoftDelete", ctx, tx, mock.Anything).Once().Return(nil)
				topicsAssignmentsRepo.On("SoftDeleteByAssignmentIDs", ctx, tx, mock.Anything).Once().Return(nil)
				assignmentStudyPlanItemRepo.On("SoftDeleteByAssigmentIDs", ctx, tx, mock.Anything).Once().Return(database.TextArray([]string{}), pgx.ErrNoRows)
			},
		},
		{
			name:         "err delete studyplan item",
			req:          validReq,
			expectedResp: nil,
			expectedErr:  status.Errorf(codes.Internal, fmt.Errorf("s.SoftDeleteByStudyPlanItemIDs: %w", pgx.ErrNoRows).Error()),
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)
				assignmentRepo.On("SoftDelete", ctx, tx, mock.Anything).Once().Return(nil)
				topicsAssignmentsRepo.On("SoftDeleteByAssignmentIDs", ctx, tx, mock.Anything).Once().Return(nil)
				assignmentStudyPlanItemRepo.On("SoftDeleteByAssigmentIDs", ctx, tx, mock.Anything).Once().Return(database.TextArray([]string{"assignment_id"}), nil)
				studyPlanItemRepo.On("SoftDeleteByStudyPlanItemIDs", ctx, tx, mock.Anything).Once().Return(pgx.ErrNoRows)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			ctx := context.Background()
			testCase.setup(ctx)
			rsp, err := AssignmentModifierService.DeleteAssignments(ctx, testCase.req.(*pb.DeleteAssignmentsRequest))
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr, err)
			}
			if testCase.expectedResp == nil {
				assert.Nil(t, testCase.expectedResp, rsp)
			} else {
				assert.Equal(t, testCase.expectedResp.(*pb.DeleteAssignmentsResponse), rsp)
			}
		})
	}
}

func TestEditAssignmentTime(t *testing.T) {
	t.Parallel()

	assignmentStudyPlanItemRepo := &mock_repositories.MockAssignmentStudyPlanItemRepo{}
	studyPlanItemRepo := &mock_repositories.MockStudyPlanItemRepo{}

	db := &mock_database.Ext{}
	AssignmentModifierService := &AssignmentModifierService{
		DB:                          db,
		AssignmentStudyPlanItemRepo: assignmentStudyPlanItemRepo,
		StudyPlanItemRepo:           studyPlanItemRepo,
	}

	testCases := []TestCase{
		{
			name: "invalid time",
			req: &pb.EditAssignmentTimeRequest{
				StudentId:        "student-id",
				StudyPlanItemIds: []string{"study-plan-id"},
				UpdateType:       pb.UpdateType_UPDATE_START_DATE,
				StartDate:        timestamppb.New(time.Now()),
			},
			expectedErr: status.Error(codes.InvalidArgument, "invalid time"),
			setup: func(ctx context.Context) {
				studyPlanItemRepo.On("FindByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return([]*entities.StudyPlanItem{
					{
						ID: database.Text("study-plan-id"),
					},
				}, nil)
				assignmentStudyPlanItemRepo.On("BulkEditAssignmentTime", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(fmt.Errorf("cannot update all study plan items"))
			},
		},
		{
			name: "edit study plan item time error",
			req: &pb.EditAssignmentTimeRequest{
				StudentId:        "student-id",
				StudyPlanItemIds: []string{"study-plan-id"},
				UpdateType:       pb.UpdateType_UPDATE_START_DATE,
				StartDate:        timestamppb.New(time.Now()),
			},
			expectedErr: status.Error(codes.Internal, fmt.Errorf("AssignmentModifierService.EditAssignmentTime: %w", fmt.Errorf("error")).Error()),
			setup: func(ctx context.Context) {
				studyPlanItemRepo.On("FindByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return([]*entities.StudyPlanItem{
					{
						ID: database.Text("study-plan-id"),
					},
				}, nil)
				assignmentStudyPlanItemRepo.On("BulkEditAssignmentTime", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(fmt.Errorf("error"))
			},
		},
		{
			name: "fetch study plan item ids error",
			req: &pb.EditAssignmentTimeRequest{
				StudentId:        "student-id",
				StudyPlanItemIds: []string{"study-plan-id"},
				UpdateType:       pb.UpdateType_UPDATE_START_DATE,
				StartDate:        timestamppb.New(time.Now()),
			},
			expectedErr: status.Error(codes.Internal, fmt.Errorf("AssignmentModifierService.EditAssignmentTime: %w", fmt.Errorf("error")).Error()),
			setup: func(ctx context.Context) {
				studyPlanItemRepo.On("FindByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, fmt.Errorf("error"))
				assignmentStudyPlanItemRepo.On("BulkEditAssignmentTime", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "missing study plan item ids",
			req: &pb.EditAssignmentTimeRequest{
				StudentId:  "student-id",
				UpdateType: pb.UpdateType_UPDATE_START_DATE,
				StartDate:  timestamppb.New(time.Now()),
			},
			expectedErr: status.Error(codes.InvalidArgument, fmt.Errorf("AssignmentModifierService.EditAssignmentTime: study plan item ids are empty").Error()),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "start date larger than end date",
			req: &pb.EditAssignmentTimeRequest{
				StudentId:        "student-id",
				StudyPlanItemIds: []string{"study-plan-id"},
				UpdateType:       pb.UpdateType_UPDATE_START_DATE_END_DATE,
				EndDate:          timestamppb.New(time.Now()),
				StartDate:        timestamppb.New(time.Now().Add(+5 * time.Minute)),
			},
			expectedErr: status.Error(codes.InvalidArgument, fmt.Errorf("AssignmentModifierService.EditAssignmentTime: start date after end date").Error()),
			setup: func(ctx context.Context) {
				studyPlanItemRepo.On("FindByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return([]*entities.StudyPlanItem{
					{
						ID: database.Text("study-plan-id"),
					},
				}, nil)
			},
		},
		{
			name: "missing student id",
			req: &pb.EditAssignmentTimeRequest{
				StudyPlanItemIds: []string{"study-plan-id"},
				UpdateType:       pb.UpdateType_UPDATE_START_DATE,
				StartDate:        timestamppb.New(time.Now()),
			},
			expectedErr: status.Error(codes.InvalidArgument, fmt.Errorf("AssignmentModifierService.EditAssignmentTime: student id is empty").Error()),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "happy case",
			req: &pb.EditAssignmentTimeRequest{
				StudentId:        "student-id",
				StudyPlanItemIds: []string{"study-plan-id"},
				UpdateType:       pb.UpdateType_UPDATE_START_DATE,
				StartDate:        timestamppb.New(time.Now()),
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				studyPlanItemRepo.On("FindByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return([]*entities.StudyPlanItem{
					{
						ID: database.Text("study-plan-id"),
					},
				}, nil)
				assignmentStudyPlanItemRepo.On("BulkEditAssignmentTime", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			ctx := context.Background()
			testCase.setup(ctx)
			_, err := AssignmentModifierService.EditAssignmentTime(ctx, testCase.req.(*pb.EditAssignmentTimeRequest))
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr, err)
			}
		})
	}
}

func TestUpsertAssignments(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}

	topicRepo := new(mock_repositories.MockTopicRepo)
	assignmentRepo := new(mock_repositories.MockAssignmentRepo)
	topicAssignmentRepo := new(mock_repositories.MockTopicsAssignmentsRepo)
	jsm := new(mock_nats.JetStreamManagement)

	req := &pb.UpsertAssignmentsRequest{
		Assignments: []*pb.Assignment{
			{
				AssignmentId:     idutil.ULIDNow(),
				Name:             "assignment-name-1",
				AssignmentStatus: pb.AssignmentStatus_ASSIGNMENT_STATUS_ACTIVE,
				Attachments:      []string{"media-id-1", "media-id-2"},
				Content: &pb.AssignmentContent{
					TopicId: "topic-id-1",
					LoId:    []string{"lo-id-1", "lo-id-2"},
				},
			},
			{
				AssignmentId:     idutil.ULIDNow(),
				Name:             "assignment-name-2",
				AssignmentStatus: pb.AssignmentStatus_ASSIGNMENT_STATUS_ACTIVE,
				Attachments:      []string{"media-id-1"},
				Content: &pb.AssignmentContent{
					TopicId: "topic-id-2",
					LoId:    []string{"lo-id-1", "lo-id-2"},
				},
			},
		},
	}
	topics := []*entities.Topic{
		{
			ID:                    database.Text("topic-id-1"),
			LODisplayOrderCounter: database.Int4(0),
		},
		{
			ID:                    database.Text("topic-id-2"),
			LODisplayOrderCounter: database.Int4(0),
		},
	}

	testCases := []TestCase{
		{
			name:        "happy case",
			ctx:         interceptors.ContextWithUserID(ctx, "admin"),
			req:         req,
			expectedErr: nil,
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything, mock.Anything).Times(2).Return(tx, nil)
				tx.On("Commit", mock.Anything).Times(2).Return(nil)
				topicRepo.On("UpdateTotalLOs", ctx, tx, mock.Anything).Times(2).Return(nil)
				topicRepo.On("UpdateLODisplayOrderCounter", ctx, tx, mock.Anything, mock.Anything).Times(2).Return(nil)
				topicRepo.On("RetrieveByIDs", ctx, mock.Anything, mock.Anything).Once().Return(topics, nil)
				topicRepo.On("RetrieveByID", ctx, tx, mock.Anything, mock.Anything).Times(2).Return(topics[0], nil)
				assignmentRepo.On("BulkUpsert", ctx, tx, mock.Anything).Times(2).Return(nil)
				assignmentRepo.On("RetrieveAssignments", ctx, db, mock.Anything).Once().Return(nil, nil)
				topicAssignmentRepo.On("BulkUpsert", ctx, tx, mock.Anything).Times(2).Return(nil)
				jsm.On("PublishContext", ctx, constants.SubjectAssignmentsCreated, mock.Anything).Once().Return(&nats.PubAck{}, nil)
			},
		},
		{
			name:        "error TopicRepo.RetrieveByIDs",
			ctx:         interceptors.ContextWithUserID(ctx, "admin"),
			req:         req,
			expectedErr: status.Errorf(codes.InvalidArgument, fmt.Errorf("unable to retrieve topics by ids: %w", pgx.ErrNoRows).Error()),
			setup: func(ctx context.Context) {
				topicRepo.On("RetrieveByIDs", ctx, mock.Anything, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
			},
		},
		{
			name:        "error TopicRepo.RetrieveByID",
			ctx:         interceptors.ContextWithUserID(ctx, "admin"),
			req:         req,
			expectedErr: status.Error(codes.Internal, fmt.Errorf("unable to retrieve topic by id: %w", pgx.ErrNoRows).Error()),
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything, mock.Anything).Once().Return(tx, nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)

				topicRepo.On("RetrieveByIDs", ctx, mock.Anything, mock.Anything).Once().Return(topics, nil)
				assignmentRepo.On("RetrieveAssignments", ctx, db, mock.Anything).Once().Return(nil, nil)
				topicRepo.On("RetrieveByID", ctx, tx, mock.Anything, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
			},
		},
		{
			name:        "error TopicRepo.isTopicsExisted",
			ctx:         interceptors.ContextWithUserID(ctx, "admin"),
			req:         req,
			expectedErr: status.Error(codes.InvalidArgument, "some topics does not exists"),
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything, mock.Anything).Once().Return(tx, nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)

				topicRepo.On("RetrieveByIDs", ctx, mock.Anything, mock.Anything).Once().Return([]*entities.Topic{
					{
						ID:                    database.Text("topic-id-3"),
						LODisplayOrderCounter: database.Int4(0),
					},
				}, nil)
			},
		},
		{
			name:        "error cm.UpdateLODisplayOrderCounter",
			ctx:         interceptors.ContextWithUserID(ctx, "admin"),
			req:         req,
			expectedErr: status.Error(codes.Internal, fmt.Errorf("unable to update lo display order counter: %w", pgx.ErrTxCommitRollback).Error()),
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything, mock.Anything).Once().Return(tx, nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)

				topicRepo.On("UpdateLODisplayOrderCounter", ctx, tx, mock.Anything, mock.Anything).Once().Return(pgx.ErrTxCommitRollback)
				topicRepo.On("RetrieveByIDs", ctx, mock.Anything, mock.Anything).Once().Return(topics, nil)
				topicRepo.On("RetrieveByID", ctx, tx, mock.Anything, mock.Anything).Once().Return(topics[0], nil)
				assignmentRepo.On("BulkUpsert", ctx, tx, mock.Anything).Once().Return(nil)
				assignmentRepo.On("RetrieveAssignments", ctx, db, mock.Anything).Once().Return(nil, nil)
				topicAssignmentRepo.On("BulkUpsert", ctx, tx, mock.Anything).Once().Return(nil)
			},
		},
		{
			name:        "error cm.updateTotalLOs",
			ctx:         interceptors.ContextWithUserID(ctx, "admin"),
			req:         req,
			expectedErr: status.Error(codes.Internal, fmt.Errorf("unable to update total learing objectives: %w", pgx.ErrTxCommitRollback).Error()),
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything, mock.Anything).Once().Return(tx, nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)
				topicRepo.On("UpdateLODisplayOrderCounter", ctx, tx, mock.Anything, mock.Anything).Once().Return(nil)
				topicRepo.On("RetrieveByIDs", ctx, mock.Anything, mock.Anything).Once().Return(topics, nil)
				topicRepo.On("RetrieveByID", ctx, tx, mock.Anything, mock.Anything).Once().Return(topics[0], nil)
				topicRepo.On("UpdateTotalLOs", ctx, tx, mock.Anything).Once().Return(pgx.ErrTxCommitRollback)
				assignmentRepo.On("BulkUpsert", ctx, tx, mock.Anything).Once().Return(nil)
				topicAssignmentRepo.On("BulkUpsert", ctx, tx, mock.Anything).Once().Return(nil)
				assignmentRepo.On("RetrieveAssignments", ctx, db, mock.Anything).Once().Return(nil, nil)

			},
		},
	}

	s := &AssignmentModifierService{
		DB:                    db,
		TopicRepo:             topicRepo,
		AssignmentRepo:        assignmentRepo,
		TopicsAssignmentsRepo: topicAssignmentRepo,
		JSM:                   jsm,
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			ctx := interceptors.NewIncomingContext(testCase.ctx)
			testCase.setup(ctx)
			req := testCase.req.(*pb.UpsertAssignmentsRequest)
			_, err := s.UpsertAssignments(ctx, req)
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, testCase.expectedErr, err)
			}
		})
	}
}

func TestAssignmentModifierService_AssignAssignmentsToTopic(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	db := &mock_database.Ext{}

	topicRepo := new(mock_repositories.MockTopicRepo)
	assignmentRepo := new(mock_repositories.MockAssignmentRepo)
	topicAssignmentRepo := new(mock_repositories.MockTopicsAssignmentsRepo)
	jsm := new(mock_nats.JetStreamManagement)

	req := &pb.AssignAssignmentsToTopicRequest{
		TopicId: "topic-id",
		Assignment: []*pb.AssignAssignmentsToTopicRequest_Assignment{
			{
				AssignmentId: "assignment-id-1",
				DisplayOrder: 0,
			},
			{
				AssignmentId: "assignment-id-2",
				DisplayOrder: 1,
			},
		},
	}

	testCases := []TestCase{
		{
			name:        "happy case",
			ctx:         interceptors.ContextWithUserID(ctx, "admin"),
			req:         req,
			expectedErr: nil,
			setup: func(ctx context.Context) {
				topicAssignmentRepo.On("BulkUpsert", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				assignmentRepo.On("RetrieveAssignments", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, nil)
				jsm.On("PublishContext", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, nil)
			},
		},
		{
			name:        "error s.TopicsAssignmentsRepo.Upsert ",
			ctx:         interceptors.ContextWithUserID(ctx, "admin"),
			req:         req,
			expectedErr: fmt.Errorf("s.TopicsAssignmentsRepo.Upsert: %w", pgx.ErrNoRows),
			setup: func(ctx context.Context) {
				topicAssignmentRepo.On("BulkUpsert", mock.Anything, mock.Anything, mock.Anything).Once().Return(pgx.ErrNoRows)
			},
		},
		{
			name:        "error s.AssignmentRepo.RetrieveAssignments ",
			ctx:         interceptors.ContextWithUserID(ctx, "admin"),
			req:         req,
			expectedErr: fmt.Errorf("s.AssignmentRepo.RetrieveAssignments: %w", pgx.ErrNoRows),
			setup: func(ctx context.Context) {
				topicAssignmentRepo.On("BulkUpsert", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				assignmentRepo.On("RetrieveAssignments", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
			},
		},
	}

	s := &AssignmentModifierService{
		DB:                    db,
		TopicRepo:             topicRepo,
		AssignmentRepo:        assignmentRepo,
		TopicsAssignmentsRepo: topicAssignmentRepo,
		JSM:                   jsm,
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			ctx := interceptors.NewIncomingContext(testCase.ctx)
			testCase.setup(ctx)
			req := testCase.req.(*pb.AssignAssignmentsToTopicRequest)
			_, err := s.AssignAssignmentsToTopic(ctx, req)
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, testCase.expectedErr, err)
			}
		})
	}
}

func TestInsertAssignment(t *testing.T) {
	t.Parallel()
	topicRepo := &mock_repositories.MockTopicRepo{}
	generalAssignmentRepo := &mock_repositories.MockGeneralAssignmentRepo{}

	mockDB := &mock_database.Ext{}
	mockTx := &mock_database.Tx{}

	assignmentService := &AssignmentService{
		DB:                    mockDB,
		TopicRepo:             topicRepo,
		GeneralAssignmentRepo: generalAssignmentRepo,
	}
	validReq := &sspb.InsertAssignmentRequest{
		Assignment: &sspb.AssignmentBase{
			Base: &sspb.LearningMaterialBase{
				TopicId: "topic-id",
				Name:    "name",
			},
			Attachments:            []string{"attachment-1", "attachment-2"},
			Instruction:            "instruction",
			MaxGrade:               10,
			IsRequiredGrade:        true,
			AllowResubmission:      true,
			RequireAttachment:      false,
			AllowLateSubmission:    false,
			RequireAssignmentNote:  false,
			RequireVideoSubmission: true,
		},
	}
	testCases := []TestCase{
		{
			name:        "err retrieve topic",
			req:         validReq,
			expectedErr: status.Errorf(codes.InvalidArgument, "topic topic-id doesn't exists"),
			setup: func(ctx context.Context) {
				topicRepo.On("RetrieveByID", mock.Anything, mockDB, mock.Anything, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
			},
		},
		{
			name: "happy case",
			req:  validReq,
			setup: func(ctx context.Context) {
				topic := &entities.Topic{
					ID:                    database.Text("topic-id"),
					LODisplayOrderCounter: database.Int4(0),
				}

				mockDB.On("Begin", mock.Anything, mock.Anything).Return(mockTx, nil)
				mockTx.On("Commit", mock.Anything).Return(nil)

				topicRepo.On("RetrieveByID", mock.Anything, mockDB, mock.Anything, mock.Anything).Once().Return(topic, nil)
				topicRepo.On("RetrieveByID", mock.Anything, mockTx, mock.Anything, mock.Anything).Once().Return(topic, nil)
				generalAssignmentRepo.On("Insert", mock.Anything, mockTx, mock.Anything).Once().Return(nil)
				topicRepo.On("UpdateLODisplayOrderCounter", mock.Anything, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				topicRepo.On("UpdateTotalLOs", mock.Anything, mockTx, mock.Anything).Once().Return(nil)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			ctx := interceptors.NewIncomingContext(context.Background())
			testCase.setup(ctx)
			rsp, err := assignmentService.InsertAssignment(ctx, testCase.req.(*sspb.InsertAssignmentRequest))
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
				assert.Nil(t, rsp, "expecting nil response")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUpdateAssignment(t *testing.T) {
	t.Parallel()
	topicRepo := &mock_repositories.MockTopicRepo{}
	generalAssignmentRepo := &mock_repositories.MockGeneralAssignmentRepo{}

	mockDB := &mock_database.Ext{}

	assignmentService := &AssignmentService{
		DB:                    mockDB,
		TopicRepo:             topicRepo,
		GeneralAssignmentRepo: generalAssignmentRepo,
	}
	validReq := &sspb.UpdateAssignmentRequest{
		Assignment: &sspb.AssignmentBase{
			Base: &sspb.LearningMaterialBase{
				LearningMaterialId: "learning-material-id",
				Name:               "name",
			},
			Attachments:            []string{"attachment-1", "attachment-2"},
			Instruction:            "instruction",
			MaxGrade:               10,
			IsRequiredGrade:        true,
			AllowResubmission:      true,
			RequireAttachment:      false,
			AllowLateSubmission:    false,
			RequireAssignmentNote:  false,
			RequireVideoSubmission: true,
		},
	}
	testCases := []TestCase{
		{
			name: "err validate updateAssignmentRequest",
			req: &sspb.UpdateAssignmentRequest{
				Assignment: &sspb.AssignmentBase{
					Base: &sspb.LearningMaterialBase{},
				},
			},
			expectedErr: status.Errorf(codes.InvalidArgument, fmt.Errorf("validateUpdateAssignmentReq: %w", status.Error(codes.InvalidArgument, "empty learning_material_id")).Error()),
			setup:       func(ctx context.Context) {},
		},
		{
			name: "happy case",
			req:  validReq,
			setup: func(ctx context.Context) {
				generalAssignmentRepo.On("Update", mock.Anything, mockDB, mock.Anything).Once().Return(nil)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			ctx := interceptors.NewIncomingContext(context.Background())
			testCase.setup(ctx)
			rsp, err := assignmentService.UpdateAssignment(ctx, testCase.req.(*sspb.UpdateAssignmentRequest))
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
				assert.Nil(t, rsp, "expecting nil response")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestToAssignmentConfig(t *testing.T) {
	testCases := []TestCase{
		{
			name: "happy case",
			req: &pb.AssignmentSetting{
				AllowLateSubmission:       true,
				AllowResubmission:         true,
				RequireAssignmentNote:     true,
				RequireAttachment:         true,
				RequireVideoSubmission:    true,
				RequireCompleteDate:       true,
				RequireDuration:           true,
				RequireCorrectness:        false,
				RequireUnderstandingLevel: true,
			},
			expectedResp: &entities.AssignmentSetting{
				AllowLateSubmission:       true,
				AllowResubmission:         true,
				RequireAssignmentNote:     true,
				RequireAttachment:         true,
				RequireVideoSubmission:    true,
				RequireCompleteDate:       true,
				RequireDuration:           true,
				RequireCorrectness:        false,
				RequireUnderstandingLevel: true,
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			res := toAssignmentConfig(testCase.req.(*pb.AssignmentSetting))
			assert.Equal(t, testCase.expectedResp, res)
		})
	}
}

func TestToAssignmentToDoList(t *testing.T) {
	m := make(map[string]bool, 0)
	m["content"] = true
	testCases := []TestCase{
		{
			name: "happy case",
			req: &pb.CheckList{
				Items: []*pb.CheckListItem{
					{
						Content:   "content",
						IsChecked: true,
					},
				},
			},
			expectedResp: &entities.AssignmentCheckList{
				CheckList: m,
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			res := toAssignmentToDoList(testCase.req.(*pb.CheckList))
			assert.Equal(t, testCase.expectedResp, res)
		})
	}
}

func Test_validateInsertAssignmentReq(t *testing.T) {
	assignmentService := &AssignmentService{}
	testCases := []TestCase{
		{
			name: "LearningMaterialId not empty",
			req: &sspb.InsertAssignmentRequest{
				Assignment: &sspb.AssignmentBase{
					Base: &sspb.LearningMaterialBase{
						LearningMaterialId: "learning_material_id",
					},
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, "learning_material_id must be empty"),
		},
		{
			name: "type not empty",
			req: &sspb.InsertAssignmentRequest{
				Assignment: &sspb.AssignmentBase{
					Base: &sspb.LearningMaterialBase{
						Type: sspb.LearningMaterialType_LEARNING_MATERIAL_GENERAL_ASSIGNMENT.String(),
					},
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, "type must be empty"),
		},
		{
			name: "name is empty",
			req: &sspb.InsertAssignmentRequest{
				Assignment: &sspb.AssignmentBase{
					Base: &sspb.LearningMaterialBase{},
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, "empty assignment name"),
		},
		{
			name: "topic is empty",
			req: &sspb.InsertAssignmentRequest{
				Assignment: &sspb.AssignmentBase{
					Base: &sspb.LearningMaterialBase{
						Name: "lm name",
					},
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, "empty topic_id"),
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			err := assignmentService.validateInsertAssignmentReq(testCase.req.(*sspb.InsertAssignmentRequest))
			assert.Equal(t, testCase.expectedErr, err)
		})
	}
}

func Test_validateUpdateAssignmentReq(t *testing.T) {
	assignmentService := &AssignmentService{}
	testCases := []TestCase{
		{
			name: "LearningMaterialId empty",
			req: &sspb.UpdateAssignmentRequest{
				Assignment: &sspb.AssignmentBase{
					Base: &sspb.LearningMaterialBase{},
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, "empty learning_material_id"),
		},
		{
			name: "type not empty",
			req: &sspb.UpdateAssignmentRequest{
				Assignment: &sspb.AssignmentBase{
					Base: &sspb.LearningMaterialBase{
						LearningMaterialId: "lm-id",
						Type:               sspb.LearningMaterialType_LEARNING_MATERIAL_GENERAL_ASSIGNMENT.String(),
					},
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, "type must be empty"),
		},
		{
			name: "topic is empty",
			req: &sspb.UpdateAssignmentRequest{
				Assignment: &sspb.AssignmentBase{
					Base: &sspb.LearningMaterialBase{
						LearningMaterialId: "lm-id",
						TopicId:            "topic-id",
					},
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, "topic_id must be empty"),
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			err := assignmentService.validateUpdateAssignmentReq(testCase.req.(*sspb.UpdateAssignmentRequest))
			assert.Equal(t, testCase.expectedErr, err)
		})
	}
}
