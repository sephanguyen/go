package services

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/eureka/services/question"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type CourseModifierService struct {
	DB database.Ext

	CourseBookRepo interface {
		Upsert(ctx context.Context, db database.Ext, cc []*entities.CoursesBooks) error
	}

	AssignmentRepo interface {
		DuplicateAssignment(ctx context.Context, db database.QueryExecer, copiedFromTopicIDs pgtype.TextArray, newTopicIDs pgtype.TextArray) error
		UpdateDisplayOrders(ctx context.Context, db database.QueryExecer, mDisplayOrder map[pgtype.Text]pgtype.Int4) error
		RetrieveAssignmentsByTopicIDs(ctx context.Context, db database.QueryExecer, topicIDs pgtype.TextArray) ([]*entities.Assignment, error)
	}
	TopicsAssignmentsRepo interface {
		BulkUpsert(ctx context.Context, db database.QueryExecer, topicsAssignmentsList []*entities.TopicsAssignments) error
		BulkUpdateDisplayOrder(
			ctx context.Context, db database.QueryExecer,
			topicsAssignments []*entities.TopicsAssignments,
		) error
	}

	BobInternalModifier interface {
		SubmitQuizAnswers(ctx context.Context, req *bpb.SubmitQuizAnswersRequest, opts ...grpc.CallOption) (*bpb.SubmitQuizAnswersResponse, error)
	}

	FlashcardProgressionRepo interface {
		Get(ctx context.Context, db database.QueryExecer, args *repositories.GetFlashcardProgressionArgs) (*entities.FlashcardProgression, error)
		UpdateCompletedAt(ctx context.Context, db database.QueryExecer, studySetID pgtype.Text) error
		DeleteByStudySetID(ctx context.Context, db database.QueryExecer, studySetID pgtype.Text) error
		Upsert(ctx context.Context, db database.QueryExecer, cc []*entities.FlashcardProgression) error
		GetByStudySetIDAndStudentID(ctx context.Context, db database.QueryExecer, studentID, studySetID pgtype.Text) (*entities.FlashcardProgression, error)
	}

	StudyPlanItemRepo interface {
		UpdateCompletedAtByID(ctx context.Context, db database.QueryExecer, id pgtype.Text, completedAt pgtype.Timestamptz) error
	}

	BookRepo interface {
		Upsert(ctx context.Context, db database.Ext, cc []*entities.Book) error
		DuplicateBook(ctx context.Context, db database.QueryExecer, bookID pgtype.Text, bookName pgtype.Text) (string, error)
		FindByIDs(ctx context.Context, db database.QueryExecer, bookIDs []string) (map[string]*entities.Book, error)
	}

	ChapterRepo interface {
		FindByIDs(ctx context.Context, db database.QueryExecer, chapterIDs []string) (map[string]*entities.Chapter, error)
		DuplicateChapters(ctx context.Context, db database.QueryExecer, bookID string, chapterIDs []string) ([]*entities.CopiedChapter, error)
	}

	BookChapterRepo interface {
		Upsert(ctx context.Context, db database.Ext, cc []*entities.BookChapter) error
		SoftDelete(ctx context.Context, db database.QueryExecer, chapterIDs, bookIDs pgtype.TextArray) error
		FindByBookIDs(ctx context.Context, db database.QueryExecer, bookIDs []string) (map[string][]*entities.BookChapter, error)
	}

	LearningObjectiveRepo interface {
		DuplicateLearningObjectives(ctx context.Context, db database.QueryExecer, topicIDs pgtype.TextArray, newTopicIDs pgtype.TextArray) ([]*entities.CopiedLearningObjective, error)
		RetrieveByTopicIDs(ctx context.Context, db database.QueryExecer, topicIds pgtype.TextArray) ([]*entities.LearningObjective, error)
	}

	TopicsLearningObjectivesRepo interface {
		BulkImport(context.Context, database.QueryExecer, []*entities.TopicsLearningObjectives) error
	}

	TopicRepo interface {
		DuplicateTopics(ctx context.Context, db database.QueryExecer, chapterIDs pgtype.TextArray, newChapterIDs pgtype.TextArray) ([]*entities.CopiedTopic, error)
	}

	StudentLearningTimeDailyRepo interface {
		Upsert(ctx context.Context, db database.QueryExecer, s *entities.StudentLearningTimeDaily) error
		Retrieve(ctx context.Context, db database.QueryExecer, studentID pgtype.Text, from, to *pgtype.Timestamptz, queryEnhancers ...repositories.QueryEnhancer) ([]*entities.StudentLearningTimeDaily, error)
	}

	UsermgmtUserReaderService interface {
		SearchBasicProfile(ctx context.Context, in *upb.SearchBasicProfileRequest, opts ...grpc.CallOption) (*upb.SearchBasicProfileResponse, error)
	}

	ShuffledQuizSetRepo interface {
		Get(context.Context, database.QueryExecer, pgtype.Text, pgtype.Int8, pgtype.Int8) (*entities.ShuffledQuizSet, error)
		GetSeed(context.Context, database.QueryExecer, pgtype.Text) (pgtype.Text, error)
		UpdateSubmissionHistory(context.Context, database.QueryExecer, pgtype.Text, pgtype.JSONB) error
		UpdateTotalCorrectness(context.Context, database.QueryExecer, pgtype.Text) error
		GetStudentID(context.Context, database.QueryExecer, pgtype.Text) (pgtype.Text, error)
		GetLoID(context.Context, database.QueryExecer, pgtype.Text) (pgtype.Text, error)
		GetScore(context.Context, database.QueryExecer, pgtype.Text) (pgtype.Int4, pgtype.Int4, error)
		IsFinishedQuizTest(context.Context, database.QueryExecer, pgtype.Text) (pgtype.Bool, error)
		GetQuizIdx(context.Context, database.QueryExecer, pgtype.Text, pgtype.Text) (pgtype.Int4, error)
		GetExternalIDsFromSubmissionHistory(ctx context.Context, db database.QueryExecer, shuffleQuizSetID pgtype.Text, isAccepted bool) (pgtype.TextArray, error)
		GenerateExamLOSubmission(ctx context.Context, db database.QueryExecer, shuffleQuizSetID pgtype.Text) (*entities.ExamLOSubmission, error)
		GetExternalIDs(ctx context.Context, db database.QueryExecer, shuffleQuizSetID pgtype.Text) (pgtype.TextArray, error)
	}

	QuizRepo interface {
		GetByExternalIDs(context.Context, database.QueryExecer, pgtype.TextArray, pgtype.Text) (entities.Quizzes, error)
		GetOptions(context.Context, database.QueryExecer, pgtype.Text, pgtype.Text) ([]*entities.QuizOption, error)
	}

	StudentsLearningObjectivesCompletenessRepo interface {
		UpsertHighestQuizScore(ctx context.Context, db database.QueryExecer, loID pgtype.Text, studentID pgtype.Text, newScore pgtype.Float4) error
		UpsertFirstQuizCompleteness(ctx context.Context, db database.QueryExecer, loID pgtype.Text, studentID pgtype.Text, firstQuizScore pgtype.Float4) error
	}

	ExamLORepo interface {
		Get(ctx context.Context, db database.QueryExecer, learningMaterialID pgtype.Text) (*entities.ExamLO, error)
	}

	ExamLOSubmissionRepo interface {
		Get(ctx context.Context, db database.QueryExecer, args *repositories.GetExamLOSubmissionArgs) (*entities.ExamLOSubmission, error)
		Insert(ctx context.Context, db database.QueryExecer, e *entities.ExamLOSubmission) error
		Update(ctx context.Context, db database.QueryExecer, e *entities.ExamLOSubmission) error
		GetTotalGradedPoint(ctx context.Context, db database.QueryExecer, submissionID pgtype.Text) (pgtype.Int4, error)
	}

	ExamLOSubmissionAnswerRepo interface {
		Upsert(ctx context.Context, db database.QueryExecer, e *entities.ExamLOSubmissionAnswer) (int, error)
	}

	LOProgressionRepo interface {
		DeleteByStudyPlanIdentity(ctx context.Context, db database.QueryExecer, args repositories.StudyPlanItemIdentity) (int64, error)
	}
	LOProgressionAnswerRepo interface {
		DeleteByStudyPlanIdentity(ctx context.Context, db database.QueryExecer, args repositories.StudyPlanItemIdentity) (int64, error)
	}
	assignmentModifierService        *AssignmentModifierService
	learningObjectiveModifierService *LearningObjectiveModifierService
}

func NewCourseModifierService(
	db database.Ext,
	bobInternalModifier bpb.InternalModifierServiceClient,
	assignmentModifierService *AssignmentModifierService,
	learningObjectiveModifierService *LearningObjectiveModifierService,
	usermgmtUserReader upb.UserReaderServiceClient,
) *CourseModifierService {
	return &CourseModifierService{
		DB:                               db,
		BookRepo:                         new(repositories.BookRepo),
		CourseBookRepo:                   new(repositories.CourseBookRepo),
		AssignmentRepo:                   &repositories.AssignmentRepo{},
		TopicsAssignmentsRepo:            &repositories.TopicsAssignmentsRepo{},
		assignmentModifierService:        assignmentModifierService,
		learningObjectiveModifierService: learningObjectiveModifierService,
		BobInternalModifier:              bobInternalModifier,
		UsermgmtUserReaderService:        usermgmtUserReader,
		FlashcardProgressionRepo:         &repositories.FlashcardProgressionRepo{},
		StudyPlanItemRepo:                &repositories.StudyPlanItemRepo{},
		ChapterRepo:                      &repositories.ChapterRepo{},
		BookChapterRepo:                  &repositories.BookChapterRepo{},
		LearningObjectiveRepo:            &repositories.LearningObjectiveRepo{},
		TopicsLearningObjectivesRepo:     &repositories.TopicsLearningObjectivesRepo{},
		TopicRepo:                        &repositories.TopicRepo{},
		StudentLearningTimeDailyRepo:     new(repositories.StudentLearningTimeDailyRepo),
		ShuffledQuizSetRepo:              &repositories.ShuffledQuizSetRepo{},
		QuizRepo:                         &repositories.QuizRepo{},
		StudentsLearningObjectivesCompletenessRepo: &repositories.StudentsLearningObjectivesCompletenessRepo{},
		ExamLORepo:                 &repositories.ExamLORepo{},
		ExamLOSubmissionRepo:       &repositories.ExamLOSubmissionRepo{},
		ExamLOSubmissionAnswerRepo: &repositories.ExamLOSubmissionAnswerRepo{},
		LOProgressionRepo:          &repositories.LOProgressionRepo{},
		LOProgressionAnswerRepo:    &repositories.LOProgressionAnswerRepo{},
	}
}

func (cm *CourseModifierService) duplicateChapter(ctx context.Context, tx pgx.Tx, chapterIDs []string, bookID string) ([]string, []string, error) {
	copiedChapters, err := cm.ChapterRepo.DuplicateChapters(ctx, tx, bookID, chapterIDs)
	if err != nil {
		return nil, nil, status.Error(codes.Internal, fmt.Errorf("cm.ChapterRepo.DuplicateChapters: %w", err).Error())
	}

	newBookChapters := make([]*entities.BookChapter, len(chapterIDs))
	orgChapterIDs := make([]string, 0, len(copiedChapters))
	newChapterIDs := make([]string, 0, len(copiedChapters))
	now := time.Now()
	for i, copiedChapter := range copiedChapters {
		orgChapterIDs = append(orgChapterIDs, copiedChapter.CopyFromID.String)
		newChapterIDs = append(newChapterIDs, copiedChapter.ID.String)
		e := &entities.BookChapter{}
		database.AllNullEntity(e)
		err = multierr.Combine(
			e.BookID.Set(bookID),
			e.ChapterID.Set(copiedChapter.ID.String),
			e.CreatedAt.Set(now),
			e.UpdatedAt.Set(now),
		)
		if err != nil {
			return nil, nil, fmt.Errorf("error converting book chapter: %w", err)
		}
		newBookChapters[i] = e
	}
	err = cm.BookChapterRepo.Upsert(ctx, tx, newBookChapters)
	if err != nil {
		return nil, nil, fmt.Errorf("cm.BookChapter.Upsert: %w", err)
	}
	return orgChapterIDs, newChapterIDs, nil
}

func (cm *CourseModifierService) duplicateTopics(ctx context.Context, tx pgx.Tx, orgChapterIDs []string, newChapterIDs []string) ([]string, []string, error) {
	copiedTopics, err := cm.TopicRepo.DuplicateTopics(ctx, tx, database.TextArray(orgChapterIDs), database.TextArray(newChapterIDs))
	if err != nil {
		return nil, nil, fmt.Errorf("cm.TopicRepo.DuplicateTopics: %w", err)
	}
	newTopicIDs := make([]string, len(copiedTopics))
	orgTopicIDs := make([]string, len(copiedTopics))
	for i, copiedTopic := range copiedTopics {
		newTopicIDs[i] = copiedTopic.ID.String
		orgTopicIDs[i] = copiedTopic.CopyFromID.String
	}
	return orgTopicIDs, newTopicIDs, nil
}

func (cm *CourseModifierService) DuplicateBook(ctx context.Context, req *pb.DuplicateBookRequest) (*pb.DuplicateBookResponse, error) {
	booksChapterMap, err := cm.BookChapterRepo.FindByBookIDs(ctx, cm.DB, []string{req.BookId})
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("cm.BookChapter.FindByBookIDs: %w", err).Error())
	}

	var chapterIDs []string
	if bookChapters, ok := booksChapterMap[req.BookId]; ok {
		chapterIDs = make([]string, len(bookChapters))
		for i, bookChapter := range bookChapters {
			chapterIDs[i] = bookChapter.ChapterID.String
		}
	}

	now := time.Now()
	var createdBook string
	var orgTopicIDs, newTopicIDs []string
	err = database.ExecInTx(ctx, cm.DB, func(ctx context.Context, tx pgx.Tx) error {
		newBookID, err := cm.BookRepo.DuplicateBook(ctx, tx, database.Text(req.BookId), database.Text(req.BookName))
		if err != nil {
			return fmt.Errorf("cm.BookRepo.DuplicateBook: %w", err)
		}
		createdBook = newBookID

		if len(chapterIDs) > 0 {
			orgChapterIDs, newChapterIDs, err := cm.duplicateChapter(ctx, tx, chapterIDs, newBookID)
			if err != nil {
				return fmt.Errorf("cm.duplicateChapter: %w", err)
			}
			orgTopicIDs, newTopicIDs, err = cm.duplicateTopics(ctx, tx, orgChapterIDs, newChapterIDs)
			if err != nil {
				return fmt.Errorf("cm.duplicateTopics: %w", err)
			}
			if _, err = cm.LearningObjectiveRepo.DuplicateLearningObjectives(ctx, tx, database.TextArray(orgTopicIDs), database.TextArray(newTopicIDs)); err != nil {
				return fmt.Errorf("cm.LearningObjectiveRepo.DuplicateLearningObjectives: %w", err)
			}

			// duplicate learning objective
			newLearningObjectives, err := cm.LearningObjectiveRepo.RetrieveByTopicIDs(ctx, tx, database.TextArray(newTopicIDs))
			if err != nil {
				return fmt.Errorf("cm.TopicsLearningObjectivesRepo.RetrieveByTopicIDs: %w", err)
			}

			var topicsLearningObjectives []*entities.TopicsLearningObjectives
			for _, oldLearningObjective := range newLearningObjectives {
				topicsLearningObjectives = append(topicsLearningObjectives, &entities.TopicsLearningObjectives{
					TopicID:      oldLearningObjective.TopicID,
					LoID:         oldLearningObjective.ID,
					DisplayOrder: oldLearningObjective.DisplayOrder,
					CreatedAt:    database.Timestamptz(now),
					UpdatedAt:    database.Timestamptz(now),
				})
			}

			if err := cm.TopicsLearningObjectivesRepo.BulkImport(ctx, tx, topicsLearningObjectives); err != nil {
				return fmt.Errorf("cm.TopicsLearningObjectivesRepo.BulkImport: %w", err)
			}

			// duplicate assignment
			if err = cm.AssignmentRepo.DuplicateAssignment(ctx, tx, database.TextArray(orgTopicIDs), database.TextArray(newTopicIDs)); err != nil {
				return fmt.Errorf("s.AssignmentRepo.DuplicateAssignment: %w", err)
			}

			newAssignments, err := cm.AssignmentRepo.RetrieveAssignmentsByTopicIDs(ctx, tx, database.TextArray(newTopicIDs))
			if err != nil {
				return fmt.Errorf("s.AssignmentRepo.RetrieveAssignmentsByTopicIDs: %w", err)
			}

			var topicsAssignments []*entities.TopicsAssignments
			for _, oldAssignment := range newAssignments {
				var assignmentContent entities.AssignmentContent
				oldAssignment.Content.AssignTo(&assignmentContent)

				topicsAssignments = append(topicsAssignments, &entities.TopicsAssignments{
					TopicID:      database.Text(assignmentContent.TopicID),
					AssignmentID: oldAssignment.ID,
					DisplayOrder: database.Int2(int16(oldAssignment.DisplayOrder.Int)),
					CreatedAt:    database.Timestamptz(now),
					UpdatedAt:    database.Timestamptz(now),
				})
			}

			if err := cm.TopicsAssignmentsRepo.BulkUpsert(ctx, tx, topicsAssignments); err != nil {
				return fmt.Errorf("s.TopicsAssignmentsRepo.BulkUpsert: %w", err)
			}

			return nil
		}

		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.DuplicateBookResponse{
		NewBookID:  createdBook,
		NewTopicId: newTopicIDs,
		OrgTopicId: orgTopicIDs,
	}, nil
}

func (s *CourseModifierService) UpsertLOsAndAssignments(
	ctx context.Context, req *pb.UpsertLOsAndAssignmentsRequest,
) (*pb.UpsertLOsAndAssignmentsResponse, error) {
	headers, _ := metadata.FromIncomingContext(ctx)
	pkg := headers["pkg"][0]
	token := headers["token"][0]
	version := headers["version"][0]

	var assignmentIDs, loIDs []string

	err := database.ExecInTxWithRetry(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		if len(req.Assignments) != 0 {
			upsertAssignmentsResponse, _err := upsertAssignment(
				ctx, &pb.UpsertAssignmentsRequest{
					Assignments: req.Assignments,
				}, tx,
				s.assignmentModifierService.AssignmentRepo,
			)
			if _err != nil {
				return fmt.Errorf("s.upsertAssignment: %w", _err)
			}
			assignmentIDs = upsertAssignmentsResponse.AssignmentIds
		}

		if len(req.LearningObjectives) != 0 {
			upsertLOsResponse, _err := s.learningObjectiveModifierService.UpsertLOs(
				metadata.AppendToOutgoingContext(ctx, "pkg", pkg, "version", version, "token", token),
				&pb.UpsertLOsRequest{
					LearningObjectives: req.LearningObjectives,
				})
			if _err != nil {
				return fmt.Errorf("s.BobCourseModifier.UpsertLOs: %w", _err)
			}
			loIDs = upsertLOsResponse.LoIds
		}

		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("ExecInTxWithRetry: %v", err))
	}

	return &pb.UpsertLOsAndAssignmentsResponse{
		AssignmentIds: assignmentIDs,
		LoIds:         loIDs,
	}, nil
}

func (s *CourseModifierService) UpdateDisplayOrdersOfLOsAndAssignments(
	ctx context.Context, req *pb.UpdateDisplayOrdersOfLOsAndAssignmentsRequest,
) (*pb.UpdateDisplayOrdersOfLOsAndAssignmentsResponse, error) {
	resp := &pb.UpdateDisplayOrdersOfLOsAndAssignmentsResponse{}

	err := database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		now := database.Timestamptz(time.Now())
		if len(req.Assignments) != 0 {
			mDisplayOrder := make(map[pgtype.Text]pgtype.Int4)
			topicsAssignments := make([]*entities.TopicsAssignments, 0, len(req.Assignments))
			for _, assignment := range req.Assignments {
				topicsAssignments = append(topicsAssignments, &entities.TopicsAssignments{
					TopicID:      database.Text(assignment.TopicId),
					AssignmentID: database.Text(assignment.AssignmentId),
					DisplayOrder: database.Int2(int16(assignment.DisplayOrder)),
					CreatedAt:    now,
					UpdatedAt:    now,
				})
				resp.Assignments = append(resp.Assignments, &pb.UpdateDisplayOrdersOfLOsAndAssignmentsResponse_Assignment{
					AssignmentId: assignment.AssignmentId,
					TopicId:      assignment.TopicId,
				})
				mDisplayOrder[database.Text(assignment.AssignmentId)] = database.Int4(assignment.DisplayOrder)
			}
			if _err := s.TopicsAssignmentsRepo.BulkUpdateDisplayOrder(ctx, tx, topicsAssignments); _err != nil {
				return fmt.Errorf("s.assignmentModifierService.TopicsAssignmentsRepo.BulkUpdateDisplayOrder: %w", _err)
			}

			if _err := s.AssignmentRepo.UpdateDisplayOrders(ctx, tx, mDisplayOrder); _err != nil {
				return fmt.Errorf("s.AssignmentRepo.UpdateDisplayOrders: %w", _err)
			}
		}

		if len(req.LearningObjectives) != 0 {
			los := make([]*pb.TopicLODisplayOrder, 0, len(req.LearningObjectives))
			for _, lo := range req.LearningObjectives {
				los = append(los, &pb.TopicLODisplayOrder{
					LoId:         lo.LoId,
					DisplayOrder: lo.DisplayOrder,
					TopicId:      lo.TopicId,
				})
			}

			updateDisplayOrdersOfLOs, err := s.learningObjectiveModifierService.UpdateDisplayOrdersOfLOs(ctx, los)
			if err != nil {
				return fmt.Errorf("s.learningObjectiveModifierService.UpdateDisplayOrdersOfLOs: %w", err)
			}

			for _, lo := range updateDisplayOrdersOfLOs {
				resp.LearningObjectives = append(resp.LearningObjectives, &pb.UpdateDisplayOrdersOfLOsAndAssignmentsResponse_LearningObjective{
					LoId:    lo.LoId,
					TopicId: lo.TopicId,
				})
			}
		}

		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("ExecInTxWithRetry: %v", err))
	}

	return resp, nil
}

func (s *CourseModifierService) CompleteStudyPlanItem(ctx context.Context, req *pb.CompleteStudyPlanItemRequest) (*pb.CompleteStudyPlanItemResponse, error) {
	if req.StudyPlanItemId == "" {
		return nil, status.Error(codes.InvalidArgument, "missing study plan item id")
	}

	if err := s.assignmentModifierService.StudyPlanItemRepo.UpdateCompletedAtByID(ctx, s.DB, database.Text(req.StudyPlanItemId), pgtype.Timestamptz{Status: pgtype.Null}); err != nil {
		return nil, status.Errorf(codes.Internal, "s.assignmentModifierService.StudyPlanItemRepo.UpdateCompletedAtByID: %v", err)
	}

	return &pb.CompleteStudyPlanItemResponse{
		IsSuccess: true,
	}, nil
}

func (s *CourseModifierService) validateCheckSubmitQuizAnswersRequest(req *pb.SubmitQuizAnswersRequest) error {
	if req.SetId == "" {
		return fmt.Errorf("req must have SetId")
	}
	for _, quizAnswer := range req.QuizAnswer {
		if err := validateQuizAnswerMessage(quizAnswer); err != nil {
			return err
		}
	}

	return nil
}

func validateQuizAnswerMessage(quizAnswer *pb.QuizAnswer) error {
	if quizAnswer.GetQuizId() == "" {
		return fmt.Errorf("req must have QuizId")
	}
	if len(quizAnswer.GetAnswer()) == 0 {
		return fmt.Errorf("req must have quizAnswer")
	}

	if !isValidAnswerMessage(quizAnswer.Answer[0]) {
		return fmt.Errorf(fmt.Sprintf("your answer of quiz_id(%s) is must not empty", quizAnswer.QuizId))
	}
	return nil
}

//nolint:gosimple
func isValidAnswerMessage(answer *pb.Answer) bool {
	format := answer.GetFormat()
	if format == nil {
		return false
	}

	return true
}

func (s *CourseModifierService) checkCorrectnessQuizzes(
	ctx context.Context,
	quizzes entities.Quizzes,
	quizAnswersMap map[string][]*pb.Answer,
	req *pb.SubmitQuizAnswersRequest,
) (_ []*entities.QuizAnswer, err error) {
	questionSrv := &question.Service{}
	answersEnt, err := questionSrv.CheckQuestionsCorrectness(quizzes, question.WithSubmitQuizAnswersRequest(req))
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, fmt.Sprintf("questionSrv.CheckQuestionsCorrectness: %v", err))
	}
	if len(quizzes) != len(answersEnt) {
		return nil, status.Error(codes.FailedPrecondition, fmt.Sprintf("got ERROR: questionSrv.CheckQuestionsCorrectness: shuffle quiz set id %s :expected %d answer but got %d \n", req.SetId, len(quizzes), len(answersEnt)))
	}

	for i, quiz := range quizzes {
		var answerEnt *entities.QuizAnswer
		answers := quizAnswersMap[quiz.ExternalID.String]

		switch quiz.Kind.String {
		case cpb.QuizType_QUIZ_TYPE_MCQ.String(), cpb.QuizType_QUIZ_TYPE_MAQ.String():
			MCQuiz := &MultipleChoiceQuiz{
				Quiz:                quiz,
				SetID:               req.SetId,
				ShuffledQuizSetRepo: s.ShuffledQuizSetRepo,
			}

			answerEnt, err = MCQuiz.CheckCorrectness(ctx, s.DB, answers)
			if err != nil {
				return nil, status.Error(codes.FailedPrecondition, fmt.Sprintf("SubmitQuizAnswers.MultipleChoice.CheckCorrectness: %v", err))
			}
		case cpb.QuizType_QUIZ_TYPE_MIQ.String():
			manualInputQuiz := &ManualInputQuiz{
				Quiz: quiz,
			}

			answerEnt, err = manualInputQuiz.CheckCorrectness(ctx, s.DB, answers)
			if err != nil {
				return nil, status.Error(codes.FailedPrecondition, fmt.Sprintf("SubmitQuizAnswers.ManualInputQuiz.CheckCorrectness: %v", err))
			}
		case cpb.QuizType_QUIZ_TYPE_FIB.String(), cpb.QuizType_QUIZ_TYPE_POW.String(), cpb.QuizType_QUIZ_TYPE_TAD.String():
			fillInTheBlankQuiz := &FillInTheBlankQuiz{
				Quiz: quiz,
			}

			answerEnt, err = fillInTheBlankQuiz.CheckCorrectness(ctx, s.DB, answers)
			if err != nil {
				return nil, status.Error(codes.FailedPrecondition, fmt.Sprintf("SubmitQuizAnswers.FillInTheBlank.CheckCorrectness: %v", err))
			}
		case cpb.QuizType_QUIZ_TYPE_ORD.String(), cpb.QuizType_QUIZ_TYPE_ESQ.String(): // skip
			continue
		default:
			return nil, status.Error(codes.FailedPrecondition, "Not supported quiz type!")
		}

		answersEnt[i] = answerEnt
	}

	return answersEnt, nil
}

func (s *CourseModifierService) SubmitQuizAnswers(ctx context.Context, req *pb.SubmitQuizAnswersRequest) (*pb.SubmitQuizAnswersResponse, error) {
	if err := s.validateCheckSubmitQuizAnswersRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	loID, err := s.ShuffledQuizSetRepo.GetLoID(ctx, s.DB, database.Text(req.SetId))
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Errorf("SubmitQuizAnswers.ShuffledQuizSetRepo.GetLoID: %v", err).Error())
	}

	quizIDs := make([]string, 0, len(req.QuizAnswer))
	quizAnswersMap := make(map[string][]*pb.Answer)
	for _, quizAnswer := range req.QuizAnswer {
		quizIDs = append(quizIDs, quizAnswer.QuizId)
		quizAnswersMap[quizAnswer.QuizId] = quizAnswer.Answer
	}
	quizzes, err := s.QuizRepo.GetByExternalIDs(ctx, s.DB, database.TextArray(quizIDs), loID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Errorf("SubmitQuizAnswers.ShuffledQuizSetRepo.GetByExternalIDs: %v", err).Error())
	}

	answersEnt, err := s.checkCorrectnessQuizzes(ctx, quizzes, quizAnswersMap, req)
	if err != nil {
		return nil, err
	}

	err = database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		if err := s.ShuffledQuizSetRepo.UpdateSubmissionHistory(ctx, tx, database.Text(req.SetId), database.JSONB(answersEnt)); err != nil {
			return fmt.Errorf("SubmitQuizAnswers.ShuffledQuizSetRepo.UpdateSubmissionHistory: %v", err)
		}

		if err := s.ShuffledQuizSetRepo.UpdateTotalCorrectness(ctx, tx, database.Text(req.SetId)); err != nil {
			return fmt.Errorf("SubmitQuizAnswers.ShuffledQuizSetRepo.UpdateTotalCorrectness: %v", err)
		}

		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	err = database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		isFinished, err := s.ShuffledQuizSetRepo.IsFinishedQuizTest(ctx, tx, database.Text(req.SetId))
		if err != nil && err != pgx.ErrNoRows {
			return fmt.Errorf("SubmitQuizAnswers.ShuffledQuizSetRepo.IsFinishedQuizTest: %v", err)
		}

		shuffledQuizSet, err := s.ShuffledQuizSetRepo.Get(ctx, tx, database.Text(req.SetId), database.Int8(1), database.Int8(1))
		if err != nil {
			return fmt.Errorf("SubmitQuizAnswers.ShuffledQuizSetRepo.Get: %v", err)
		}
		isRetry := shuffledQuizSet.OriginalShuffleQuizSetID.Status == pgtype.Present

		if isRetry || isFinished.Bool {
			studentID, err := s.ShuffledQuizSetRepo.GetStudentID(ctx, tx, database.Text(req.SetId))
			if err != nil {
				return fmt.Errorf("SubmitQuizAnswers.ShuffledQuizSetRepo.GetStudentID: %v", err)
			}

			totalCorrectness, totalQuiz, err := s.ShuffledQuizSetRepo.GetScore(ctx, tx, database.Text(req.SetId))
			if err != nil {
				return fmt.Errorf("SubmitQuizAnswers.ShuffledQuizSetRepo.GetScore: %v", err)
			}

			if isRetry {
				externalIDsFromSubmissionHistory, err := s.ShuffledQuizSetRepo.GetExternalIDsFromSubmissionHistory(ctx, tx, shuffledQuizSet.OriginalShuffleQuizSetID, false)
				if err != nil {
					return fmt.Errorf("SubmitQuizAnswers.ShuffledQuizSetRepo.GetExternalIDsFromSubmissionHistory: %w", err)
				}
				externalQuizIDs := make([]string, 0)
				for _, e := range externalIDsFromSubmissionHistory.Elements {
					if e.Status == pgtype.Present {
						externalQuizIDs = append(externalQuizIDs, e.String)
					}
				}
				for _, e := range shuffledQuizSet.QuizExternalIDs.Elements {
					if e.Status == pgtype.Present {
						externalQuizIDs = append(externalQuizIDs, e.String)
					}
				}
				totalQuiz = database.Int4(int32(len(golibs.GetUniqueElementStringArray(externalQuizIDs))))
			}

			score := float32(math.Floor(float64(totalCorrectness.Int) / float64(totalQuiz.Int) * 100))
			if err = s.StudentsLearningObjectivesCompletenessRepo.UpsertFirstQuizCompleteness(ctx, tx, loID, studentID, database.Float4(score)); err != nil {
				return fmt.Errorf("SubmitQuizAnswers.StudentsLearningObjectivesCompletenessRepo.UpsertFirstQuizCompleteness: %v", err)
			}

			if err = s.StudentsLearningObjectivesCompletenessRepo.UpsertHighestQuizScore(ctx, tx, loID, studentID, database.Float4(score)); err != nil {
				return fmt.Errorf("SubmitQuizAnswers.StudentsLearningObjectivesCompletenessRepo.UpsertHighestQuizScore: %v", err)
			}
		}

		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Refactor SubmitQuizAnswers to support Manual grading
	var totalGradedPoint, totalPoint uint32
	var totalCorrectAnswer, totalQuestion int32
	var submissionResult pb.ExamLOSubmissionResult

	examLOSubmission, err := s.ExamLOSubmissionRepo.Get(ctx, s.DB, &repositories.GetExamLOSubmissionArgs{
		SubmissionID:      pgtype.Text{Status: pgtype.Null},
		ShuffledQuizSetID: database.Text(req.SetId),
	})
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, status.Error(codes.Internal, fmt.Errorf("ExamLOSubmissionRepo.Get: %w", err).Error())
	}

	externalIDs, err := s.ShuffledQuizSetRepo.GetExternalIDs(ctx, s.DB, database.Text(req.SetId))
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("ShuffledQuizSetRepo.GetExternalIDs: %w", err).Error())
	}

	err = database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		if examLOSubmission != nil {
			examLO, err := s.ExamLORepo.Get(ctx, tx, examLOSubmission.LearningMaterialID)
			if err != nil {
				return fmt.Errorf("ExamLORepo.Get: %w", err)
			}

			resultTotalGradedPoint, err := s.ExamLOSubmissionRepo.GetTotalGradedPoint(ctx, tx, examLOSubmission.SubmissionID)
			if err != nil {
				return fmt.Errorf("ExamLOSubmissionRepo.GetTotalGradedPoint: %w", err)
			}
			totalGradedPoint = uint32(resultTotalGradedPoint.Int)

			if examLO.ManualGrading.Bool {
				examLOSubmission.Status = database.Text(pb.SubmissionStatus_SUBMISSION_STATUS_NOT_MARKED.String())
				examLOSubmission.Result = database.Text(pb.ExamLOSubmissionResult_EXAM_LO_SUBMISSION_WAITING_FOR_GRADE.String())
			} else {
				examLOSubmission.Status = database.Text(pb.SubmissionStatus_SUBMISSION_STATUS_RETURNED.String())
				if examLO.GradeToPass.Status == pgtype.Null {
					examLOSubmission.Result = database.Text(pb.ExamLOSubmissionResult_EXAM_LO_SUBMISSION_COMPLETED.String())
				} else {
					if resultTotalGradedPoint.Int >= examLO.GradeToPass.Int {
						examLOSubmission.Result = database.Text(pb.ExamLOSubmissionResult_EXAM_LO_SUBMISSION_PASSED.String())
					} else {
						examLOSubmission.Result = database.Text(pb.ExamLOSubmissionResult_EXAM_LO_SUBMISSION_FAILED.String())
					}
				}
			}

			examLOSubmission.UpdatedAt = database.Timestamptz(time.Now())
			err = s.ExamLOSubmissionRepo.Update(ctx, tx, examLOSubmission)
			if err != nil {
				return fmt.Errorf("s.ExamLOSubmissionRepo.Update: %w", err)
			}

			for _, e := range externalIDs.Elements {
				if e.Status == pgtype.Present {
					totalQuestion++
				}
			}

			for _, answerEnt := range answersEnt {
				if answerEnt.IsAccepted {
					totalCorrectAnswer++
				}
			}

			totalPoint = uint32(examLOSubmission.TotalPoint.Int)
			submissionResult = pb.ExamLOSubmissionResult(pb.ExamLOSubmissionResult_value[examLOSubmission.Result.String])
		} else {
			examLOSubmission, err = s.ShuffledQuizSetRepo.GenerateExamLOSubmission(ctx, tx, database.Text(req.SetId))
			if err != nil {
				return fmt.Errorf("ShuffledQuizSetRepo.GenerateExamLOSubmission: %w", err)
			}

			if err = s.ExamLOSubmissionRepo.Insert(ctx, tx, examLOSubmission); err != nil {
				return fmt.Errorf("ExamLOSubmissionRepo.Insert: %w", err)
			}

			cloneAnswersEnt := make([]*entities.QuizAnswer, len(answersEnt))
			copy(cloneAnswersEnt, answersEnt)

			for _, e := range externalIDs.Elements {
				if e.Status == pgtype.Present {
					totalQuestion++

					answerEnt := &entities.QuizAnswer{QuizID: e.String, Correctness: make([]bool, 0)}
					for i, value := range cloneAnswersEnt {
						if e.String == value.QuizID {
							answerEnt = value
							cloneAnswersEnt = RemoveQuizAnswerByIndex(cloneAnswersEnt, i)
							break
						}
					}

					answer, err := s.toExamLOSubmissionAnswerEntityBySubmission(examLOSubmission, answerEnt)
					if err != nil {
						return fmt.Errorf("toExamLOSubmissionAnswerEntityBySubmission: %w", err)
					}

					if _, err = s.ExamLOSubmissionAnswerRepo.Upsert(ctx, tx, answer); err != nil {
						return fmt.Errorf("ExamLOSubmissionAnswerRepo.Upsert: %w", err)
					}

					if answer.IsAccepted.Bool {
						totalCorrectAnswer++
					}
				}
			}

			examLO, err := s.ExamLORepo.Get(ctx, tx, examLOSubmission.LearningMaterialID)
			if err != nil {
				return fmt.Errorf("ExamLORepo.Get: %w", err)
			}

			resultTotalGradedPoint, err := s.ExamLOSubmissionRepo.GetTotalGradedPoint(ctx, tx, examLOSubmission.SubmissionID)
			if err != nil {
				return fmt.Errorf("ExamLOSubmissionRepo.GetTotalGradedPoint: %w", err)
			}
			totalGradedPoint = uint32(resultTotalGradedPoint.Int)

			if examLO.ManualGrading.Bool {
				examLOSubmission.Status = database.Text(pb.SubmissionStatus_SUBMISSION_STATUS_NOT_MARKED.String())
				examLOSubmission.Result = database.Text(pb.ExamLOSubmissionResult_EXAM_LO_SUBMISSION_WAITING_FOR_GRADE.String())
			} else {
				examLOSubmission.Status = database.Text(pb.SubmissionStatus_SUBMISSION_STATUS_RETURNED.String())
				if examLO.GradeToPass.Status == pgtype.Null {
					examLOSubmission.Result = database.Text(pb.ExamLOSubmissionResult_EXAM_LO_SUBMISSION_COMPLETED.String())
				} else {
					if resultTotalGradedPoint.Int >= examLO.GradeToPass.Int {
						examLOSubmission.Result = database.Text(pb.ExamLOSubmissionResult_EXAM_LO_SUBMISSION_PASSED.String())
					} else {
						examLOSubmission.Result = database.Text(pb.ExamLOSubmissionResult_EXAM_LO_SUBMISSION_FAILED.String())
					}
				}
			}
			examLOSubmission.UpdatedAt = database.Timestamptz(time.Now())

			err = s.ExamLOSubmissionRepo.Update(ctx, tx, examLOSubmission)
			if err != nil {
				return fmt.Errorf("s.ExamLOSubmissionRepo.Update: %w", err)
			}

			totalPoint = uint32(examLOSubmission.TotalPoint.Int)
			submissionResult = pb.ExamLOSubmissionResult(pb.ExamLOSubmissionResult_value[examLOSubmission.Result.String])
		}
		// soft delete lo progression and lo progression answer records
		if _, err = s.LOProgressionRepo.DeleteByStudyPlanIdentity(ctx, tx, repositories.StudyPlanItemIdentity{
			StudentID:          examLOSubmission.StudentID,
			StudyPlanID:        examLOSubmission.StudyPlanID,
			LearningMaterialID: examLOSubmission.LearningMaterialID,
		}); err != nil {
			return fmt.Errorf("s.LOProgressionRepo.DeleteByStudyPlanIdentity: %w", err)
		}
		if _, err = s.LOProgressionAnswerRepo.DeleteByStudyPlanIdentity(ctx, tx, repositories.StudyPlanItemIdentity{
			StudentID:          examLOSubmission.StudentID,
			StudyPlanID:        examLOSubmission.StudyPlanID,
			LearningMaterialID: examLOSubmission.LearningMaterialID,
		}); err != nil {
			return fmt.Errorf("s.LOProgressionAnswerRepo.DeleteByStudyPlanIdentity: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("database.ExecInTx: %w", err).Error())
	}

	resp := &pb.SubmitQuizAnswersResponse{}

	for _, answerEnt := range answersEnt {
		resp.Logs = append(resp.Logs, answerEnt.ToAnswerLogProtoMessage())
	}

	resp.TotalGradedPoint = wrapperspb.UInt32(totalGradedPoint)
	resp.TotalPoint = wrapperspb.UInt32(totalPoint)
	resp.TotalCorrectAnswer = totalCorrectAnswer
	resp.SubmissionResult = submissionResult
	resp.TotalQuestion = totalQuestion
	return resp, nil
}

func RemoveQuizAnswerByIndex(s []*entities.QuizAnswer, index int) []*entities.QuizAnswer {
	return append(s[:index], s[index+1:]...)
}

func (s *CourseModifierService) toExamLOSubmissionAnswerEntityBySubmission(examLOSubmission *entities.ExamLOSubmission, answerEnt *entities.QuizAnswer) (*entities.ExamLOSubmissionAnswer, error) {
	e := &entities.ExamLOSubmissionAnswer{}
	database.AllNullEntity(e)

	err := multierr.Combine(
		e.StudentID.Set(examLOSubmission.StudentID),
		e.QuizID.Set(answerEnt.QuizID),
		e.SubmissionID.Set(examLOSubmission.SubmissionID),
		e.StudyPlanID.Set(examLOSubmission.StudyPlanID),
		e.LearningMaterialID.Set(examLOSubmission.LearningMaterialID),
		e.ShuffledQuizSetID.Set(examLOSubmission.ShuffledQuizSetID),
		e.StudentTextAnswer.Set(answerEnt.FilledText),
		e.CorrectTextAnswer.Set(answerEnt.CorrectText),
		e.StudentIndexAnswer.Set(answerEnt.SelectedIndex),
		e.CorrectIndexAnswer.Set(answerEnt.CorrectIndex),
		e.IsCorrect.Set(answerEnt.Correctness),
		e.Point.Set(answerEnt.Point),
		e.CreatedAt.Set(examLOSubmission.CreatedAt),
		e.UpdatedAt.Set(examLOSubmission.UpdatedAt),
		e.CorrectKeysAnswer.Set(answerEnt.CorrectKeys),
		e.SubmittedKeysAnswer.Set(answerEnt.SubmittedKeys),
	)

	if len(answerEnt.Correctness) > 0 {
		err = multierr.Append(err, multierr.Combine(
			e.IsAccepted.Set(answerEnt.IsAccepted),
			e.IsCorrect.Set(answerEnt.Correctness),
		))
	} else {
		e.IsCorrect = pgtype.BoolArray{
			Elements: []pgtype.Bool{},
			Status:   pgtype.Present,
		}
	}

	if err != nil {
		return nil, fmt.Errorf("toExamLOSubmissionAnswerEntityBySubmission: %w", err)
	}

	return e, nil
}

func (s *CourseModifierService) FinishFlashCardStudyProgress(ctx context.Context, req *pb.FinishFlashCardStudyProgressRequest) (*pb.FinishFlashCardStudyProgressResponse, error) {
	if req.StudySetId == "" {
		return nil, status.Error(codes.InvalidArgument, "req must have study set id")
	}
	if req.StudentId == "" {
		return nil, status.Error(codes.InvalidArgument, "req must have student id")
	}
	if req.LoId == "" {
		return nil, status.Error(codes.InvalidArgument, "req must have lo id")
	}

	studyPlanItemID := pgtype.Text{Status: pgtype.Null}
	if req.StudyPlanItemId != "" {
		studyPlanItemID = database.Text(req.StudyPlanItemId)
	}

	flashcardProgress, err := s.FlashcardProgressionRepo.Get(ctx, s.DB, &repositories.GetFlashcardProgressionArgs{
		StudySetID:      database.Text(req.StudySetId),
		StudentID:       database.Text(req.StudentId),
		LoID:            database.Text(req.LoId),
		StudyPlanItemID: studyPlanItemID,
		LmID:            pgtype.Text{Status: pgtype.Null},
		StudyPlanID:     pgtype.Text{Status: pgtype.Null},
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "s.FlashcardProgressionRepo.Get: %v", err)
	}

	if !req.IsRestart && flashcardProgress.CompletedAt.Time.IsZero() {
		err := database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
			if err := s.FlashcardProgressionRepo.UpdateCompletedAt(ctx, tx, database.Text(req.StudySetId)); err != nil {
				return status.Errorf(codes.Internal, "s.FlashcardProgressionRepo.UpdateCompletedAt: %v", err)
			}

			if len(flashcardProgress.QuizExternalIDs.Elements) == len(flashcardProgress.RememberedQuestionIDs.Elements) && req.StudyPlanItemId != "" {
				if err := s.StudyPlanItemRepo.UpdateCompletedAtByID(ctx, tx, studyPlanItemID, pgtype.Timestamptz{Status: pgtype.Null}); err != nil {
					return status.Errorf(codes.Internal, "s.StudyPlanItemRepo.UpdateCompletedAtByID: %v", err)
				}
			}

			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	if req.IsRestart {
		if err := s.FlashcardProgressionRepo.DeleteByStudySetID(ctx, s.DB, database.Text(req.StudySetId)); err != nil {
			return nil, status.Errorf(codes.Internal, "s.FlashcardProgressionRepo.DeleteByStudySetID: %v", err)
		}
	}

	return &pb.FinishFlashCardStudyProgressResponse{
		IsSuccess: true,
	}, nil
}

func (s *CourseModifierService) UpdateFlashCardStudyProgress(ctx context.Context, req *pb.UpdateFlashCardStudyProgressRequest) (*pb.UpdateFlashCardStudyProgressResponse, error) {
	if req.StudySetId == "" {
		return nil, status.Error(codes.InvalidArgument, "req must have study set id")
	}
	if req.StudentId == "" {
		return nil, status.Error(codes.InvalidArgument, "req must have student id")
	}

	if int32(len(req.SkippedQuestionIds)+len(req.RememberedQuestionIds)) != req.StudyingIndex {
		return nil, status.Error(codes.InvalidArgument, "req must have studying index equal len(skippedQuestionIds) + len(rememberedQuestionIds)")
	}

	flashcardProgress, err := s.FlashcardProgressionRepo.GetByStudySetIDAndStudentID(ctx, s.DB, database.Text(req.StudentId), database.Text(req.StudySetId))
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("s.FlashcardProgressionRepo.GetByStudySetID: %v", err))
	}

	if !flashcardProgress.CompletedAt.Time.IsZero() {
		return nil, status.Error(codes.FailedPrecondition, "can't update FlashcardProgress when it was completed")
	}

	mQuizExternalIDs := make(map[string]bool)
	for _, quizExternalID := range flashcardProgress.QuizExternalIDs.Elements {
		mQuizExternalIDs[quizExternalID.String] = true
	}
	for _, skippedQuestionId := range req.SkippedQuestionIds {
		if _, ok := mQuizExternalIDs[skippedQuestionId]; !ok {
			return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("flashcardProgress don't contain skippedQuestionId %v", skippedQuestionId))
		}
	}
	for _, rememberedQuestionId := range req.RememberedQuestionIds {
		if _, ok := mQuizExternalIDs[rememberedQuestionId]; !ok {
			return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("flashcardProgress don't contain rememberedQuestionId %v", rememberedQuestionId))
		}
	}

	flashcardProgressUpdating := &entities.FlashcardProgression{
		OriginalQuizSetID:     flashcardProgress.OriginalQuizSetID,
		StudySetID:            flashcardProgress.StudySetID,
		OriginalStudySetID:    flashcardProgress.OriginalStudySetID,
		StudentID:             flashcardProgress.StudentID,
		StudyPlanItemID:       flashcardProgress.StudyPlanItemID,
		LoID:                  flashcardProgress.LoID,
		QuizExternalIDs:       flashcardProgress.QuizExternalIDs,
		StudyingIndex:         database.Int4(req.StudyingIndex),
		SkippedQuestionIDs:    database.TextArray(req.SkippedQuestionIds),
		RememberedQuestionIDs: database.TextArray(req.RememberedQuestionIds),
		UpdatedAt:             database.Timestamptz(time.Now()),
		CompletedAt:           flashcardProgress.CompletedAt,
		DeletedAt:             flashcardProgress.DeletedAt,
	}
	flashcardProgressUpdating.StudyPlanID.Set(nil)
	flashcardProgressUpdating.LearningMaterialID.Set(nil)
	if err := s.FlashcardProgressionRepo.Upsert(ctx, s.DB, []*entities.FlashcardProgression{flashcardProgressUpdating}); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("s.FlashcardProgressionRepo.Upsert: %v", err))
	}

	return &pb.UpdateFlashCardStudyProgressResponse{
		IsSuccess: true,
	}, nil
}

func (s *CourseModifierService) validateCourseBooks(ctx context.Context, courseID string, bookIDs []string) error {
	if len(bookIDs) == 0 {
		return status.Error(codes.InvalidArgument, "missing book id")
	}

	if courseID == "" {
		return status.Error(codes.InvalidArgument, "missing course id")
	}

	books, err := s.BookRepo.FindByIDs(ctx, s.DB, bookIDs)
	if err != nil {
		return status.Errorf(codes.Internal, "BookRepo.FindByIDs: %s", err.Error())
	}

	if len(books) != len(bookIDs) {
		return status.Error(codes.NotFound, "not found books")
	}

	return nil
}

func (s *CourseModifierService) toCoursesBooksEntity(courseID string, bookIDs []string) ([]*entities.CoursesBooks, error) {
	result := []*entities.CoursesBooks{}
	for _, bookID := range bookIDs {
		if bookID == "" {
			continue
		}

		r := &entities.CoursesBooks{}
		database.AllNullEntity(r)
		if err := multierr.Combine(
			r.BookID.Set(bookID),
			r.CourseID.Set(courseID),
		); err != nil {
			return nil, err
		}

		result = append(result, r)
	}

	return result, nil
}

func (s *CourseModifierService) AddBooks(ctx context.Context, req *pb.AddBooksRequest) (*pb.AddBooksResponse, error) {
	err := s.validateCourseBooks(ctx, req.CourseId, req.BookIds)
	if err != nil {
		return nil, err
	}

	coursesBooksEntities, err := s.toCoursesBooksEntity(req.CourseId, req.BookIds)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "toCoursesBooksEntity: %s", err.Error())
	}

	if err = s.CourseBookRepo.Upsert(ctx, s.DB, coursesBooksEntities); err != nil {
		return nil, status.Errorf(codes.Internal, "CourseBookRepo.Upsert: %s", err.Error())
	}

	return &pb.AddBooksResponse{
		Successful: true,
	}, nil
}

func (s *CourseModifierService) getStudentCountries(ctx context.Context, studentIDs []string) (map[string]string, error) {
	m := make(map[string]string)
	uniqueStudentIDs := golibs.GetUniqueElementStringArray(studentIDs)
	mdCtx, err := interceptors.GetOutgoingContext(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, codes.Unauthenticated.String())
	}
	resp, err := s.UsermgmtUserReaderService.SearchBasicProfile(mdCtx, &upb.SearchBasicProfileRequest{
		UserIds: uniqueStudentIDs,
		Paging:  &cpb.Paging{Limit: uint32(len(uniqueStudentIDs))},
	})
	if err != nil {
		return nil, err
	}
	for _, student := range resp.Profiles {
		m[student.UserId] = student.Country.String()
	}
	return m, nil
}
