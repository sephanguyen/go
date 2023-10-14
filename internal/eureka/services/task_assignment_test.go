package services

import (
	"context"
	"fmt"
	"testing"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	mock_repositories "github.com/manabie-com/backend/mock/eureka/repositories"
	mock_services "github.com/manabie-com/backend/mock/eureka/services"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestTaskAssignmentService_InsertTaskAssignment(t *testing.T) {
	ctx := context.Background()
	mockDB := &mock_database.Ext{}
	mockTx := &mock_database.Tx{}

	mockTopicRepo := &mock_repositories.MockTopicRepo{}
	mockTaskAssignmentRepo := &mock_repositories.MockTaskAssignmentRepo{}

	taskAssignmentService := &TaskAssignmentService{
		DB:                 mockDB,
		TopicRepo:          mockTopicRepo,
		TaskAssignmentRepo: mockTaskAssignmentRepo,
	}
	req := &sspb.InsertTaskAssignmentRequest{
		TaskAssignment: &sspb.TaskAssignmentBase{
			Base: &sspb.LearningMaterialBase{
				TopicId: "topic-id-1",
				Name:    "task-assignment-1",
				Type:    sspb.LearningMaterialType_LEARNING_MATERIAL_TASK_ASSIGNMENT.String(),
			},
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
				mockTaskAssignmentRepo.On("Insert", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				mockTopicRepo.On("UpdateLODisplayOrderCounter", mock.Anything, mock.Anything, database.Text("topic-id-1"), database.Int4(1)).Once().Return(nil)
				mockTx.On("Commit", mock.Anything).Return(nil)
			},
			req:          req,
			expectedResp: &sspb.InsertTaskAssignmentResponse{},
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
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			resp, err := taskAssignmentService.InsertTaskAssignment(ctx, testCase.req.(*sspb.InsertTaskAssignmentRequest))
			assert.Equal(t, testCase.expectedErr, err)
			if testCase.expectedResp != nil {
				assert.IsType(t, testCase.expectedResp, resp)
			}
		})
	}
}

func Test_validateInsertTaskAssignmentReq(t *testing.T) {
	testCases := []TestCase{
		{
			name: "LearningMaterialId not empty",
			req: &sspb.TaskAssignmentBase{
				Base: &sspb.LearningMaterialBase{
					LearningMaterialId: "learning_material_id",
				},
			},
			expectedErr: fmt.Errorf("LearningMaterialId must be empty"),
		},
		{
			name: "Empty TopicId",
			req: &sspb.TaskAssignmentBase{
				Base: &sspb.LearningMaterialBase{
					TopicId: "",
				},
			},
			expectedErr: fmt.Errorf("TopicId must not be empty"),
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			err := validateInsertTaskAssignmentReq(testCase.req.(*sspb.TaskAssignmentBase))
			assert.Equal(t, testCase.expectedErr, err)
		})
	}
}

func Test_validateUpdateTaskAssignmentReq(t *testing.T) {
	testCases := []TestCase{
		{
			name: "LearningMaterialId empty",
			req: &sspb.TaskAssignmentBase{
				Base: &sspb.LearningMaterialBase{
					LearningMaterialId: "",
				},
			},
			expectedErr: fmt.Errorf("empty learning_material_id"),
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			err := validateUpdateTaskAssignmentReq(testCase.req.(*sspb.TaskAssignmentBase))
			assert.Equal(t, testCase.expectedErr, err)
		})
	}
}

func TestTaskAssignmentService_UpdateTaskAssignment(t *testing.T) {
	t.Parallel()
	topicRepo := &mock_repositories.MockTopicRepo{}
	taskAssignmentRepo := &mock_repositories.MockTaskAssignmentRepo{}

	mockDB := &mock_database.Ext{}

	taskAssignmentService := &TaskAssignmentService{
		DB:                 mockDB,
		TopicRepo:          topicRepo,
		TaskAssignmentRepo: taskAssignmentRepo,
	}
	validReq := &sspb.UpdateTaskAssignmentRequest{
		TaskAssignment: &sspb.TaskAssignmentBase{
			Base: &sspb.LearningMaterialBase{
				LearningMaterialId: "learning-material-id",
				Name:               "name",
			},
			Attachments:               []string{"attachment-1", "attachment-2"},
			Instruction:               "instruction",
			RequireDuration:           true,
			RequireCompleteDate:       true,
			RequireUnderstandingLevel: true,
			RequireCorrectness:        true,
			RequireAttachment:         false,
			RequireAssignmentNote:     false,
		},
	}
	testCases := []TestCase{
		{
			name: "err validate updateTaskAssignmentRequest",
			req: &sspb.UpdateTaskAssignmentRequest{
				TaskAssignment: &sspb.TaskAssignmentBase{
					Base: &sspb.LearningMaterialBase{},
				},
			},
			expectedErr: status.Errorf(codes.InvalidArgument, fmt.Errorf("validateUpdateTaskAssignmentReq: empty learning_material_id").Error()),
			setup:       func(ctx context.Context) {},
		},
		{
			name: "happy case",
			req:  validReq,
			setup: func(ctx context.Context) {
				taskAssignmentRepo.On("Update", mock.Anything, mockDB, mock.Anything).Once().Return(nil)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			ctx := interceptors.NewIncomingContext(context.Background())
			testCase.setup(ctx)
			rsp, err := taskAssignmentService.UpdateTaskAssignment(ctx, testCase.req.(*sspb.UpdateTaskAssignmentRequest))
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
				assert.Nil(t, rsp, "expecting nil response")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTaskAssignmentService_ListTaskAssignment(t *testing.T) {
	ctx := context.Background()
	mockDB := &mock_database.Ext{}
	mockTaskAssignmentRepo := &mock_repositories.MockTaskAssignmentRepo{}

	var (
		LearningMaterialIds = database.TextArray([]string{"id1", "id2"})
	)

	taskAssignmentService := &TaskAssignmentService{
		DB:                 mockDB,
		TaskAssignmentRepo: mockTaskAssignmentRepo,
	}

	testCases := []TestCase{
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				taskAssginment := []*entities.TaskAssignment{
					{
						LearningMaterial: entities.LearningMaterial{
							ID:           database.Text("id1"),
							TopicID:      database.Text("topic-id"),
							Name:         database.Text("sid"),
							Type:         database.Text(sspb.LearningMaterialType_LEARNING_MATERIAL_TASK_ASSIGNMENT.String()),
							DisplayOrder: database.Int2(1),
						},
						Attachments:               database.TextArray([]string{"attachment-1", "attachment-2"}),
						Instruction:               database.Text("instruction"),
						RequireDuration:           database.Bool(true),
						RequireCompleteDate:       database.Bool(true),
						RequireUnderstandingLevel: database.Bool(false),
						RequireCorrectness:        database.Bool(false),
						RequireAttachment:         database.Bool(false),
						RequireAssignmentNote:     database.Bool(true),
					},

					{
						LearningMaterial: entities.LearningMaterial{
							ID:           database.Text("id2"),
							TopicID:      database.Text("topic-id"),
							Name:         database.Text("sid"),
							Type:         database.Text(sspb.LearningMaterialType_LEARNING_MATERIAL_TASK_ASSIGNMENT.String()),
							DisplayOrder: database.Int2(2),
						},
						Attachments:               database.TextArray([]string{"attachment-1", "attachment-2"}),
						Instruction:               database.Text("instruction"),
						RequireDuration:           database.Bool(true),
						RequireCompleteDate:       database.Bool(true),
						RequireUnderstandingLevel: database.Bool(false),
						RequireCorrectness:        database.Bool(false),
						RequireAttachment:         database.Bool(false),
						RequireAssignmentNote:     database.Bool(true),
					},
				}
				mockTaskAssignmentRepo.On("List", mock.Anything, mock.Anything, LearningMaterialIds).Once().Return(taskAssginment, nil)
			},
			req: &sspb.ListTaskAssignmentRequest{
				LearningMaterialIds: []string{"id1", "id2"},
			},
			expectedErr: nil,
			expectedResp: &sspb.ListTaskAssignmentResponse{
				TaskAssignments: []*sspb.TaskAssignmentBase{
					{
						Base: &sspb.LearningMaterialBase{
							LearningMaterialId: "id1",
							TopicId:            "topic-id",
							Name:               "sid",
							Type:               sspb.LearningMaterialType_LEARNING_MATERIAL_TASK_ASSIGNMENT.String(),
							DisplayOrder: &wrapperspb.Int32Value{
								Value: 1,
							},
						},
						Attachments:               []string{"attachment-1", "attachment-2"},
						Instruction:               "instruction",
						RequireDuration:           true,
						RequireCompleteDate:       true,
						RequireUnderstandingLevel: false,
						RequireCorrectness:        false,
						RequireAttachment:         false,
						RequireAssignmentNote:     true,
					},
					{
						Base: &sspb.LearningMaterialBase{
							LearningMaterialId: "id2",
							TopicId:            "topic-id",
							Name:               "sid",
							Type:               sspb.LearningMaterialType_LEARNING_MATERIAL_TASK_ASSIGNMENT.String(),
							DisplayOrder: &wrapperspb.Int32Value{
								Value: 2,
							},
						},
						Attachments:               []string{"attachment-1", "attachment-2"},
						Instruction:               "instruction",
						RequireDuration:           true,
						RequireCompleteDate:       true,
						RequireUnderstandingLevel: false,
						RequireCorrectness:        false,
						RequireAttachment:         false,
						RequireAssignmentNote:     true,
					},
				},
			},
		},

		{
			name: "Not found",
			setup: func(ctx context.Context) {
				mockTaskAssignmentRepo.On("List", mock.Anything, mock.Anything, LearningMaterialIds).Once().Return(nil, pgx.ErrNoRows)
			},
			expectedErr: status.Errorf(codes.NotFound, "task assignment not found: %v", pgx.ErrNoRows),
			req: &sspb.ListTaskAssignmentRequest{
				LearningMaterialIds: []string{"id1", "id2"},
			},
		},
		{
			name: "invalid request",
			setup: func(ctx context.Context) {
				mockTaskAssignmentRepo.On("List", mock.Anything, mock.Anything, LearningMaterialIds).Once().Return(nil, pgx.ErrNoRows)
			},
			expectedErr: status.Errorf(codes.InvalidArgument, "LearningMaterialIds must not be empty"),
			req: &sspb.ListTaskAssignmentRequest{
				LearningMaterialIds: []string{},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			resp, err := taskAssignmentService.ListTaskAssignment(ctx, testCase.req.(*sspb.ListTaskAssignmentRequest))
			assert.Equal(t, testCase.expectedErr, err)
			if testCase.expectedResp != nil {
				assert.IsType(t, testCase.expectedResp, resp)
				assert.Equal(t, testCase.expectedResp, resp)
			}
		})
	}
}

func TestTaskAssignmentService_UpsertAdhocTaskAssignment(t *testing.T) {
	ctx := interceptors.NewIncomingContext(context.Background())

	mockDB := &mock_database.Ext{}
	tx := &mock_database.Tx{}

	mockTaskAssignmentRepo := new(mock_repositories.MockTaskAssignmentRepo)
	mockBookRepo := new(mock_repositories.MockBookRepo)
	mockChapterRepo := new(mock_repositories.MockChapterRepo)
	mockBookChapterRepo := new(mock_repositories.MockBookChapterRepo)
	mockTopicRepo := new(mock_repositories.MockTopicRepo)
	mockCourseBookRepo := new(mock_repositories.MockCourseBookRepo)
	mockStudyPlanRepo := new(mock_repositories.MockStudyPlanRepo)
	mockStudentStudyPlanRepo := new(mock_repositories.MockStudentStudyPlanRepo)
	mockAssignmentRepo := new(mock_repositories.MockAssignmentRepo)
	mockStudyPlanItemRepo := new(mock_repositories.MockStudyPlanItemRepo)
	mockAssignmentStudyPlanItemRepo := new(mock_repositories.MockAssignmentStudyPlanItemRepo)
	mockLoStudyPlanItemRepo := new(mock_repositories.MockLoStudyPlanItemRepo)
	mockLearningObjectiveRepo := new(mock_repositories.MockLearningObjectiveRepo)
	mockBobStudentReaderService := new(mock_services.BobStudentReaderServiceClient)
	mockIndividualStudyPlanRepo := new(mock_repositories.MockIndividualStudyPlan)
	mockMasterStudyPlanRepo := new(mock_repositories.MockMasterStudyPlanRepo)

	taskAssignmentService := &TaskAssignmentService{
		DB:                          mockDB,
		TaskAssignmentRepo:          mockTaskAssignmentRepo,
		BookRepo:                    mockBookRepo,
		ChapterRepo:                 mockChapterRepo,
		BookChapterRepo:             mockBookChapterRepo,
		TopicRepo:                   mockTopicRepo,
		CourseBookRepo:              mockCourseBookRepo,
		StudyPlanRepo:               mockStudyPlanRepo,
		StudentStudyPlanRepo:        mockStudentStudyPlanRepo,
		AssignmentRepo:              mockAssignmentRepo,
		StudyPlanItemRepo:           mockStudyPlanItemRepo,
		AssignmentStudyPlanItemRepo: mockAssignmentStudyPlanItemRepo,
		LoStudyPlanItemRepo:         mockLoStudyPlanItemRepo,
		LearningObjectiveRepo:       mockLearningObjectiveRepo,
		BobStudentReaderService:     mockBobStudentReaderService,
		IndividualStudyPlanRepo:     mockIndividualStudyPlanRepo,
		MasterStudyPlanRepo:         mockMasterStudyPlanRepo,
	}

	validReq := &sspb.UpsertAdhocTaskAssignmentRequest{
		StudentId: "student-id",
		CourseId:  "course-id",
		StartDate: timestamppb.Now(),
		TaskAssignment: &sspb.TaskAssignmentBase{
			Base: &sspb.LearningMaterialBase{LearningMaterialId: "lm-id"},
		},
	}

	testCases := []TestCase{
		{
			name:  "error validate request",
			setup: func(ctx context.Context) {},
			req: &sspb.UpsertAdhocTaskAssignmentRequest{
				CourseId: "",
			},
			expectedErr: status.Error(codes.InvalidArgument, "s.verifyUpsertAdHocTaskAssignmentRequest: CourseId must not be empty"),
		},
		{
			name: "error retrieve student profile",
			setup: func(ctx context.Context) {
				mockBobStudentReaderService.On("RetrieveStudentProfile", mock.Anything, mock.Anything).Once().Return(nil, fmt.Errorf("bob error"))
			},
			req:         validReq,
			expectedErr: status.Error(codes.Internal, "s.RetrieveStudentProfile: s.BobStudentReaderService.RetrieveStudentProfile: bob error"),
		},
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				mockBobStudentReaderService.On("RetrieveStudentProfile", mock.Anything, mock.Anything).Once().Return(&bpb.RetrieveStudentProfileResponse{
					Items: []*bpb.RetrieveStudentProfileResponse_Data{{
						Profile: &bpb.StudentProfile{
							Id:      "student-id",
							Name:    "Alice",
							Grade:   "Lớp 1",
							Country: cpb.Country_COUNTRY_VN,
							School:  &bpb.School{Id: constants.ManabieSchool}},
					}},
				}, nil)
				mockDB.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
				tx.On("Commit", mock.Anything).Return(nil)
				tx.On("Begin", mock.Anything).Return(tx, nil)
				mockBookRepo.On("RetrieveAdHocBookByCourseIDAndStudentID", mock.Anything, mock.Anything, database.Text("course-id"), database.Text("student-id")).
					Once().Return(&entities.Book{ID: database.Text("book-id")}, nil)
				mockTopicRepo.On("FindByBookIDs", mock.Anything, mock.Anything,
					database.TextArray([]string{"book-id"}),
					pgtype.TextArray{Status: pgtype.Null},
					pgtype.Int4{Status: pgtype.Null},
					pgtype.Int4{Status: pgtype.Null},
				).Once().Return([]*entities.Topic{{ID: database.Text("topic-id")}}, nil)
				mockCourseBookRepo.On("Upsert", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)

				mockCourseBookRepo.On("FindByCourseIDAndBookID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, nil)
				mockBookRepo.On("FindByID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(&entities.Book{
					ID:       database.Text("book-id"),
					BookType: database.Text(cpb.BookType_BOOK_TYPE_ADHOC.String()),
				}, nil)
				mockStudyPlanRepo.On("BulkUpsert", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				mockStudentStudyPlanRepo.On("BulkUpsert", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				mockBookRepo.On("RetrieveBookTreeByBookID", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, nil)
				mockAssignmentRepo.On("RetrieveAssignmentsByTopicIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, nil)
				mockLearningObjectiveRepo.On("RetrieveLearningObjectivesByTopicIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, nil)
				mockTaskAssignmentRepo.On("Upsert", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				mockIndividualStudyPlanRepo.On("BulkSync", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, nil)
				mockMasterStudyPlanRepo.On("BulkUpsert", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
			},
			req: validReq,
			expectedResp: &sspb.UpsertAdhocTaskAssignmentResponse{
				LearningMaterialId: "lm-id",
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			resp, err := taskAssignmentService.UpsertAdhocTaskAssignment(ctx, testCase.req.(*sspb.UpsertAdhocTaskAssignmentRequest))
			assert.Equal(t, testCase.expectedErr, err)
			if testCase.expectedResp != nil {
				assert.IsType(t, testCase.expectedResp, resp)
				assert.Equal(t, testCase.expectedResp, resp)
			}
		})
	}
}

func TestTaskAssignmentService_verifyUpsertAdHocTaskAssignmentRequest(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	taskAssignmentService := &TaskAssignmentService{}
	testCases := []TestCase{
		{
			name:        "error empty CourseId",
			setup:       func(ctx context.Context) {},
			req:         &sspb.UpsertAdhocTaskAssignmentRequest{},
			expectedErr: fmt.Errorf("CourseId must not be empty"),
		},
		{
			name:  "error empty StudentId",
			setup: func(ctx context.Context) {},
			req: &sspb.UpsertAdhocTaskAssignmentRequest{
				CourseId: "course-id",
			},
			expectedErr: fmt.Errorf("StudentId must not be empty"),
		},
		{
			name:  "error empty StartDate",
			setup: func(ctx context.Context) {},
			req: &sspb.UpsertAdhocTaskAssignmentRequest{
				CourseId:  "course-id",
				StudentId: "student-id",
			},
			expectedErr: fmt.Errorf("StartDate must not be empty"),
		},
		{
			name:  "error empty TaskAssignment",
			setup: func(ctx context.Context) {},
			req: &sspb.UpsertAdhocTaskAssignmentRequest{
				CourseId:  "course-id",
				StudentId: "student-id",
				StartDate: timestamppb.Now(),
			},
			expectedErr: fmt.Errorf("TaskAssignment must not be empty"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			err := taskAssignmentService.verifyUpsertAdHocTaskAssignmentRequest(testCase.req.(*sspb.UpsertAdhocTaskAssignmentRequest))
			assert.Equal(t, testCase.expectedErr, err)
		})
	}
}
