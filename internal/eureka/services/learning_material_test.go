package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	mock_repositories "github.com/manabie-com/backend/mock/eureka/repositories"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestLearningMaterialService_DeleteLearningMaterial(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	lmRepo := &mock_repositories.MockLearningMaterialRepo{}

	lmService := &LearningMaterialService{
		LearningMaterialRepo: lmRepo,
	}

	testCases := []TestCase{
		{
			name: "happy case",
			req: &sspb.DeleteLearningMaterialRequest{
				LearningMaterialId: "learning_material_id",
			},
			setup: func(ctx context.Context) {
				lmRepo.On("Delete", ctx, mock.Anything, mock.Anything).Once().Return(nil)
			},
			expectedResp: &sspb.DeleteLearningMaterialResponse{},
			expectedErr:  nil,
		},
		{
			name: "missing learning_material_id",
			req: &sspb.DeleteLearningMaterialRequest{
				LearningMaterialId: "",
			},
			setup: func(ctx context.Context) {
				// missing learning_material_id will throw error
			},
			expectedResp: nil,
			expectedErr:  status.Errorf(codes.InvalidArgument, fmt.Errorf("validateLearningMaterialReq: LearningMaterial ID must not be empty").Error()),
		},
		{
			name: "error repository",
			req: &sspb.DeleteLearningMaterialRequest{
				LearningMaterialId: "learning_material_id",
			},
			setup: func(ctx context.Context) {
				lmRepo.On("Delete", ctx, mock.Anything, mock.Anything).Once().Return(fmt.Errorf("db.Exec: error repository"))
			},
			expectedResp: nil,
			expectedErr:  status.Errorf(codes.Internal, fmt.Errorf("LearningMaterialRepo.Delete: db.Exec: error repository").Error()),
		},
		{
			name: "not exists learning_material_id",
			req: &sspb.DeleteLearningMaterialRequest{
				LearningMaterialId: "not_exists_learning_material_id",
			},
			setup: func(ctx context.Context) {
				lmRepo.On("Delete", ctx, mock.Anything, mock.Anything).Once().Return(fmt.Errorf("not found any learning material to delete: %w", pgx.ErrNoRows))
			},
			expectedResp: nil,
			expectedErr:  status.Errorf(codes.Internal, fmt.Errorf("LearningMaterialRepo.Delete: not found any learning material to delete: %w", pgx.ErrNoRows).Error()),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			resp, err := lmService.DeleteLearningMaterial(ctx, testCase.req.(*sspb.DeleteLearningMaterialRequest))
			assert.Equal(t, testCase.expectedErr, err)
			if testCase.expectedResp != nil {
				assert.IsType(t, testCase.expectedResp, resp)
			}
		})
	}
}

func TestLearningMaterialService_SwapDisplayOrder(t *testing.T) {
	ctx := context.Background()
	mockDB := &mock_database.Ext{}
	mockTx := &mock_database.Tx{}
	mockLMRepo := &mock_repositories.MockLearningMaterialRepo{}
	mockTopicRepo := &mock_repositories.MockTopicRepo{}

	s := &LearningMaterialService{
		DB:                   mockDB,
		LearningMaterialRepo: mockLMRepo,
		TopicRepo:            mockTopicRepo,
	}

	testCases := []TestCase{
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				lms := []*entities.LearningMaterial{
					{
						ID:           database.Text("lm-id-1"),
						TopicID:      database.Text("topic-id-1"),
						DisplayOrder: database.Int2(1),
					},
					{
						ID:           database.Text("lm-id-2"),
						TopicID:      database.Text("topic-id-1"),
						DisplayOrder: database.Int2(2),
					},
				}
				mockLMRepo.On("FindByIDs", mock.Anything, mockDB, database.TextArray([]string{"lm-id-1", "lm-id-2"})).Once().Return(lms, nil)
				mockDB.On("Begin", mock.Anything, mock.Anything).Return(mockTx, nil)
				mockTopicRepo.On("RetrieveByID", mock.Anything, mock.Anything, database.Text("topic-id-1"), mock.AnythingOfType("[]repositories.QueryEnhancer")).Once().Return(&entities.Topic{}, nil)
				mockLMRepo.On("UpdateDisplayOrders", mock.Anything, mock.Anything, lms).Once().Return(nil)
				mockTx.On("Commit", mock.Anything).Return(nil)
			},
			req: &sspb.SwapDisplayOrderRequest{
				FirstLearningMaterialId:  "lm-id-1",
				SecondLearningMaterialId: "lm-id-2",
			},
			expectedResp: &sspb.SwapDisplayOrderResponse{},
		},
		{
			name: "failed topic validation",
			setup: func(ctx context.Context) {
				lms := []*entities.LearningMaterial{
					{
						ID:           database.Text("lm-id-1"),
						TopicID:      database.Text("topic-id-1"),
						DisplayOrder: database.Int2(1),
					},
					{
						ID:           database.Text("lm-id-2"),
						TopicID:      database.Text("topic-id-2"),
						DisplayOrder: database.Int2(2),
					},
				}
				mockLMRepo.On("FindByIDs", mock.Anything, mockDB, database.TextArray([]string{"lm-id-1", "lm-id-2"})).Once().Return(lms, nil)
			},
			req: &sspb.SwapDisplayOrderRequest{
				FirstLearningMaterialId:  "lm-id-1",
				SecondLearningMaterialId: "lm-id-2",
			},
			expectedErr: status.Error(codes.InvalidArgument, fmt.Errorf("validateSwapDisplayOrderLearningMaterial: LearningMaterials not in the same topic").Error()),
		},
		{
			name: "missing 1 learning material",
			setup: func(ctx context.Context) {
				lms := []*entities.LearningMaterial{
					{
						ID:           database.Text("lm-id-1"),
						TopicID:      database.Text("topic-id-1"),
						DisplayOrder: database.Int2(1),
					},
				}
				mockLMRepo.On("FindByIDs", mock.Anything, mockDB, database.TextArray([]string{"lm-id-1", "lm-id-2"})).Once().Return(lms, nil)
			},
			req: &sspb.SwapDisplayOrderRequest{
				FirstLearningMaterialId:  "lm-id-1",
				SecondLearningMaterialId: "lm-id-2",
			},
			expectedErr: status.Error(codes.InvalidArgument, fmt.Errorf("validateSwapDisplayOrderLearningMaterial: missing LearningMaterials").Error()),
		},
		{
			name:  "missing FirstLearningMaterialId",
			setup: func(ctx context.Context) {},
			req: &sspb.SwapDisplayOrderRequest{
				SecondLearningMaterialId: "lm-id-2",
			},
			expectedErr: status.Error(codes.InvalidArgument, fmt.Errorf("validateSwapDisplayOrderRequest: missing FirstLearningMaterialId").Error()),
		},
		{
			name:  "missing SecondLearningMaterialId",
			setup: func(ctx context.Context) {},
			req: &sspb.SwapDisplayOrderRequest{
				FirstLearningMaterialId: "lm-id-1",
			},
			expectedErr: status.Error(codes.InvalidArgument, fmt.Errorf("validateSwapDisplayOrderRequest: missing SecondLearningMaterialId").Error()),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			resp, err := s.SwapDisplayOrder(ctx, testCase.req.(*sspb.SwapDisplayOrderRequest))
			assert.Equal(t, testCase.expectedErr, err)
			if testCase.expectedResp != nil {
				assert.IsType(t, testCase.expectedResp, resp)
			}
		})
	}
}

func TestLearningMaterialService_DuplicateBook(t *testing.T) {
	ctx := context.Background()
	mockDB := &mock_database.Ext{}
	mockTx := &mock_database.Tx{}
	mockLMRepo := &mock_repositories.MockLearningMaterialRepo{}
	mockBookRepo := &mock_repositories.MockBookRepo{}
	mockBookChapterRepo := &mock_repositories.MockBookChapterRepo{}
	mockChapterRepo := &mock_repositories.MockChapterRepo{}
	mockTopicRepo := &mock_repositories.MockTopicRepo{}
	mockGeneralAssignmentRepo := &mock_repositories.MockGeneralAssignmentRepo{}
	mockFlashCardRepo := &mock_repositories.MockFlashcardRepo{}
	mockLORepo := &mock_repositories.MockLearningObjectiveRepoV2{}
	mockExamLO := &mock_repositories.MockExamLORepo{}
	mockTaskAssignment := &mock_repositories.MockTaskAssignmentRepo{}

	s := &LearningMaterialService{
		DB:                      mockDB,
		LearningMaterialRepo:    mockLMRepo,
		BookRepo:                mockBookRepo,
		BookChapterRepo:         mockBookChapterRepo,
		ChapterRepo:             mockChapterRepo,
		TopicRepo:               mockTopicRepo,
		GeneralAssignmentRepo:   mockGeneralAssignmentRepo,
		FlashcardRepo:           mockFlashCardRepo,
		LearningObjectiveRepoV2: mockLORepo,
		ExamLORepo:              mockExamLO,
		TaskAssignmentRepo:      mockTaskAssignment,
	}

	book := entities.Book{
		ID:   database.Text("book-id"),
		Name: database.Text("Clone book-name"),
	}
	var mapBook = make(map[string]*entities.Book, 0)

	mapBook[book.ID.String] = &entities.Book{
		ID:   database.Text(book.ID.String),
		Name: database.Text("book-name"),
	}
	newBook := entities.Book{
		ID:   database.Text("new-book-id"),
		Name: database.Text("book-name"),
	}
	chapter := entities.Chapter{
		ID:     database.Text("chapter-id"),
		Name:   database.Text("chapter-name"),
		BookID: book.ID,
	}
	copiedChapters := []*entities.CopiedChapter{
		{
			ID:         database.Text("new-chapter-id-1"),
			CopyFromID: chapter.ID,
		},
	}
	newChapter := entities.Chapter{
		ID:     copiedChapters[0].ID,
		Name:   database.Text("chapter-name"),
		BookID: newBook.ID,
	}

	var mapChapter = make(map[string]*entities.Chapter, 0)
	var orgMapChapter = make(map[string]*entities.Chapter, 0)
	var orgChapterIDs = make([]string, 0)
	mapChapter[newChapter.ID.String] = &newChapter
	orgMapChapter[chapter.ID.String] = &chapter

	for id := range orgMapChapter {
		orgChapterIDs = append(orgChapterIDs, id)
	}

	topic := entities.Topic{
		ID:        database.Text("topic-id"),
		Name:      database.Text("topic-name"),
		ChapterID: chapter.ID,
	}
	newTopic := entities.Topic{
		ID:        database.Text("new-topic-id"),
		Name:      database.Text("topic-name"),
		ChapterID: newChapter.ID,
	}
	copiedTopics := []*entities.CopiedTopic{
		{
			ID:         newTopic.ID,
			CopyFromID: topic.ID,
		},
	}
	generalAssignment := []*entities.GeneralAssignment{
		{
			LearningMaterial: entities.LearningMaterial{
				ID:      database.Text("general-assignment-id"),
				Name:    database.Text("general-assignment-name"),
				TopicID: topic.ID,
			},
		},
	}
	flashcard := []*entities.Flashcard{
		{
			LearningMaterial: entities.LearningMaterial{
				ID:      database.Text("flashcard-id"),
				Name:    database.Text("flashcard-name"),
				TopicID: topic.ID,
			},
		},
	}
	lo := []*entities.LearningObjectiveV2{
		{
			LearningMaterial: entities.LearningMaterial{
				ID:      database.Text("learning-objective-id"),
				Name:    database.Text("learning-objective-name"),
				TopicID: topic.ID,
			},
		},
	}
	examLO := []*entities.ExamLO{
		{
			LearningMaterial: entities.LearningMaterial{
				ID:      database.Text("exam-lo-id"),
				Name:    database.Text("exam-lo-name"),
				TopicID: topic.ID,
			},
		},
	}
	taskAssignment := []*entities.TaskAssignment{
		{
			LearningMaterial: entities.LearningMaterial{
				ID:      database.Text("task-assignment-id"),
				Name:    database.Text("task-assignment-name"),
				TopicID: topic.ID,
			},
		},
	}

	orgChapterIDsFromCopied := make([]string, 0, len(copiedChapters))
	newChapterIDsFromCopied := make([]string, 0, len(copiedChapters))

	for _, copiedChapter := range copiedChapters {
		orgChapterIDsFromCopied = append(orgChapterIDsFromCopied, copiedChapter.CopyFromID.String)
		newChapterIDsFromCopied = append(newChapterIDsFromCopied, copiedChapter.ID.String)
	}

	testCases := []TestCase{
		{
			name: "happy case",
			setup: func(ctx context.Context) {

				mockBookRepo.On("FindByIDs", mock.Anything, mockDB, []string{book.ID.String}).Once().Return(mapBook, nil)
				mockChapterRepo.On("FindByBookID", mock.Anything, mockDB, book.ID.String).Return(orgMapChapter, nil)
				mockDB.On("Begin", mock.Anything, mock.Anything).Return(mockTx, nil)

				mockBookRepo.On("DuplicateBook", mock.Anything, mockTx, book.ID, book.Name).Return(newBook.ID.String, nil)
				mockChapterRepo.On("DuplicateChapters", mock.Anything, mockTx, newBook.ID.String, orgChapterIDs).Once().Return(copiedChapters, nil)
				mockTopicRepo.On("DuplicateTopics", mock.Anything, mockTx, database.TextArray(orgChapterIDsFromCopied), database.TextArray(newChapterIDsFromCopied)).Return(copiedTopics, nil)
				mockGeneralAssignmentRepo.On("ListByTopicIDs", mock.Anything, mockTx, database.TextArray([]string{topic.ID.String})).Return(generalAssignment, nil)
				mockGeneralAssignmentRepo.On("BulkInsert", mock.Anything, mockTx, generalAssignment).Return(nil)
				mockFlashCardRepo.On("ListByTopicIDs", mock.Anything, mockTx, database.TextArray([]string{topic.ID.String})).Return(flashcard, nil)
				mockFlashCardRepo.On("BulkInsert", mock.Anything, mockTx, flashcard).Return(nil)
				mockExamLO.On("ListByTopicIDs", mock.Anything, mockTx, database.TextArray([]string{topic.ID.String})).Return(examLO, nil)
				mockExamLO.On("BulkInsert", mock.Anything, mockTx, examLO).Return(nil)
				mockLORepo.On("ListByTopicIDs", mock.Anything, mockTx, database.TextArray([]string{topic.ID.String})).Return(lo, nil)
				mockLORepo.On("BulkInsert", mock.Anything, mockTx, lo).Return(nil)
				mockTaskAssignment.On("ListByTopicIDs", mock.Anything, mockTx, database.TextArray([]string{topic.ID.String})).Return(taskAssignment, nil)
				mockTaskAssignment.On("BulkInsert", mock.Anything, mockTx, taskAssignment).Return(nil)

				mockTx.On("Commit", mock.Anything).Return(nil)
			},
			req: &sspb.DuplicateBookRequest{
				BookId:   book.ID.String,
				BookName: book.Name.String,
			},
			expectedResp: &sspb.DuplicateBookResponse{
				NewBookID:  newBook.ID.String,
				OldTopicId: []string{topic.ID.String},
				NewTopicId: []string{newTopic.ID.String},
			},
		},
		{
			name: "missing book_id",
			setup: func(ctx context.Context) {
				mockBookRepo.On("FindByIDs", mock.Anything, mockDB, []string{book.ID.String}).Once().Return(mapBook, nil)
			},
			req:         &sspb.DuplicateBookRequest{},
			expectedErr: status.Error(codes.InvalidArgument, fmt.Errorf("validateDuplicateBookRequest: missing BookId").Error()),
		},
		{
			name: "err book repo FindByIDs",
			setup: func(ctx context.Context) {
				mockBookRepo.On("FindByIDs", mock.Anything, mockDB, []string{"bookRandomID"}).Once().Return(nil, fmt.Errorf("database.Select: %w", pgx.ErrNoRows))
			},
			req: &sspb.DuplicateBookRequest{
				BookId:   "bookRandomID",
				BookName: "bookRandomName",
			},
			expectedErr: fmt.Errorf("s.BookRepo.FindByIDs: %w", fmt.Errorf("database.Select: %w", pgx.ErrNoRows)),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			resp, err := s.DuplicateBook(ctx, testCase.req.(*sspb.DuplicateBookRequest))
			assert.Equal(t, testCase.expectedErr, err)
			if testCase.expectedResp != nil {
				assert.IsType(t, testCase.expectedResp, resp)
				assert.Equal(t, testCase.expectedResp, resp)
			}
		})
	}
}

func TestValidateDuplicateBookRequest(t *testing.T) {
	testCases := map[string]TestCase{
		"missing book id": {
			req: &sspb.DuplicateBookRequest{
				BookId:   "",
				BookName: "random book name",
			},
			expectedErr: fmt.Errorf("missing BookId"),
		},

		"happy case": {
			req: &sspb.DuplicateBookRequest{
				BookId:   idutil.ULIDNow(),
				BookName: "random book name",
			},
		},
	}
	for caseName, testCase := range testCases {
		t.Run(caseName, func(t *testing.T) {
			bookReq := testCase.req.(*sspb.DuplicateBookRequest)
			if err := validateDuplicateBookRequest(bookReq); testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, testCase.expectedErr, err)
			}
		})
	}
}

func TestValidateListLearningMaterialRequest(t *testing.T) {
	testCases := map[string]TestCase{
		"missing message": {
			req: &sspb.ListLearningMaterialRequest{
				Message: nil,
			},
			expectedErr: fmt.Errorf("missing ListLearningMaterialsRequest"),
		},

		"happy case": {
			req: &sspb.ListLearningMaterialRequest{
				Message: &sspb.ListLearningMaterialRequest_Assignment{
					Assignment: &sspb.ListAssignmentRequest{
						LearningMaterialIds: []string{idutil.ULIDNow()},
					},
				},
			},
		},
	}
	for caseName, testCase := range testCases {
		t.Run(caseName, func(t *testing.T) {
			lmReq := testCase.req.(*sspb.ListLearningMaterialRequest)
			if err := validateListLearningMaterialRequest(lmReq); testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, testCase.expectedErr, err)
			}
		})
	}
}

func TestLearningMaterialService_ListLearningMaterial(t *testing.T) {
	ctx := context.Background()
	mockDB := &mock_database.Ext{}

	mockLearningMaterialRepo := &mock_repositories.MockLearningMaterialRepo{}
	mockFlashcardRepo := &mock_repositories.MockFlashcardRepo{}
	mockLearningObjectiveRepoV2 := &mock_repositories.MockLearningObjectiveRepoV2{}
	mockGeneralAssignmentRepo := &mock_repositories.MockGeneralAssignmentRepo{}
	mockTaskAssignmentRepo := &mock_repositories.MockTaskAssignmentRepo{}
	mockExamLORepo := &mock_repositories.MockExamLORepo{}

	lmService := &LearningMaterialService{
		DB:                      mockDB,
		LearningMaterialRepo:    mockLearningMaterialRepo,
		FlashcardRepo:           mockFlashcardRepo,
		LearningObjectiveRepoV2: mockLearningObjectiveRepoV2,
		GeneralAssignmentRepo:   mockGeneralAssignmentRepo,
		TaskAssignmentRepo:      mockTaskAssignmentRepo,
		ExamLORepo:              mockExamLORepo,
	}

	var (
		baseAssignments     = make([]*sspb.AssignmentBase, 0, 2)
		baseFlashcards      = make([]*sspb.FlashcardBase, 0, 2)
		baseLOs             = make([]*sspb.LearningObjectiveBase, 0, 2)
		baseTaskAssignments = make([]*sspb.TaskAssignmentBase, 0, 2)
		baseExamLOs         = make([]*sspb.ExamLOBase, 0, 2)
	)

	assignments := []*entities.GeneralAssignment{
		{
			LearningMaterial: entities.LearningMaterial{
				ID:   database.Text("assignment-id-1"),
				Name: database.Text("assignment-name-1"),
			},
			Attachments: database.TextArray([]string{"attachment-1", "attachment-2"}),
		},
		{
			LearningMaterial: entities.LearningMaterial{
				ID:   database.Text("assignment-id-2"),
				Name: database.Text("assignment-name-2"),
			},
			Attachments: database.TextArray([]string{"attachment-1", "attachment-2"}),
		},
	}
	examLOs := []*entities.ExamLO{
		{
			LearningMaterial: entities.LearningMaterial{
				ID:   database.Text("exam-lo-id-1"),
				Name: database.Text("exam-lo-name-1"),
			},
		},
		{
			LearningMaterial: entities.LearningMaterial{
				ID:   database.Text("exam-lo-id-2"),
				Name: database.Text("exam-lo-name-2"),
			},
		},
	}
	taskAssignments := []*entities.TaskAssignment{
		{
			LearningMaterial: entities.LearningMaterial{
				ID:   database.Text("task-assignment-id-1"),
				Name: database.Text("task-assignment-name-1"),
			},
			Attachments: database.TextArray([]string{"attachment-1", "attachment-2"}),
		},
		{
			LearningMaterial: entities.LearningMaterial{
				ID:   database.Text("task-assignment-id-2"),
				Name: database.Text("task-assignment-name-2"),
			},
			Attachments: database.TextArray([]string{"attachment-1", "attachment-2"}),
		},
	}
	los := []*entities.LearningObjectiveV2{
		{
			LearningMaterial: entities.LearningMaterial{
				ID:   database.Text("lo-id-1"),
				Name: database.Text("lo-name-1"),
			},
		},
		{
			LearningMaterial: entities.LearningMaterial{
				ID:   database.Text("lo-id-2"),
				Name: database.Text("lo-name-2"),
			},
		},
	}
	flashcards := []*entities.Flashcard{
		{
			LearningMaterial: entities.LearningMaterial{
				ID:   database.Text("flashcard-id-1"),
				Name: database.Text("flashcard-name-1"),
			},
		},
		{
			LearningMaterial: entities.LearningMaterial{
				ID:   database.Text("flashcard-id-2"),
				Name: database.Text("flashcard-name-2"),
			},
		},
	}

	for i := 0; i < len(flashcards); i++ {
		baseAssignment, _ := ToAssignmentPb(assignments[i])
		baseAssignments = append(baseAssignments, baseAssignment)
		baseFlashcards = append(baseFlashcards, ToBaseFlashcard(flashcards[i]))
		baseLOs = append(baseLOs, ToBaseLearningObjectiveV2(los[i]))
		baseExamLOs = append(baseExamLOs, ToBaseExamLO(examLOs[i]))
		baseTaskAssignment, _ := ToTaskAssignmentPb(taskAssignments[i])
		baseTaskAssignments = append(baseTaskAssignments, baseTaskAssignment)
	}

	testCases := []TestCase{
		{
			name: "happy case assignment",
			setup: func(ctx context.Context) {
				mockGeneralAssignmentRepo.On("List", mock.Anything, mockDB, database.TextArray([]string{"assignment-id-1", "assignment-id-2"})).Once().Return(assignments, nil)
			},
			req: &sspb.ListLearningMaterialRequest{
				Message: &sspb.ListLearningMaterialRequest_Assignment{
					Assignment: &sspb.ListAssignmentRequest{
						LearningMaterialIds: []string{"assignment-id-1", "assignment-id-2"},
					},
				},
			},
			expectedResp: &sspb.ListLearningMaterialResponse{
				Message: &sspb.ListLearningMaterialResponse_Assignment{
					Assignment: &sspb.ListAssignmentResponse{
						Assignments: baseAssignments,
					},
				},
			},
		},
		{
			name: "happy case flashcard",
			setup: func(ctx context.Context) {
				mockFlashcardRepo.On("ListFlashcard", mock.Anything, mockDB, &repositories.ListFlashcardArgs{
					LearningMaterialIDs: database.TextArray([]string{
						"flashcard-id-1",
						"flashcard-id-2",
					}),
				}).Once().Return(flashcards, nil)
			},
			req: &sspb.ListLearningMaterialRequest{
				Message: &sspb.ListLearningMaterialRequest_Flashcard{
					Flashcard: &sspb.ListFlashcardRequest{
						LearningMaterialIds: []string{"flashcard-id-1", "flashcard-id-2"},
					},
				},
			},
			expectedResp: &sspb.ListLearningMaterialResponse{
				Message: &sspb.ListLearningMaterialResponse_Flashcard{
					Flashcard: &sspb.ListFlashcardResponse{
						Flashcards: baseFlashcards,
					},
				},
			},
		},
		{
			name: "happy case lo",
			setup: func(ctx context.Context) {
				mockLearningObjectiveRepoV2.On("ListByIDs", mock.Anything, mockDB, database.TextArray([]string{"lo-id-1", "lo-id-2"})).Once().Return(los, nil)
			},
			req: &sspb.ListLearningMaterialRequest{
				Message: &sspb.ListLearningMaterialRequest_LearningObjective{
					LearningObjective: &sspb.ListLearningObjectiveRequest{
						LearningMaterialIds: []string{"lo-id-1", "lo-id-2"},
					},
				},
			},
			expectedResp: &sspb.ListLearningMaterialResponse{
				Message: &sspb.ListLearningMaterialResponse_LearningObjective{
					LearningObjective: &sspb.ListLearningObjectiveResponse{
						LearningObjectives: baseLOs,
					},
				},
			},
		},
		{
			name: "happy case exam lo",
			setup: func(ctx context.Context) {
				mockExamLORepo.On("ListByIDs", mock.Anything, mockDB, database.TextArray([]string{"exam-lo-id-1", "exam-lo-id-2"})).Once().Return(examLOs, nil)
			},
			req: &sspb.ListLearningMaterialRequest{
				Message: &sspb.ListLearningMaterialRequest_ExamLo{
					ExamLo: &sspb.ListExamLORequest{
						LearningMaterialIds: []string{"exam-lo-id-1", "exam-lo-id-2"},
					},
				},
			},
			expectedResp: &sspb.ListLearningMaterialResponse{
				Message: &sspb.ListLearningMaterialResponse_ExamLo{
					ExamLo: &sspb.ListExamLOResponse{
						ExamLos: baseExamLOs,
					},
				},
			},
		},
		{
			name: "happy case task assignment",
			setup: func(ctx context.Context) {
				mockTaskAssignmentRepo.On("List", mock.Anything, mockDB, database.TextArray([]string{"task-assignment-id-1", "task-assignment-id-2"})).Once().Return(taskAssignments, nil)
			},
			req: &sspb.ListLearningMaterialRequest{
				Message: &sspb.ListLearningMaterialRequest_TaskAssignment{
					TaskAssignment: &sspb.ListTaskAssignmentRequest{
						LearningMaterialIds: []string{"task-assignment-id-1", "task-assignment-id-2"},
					},
				},
			},
			expectedResp: &sspb.ListLearningMaterialResponse{
				Message: &sspb.ListLearningMaterialResponse_TaskAssignment{
					TaskAssignment: &sspb.ListTaskAssignmentResponse{
						TaskAssignments: baseTaskAssignments,
					},
				},
			},
		},
		{
			name: "no row task assignment",
			setup: func(ctx context.Context) {
				mockTaskAssignmentRepo.On("List", mock.Anything, mockDB, database.TextArray([]string{"randomID"})).Once().Return(nil, fmt.Errorf("%w", pgx.ErrNoRows))

			},
			req: &sspb.ListLearningMaterialRequest{
				Message: &sspb.ListLearningMaterialRequest_TaskAssignment{
					TaskAssignment: &sspb.ListTaskAssignmentRequest{
						LearningMaterialIds: []string{"randomID"},
					},
				},
			},
			expectedErr: status.Error(codes.NotFound, fmt.Errorf("s.TaskAssignmentRepo.List: %w", pgx.ErrNoRows).Error()),
		},
		{
			name: "internal error task assignment",
			setup: func(ctx context.Context) {
				mockTaskAssignmentRepo.On("List", mock.Anything, mockDB, mock.Anything).Once().Return(nil, fmt.Errorf("%w", pgx.ErrTxClosed))

			},
			req: &sspb.ListLearningMaterialRequest{
				Message: &sspb.ListLearningMaterialRequest_TaskAssignment{
					TaskAssignment: &sspb.ListTaskAssignmentRequest{
						LearningMaterialIds: []string{},
					},
				},
			},
			expectedErr: status.Error(codes.Internal, fmt.Errorf("s.TaskAssignmentRepo.List: %w", pgx.ErrTxClosed).Error()),
		},
		{
			name: "no row assignment",
			setup: func(ctx context.Context) {
				mockGeneralAssignmentRepo.On("List", mock.Anything, mockDB, database.TextArray([]string{"randomID"})).Once().Return(nil, fmt.Errorf("%w", pgx.ErrNoRows))

			},
			req: &sspb.ListLearningMaterialRequest{
				Message: &sspb.ListLearningMaterialRequest_Assignment{
					Assignment: &sspb.ListAssignmentRequest{
						LearningMaterialIds: []string{"randomID"},
					},
				},
			},
			expectedErr: status.Error(codes.NotFound, fmt.Errorf("s.GeneralAssignmentRepo.List: %w", pgx.ErrNoRows).Error()),
		},
		{
			name: "internal error assignment",
			setup: func(ctx context.Context) {
				mockGeneralAssignmentRepo.On("List", mock.Anything, mockDB, mock.Anything).Once().Return(nil, fmt.Errorf("%w", pgx.ErrTxClosed))

			},
			req: &sspb.ListLearningMaterialRequest{
				Message: &sspb.ListLearningMaterialRequest_Assignment{
					Assignment: &sspb.ListAssignmentRequest{
						LearningMaterialIds: []string{},
					},
				},
			},
			expectedErr: status.Error(codes.Internal, fmt.Errorf("s.GeneralAssignmentRepo.List: %w", pgx.ErrTxClosed).Error()),
		},
		{
			name: "no row flashcard",
			setup: func(ctx context.Context) {
				mockFlashcardRepo.On("ListFlashcard", mock.Anything, mockDB, &repositories.ListFlashcardArgs{
					LearningMaterialIDs: database.TextArray([]string{"randomID"}),
				}).Once().Return(nil, fmt.Errorf("database.Select: %w", pgx.ErrNoRows))

			},
			req: &sspb.ListLearningMaterialRequest{
				Message: &sspb.ListLearningMaterialRequest_Flashcard{
					Flashcard: &sspb.ListFlashcardRequest{
						LearningMaterialIds: []string{"randomID"},
					},
				},
			},
			expectedErr: status.Error(codes.NotFound, fmt.Errorf("s.FlashcardRepo.ListFlashcard: database.Select: %w", pgx.ErrNoRows).Error()),
		},
		{
			name: "internal error flashcard",
			setup: func(ctx context.Context) {
				mockFlashcardRepo.On("ListFlashcard", mock.Anything, mockDB, mock.Anything).Once().Return(nil, fmt.Errorf("%w", pgx.ErrTxClosed))

			},
			req: &sspb.ListLearningMaterialRequest{
				Message: &sspb.ListLearningMaterialRequest_Flashcard{
					Flashcard: &sspb.ListFlashcardRequest{
						LearningMaterialIds: []string{},
					},
				},
			},
			expectedErr: status.Error(codes.Internal, fmt.Errorf("s.FlashcardRepo.ListFlashcard: %w", pgx.ErrTxClosed).Error()),
		},
		{
			name: "no row exam lo",
			setup: func(ctx context.Context) {
				mockExamLORepo.On("ListByIDs", mock.Anything, mockDB, database.TextArray([]string{"randomID"})).Once().Return(nil, fmt.Errorf("database.Select: %w", pgx.ErrNoRows))

			},
			req: &sspb.ListLearningMaterialRequest{
				Message: &sspb.ListLearningMaterialRequest_ExamLo{
					ExamLo: &sspb.ListExamLORequest{
						LearningMaterialIds: []string{"randomID"},
					},
				},
			},
			expectedErr: status.Error(codes.NotFound, fmt.Errorf("s.ExamLORepo.ListByIDs: database.Select: %w", pgx.ErrNoRows).Error()),
		},
		{
			name: "internal error exam lo",
			setup: func(ctx context.Context) {
				mockExamLORepo.On("ListByIDs", mock.Anything, mockDB, mock.Anything).Once().Return(nil, fmt.Errorf("%w", pgx.ErrTxClosed))

			},
			req: &sspb.ListLearningMaterialRequest{
				Message: &sspb.ListLearningMaterialRequest_ExamLo{
					ExamLo: &sspb.ListExamLORequest{
						LearningMaterialIds: []string{},
					},
				},
			},
			expectedErr: status.Error(codes.Internal, fmt.Errorf("s.ExamLORepo.ListByIDs: %w", pgx.ErrTxClosed).Error()),
		},
		{
			name: "no row lo",
			setup: func(ctx context.Context) {
				mockLearningObjectiveRepoV2.On("ListByIDs", mock.Anything, mockDB, database.TextArray([]string{"randomID"})).Once().Return(nil, fmt.Errorf("database.Select: %w", pgx.ErrNoRows))

			},
			req: &sspb.ListLearningMaterialRequest{
				Message: &sspb.ListLearningMaterialRequest_LearningObjective{
					LearningObjective: &sspb.ListLearningObjectiveRequest{
						LearningMaterialIds: []string{"randomID"},
					},
				},
			},
			expectedErr: status.Error(codes.NotFound, fmt.Errorf("s.LearningObjectiveRepoV2.ListByIDs: database.Select: %w", pgx.ErrNoRows).Error()),
		},
		{
			name: "internal error lo",
			setup: func(ctx context.Context) {
				mockLearningObjectiveRepoV2.On("ListByIDs", mock.Anything, mockDB, mock.Anything).Once().Return(nil, fmt.Errorf("%w", pgx.ErrTxClosed))

			},
			req: &sspb.ListLearningMaterialRequest{
				Message: &sspb.ListLearningMaterialRequest_LearningObjective{
					LearningObjective: &sspb.ListLearningObjectiveRequest{
						LearningMaterialIds: []string{},
					},
				},
			},
			expectedErr: status.Error(codes.Internal, fmt.Errorf("s.LearningObjectiveRepoV2.ListByIDs: %w", pgx.ErrTxClosed).Error()),
		},
		{
			name: "missing learning material ids",
			setup: func(ctx context.Context) {
			},
			req:         &sspb.ListLearningMaterialRequest{},
			expectedErr: status.Errorf(codes.InvalidArgument, "validateListLearningMaterialRequest: missing ListLearningMaterialsRequest"),
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			resp, err := lmService.ListLearningMaterial(ctx, testCase.req.(*sspb.ListLearningMaterialRequest))
			assert.Equal(t, testCase.expectedErr, err)
			if testCase.expectedResp != nil {
				assert.IsType(t, testCase.expectedResp, resp)
				assert.Equal(t, testCase.expectedResp, resp)
			}
		})
	}
}

type AssertDuplicateLearningMaterialArgument struct {
	IsPublished pgtype.Bool
}

func assertDuplicateLearningMaterialCommonFields(t *testing.T, fieldObj AssertDuplicateLearningMaterialArgument) {
	// Spec: The newly duplicated LM publishing status are Unpublished by default.
	assert.Equal(t, fieldObj.IsPublished.Bool, false)
}

func Test_duplicateGeneralAssignment(t *testing.T) {
	ctx := context.Background()
	mockDB := &mock_database.Ext{}

	mockGeneralAssignmentRepo := &mock_repositories.MockGeneralAssignmentRepo{}
	topicID := idutil.ULIDNow()
	newTopicID := idutil.ULIDNow()
	s := &LearningMaterialService{
		DB:                    mockDB,
		GeneralAssignmentRepo: mockGeneralAssignmentRepo,
	}

	generalAssignment := []*entities.GeneralAssignment{
		{
			LearningMaterial: entities.LearningMaterial{
				ID:          database.Text("general-assignment-id"),
				Name:        database.Text("general-assignment-name"),
				TopicID:     database.Text(topicID),
				IsPublished: database.Bool(true),
			},
		},
	}
	var resp = make(map[string]string, 1)
	resp[topicID] = newTopicID
	testCases := []TestCase{
		{
			req:  topicID,
			name: "happy case",
			setup: func(ctx context.Context) {

				mockGeneralAssignmentRepo.On("ListByTopicIDs", mock.Anything, mockDB, database.TextArray([]string{topicID})).Return(generalAssignment, nil)
				mockGeneralAssignmentRepo.On("BulkInsert", mock.Anything, mockDB, generalAssignment).Run(func(args mock.Arguments) {
					generalAssignments := args[2].([]*entities.GeneralAssignment)

					for _, item := range generalAssignments {
						assertDuplicateLearningMaterialCommonFields(t, AssertDuplicateLearningMaterialArgument{
							IsPublished: item.IsPublished,
						})
					}

				}).Return(nil)
			},
			expectedErr: nil,
		},
		{
			req:  "randomID",
			name: "no row",
			setup: func(ctx context.Context) {
				mockGeneralAssignmentRepo.On("ListByTopicIDs", mock.Anything, mockDB, database.TextArray([]string{"randomID"})).Return(nil, fmt.Errorf("database.Select: %w", pgx.ErrNoRows))
			},
			expectedErr: fmt.Errorf("s.GeneralAssignmentRepo.ListByTopicIDs: %w", fmt.Errorf("database.Select: %w", pgx.ErrNoRows)),
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			err := s.duplicateGeneralAssignment(ctx, mockDB, []string{testCase.req.(string)}, resp)
			assert.Equal(t, testCase.expectedErr, err)
		})
	}
}

func Test_duplicateFlashcard(t *testing.T) {
	ctx := context.Background()
	mockDB := &mock_database.Ext{}

	mockFlashcardRepo := &mock_repositories.MockFlashcardRepo{}
	topicID := idutil.ULIDNow()
	newTopicID := idutil.ULIDNow()
	s := &LearningMaterialService{
		DB:            mockDB,
		FlashcardRepo: mockFlashcardRepo,
	}

	flashcard := []*entities.Flashcard{
		{
			LearningMaterial: entities.LearningMaterial{
				ID:          database.Text("flashcard-id"),
				Name:        database.Text("flashcard-name"),
				TopicID:     database.Text(topicID),
				IsPublished: database.Bool(true),
			},
		},
	}
	var resp = make(map[string]string, 1)
	resp[topicID] = newTopicID
	testCases := []TestCase{
		{
			req:  topicID,
			name: "happy case",
			setup: func(ctx context.Context) {

				mockFlashcardRepo.On("ListByTopicIDs", mock.Anything, mockDB, database.TextArray([]string{topicID})).Return(flashcard, nil)
				mockFlashcardRepo.On("BulkInsert", mock.Anything, mockDB, flashcard).Run(func(args mock.Arguments) {
					flashcards := args[2].([]*entities.Flashcard)

					for _, item := range flashcards {
						assertDuplicateLearningMaterialCommonFields(t, AssertDuplicateLearningMaterialArgument{
							IsPublished: item.IsPublished,
						})
					}

				}).Return(nil)
			},
			expectedErr: nil,
		},
		{
			req:  "randomID",
			name: "no row",
			setup: func(ctx context.Context) {
				mockFlashcardRepo.On("ListByTopicIDs", mock.Anything, mockDB, database.TextArray([]string{"randomID"})).Return(nil, fmt.Errorf("database.Select: %w", pgx.ErrNoRows))
			},
			expectedErr: fmt.Errorf("s.FlashcardRepo.ListByTopicIDs: %w", fmt.Errorf("database.Select: %w", pgx.ErrNoRows)),
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			err := s.duplicateFlashcard(ctx, mockDB, []string{testCase.req.(string)}, resp)
			assert.Equal(t, testCase.expectedErr, err)
		})
	}
}

func Test_duplicateLO(t *testing.T) {
	ctx := context.Background()
	mockDB := &mock_database.Ext{}

	mockLORepo := &mock_repositories.MockLearningObjectiveRepoV2{}
	topicID := idutil.ULIDNow()
	newTopicID := idutil.ULIDNow()
	s := &LearningMaterialService{
		DB:                      mockDB,
		LearningObjectiveRepoV2: mockLORepo,
	}

	lo := []*entities.LearningObjectiveV2{
		{
			LearningMaterial: entities.LearningMaterial{
				ID:          database.Text("learning-objective-id"),
				Name:        database.Text("learning-objective-name"),
				TopicID:     database.Text(topicID),
				IsPublished: database.Bool(true),
			},
		},
	}
	var resp = make(map[string]string, 1)
	resp[topicID] = newTopicID
	testCases := []TestCase{
		{
			req:  topicID,
			name: "happy case",
			setup: func(ctx context.Context) {

				mockLORepo.On("ListByTopicIDs", mock.Anything, mockDB, database.TextArray([]string{topicID})).Return(lo, nil)
				mockLORepo.On("BulkInsert", mock.Anything, mockDB, lo).Run(func(args mock.Arguments) {
					los := args[2].([]*entities.LearningObjectiveV2)

					for _, item := range los {
						assertDuplicateLearningMaterialCommonFields(t, AssertDuplicateLearningMaterialArgument{
							IsPublished: item.IsPublished,
						})
					}

				}).Return(nil)
			},
			expectedErr: nil,
		},
		{
			req:  "randomID",
			name: "no row",
			setup: func(ctx context.Context) {
				mockLORepo.On("ListByTopicIDs", mock.Anything, mockDB, database.TextArray([]string{"randomID"})).Return(nil, fmt.Errorf("database.Select: %w", pgx.ErrNoRows))
			},
			expectedErr: fmt.Errorf("s.LearningObjectiveRepoV2.ListByTopicIDs: %w", fmt.Errorf("database.Select: %w", pgx.ErrNoRows)),
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			err := s.duplicateLO(ctx, mockDB, []string{testCase.req.(string)}, resp)
			assert.Equal(t, testCase.expectedErr, err)
		})
	}
}

func Test_duplicateExamLO(t *testing.T) {
	ctx := context.Background()
	mockDB := &mock_database.Ext{}

	mockExamLORepo := &mock_repositories.MockExamLORepo{}
	topicID := idutil.ULIDNow()
	newTopicID := idutil.ULIDNow()
	s := &LearningMaterialService{
		DB:         mockDB,
		ExamLORepo: mockExamLORepo,
	}

	examLO := []*entities.ExamLO{
		{
			LearningMaterial: entities.LearningMaterial{
				ID:          database.Text("learning-objective-id"),
				Name:        database.Text("learning-objective-name"),
				TopicID:     database.Text(topicID),
				IsPublished: database.Bool(true),
			},
		},
	}
	var resp = make(map[string]string, 0)
	resp[topicID] = newTopicID

	testCases := []TestCase{
		{
			req:  topicID,
			name: "happy case",
			setup: func(ctx context.Context) {

				mockExamLORepo.On("ListByTopicIDs", mock.Anything, mockDB, database.TextArray([]string{topicID})).Return(examLO, nil)
				mockExamLORepo.On("BulkInsert", mock.Anything, mockDB, examLO).Run(func(args mock.Arguments) {
					examLOs := args[2].([]*entities.ExamLO)

					for _, item := range examLOs {
						assertDuplicateLearningMaterialCommonFields(t, AssertDuplicateLearningMaterialArgument{
							IsPublished: item.IsPublished,
						})
					}

				}).Return(nil)
			},
			expectedErr: nil,
		},
		{
			req:  "randomID",
			name: "no row",
			setup: func(ctx context.Context) {
				mockExamLORepo.On("ListByTopicIDs", mock.Anything, mockDB, database.TextArray([]string{"randomID"})).Return(nil, fmt.Errorf("database.Select: %w", pgx.ErrNoRows))
			},
			expectedErr: fmt.Errorf("s.ExamLORepo.ListByTopicIDs: %w", fmt.Errorf("database.Select: %w", pgx.ErrNoRows)),
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			err := s.duplicateExamLO(ctx, mockDB, []string{testCase.req.(string)}, resp)
			assert.Equal(t, testCase.expectedErr, err)
		})
	}
}

func Test_duplicateTaskAssignment(t *testing.T) {
	ctx := context.Background()
	mockDB := &mock_database.Ext{}

	mockTaskAssignmentRepo := &mock_repositories.MockTaskAssignmentRepo{}
	topicID := idutil.ULIDNow()
	newTopicID := idutil.ULIDNow()
	s := &LearningMaterialService{
		DB:                 mockDB,
		TaskAssignmentRepo: mockTaskAssignmentRepo,
	}

	taskAssignment := []*entities.TaskAssignment{
		{
			LearningMaterial: entities.LearningMaterial{
				ID:          database.Text("task-assignment-id"),
				Name:        database.Text("task-assignment-name"),
				TopicID:     database.Text(topicID),
				IsPublished: database.Bool(true),
			},
		},
	}
	var resp = make(map[string]string, 1)
	resp[topicID] = newTopicID
	testCases := []TestCase{
		{
			req:  topicID,
			name: "happy case",
			setup: func(ctx context.Context) {

				mockTaskAssignmentRepo.On("ListByTopicIDs", mock.Anything, mockDB, database.TextArray([]string{topicID})).Return(taskAssignment, nil)
				mockTaskAssignmentRepo.On("BulkInsert", mock.Anything, mockDB, taskAssignment).Run(func(args mock.Arguments) {
					taskAssignments := args[2].([]*entities.TaskAssignment)

					for _, item := range taskAssignments {
						assertDuplicateLearningMaterialCommonFields(t, AssertDuplicateLearningMaterialArgument{
							IsPublished: item.IsPublished,
						})
					}

				}).Return(nil)
			},
		},
		{
			req:  "randomID",
			name: "no row",
			setup: func(ctx context.Context) {
				mockTaskAssignmentRepo.On("ListByTopicIDs", mock.Anything, mockDB, database.TextArray([]string{"randomID"})).Return(nil, fmt.Errorf("database.Select: %w", pgx.ErrNoRows))
			},
			expectedErr: fmt.Errorf("s.TaskAssignmentRepo.ListByTopicIDs: %w", fmt.Errorf("database.Select: %w", pgx.ErrNoRows)),
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			err := s.duplicateTaskAssignment(ctx, mockDB, []string{testCase.req.(string)}, resp)
			assert.Equal(t, testCase.expectedErr, err)
		})
	}
}

func TestUpdateLMNameRequest(t *testing.T) {
	testCases := map[string]TestCase{
		"missing lm id": {
			req: &sspb.UpdateLearningMaterialNameRequest{
				LearningMaterialId: "",
			},
			expectedErr: fmt.Errorf("missing field LearningMaterialId"),
		},
		"missing lm name": {
			req: &sspb.UpdateLearningMaterialNameRequest{
				LearningMaterialId: "randomID",
			},
			expectedErr: fmt.Errorf("missing field NewLearningMaterialName"),
		},

		"happy case": {
			req: &sspb.UpdateLearningMaterialNameRequest{
				LearningMaterialId:      "id",
				NewLearningMaterialName: "name",
			},
		},
	}
	for caseName, testCase := range testCases {
		t.Run(caseName, func(t *testing.T) {
			lmReq := testCase.req.(*sspb.UpdateLearningMaterialNameRequest)
			if err := validateUpdateLearningMaterialNameRequest(lmReq); testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, testCase.expectedErr, err)
			}
		})
	}
}

func TestLearningMaterialService_UpdateLearningMaterialName(t *testing.T) {
	ctx := context.Background()
	mockDB := &mock_database.Ext{}

	mockLearningMaterialRepo := &mock_repositories.MockLearningMaterialRepo{}
	lmService := &LearningMaterialService{
		DB:                   mockDB,
		LearningMaterialRepo: mockLearningMaterialRepo,
	}
	testCases := []TestCase{
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				mockLearningMaterialRepo.On("UpdateName", mock.Anything, mockDB, database.Text("id"), database.Text("name")).Return(int64(1), nil)
			},
			req: &sspb.UpdateLearningMaterialNameRequest{
				LearningMaterialId:      "id",
				NewLearningMaterialName: "name",
			},
		},
		{
			name: "no row",
			setup: func(ctx context.Context) {
				mockLearningMaterialRepo.On("UpdateName", mock.Anything, mockDB, database.Text("id-1"), database.Text("name")).Return(int64(0), nil)
			},
			req: &sspb.UpdateLearningMaterialNameRequest{
				LearningMaterialId:      "id-1",
				NewLearningMaterialName: "name",
			},
			expectedErr: status.Error(codes.NotFound, fmt.Errorf("s.LearningMaterialRepo.UpdateName not found any learning material to update name: %w", pgx.ErrNoRows).Error()),
		},
		{
			name: "internal error",
			setup: func(ctx context.Context) {
				mockLearningMaterialRepo.On("UpdateName", mock.Anything, mockDB, database.Text("id-2"), database.Text("name")).Return(int64(0), pgx.ErrTxClosed)
			},
			req: &sspb.UpdateLearningMaterialNameRequest{
				LearningMaterialId:      "id-2",
				NewLearningMaterialName: "name",
			},
			expectedErr: status.Error(codes.Internal, fmt.Errorf("s.LearningMaterialRepo.UpdateName: %w", pgx.ErrTxClosed).Error()),
		},
		{
			name: "missing learning material id",
			setup: func(ctx context.Context) {
			},
			req:         &sspb.UpdateLearningMaterialNameRequest{},
			expectedErr: status.Errorf(codes.InvalidArgument, "validateUpdateLearningMaterialNameRequest: missing field LearningMaterialId"),
		},
		{
			name: "missing learning material name",
			setup: func(ctx context.Context) {
			},
			req: &sspb.UpdateLearningMaterialNameRequest{
				LearningMaterialId: "randomID",
			},
			expectedErr: status.Errorf(codes.InvalidArgument, "validateUpdateLearningMaterialNameRequest: missing field NewLearningMaterialName"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			_, err := lmService.UpdateLearningMaterialName(ctx, testCase.req.(*sspb.UpdateLearningMaterialNameRequest))
			assert.Equal(t, testCase.expectedErr, err)
		})
	}
}
