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
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestLearningObjectiveModifierService_toLearningObjectiveEnt(t *testing.T) {
	t.Parallel()

	testCases := []TestCase{
		{
			name:         "set vendor type",
			req:          sspb.LearningMaterialVendorType_LM_VENDOR_TYPE_MANABIE,
			expectedResp: sspb.LearningMaterialVendorType_LM_VENDOR_TYPE_MANABIE.String(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			req := &sspb.InsertLearningObjectiveRequest{
				LearningObjective: &sspb.LearningObjectiveBase{
					Base: &sspb.LearningMaterialBase{
						TopicId:    "topic-id-1",
						Name:       "learningObjective-1",
						Type:       sspb.LearningMaterialType_LEARNING_MATERIAL_FLASH_CARD.String(),
						VendorType: testCase.req.(sspb.LearningMaterialVendorType),
					},
				},
			}

			e, err := toInsertLearningObjectiveEnt(req.LearningObjective)

			assert.NoError(t, err)
			assert.Equal(t, testCase.expectedResp, e.VendorType.String)

			// Default values when insert learning objective
			assert.Equal(t, e.IsPublished.Bool, false)
			assert.Equal(t, e.Type.String, sspb.LearningMaterialType_LEARNING_MATERIAL_LEARNING_OBJECTIVE.String())
		})
	}

}

func TestLearningObjectiveModifierService_toUpdateLearningObjectiveEnt(t *testing.T) {
	t.Parallel()

	t.Run("Happy case", func(t *testing.T) {
		req := &sspb.UpdateLearningObjectiveRequest{
			LearningObjective: &sspb.LearningObjectiveBase{
				Base: &sspb.LearningMaterialBase{
					LearningMaterialId: "LM_ID",
					Name:               "learningObjective-1",
					Type:               sspb.LearningMaterialType_LEARNING_MATERIAL_FLASH_CARD.String(),
					VendorType:         sspb.LearningMaterialVendorType_LM_VENDOR_TYPE_MANABIE,
				},
				VideoId:    "video_id_1",
				StudyGuide: "study_guide_1",
			},
		}

		e, err := toUpdateLearningObjectiveEnt(req.LearningObjective)

		assert.NoError(t, err)

		// Updated fields
		assert.Equal(t, e.ID.String, "LM_ID")
		assert.Equal(t, e.Name.String, "learningObjective-1")
		assert.Equal(t, e.Video.String, "video_id_1")
		assert.Equal(t, e.StudyGuide.String, "study_guide_1")

		// Don't update these fields
		assert.Equal(t, e.Type.Status, pgtype.Null)
		assert.Equal(t, e.IsPublished.Status, pgtype.Null)
	})

}

func TestLearningObjectiveModifierService_InsertLearningObjective(t *testing.T) {
	ctx := context.Background()
	mockDB := &mock_database.Ext{}
	mockTx := &mock_database.Tx{}

	mockTopicRepo := &mock_repositories.MockTopicRepo{}
	mockLearningObjectiveRepo := &mock_repositories.MockLearningObjectiveRepoV2{}

	learningObjectiveService := &LearningObjectiveService{
		DB:                    mockDB,
		TopicRepo:             mockTopicRepo,
		LearningObjectiveRepo: mockLearningObjectiveRepo,
	}
	req := &sspb.InsertLearningObjectiveRequest{
		LearningObjective: &sspb.LearningObjectiveBase{
			Base: &sspb.LearningMaterialBase{
				TopicId:    "topic-id-1",
				Name:       "learningObjective-1",
				Type:       sspb.LearningMaterialType_LEARNING_MATERIAL_FLASH_CARD.String(),
				VendorType: sspb.LearningMaterialVendorType_LM_VENDOR_TYPE_LEARNOSITY,
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
				mockLearningObjectiveRepo.On("Insert", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				mockTopicRepo.On("UpdateLODisplayOrderCounter", mock.Anything, mock.Anything, database.Text("topic-id-1"), database.Int4(1)).Once().Return(nil)
				mockTx.On("Commit", mock.Anything).Return(nil)
			},
			req:          req,
			expectedResp: &sspb.InsertLearningObjectiveResponse{},
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
			resp, err := learningObjectiveService.InsertLearningObjective(ctx, testCase.req.(*sspb.InsertLearningObjectiveRequest))
			assert.Equal(t, testCase.expectedErr, err)
			if testCase.expectedResp != nil {
				assert.IsType(t, testCase.expectedResp, resp)
			}
		})
	}
}

func TestLearningObjectiveModifierService_UpdateLearningObjective(t *testing.T) {
	ctx := context.Background()
	mockDB := &mock_database.Ext{}
	mockTx := &mock_database.Tx{}

	mockLearningObjectiveRepo := &mock_repositories.MockLearningObjectiveRepoV2{}

	learningObjectiveService := &LearningObjectiveService{
		DB:                    mockDB,
		LearningObjectiveRepo: mockLearningObjectiveRepo,
	}

	testCases := []TestCase{
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				mockDB.On("Begin", mock.Anything, mock.Anything).Return(mockTx, nil)
				mockLearningObjectiveRepo.On("Update", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", mock.Anything).Return(nil)
			},
			req: &sspb.UpdateLearningObjectiveRequest{
				LearningObjective: &sspb.LearningObjectiveBase{
					Base: &sspb.LearningMaterialBase{
						LearningMaterialId: "learning-objective-id-1",
						Name:               "learning-objective updated",
					},
				},
			},
			expectedResp: &sspb.UpdateLearningObjectiveResponse{},
		},
		{
			name:  "missing LearningMaterialID",
			setup: func(ctx context.Context) {},
			req: &sspb.UpdateLearningObjectiveRequest{
				LearningObjective: &sspb.LearningObjectiveBase{
					Base: &sspb.LearningMaterialBase{
						LearningMaterialId: "",
						Name:               "learning-objective updated",
					},
				},
			},
			expectedErr: status.Errorf(codes.InvalidArgument, fmt.Errorf("validateLearningObjectiveUpdateReq: LearningMaterialId must not be empty").Error()),
		},
		{
			name: "learning objective not found",
			setup: func(ctx context.Context) {
				mockDB.On("Begin", mock.Anything, mock.Anything).Return(mockTx, nil)
				mockLearningObjectiveRepo.On("Update", mock.Anything, mock.Anything, mock.Anything).Once().Return(pgx.ErrNoRows)
				mockTx.On("Rollback", mock.Anything).Return(nil)
			},
			req: &sspb.UpdateLearningObjectiveRequest{
				LearningObjective: &sspb.LearningObjectiveBase{
					Base: &sspb.LearningMaterialBase{
						LearningMaterialId: "learning-objective-id-2",
						Name:               "learning-objective updated",
					},
				},
			},
			expectedErr: status.Errorf(codes.Internal, fmt.Errorf("database.ExecInTx: s.LearningObjectiveRepo.Update: %w", pgx.ErrNoRows).Error()),
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			resp, err := learningObjectiveService.UpdateLearningObjective(ctx, testCase.req.(*sspb.UpdateLearningObjectiveRequest))
			assert.Equal(t, testCase.expectedErr, err)
			if testCase.expectedResp != nil {
				assert.IsType(t, testCase.expectedResp, resp)
			}
		})
	}
}

func TestLearningObjectiveModifierService_ListLearningObjective(t *testing.T) {
	ctx := context.Background()
	mockDB := &mock_database.Ext{}
	mockTx := &mock_database.Tx{}

	mockLearningObjectiveRepo := &mock_repositories.MockLearningObjectiveRepoV2{}

	learningObjectiveService := &LearningObjectiveService{
		DB:                    mockDB,
		LearningObjectiveRepo: mockLearningObjectiveRepo,
	}

	testCases := []TestCase{
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				mockDB.On("Begin", mock.Anything, mock.Anything).Return(mockTx, nil)
				mockLearningObjectiveRepo.On("ListLOBaseByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return([]*entities.LearningObjectiveBaseV2{
					{
						LearningObjectiveV2: entities.LearningObjectiveV2{
							LearningMaterial: entities.LearningMaterial{
								ID:           database.Text("learning-objective-id-1"),
								Name:         database.Text("learning-objective-1"),
								TopicID:      database.Text("topic-id-1"),
								Type:         database.Text("learning-objective"),
								DisplayOrder: database.Int2(1),
							},
							Video:         database.Text("video-1"),
							StudyGuide:    database.Text("study-guide-1"),
							VideoScript:   database.Text("video-script-1"),
							ManualGrading: database.Bool(true),
						},
						TotalQuestion: database.Int4(1),
					},
					{
						LearningObjectiveV2: entities.LearningObjectiveV2{
							LearningMaterial: entities.LearningMaterial{
								ID:           database.Text("learning-objective-id-2"),
								Name:         database.Text("learning-objective-2"),
								TopicID:      database.Text("topic-id-1"),
								Type:         database.Text("learning-objective"),
								DisplayOrder: database.Int2(2),
							},
							Video:         database.Text("video-2"),
							StudyGuide:    database.Text("study-guide-2"),
							VideoScript:   database.Text("video-script-2"),
							ManualGrading: database.Bool(false),
						},
						TotalQuestion: database.Int4(2),
					},
				}, nil)
				mockTx.On("Commit", mock.Anything).Return(nil)
			},
			req: &sspb.ListLearningObjectiveRequest{
				LearningMaterialIds: []string{"learning-objective-id-1", "learning-objective-id-2"},
			},
			expectedResp: &sspb.ListLearningObjectiveResponse{
				LearningObjectives: []*sspb.LearningObjectiveBase{
					{
						Base: &sspb.LearningMaterialBase{
							LearningMaterialId: "learning-objective-id-1",
							Name:               "learning-objective-1",
							TopicId:            "topic-id-1",
							Type:               "learning-objective",
							DisplayOrder:       wrapperspb.Int32(int32(1)),
						},
						VideoId:       "video-1",
						StudyGuide:    "study-guide-1",
						VideoScript:   "video-script-1",
						TotalQuestion: int32(1),
						ManualGrading: true,
					},
					{
						Base: &sspb.LearningMaterialBase{
							LearningMaterialId: "learning-objective-id-2",
							Name:               "learning-objective-2",
							TopicId:            "topic-id-1",
							Type:               "learning-objective",
							DisplayOrder:       wrapperspb.Int32(int32(2)),
						},
						VideoId:       "video-2",
						StudyGuide:    "study-guide-2",
						VideoScript:   "video-script-2",
						TotalQuestion: int32(2),
						ManualGrading: false,
					},
				},
			},
		},
		{
			name: "not found",
			setup: func(ctx context.Context) {
				mockDB.On("Begin", mock.Anything, mock.Anything).Return(mockTx, nil)
				mockLearningObjectiveRepo.On("ListLOBaseByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
				mockTx.On("Rollback", mock.Anything).Return(nil)
			},
			req: &sspb.ListLearningObjectiveRequest{
				LearningMaterialIds: []string{"learning-objective-id-1", "learning-objective-id-2"},
			},
			expectedErr: status.Errorf(codes.NotFound, fmt.Errorf("s.LearningObjectiveRepo.ListLOBaseByIDs: %w", pgx.ErrNoRows).Error()),
		},
		{
			name: "missing learning material ids",
			setup: func(ctx context.Context) {
				mockDB.On("Begin", mock.Anything, mock.Anything).Return(mockTx, nil)
				mockLearningObjectiveRepo.On("ListLOBaseByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
				mockTx.On("Rollback", mock.Anything).Return(nil)
			},
			req: &sspb.ListLearningObjectiveRequest{
				LearningMaterialIds: []string{},
			},
			expectedErr: status.Errorf(codes.InvalidArgument, "LearningMaterialIds must not be empty"),
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			resp, err := learningObjectiveService.ListLearningObjective(ctx, testCase.req.(*sspb.ListLearningObjectiveRequest))
			assert.Equal(t, testCase.expectedErr, err)
			if testCase.expectedResp != nil {
				assert.IsType(t, testCase.expectedResp, resp)
				assert.Equal(t, testCase.expectedResp, resp)
			}
		})
	}
}

func Test_ToBaseLearningObjective(t *testing.T) {
	testCases := []TestCase{
		{
			name: "happy case",
			req: &entities.LearningObjectiveV2{
				LearningMaterial: entities.LearningMaterial{
					ID:           database.Text("learning-objective-id"),
					Name:         database.Text("learning-objective-name"),
					TopicID:      database.Text("learning-objective-topic-id"),
					Type:         database.Text(sspb.LearningMaterialType_LEARNING_MATERIAL_LEARNING_OBJECTIVE.String()),
					DisplayOrder: database.Int2(0),
				},
			},
			expectedResp: &sspb.LearningObjectiveBase{
				Base: &sspb.LearningMaterialBase{
					LearningMaterialId: "learning-objective-id",
					TopicId:            "learning-objective-topic-id",
					Name:               "learning-objective-name",
					Type:               sspb.LearningMaterialType_LEARNING_MATERIAL_LEARNING_OBJECTIVE.String(),
					DisplayOrder: &wrapperspb.Int32Value{
						Value: int32(0),
					},
				},
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			res := ToBaseLearningObjectiveV2(testCase.req.(*entities.LearningObjectiveV2))
			assert.Equal(t, testCase.expectedResp, res)
		})
	}
}

func TestLearningObjectiveService_RetrieveLOProgression(t *testing.T) {
	mockDB := &mock_database.Ext{}

	quizRepo := new(mock_repositories.MockQuizRepo)
	lOProgressionRepo := new(mock_repositories.MockLOProgressionRepo)
	lOProgressionAnswerRepo := new(mock_repositories.MockLOProgressionAnswerRepo)
	questionGroupRepo := new(mock_repositories.MockQuestionGroupRepo)
	shuffledQuizSetRepo := new(mock_repositories.MockShuffledQuizSetRepo)

	svc := &LearningObjectiveService{
		DB:                      mockDB,
		QuizRepo:                quizRepo,
		LOProgressionRepo:       lOProgressionRepo,
		LOProgressionAnswerRepo: lOProgressionAnswerRepo,
		QuestionGroupRepo:       questionGroupRepo,
		ShuffledQuizSetRepo:     shuffledQuizSetRepo,
	}

	request := &sspb.RetrieveLOProgressionRequest{
		StudyPlanItemIdentity: &sspb.StudyPlanItemIdentity{
			StudyPlanId: "whatever-1", LearningMaterialId: "whatever-2", StudentId: wrapperspb.String("whatever-3"),
		},
		Paging: &cpb.Paging{Limit: 10},
	}

	progression := &entities.LOProgression{
		ProgressionID:      database.Text("ProgressionID"),
		ShuffledQuizSetID:  database.Text("ShuffledQuizSetID"),
		StudentID:          database.Text("StudentID"),
		StudyPlanID:        database.Text("StudyPlanID"),
		LearningMaterialID: database.Text("LearningMaterialID"),
		LastIndex:          database.Int4(4),
		QuizExternalIDs:    database.TextArray([]string{"External01", "External02", "External03"}),
	}
	progressionAnswers := entities.LOProgressionAnswers{
		{
			ProgressionAnswerID: database.Text("ProgressionAnswerID-01"),
			QuizExternalID:      database.Text("External01"),
			ShuffledQuizSetID:   database.Text("ShuffledQuizSetID"),
			ProgressionID:       database.Text("ProgressionID"),
			StudentIndexAnswers: database.Int4Array([]int32{1, 2}),
			StudentTextAnswers:  database.TextArray(nil),
			SubmittedKeysAnswer: database.TextArray(nil),
		},
		{
			ProgressionAnswerID: database.Text("ProgressionAnswerID-02"),
			QuizExternalID:      database.Text("External02"),
			ShuffledQuizSetID:   database.Text("ShuffledQuizSetID"),
			ProgressionID:       database.Text("ProgressionID"),
			StudentTextAnswers:  database.TextArray([]string{"whatever-1", "whatever-2"}),
			StudentIndexAnswers: database.Int4Array(nil),
			SubmittedKeysAnswer: database.TextArray(nil),
		},
	}
	quizzes := entities.Quizzes{
		{
			ID:              database.Text("Quiz-01"),
			ExternalID:      database.Text("External01"),
			QuestionGroupID: database.Text("QuestionGroupID-01"),
			Question:        database.JSONB("{}"),
			Explanation:     database.JSONB("{}"),
			Options:         database.JSONB("[{}]"),
		},
		{
			ID:              database.Text("Quiz-02"),
			ExternalID:      database.Text("External02"),
			QuestionGroupID: database.Text("QuestionGroupID-02"),
			Question:        database.JSONB("{}"),
			Explanation:     database.JSONB("{}"),
			Options:         database.JSONB("[{}]"),
		},
		{
			ID:          database.Text("Quiz-03"),
			ExternalID:  database.Text("External03"),
			Question:    database.JSONB("{}"),
			Explanation: database.JSONB("{}"),
			Options:     database.JSONB("[{}]"),
		},
	}
	randomSeed := database.Text("1673576588317787428")
	questionGroups := entities.QuestionGroups{
		{
			QuestionGroupID: database.Text("QuestionGroupID-01"),
			Name:            database.Text("01"),
			RichDescription: database.JSONB(&entities.RichText{Raw: "raw rich text 01", RenderedURL: "01"}),
		},
		{
			QuestionGroupID: database.Text("QuestionGroupID-02"),
			Name:            database.Text("02"),
			RichDescription: database.JSONB(&entities.RichText{Raw: "raw rich text 02", RenderedURL: "02"}),
		},
	}

	testCases := []TestCase{
		{
			name: "happy case",
			req:  request,
			setup: func(ctx context.Context) {
				lOProgressionRepo.On("GetByStudyPlanItemIdentity", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(progression, nil)
				lOProgressionAnswerRepo.On("ListByProgressionAndExternalIDs", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(progressionAnswers, nil)
				quizRepo.On("GetByExternalIDsAndLmID", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(quizzes, nil)
				shuffledQuizSetRepo.On("GetSeed", ctx, mock.Anything, mock.Anything).Once().Return(randomSeed, nil)
				questionGroupRepo.On("GetQuestionGroupsByIDs", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(questionGroups, nil)
			},
			expectedResp: &sspb.RetrieveLOProgressionResponse{
				OriginalShuffledQuizSetId: progression.ShuffledQuizSetID.String,
				LastIndex:                 4,
				Items:                     []*sspb.QuizAnswerInfo{{}, {}, {}},
				QuestionGroups:            []*cpb.QuestionGroup{{}, {}},
			},
		},
		{
			name:        "validate",
			req:         &sspb.RetrieveLOProgressionRequest{},
			setup:       func(ctx context.Context) {},
			expectedErr: status.Error(codes.InvalidArgument, "Validate: req must have student id"),
		},
		{
			name: "don't have progression",
			req:  request,
			setup: func(ctx context.Context) {
				lOProgressionRepo.On("GetByStudyPlanItemIdentity", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, fmt.Errorf("no rows"))
			},
			expectedErr: status.Errorf(codes.NotFound, fmt.Errorf("LOProgressionRepo.GetByStudyPlanItemIdentity: %w", fmt.Errorf("no rows")).Error()),
		},
		{
			name: "don't have progression answer",
			req:  request,
			setup: func(ctx context.Context) {
				lOProgressionRepo.On("GetByStudyPlanItemIdentity", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(progression, nil)
				lOProgressionAnswerRepo.On("ListByProgressionAndExternalIDs", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(entities.LOProgressionAnswers{}, nil)
				quizRepo.On("GetByExternalIDsAndLmID", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(quizzes, nil)
				shuffledQuizSetRepo.On("GetSeed", ctx, mock.Anything, mock.Anything).Once().Return(randomSeed, nil)
				questionGroupRepo.On("GetQuestionGroupsByIDs", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(questionGroups, nil)
			},
			expectedResp: &sspb.RetrieveLOProgressionResponse{
				OriginalShuffledQuizSetId: progression.ShuffledQuizSetID.String,
				LastIndex:                 4,
				Items:                     []*sspb.QuizAnswerInfo{{}, {}, {}},
				QuestionGroups:            []*cpb.QuestionGroup{{}, {}},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			ctx := context.Background()
			testCase.setup(ctx)

			res, err := svc.RetrieveLOProgression(ctx, testCase.req.(*sspb.RetrieveLOProgressionRequest))
			assert.Equal(t, testCase.expectedErr, err)
			if exp, ok := testCase.expectedResp.(*sspb.RetrieveLOProgressionResponse); ok {
				assert.Equal(t, exp.GetOriginalShuffledQuizSetId(), res.GetOriginalShuffledQuizSetId())
				assert.Equal(t, exp.GetLastIndex(), res.GetLastIndex())
				assert.Equal(t, len(exp.GetQuestionGroups()), len(res.GetQuestionGroups()))
				assert.Equal(t, len(exp.GetItems()), len(res.GetItems()))
			}
		})
	}
}

func Test_progressionAnswer2QuizAnswer(t *testing.T) {
	testCases := []TestCase{
		{
			name: "selected index",
			req: &entities.LOProgressionAnswer{
				QuizExternalID:      database.Text("external-id"),
				StudentIndexAnswers: database.Int4Array([]int32{1, 2}),
				StudentTextAnswers:  database.TextArray(nil),
				SubmittedKeysAnswer: database.TextArray(nil),
			},
			expectedResp: []*sspb.Answer{
				{Format: &sspb.Answer_SelectedIndex{SelectedIndex: 1}},
				{Format: &sspb.Answer_SelectedIndex{SelectedIndex: 2}},
			},
		},
		{
			name: "filled text",
			req: &entities.LOProgressionAnswer{
				QuizExternalID:      database.Text("external-id"),
				StudentIndexAnswers: database.Int4Array(nil),
				StudentTextAnswers:  database.TextArray([]string{"whatever-1", "whatever-2"}),
				SubmittedKeysAnswer: database.TextArray(nil),
			},
			expectedResp: []*sspb.Answer{
				{Format: &sspb.Answer_FilledText{FilledText: "whatever-1"}},
				{Format: &sspb.Answer_FilledText{FilledText: "whatever-2"}},
			},
		},
		{
			name: "submitted keys",
			req: &entities.LOProgressionAnswer{
				QuizExternalID:      database.Text("external-id"),
				StudentIndexAnswers: database.Int4Array(nil),
				StudentTextAnswers:  database.TextArray(nil),
				SubmittedKeysAnswer: database.TextArray([]string{"whatever-1", "whatever-2"}),
			},
			expectedResp: []*sspb.Answer{
				{Format: &sspb.Answer_SubmittedKey{SubmittedKey: "whatever-1"}},
				{Format: &sspb.Answer_SubmittedKey{SubmittedKey: "whatever-2"}},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			answer, err := progressionAnswer2QuizAnswer(tc.req.(*entities.LOProgressionAnswer))
			assert.Equal(t, tc.expectedErr, err)
			assert.Equal(t, len(tc.expectedResp.([]*sspb.Answer)), len(answer.Answer))
		})
	}
}
