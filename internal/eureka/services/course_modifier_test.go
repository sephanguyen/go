package services

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	mock_repositories "github.com/manabie-com/backend/mock/eureka/repositories"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestCourseModifierService_UpdateDisplayOrdersOfLOsAndAssignments(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}

	topicsAssignmentsRepo := new(mock_repositories.MockTopicsAssignmentsRepo)
	assignmentRepo := new(mock_repositories.MockAssignmentRepo)

	testCases := []TestCase{
		{
			name: "happy case",
			ctx:  interceptors.ContextWithUserID(ctx, "student"),
			req: &pb.UpdateDisplayOrdersOfLOsAndAssignmentsRequest{
				Assignments: []*pb.UpdateDisplayOrdersOfLOsAndAssignmentsRequest_Assignment{
					{
						AssignmentId: "assignment-id",
					},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)

				topicsAssignmentsRepo.On("BulkUpdateDisplayOrder", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)

				assignmentRepo.On("UpdateDisplayOrders", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
			},
		},
	}

	s := &CourseModifierService{
		DB:                    db,
		TopicsAssignmentsRepo: topicsAssignmentsRepo,
		AssignmentRepo:        assignmentRepo,
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			req := testCase.req.(*pb.UpdateDisplayOrdersOfLOsAndAssignmentsRequest)
			_, err := s.UpdateDisplayOrdersOfLOsAndAssignments(testCase.ctx, req)
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCourseModifierService_FinishFlashCardStudyProgress(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}

	flashCardProgressionRepo := new(mock_repositories.MockFlashcardProgressionRepo)
	studyPlanItemRepo := new(mock_repositories.MockStudyPlanItemRepo)

	testCases := []TestCase{
		{
			name: "happy case with isRestart false",
			ctx:  interceptors.ContextWithUserID(ctx, "student"),
			req: &pb.FinishFlashCardStudyProgressRequest{
				StudySetId:      "study-set-id",
				StudentId:       "student-id",
				LoId:            "lo-id",
				StudyPlanItemId: "study-plan-item-id",
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				flashCardProgressionRepo.On("Get", mock.Anything, mock.Anything, mock.Anything).Once().Return(&entities.FlashcardProgression{
					CompletedAt: pgtype.Timestamptz{
						Status: pgtype.Null,
					},
					QuizExternalIDs:       database.TextArray([]string{}),
					RememberedQuestionIDs: database.TextArray([]string{}),
				}, nil)

				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
				flashCardProgressionRepo.On("UpdateCompletedAt", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				studyPlanItemRepo.On("UpdateCompletedAtByID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "happy case with isRestart true",
			ctx:  interceptors.ContextWithUserID(ctx, "student"),
			req: &pb.FinishFlashCardStudyProgressRequest{
				StudySetId:      "study-set-id",
				StudentId:       "student-id",
				LoId:            "lo-id",
				StudyPlanItemId: "study-plan-item-id",
				IsRestart:       true,
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				flashCardProgressionRepo.On("Get", mock.Anything, mock.Anything, mock.Anything).Once().Return(&entities.FlashcardProgression{
					CompletedAt: pgtype.Timestamptz{
						Status: pgtype.Null,
					},
					QuizExternalIDs:       database.TextArray([]string{}),
					RememberedQuestionIDs: database.TextArray([]string{}),
				}, nil)

				flashCardProgressionRepo.On("DeleteByStudySetID", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
			},
		},
	}

	s := &CourseModifierService{
		DB:                       db,
		FlashcardProgressionRepo: flashCardProgressionRepo,
		StudyPlanItemRepo:        studyPlanItemRepo,
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			req := testCase.req.(*pb.FinishFlashCardStudyProgressRequest)
			_, err := s.FinishFlashCardStudyProgress(testCase.ctx, req)
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCourseModifierService_UpdateFlashCardStudyProgress(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	db := &mock_database.Ext{}

	flashCardProgressionRepo := new(mock_repositories.MockFlashcardProgressionRepo)

	testCases := []TestCase{
		{
			name: "happy case ",
			ctx:  interceptors.ContextWithUserID(ctx, "student"),
			req: &pb.UpdateFlashCardStudyProgressRequest{
				StudySetId: "study-set-id",
				StudentId:  "student-id",
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				flashCardProgressionRepo.On("GetByStudySetIDAndStudentID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(&entities.FlashcardProgression{
					CompletedAt: pgtype.Timestamptz{
						Status: pgtype.Null,
					},
				}, nil)
				flashCardProgressionRepo.On("Upsert", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "error upsert ",
			ctx:  interceptors.ContextWithUserID(ctx, "student"),
			req: &pb.UpdateFlashCardStudyProgressRequest{
				StudySetId: "study-set-id",
				StudentId:  "student-id",
			},
			expectedErr: status.Error(codes.Internal, fmt.Sprintf("s.FlashcardProgressionRepo.Upsert: %v", pgx.ErrNoRows)),
			setup: func(ctx context.Context) {
				flashCardProgressionRepo.On("GetByStudySetIDAndStudentID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(&entities.FlashcardProgression{
					CompletedAt: pgtype.Timestamptz{
						Status: pgtype.Null,
					},
				}, nil)
				flashCardProgressionRepo.On("Upsert", mock.Anything, mock.Anything, mock.Anything).Once().Return(pgx.ErrNoRows)
			},
		},
	}

	s := &CourseModifierService{
		DB:                       db,
		FlashcardProgressionRepo: flashCardProgressionRepo,
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			req := testCase.req.(*pb.UpdateFlashCardStudyProgressRequest)
			_, err := s.UpdateFlashCardStudyProgress(testCase.ctx, req)
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCourseModifierService_CompleteStudyPlanItem(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	db := &mock_database.Ext{}

	studyPlanItemRepo := new(mock_repositories.MockStudyPlanItemRepo)

	testCases := []TestCase{
		{
			name:        "error req missing study plan item id",
			ctx:         interceptors.ContextWithUserID(ctx, "student"),
			req:         &pb.CompleteStudyPlanItemRequest{},
			expectedErr: status.Error(codes.InvalidArgument, "missing study plan item id"),
			setup:       func(ctx context.Context) {},
		},
		{
			name: "error ErrNoRows ",
			ctx:  interceptors.ContextWithUserID(ctx, "student"),
			req: &pb.CompleteStudyPlanItemRequest{
				StudyPlanItemId: "study-plan-item-id",
			},
			expectedErr: status.Errorf(codes.Internal, "s.assignmentModifierService.StudyPlanItemRepo.UpdateCompletedAtByID: %v", pgx.ErrNoRows),
			setup: func(ctx context.Context) {
				studyPlanItemRepo.On("UpdateCompletedAtByID", ctx, db, database.Text("study-plan-item-id"), mock.Anything).Once().Return(pgx.ErrNoRows)
			},
		},
		{
			name: "happy case ",
			ctx:  interceptors.ContextWithUserID(ctx, "student"),
			req: &pb.CompleteStudyPlanItemRequest{
				StudyPlanItemId: "study-plan-item-id",
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				studyPlanItemRepo.On("UpdateCompletedAtByID", ctx, db, database.Text("study-plan-item-id"), mock.Anything).Once().Return(nil)
			},
		},
	}

	s := &CourseModifierService{
		DB: db,
		assignmentModifierService: &AssignmentModifierService{
			StudyPlanItemRepo: studyPlanItemRepo,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			req := testCase.req.(*pb.CompleteStudyPlanItemRequest)
			_, err := s.CompleteStudyPlanItem(testCase.ctx, req)
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDuplicateBook(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	bookChapterRepo := &mock_repositories.MockBookChapterRepo{}
	bookRepo := &mock_repositories.MockBookRepo{}
	topicRepo := &mock_repositories.MockTopicRepo{}
	chapterRepo := &mock_repositories.MockChapterRepo{}
	learningObjectiveRepo := &mock_repositories.MockLearningObjectiveRepo{}
	topicLearningObjectiveRepo := &mock_repositories.MockTopicsLearningObjectivesRepo{}
	assignmentRepo := &mock_repositories.MockAssignmentRepo{}
	topicAssignmentRepo := &mock_repositories.MockTopicsAssignmentsRepo{}
	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}

	s := &CourseModifierService{
		DB:                           db,
		BookChapterRepo:              bookChapterRepo,
		BookRepo:                     bookRepo,
		ChapterRepo:                  chapterRepo,
		LearningObjectiveRepo:        learningObjectiveRepo,
		TopicsLearningObjectivesRepo: topicLearningObjectiveRepo,
		TopicRepo:                    topicRepo,
		AssignmentRepo:               assignmentRepo,
		TopicsAssignmentsRepo:        topicAssignmentRepo,
	}
	now := time.Now()
	nowTimestamptz := database.Timestamptz(now)
	bookChapter := &entities.BookChapter{
		BookID:    database.Text("book-id"),
		ChapterID: database.Text("chapter-id"),
		CreatedAt: nowTimestamptz,
		UpdatedAt: nowTimestamptz,
	}
	bookChapterMap := map[string][]*entities.BookChapter{
		"book-id": {bookChapter},
	}
	copiedChapter := []*entities.CopiedChapter{
		{
			ID:         database.Text("chapter-id"),
			CopyFromID: database.Text("org-chapter"),
		},
	}

	orgChapterIDs := []string{"chapter-id"}

	copiedTopics := []*entities.CopiedTopic{
		{
			CopyFromID: database.Text("org-topic-id"),
			ID:         database.Text("topic-id"),
		},
	}

	testCases := map[string]TestCase{
		"error finding book chapter": {
			ctx: ctx,
			req: &pb.DuplicateBookRequest{
				BookId:   "book-id",
				BookName: "book-name",
			},
			expectedErr: status.Error(codes.Internal, fmt.Errorf("cm.BookChapter.FindByBookIDs: %w", pgx.ErrNoRows).Error()),
			setup: func(ctx context.Context) {
				bookChapterRepo.On("FindByBookIDs", ctx, db, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
			},
		},
		"error duplicating book": {
			ctx: ctx,
			req: &pb.DuplicateBookRequest{
				BookId:   "book-id",
				BookName: "book-name",
			},
			expectedErr: status.Error(codes.Internal, fmt.Errorf("cm.BookRepo.DuplicateBook: %w", pgx.ErrNoRows).Error()),
			setup: func(ctx context.Context) {
				bookChapterRepo.On("FindByBookIDs", ctx, db, mock.Anything).Once().Return(bookChapterMap, nil)
				bookRepo.On("DuplicateBook", ctx, tx, mock.Anything, mock.Anything).Once().Return("", pgx.ErrNoRows)
				tx.On("Rollback", mock.Anything).Return(nil)

				db.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
			},
		},
		"err duplicate chapter": {
			ctx: ctx,
			req: &pb.DuplicateBookRequest{
				BookId:   "book-id",
				BookName: "book-name",
			},
			expectedErr: status.Error(codes.Internal, fmt.Errorf("cm.duplicateChapter: %w", status.Error(codes.Internal, fmt.Errorf("cm.ChapterRepo.DuplicateChapters: %w", pgx.ErrNoRows).Error())).Error()),
			setup: func(ctx context.Context) {
				bookChapterRepo.On("FindByBookIDs", ctx, db, mock.Anything).Once().Return(bookChapterMap, nil)
				bookRepo.On("DuplicateBook", ctx, tx, mock.Anything, mock.Anything).Once().Return("new-book-id", nil)
				chapterRepo.On("DuplicateChapters", ctx, tx, mock.Anything, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
				tx.On("Rollback", mock.Anything).Return(nil)

				db.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
			},
		},
		"err upsert chapter": {
			ctx: ctx,
			req: &pb.DuplicateBookRequest{
				BookId:   "book-id",
				BookName: "book-name",
			},
			expectedErr: status.Error(codes.Internal, fmt.Errorf("cm.duplicateChapter: %w", fmt.Errorf("cm.BookChapter.Upsert: %w", pgx.ErrNoRows)).Error()),
			setup: func(ctx context.Context) {
				bookChapterRepo.On("FindByBookIDs", ctx, db, mock.Anything).Once().Return(bookChapterMap, nil)
				bookRepo.On("DuplicateBook", ctx, tx, mock.Anything, mock.Anything).Once().Return("new-book-id", nil)
				chapterRepo.On("DuplicateChapters", ctx, tx, mock.Anything, mock.Anything).Once().Return(copiedChapter, nil)
				bookChapterRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(pgx.ErrNoRows)
				tx.On("Rollback", mock.Anything).Return(nil)

				db.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
			},
		},
		"err duplicate topic": {
			ctx: ctx,
			req: &pb.DuplicateBookRequest{
				BookId:   "book-id",
				BookName: "book-name",
			},
			expectedErr: status.Error(codes.Internal, fmt.Errorf("cm.duplicateTopics: %w", fmt.Errorf("cm.TopicRepo.DuplicateTopics: %w", pgx.ErrNoRows)).Error()),
			setup: func(ctx context.Context) {
				bookChapterRepo.On("FindByBookIDs", ctx, db, mock.Anything).Once().Return(bookChapterMap, nil)
				bookRepo.On("DuplicateBook", ctx, tx, mock.Anything, mock.Anything).Once().Return("new-book-id", nil)
				chapterRepo.On("DuplicateChapters", ctx, tx, mock.Anything, mock.Anything).Once().Return(copiedChapter, nil)
				bookChapterRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				topicRepo.On("DuplicateTopics", ctx, tx, mock.Anything, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
				tx.On("Rollback", mock.Anything).Return(nil)

				db.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
			},
		},
		"err duplicate learning objectives": {
			ctx: ctx,
			req: &pb.DuplicateBookRequest{
				BookId:   "book-id",
				BookName: "book-name",
			},
			expectedErr: status.Error(codes.Internal, fmt.Errorf("cm.LearningObjectiveRepo.DuplicateLearningObjectives: %w", pgx.ErrNoRows).Error()),
			setup: func(ctx context.Context) {
				bookChapterRepo.On("FindByBookIDs", ctx, db, mock.Anything).Once().Return(bookChapterMap, nil)
				bookRepo.On("DuplicateBook", ctx, tx, mock.Anything, mock.Anything).Once().Return("new-book-id", nil)
				chapterRepo.On("DuplicateChapters", ctx, tx, mock.Anything, mock.Anything).Once().Return(copiedChapter, nil)
				bookChapterRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				topicRepo.On("DuplicateTopics", ctx, tx, mock.Anything, mock.Anything).Once().Return(copiedTopics, nil)
				learningObjectiveRepo.On("DuplicateLearningObjectives", ctx, tx, mock.Anything, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
				tx.On("Rollback", mock.Anything).Return(nil)

				db.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
			},
		},
		"happy case": {
			ctx: ctx,
			req: &pb.DuplicateBookRequest{
				BookId:   "book-id",
				BookName: "book-name",
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				bookChapterRepo.On("FindByBookIDs", ctx, db, mock.Anything).Once().Return(bookChapterMap, nil)
				bookRepo.On("DuplicateBook", ctx, tx, mock.Anything, mock.Anything).Once().Return("new-book-id", nil)
				chapterRepo.On("DuplicateChapters", ctx, tx, "new-book-id", orgChapterIDs).Once().Return(copiedChapter, nil)
				bookChapterRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				topicRepo.On("DuplicateTopics", ctx, tx, mock.Anything, mock.Anything).Once().Return(copiedTopics, nil)
				learningObjectiveRepo.On("DuplicateLearningObjectives", ctx, tx, mock.Anything, mock.Anything).Once().Return([]*entities.CopiedLearningObjective{
					{
						CopiedLoID: database.Text("org-lo-id"),
						LoID:       database.Text("lo-id"),
					},
				}, nil)
				learningObjectiveRepo.On("RetrieveByTopicIDs", ctx, tx, mock.Anything).Once().Return([]*entities.LearningObjective{
					{
						TopicID:      database.Text("topic-id"),
						ID:           database.Text("id"),
						DisplayOrder: database.Int2(12),
					},
				}, nil)
				topicLearningObjectiveRepo.On("BulkImport", ctx, tx, mock.Anything).Once().Return(nil)

				assignmentRepo.On("DuplicateAssignment", ctx, tx, mock.Anything, mock.Anything).Once().Return(nil)
				assignmentRepo.On("RetrieveAssignmentsByTopicIDs", ctx, tx, mock.Anything).Once().Return([]*entities.Assignment{
					{},
				}, nil)
				topicAssignmentRepo.On("BulkUpsert", ctx, tx, mock.Anything).Once().Return(nil)

				tx.On("Rollback", mock.Anything).Return(nil)

				db.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
			},
		},
	}

	for name, testCase := range testCases {
		t.Log("Test case: " + name)
		testCase.setup(testCase.ctx)
		_, err := s.DuplicateBook(testCase.ctx, testCase.req.(*pb.DuplicateBookRequest))
		assert.Equal(t, testCase.expectedErr, err)
	}
}

func TestCourseModifierService_AddBooks(t *testing.T) {
	ctx := context.Background()
	db := &mock_database.Ext{}
	bookRepo := &mock_repositories.MockBookRepo{}
	courseBookRepo := &mock_repositories.MockCourseBookRepo{}
	s := &CourseModifierService{
		DB:             db,
		BookRepo:       bookRepo,
		CourseBookRepo: courseBookRepo,
	}

	bookIDs := []string{"1", "2", "3"}
	booksMap := map[string]*entities.Book{
		"1": {ID: database.Text("1")},
		"2": {ID: database.Text("2")},
		"3": {ID: database.Text("3")},
	}
	courseID := "1"
	testCases := []TestCase{
		{
			name:  "empty bookIDs",
			ctx:   ctx,
			setup: func(ctx context.Context) {},
			req: &pb.AddBooksRequest{
				BookIds: []string{},
			},
			expectedErr: status.Error(codes.InvalidArgument, "missing book id"),
		},
		{
			name:  "empty courseID",
			ctx:   ctx,
			setup: func(ctx context.Context) {},
			req: &pb.AddBooksRequest{
				BookIds:  bookIDs,
				CourseId: "",
			},
			expectedErr: status.Error(codes.InvalidArgument, "missing course id"),
		},
		{
			name: "non-existed bookIDs",
			ctx:  ctx,
			setup: func(ctx context.Context) {
				bookRepo.On("FindByIDs", ctx, db, []string{"1", "2", "3", "missing"}).Return(booksMap, nil)
			},
			req: &pb.AddBooksRequest{
				BookIds:  []string{"1", "2", "3", "missing"},
				CourseId: courseID,
			},
			expectedErr: status.Error(codes.NotFound, "not found books"),
		},
		{
			name: "upsert error",
			ctx:  ctx,
			setup: func(ctx context.Context) {
				bookRepo.On("FindByIDs", ctx, db, bookIDs).Return(booksMap, nil)
				courseBookRepo.On("Upsert", ctx, db, mock.Anything).Return(ErrSomethingWentWrong)
			},
			req: &pb.AddBooksRequest{
				BookIds:  bookIDs,
				CourseId: courseID,
			},
			expectedErr: status.Errorf(codes.Internal, "CourseBookRepo.Upsert: %s", ErrSomethingWentWrong.Error()),
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			req := testCase.req.(*pb.AddBooksRequest)
			resp, err := s.AddBooks(testCase.ctx, req)
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, testCase.expectedResp.(*pb.AddBooksResponse), resp)
			}
		})
	}
}

func TestCourseModifierService_SubmitQuizAnswers(t *testing.T) {
	ctx := context.Background()
	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	bookRepo := &mock_repositories.MockBookRepo{}
	courseBookRepo := &mock_repositories.MockCourseBookRepo{}
	shuffledQuizSetRepo := &mock_repositories.MockShuffledQuizSetRepo{}
	quizRepo := &mock_repositories.MockQuizRepo{}
	studentsLearningObjectivesCompletenessRepo := &mock_repositories.MockStudentsLearningObjectivesCompletenessRepo{}
	examLORepo := &mock_repositories.MockExamLORepo{}
	examLOSubmissionRepo := &mock_repositories.MockExamLOSubmissionRepo{}
	examLOSubmissionAnswerRepo := &mock_repositories.MockExamLOSubmissionAnswerRepo{}
	loProgressionRepo := &mock_repositories.MockLOProgressionRepo{}
	loProgressionAnswerRepo := &mock_repositories.MockLOProgressionAnswerRepo{}

	s := &CourseModifierService{
		DB:                         db,
		BookRepo:                   bookRepo,
		CourseBookRepo:             courseBookRepo,
		ShuffledQuizSetRepo:        shuffledQuizSetRepo,
		QuizRepo:                   quizRepo,
		ExamLORepo:                 examLORepo,
		ExamLOSubmissionRepo:       examLOSubmissionRepo,
		ExamLOSubmissionAnswerRepo: examLOSubmissionAnswerRepo,
		StudentsLearningObjectivesCompletenessRepo: studentsLearningObjectivesCompletenessRepo,
		LOProgressionRepo:                          loProgressionRepo,
		LOProgressionAnswerRepo:                    loProgressionAnswerRepo,
	}
	quizzes := getQuizzes(
		5,
		cpb.QuizType_QUIZ_TYPE_MCQ.String(),
		cpb.QuizType_QUIZ_TYPE_MAQ.String(),
		cpb.QuizType_QUIZ_TYPE_MIQ.String(),
		cpb.QuizType_QUIZ_TYPE_ORD.String(),
		cpb.QuizType_QUIZ_TYPE_ESQ.String(),
		//cpb.QuizType_QUIZ_TYPE_FIB.String(),  TODO: add later
		//cpb.QuizType_QUIZ_TYPE_POW.String(),
		//cpb.QuizType_QUIZ_TYPE_TAD.String(),
	)

	getExternalIDFromQuiz := func(qz entities.Quizzes) []string {
		res := make([]string, 0, len(qz))
		for _, q := range qz {
			res = append(res, q.ExternalID.String)
		}

		return res
	}
	totalPoint := func(qz entities.Quizzes) int32 {
		var res int32
		for _, q := range qz {
			res += q.Point.Int
		}
		return res
	}

	seed := int64(1)
	expectedRepoMethodCallForCheckCorrectness := func(expectedCTX context.Context, expectedDB database.Ext, qzs entities.Quizzes) {
		for i, quiz := range qzs {
			switch quiz.Kind.String {
			case cpb.QuizType_QUIZ_TYPE_MCQ.String(), cpb.QuizType_QUIZ_TYPE_MAQ.String():
				shuffledQuizSetRepo.On("GetSeed", expectedCTX, expectedDB, database.Text("shuffle-quiz-set-id")).
					Return(database.Text(fmt.Sprintf("%d", seed)), nil).Once()
				shuffledQuizSetRepo.On("GetQuizIdx", expectedCTX, expectedDB, database.Text("shuffle-quiz-set-id"), database.Text(quiz.ExternalID.String)).
					Return(database.Int4(int32(i)), nil).Once()
			case cpb.QuizType_QUIZ_TYPE_MIQ.String():
			case cpb.QuizType_QUIZ_TYPE_FIB.String(), cpb.QuizType_QUIZ_TYPE_POW.String(), cpb.QuizType_QUIZ_TYPE_TAD.String():
			case cpb.QuizType_QUIZ_TYPE_ORD.String(), cpb.QuizType_QUIZ_TYPE_ESQ.String():
			default:
				assert.Fail(t, fmt.Sprintf("quiz kind %s not yet is supported", quiz.Kind.String))
			}
		}
	}

	// for QuizType_QUIZ_TYPE_MCQ, QuizType_QUIZ_TYPE_MAQ type
	correctIndexAnswerBySeed := func(seed int64) []*pb.Answer {
		r := rand.New(rand.NewSource(seed))
		options := []int{1, 2, 3}
		r.Shuffle(len(options), func(i, j int) { options[i], options[j] = options[j], options[i] })

		res := make([]*pb.Answer, 0, 2)
		for i, opt := range options {
			if opt == 1 || opt == 2 {
				res = append(res, &pb.Answer{
					Format: &pb.Answer_SelectedIndex{SelectedIndex: uint32(i + 1)},
				})
			}
		}
		return res
	}

	// for QuizType_QUIZ_TYPE_MCQ, QuizType_QUIZ_TYPE_MAQ type
	returnCorrectIndexBySeed := func(seed int64) []uint32 {
		r := rand.New(rand.NewSource(seed))
		options := []int{1, 2, 3}
		r.Shuffle(len(options), func(i, j int) { options[i], options[j] = options[j], options[i] })

		res := make([]uint32, 0, 2)
		for i, opt := range options {
			if opt == 1 || opt == 2 {
				res = append(res, uint32(i+1))
			}
		}
		return res
	}

	// TODO: Please add expected for all these test cases include
	// args and return value for each repo method and response received,
	// if needed plz add more cases
	now := time.Date(2023, 01, 01, 0, 0, 0, 0, time.UTC)
	testCases := []TestCase{
		{
			name:  "empty set_id",
			ctx:   ctx,
			setup: func(ctx context.Context) {},
			req: &pb.SubmitQuizAnswersRequest{
				SetId: "",
			},
			expectedErr: status.Error(codes.InvalidArgument, "req must have SetId"),
		},
		{
			name:  "empty quiz_id",
			ctx:   ctx,
			setup: func(ctx context.Context) {},
			req: &pb.SubmitQuizAnswersRequest{
				SetId: "set-id",
				QuizAnswer: []*pb.QuizAnswer{
					{
						QuizId: "",
					},
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, "req must have QuizId"),
		},
		{
			name:  "without answer",
			ctx:   ctx,
			setup: func(ctx context.Context) {},
			req: &pb.SubmitQuizAnswersRequest{
				SetId: "set-id",
				QuizAnswer: []*pb.QuizAnswer{
					{
						QuizId: "quiz-id",
						Answer: nil,
					},
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, "req must have quizAnswer"),
		},
		{
			name:  "not a multiple choice neither fill in the blank",
			ctx:   ctx,
			setup: func(ctx context.Context) {},
			req: &pb.SubmitQuizAnswersRequest{
				SetId: "set-id",
				QuizAnswer: []*pb.QuizAnswer{
					{
						QuizId: "quiz-id",
						Answer: []*pb.Answer{
							{},
						},
					},
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, "your answer of quiz_id(quiz-id) is must not empty"),
		},
		{
			name: "happy case - trigger is executed",
			ctx:  ctx,
			setup: func(ctx context.Context) {
				shuffledQuizSetRepo.On("GetLoID", mock.Anything, mock.Anything, mock.Anything).Once().Return(database.Text("lo-id"), nil)
				quizRepo.On("GetByExternalIDs", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(entities.Quizzes{}, nil)
				shuffledQuizSetRepo.On("UpdateSubmissionHistory", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				shuffledQuizSetRepo.On("UpdateTotalCorrectness", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				shuffledQuizSetRepo.On("IsFinishedQuizTest", mock.Anything, mock.Anything, mock.Anything).Once().Return(database.Bool(false), nil)
				shuffledQuizSetRepo.On("Get", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(&entities.ShuffledQuizSet{}, nil)
				examLOSubmissionRepo.On("Get", mock.Anything, mock.Anything, mock.Anything).Once().Return(&entities.ExamLOSubmission{
					SubmissionID:      pgtype.Text{Status: pgtype.Null},
					ShuffledQuizSetID: database.Text("shuffled_quiz_set_id"),
				}, nil)
				examLORepo.On("Get", mock.Anything, mock.Anything, mock.Anything).Once().Return(&entities.ExamLO{}, nil)
				shuffledQuizSetRepo.On("GetExternalIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(pgtype.TextArray{}, nil)
				examLOSubmissionRepo.On("GetTotalGradedPoint", mock.Anything, mock.Anything, mock.Anything).Once().Return(database.Int4(2), nil)
				examLOSubmissionRepo.On("Update", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				loProgressionRepo.On("DeleteByStudyPlanIdentity", mock.Anything, mock.Anything, mock.Anything).Once().Return(int64(0), nil)
				loProgressionAnswerRepo.On("DeleteByStudyPlanIdentity", mock.Anything, mock.Anything, mock.Anything).Once().Return(int64(0), nil)

				db.On("Begin", ctx).Return(tx, nil).Times(3)
				tx.On("Commit", mock.Anything).Return(nil).Times(3)
			},
			req: &pb.SubmitQuizAnswersRequest{
				SetId: "set-id",
				QuizAnswer: []*pb.QuizAnswer{
					{
						QuizId: "quiz-id",
						Answer: []*pb.Answer{
							{
								Format: &pb.Answer_FilledText{},
							},
						},
					},
				},
			},
			expectedErr: nil,
		},
		{
			name: "happy case",
			ctx:  ctx,
			setup: func(ctx context.Context) {
				shuffledQuizSetRepo.On("GetLoID", mock.Anything, mock.Anything, mock.Anything).Once().Return(database.Text("lo-id"), nil)
				quizRepo.On("GetByExternalIDs", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(entities.Quizzes{}, nil)
				shuffledQuizSetRepo.On("UpdateSubmissionHistory", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				shuffledQuizSetRepo.On("UpdateTotalCorrectness", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				shuffledQuizSetRepo.On("IsFinishedQuizTest", mock.Anything, mock.Anything, mock.Anything).Once().Return(database.Bool(false), nil)
				shuffledQuizSetRepo.On("Get", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(&entities.ShuffledQuizSet{}, nil)
				examLOSubmissionRepo.On("Get", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, nil)
				examLORepo.On("Get", mock.Anything, mock.Anything, mock.Anything).Once().Return(&entities.ExamLO{}, nil)
				shuffledQuizSetRepo.On("GetExternalIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(pgtype.TextArray{}, nil)
				shuffledQuizSetRepo.On("GenerateExamLOSubmission", mock.Anything, mock.Anything, mock.Anything).Once().Return(&entities.ExamLOSubmission{}, nil)
				examLOSubmissionRepo.On("Insert", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				examLOSubmissionRepo.On("GetTotalGradedPoint", mock.Anything, mock.Anything, mock.Anything).Once().Return(database.Int4(2), nil)
				examLOSubmissionRepo.On("Update", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				loProgressionRepo.On("DeleteByStudyPlanIdentity", mock.Anything, mock.Anything, mock.Anything).Once().Return(int64(0), nil)
				loProgressionAnswerRepo.On("DeleteByStudyPlanIdentity", mock.Anything, mock.Anything, mock.Anything).Once().Return(int64(0), nil)

				db.On("Begin", ctx).Return(tx, nil).Times(3)
				tx.On("Commit", mock.Anything).Return(nil).Times(3)
			},
			req: &pb.SubmitQuizAnswersRequest{
				SetId: "set-id",
				QuizAnswer: []*pb.QuizAnswer{
					{
						QuizId: "quiz-id",
						Answer: []*pb.Answer{
							{
								Format: &pb.Answer_FilledText{},
							},
						},
					},
				},
			},
			expectedErr: nil,
		},

		{
			name: "submit correct answer all quiz type",
			ctx:  ctx,
			setup: func(ctx context.Context) {
				shuffledQuizSetRepo.On("GetLoID", ctx, db, database.Text("shuffle-quiz-set-id")).
					Return(database.Text("lo-id"), nil).Once()
				quizRepo.On("GetByExternalIDs", ctx, db, database.TextArray(getExternalIDFromQuiz(quizzes)), database.Text("lo-id")).
					Return(quizzes, nil).Once()
				expectedRepoMethodCallForCheckCorrectness(ctx, db, quizzes)

				shuffledQuizSetRepo.On("UpdateSubmissionHistory", ctx, tx, database.Text("shuffle-quiz-set-id"), mock.Anything).
					Run(func(args mock.Arguments) {
						data := args[3].(pgtype.JSONB)
						var answersEnt []*entities.QuizAnswer
						err := data.AssignTo(&answersEnt)
						require.NoError(t, err)
						assert.Len(t, answersEnt, len(quizzes))

						// gen from quizzes
						expectedAnswersEnt := []*entities.QuizAnswer{
							{
								QuizID:       quizzes[0].ExternalID.String,
								QuizType:     quizzes[0].Kind.String,
								Correctness:  []bool{true, true},
								IsAccepted:   true,
								IsAllCorrect: true,
								Point:        uint32(quizzes[0].Point.Int),
							},
							{
								QuizID:       quizzes[1].ExternalID.String,
								QuizType:     quizzes[1].Kind.String,
								Correctness:  []bool{true, true},
								IsAccepted:   true,
								IsAllCorrect: true,
								Point:        uint32(quizzes[1].Point.Int),
							},
							{
								QuizID:       quizzes[2].ExternalID.String,
								QuizType:     quizzes[2].Kind.String,
								Correctness:  []bool{true, true},
								IsAccepted:   true,
								IsAllCorrect: true,
								Point:        uint32(quizzes[2].Point.Int),
							},
							{
								QuizID:       quizzes[3].ExternalID.String,
								QuizType:     quizzes[3].Kind.String,
								Correctness:  []bool{true, true, true},
								IsAccepted:   true,
								IsAllCorrect: true,
								Point:        uint32(quizzes[3].Point.Int),
							},
							{
								QuizID:     quizzes[4].ExternalID.String,
								QuizType:   quizzes[4].Kind.String,
								FilledText: []string{"sample-text"},
							},
						}
						for i, ans := range answersEnt {
							assert.NotZero(t, ans.SubmittedAt)
							switch quizzes[i].Kind.String {
							case cpb.QuizType_QUIZ_TYPE_MCQ.String(), cpb.QuizType_QUIZ_TYPE_MAQ.String(), cpb.QuizType_QUIZ_TYPE_MIQ.String():
								assert.Len(t, ans.SelectedIndex, 2)
								assert.Len(t, ans.CorrectIndex, 2)
							case cpb.QuizType_QUIZ_TYPE_ORD.String():
								assert.Len(t, ans.SubmittedKeys, 3)
								assert.Len(t, ans.CorrectKeys, 3)
							}
							expectedAnswersEnt[i].SelectedIndex = ans.SelectedIndex
							expectedAnswersEnt[i].CorrectIndex = ans.CorrectIndex
							expectedAnswersEnt[i].FilledText = ans.FilledText
							expectedAnswersEnt[i].CorrectText = ans.CorrectText
							expectedAnswersEnt[i].SubmittedKeys = ans.SubmittedKeys
							expectedAnswersEnt[i].CorrectKeys = ans.CorrectKeys
							expectedAnswersEnt[i].SubmittedAt = ans.SubmittedAt
							assert.Equal(t, expectedAnswersEnt[i], ans)
						}
					}).
					Return(nil).Once()
				shuffledQuizSetRepo.On("UpdateTotalCorrectness", ctx, tx, database.Text("shuffle-quiz-set-id")).
					Once().Return(nil)
				shuffledQuizSetRepo.On("IsFinishedQuizTest", ctx, tx, database.Text("shuffle-quiz-set-id")).
					Once().Return(database.Bool(true), nil)
				shuffledQuizSetRepo.On("Get", ctx, tx, database.Text("shuffle-quiz-set-id"), database.Int8(1), database.Int8(1)).
					Once().Return(&entities.ShuffledQuizSet{}, nil)
				shuffledQuizSetRepo.On("GetStudentID", ctx, tx, database.Text("shuffle-quiz-set-id")).
					Once().Return(database.Text("student-id-0"), nil)
				shuffledQuizSetRepo.On("GetScore", ctx, tx, database.Text("shuffle-quiz-set-id")).
					Once().Return(database.Int4(int32(len(quizzes))), database.Int4(int32(len(quizzes))), nil)

				studentsLearningObjectivesCompletenessRepo.On("UpsertFirstQuizCompleteness", ctx, tx, database.Text("lo-id"), database.Text("student-id-0"), database.Float4(100)).
					Once().Return(nil)
				studentsLearningObjectivesCompletenessRepo.On("UpsertHighestQuizScore", ctx, tx, database.Text("lo-id"), database.Text("student-id-0"), database.Float4(100)).
					Once().Return(nil)

				examLOSubmissionRepo.
					On("Get", ctx, db, &repositories.GetExamLOSubmissionArgs{
						SubmissionID:      pgtype.Text{Status: pgtype.Null},
						ShuffledQuizSetID: database.Text("shuffle-quiz-set-id"),
					}).
					Once().Return(nil, pgx.ErrNoRows)
				shuffledQuizSetRepo.On("GetExternalIDs", ctx, db, database.Text("shuffle-quiz-set-id")).
					Once().Return(database.TextArray(getExternalIDFromQuiz(quizzes)), nil)
				shuffledQuizSetRepo.On("GenerateExamLOSubmission", ctx, tx, database.Text("shuffle-quiz-set-id")).
					Once().
					Return(
						&entities.ExamLOSubmission{
							BaseEntity: entities.BaseEntity{
								CreatedAt: database.Timestamptz(now),
								UpdatedAt: database.Timestamptz(now),
							},
							SubmissionID:       database.Text("submission-id-0"),
							StudentID:          database.Text("student-id-0"),
							StudyPlanID:        database.Text("study-plan-id-0"),
							LearningMaterialID: database.Text("learning-material-id-0"),
							ShuffledQuizSetID:  database.Text("shuffle-quiz-set-id"),
							TotalPoint:         database.Int4(totalPoint(quizzes)),
						},
						nil,
					)
				examLOSubmissionRepo.On("Insert", ctx, tx, &entities.ExamLOSubmission{
					BaseEntity: entities.BaseEntity{
						CreatedAt: database.Timestamptz(now),
						UpdatedAt: database.Timestamptz(now),
					},
					SubmissionID:       database.Text("submission-id-0"),
					StudentID:          database.Text("student-id-0"),
					StudyPlanID:        database.Text("study-plan-id-0"),
					LearningMaterialID: database.Text("learning-material-id-0"),
					ShuffledQuizSetID:  database.Text("shuffle-quiz-set-id"),
					TotalPoint:         database.Int4(totalPoint(quizzes)),
				}).Once().Return(nil)
				examLOSubmissionAnswerRepo.On("Upsert", ctx, tx, mock.Anything).
					Run(func(args mock.Arguments) {
						answer := args[2].(*entities.ExamLOSubmissionAnswer)
						if answer.QuizID.String == quizzes[3].ExternalID.String {
							assert.Equal(t, []string{"key-1", "key-2", "key-3"}, database.FromTextArray(answer.SubmittedKeysAnswer))
							assert.Equal(t, []string{"key-1", "key-2", "key-3"}, database.FromTextArray(answer.CorrectKeysAnswer))
						}
						// TODO: add expected for another fields
					}).Return(len(quizzes), nil).Times(len(quizzes))
				examLORepo.On("Get", ctx, tx, database.Text("learning-material-id-0")).
					Once().Return(&entities.ExamLO{
					GradeToPass: database.Int4(totalPoint(quizzes)),
				}, nil)
				examLOSubmissionRepo.On("GetTotalGradedPoint", ctx, tx, database.Text("submission-id-0")).
					Once().Return(database.Int4(totalPoint(quizzes)), nil)
				examLOSubmissionRepo.On("Update", ctx, tx, mock.Anything).
					Run(func(args mock.Arguments) {
						examLOSubmission := args[2].(*entities.ExamLOSubmission)
						examLOSubmission.CreatedAt = database.Timestamptz(now)
						examLOSubmission.UpdatedAt = database.Timestamptz(now)
						expected := &entities.ExamLOSubmission{
							BaseEntity: entities.BaseEntity{
								CreatedAt: database.Timestamptz(now),
								UpdatedAt: database.Timestamptz(now),
							},
							SubmissionID:       database.Text("submission-id-0"),
							StudentID:          database.Text("student-id-0"),
							StudyPlanID:        database.Text("study-plan-id-0"),
							LearningMaterialID: database.Text("learning-material-id-0"),
							ShuffledQuizSetID:  database.Text("shuffle-quiz-set-id"),
							TotalPoint:         database.Int4(totalPoint(quizzes)),
							Result:             database.Text(pb.ExamLOSubmissionResult_EXAM_LO_SUBMISSION_PASSED.String()),
							Status:             database.Text(pb.SubmissionStatus_SUBMISSION_STATUS_RETURNED.String()),
						}
						assert.Equal(t, expected, examLOSubmission)
					}).Once().Return(nil)
				loProgressionRepo.On("DeleteByStudyPlanIdentity", mock.Anything, mock.Anything, repositories.StudyPlanItemIdentity{
					StudentID:          database.Text("student-id-0"),
					StudyPlanID:        database.Text("study-plan-id-0"),
					LearningMaterialID: database.Text("learning-material-id-0"),
				}).Once().Return(int64(0), nil)
				loProgressionAnswerRepo.On("DeleteByStudyPlanIdentity", mock.Anything, mock.Anything, repositories.StudyPlanItemIdentity{
					StudentID:          database.Text("student-id-0"),
					StudyPlanID:        database.Text("study-plan-id-0"),
					LearningMaterialID: database.Text("learning-material-id-0"),
				}).Once().Return(int64(0), nil)

				db.On("Begin", ctx).Return(tx, nil).Times(3)
				tx.On("Commit", ctx).Return(nil).Times(3)
			},
			req: &pb.SubmitQuizAnswersRequest{
				SetId: "shuffle-quiz-set-id",
				QuizAnswer: []*pb.QuizAnswer{
					{
						QuizId: quizzes[0].ExternalID.String,
						Answer: correctIndexAnswerBySeed(seed),
					},
					{
						QuizId: quizzes[1].ExternalID.String,
						Answer: correctIndexAnswerBySeed(seed + 1),
					},
					{
						QuizId: quizzes[2].ExternalID.String,
						Answer: []*pb.Answer{
							{Format: &pb.Answer_SelectedIndex{SelectedIndex: 1}},
							{Format: &pb.Answer_SelectedIndex{SelectedIndex: 2}},
						},
					},
					{
						QuizId: quizzes[3].ExternalID.String,
						Answer: []*pb.Answer{
							{Format: &pb.Answer_SubmittedKey{SubmittedKey: "key-1"}},
							{Format: &pb.Answer_SubmittedKey{SubmittedKey: "key-2"}},
							{Format: &pb.Answer_SubmittedKey{SubmittedKey: "key-3"}},
						},
					},
					{
						QuizId: quizzes[4].ExternalID.String,
						Answer: []*pb.Answer{
							{Format: &pb.Answer_FilledText{FilledText: "sample-text"}},
						},
					},
				},
			},
			expectedResp: &pb.SubmitQuizAnswersResponse{
				Logs: []*cpb.AnswerLog{
					{
						QuizId:        quizzes[0].ExternalID.String,
						QuizType:      cpb.QuizType(cpb.QuizType_value[quizzes[0].Kind.String]),
						SelectedIndex: returnCorrectIndexBySeed(seed),
						CorrectIndex:  returnCorrectIndexBySeed(seed),
						Correctness:   []bool{true, true},
						IsAccepted:    true,
					},
					{
						QuizId:        quizzes[1].ExternalID.String,
						QuizType:      cpb.QuizType(cpb.QuizType_value[quizzes[1].Kind.String]),
						SelectedIndex: returnCorrectIndexBySeed(seed + 1),
						CorrectIndex:  returnCorrectIndexBySeed(seed + 1),
						Correctness:   []bool{true, true},
						IsAccepted:    true,
					},
					{
						QuizId:        quizzes[2].ExternalID.String,
						QuizType:      cpb.QuizType(cpb.QuizType_value[quizzes[2].Kind.String]),
						SelectedIndex: []uint32{1, 2},
						CorrectIndex:  []uint32{1, 2},
						Correctness:   []bool{true, true},
						IsAccepted:    true,
					},
					{
						QuizId:   quizzes[3].ExternalID.String,
						QuizType: cpb.QuizType(cpb.QuizType_value[quizzes[3].Kind.String]),
						Result: &cpb.AnswerLog_OrderingResult{OrderingResult: &cpb.OrderingResult{
							SubmittedKeys: []string{"key-1", "key-2", "key-3"},
							CorrectKeys:   []string{"key-1", "key-2", "key-3"},
						}},
						Correctness: []bool{true, true, true},
						IsAccepted:  true,
					},
					{
						QuizId:     quizzes[4].ExternalID.String,
						QuizType:   cpb.QuizType(cpb.QuizType_value[quizzes[4].Kind.String]),
						FilledText: []string{"sample-text"},
					},
				},
				TotalGradedPoint:   wrapperspb.UInt32(uint32(totalPoint(quizzes))),
				TotalPoint:         wrapperspb.UInt32(uint32(totalPoint(quizzes))),
				TotalCorrectAnswer: int32(len(quizzes) - 1),
				TotalQuestion:      int32(len(quizzes)),
				SubmissionResult:   pb.ExamLOSubmissionResult_EXAM_LO_SUBMISSION_PASSED,
			},
			expectedErr: nil,
		},
		{
			name: "submit correct answer ordering quiz",
			ctx:  ctx,
			setup: func(ctx context.Context) {
				quiz := quizzes[3:4]
				shuffledQuizSetRepo.On("GetLoID", ctx, db, database.Text("shuffle-quiz-set-id")).
					Return(database.Text("lo-id"), nil).Once()
				quizRepo.On("GetByExternalIDs", ctx, db, database.TextArray(getExternalIDFromQuiz(quiz)), database.Text("lo-id")).
					Return(quiz, nil).Once()
				expectedRepoMethodCallForCheckCorrectness(ctx, db, quiz)

				shuffledQuizSetRepo.On("UpdateSubmissionHistory", ctx, tx, database.Text("shuffle-quiz-set-id"), mock.Anything).
					Run(func(args mock.Arguments) {
						data := args[3].(pgtype.JSONB)
						var answersEnt []*entities.QuizAnswer
						err := data.AssignTo(&answersEnt)
						require.NoError(t, err)
						assert.Len(t, answersEnt, len(quiz))

						// gen from quizzes
						expectedAnswersEnt := []*entities.QuizAnswer{
							{
								QuizID:        quizzes[3].ExternalID.String,
								QuizType:      quizzes[3].Kind.String,
								SubmittedKeys: []string{"key-1", "key-2", "key-3"},
								CorrectKeys:   []string{"key-1", "key-2", "key-3"},
								Correctness:   []bool{true, true, true},
								IsAccepted:    true,
								IsAllCorrect:  true,
								Point:         uint32(quizzes[3].Point.Int),
							},
						}
						for i, ans := range answersEnt {
							assert.NotZero(t, ans.SubmittedAt)
							assert.Equal(t, cpb.QuizType_QUIZ_TYPE_ORD.String(), ans.QuizType)
							assert.Len(t, ans.SubmittedKeys, 3)
							assert.Len(t, ans.CorrectKeys, 3)
							expectedAnswersEnt[i].SubmittedAt = ans.SubmittedAt
							assert.Equal(t, expectedAnswersEnt[i], ans)
						}
					}).
					Return(nil).Once()
				shuffledQuizSetRepo.On("UpdateTotalCorrectness", ctx, tx, database.Text("shuffle-quiz-set-id")).
					Once().Return(nil)
				shuffledQuizSetRepo.On("IsFinishedQuizTest", ctx, tx, database.Text("shuffle-quiz-set-id")).
					Once().Return(database.Bool(true), nil)
				shuffledQuizSetRepo.On("Get", ctx, tx, database.Text("shuffle-quiz-set-id"), database.Int8(1), database.Int8(1)).
					Once().Return(&entities.ShuffledQuizSet{}, nil)
				shuffledQuizSetRepo.On("GetStudentID", ctx, tx, database.Text("shuffle-quiz-set-id")).
					Once().Return(database.Text("student-id-0"), nil)
				shuffledQuizSetRepo.On("GetScore", ctx, tx, database.Text("shuffle-quiz-set-id")).
					Once().Return(database.Int4(int32(len(quiz))), database.Int4(int32(len(quiz))), nil)

				studentsLearningObjectivesCompletenessRepo.On("UpsertFirstQuizCompleteness", ctx, tx, database.Text("lo-id"), database.Text("student-id-0"), database.Float4(100)).
					Once().Return(nil)
				studentsLearningObjectivesCompletenessRepo.On("UpsertHighestQuizScore", ctx, tx, database.Text("lo-id"), database.Text("student-id-0"), database.Float4(100)).
					Once().Return(nil)

				examLOSubmissionRepo.
					On("Get", ctx, db, &repositories.GetExamLOSubmissionArgs{
						SubmissionID:      pgtype.Text{Status: pgtype.Null},
						ShuffledQuizSetID: database.Text("shuffle-quiz-set-id"),
					}).
					Once().Return(nil, pgx.ErrNoRows)
				shuffledQuizSetRepo.On("GetExternalIDs", ctx, db, database.Text("shuffle-quiz-set-id")).
					Once().Return(database.TextArray(getExternalIDFromQuiz(quiz)), nil)
				shuffledQuizSetRepo.On("GenerateExamLOSubmission", ctx, tx, database.Text("shuffle-quiz-set-id")).
					Once().
					Return(
						&entities.ExamLOSubmission{
							BaseEntity: entities.BaseEntity{
								CreatedAt: database.Timestamptz(now),
								UpdatedAt: database.Timestamptz(now),
							},
							SubmissionID:       database.Text("submission-id-0"),
							StudentID:          database.Text("student-id-0"),
							StudyPlanID:        database.Text("study-plan-id-0"),
							LearningMaterialID: database.Text("learning-material-id-0"),
							ShuffledQuizSetID:  database.Text("shuffle-quiz-set-id"),
							TotalPoint:         database.Int4(totalPoint(quiz)),
						},
						nil,
					)
				examLOSubmissionRepo.On("Insert", ctx, tx, &entities.ExamLOSubmission{
					BaseEntity: entities.BaseEntity{
						CreatedAt: database.Timestamptz(now),
						UpdatedAt: database.Timestamptz(now),
					},
					SubmissionID:       database.Text("submission-id-0"),
					StudentID:          database.Text("student-id-0"),
					StudyPlanID:        database.Text("study-plan-id-0"),
					LearningMaterialID: database.Text("learning-material-id-0"),
					ShuffledQuizSetID:  database.Text("shuffle-quiz-set-id"),
					TotalPoint:         database.Int4(totalPoint(quiz)),
				}).Once().Return(nil)
				examLOSubmissionAnswerRepo.On("Upsert", ctx, tx, mock.Anything).
					Run(func(args mock.Arguments) {
						answer := args[2].(*entities.ExamLOSubmissionAnswer)
						assert.Equal(t, []string{"key-1", "key-2", "key-3"}, database.FromTextArray(answer.SubmittedKeysAnswer))
						assert.Equal(t, []string{"key-1", "key-2", "key-3"}, database.FromTextArray(answer.CorrectKeysAnswer))
						// TODO: add expected for another fields
					}).Return(len(quiz), nil).Times(len(quiz))
				examLORepo.On("Get", ctx, tx, database.Text("learning-material-id-0")).
					Once().Return(&entities.ExamLO{
					GradeToPass: database.Int4(totalPoint(quiz)),
				}, nil)
				examLOSubmissionRepo.On("GetTotalGradedPoint", ctx, tx, database.Text("submission-id-0")).
					Once().Return(database.Int4(totalPoint(quiz)), nil)
				examLOSubmissionRepo.On("Update", ctx, tx, mock.Anything).
					Run(func(args mock.Arguments) {
						examLOSubmission := args[2].(*entities.ExamLOSubmission)
						examLOSubmission.CreatedAt = database.Timestamptz(now)
						examLOSubmission.UpdatedAt = database.Timestamptz(now)
						expected := &entities.ExamLOSubmission{
							BaseEntity: entities.BaseEntity{
								CreatedAt: database.Timestamptz(now),
								UpdatedAt: database.Timestamptz(now),
							},
							SubmissionID:       database.Text("submission-id-0"),
							StudentID:          database.Text("student-id-0"),
							StudyPlanID:        database.Text("study-plan-id-0"),
							LearningMaterialID: database.Text("learning-material-id-0"),
							ShuffledQuizSetID:  database.Text("shuffle-quiz-set-id"),
							TotalPoint:         database.Int4(totalPoint(quiz)),
							Result:             database.Text(pb.ExamLOSubmissionResult_EXAM_LO_SUBMISSION_PASSED.String()),
							Status:             database.Text(pb.SubmissionStatus_SUBMISSION_STATUS_RETURNED.String()),
						}
						assert.Equal(t, expected, examLOSubmission)
					}).Once().Return(nil)
				loProgressionRepo.On("DeleteByStudyPlanIdentity", mock.Anything, mock.Anything, repositories.StudyPlanItemIdentity{
					StudentID:          database.Text("student-id-0"),
					StudyPlanID:        database.Text("study-plan-id-0"),
					LearningMaterialID: database.Text("learning-material-id-0"),
				}).Once().Return(int64(0), nil)
				loProgressionAnswerRepo.On("DeleteByStudyPlanIdentity", mock.Anything, mock.Anything, repositories.StudyPlanItemIdentity{
					StudentID:          database.Text("student-id-0"),
					StudyPlanID:        database.Text("study-plan-id-0"),
					LearningMaterialID: database.Text("learning-material-id-0"),
				}).Once().Return(int64(0), nil)

				db.On("Begin", ctx).Return(tx, nil).Times(3)
				tx.On("Commit", ctx).Return(nil).Times(3)
			},
			req: &pb.SubmitQuizAnswersRequest{
				SetId: "shuffle-quiz-set-id",
				QuizAnswer: []*pb.QuizAnswer{
					{
						QuizId: quizzes[3].ExternalID.String,
						Answer: []*pb.Answer{
							{Format: &pb.Answer_SubmittedKey{SubmittedKey: "key-1"}},
							{Format: &pb.Answer_SubmittedKey{SubmittedKey: "key-2"}},
							{Format: &pb.Answer_SubmittedKey{SubmittedKey: "key-3"}},
						},
					},
				},
			},
			expectedResp: &pb.SubmitQuizAnswersResponse{
				Logs: []*cpb.AnswerLog{
					{
						QuizId:   quizzes[3].ExternalID.String,
						QuizType: cpb.QuizType(cpb.QuizType_value[quizzes[3].Kind.String]),
						Result: &cpb.AnswerLog_OrderingResult{OrderingResult: &cpb.OrderingResult{
							SubmittedKeys: []string{"key-1", "key-2", "key-3"},
							CorrectKeys:   []string{"key-1", "key-2", "key-3"},
						}},
						Correctness: []bool{true, true, true},
						IsAccepted:  true,
					},
				},
				TotalGradedPoint:   wrapperspb.UInt32(uint32(totalPoint(quizzes[3:4]))),
				TotalPoint:         wrapperspb.UInt32(uint32(totalPoint(quizzes[3:4]))),
				TotalCorrectAnswer: int32(len(quizzes[3:4])),
				TotalQuestion:      int32(len(quizzes[3:4])),
				SubmissionResult:   pb.ExamLOSubmissionResult_EXAM_LO_SUBMISSION_PASSED,
			},
			expectedErr: nil,
		},
		{
			name: "submit incorrect answer ordering quiz",
			ctx:  ctx,
			setup: func(ctx context.Context) {
				quiz := quizzes[3:4]
				shuffledQuizSetRepo.On("GetLoID", ctx, db, database.Text("shuffle-quiz-set-id")).
					Return(database.Text("lo-id"), nil).Once()
				quizRepo.On("GetByExternalIDs", ctx, db, database.TextArray(getExternalIDFromQuiz(quiz)), database.Text("lo-id")).
					Return(quiz, nil).Once()
				expectedRepoMethodCallForCheckCorrectness(ctx, db, quiz)

				shuffledQuizSetRepo.On("UpdateSubmissionHistory", ctx, tx, database.Text("shuffle-quiz-set-id"), mock.Anything).
					Run(func(args mock.Arguments) {
						data := args[3].(pgtype.JSONB)
						var answersEnt []*entities.QuizAnswer
						err := data.AssignTo(&answersEnt)
						require.NoError(t, err)
						assert.Len(t, answersEnt, len(quiz))

						// gen from quizzes
						expectedAnswersEnt := []*entities.QuizAnswer{
							{
								QuizID:        quizzes[3].ExternalID.String,
								QuizType:      quizzes[3].Kind.String,
								SubmittedKeys: []string{"key-2", "key-1", "key-3"},
								CorrectKeys:   []string{"key-1", "key-2", "key-3"},
								Correctness:   []bool{false, false, true},
								IsAccepted:    false,
								IsAllCorrect:  false,
								Point:         0,
							},
						}
						for i, ans := range answersEnt {
							assert.NotZero(t, ans.SubmittedAt)
							assert.Equal(t, cpb.QuizType_QUIZ_TYPE_ORD.String(), ans.QuizType)
							assert.Len(t, ans.SubmittedKeys, 3)
							assert.Len(t, ans.CorrectKeys, 3)
							expectedAnswersEnt[i].SubmittedAt = ans.SubmittedAt
							assert.Equal(t, expectedAnswersEnt[i], ans)
						}
					}).
					Return(nil).Once()
				shuffledQuizSetRepo.On("UpdateTotalCorrectness", ctx, tx, database.Text("shuffle-quiz-set-id")).
					Once().Return(nil)
				shuffledQuizSetRepo.On("IsFinishedQuizTest", ctx, tx, database.Text("shuffle-quiz-set-id")).
					Once().Return(database.Bool(true), nil)
				shuffledQuizSetRepo.On("Get", ctx, tx, database.Text("shuffle-quiz-set-id"), database.Int8(1), database.Int8(1)).
					Once().Return(&entities.ShuffledQuizSet{}, nil)
				shuffledQuizSetRepo.On("GetStudentID", ctx, tx, database.Text("shuffle-quiz-set-id")).
					Once().Return(database.Text("student-id-0"), nil)
				shuffledQuizSetRepo.On("GetScore", ctx, tx, database.Text("shuffle-quiz-set-id")).
					Once().Return(database.Int4(int32(len(quiz))), database.Int4(int32(len(quiz))), nil)

				studentsLearningObjectivesCompletenessRepo.On("UpsertFirstQuizCompleteness", ctx, tx, database.Text("lo-id"), database.Text("student-id-0"), database.Float4(100)).
					Once().Return(nil)
				studentsLearningObjectivesCompletenessRepo.On("UpsertHighestQuizScore", ctx, tx, database.Text("lo-id"), database.Text("student-id-0"), database.Float4(100)).
					Once().Return(nil)

				examLOSubmissionRepo.
					On("Get", ctx, db, &repositories.GetExamLOSubmissionArgs{
						SubmissionID:      pgtype.Text{Status: pgtype.Null},
						ShuffledQuizSetID: database.Text("shuffle-quiz-set-id"),
					}).
					Once().Return(nil, pgx.ErrNoRows)
				shuffledQuizSetRepo.On("GetExternalIDs", ctx, db, database.Text("shuffle-quiz-set-id")).
					Once().Return(database.TextArray(getExternalIDFromQuiz(quiz)), nil)
				shuffledQuizSetRepo.On("GenerateExamLOSubmission", ctx, tx, database.Text("shuffle-quiz-set-id")).
					Once().
					Return(
						&entities.ExamLOSubmission{
							BaseEntity: entities.BaseEntity{
								CreatedAt: database.Timestamptz(now),
								UpdatedAt: database.Timestamptz(now),
							},
							SubmissionID:       database.Text("submission-id-0"),
							StudentID:          database.Text("student-id-0"),
							StudyPlanID:        database.Text("study-plan-id-0"),
							LearningMaterialID: database.Text("learning-material-id-0"),
							ShuffledQuizSetID:  database.Text("shuffle-quiz-set-id"),
							TotalPoint:         database.Int4(totalPoint(quiz)),
						},
						nil,
					)
				examLOSubmissionRepo.On("Insert", ctx, tx, &entities.ExamLOSubmission{
					BaseEntity: entities.BaseEntity{
						CreatedAt: database.Timestamptz(now),
						UpdatedAt: database.Timestamptz(now),
					},
					SubmissionID:       database.Text("submission-id-0"),
					StudentID:          database.Text("student-id-0"),
					StudyPlanID:        database.Text("study-plan-id-0"),
					LearningMaterialID: database.Text("learning-material-id-0"),
					ShuffledQuizSetID:  database.Text("shuffle-quiz-set-id"),
					TotalPoint:         database.Int4(totalPoint(quiz)),
				}).Once().Return(nil)
				// TODO: add expected for last arg
				examLOSubmissionAnswerRepo.On("Upsert", ctx, tx, mock.Anything).
					Run(func(args mock.Arguments) {
						answer := args[2].(*entities.ExamLOSubmissionAnswer)
						assert.Equal(t, []string{"key-2", "key-1", "key-3"}, database.FromTextArray(answer.SubmittedKeysAnswer))
						assert.Equal(t, []string{"key-1", "key-2", "key-3"}, database.FromTextArray(answer.CorrectKeysAnswer))
						// TODO: add expected for another fields
					}).Return(len(quiz), nil).Times(len(quiz))
				examLORepo.On("Get", ctx, tx, database.Text("learning-material-id-0")).
					Once().Return(&entities.ExamLO{
					GradeToPass: database.Int4(totalPoint(quiz)),
				}, nil)
				examLOSubmissionRepo.On("GetTotalGradedPoint", ctx, tx, database.Text("submission-id-0")).
					Once().Return(database.Int4(0), nil)
				examLOSubmissionRepo.On("Update", ctx, tx, mock.Anything).
					Run(func(args mock.Arguments) {
						examLOSubmission := args[2].(*entities.ExamLOSubmission)
						examLOSubmission.CreatedAt = database.Timestamptz(now)
						examLOSubmission.UpdatedAt = database.Timestamptz(now)
						expected := &entities.ExamLOSubmission{
							BaseEntity: entities.BaseEntity{
								CreatedAt: database.Timestamptz(now),
								UpdatedAt: database.Timestamptz(now),
							},
							SubmissionID:       database.Text("submission-id-0"),
							StudentID:          database.Text("student-id-0"),
							StudyPlanID:        database.Text("study-plan-id-0"),
							LearningMaterialID: database.Text("learning-material-id-0"),
							ShuffledQuizSetID:  database.Text("shuffle-quiz-set-id"),
							TotalPoint:         database.Int4(totalPoint(quiz)),
							Result:             database.Text(pb.ExamLOSubmissionResult_EXAM_LO_SUBMISSION_FAILED.String()),
							Status:             database.Text(pb.SubmissionStatus_SUBMISSION_STATUS_RETURNED.String()),
						}
						assert.Equal(t, expected, examLOSubmission)
					}).Once().Return(nil)
				loProgressionRepo.On("DeleteByStudyPlanIdentity", mock.Anything, mock.Anything, repositories.StudyPlanItemIdentity{
					StudentID:          database.Text("student-id-0"),
					StudyPlanID:        database.Text("study-plan-id-0"),
					LearningMaterialID: database.Text("learning-material-id-0"),
				}).Once().Return(int64(0), nil)
				loProgressionAnswerRepo.On("DeleteByStudyPlanIdentity", mock.Anything, mock.Anything, repositories.StudyPlanItemIdentity{
					StudentID:          database.Text("student-id-0"),
					StudyPlanID:        database.Text("study-plan-id-0"),
					LearningMaterialID: database.Text("learning-material-id-0"),
				}).Once().Return(int64(0), nil)

				db.On("Begin", ctx).Return(tx, nil).Times(3)
				tx.On("Commit", ctx).Return(nil).Times(3)
			},
			req: &pb.SubmitQuizAnswersRequest{
				SetId: "shuffle-quiz-set-id",
				QuizAnswer: []*pb.QuizAnswer{
					{
						QuizId: quizzes[3].ExternalID.String,
						Answer: []*pb.Answer{
							{Format: &pb.Answer_SubmittedKey{SubmittedKey: "key-2"}},
							{Format: &pb.Answer_SubmittedKey{SubmittedKey: "key-1"}},
							{Format: &pb.Answer_SubmittedKey{SubmittedKey: "key-3"}},
						},
					},
				},
			},
			expectedResp: &pb.SubmitQuizAnswersResponse{
				Logs: []*cpb.AnswerLog{
					{
						QuizId:   quizzes[3].ExternalID.String,
						QuizType: cpb.QuizType(cpb.QuizType_value[quizzes[3].Kind.String]),
						Result: &cpb.AnswerLog_OrderingResult{OrderingResult: &cpb.OrderingResult{
							SubmittedKeys: []string{"key-2", "key-1", "key-3"},
							CorrectKeys:   []string{"key-1", "key-2", "key-3"},
						}},
						Correctness: []bool{false, false, true},
						IsAccepted:  false,
					},
				},
				TotalGradedPoint:   wrapperspb.UInt32(0),
				TotalPoint:         wrapperspb.UInt32(uint32(totalPoint(quizzes[3:4]))),
				TotalCorrectAnswer: 0,
				TotalQuestion:      int32(len(quizzes[3:4])),
				SubmissionResult:   pb.ExamLOSubmissionResult_EXAM_LO_SUBMISSION_FAILED,
			},
			expectedErr: nil,
		},
		{
			name: "submit invalid answer ordering quiz",
			ctx:  ctx,
			setup: func(ctx context.Context) {
				quiz := quizzes[3:4]
				shuffledQuizSetRepo.On("GetLoID", ctx, db, database.Text("shuffle-quiz-set-id")).
					Return(database.Text("lo-id"), nil).Once()
				quizRepo.On("GetByExternalIDs", ctx, db, database.TextArray(getExternalIDFromQuiz(quiz)), database.Text("lo-id")).
					Return(quiz, nil).Once()
				expectedRepoMethodCallForCheckCorrectness(ctx, db, quiz)
			},
			req: &pb.SubmitQuizAnswersRequest{
				SetId: "shuffle-quiz-set-id",
				QuizAnswer: []*pb.QuizAnswer{
					{
						QuizId: quizzes[3].ExternalID.String,
						Answer: []*pb.Answer{
							{Format: &pb.Answer_SubmittedKey{SubmittedKey: "key-1"}},
							{Format: &pb.Answer_FilledText{FilledText: "key-2"}},
							{Format: &pb.Answer_SubmittedKey{SubmittedKey: "key-3"}},
						},
					},
				},
			},
			expectedErr: status.Error(
				codes.FailedPrecondition,
				fmt.Sprintf(
					"questionSrv.CheckQuestionsCorrectness: %v",
					fmt.Errorf("run opt: WithSubmitQuizAnswersRequest.executor.GetUserAnswerFromSubmitQuizAnswersRequest: %w",
						fmt.Errorf("your answer is not the ordering type, question %s (external_id), %s (quiz_id)", quizzes[3].ExternalID.String, quizzes[3].ID.String),
					)),
			),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			req := testCase.req.(*pb.SubmitQuizAnswersRequest)
			res, err := s.SubmitQuizAnswers(testCase.ctx, req)
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else if testCase.expectedResp != nil {
				require.NoError(t, err)
				expected := testCase.expectedResp.(*pb.SubmitQuizAnswersResponse)
				assert.NotNil(t, res)
				assert.Equal(t, expected.TotalGradedPoint, res.TotalGradedPoint)
				assert.Equal(t, expected.TotalPoint, res.TotalPoint)
				assert.Equal(t, expected.TotalCorrectAnswer, res.TotalCorrectAnswer)
				assert.Equal(t, expected.TotalQuestion, res.TotalQuestion)
				assert.Equal(t, expected.SubmissionResult, res.SubmissionResult)
				require.Len(t, expected.Logs, len(res.Logs))

				for i, log := range expected.Logs {
					assert.NotZero(t, res.Logs[i].SubmittedAt)
					assert.Nil(t, res.Logs[i].Core)
					log.SubmittedAt = res.Logs[i].SubmittedAt
					assert.Equal(t, log, res.Logs[i])
				}
			}

			mock.AssertExpectationsForObjects(
				t,
				db,
				tx,
				bookRepo,
				courseBookRepo,
				shuffledQuizSetRepo,
				quizRepo,
				studentsLearningObjectivesCompletenessRepo,
				examLORepo,
				examLOSubmissionRepo,
				examLOSubmissionAnswerRepo,
			)
		})
	}
}
