package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/eureka/constants"
	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	mock_repositories "github.com/manabie-com/backend/mock/eureka/repositories"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"
	"github.com/pkg/errors"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestExamLOService_InsertExamLO(t *testing.T) {
	ctx := context.Background()
	mockDB := &mock_database.Ext{}
	mockTx := &mock_database.Tx{}

	mockTopicRepo := &mock_repositories.MockTopicRepo{}
	mockExamLORepo := &mock_repositories.MockExamLORepo{}

	examLOService := &ExamLOService{
		DB:         mockDB,
		TopicRepo:  mockTopicRepo,
		ExamLORepo: mockExamLORepo,
	}
	req := &sspb.InsertExamLORequest{
		ExamLo: &sspb.ExamLOBase{
			Base: &sspb.LearningMaterialBase{
				TopicId: "topic-id-1",
				Name:    "exam-lo-1",
				Type:    sspb.LearningMaterialType_LEARNING_MATERIAL_EXAM_LO.String(),
			},
			Instruction:    "instruction",
			GradeToPass:    wrapperspb.Int32(1),
			ManualGrading:  true,
			TimeLimit:      wrapperspb.Int32(1),
			MaximumAttempt: wrapperspb.Int32(10),
		},
	}
	testCases := []TestCase{
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				mockDB.On("Begin", mock.Anything, mock.Anything).Return(mockTx, nil)
				mockTopicRepo.On("RetrieveByID", mock.Anything, mock.Anything, database.Text("topic-id-1"), mock.Anything).Once().Return(&entities.Topic{
					ID:                    database.Text("topic-id-1"),
					LODisplayOrderCounter: database.Int4(0),
				}, nil)
				mockExamLORepo.On("Insert", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				mockTopicRepo.On("UpdateLODisplayOrderCounter", mock.Anything, mock.Anything, database.Text("topic-id-1"), database.Int4(1)).Once().Return(nil)
				mockTx.On("Commit", mock.Anything).Return(nil)
			},
			req:          req,
			expectedResp: &sspb.InsertExamLOResponse{},
		},
		{
			name: "some field be ignored",
			setup: func(ctx context.Context) {
				mockDB.On("Begin", mock.Anything, mock.Anything).Return(mockTx, nil)
				mockTopicRepo.On("RetrieveByID", mock.Anything, mock.Anything, database.Text("topic-id-1"), mock.Anything).Once().Return(&entities.Topic{
					ID:                    database.Text("topic-id-1"),
					LODisplayOrderCounter: database.Int4(0),
				}, nil)
				mockExamLORepo.On("Insert", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				mockTopicRepo.On("UpdateLODisplayOrderCounter", mock.Anything, mock.Anything, database.Text("topic-id-1"), database.Int4(1)).Once().Return(nil)
				mockTx.On("Commit", mock.Anything).Return(nil)
			},
			req: &sspb.InsertExamLORequest{
				ExamLo: &sspb.ExamLOBase{
					Base: &sspb.LearningMaterialBase{
						TopicId: "topic-id-1",
						Name:    "exam-lo-1",
						Type:    sspb.LearningMaterialType_LEARNING_MATERIAL_EXAM_LO.String(),
					},
				},
			},
			expectedResp: &sspb.InsertExamLOResponse{},
		},
		{
			name: "topic not found",
			setup: func(ctx context.Context) {
				mockDB.On("Begin", mock.Anything, mock.Anything).Return(mockTx, nil)
				mockTopicRepo.On("RetrieveByID", mock.Anything, mock.Anything, database.Text("topic-id-1"), mock.Anything).Once().Return(nil, pgx.ErrNoRows)
				mockTx.On("Rollback", mock.Anything).Return(nil)
			},
			req:         req,
			expectedErr: status.Errorf(codes.Internal, fmt.Errorf("database.ExecInTx: s.TopicRepo.RetrieveByID: %w", pgx.ErrNoRows).Error()),
		},
		{
			name:  "validate error maximum_attempt < 1",
			setup: func(ctx context.Context) {},
			req: &sspb.InsertExamLORequest{
				ExamLo: &sspb.ExamLOBase{
					Base: &sspb.LearningMaterialBase{
						TopicId: "topic-id-1",
						Name:    "exam-lo-1",
						Type:    sspb.LearningMaterialType_LEARNING_MATERIAL_EXAM_LO.String(),
					},
					Instruction:    "instruction",
					GradeToPass:    wrapperspb.Int32(1),
					ManualGrading:  true,
					TimeLimit:      wrapperspb.Int32(1),
					MaximumAttempt: wrapperspb.Int32(0),
				},
			},
			expectedErr: status.Errorf(codes.InvalidArgument, "validateExamLOReq: maximum_attempt must be null or between 1 to 99"),
		},
		{
			name:  "validate error maximum_attempt > 99",
			setup: func(ctx context.Context) {},
			req: &sspb.InsertExamLORequest{
				ExamLo: &sspb.ExamLOBase{
					Base: &sspb.LearningMaterialBase{
						TopicId: "topic-id-1",
						Name:    "exam-lo-1",
						Type:    sspb.LearningMaterialType_LEARNING_MATERIAL_EXAM_LO.String(),
					},
					Instruction:    "instruction",
					GradeToPass:    wrapperspb.Int32(1),
					ManualGrading:  true,
					TimeLimit:      wrapperspb.Int32(1),
					MaximumAttempt: wrapperspb.Int32(100),
				},
			},
			expectedErr: status.Errorf(codes.InvalidArgument, "validateExamLOReq: maximum_attempt must be null or between 1 to 99"),
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			resp, err := examLOService.InsertExamLO(ctx, testCase.req.(*sspb.InsertExamLORequest))
			assert.Equal(t, testCase.expectedErr, err)
			if testCase.expectedResp != nil {
				assert.IsType(t, testCase.expectedResp, resp)
			}
		})
	}
}

func Test_validateInsertExamLOReq(t *testing.T) {
	testCases := []TestCase{
		{
			name: "LearningMaterialId not empty",
			req: &sspb.ExamLOBase{
				Base: &sspb.LearningMaterialBase{
					LearningMaterialId: "learning_material_id",
				},
			},
			expectedErr: fmt.Errorf("LearningMaterialId must be empty"),
		},
		{
			name: "Empty TopicId",
			req: &sspb.ExamLOBase{
				Base: &sspb.LearningMaterialBase{
					TopicId: "",
				},
			},
			expectedErr: fmt.Errorf("Topic ID must not be empty"),
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			err := validateInsertExamLOReq(testCase.req.(*sspb.ExamLOBase))
			assert.Equal(t, testCase.expectedErr, err)
		})
	}
}

func TestExamLOService_UpdateExamLO(t *testing.T) {
	ctx := context.Background()
	mockDB := &mock_database.Ext{}
	mockTx := &mock_database.Tx{}

	mockExamLORepo := &mock_repositories.MockExamLORepo{}

	s := &ExamLOService{
		DB:         mockDB,
		ExamLORepo: mockExamLORepo,
	}

	testCases := []TestCase{
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				mockDB.On("Begin", mock.Anything, mock.Anything).Return(mockTx, nil)
				mockExamLORepo.On("Update", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", mock.Anything).Return(nil)
			},
			req: &sspb.UpdateExamLORequest{
				ExamLo: &sspb.ExamLOBase{
					Base: &sspb.LearningMaterialBase{
						LearningMaterialId: "exam-lo-id-1",
						Name:               "ExamLO updated",
					},
					MaximumAttempt: wrapperspb.Int32(10),
				},
			},
			expectedResp: &sspb.UpdateExamLOResponse{},
		},
		{
			name:  "missing LearningMaterialID",
			setup: func(ctx context.Context) {},
			req: &sspb.UpdateExamLORequest{
				ExamLo: &sspb.ExamLOBase{
					Base: &sspb.LearningMaterialBase{
						LearningMaterialId: "",
						Name:               "ExamLO updated",
					},
				},
			},
			expectedErr: status.Errorf(codes.InvalidArgument, fmt.Errorf("validateUpdateExamLOReq: LearningMaterialId must not be empty").Error()),
		},
		{
			name: "exam-lo not found",
			setup: func(ctx context.Context) {
				mockDB.On("Begin", mock.Anything, mock.Anything).Return(mockTx, nil)
				mockExamLORepo.On("Update", mock.Anything, mock.Anything, mock.Anything).Once().Return(pgx.ErrNoRows)
				mockTx.On("Rollback", mock.Anything).Return(nil)
			},
			req: &sspb.UpdateExamLORequest{
				ExamLo: &sspb.ExamLOBase{
					Base: &sspb.LearningMaterialBase{
						LearningMaterialId: "exam-lo-id-2",
						Name:               "ExamLO updated",
					},
				},
			},
			expectedErr: status.Errorf(codes.Internal, fmt.Errorf("database.ExecInTx: s.ExamLORepo.Update: %w", pgx.ErrNoRows).Error()),
		},
		{
			name:  "validate error maximum_attempt < 1",
			setup: func(ctx context.Context) {},
			req: &sspb.UpdateExamLORequest{
				ExamLo: &sspb.ExamLOBase{
					Base: &sspb.LearningMaterialBase{
						LearningMaterialId: "exam-lo-id-1",
						Name:               "ExamLO updated",
					},
					MaximumAttempt: wrapperspb.Int32(0),
				},
			},
			expectedErr: status.Errorf(codes.InvalidArgument, "validateUpdateExamLOReq: maximum_attempt must be null or between 1 to 99"),
		},
		{
			name:  "validate error maximum_attempt > 99",
			setup: func(ctx context.Context) {},
			req: &sspb.UpdateExamLORequest{
				ExamLo: &sspb.ExamLOBase{
					Base: &sspb.LearningMaterialBase{
						LearningMaterialId: "exam-lo-id-1",
						Name:               "ExamLO updated",
					},
					MaximumAttempt: wrapperspb.Int32(100),
				},
			},
			expectedErr: status.Errorf(codes.InvalidArgument, "validateUpdateExamLOReq: maximum_attempt must be null or between 1 to 99"),
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			resp, err := s.UpdateExamLO(ctx, testCase.req.(*sspb.UpdateExamLORequest))
			assert.Equal(t, testCase.expectedErr, err)
			if testCase.expectedResp != nil {
				assert.IsType(t, testCase.expectedResp, resp)
			}
		})
	}
}

func TestExamLOService_ListExamLO(t *testing.T) {
	ctx := context.Background()
	mockDB := &mock_database.Ext{}
	mockTx := &mock_database.Tx{}

	mockExamLORepo := &mock_repositories.MockExamLORepo{}

	examLOService := &ExamLOService{
		DB:         mockDB,
		ExamLORepo: mockExamLORepo,
	}

	testCases := []TestCase{
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				mockDB.On("Begin", mock.Anything, mock.Anything).Return(mockTx, nil)
				mockExamLORepo.On("ListExamLOBaseByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return([]*entities.ExamLOBase{
					{
						ExamLO: entities.ExamLO{
							LearningMaterial: entities.LearningMaterial{
								ID:           database.Text("exam-lo-id-1"),
								TopicID:      database.Text("topic-id-1"),
								Name:         database.Text("exam-lo-1"),
								Type:         database.Text("exam-lo"),
								DisplayOrder: database.Int2(1),
							},
							Instruction:   database.Text("instruction-1"),
							GradeToPass:   database.Int4(8),
							ManualGrading: database.Bool(true),
							TimeLimit:     database.Int4(1),
						},
						TotalQuestion: database.Int4(1),
					},
					{
						ExamLO: entities.ExamLO{
							LearningMaterial: entities.LearningMaterial{
								ID:           database.Text("exam-lo-id-2"),
								TopicID:      database.Text("topic-id-1"),
								Name:         database.Text("exam-lo-2"),
								Type:         database.Text("exam-lo"),
								DisplayOrder: database.Int2(2),
							},
							Instruction:   database.Text("instruction-2"),
							GradeToPass:   database.Int4(8),
							ManualGrading: database.Bool(true),
							TimeLimit:     database.Int4(1),
						},
						TotalQuestion: database.Int4(2),
					},
					{
						ExamLO: entities.ExamLO{
							LearningMaterial: entities.LearningMaterial{
								ID:           database.Text("exam-lo-id-3"),
								TopicID:      database.Text("topic-id-1"),
								Name:         database.Text("exam-lo-3"),
								Type:         database.Text("exam-lo"),
								DisplayOrder: database.Int2(3),
							},
							Instruction:   database.Text("instruction-3"),
							ManualGrading: database.Bool(true),
						},
						TotalQuestion: database.Int4(3),
					},
				}, nil)
				mockTx.On("Commit", mock.Anything).Return(nil)
			},
			req: &sspb.ListExamLORequest{
				LearningMaterialIds: []string{"exam-lo-id-1", "exam-lo-id-2"},
			},
			expectedResp: &sspb.ListExamLOResponse{
				ExamLos: []*sspb.ExamLOBase{
					{
						Base: &sspb.LearningMaterialBase{
							LearningMaterialId: "exam-lo-id-1",
							TopicId:            "topic-id-1",
							Name:               "exam-lo-1",
							Type:               "exam-lo",
							DisplayOrder:       wrapperspb.Int32(1),
						},
						Instruction:   "instruction-1",
						GradeToPass:   wrapperspb.Int32(8),
						ManualGrading: true,
						TimeLimit:     wrapperspb.Int32(1),
						TotalQuestion: int32(1),
					},
					{
						Base: &sspb.LearningMaterialBase{
							LearningMaterialId: "exam-lo-id-2",
							TopicId:            "topic-id-1",
							Name:               "exam-lo-2",
							Type:               "exam-lo",
							DisplayOrder:       wrapperspb.Int32(2),
						},
						Instruction:   "instruction-2",
						GradeToPass:   wrapperspb.Int32(8),
						ManualGrading: true,
						TimeLimit:     wrapperspb.Int32(1),
						TotalQuestion: int32(2),
					},
					{
						Base: &sspb.LearningMaterialBase{
							LearningMaterialId: "exam-lo-id-3",
							TopicId:            "topic-id-1",
							Name:               "exam-lo-3",
							Type:               "exam-lo",
							DisplayOrder:       wrapperspb.Int32(3),
						},
						Instruction:   "instruction-3",
						GradeToPass:   &wrapperspb.Int32Value{},
						ManualGrading: true,
						TimeLimit:     &wrapperspb.Int32Value{},
						TotalQuestion: int32(3),
					},
				},
			},
		},
		{
			name: "not found",
			setup: func(ctx context.Context) {
				mockDB.On("Begin", mock.Anything, mock.Anything).Return(mockTx, nil)
				mockExamLORepo.On("ListExamLOBaseByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
				mockTx.On("Commit", mock.Anything).Return(nil)
			},
			req: &sspb.ListExamLORequest{
				LearningMaterialIds: []string{"exam-lo-id-1", "exam-lo-id-2"},
			},
			expectedErr: status.Errorf(codes.NotFound, fmt.Errorf("s.ExamLORepo.ListExamLOBaseByIDs: %w", pgx.ErrNoRows).Error()),
		},
		{
			name: "missing learning material ids",
			setup: func(ctx context.Context) {
				mockDB.On("Begin", mock.Anything, mock.Anything).Return(mockTx, nil)
				mockExamLORepo.On("ListExamLOBaseByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
				mockTx.On("Commit", mock.Anything).Return(nil)
			},
			req: &sspb.ListExamLORequest{
				LearningMaterialIds: []string{},
			},
			expectedErr: status.Errorf(codes.InvalidArgument, "LearningMaterialIds must not be empty"),
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			resp, err := examLOService.ListExamLO(ctx, testCase.req.(*sspb.ListExamLORequest))
			assert.Equal(t, testCase.expectedErr, err)
			if testCase.expectedResp != nil {
				assert.IsType(t, testCase.expectedResp, resp)
				assert.Equal(t, testCase.expectedResp, resp)
			}
		})
	}
}

func TestExamLOService_ListHighestResultExamLOSubmission(t *testing.T) {
	ctx := context.Background()
	mockDB := &mock_database.Ext{}

	mockExamLORepo := &mock_repositories.MockExamLORepo{}
	mockExamLOSubmissionRepo := &mock_repositories.MockExamLOSubmissionRepo{}

	examLOService := &ExamLOService{
		DB:                   mockDB,
		ExamLORepo:           mockExamLORepo,
		ExamLOSubmissionRepo: mockExamLOSubmissionRepo,
	}

	testCases := []TestCase{
		{
			name: "submissions: contains completed, passed, failed",
			req: &sspb.ListHighestResultExamLOSubmissionRequest{
				StudyPlanItemIdentities: []*sspb.StudyPlanItemIdentity{
					{
						StudyPlanId:        "study-plan-id-1",
						StudentId:          wrapperspb.String("student-id-1"),
						LearningMaterialId: "learning-material-id-1",
					},
				},
			},
			expectedResp: &sspb.ListHighestResultExamLOSubmissionResponse{
				StudyPlanItemResults: []*sspb.ListHighestResultExamLOSubmissionResponse_StudyPlanItemResult{
					{
						StudyPlanItemIdentity: &sspb.StudyPlanItemIdentity{
							StudyPlanId:        "study-plan-id-1",
							StudentId:          wrapperspb.String("student-id-1"),
							LearningMaterialId: "learning-material-id-1",
						},
						LatestExamLoSubmissionResult: sspb.ExamLOSubmissionResult_EXAM_LO_SUBMISSION_PASSED,
					},
				},
			},
			setup: func(ctx context.Context) {
				mockExamLOSubmissionRepo.On("ListByStudyPlanItemIdentities", mock.Anything, mock.Anything, mock.Anything).Once().Return([]*entities.ExamLOSubmission{
					{
						LearningMaterialID: database.Text("learning-material-id-1"),
						StudentID:          database.Text("student-id-1"),
						StudyPlanID:        database.Text("study-plan-id-1"),
						Result:             database.Text(sspb.ExamLOSubmissionResult_EXAM_LO_SUBMISSION_COMPLETED.String()),
					},
					{
						LearningMaterialID: database.Text("learning-material-id-1"),
						StudentID:          database.Text("student-id-1"),
						StudyPlanID:        database.Text("study-plan-id-1"),
						Result:             database.Text(sspb.ExamLOSubmissionResult_EXAM_LO_SUBMISSION_PASSED.String()),
					},
					{
						LearningMaterialID: database.Text("learning-material-id-1"),
						StudentID:          database.Text("student-id-1"),
						StudyPlanID:        database.Text("study-plan-id-1"),
						Result:             database.Text(sspb.ExamLOSubmissionResult_EXAM_LO_SUBMISSION_FAILED.String()),
					},
				}, nil)
			},
		},
		{
			name: "submissions: contains passed, failed",
			req: &sspb.ListHighestResultExamLOSubmissionRequest{
				StudyPlanItemIdentities: []*sspb.StudyPlanItemIdentity{
					{
						StudyPlanId:        "study-plan-id-1",
						StudentId:          wrapperspb.String("student-id-1"),
						LearningMaterialId: "learning-material-id-1",
					},
				},
			},
			expectedResp: &sspb.ListHighestResultExamLOSubmissionResponse{
				StudyPlanItemResults: []*sspb.ListHighestResultExamLOSubmissionResponse_StudyPlanItemResult{
					{
						StudyPlanItemIdentity: &sspb.StudyPlanItemIdentity{
							StudyPlanId:        "study-plan-id-1",
							StudentId:          wrapperspb.String("student-id-1"),
							LearningMaterialId: "learning-material-id-1",
						},
						LatestExamLoSubmissionResult: sspb.ExamLOSubmissionResult_EXAM_LO_SUBMISSION_PASSED,
					},
				},
			},
			setup: func(ctx context.Context) {
				mockExamLOSubmissionRepo.On("ListByStudyPlanItemIdentities", mock.Anything, mock.Anything, mock.Anything).Once().Return([]*entities.ExamLOSubmission{
					{
						LearningMaterialID: database.Text("learning-material-id-1"),
						StudentID:          database.Text("student-id-1"),
						StudyPlanID:        database.Text("study-plan-id-1"),
						Result:             database.Text(sspb.ExamLOSubmissionResult_EXAM_LO_SUBMISSION_PASSED.String()),
					},
					{
						LearningMaterialID: database.Text("learning-material-id-1"),
						StudentID:          database.Text("student-id-1"),
						StudyPlanID:        database.Text("study-plan-id-1"),
						Result:             database.Text(sspb.ExamLOSubmissionResult_EXAM_LO_SUBMISSION_FAILED.String()),
					},
				}, nil)
			},
		},
		{
			name: "submissions: contains failed, none",
			req: &sspb.ListHighestResultExamLOSubmissionRequest{
				StudyPlanItemIdentities: []*sspb.StudyPlanItemIdentity{
					{
						StudyPlanId:        "study-plan-id-1",
						StudentId:          wrapperspb.String("student-id-1"),
						LearningMaterialId: "learning-material-id-1",
					},
				},
			},
			expectedResp: &sspb.ListHighestResultExamLOSubmissionResponse{
				StudyPlanItemResults: []*sspb.ListHighestResultExamLOSubmissionResponse_StudyPlanItemResult{
					{
						StudyPlanItemIdentity: &sspb.StudyPlanItemIdentity{
							StudyPlanId:        "study-plan-id-1",
							StudentId:          wrapperspb.String("student-id-1"),
							LearningMaterialId: "learning-material-id-1",
						},
						LatestExamLoSubmissionResult: sspb.ExamLOSubmissionResult_EXAM_LO_SUBMISSION_FAILED,
					},
				},
			},
			setup: func(ctx context.Context) {
				mockExamLOSubmissionRepo.On("ListByStudyPlanItemIdentities", mock.Anything, mock.Anything, mock.Anything).Once().Return([]*entities.ExamLOSubmission{}, nil)
				mockExamLORepo.On("ListByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return([]*entities.ExamLO{}, nil)
			},
		},
		{
			name: "submissions: contains passed, failed",
			req: &sspb.ListHighestResultExamLOSubmissionRequest{
				StudyPlanItemIdentities: []*sspb.StudyPlanItemIdentity{
					{
						StudyPlanId:        "study-plan-id-1",
						StudentId:          wrapperspb.String("student-id-1"),
						LearningMaterialId: "learning-material-id-1",
					},
				},
			},
			expectedResp: &sspb.ListHighestResultExamLOSubmissionResponse{
				StudyPlanItemResults: []*sspb.ListHighestResultExamLOSubmissionResponse_StudyPlanItemResult{
					{
						StudyPlanItemIdentity: &sspb.StudyPlanItemIdentity{
							StudyPlanId:        "study-plan-id-1",
							StudentId:          wrapperspb.String("student-id-1"),
							LearningMaterialId: "learning-material-id-1",
						},
						LatestExamLoSubmissionResult: sspb.ExamLOSubmissionResult_EXAM_LO_SUBMISSION_PASSED,
					},
				},
			},
			setup: func(ctx context.Context) {
				mockExamLOSubmissionRepo.On("ListByStudyPlanItemIdentities", mock.Anything, mock.Anything, mock.Anything).Once().Return([]*entities.ExamLOSubmission{
					{
						LearningMaterialID: database.Text("learning-material-id-1"),
						StudentID:          database.Text("student-id-1"),
						StudyPlanID:        database.Text("study-plan-id-1"),
						Result:             database.Text(sspb.ExamLOSubmissionResult_EXAM_LO_SUBMISSION_PASSED.String()),
					},
					{
						LearningMaterialID: database.Text("learning-material-id-1"),
						StudentID:          database.Text("student-id-1"),
						StudyPlanID:        database.Text("study-plan-id-1"),
						Result:             database.Text(sspb.ExamLOSubmissionResult_EXAM_LO_SUBMISSION_FAILED.String()),
					},
				}, nil)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			resp, err := examLOService.ListHighestResultExamLOSubmission(ctx, testCase.req.(*sspb.ListHighestResultExamLOSubmissionRequest))
			assert.Equal(t, testCase.expectedErr, err)
			if testCase.expectedResp != nil {
				assert.IsType(t, testCase.expectedResp, resp)

				expectedResult := testCase.expectedResp.(*sspb.ListHighestResultExamLOSubmissionResponse)
				for i, result := range expectedResult.StudyPlanItemResults {
					assert.Equal(t, result.StudyPlanItemIdentity.LearningMaterialId, expectedResult.StudyPlanItemResults[i].StudyPlanItemIdentity.LearningMaterialId)
					assert.Equal(t, result.StudyPlanItemIdentity.StudentId, expectedResult.StudyPlanItemResults[i].StudyPlanItemIdentity.StudentId)
					assert.Equal(t, result.StudyPlanItemIdentity.StudyPlanId, expectedResult.StudyPlanItemResults[i].StudyPlanItemIdentity.StudyPlanId)
					assert.Equal(t, result.LatestExamLoSubmissionResult, expectedResult.StudyPlanItemResults[i].LatestExamLoSubmissionResult)
				}
			}
		})
	}
}

func TestListExamLOSubmissions(t *testing.T) {
	t.Parallel()
	examLOSubmissionRepo := &mock_repositories.MockExamLOSubmissionRepo{}
	studentRepo := &mock_repositories.MockStudentRepo{}
	mockDB := &mock_database.Ext{}
	svc := &ExamLOService{
		DB:                   mockDB,
		StudentRepo:          studentRepo,
		ExamLOSubmissionRepo: examLOSubmissionRepo,
	}

	validReq := &sspb.ListExamLOSubmissionRequest{
		ClassIds: []string{"class-id-1,class-id-2"},
		CourseId: wrapperspb.String("course-id-1"),
		Start:    timestamppb.Now(),
		End:      timestamppb.Now(),
		Statuses: []sspb.SubmissionStatus{
			sspb.SubmissionStatus_SUBMISSION_STATUS_IN_PROGRESS,
			sspb.SubmissionStatus_SUBMISSION_STATUS_MARKED,
		},
		Paging: &cpb.Paging{},
	}

	testCases := []TestCase{
		{
			name:        "error no rows find student by class IDs",
			req:         validReq,
			expectedErr: status.Errorf(codes.Internal, fmt.Errorf("s.getStudentIDsV2: %w", pgx.ErrNoRows).Error()),
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
			req: &sspb.ListExamLOSubmissionRequest{
				CourseId: wrapperspb.String("course-id-1"),
				Start:    timestamppb.Now(),
				End:      timestamppb.Now(),
				Statuses: []sspb.SubmissionStatus{
					sspb.SubmissionStatus_SUBMISSION_STATUS_IN_PROGRESS,
					sspb.SubmissionStatus_SUBMISSION_STATUS_MARKED,
				},
				Paging: &cpb.Paging{},
			},
			expectedErr: status.Errorf(codes.Internal, fmt.Errorf("s.getStudentIDsV2: %w", pgx.ErrNoRows).Error()),
			setup: func(ctx context.Context) {
				studentIDs := database.TextArray([]string{"student-id-1", "student-id-2"})
				studentRepo.On("FindStudentsByCourseLocation", ctx, mockDB, mock.Anything, mock.Anything).Once().Return(
					&studentIDs,
					pgx.ErrNoRows,
				)
			},
		},
		{
			name:        "error list submission",
			req:         validReq,
			expectedErr: status.Errorf(codes.Internal, fmt.Errorf("s.ExamLOSubmissionRepo.List: %w", pgx.ErrNoRows).Error()),
			setup: func(ctx context.Context) {
				studentIDs := database.TextArray([]string{"student-id-1", "student-id-2"})
				studentRepo.On("FindStudentsByClassIDs", ctx, mockDB, mock.Anything).Once().Return(
					&studentIDs,
					nil,
				)
				submissions := []*repositories.ExtendedExamLOSubmission{}
				examLOSubmissionRepo.On("List", ctx, mockDB, mock.Anything).Once().Return(
					submissions, pgx.ErrNoRows,
				)
			},
		},
		{
			name:        "happy case",
			req:         validReq,
			expectedErr: nil,
			setup: func(ctx context.Context) {
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
				submissions := []*repositories.ExtendedExamLOSubmission{
					{
						ExamLOSubmission: entities.ExamLOSubmission{
							LearningMaterialID: database.Text("learning-material-id-1"),
							StudentID:          database.Text("student-id"),
						},
					},
					{
						ExamLOSubmission: entities.ExamLOSubmission{
							LearningMaterialID: database.Text("learning-material-id-2"),
							StudentID:          database.Text("student-id"),
						},
					},
					{
						ExamLOSubmission: entities.ExamLOSubmission{
							LearningMaterialID: database.Text("learning-material-id-3"),
							StudentID:          database.Text("student-id"),
						},
					},
				}
				examLOSubmissionRepo.On("List", ctx, mockDB, mock.Anything).Once().Return(
					submissions, nil,
				)
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		_, err := svc.ListExamLOSubmission(ctx, testCase.req.(*sspb.ListExamLOSubmissionRequest))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}

func TestListExamLOSubmissionResult(t *testing.T) {
	t.Parallel()
	examLOSubmissionRepo := &mock_repositories.MockExamLOSubmissionRepo{}
	examLORepo := &mock_repositories.MockExamLORepo{}
	shuffledQuizSetRepo := &mock_repositories.MockShuffledQuizSetRepo{}
	studentEventLogRepo := &mock_repositories.MockStudentEventLogRepo{}
	studentRepo := &mock_repositories.MockStudentRepo{}
	mockDB := &mock_database.Ext{}
	svc := &ExamLOService{
		DB:                        mockDB,
		StudentRepo:               studentRepo,
		ShuffledQuizSetRepo:       shuffledQuizSetRepo,
		ExamLOSubmissionRepo:      examLOSubmissionRepo,
		ExamLORepo:                examLORepo,
		StudentEventLogRepo:       studentEventLogRepo,
		LearningTimeCalculatorSvc: &LearningTimeCalculator{},
	}

	testCases := []TestCase{
		{
			name: "error no rows find student by class IDs",
			req: &sspb.ListExamLOSubmissionResultRequest{
				StudyPlanItemIdentities: []*sspb.StudyPlanItemIdentity{
					{
						StudyPlanId:        "study-plan-id",
						LearningMaterialId: "learning-material-id",
						StudentId:          wrapperspb.String("student-id"),
					},
				},
			},
			expectedErr: status.Errorf(codes.Internal, fmt.Errorf("s.ExamLOSubmissionRepo.ListExamLOSubmissionWithDates: %w", pgx.ErrNoRows).Error()),
			setup: func(ctx context.Context) {
				examLOSubmissionRepo.On("ListExamLOSubmissionWithDates", ctx, mockDB, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
			},
		},
		{
			name:        "empty request",
			req:         &sspb.ListExamLOSubmissionResultRequest{},
			expectedErr: nil,
			setup:       func(ctx context.Context) {},
		},
		{
			name: "happy case",
			req: &sspb.ListExamLOSubmissionResultRequest{
				StudyPlanItemIdentities: []*sspb.StudyPlanItemIdentity{
					{
						StudyPlanId:        "study-plan-id-1",
						LearningMaterialId: "learning-material-id-1",
						StudentId:          wrapperspb.String("student-id"),
					},
					{
						StudyPlanId:        "study-plan-id-2",
						LearningMaterialId: "learning-material-id-2",
						StudentId:          wrapperspb.String("student-id"),
					},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				examLORepo.On("ListByIDs", ctx, mockDB, mock.Anything).Once().Return([]*entities.ExamLO{
					{
						LearningMaterial: entities.LearningMaterial{
							ID: database.Text("learning-material-id-1"),
						},
						ReviewOption: database.Text(sspb.ExamLOReviewOption_EXAM_LO_REVIEW_OPTION_AFTER_DUE_DATE.String()),
					},
					{
						LearningMaterial: entities.LearningMaterial{
							ID: database.Text("learning-material-id-2"),
						},
						ReviewOption: database.Text(sspb.ExamLOReviewOption_EXAM_LO_REVIEW_OPTION_AFTER_DUE_DATE.String()),
					},
				}, nil)
				examLOSubmissionRepo.On("ListExamLOSubmissionWithDates", ctx, mockDB, mock.Anything).Once().Return([]*repositories.ExtendedExamLOSubmission{
					{
						ExamLOSubmission: entities.ExamLOSubmission{
							SubmissionID:       database.Text("submission-id-1"),
							StudentID:          database.Text("student-id"),
							StudyPlanID:        database.Text("study-plan-id-1"),
							LearningMaterialID: database.Text("learning-material-id-1"),
							ShuffledQuizSetID:  database.Text("shuffled-quiz-set-id-1"),
						},
						EndDate: database.Timestamptz(time.Now().Add(-100 * time.Hour)),
					},
					{
						ExamLOSubmission: entities.ExamLOSubmission{
							SubmissionID:       database.Text("submission-id-2"),
							StudentID:          database.Text("student-id"),
							StudyPlanID:        database.Text("study-plan-id-2"),
							LearningMaterialID: database.Text("learning-material-id-2"),
							ShuffledQuizSetID:  database.Text("shuffled-quiz-set-id-2"),
						},
						EndDate: database.Timestamptz(time.Now().Add(100 * time.Hour)),
					},
				}, nil)
				examLOSubmissionRepo.On("ListTotalGradePoints", ctx, mockDB, mock.Anything).Once().Return([]*repositories.ExamLOSubmissionWithGrade{
					{
						ExamLOSubmission: entities.ExamLOSubmission{
							SubmissionID: database.Text("submission-id-1"),
						},
						TotalGradePoint: database.Int2(4),
					},
					{
						ExamLOSubmission: entities.ExamLOSubmission{
							SubmissionID: database.Text("submission-id-2"),
						},
						TotalGradePoint: database.Int2(6),
					},
				}, nil)
				shuffledQuizSetRepo.On("Retrieve", ctx, mockDB, mock.Anything).Once().Return([]*entities.ShuffledQuizSet{
					{
						StudentID:          database.Text("student-id"),
						StudyPlanID:        database.Text("study-plan-id-1"),
						LearningMaterialID: database.Text("learning-material-id-1"),
					},
					{
						StudentID:          database.Text("student-id"),
						StudyPlanID:        database.Text("study-plan-id-2"),
						LearningMaterialID: database.Text("learning-material-id-2"),
					},
				}, nil)
				studentEventLogRepo.On("RetrieveStudentEventLogsByStudyPlanIdentities", ctx, mockDB, mock.Anything).Once().Return(nil, nil)
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		_, err := svc.ListExamLOSubmissionResult(ctx, testCase.req.(*sspb.ListExamLOSubmissionResultRequest))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}

func Test_ListExamLOSubmissionScore(t *testing.T) {
	examLOSubmissionRepo := &mock_repositories.MockExamLOSubmissionRepo{}
	examLOSubmissionScoreRepo := &mock_repositories.MockExamLOSubmissionScoreRepo{}
	examLOSubmissionAnswerRepo := &mock_repositories.MockExamLOSubmissionAnswerRepo{}
	quizRepo := &mock_repositories.MockQuizRepo{}
	shuffledQuizSetRepo := &mock_repositories.MockShuffledQuizSetRepo{}
	studentRepo := &mock_repositories.MockStudentRepo{}
	questionGroupRepo := &mock_repositories.MockQuestionGroupRepo{}

	mockDB := &mock_database.Ext{}
	svc := &ExamLOService{
		DB:                         mockDB,
		StudentRepo:                studentRepo,
		ExamLOSubmissionRepo:       examLOSubmissionRepo,
		ExamLOSubmissionScoreRepo:  examLOSubmissionScoreRepo,
		ExamLOSubmissionAnswerRepo: examLOSubmissionAnswerRepo,
		QuizRepo:                   quizRepo,
		ShuffledQuizSetRepo:        shuffledQuizSetRepo,
		QuestionGroupRepo:          questionGroupRepo,
	}

	quizzes := getQuizzes(4,
		cpb.QuizType_QUIZ_TYPE_MCQ.String(),
		cpb.QuizType_QUIZ_TYPE_FIB.String(),
		cpb.QuizType_QUIZ_TYPE_ORD.String(),
		cpb.QuizType_QUIZ_TYPE_ESQ.String(),
	)
	quizzes[0].ID = database.Text("quiz-id-0")
	quizzes[0].ExternalID = database.Text("external-quiz-id-0")

	quizzes[1].ID = database.Text("quiz-id-1")
	quizzes[1].ExternalID = database.Text("external-quiz-id-1")

	quizzes[2].ID = database.Text("quiz-id-2")
	quizzes[2].ExternalID = database.Text("external-quiz-id-2")

	quizzes[3].ID = database.Text("quiz-id-3")
	quizzes[3].ExternalID = database.Text("external-quiz-id-3")

	loID := database.Text("lo_id")
	url := "sample-text"

	questionGroups := entities.QuestionGroups{
		{
			BaseEntity:         entities.BaseEntity{},
			QuestionGroupID:    database.Text("group-1"),
			LearningMaterialID: loID,
			Name:               database.Text("name 1"),
			Description:        database.Text("description 1"),
			RichDescription: database.JSONB(&entities.RichText{
				Raw:         "raw rich text",
				RenderedURL: url,
			}),
		},
		{
			BaseEntity:         entities.BaseEntity{},
			QuestionGroupID:    database.Text("group-2"),
			LearningMaterialID: loID,
			Name:               database.Text("name 2"),
			Description:        database.Text("description 2"),
			RichDescription: database.JSONB(&entities.RichText{
				Raw:         "raw rich text",
				RenderedURL: url,
			}),
		},
	}
	questionGroups[0].SetTotalChildrenAndPoints(2, 3)
	questionGroups[1].SetTotalChildrenAndPoints(5, 6)

	examLOsubmissionScores := []*entities.ExamLOSubmissionScore{
		{
			SubmissionID:      database.Text("submission-id"),
			ShuffledQuizSetID: database.Text("shuffled-quiz-set-id"),
			Point:             database.Int4(1),
			TeacherComment:    database.Text("teacher-comment-0"),
			QuizID:            database.Text("external-quiz-id-0"),
		},
		{
			SubmissionID:      database.Text("submission-id"),
			ShuffledQuizSetID: database.Text("shuffled-quiz-set-id"),
			Point:             database.Int4(5),
			TeacherComment:    database.Text("teacher-comment-1"),
			QuizID:            database.Text("external-quiz-id-1"),
		},
		{
			SubmissionID:      database.Text("submission-id"),
			ShuffledQuizSetID: database.Text("shuffled-quiz-set-id"),
			Point:             database.Int4(1),
			TeacherComment:    database.Text("teacher-comment-2"),
			QuizID:            database.Text("external-quiz-id-2"),
		},
		{
			SubmissionID:      database.Text("submission-id"),
			ShuffledQuizSetID: database.Text("shuffled-quiz-set-id"),
			Point:             database.Int4(1),
			TeacherComment:    database.Text("teacher-comment-3"),
			QuizID:            database.Text("external-quiz-id-3"),
		},
	}
	examLOsubmissionAnswers := []*entities.ExamLOSubmissionAnswer{
		{
			SubmissionID:      database.Text("submission-id"),
			ShuffledQuizSetID: database.Text("shuffled-quiz-set-id"),
			QuizID:            database.Text("external-quiz-id-0"),
			IsCorrect: pgtype.BoolArray{
				Elements: []pgtype.Bool{database.Bool(true), database.Bool(true), database.Bool(true)},
				Status:   pgtype.Present,
			},
			StudentIndexAnswer: database.Int4Array([]int32{1, 2, 3}),
			CorrectIndexAnswer: database.Int4Array([]int32{1, 2, 3}),
			CorrectTextAnswer:  database.TextArray([]string{"1", "2", "3"}),
			StudentTextAnswer:  database.TextArray([]string{}),
			IsAccepted:         database.Bool(true),
			Point:              database.Int4(1),
		},
		{
			SubmissionID:      database.Text("submission-id"),
			ShuffledQuizSetID: database.Text("shuffled-quiz-set-id"),
			QuizID:            database.Text("external-quiz-id-1"),
			IsCorrect: pgtype.BoolArray{
				Elements: []pgtype.Bool{database.Bool(true), database.Bool(true), database.Bool(true)},
				Status:   pgtype.Present,
			},
			StudentIndexAnswer: database.Int4Array([]int32{}),
			CorrectIndexAnswer: database.Int4Array([]int32{1, 2, 3}),
			CorrectTextAnswer:  database.TextArray([]string{"1", "2", "3"}),
			StudentTextAnswer:  database.TextArray([]string{"1", "2", "3"}),
			IsAccepted:         database.Bool(true),
			Point:              database.Int4(5),
		},
		{
			SubmissionID:      database.Text("submission-id"),
			ShuffledQuizSetID: database.Text("shuffled-quiz-set-id"),
			QuizID:            database.Text("external-quiz-id-2"),
			IsCorrect: pgtype.BoolArray{
				Elements: []pgtype.Bool{database.Bool(true), database.Bool(true), database.Bool(true)},
				Status:   pgtype.Present,
			},
			StudentIndexAnswer:  database.Int4Array([]int32{}),
			CorrectIndexAnswer:  database.Int4Array([]int32{1, 2, 3}),
			CorrectTextAnswer:   database.TextArray([]string{"1", "2", "3"}),
			StudentTextAnswer:   database.TextArray([]string{}),
			CorrectKeysAnswer:   database.TextArray([]string{"1", "2", "3"}),
			SubmittedKeysAnswer: database.TextArray([]string{"key-1", "key-2", "key-3"}),
			IsAccepted:          database.Bool(true),
			Point:               database.Int4(1),
		},
		{
			SubmissionID:      database.Text("submission-id"),
			ShuffledQuizSetID: database.Text("shuffled-quiz-set-id"),
			QuizID:            database.Text("external-quiz-id-3"),
			IsCorrect: pgtype.BoolArray{
				Elements: []pgtype.Bool{database.Bool(true), database.Bool(true), database.Bool(true)},
				Status:   pgtype.Present,
			},
			StudentIndexAnswer:  database.Int4Array([]int32{}),
			CorrectIndexAnswer:  database.Int4Array([]int32{1, 2, 3}),
			CorrectTextAnswer:   database.TextArray([]string{"1", "2", "3"}),
			StudentTextAnswer:   database.TextArray([]string{}),
			CorrectKeysAnswer:   database.TextArray([]string{"1", "2", "3"}),
			SubmittedKeysAnswer: database.TextArray([]string{"key-1", "key-2", "key-3"}),
			IsAccepted:          database.Bool(true),
			Point:               database.Int4(1),
		},
	}
	tagsPerQuiz := map[string][]string{
		"external-quiz-id-0": {"tag-name-0", "tag-name-1"},
		"external-quiz-id-1": {"tag-name-0", "tag-name-1"},
		"external-quiz-id-2": {"tag-name-0", "tag-name-1"},
		"external-quiz-id-3": {"tag-name-0", "tag-name-1"},
	}
	respQuestionGroups, _ := entities.QuestionGroupsToQuestionGroupProtoBufMess(questionGroups)
	respNilQuestionGroups, _ := entities.QuestionGroupsToQuestionGroupProtoBufMess(nil)

	testCases := []TestCase{
		{
			name: "error empty submission_id",
			req: &sspb.ListExamLOSubmissionScoreRequest{
				SubmissionId: "",
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "empty submission_id"),
			setup:        func(ctx context.Context) {},
		},
		{
			name: "error empty shuffled_quiz_set_id",
			req: &sspb.ListExamLOSubmissionScoreRequest{
				SubmissionId:      "submission-id",
				ShuffledQuizSetId: "",
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "empty shuffled_quiz_set_id"),
			setup:        func(ctx context.Context) {},
		},
		{
			name: "success",
			req: &sspb.ListExamLOSubmissionScoreRequest{
				SubmissionId:      "submission-id",
				ShuffledQuizSetId: "shuffled-quiz-set-id",
			},
			expectedResp: &sspb.ListExamLOSubmissionScoreResponse{
				SubmissionScores: []*sspb.ExamLOSubmissionScore{
					{
						ShuffleQuizSetId: "shuffled-quiz-set-id",
						QuizType:         cpb.QuizType_QUIZ_TYPE_MCQ,
						SelectedIndex:    []uint32{1, 2, 3},
						CorrectIndex:     []uint32{1, 2},
						FilledText:       []string{},
						CorrectText:      []string{},
						Correctness:      []bool{true, true, true},
						IsAccepted:       true,
						Core:             nil,
						TeacherComment:   "teacher-comment-0",
						GradedPoint:      wrapperspb.UInt32(1),
						Point:            wrapperspb.UInt32(1),
					},
					{
						ShuffleQuizSetId: "shuffled-quiz-set-id",
						QuizType:         cpb.QuizType_QUIZ_TYPE_FIB,
						SelectedIndex:    []uint32{},
						CorrectIndex:     []uint32{},
						FilledText:       []string{"1", "2", "3"},
						CorrectText:      []string{"3213213", "3213214", "3213215"},
						Correctness:      []bool{true, true, true},
						IsAccepted:       true,
						Core:             nil,
						TeacherComment:   "teacher-comment-1",
						GradedPoint:      wrapperspb.UInt32(5),
						Point:            wrapperspb.UInt32(5),
					},
					{
						ShuffleQuizSetId: "shuffled-quiz-set-id",
						QuizType:         cpb.QuizType_QUIZ_TYPE_ORD,
						SelectedIndex:    []uint32{},
						CorrectIndex:     []uint32{},
						FilledText:       []string{},
						CorrectText:      []string{},
						Correctness:      []bool{true, true, true},
						Core:             nil,
						TeacherComment:   "teacher-comment-2",
						GradedPoint:      wrapperspb.UInt32(1),
						Point:            wrapperspb.UInt32(1),
						IsAccepted:       true,
						Result: &sspb.ExamLOSubmissionScore_OrderingResult{
							OrderingResult: &cpb.OrderingResult{
								CorrectKeys:   []string{"key-1", "key-2", "key-3"},
								SubmittedKeys: []string{"key-1", "key-2", "key-3"},
							},
						},
					},
					{
						ShuffleQuizSetId: "shuffled-quiz-set-id",
						QuizType:         cpb.QuizType_QUIZ_TYPE_ESQ,
						SelectedIndex:    []uint32{},
						CorrectIndex:     []uint32{},
						FilledText:       []string{},
						CorrectText:      []string{},
						Correctness:      []bool{true, true, true},
						Core:             nil,
						TeacherComment:   "teacher-comment-3",
						GradedPoint:      wrapperspb.UInt32(1),
						Point:            wrapperspb.UInt32(1),
						IsAccepted:       true,
					},
				},
				SubmissionStatus: sspb.SubmissionStatus_SUBMISSION_STATUS_IN_PROGRESS,
				SubmissionResult: sspb.ExamLOSubmissionResult_EXAM_LO_SUBMISSION_PASSED,
				TeacherFeedback:  "very good !",
				TotalGradedPoint: wrapperspb.UInt32(7),
				TotalPoint:       wrapperspb.UInt32(7),
				QuestionGroups:   respNilQuestionGroups,
			},
			setup: func(ctx context.Context) {
				examLOSubmissionRepo.
					On("Get", ctx, mockDB, &repositories.GetExamLOSubmissionArgs{
						SubmissionID:      database.Text("submission-id"),
						ShuffledQuizSetID: database.Text("shuffled-quiz-set-id"),
					}).
					Once().
					Return(&entities.ExamLOSubmission{
						SubmissionID:      database.Text("submission-id"),
						ShuffledQuizSetID: database.Text("shuffled-quiz-set-id"),
						Status:            database.Text(sspb.SubmissionStatus_SUBMISSION_STATUS_IN_PROGRESS.String()),
						Result:            database.Text(sspb.ExamLOSubmissionResult_EXAM_LO_SUBMISSION_PASSED.String()),
						TotalPoint:        database.Int4(7),
						TeacherFeedback:   database.Text("very good !"),
					}, nil)
				examLOSubmissionScoreRepo.
					On("List", ctx, mockDB, &repositories.ExamLOSubmissionScoreFilter{
						SubmissionID:      database.Text("submission-id"),
						ShuffledQuizSetID: database.Text("shuffled-quiz-set-id"),
					}).
					Once().Return(examLOsubmissionScores, nil)
				examLOSubmissionAnswerRepo.
					On("List", ctx, mockDB, &repositories.ExamLOSubmissionAnswerFilter{
						SubmissionID:      database.Text("submission-id"),
						ShuffledQuizSetID: database.Text("shuffled-quiz-set-id"),
					}).
					Once().Return(examLOsubmissionAnswers, nil)
				shuffledQuizSetRepo.
					On("GetSeed", ctx, mockDB, database.Text("shuffled-quiz-set-id")).
					Twice().
					Return(database.Text("1668152513982496812"), nil)
				shuffledQuizSetRepo.
					On("GetQuizIdx", ctx, mockDB, database.Text("shuffled-quiz-set-id"), database.Text("external-quiz-id-0")).
					Once().
					Return(database.Int4(1), nil)
				shuffledQuizSetRepo.
					On("GetQuizIdx", ctx, mockDB, database.Text("shuffled-quiz-set-id"), database.Text("external-quiz-id-2")).
					Once().
					Return(database.Int4(3), nil)
				quizRepo.
					On("Search", ctx, mockDB, repositories.QuizFilter{
						ExternalIDs: database.TextArray([]string{"external-quiz-id-0", "external-quiz-id-1", "external-quiz-id-2", "external-quiz-id-3"}),
						Status: pgtype.Text{
							Status: pgtype.Null,
						},
						Limit: uint(4),
					}).
					Once().
					Return(quizzes, nil)
				quizRepo.
					On("GetTagNames", ctx, mockDB, database.TextArray([]string{"external-quiz-id-0", "external-quiz-id-1", "external-quiz-id-2", "external-quiz-id-3"})).
					Once().
					Return(tagsPerQuiz, nil)
				questionGroupRepo.On("GetQuestionGroupsByIDs", ctx, mockDB, []string{}).Once().Return(entities.QuestionGroups{}, nil)
			},
		},
		{
			name: "success with question group",
			req: &sspb.ListExamLOSubmissionScoreRequest{
				SubmissionId:      "submission-id",
				ShuffledQuizSetId: "shuffled-quiz-set-id",
			},
			expectedResp: &sspb.ListExamLOSubmissionScoreResponse{
				SubmissionScores: []*sspb.ExamLOSubmissionScore{
					{
						ShuffleQuizSetId: "shuffled-quiz-set-id",
						QuizType:         cpb.QuizType_QUIZ_TYPE_MCQ,
						SelectedIndex:    []uint32{1, 2, 3},
						CorrectIndex:     []uint32{1, 2},
						FilledText:       []string{},
						CorrectText:      []string{},
						Correctness:      []bool{true, true, true},
						IsAccepted:       true,
						Core:             nil,
						TeacherComment:   "teacher-comment-0",
						GradedPoint:      wrapperspb.UInt32(1),
						Point:            wrapperspb.UInt32(1),
					},
					{
						ShuffleQuizSetId: "shuffled-quiz-set-id",
						QuizType:         cpb.QuizType_QUIZ_TYPE_FIB,
						SelectedIndex:    []uint32{},
						CorrectIndex:     []uint32{},
						FilledText:       []string{"1", "2", "3"},
						CorrectText:      []string{"3213213", "3213214", "3213215"},
						Correctness:      []bool{true, true, true},
						IsAccepted:       true,
						Core:             nil,
						TeacherComment:   "teacher-comment-1",
						GradedPoint:      wrapperspb.UInt32(5),
						Point:            wrapperspb.UInt32(5),
					},
					{
						ShuffleQuizSetId: "shuffled-quiz-set-id",
						QuizType:         cpb.QuizType_QUIZ_TYPE_ORD,
						SelectedIndex:    []uint32{},
						CorrectIndex:     []uint32{},
						FilledText:       []string{},
						CorrectText:      []string{},
						Correctness:      []bool{true, true, true},
						Core:             nil,
						TeacherComment:   "teacher-comment-2",
						GradedPoint:      wrapperspb.UInt32(1),
						Point:            wrapperspb.UInt32(1),
						IsAccepted:       true,
						Result: &sspb.ExamLOSubmissionScore_OrderingResult{
							OrderingResult: &cpb.OrderingResult{
								CorrectKeys:   []string{"key-1", "key-2", "key-3"},
								SubmittedKeys: []string{"key-1", "key-2", "key-3"},
							},
						},
					},
					{
						ShuffleQuizSetId: "shuffled-quiz-set-id",
						QuizType:         cpb.QuizType_QUIZ_TYPE_ESQ,
						SelectedIndex:    []uint32{},
						CorrectIndex:     []uint32{},
						FilledText:       []string{},
						CorrectText:      []string{},
						Correctness:      []bool{true, true, true},
						Core:             nil,
						TeacherComment:   "teacher-comment-3",
						GradedPoint:      wrapperspb.UInt32(1),
						Point:            wrapperspb.UInt32(1),
						IsAccepted:       true,
					},
				},
				SubmissionStatus: sspb.SubmissionStatus_SUBMISSION_STATUS_IN_PROGRESS,
				SubmissionResult: sspb.ExamLOSubmissionResult_EXAM_LO_SUBMISSION_PASSED,
				TeacherFeedback:  "very good !",
				TotalGradedPoint: wrapperspb.UInt32(7),
				TotalPoint:       wrapperspb.UInt32(7),
				QuestionGroups:   respQuestionGroups,
			},
			setup: func(ctx context.Context) {
				groupedQuizzes := getQuizzes(4,
					cpb.QuizType_QUIZ_TYPE_MCQ.String(),
					cpb.QuizType_QUIZ_TYPE_FIB.String(),
					cpb.QuizType_QUIZ_TYPE_ORD.String(),
					cpb.QuizType_QUIZ_TYPE_ESQ.String(),
				)
				groupedQuizzes[0].ID = database.Text("quiz-id-0")
				groupedQuizzes[0].ExternalID = database.Text("external-quiz-id-0")
				groupedQuizzes[0].QuestionGroupID = database.Text("group-1")

				groupedQuizzes[1].ID = database.Text("quiz-id-1")
				groupedQuizzes[1].ExternalID = database.Text("external-quiz-id-1")
				groupedQuizzes[1].QuestionGroupID = database.Text("group-1")

				groupedQuizzes[2].ID = database.Text("quiz-id-2")
				groupedQuizzes[2].ExternalID = database.Text("external-quiz-id-2")

				groupedQuizzes[3].ID = database.Text("quiz-id-3")
				groupedQuizzes[3].ExternalID = database.Text("external-quiz-id-3")
				groupedQuizzes[3].QuestionGroupID = database.Text("group-2")

				examLOSubmissionRepo.
					On("Get", ctx, mockDB, &repositories.GetExamLOSubmissionArgs{
						SubmissionID:      database.Text("submission-id"),
						ShuffledQuizSetID: database.Text("shuffled-quiz-set-id"),
					}).
					Once().
					Return(&entities.ExamLOSubmission{
						SubmissionID:      database.Text("submission-id"),
						ShuffledQuizSetID: database.Text("shuffled-quiz-set-id"),
						Status:            database.Text(sspb.SubmissionStatus_SUBMISSION_STATUS_IN_PROGRESS.String()),
						Result:            database.Text(sspb.ExamLOSubmissionResult_EXAM_LO_SUBMISSION_PASSED.String()),
						TotalPoint:        database.Int4(7),
						TeacherFeedback:   database.Text("very good !"),
					}, nil)
				examLOSubmissionScoreRepo.
					On("List", ctx, mockDB, &repositories.ExamLOSubmissionScoreFilter{
						SubmissionID:      database.Text("submission-id"),
						ShuffledQuizSetID: database.Text("shuffled-quiz-set-id"),
					}).
					Once().Return(examLOsubmissionScores, nil)
				examLOSubmissionAnswerRepo.
					On("List", ctx, mockDB, &repositories.ExamLOSubmissionAnswerFilter{
						SubmissionID:      database.Text("submission-id"),
						ShuffledQuizSetID: database.Text("shuffled-quiz-set-id"),
					}).
					Once().Return(examLOsubmissionAnswers, nil)
				shuffledQuizSetRepo.
					On("GetSeed", ctx, mockDB, database.Text("shuffled-quiz-set-id")).
					Twice().
					Return(database.Text("1668152513982496812"), nil)
				shuffledQuizSetRepo.
					On("GetQuizIdx", ctx, mockDB, database.Text("shuffled-quiz-set-id"), database.Text("external-quiz-id-0")).
					Once().
					Return(database.Int4(1), nil)
				shuffledQuizSetRepo.
					On("GetQuizIdx", ctx, mockDB, database.Text("shuffled-quiz-set-id"), database.Text("external-quiz-id-2")).
					Once().
					Return(database.Int4(3), nil)
				quizRepo.
					On("Search", ctx, mockDB, repositories.QuizFilter{
						ExternalIDs: database.TextArray([]string{"external-quiz-id-0", "external-quiz-id-1", "external-quiz-id-2", "external-quiz-id-3"}),
						Status: pgtype.Text{
							Status: pgtype.Null,
						},
						Limit: uint(4),
					}).
					Once().
					Return(groupedQuizzes, nil)
				quizRepo.
					On("GetTagNames", ctx, mockDB, database.TextArray([]string{"external-quiz-id-0", "external-quiz-id-1", "external-quiz-id-2", "external-quiz-id-3"})).
					Once().
					Return(tagsPerQuiz, nil)
				questionGroupRepo.On("GetQuestionGroupsByIDs", ctx, mockDB, []string{"group-1", "group-2"}).Once().Return(questionGroups, nil)
			},
		},
		{
			name: "err get question group by ids",
			req: &sspb.ListExamLOSubmissionScoreRequest{
				SubmissionId:      "submission-id",
				ShuffledQuizSetId: "shuffled-quiz-set-id",
			},
			expectedErr: status.Error(codes.Internal, "QuestionGroupRepo.GetQuestionGroupsByIDs: tx is closed"),
			setup: func(ctx context.Context) {
				groupedQuizzes := getQuizzes(4,
					cpb.QuizType_QUIZ_TYPE_MCQ.String(),
					cpb.QuizType_QUIZ_TYPE_FIB.String(),
					cpb.QuizType_QUIZ_TYPE_ORD.String(),
					cpb.QuizType_QUIZ_TYPE_ESQ.String(),
				)
				groupedQuizzes[0].ID = database.Text("quiz-id-0")
				groupedQuizzes[0].ExternalID = database.Text("external-quiz-id-0")
				groupedQuizzes[0].QuestionGroupID = database.Text("group-1")

				groupedQuizzes[1].ID = database.Text("quiz-id-1")
				groupedQuizzes[1].ExternalID = database.Text("external-quiz-id-1")
				groupedQuizzes[1].QuestionGroupID = database.Text("group-1")

				groupedQuizzes[2].ID = database.Text("quiz-id-2")
				groupedQuizzes[2].ExternalID = database.Text("external-quiz-id-2")

				groupedQuizzes[3].ID = database.Text("quiz-id-3")
				groupedQuizzes[3].ExternalID = database.Text("external-quiz-id-3")
				groupedQuizzes[3].QuestionGroupID = database.Text("group-2")

				examLOSubmissionRepo.
					On("Get", ctx, mockDB, &repositories.GetExamLOSubmissionArgs{
						SubmissionID:      database.Text("submission-id"),
						ShuffledQuizSetID: database.Text("shuffled-quiz-set-id"),
					}).
					Once().
					Return(&entities.ExamLOSubmission{
						SubmissionID:      database.Text("submission-id"),
						ShuffledQuizSetID: database.Text("shuffled-quiz-set-id"),
						Status:            database.Text(sspb.SubmissionStatus_SUBMISSION_STATUS_IN_PROGRESS.String()),
						Result:            database.Text(sspb.ExamLOSubmissionResult_EXAM_LO_SUBMISSION_PASSED.String()),
						TotalPoint:        database.Int4(7),
						TeacherFeedback:   database.Text("very good !"),
					}, nil)
				examLOSubmissionScoreRepo.
					On("List", ctx, mockDB, &repositories.ExamLOSubmissionScoreFilter{
						SubmissionID:      database.Text("submission-id"),
						ShuffledQuizSetID: database.Text("shuffled-quiz-set-id"),
					}).
					Once().Return(examLOsubmissionScores, nil)
				examLOSubmissionAnswerRepo.
					On("List", ctx, mockDB, &repositories.ExamLOSubmissionAnswerFilter{
						SubmissionID:      database.Text("submission-id"),
						ShuffledQuizSetID: database.Text("shuffled-quiz-set-id"),
					}).
					Once().Return(examLOsubmissionAnswers, nil)
				shuffledQuizSetRepo.
					On("GetSeed", ctx, mockDB, database.Text("shuffled-quiz-set-id")).
					Twice().
					Return(database.Text("1668152513982496812"), nil)
				shuffledQuizSetRepo.
					On("GetQuizIdx", ctx, mockDB, database.Text("shuffled-quiz-set-id"), database.Text("external-quiz-id-0")).
					Once().
					Return(database.Int4(1), nil)
				shuffledQuizSetRepo.
					On("GetQuizIdx", ctx, mockDB, database.Text("shuffled-quiz-set-id"), database.Text("external-quiz-id-2")).
					Once().
					Return(database.Int4(3), nil)
				quizRepo.
					On("Search", ctx, mockDB, repositories.QuizFilter{
						ExternalIDs: database.TextArray([]string{"external-quiz-id-0", "external-quiz-id-1", "external-quiz-id-2", "external-quiz-id-3"}),
						Status: pgtype.Text{
							Status: pgtype.Null,
						},
						Limit: uint(4),
					}).
					Once().
					Return(groupedQuizzes, nil)
				quizRepo.
					On("GetTagNames", ctx, mockDB, database.TextArray([]string{"external-quiz-id-0", "external-quiz-id-1", "external-quiz-id-2", "external-quiz-id-3"})).
					Once().
					Return(tagsPerQuiz, nil)
				questionGroupRepo.On("GetQuestionGroupsByIDs", ctx, mockDB, []string{"group-1", "group-2"}).Once().Return(entities.QuestionGroups{}, pgx.ErrTxClosed)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			ctx := context.Background()
			testCase.setup(ctx)
			resp, err := svc.ListExamLOSubmissionScore(ctx, testCase.req.(*sspb.ListExamLOSubmissionScoreRequest))
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				expected := testCase.expectedResp.(*sspb.ListExamLOSubmissionScoreResponse)

				for i, scr := range expected.SubmissionScores {
					scr.Core = resp.SubmissionScores[i].Core
					assert.Equal(t, scr, resp.SubmissionScores[i])
				}

				assert.Equal(t, expected.GetQuestionGroups(), resp.QuestionGroups)
			}
		})
	}
}

func Test_ToBaseExamLO(t *testing.T) {
	testCases := []TestCase{
		{
			name: "happy case",
			req: &entities.ExamLO{
				LearningMaterial: entities.LearningMaterial{
					ID:           database.Text("exam-lo-id"),
					Name:         database.Text("exam-lo-name"),
					TopicID:      database.Text("exam-lo-topic-id"),
					Type:         database.Text(sspb.LearningMaterialType_LEARNING_MATERIAL_EXAM_LO.String()),
					DisplayOrder: database.Int2(0),
				},
			},
			expectedResp: &sspb.ExamLOBase{
				Base: &sspb.LearningMaterialBase{
					LearningMaterialId: "exam-lo-id",
					TopicId:            "exam-lo-topic-id",
					Name:               "exam-lo-name",
					Type:               sspb.LearningMaterialType_LEARNING_MATERIAL_EXAM_LO.String(),
					DisplayOrder: &wrapperspb.Int32Value{
						Value: int32(0),
					},
				},
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			res := ToBaseExamLO(testCase.req.(*entities.ExamLO))
			assert.Equal(t, testCase.expectedResp, res)
		})
	}
}

func Test_GradeAManualGradingExamSubmission(t *testing.T) {
	mockDB := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	examLORepo := &mock_repositories.MockExamLORepo{}
	examLOSubmissionRepo := &mock_repositories.MockExamLOSubmissionRepo{}
	examLOSubmissionScoreRepo := &mock_repositories.MockExamLOSubmissionScoreRepo{}
	allocateMarkerRepo := &mock_repositories.MockAllocateMarkerRepo{}
	ctx := context.Background()

	svc := &ExamLOService{
		DB:                        mockDB,
		ExamLORepo:                examLORepo,
		ExamLOSubmissionRepo:      examLOSubmissionRepo,
		ExamLOSubmissionScoreRepo: examLOSubmissionScoreRepo,
		AllocateMarkerRepo:        allocateMarkerRepo,
	}

	testCases := []TestCase{
		{
			name: "cannot grade the old submission",
			ctx:  interceptors.ContextWithUserGroup(interceptors.ContextWithUserID(ctx, "teacher_id"), constants.RoleTeacher),
			req: &sspb.GradeAManualGradingExamSubmissionRequest{
				SubmissionId:      "submission_id",
				ShuffledQuizSetId: "shuffled_quiz_set_id",
				TeacherFeedback:   "teacher_feedback",
				SubmissionStatus:  sspb.SubmissionStatus_SUBMISSION_STATUS_RETURNED,
				TeacherExamGrades: []*sspb.TeacherExamGrade{
					{
						QuizId:            "quiz_id",
						TeacherPointGiven: wrapperspb.UInt32(2),
						TeacherComment:    "teacher_comment",
						Correctness:       []bool{true},
						IsAccepted:        true,
					},
				},
			},
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Return(tx, nil)
				tx.On("Commit", mock.Anything).Return(nil)

				examLOSubmissionRepo.On("GetLatestSubmissionID", ctx, mockDB, mock.Anything).Once().Return(database.Text("submission_id_tmp"), nil)
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.FailedPrecondition, "cannot grade the old submission"),
		},
		{
			name: "happy case",
			ctx:  interceptors.ContextWithUserGroup(interceptors.ContextWithUserID(ctx, "teacher_id"), constants.RoleTeacher),
			req: &sspb.GradeAManualGradingExamSubmissionRequest{
				SubmissionId:      "submission_id",
				ShuffledQuizSetId: "shuffled_quiz_set_id",
				TeacherFeedback:   "teacher_feedback",
				SubmissionStatus:  sspb.SubmissionStatus_SUBMISSION_STATUS_RETURNED,
				TeacherExamGrades: []*sspb.TeacherExamGrade{
					{
						QuizId:            "quiz_id",
						TeacherPointGiven: wrapperspb.UInt32(2),
						TeacherComment:    "teacher_comment",
						Correctness:       []bool{true},
						IsAccepted:        true,
					},
				},
			},
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Return(tx, nil)
				tx.On("Commit", mock.Anything).Return(nil)

				examLOSubmissionRepo.On("GetLatestSubmissionID", ctx, mockDB, mock.Anything).Once().Return(database.Text("submission_id"), nil)
				examLOSubmissionRepo.On("Get", ctx, mockDB, mock.Anything).Once().Return(&entities.ExamLOSubmission{
					SubmissionID:      database.Text("submission_id"),
					ShuffledQuizSetID: database.Text("shuffled_quiz_set_id"),
					Status:            database.Text(sspb.SubmissionStatus_SUBMISSION_STATUS_NOT_MARKED.String()),
				}, nil)
				examLORepo.On("Get", ctx, mockDB, mock.Anything).Once().Return(&entities.ExamLO{
					ApproveGrading: database.Bool(false),
				}, nil)
				allocateMarkerRepo.On("GetTeacherID", ctx, mockDB, mock.Anything).Once().Return(database.Text("teacher_id"), nil)
				examLOSubmissionScoreRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(1, nil)
				examLOSubmissionRepo.On("GetTotalGradedPoint", ctx, tx, mock.Anything).Once().Return(database.Int4(2), nil)
				examLOSubmissionRepo.On("Update", ctx, tx, mock.Anything).Once().Return(nil)
			},
			expectedResp: nil,
		},
	}

	for _, testCase := range testCases {
		testCase.setup(testCase.ctx)
		_, err := svc.GradeAManualGradingExamSubmission(testCase.ctx, testCase.req.(*sspb.GradeAManualGradingExamSubmissionRequest))
		if testCase.expectedErr != nil {
			fmt.Println(err.Error())
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}

func Test_validateGradeAManualGradingExamSubmissionRequest(t *testing.T) {
	testCases := []TestCase{
		{
			name: "Empty SubmissionId",
			req: &sspb.GradeAManualGradingExamSubmissionRequest{
				SubmissionId:      "",
				ShuffledQuizSetId: "shuffled_quiz_set_id",
			},
			expectedErr: fmt.Errorf("req must have SubmissionId"),
		},
		{
			name: "Empty ShuffledQuizSetId",
			req: &sspb.GradeAManualGradingExamSubmissionRequest{
				SubmissionId:      "submission_id",
				ShuffledQuizSetId: "",
			},
			expectedErr: fmt.Errorf("req must have ShuffledQuizSetId"),
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			err := validateGradeAManualGradingExamSubmissionRequest(testCase.req.(*sspb.GradeAManualGradingExamSubmissionRequest))
			assert.Equal(t, testCase.expectedErr, err)
		})
	}
}

func Test_DeleteExamLOSubmission(t *testing.T) {
	mockDB := &mock_database.Ext{}
	mockTx := &mock_database.Tx{}
	examLORepo := &mock_repositories.MockExamLORepo{}
	examLOSubmissionRepo := &mock_repositories.MockExamLOSubmissionRepo{}
	examLOSubmissionAnswerRepo := &mock_repositories.MockExamLOSubmissionAnswerRepo{}
	examLOSubmissionScoreRepo := &mock_repositories.MockExamLOSubmissionScoreRepo{}
	studyPlanItemRepo := &mock_repositories.MockStudyPlanItemRepo{}
	studentEventLogsRepo := &mock_repositories.MockStudentEventLogRepo{}

	svc := &ExamLOService{
		DB:                         mockDB,
		ExamLORepo:                 examLORepo,
		ExamLOSubmissionRepo:       examLOSubmissionRepo,
		ExamLOSubmissionAnswerRepo: examLOSubmissionAnswerRepo,
		ExamLOSubmissionScoreRepo:  examLOSubmissionScoreRepo,
		StudyPlanItemRepo:          studyPlanItemRepo,
		StudentEventLogRepo:        studentEventLogsRepo,
	}

	examLOSubmission := entities.ExamLOSubmission{
		SubmissionID:       database.Text("examLOSubmission-1-submission-id"),
		ShuffledQuizSetID:  database.Text("examLOSubmission-shuffle-quiz-set-id"),
		StudentID:          database.Text("examLOSubmission-student-id"),
		StudyPlanID:        database.Text("examLOSubmission-study-plan-id"),
		LearningMaterialID: database.Text("examLOSubmission-learning-material-id"),
	}
	examLOSubmissionWrongID := entities.ExamLOSubmission{
		SubmissionID: database.Text("examLOSubmission-1-submission-id-wrong"),
	}
	testCases := []TestCase{
		{
			name: "happy case",
			req: &sspb.DeleteExamLOSubmissionRequest{
				SubmissionId: examLOSubmission.SubmissionID.String,
			},
			setup: func(ctx context.Context) {
				examLOSubmissionRepo.On("GetLatestExamLOSubmission", ctx, mockDB, examLOSubmission.SubmissionID).Return(examLOSubmission, nil)
				mockDB.On("Begin", mock.Anything, mock.Anything).Return(mockTx, nil)
				examLOSubmissionRepo.On("Delete", ctx, mockTx, examLOSubmission.SubmissionID).Once().Return(int64(1), nil)
				examLOSubmissionAnswerRepo.On("Delete", ctx, mockTx, examLOSubmission.SubmissionID).Once().Return(int64(2), nil)
				examLOSubmissionScoreRepo.On("Delete", ctx, mockTx, examLOSubmission.SubmissionID).Once().Return(int64(2), nil)
				examLOSubmissionRepo.On("GetLatestExamLOSubmission", ctx, mockTx, examLOSubmission.SubmissionID).Return(entities.ExamLOSubmission{}, fmt.Errorf("database Select: %w", pgx.ErrNoRows))
				studyPlanItemRepo.On("UpdateCompletedAtToNullByStudyPlanItemIdentity", ctx, mockTx, repositories.StudyPlanItemIdentity{
					StudentID:          examLOSubmission.StudentID,
					StudyPlanID:        examLOSubmission.StudyPlanID,
					LearningMaterialID: examLOSubmission.LearningMaterialID,
				}).Return(int64(1), nil)
				studentEventLogsRepo.On("DeleteByStudyPlanIdentities", ctx, mockTx, repositories.StudyPlanItemIdentity{
					StudentID:          examLOSubmission.StudentID,
					StudyPlanID:        examLOSubmission.StudyPlanID,
					LearningMaterialID: examLOSubmission.LearningMaterialID,
				}).Return(int64(1), nil)
				mockTx.On("Commit", mock.Anything).Return(nil)
			},
			expectedErr: nil,
		},
		{
			name: "id not latest",
			req: &sspb.DeleteExamLOSubmissionRequest{
				SubmissionId: examLOSubmissionWrongID.SubmissionID.String,
			},
			setup: func(ctx context.Context) {
				examLOSubmissionRepo.On("GetLatestExamLOSubmission", ctx, mockDB, examLOSubmissionWrongID.SubmissionID).Once().Return(examLOSubmission, nil)
			},
			expectedErr: status.Error(codes.InvalidArgument, fmt.Errorf("S.DeleteExamLOSubmission, err: exam lo submission not the latest, expected %s, get %s", examLOSubmission.SubmissionID.String, examLOSubmissionWrongID.SubmissionID.String).Error()),
		},
		{
			name: "internal error",
			req: &sspb.DeleteExamLOSubmissionRequest{
				SubmissionId: examLOSubmission.SubmissionID.String,
			},
			setup: func(ctx context.Context) {
				examLOSubmissionRepo.On("GetLatestExamLOSubmission", ctx, mockDB, examLOSubmission.SubmissionID).Once().Return(examLOSubmission, nil)
				mockDB.On("Begin", mock.Anything, mock.Anything).Return(mockTx, nil)
				examLOSubmissionRepo.On("Delete", ctx, mockTx, examLOSubmission.SubmissionID).Once().Return(int64(0), pgx.ErrTxClosed)
				mockTx.On("Rollback", mock.Anything).Return(status.Error(codes.Internal, fmt.Errorf("s.ExamLOSubmissionRepo.Delete, err: %w", pgx.ErrTxClosed).Error()))
			},
			expectedErr: status.Error(codes.Internal, fmt.Errorf("s.ExamLOSubmissionRepo.Delete, err: %w", pgx.ErrTxClosed).Error()),
		},
	}

	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		_, err := svc.DeleteExamLOSubmission(ctx, testCase.req.(*sspb.DeleteExamLOSubmissionRequest))
		assert.Equal(t, testCase.expectedErr, err)
	}
}

func Test_validateDeleteExamLOReq(t *testing.T) {
	testCases := []TestCase{
		{
			name: "happy case",
			req: &sspb.DeleteExamLOSubmissionRequest{
				SubmissionId: "randomID",
			},
			expectedErr: nil,
		},
		{
			name: "Empty submission id",
			req: &sspb.DeleteExamLOSubmissionRequest{
				SubmissionId: "",
			},
			expectedErr: fmt.Errorf("cannot empty submission_id"),
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			err := validateDeleteExamLOReq(testCase.req.(*sspb.DeleteExamLOSubmissionRequest))
			assert.Equal(t, testCase.expectedErr, err)
		})
	}
}

func Test_validateStatusChangeApproveGradingSetup(t *testing.T) {
	notMarked := sspb.SubmissionStatus_SUBMISSION_STATUS_NOT_MARKED.String()
	inProgress := sspb.SubmissionStatus_SUBMISSION_STATUS_IN_PROGRESS.String()
	returned := sspb.SubmissionStatus_SUBMISSION_STATUS_RETURNED.String()

	testCases := []TestCase{
		{
			name: "Teacher/HQ Staff - Assigned - Approve Grading On - Not Marked to Returned",
			req: &ApproveGradingSetup{
				ApproveGrading: true,
				Role:           constants.RoleTeacher,
				IsAssigned:     true,
				Status:         notMarked,
				StatusChange:   returned,
			},
			expectedErr: fmt.Errorf("changing %s status to %s status is not allowed", notMarked, returned),
		},
		{
			name: "Teacher/HQ Staff - Assigned - Approve Grading On - In Progress to Returned",
			req: &ApproveGradingSetup{
				ApproveGrading: true,
				Role:           constants.RoleTeacher,
				IsAssigned:     true,
				Status:         sspb.SubmissionStatus_SUBMISSION_STATUS_IN_PROGRESS.String(),
				StatusChange:   sspb.SubmissionStatus_SUBMISSION_STATUS_RETURNED.String(),
			},
			expectedErr: fmt.Errorf("changing %s status to %s status is not allowed", inProgress, returned),
		},
		{
			name: "Teacher/HQ Staff - Assigned - Approve Grading On - Marked to Any",
			req: &ApproveGradingSetup{
				ApproveGrading: true,
				Role:           constants.RoleTeacher,
				IsAssigned:     true,
				Status:         sspb.SubmissionStatus_SUBMISSION_STATUS_MARKED.String(),
				StatusChange:   sspb.SubmissionStatus_SUBMISSION_STATUS_RETURNED.String(),
			},
			expectedErr: fmt.Errorf("user is not assigned to change the information"),
		},
		{
			name: "Teacher/HQ Staff - Assigned - Approve Grading On - Returned to Any",
			req: &ApproveGradingSetup{
				ApproveGrading: true,
				Role:           constants.RoleTeacher,
				IsAssigned:     true,
				Status:         sspb.SubmissionStatus_SUBMISSION_STATUS_RETURNED.String(),
				StatusChange:   sspb.SubmissionStatus_SUBMISSION_STATUS_IN_PROGRESS.String(),
			},
			expectedErr: fmt.Errorf("user is not assigned to change the information"),
		},
		{
			name: "Teacher/HQ Staff - Not Assigned - Approve Grading On - Not Marked to Any",
			req: &ApproveGradingSetup{
				ApproveGrading: true,
				Role:           constants.RoleTeacher,
				IsAssigned:     false,
				Status:         sspb.SubmissionStatus_SUBMISSION_STATUS_NOT_MARKED.String(),
				StatusChange:   sspb.SubmissionStatus_SUBMISSION_STATUS_IN_PROGRESS.String(),
			},
			expectedErr: fmt.Errorf("user is not assigned to change the information"),
		},
		{
			name: "Teacher/HQ Staff - Not Assigned - Approve Grading Off - Not Marked to Any",
			req: &ApproveGradingSetup{
				ApproveGrading: false,
				Role:           constants.RoleTeacher,
				IsAssigned:     false,
				Status:         sspb.SubmissionStatus_SUBMISSION_STATUS_NOT_MARKED.String(),
				StatusChange:   sspb.SubmissionStatus_SUBMISSION_STATUS_IN_PROGRESS.String(),
			},
			expectedErr: fmt.Errorf("user is not assigned to change the information"),
		},
		{
			name: "Student - Not Assigned - Approve Grading Off - Not Marked to Any",
			req: &ApproveGradingSetup{
				ApproveGrading: false,
				Role:           constants.RoleStudent,
				IsAssigned:     false,
				Status:         sspb.SubmissionStatus_SUBMISSION_STATUS_NOT_MARKED.String(),
				StatusChange:   sspb.SubmissionStatus_SUBMISSION_STATUS_IN_PROGRESS.String(),
			},
			expectedErr: fmt.Errorf("role %s does not allow change information", constants.RoleStudent),
		},
		{
			name: "happy case - Teacher - Assigned - Approve Grading Off - Not Marked to Any",
			req: &ApproveGradingSetup{
				ApproveGrading: false,
				Role:           constants.RoleTeacher,
				IsAssigned:     true,
				Status:         sspb.SubmissionStatus_SUBMISSION_STATUS_NOT_MARKED.String(),
				StatusChange:   sspb.SubmissionStatus_SUBMISSION_STATUS_RETURNED.String(),
			},
			expectedErr: nil,
		},
		{
			name: "happy case - HQ Staff - Assigned - Approve Grading Off - Not Marked to Any",
			req: &ApproveGradingSetup{
				ApproveGrading: false,
				Role:           constants.RoleHQStaff,
				IsAssigned:     true,
				Status:         sspb.SubmissionStatus_SUBMISSION_STATUS_NOT_MARKED.String(),
				StatusChange:   sspb.SubmissionStatus_SUBMISSION_STATUS_RETURNED.String(),
			},
			expectedErr: nil,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			err := validateStatusChangeApproveGradingSetup(testCase.req.(*ApproveGradingSetup))
			assert.Equal(t, testCase.expectedErr, err)
		})
	}
}

func Test_BulkApproveRejectSubmission(t *testing.T) {
	mockDB := &mock_database.Ext{}
	mockTx := &mock_database.Tx{}
	examLOSubmissionRepo := &mock_repositories.MockExamLOSubmissionRepo{}
	examLORepo := &mock_repositories.MockExamLORepo{}

	svc := &ExamLOService{
		DB:                   mockDB,
		ExamLOSubmissionRepo: examLOSubmissionRepo,
		ExamLORepo:           examLORepo,
	}

	testCases := []TestCase{
		{
			name: "req must have SubmissionIds",
			req: &sspb.BulkApproveRejectSubmissionRequest{
				ApproveGradingAction: sspb.ApproveGradingAction_APPROVE_ACTION_APPROVED,
				SubmissionIds:        []string{},
			},
			setup:        func(ctx context.Context) {},
			expectedResp: nil,
			expectedErr:  status.Errorf(codes.InvalidArgument, errors.New("req must have SubmissionIds").Error()),
		},
		{
			name: "there are invalid submission IDs",
			req: &sspb.BulkApproveRejectSubmissionRequest{
				ApproveGradingAction: sspb.ApproveGradingAction_APPROVE_ACTION_APPROVED,
				SubmissionIds:        []string{"submission_id_1", "submission_id_2"},
			},
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Return(mockTx, nil)
				mockTx.On("Rollback", ctx).Once().Return(nil)

				examLOSubmissionRepo.On("BulkUpdateApproveReject", ctx, mockTx, mock.Anything).Once().Return(1, nil)
				examLOSubmissionRepo.On("GetInvalidIDsByBulkApproveReject", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(
					database.TextArray([]string{"submission_id_1"}), nil)
			},
			expectedResp: &sspb.BulkApproveRejectSubmissionResponse{
				InvalidSubmissionIds: []string{"submission_id_1"},
			},
			expectedErr: status.Errorf(codes.Internal, errors.New("bulk approved is only allowed which submission has status is Marked").Error()),
		},
		{
			name: "happy case - approve",
			req: &sspb.BulkApproveRejectSubmissionRequest{
				ApproveGradingAction: sspb.ApproveGradingAction_APPROVE_ACTION_APPROVED,
				SubmissionIds:        []string{"submission_id_1", "submission_id_2"},
			},
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Return(mockTx, nil)
				mockTx.On("Commit", mock.Anything).Return(nil)

				examLOSubmissionRepo.On("BulkUpdateApproveReject", ctx, mockTx, mock.Anything).Once().Return(2, nil)
			},
			expectedResp: nil,
		},
		{
			name: "happy case - reject",
			req: &sspb.BulkApproveRejectSubmissionRequest{
				ApproveGradingAction: sspb.ApproveGradingAction_APPROVE_ACTION_REJECTED,
				SubmissionIds:        []string{"submission_id_1", "submission_id_2"},
			},
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Return(mockTx, nil)
				mockTx.On("Commit", mock.Anything).Return(nil)

				examLOSubmissionRepo.On("BulkUpdateApproveReject", ctx, mockTx, mock.Anything).Once().Return(2, nil)
			},
			expectedResp: nil,
		},
	}

	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		_, err := svc.BulkApproveRejectSubmission(ctx, testCase.req.(*sspb.BulkApproveRejectSubmissionRequest))
		if testCase.expectedErr != nil {
			fmt.Println(err.Error())
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}

func Test_RetrieveMetadataTaggingResult(t *testing.T) {
	mockDB := &mock_database.Ext{}
	examLORepo := &mock_repositories.MockExamLORepo{}
	questionTagRepo := &mock_repositories.MockQuestionTagRepo{}
	examLOSubmissionRepo := &mock_repositories.MockExamLOSubmissionRepo{}

	svc := &ExamLOService{
		DB:                   mockDB,
		ExamLORepo:           examLORepo,
		ExamLOSubmissionRepo: examLOSubmissionRepo,
		QuestionTagRepo:      questionTagRepo,
	}

	testCases := []TestCase{
		{
			name:        "happy case",
			req:         &sspb.RetrieveMetadataTaggingResultRequest{SubmissionId: "123"},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				questionTagRepo.On("GetPointPerTagBySubmissionID", ctx, mockDB, mock.Anything).Once().Return([]repositories.GetPointPerTagBySubmissionIDData{}, nil)
			},
		},
		{
			name:        "missing submission id",
			req:         &sspb.RetrieveMetadataTaggingResultRequest{SubmissionId: ""},
			expectedErr: status.Error(codes.InvalidArgument, fmt.Sprintf("validateRetrieveMetadataTaggingResultRequest: %s", fmt.Errorf("missing submission id").Error())),
		},
		{
			name:        "cannot get point per tag",
			req:         &sspb.RetrieveMetadataTaggingResultRequest{SubmissionId: "123"},
			expectedErr: status.Errorf(codes.Internal, "s.QuestionTagRepo.GetPointPerTagBySubmissionID: %s", fmt.Errorf("error")),
			setup: func(ctx context.Context) {
				questionTagRepo.On("GetPointPerTagBySubmissionID", ctx, mockDB, mock.Anything).Once().Return(nil, fmt.Errorf("error"))
			},
		},
	}

	for _, testCase := range testCases {
		ctx := context.Background()
		if testCase.setup != nil {
			testCase.setup(ctx)
		}
		_, err := svc.RetrieveMetadataTaggingResult(ctx, testCase.req.(*sspb.RetrieveMetadataTaggingResultRequest))
		assert.Equal(t, testCase.expectedErr, err)
	}
}
