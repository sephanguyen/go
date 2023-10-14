package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/eureka/entities"
	eureka_repo "github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	mock_repositories "github.com/manabie-com/backend/mock/eureka/repositories"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"

	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestFlashcardService_InsertFlashcard(t *testing.T) {
	ctx := context.Background()
	mockDB := &mock_database.Ext{}
	mockTx := &mock_database.Tx{}

	mockTopicRepo := &mock_repositories.MockTopicRepo{}
	mockFlashcardRepo := &mock_repositories.MockFlashcardRepo{}

	flashcardService := &FlashcardService{
		DB:            mockDB,
		TopicRepo:     mockTopicRepo,
		FlashcardRepo: mockFlashcardRepo,
	}
	req := &sspb.InsertFlashcardRequest{
		Flashcard: &sspb.FlashcardBase{
			Base: &sspb.LearningMaterialBase{
				TopicId: "topic-id-1",
				Name:    "flashcard-1",
				Type:    sspb.LearningMaterialType_LEARNING_MATERIAL_FLASH_CARD.String(),
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
				mockFlashcardRepo.On("Insert", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				mockTopicRepo.On("UpdateLODisplayOrderCounter", mock.Anything, mock.Anything, database.Text("topic-id-1"), database.Int4(1)).Once().Return(nil)
				mockTx.On("Commit", mock.Anything).Return(nil)
			},
			req:          req,
			expectedResp: &sspb.InsertFlashcardResponse{},
		},
		{
			name: "happy case - with learnosity vendor type",
			setup: func(ctx context.Context) {
				mockDB.On("Begin", mock.Anything, mock.Anything).Return(mockTx, nil)
				mockTopicRepo.On("RetrieveByID", mock.Anything, mock.Anything, database.Text("topic-id-1"), mock.Anything).Once().Return(&entities.Topic{
					ID:                    database.Text("topic-id-1"),
					LODisplayOrderCounter: database.Int4(0),
				}, nil)
				mockFlashcardRepo.On("Insert", mock.Anything, mock.Anything,
					mock.MatchedBy(func(flashcard *entities.Flashcard) bool {
						return flashcard.VendorType.String == cpb.LearningMaterialVendorType_LM_VENDOR_TYPE_LEARNOSITY.String()
					})).Once().Return(nil)
				mockTopicRepo.On("UpdateLODisplayOrderCounter", mock.Anything, mock.Anything, database.Text("topic-id-1"), database.Int4(1)).Once().Return(nil)
				mockTx.On("Commit", mock.Anything).Return(nil)
			},
			req: &sspb.InsertFlashcardRequest{
				Flashcard: &sspb.FlashcardBase{
					Base: &sspb.LearningMaterialBase{
						TopicId:    "topic-id-1",
						Name:       "flashcard-1",
						Type:       sspb.LearningMaterialType_LEARNING_MATERIAL_FLASH_CARD.String(),
						VendorType: sspb.LearningMaterialVendorType_LM_VENDOR_TYPE_LEARNOSITY,
					},
				},
			},
			expectedResp: &sspb.InsertFlashcardResponse{},
		},
		{
			name: "missing topic_id",
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
			resp, err := flashcardService.InsertFlashcard(ctx, testCase.req.(*sspb.InsertFlashcardRequest))
			assert.Equal(t, testCase.expectedErr, err)
			if testCase.expectedResp != nil {
				assert.IsType(t, testCase.expectedResp, resp)
			}
		})
	}
}

func TestFlashcardService_UpdateFlashcard(t *testing.T) {
	ctx := context.Background()
	mockDB := &mock_database.Ext{}
	mockTx := &mock_database.Tx{}

	mockFlashcardRepo := &mock_repositories.MockFlashcardRepo{}

	flashcardService := &FlashcardService{
		DB:            mockDB,
		FlashcardRepo: mockFlashcardRepo,
	}

	testCases := []TestCase{
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				mockDB.On("Begin", mock.Anything, mock.Anything).Return(mockTx, nil)
				mockFlashcardRepo.On("Update", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", mock.Anything).Return(nil)
			},
			req: &sspb.UpdateFlashcardRequest{
				Flashcard: &sspb.FlashcardBase{
					Base: &sspb.LearningMaterialBase{
						LearningMaterialId: "flashcard-id-1",
						Name:               "flashcard updated",
					},
				},
			},
			expectedResp: &sspb.UpdateFlashcardResponse{},
		},
		{
			name:  "missing LearningMaterialID",
			setup: func(ctx context.Context) {},
			req: &sspb.UpdateFlashcardRequest{
				Flashcard: &sspb.FlashcardBase{
					Base: &sspb.LearningMaterialBase{
						LearningMaterialId: "",
						Name:               "flashcard updated",
					},
				},
			},
			expectedErr: status.Errorf(codes.InvalidArgument, fmt.Errorf("validateFlashcardUpdateReq: LearningMaterialId must not be empty").Error()),
		},
		{
			name: "flashcard not found",
			setup: func(ctx context.Context) {
				mockDB.On("Begin", mock.Anything, mock.Anything).Return(mockTx, nil)
				mockFlashcardRepo.On("Update", mock.Anything, mock.Anything, mock.Anything).Once().Return(pgx.ErrNoRows)
				mockTx.On("Rollback", mock.Anything).Return(nil)
			},
			req: &sspb.UpdateFlashcardRequest{
				Flashcard: &sspb.FlashcardBase{
					Base: &sspb.LearningMaterialBase{
						LearningMaterialId: "flashcard-id-2",
						Name:               "flashcard updated",
					},
				},
			},
			expectedErr: status.Errorf(codes.Internal, fmt.Errorf("database.ExecInTx: s.FlashcardRepo.Update: %w", pgx.ErrNoRows).Error()),
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			resp, err := flashcardService.UpdateFlashcard(ctx, testCase.req.(*sspb.UpdateFlashcardRequest))
			assert.Equal(t, testCase.expectedErr, err)
			if testCase.expectedResp != nil {
				assert.IsType(t, testCase.expectedResp, resp)
			}
		})
	}
}

func TestFlashcardService_ListFlashcard(t *testing.T) {
	ctx := context.Background()
	mockDB := &mock_database.Ext{}
	mockFlashcardRepo := &mock_repositories.MockFlashcardRepo{}
	var (
		validateListFlashcardReqErr string = "LearningMaterialIds must greater than 0"
		notFoundErr                 string = "cannot find flashcards"
	)
	flashcardService := &FlashcardService{
		DB:            mockDB,
		FlashcardRepo: mockFlashcardRepo,
	}
	testCases := []TestCase{
		{
			name:         "invalid arguments",
			setup:        func(ctx context.Context) {},
			expectedErr:  status.Error(codes.InvalidArgument, fmt.Errorf("validateListFlashcardReq: %s", validateListFlashcardReqErr).Error()),
			expectedResp: (*sspb.ListFlashcardResponse)(nil),
			req: &sspb.ListFlashcardRequest{
				LearningMaterialIds: []string{},
			},
		},
		{
			name: "flashcards not found",
			setup: func(ctx context.Context) {
				mockFlashcardRepo.On("ListFlashcardBase", mock.Anything, mock.Anything, &eureka_repo.ListFlashcardArgs{
					LearningMaterialIDs: database.TextArray([]string{"fake-id1", "fake-id2"}),
				}).Once().Return(nil, pgx.ErrNoRows)
			},
			expectedErr:  status.Error(codes.NotFound, notFoundErr),
			expectedResp: (*sspb.ListFlashcardResponse)(nil),
			req: &sspb.ListFlashcardRequest{
				LearningMaterialIds: []string{"fake-id1", "fake-id2"},
			},
		},
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				flashcards := []*entities.FlashcardBase{
					{
						Flashcard: entities.Flashcard{
							LearningMaterial: entities.LearningMaterial{
								ID:           database.Text("id1"),
								TopicID:      database.Text("topic-id"),
								Name:         database.Text("sid"),
								Type:         database.Text(sspb.LearningMaterialType_LEARNING_MATERIAL_FLASH_CARD.String()),
								DisplayOrder: database.Int2(1),
							},
						},
						TotalQuestion: database.Int4(1),
					},
					{
						Flashcard: entities.Flashcard{
							LearningMaterial: entities.LearningMaterial{
								ID:           database.Text("id2"),
								TopicID:      database.Text("topic-id"),
								Name:         database.Text("sid"),
								Type:         database.Text(sspb.LearningMaterialType_LEARNING_MATERIAL_FLASH_CARD.String()),
								DisplayOrder: database.Int2(2),
							},
						},
						TotalQuestion: database.Int4(2),
					},
				}
				mockFlashcardRepo.On("ListFlashcardBase", mock.Anything, mock.Anything, &eureka_repo.ListFlashcardArgs{
					LearningMaterialIDs: database.TextArray([]string{"id1", "id2"}),
				}).Once().Return(flashcards, nil)
			},
			req: &sspb.ListFlashcardRequest{
				LearningMaterialIds: []string{"id1", "id2"},
			},
			expectedErr: nil,
			expectedResp: &sspb.ListFlashcardResponse{
				Flashcards: []*sspb.FlashcardBase{
					{
						Base: &sspb.LearningMaterialBase{
							LearningMaterialId: "id1",
							TopicId:            "topic-id",
							Name:               "sid",
							Type:               sspb.LearningMaterialType_LEARNING_MATERIAL_FLASH_CARD.String(),
							DisplayOrder: &wrapperspb.Int32Value{
								Value: 1,
							},
						},
						TotalQuestion: int32(1),
					},
					{
						Base: &sspb.LearningMaterialBase{
							LearningMaterialId: "id2",
							TopicId:            "topic-id",
							Name:               "sid",
							Type:               sspb.LearningMaterialType_LEARNING_MATERIAL_FLASH_CARD.String(),
							DisplayOrder: &wrapperspb.Int32Value{
								Value: 2,
							},
						},
						TotalQuestion: int32(2),
					},
				},
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			resp, err := flashcardService.ListFlashcard(ctx, testCase.req.(*sspb.ListFlashcardRequest))
			assert.Equal(t, testCase.expectedErr, err)
			assert.Equal(t, testCase.expectedResp.(*sspb.ListFlashcardResponse), resp)

		})
	}
}

func TestFlashcardService_CreateFlashCardStudy(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	db := &mock_database.Ext{}

	req := &sspb.CreateFlashCardStudyRequest{
		StudyPlanId: "study-plan-id",
		LmId:        "lm-id",
		StudentId:   "student-id",
		KeepOrder:   false,
		Paging: &cpb.Paging{
			Limit: 10,
			Offset: &cpb.Paging_OffsetInteger{
				OffsetInteger: 1,
			},
		},
	}
	quizSetRepo := new(mock_repositories.MockQuizSetRepo)
	quizRepo := new(mock_repositories.MockQuizRepo)
	flashcardProgressionRepo := new(mock_repositories.MockFlashcardProgressionRepo)
	s := FlashcardService{
		DB:                       db,
		QuizRepo:                 quizRepo,
		QuizSetRepo:              quizSetRepo,
		FlashcardProgressionRepo: flashcardProgressionRepo,
	}

	testCases := []TestCase{
		{
			name:        "happy case",
			ctx:         ctx,
			req:         req,
			expectedErr: nil,
			setup: func(ctx context.Context) {
				quizSetRepo.On("GetQuizSetByLoID", ctx, db, mock.Anything, mock.Anything).Once().Return(&entities.QuizSet{ID: database.Text("quiz-set-id")}, nil)
				flashcardProgressionRepo.On("Create", ctx, db, mock.Anything).Once().Return(database.Text("study-set-id"), nil)
				flashcardProgressionRepo.On("Get", ctx, db, mock.Anything).Once().Return(&entities.FlashcardProgression{}, nil)
				quizRepo.On("GetByExternalIDsAndLmID", ctx, db, mock.Anything, mock.Anything).Once().Return(entities.Quizzes{}, nil)
			},
		},
		{
			name:        "err quizSetRepo.GetQuizSetByLoID",
			ctx:         ctx,
			req:         req,
			expectedErr: status.Errorf(codes.Internal, fmt.Errorf("s.QuizSetRepo.GetQuizSetByLoID: %v", ErrSomethingWentWrong).Error()),
			setup: func(ctx context.Context) {
				quizSetRepo.On("GetQuizSetByLoID", ctx, db, mock.Anything, mock.Anything).Once().Return(nil, ErrSomethingWentWrong)
			},
		},
		{
			name:        "err flashcardProgressionRepo.Create",
			ctx:         ctx,
			req:         req,
			expectedErr: status.Errorf(codes.Internal, fmt.Errorf("s.FlashcardProgressionRepo.Create: %v", ErrSomethingWentWrong).Error()),
			setup: func(ctx context.Context) {
				quizSetRepo.On("GetQuizSetByLoID", ctx, db, mock.Anything, mock.Anything).Once().Return(&entities.QuizSet{ID: database.Text("quiz-set-id")}, nil)
				flashcardProgressionRepo.On("Create", ctx, db, mock.Anything).Once().Return(database.Text("study-set-id"), ErrSomethingWentWrong)
			},
		},
		{
			name:        "err flashcardProgressionRepo.Get",
			ctx:         ctx,
			req:         req,
			expectedErr: status.Errorf(codes.Internal, "s.getFlashcardProgressionWithPaging: s.FlashcardProgressionRepo.Get: %v", ErrSomethingWentWrong),
			setup: func(ctx context.Context) {
				quizSetRepo.On("GetQuizSetByLoID", ctx, db, mock.Anything, mock.Anything).Once().Return(&entities.QuizSet{ID: database.Text("quiz-set-id")}, nil)
				flashcardProgressionRepo.On("Create", ctx, db, mock.Anything).Once().Return(database.Text("study-set-id"), nil)
				flashcardProgressionRepo.On("Get", ctx, db, mock.Anything).Once().Return(nil, ErrSomethingWentWrong)
			},
		},
		{
			name:        "err quizRepo.GetByExternalIDsAndLmID",
			ctx:         ctx,
			req:         req,
			expectedErr: status.Errorf(codes.Internal, "s.getFlashcardProgressionWithPaging: s.QuizRepo.GetByExternalIDsAndLmID: %v", ErrSomethingWentWrong),
			setup: func(ctx context.Context) {
				quizSetRepo.On("GetQuizSetByLoID", ctx, db, mock.Anything, mock.Anything).Once().Return(&entities.QuizSet{ID: database.Text("quiz-set-id")}, nil)
				flashcardProgressionRepo.On("Create", ctx, db, mock.Anything).Once().Return(database.Text("study-set-id"), nil)
				flashcardProgressionRepo.On("Get", ctx, db, mock.Anything).Once().Return(&entities.FlashcardProgression{}, nil)
				quizRepo.On("GetByExternalIDsAndLmID", ctx, db, mock.Anything, mock.Anything).Once().Return(entities.Quizzes{}, ErrSomethingWentWrong)
			},
		},
	}
	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			_, err := s.CreateFlashCardStudy(testCase.ctx, testCase.req.(*sspb.CreateFlashCardStudyRequest))
			assert.Equal(t, testCase.expectedErr, err)
		})

	}
}

func Test_ToBaseFlashcard(t *testing.T) {
	testCases := []TestCase{
		{
			name: "happy case",
			req: &entities.Flashcard{
				LearningMaterial: entities.LearningMaterial{
					ID:           database.Text("flashcard-id"),
					Name:         database.Text("flashcard-name"),
					TopicID:      database.Text("flashcard-topic-id"),
					Type:         database.Text(sspb.LearningMaterialType_LEARNING_MATERIAL_FLASH_CARD.String()),
					DisplayOrder: database.Int2(0),
				},
			},
			expectedResp: &sspb.FlashcardBase{
				Base: &sspb.LearningMaterialBase{
					LearningMaterialId: "flashcard-id",
					TopicId:            "flashcard-topic-id",
					Name:               "flashcard-name",
					Type:               sspb.LearningMaterialType_LEARNING_MATERIAL_FLASH_CARD.String(),
					DisplayOrder: &wrapperspb.Int32Value{
						Value: int32(0),
					},
				},
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			res := ToBaseFlashcard(testCase.req.(*entities.Flashcard))
			assert.Equal(t, testCase.expectedResp, res)
		})
	}
}

func TestFlashcardService_validateFinishFlashCardStudyRequest(t *testing.T) {
	t.Parallel()
	service := FlashcardService{}
	testCases := []TestCase{
		{
			name: "error missing StudySetId",
			req: &sspb.FinishFlashCardStudyRequest{
				StudySetId: "",
			},
			expectedErr: fmt.Errorf("StudySetId must not be empty"),
		},
		{
			name: "error missing StudyPlanItemIdentity",
			req: &sspb.FinishFlashCardStudyRequest{
				StudySetId:            "study-set-id",
				StudyPlanItemIdentity: nil,
			},
			expectedErr: fmt.Errorf("StudyPlanItemIdentity must not be empty"),
		},
		{
			name: "error missing StudyPlanId",
			req: &sspb.FinishFlashCardStudyRequest{
				StudySetId: "study-set-id",
				StudyPlanItemIdentity: &sspb.StudyPlanItemIdentity{
					StudyPlanId: "",
				},
			},
			expectedErr: fmt.Errorf("StudyPlanId must not be empty"),
		},
		{
			name: "error missing LearningMaterialId",
			req: &sspb.FinishFlashCardStudyRequest{
				StudySetId: "study-set-id",
				StudyPlanItemIdentity: &sspb.StudyPlanItemIdentity{
					StudyPlanId:        "study-plan-id",
					LearningMaterialId: "",
				},
			},
			expectedErr: fmt.Errorf("LearningMaterialId must not be empty"),
		},
		{
			name: "error missing StudentId",
			req: &sspb.FinishFlashCardStudyRequest{
				StudySetId: "study-set-id",
				StudyPlanItemIdentity: &sspb.StudyPlanItemIdentity{
					StudyPlanId:        "study-plan-id",
					LearningMaterialId: "lm-id",
					StudentId:          nil,
				},
			},
			expectedErr: fmt.Errorf("StudentId must not be empty"),
		},
		{
			name: "error empty StudentId",
			req: &sspb.FinishFlashCardStudyRequest{
				StudySetId: "study-set-id",
				StudyPlanItemIdentity: &sspb.StudyPlanItemIdentity{
					StudyPlanId:        "study-plan-id",
					LearningMaterialId: "lm-id",
					StudentId:          wrapperspb.String(""),
				},
			},
			expectedErr: fmt.Errorf("StudentId must not be empty"),
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			err := service.validateFinishFlashCardStudyRequest(testCase.req.(*sspb.FinishFlashCardStudyRequest))
			assert.Equal(t, testCase.expectedErr, err)
		})
	}
}

func TestFlashcardService_FinishFlashCardStudyProgress(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockDB := &mock_database.Ext{}
	mockTx := &mock_database.Tx{}

	mockFlashcardProgressionRepo := new(mock_repositories.MockFlashcardProgressionRepo)
	service := FlashcardService{
		DB:                       mockDB,
		FlashcardProgressionRepo: mockFlashcardProgressionRepo,
	}

	testCases := []TestCase{
		{
			name:        "error invalid request",
			setup:       func(ctx context.Context) {},
			req:         &sspb.FinishFlashCardStudyRequest{},
			expectedErr: status.Error(codes.InvalidArgument, "s.ValidateFinishFlashCardStudyRequest: StudySetId must not be empty"),
		},
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				mockFlashcardProgressionRepo.On("Get", mock.Anything, mock.Anything, mock.Anything).Once().Return(&entities.FlashcardProgression{
					CompletedAt: database.Timestamptz(time.Time{}),
				}, nil)
				mockDB.On("Begin", mock.Anything, mock.Anything).Once().Return(mockTx, nil)
				mockFlashcardProgressionRepo.On("UpdateCompletedAt", mock.Anything, mock.Anything, database.Text("study-set-id")).Once().Return(nil)
				mockTx.On("Commit", mock.Anything).Once().Return(nil)
			},
			req: &sspb.FinishFlashCardStudyRequest{
				StudyPlanItemIdentity: &sspb.StudyPlanItemIdentity{
					StudyPlanId:        "study-plan-id",
					LearningMaterialId: "learning-material-id",
					StudentId:          wrapperspb.String("student-id"),
				},
				StudySetId: "study-set-id",
			},
			expectedResp: &sspb.FinishFlashCardStudyResponse{},
		},
		{
			name: "happy case restart",
			setup: func(ctx context.Context) {
				mockFlashcardProgressionRepo.On("Get", mock.Anything, mock.Anything, mock.Anything).Once().Return(&entities.FlashcardProgression{
					CompletedAt: database.Timestamptz(time.Time{}),
				}, nil)
				mockDB.On("Begin", mock.Anything, mock.Anything).Once().Return(mockTx, nil)
				mockFlashcardProgressionRepo.On("DeleteByStudySetID", mock.Anything, mock.Anything, database.Text("study-set-id")).Once().Return(nil)
				mockTx.On("Commit", mock.Anything).Once().Return(nil)
			},
			req: &sspb.FinishFlashCardStudyRequest{
				StudyPlanItemIdentity: &sspb.StudyPlanItemIdentity{
					StudyPlanId:        "study-plan-id",
					LearningMaterialId: "learning-material-id",
					StudentId:          wrapperspb.String("student-id"),
				},
				StudySetId: "study-set-id",
				IsRestart:  true,
			},
			expectedResp: &sspb.FinishFlashCardStudyResponse{},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			resp, err := service.FinishFlashCardStudy(ctx, testCase.req.(*sspb.FinishFlashCardStudyRequest))
			assert.Equal(t, testCase.expectedErr, err)
			if testCase.expectedResp != nil {
				assert.Equal(t, testCase.expectedResp.(*sspb.FinishFlashCardStudyResponse), resp)
			}
		})
	}
}
