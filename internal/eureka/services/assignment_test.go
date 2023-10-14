package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/invoicemgmt/constant"
	mock_repositories "github.com/manabie-com/backend/mock/eureka/repositories"
	mock_services "github.com/manabie-com/backend/mock/eureka/services"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestListAssignment(t *testing.T) {
	ctx := context.Background()
	mockDB := &mock_database.Ext{}
	mockGeneralAssignmentRepo := &mock_repositories.MockGeneralAssignmentRepo{}

	var (
		LearningMaterialIds = database.TextArray([]string{"id1", "id2"})
	)

	assignmentService := &AssignmentService{
		DB:                    mockDB,
		GeneralAssignmentRepo: mockGeneralAssignmentRepo,
	}

	testCases := []TestCase{
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				generalAssignments := []*entities.GeneralAssignment{
					{
						LearningMaterial: entities.LearningMaterial{
							ID:           database.Text("id1"),
							TopicID:      database.Text("topic-id"),
							Name:         database.Text("sid"),
							Type:         database.Text(sspb.LearningMaterialType_LEARNING_MATERIAL_GENERAL_ASSIGNMENT.String()),
							DisplayOrder: database.Int2(1),
						},
						Attachments:            database.TextArray([]string{"attachment-1", "attachment-2"}),
						Instruction:            database.Text("instruction"),
						MaxGrade:               database.Int4(10),
						IsRequiredGrade:        database.Bool(true),
						AllowResubmission:      database.Bool(true),
						RequireAttachment:      database.Bool(false),
						AllowLateSubmission:    database.Bool(false),
						RequireAssignmentNote:  database.Bool(false),
						RequireVideoSubmission: database.Bool(true),
					},

					{
						LearningMaterial: entities.LearningMaterial{
							ID:           database.Text("id2"),
							TopicID:      database.Text("topic-id"),
							Name:         database.Text("sid"),
							Type:         database.Text(sspb.LearningMaterialType_LEARNING_MATERIAL_GENERAL_ASSIGNMENT.String()),
							DisplayOrder: database.Int2(2),
						},
						Attachments:            database.TextArray([]string{"attachment-1", "attachment-2"}),
						Instruction:            database.Text("instruction"),
						MaxGrade:               database.Int4(10),
						IsRequiredGrade:        database.Bool(true),
						AllowResubmission:      database.Bool(true),
						RequireAttachment:      database.Bool(false),
						AllowLateSubmission:    database.Bool(false),
						RequireAssignmentNote:  database.Bool(false),
						RequireVideoSubmission: database.Bool(true),
					},
				}
				mockGeneralAssignmentRepo.On("List", mock.Anything, mock.Anything, LearningMaterialIds).Once().Return(generalAssignments, nil)
			},
			req: &sspb.ListAssignmentRequest{
				LearningMaterialIds: []string{"id1", "id2"},
			},
			expectedErr: nil,
			expectedResp: &sspb.ListAssignmentResponse{
				Assignments: []*sspb.AssignmentBase{
					{
						Base: &sspb.LearningMaterialBase{
							LearningMaterialId: "id1",
							TopicId:            "topic-id",
							Name:               "sid",
							Type:               sspb.LearningMaterialType_LEARNING_MATERIAL_GENERAL_ASSIGNMENT.String(),
							DisplayOrder: &wrapperspb.Int32Value{
								Value: 1,
							},
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
					{
						Base: &sspb.LearningMaterialBase{
							LearningMaterialId: "id2",
							TopicId:            "topic-id",
							Name:               "sid",
							Type:               sspb.LearningMaterialType_LEARNING_MATERIAL_GENERAL_ASSIGNMENT.String(),
							DisplayOrder: &wrapperspb.Int32Value{
								Value: 2,
							},
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
				},
			},
		},

		{
			name: "Not found",
			setup: func(ctx context.Context) {
				mockGeneralAssignmentRepo.On("List", mock.Anything, mock.Anything, LearningMaterialIds).Once().Return(nil, pgx.ErrNoRows)
			},
			expectedErr: status.Errorf(codes.NotFound, "assignment not found: %v", pgx.ErrNoRows),
			req: &sspb.ListAssignmentRequest{
				LearningMaterialIds: []string{"id1", "id2"},
			},
		},
		{
			name: "invalid request",
			setup: func(ctx context.Context) {
				mockGeneralAssignmentRepo.On("List", mock.Anything, mock.Anything, LearningMaterialIds).Once().Return(nil, pgx.ErrNoRows)
			},
			expectedErr: status.Errorf(codes.InvalidArgument, "LearningMaterialIds must not be empty"),
			req: &sspb.ListAssignmentRequest{
				LearningMaterialIds: []string{},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			resp, err := assignmentService.ListAssignment(ctx, testCase.req.(*sspb.ListAssignmentRequest))
			assert.Equal(t, testCase.expectedErr, err)
			if testCase.expectedResp != nil {
				assert.IsType(t, testCase.expectedResp, resp)
				assert.Equal(t, testCase.expectedResp, resp)
			}
		})
	}
}

func TestAssignmentService_SubmitAssignment(t *testing.T) {
	t.Parallel()
	submissionRepo := &mock_repositories.MockStudentSubmissionRepo{}
	latestSubmissionRepo := &mock_repositories.MockStudentLatestSubmissionRepo{}
	individualPlanRepo := &mock_repositories.MockIndividualStudyPlan{}
	assignmentRepo := &mock_repositories.MockAssignmentRepo{}
	studentLearningTimeDailyRepo := &mock_repositories.MockStudentLearningTimeDailyRepo{}
	usermgmtUserReaderService := &mock_services.MockUserMgmtService{}

	mockDB := &mock_database.Ext{}
	tx := &mock_database.Tx{}

	srv := &AssignmentService{
		DB:                           mockDB,
		AssignmentRepo:               assignmentRepo,
		SubmissionRepo:               submissionRepo,
		StudentLatestSubmissionRepo:  latestSubmissionRepo,
		StudentLearningTimeDailyRepo: studentLearningTimeDailyRepo,
		UsermgmtUserReaderService:    usermgmtUserReaderService,
	}

	now := time.Now()
	studentID := "student-id"
	validReq := &sspb.SubmitAssignmentRequest{
		Submission: &sspb.StudentSubmission{
			SubmissionId: "submission-id",
			StudyPlanItemIdentity: &sspb.StudyPlanItemIdentity{
				LearningMaterialId: "learning-material-id",
				StudyPlanId:        "study-plan-id",
				StudentId:          wrapperspb.String(studentID),
			},
			Note: "note",
			SubmissionContent: []*sspb.SubmissionContent{
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
				assignmentRepo.On("IsStudentAssignedV2", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(true, nil)
				mockDB.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
				tx.On("Commit", mock.Anything).Return(nil)
				usermgmtUserReaderService.On("SearchBasicProfile", mock.Anything, mock.Anything).Once().Return(searchBasicProfileResp, nil)
				studentLearningTimeDailyRepo.On("Retrieve", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Once().Return([]*entities.StudentLearningTimeDaily{}, nil)
				studentLearningTimeDailyRepo.On("UpsertTaskAssignment", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				submissionRepo.On("Create", ctx, tx, mock.Anything).Once().Return(nil)
				individualPlanRepo.On("MarkCompleted", ctx, tx, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				latestSubmissionRepo.On("UpsertV2", ctx, tx, mock.Anything).Return(nil)
			},
		},
		{
			ctx:         interceptors.ContextWithUserID(interceptors.ContextWithUserGroup(ctx, cpb.UserGroup_USER_GROUP_TEACHER.String()), "teacher-id-1"),
			name:        "happy case teacher",
			req:         validReq,
			expectedErr: nil,
			setup: func(ctx context.Context) {
				assignmentRepo.On("IsStudentAssignedV2", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(true, nil)
				mockDB.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
				tx.On("Commit", mock.Anything).Return(nil)
				usermgmtUserReaderService.On("SearchBasicProfile", mock.Anything, mock.Anything).Once().Return(searchBasicProfileResp, nil)
				studentLearningTimeDailyRepo.On("Retrieve", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Once().Return([]*entities.StudentLearningTimeDaily{}, nil)
				studentLearningTimeDailyRepo.On("UpsertTaskAssignment", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				submissionRepo.On("Create", ctx, tx, mock.Anything).Once().Return(nil)
				individualPlanRepo.On("MarkCompleted", ctx, tx, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				latestSubmissionRepo.On("UpsertV2", ctx, tx, mock.Anything).Return(nil)
			},
		},
		{
			name:        "ErrTxClosed case",
			ctx:         interceptors.ContextWithUserID(interceptors.ContextWithUserGroup(ctx, cpb.UserGroup_USER_GROUP_STUDENT.String()), "student-id-1"),
			req:         validReq,
			expectedErr: status.Error(codes.Internal, fmt.Errorf("database.ExecInTx: s.SubmissionRepo.Create: %w", pgx.ErrTxClosed).Error()),
			setup: func(ctx context.Context) {
				assignmentRepo.On("IsStudentAssignedV2", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(true, nil)
				mockDB.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
				tx.On("Rollback", mock.Anything).Return(nil)
				usermgmtUserReaderService.On("SearchBasicProfile", mock.Anything, mock.Anything).Once().Return(searchBasicProfileResp, nil)
				studentLearningTimeDailyRepo.On("Retrieve", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Once().Return([]*entities.StudentLearningTimeDaily{}, nil)
				studentLearningTimeDailyRepo.On("UpsertTaskAssignment", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				submissionRepo.On("Create", ctx, tx, mock.Anything).Once().Return(pgx.ErrTxClosed)
				individualPlanRepo.On("MarkCompleted", ctx, tx, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				latestSubmissionRepo.On("UpsertV2", ctx, tx, mock.Anything).Return(nil)
			},
		},
		{
			name:        "error call assignmentRepo.IsStudentAssignedV2",
			req:         validReq,
			ctx:         interceptors.ContextWithUserID(interceptors.ContextWithUserGroup(ctx, cpb.UserGroup_USER_GROUP_STUDENT.String()), "student-id-1"),
			expectedErr: status.Error(codes.PermissionDenied, fmt.Errorf("validateSubmitAssignmentPermission: error call assignmentRepo.IsStudentAssignedV2").Error()),
			setup: func(ctx context.Context) {
				assignmentRepo.
					On("IsStudentAssignedV2", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Once().Return(true, fmt.Errorf("error call assignmentRepo.IsStudentAssignedV2"))
			},
		},
		{
			name:        "non-assigned assignment",
			req:         validReq,
			ctx:         interceptors.ContextWithUserID(interceptors.ContextWithUserGroup(ctx, cpb.UserGroup_USER_GROUP_STUDENT.String()), "student-id-1"),
			expectedErr: status.Error(codes.PermissionDenied, "validateSubmitAssignmentPermission: non-assigned assignment"),
			setup: func(ctx context.Context) {
				assignmentRepo.On("IsStudentAssignedV2", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(false, nil)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			_, err := srv.SubmitAssignment(ctx, testCase.req.(*sspb.SubmitAssignmentRequest))
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, testCase.expectedErr, err)
			}
		})

	}
}

func TestAssignmentService_validateSubmitAssignmentRequest(t *testing.T) {
	s := &AssignmentService{}
	testCases := []TestCase{
		{
			name: "empty StudyPlanItemIdentity",
			req: &sspb.SubmitAssignmentRequest{
				Submission: &sspb.StudentSubmission{
					StudyPlanItemIdentity: nil,
				},
			},
			expectedErr: fmt.Errorf("studyPlanItemIdentity can not be empty"),
		},
		{
			name: "empty StudyPlanId",
			req: &sspb.SubmitAssignmentRequest{
				Submission: &sspb.StudentSubmission{
					StudyPlanItemIdentity: &sspb.StudyPlanItemIdentity{
						StudyPlanId:        "",
						LearningMaterialId: "learning-material-id",
						StudentId:          wrapperspb.String("student-id"),
					},
				},
			},
			expectedErr: fmt.Errorf("studyPlanId can not be empty"),
		},
		{
			name: "empty LearningMaterialId",
			req: &sspb.SubmitAssignmentRequest{
				Submission: &sspb.StudentSubmission{
					StudyPlanItemIdentity: &sspb.StudyPlanItemIdentity{
						StudyPlanId:        "study-plan-id",
						LearningMaterialId: "",
						StudentId:          wrapperspb.String("student-id"),
					},
				},
			},
			expectedErr: fmt.Errorf("learningMaterialId can not be empty"),
		},
		{
			name: "empty StudentId",
			req: &sspb.SubmitAssignmentRequest{
				Submission: &sspb.StudentSubmission{
					StudyPlanItemIdentity: &sspb.StudyPlanItemIdentity{
						StudyPlanId:        "study-plan-id",
						LearningMaterialId: "learning-material-id",
						StudentId:          wrapperspb.String(""),
					},
				},
			},
			expectedErr: fmt.Errorf("studentId can not be empty"),
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			err := s.validateSubmitAssignmentRequest(testCase.req.(*sspb.SubmitAssignmentRequest))
			assert.Equal(t, err, testCase.expectedErr)
		})
	}
}

func TestAssignmentService_submitAssignmentRequestToEnt(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	s := &AssignmentService{}
	t.Run("default for complete date", func(t *testing.T) {
		ctx := interceptors.ContextWithUserGroup(ctx, constant.RoleStudent)
		req := &sspb.SubmitAssignmentRequest{Submission: &sspb.StudentSubmission{
			SubmissionId: "submission-id",
			StudyPlanItemIdentity: &sspb.StudyPlanItemIdentity{
				StudyPlanId:        "study-plan-id",
				LearningMaterialId: "learning-material-id",
				StudentId:          wrapperspb.String("student-id"),
			},
		}}
		res, err := s.submitAssignmentRequestToEnt(ctx, req)
		assert.Equal(t, nil, err)
		assert.NotNil(t, res.CompleteDate.Time)
	})
}
