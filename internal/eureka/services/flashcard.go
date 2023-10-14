package services

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/eureka/constants"
	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/timeutil"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type FlashcardService struct {
	sspb.UnimplementedFlashcardServer
	DB database.Ext

	TopicRepo interface {
		RetrieveByID(ctx context.Context, db database.QueryExecer, id pgtype.Text, enhancers ...repositories.QueryEnhancer) (*entities.Topic, error)
		UpdateLODisplayOrderCounter(ctx context.Context, db database.QueryExecer, topicID pgtype.Text, number pgtype.Int4) error
	}

	FlashcardRepo interface {
		Insert(ctx context.Context, db database.QueryExecer, e *entities.Flashcard) error
		Update(ctx context.Context, db database.QueryExecer, e *entities.Flashcard) error
		ListFlashcard(ctx context.Context, db database.QueryExecer, args *repositories.ListFlashcardArgs) ([]*entities.Flashcard, error)
		ListFlashcardBase(ctx context.Context, db database.QueryExecer, args *repositories.ListFlashcardArgs) ([]*entities.FlashcardBase, error)
	}

	QuizSetRepo interface {
		GetQuizSetByLoID(context.Context, database.QueryExecer, pgtype.Text) (*entities.QuizSet, error)
	}

	QuizRepo interface {
		GetByExternalIDsAndLmID(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray, lmID pgtype.Text) (entities.Quizzes, error)
	}

	FlashcardProgressionRepo interface {
		Create(ctx context.Context, db database.QueryExecer, flashcardProgression *entities.FlashcardProgression) (pgtype.Text, error)
		Get(ctx context.Context, db database.QueryExecer, args *repositories.GetFlashcardProgressionArgs) (*entities.FlashcardProgression, error)
		GetByStudySetID(ctx context.Context, db database.QueryExecer, studySetID pgtype.Text) (*entities.FlashcardProgression, error)
		UpdateCompletedAt(ctx context.Context, db database.QueryExecer, studySetID pgtype.Text) error
		DeleteByStudySetID(ctx context.Context, db database.QueryExecer, studySetID pgtype.Text) error
		GetLastFlashcardProgressionV2(ctx context.Context, db database.QueryExecer, itemIdentity *repositories.StudyPlanItemIdentity, isCompleted pgtype.Bool) (*entities.FlashcardProgression, error)
	}
}

func NewFlashcardService(db database.Ext) sspb.FlashcardServer {
	return &FlashcardService{
		DB:                       db,
		TopicRepo:                new(repositories.TopicRepo),
		FlashcardRepo:            new(repositories.FlashcardRepo),
		QuizSetRepo:              new(repositories.QuizSetRepo),
		QuizRepo:                 new(repositories.QuizRepo),
		FlashcardProgressionRepo: new(repositories.FlashcardProgressionRepo),
	}
}

func validateFlashcardReq(req *sspb.FlashcardBase) error {
	if req.Base.TopicId == "" {
		return fmt.Errorf("Topic ID must not be empty")
	}
	return nil
}

func toFlashcardEnt(req *sspb.FlashcardBase) (*entities.Flashcard, error) {
	e := &entities.Flashcard{}
	database.AllNullEntity(e)
	id := req.Base.LearningMaterialId
	if id == "" {
		id = idutil.ULIDNow()
	}
	now := time.Now()
	if err := multierr.Combine(
		e.ID.Set(id),
		e.TopicID.Set(req.Base.TopicId),
		e.Name.Set(req.Base.Name),
		e.Type.Set(sspb.LearningMaterialType_LEARNING_MATERIAL_FLASH_CARD.String()),
		e.IsPublished.Set(false),
		e.VendorType.Set(req.Base.VendorType.String()),
		e.CreatedAt.Set(now),
		e.UpdatedAt.Set(now),
	); err != nil {
		return nil, err
	}
	return e, nil
}

func (s *FlashcardService) InsertFlashcard(ctx context.Context, req *sspb.InsertFlashcardRequest) (*sspb.InsertFlashcardResponse, error) {
	if err := validateFlashcardReq(req.Flashcard); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Errorf("validateFlashcardReq: %w", err).Error())
	}
	fc, err := toFlashcardEnt(req.Flashcard)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Errorf("toFlashcardEnt: %w", err).Error())
	}
	if err := database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		tp, err := s.TopicRepo.RetrieveByID(ctx, tx, fc.TopicID, repositories.WithUpdateLock())
		if err != nil {
			return fmt.Errorf("s.TopicRepo.RetrieveByID: %w", err)
		}
		if err := fc.DisplayOrder.Set(tp.LODisplayOrderCounter.Int + 1); err != nil {
			return fmt.Errorf("fc.DisplayOrder.Set: %w", err)
		}
		if err := s.FlashcardRepo.Insert(ctx, tx, fc); err != nil {
			return fmt.Errorf("s.FlashcardRepo.Insert: %w", err)
		}
		if err := s.TopicRepo.UpdateLODisplayOrderCounter(ctx, tx, tp.ID, database.Int4(1)); err != nil {
			return fmt.Errorf("s.TopicRepo.UpdateLODisplayOrderCounter: %w", err)
		}
		return nil
	}); err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Errorf("database.ExecInTx: %w", err).Error())
	}
	return &sspb.InsertFlashcardResponse{
		LearningMaterialId: fc.LearningMaterial.ID.String,
	}, nil
}

func validateFlashcardUpdateReq(req *sspb.FlashcardBase) error {
	if req.Base.LearningMaterialId == "" {
		return fmt.Errorf("LearningMaterialId must not be empty")
	}
	return nil
}

func (s *FlashcardService) UpdateFlashcard(ctx context.Context, req *sspb.UpdateFlashcardRequest) (*sspb.UpdateFlashcardResponse, error) {
	if err := validateFlashcardUpdateReq(req.Flashcard); err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Errorf("validateFlashcardUpdateReq: %w", err).Error())
	}
	fc, err := toFlashcardEnt(req.Flashcard)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Errorf("toFlashcardEnt: %w", err).Error())
	}
	if err := database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		if err := s.FlashcardRepo.Update(ctx, tx, fc); err != nil {
			return fmt.Errorf("s.FlashcardRepo.Update: %w", err)
		}
		return nil
	}); err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Errorf("database.ExecInTx: %w", err).Error())
	}
	return &sspb.UpdateFlashcardResponse{}, nil
}

func validateListFlashcardReq(req *sspb.ListFlashcardRequest) error {
	if len(req.LearningMaterialIds) == 0 {
		return fmt.Errorf("LearningMaterialIds must greater than 0")
	}
	return nil
}

func ToBaseFlashcard(e *entities.Flashcard) *sspb.FlashcardBase {
	return &sspb.FlashcardBase{
		Base: &sspb.LearningMaterialBase{
			LearningMaterialId: e.ID.String,
			TopicId:            e.TopicID.String,
			Name:               e.Name.String,
			Type:               e.Type.String,
			DisplayOrder: &wrapperspb.Int32Value{
				Value: int32(e.DisplayOrder.Int),
			},
		},
	}
}

func (s *FlashcardService) ListFlashcard(ctx context.Context, req *sspb.ListFlashcardRequest) (*sspb.ListFlashcardResponse, error) {
	if err := validateListFlashcardReq(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Errorf("validateListFlashcardReq: %w", err).Error())
	}
	args := &repositories.ListFlashcardArgs{
		LearningMaterialIDs: pgtype.TextArray{Status: pgtype.Null},
	}
	if err := args.LearningMaterialIDs.Set(req.LearningMaterialIds); err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Errorf("args.LearningMaterialIDs.Set: %w", err).Error())
	}

	flashcards, err := s.FlashcardRepo.ListFlashcardBase(ctx, s.DB, args)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "cannot find flashcards")
		}
		return nil, status.Errorf(codes.Internal, fmt.Errorf("s.FlashcardRepo.ListFlashcardBase: %w", err).Error())
	}

	baseFlashcards := make([]*sspb.FlashcardBase, 0, len(flashcards))
	for i, flashcard := range flashcards {
		fc := ToBaseFlashcard(&(flashcards[i].Flashcard))
		fc.TotalQuestion = flashcard.TotalQuestion.Int

		baseFlashcards = append(baseFlashcards, fc)
	}

	return &sspb.ListFlashcardResponse{
		Flashcards: baseFlashcards,
	}, nil
}

func (s *FlashcardService) CreateFlashCardStudy(ctx context.Context, req *sspb.CreateFlashCardStudyRequest) (*sspb.CreateFlashCardStudyResponse, error) {
	var (
		quizExternalIDs   pgtype.TextArray
		originalQuizSetID pgtype.Text
		shouldCreate      bool
	)

	if err := s.validateCreateFlashcardStudyRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	flashcardProgression := &entities.FlashcardProgression{}

	// create flashcard study when study_set_id is empty or old flashcard study have value with field completed_at and length of skipped_questions_ids > 0
	// return old value flashcard study when study_set_id (req) is not empty and isn't complete
	if req.StudySetId == "" {
		// get quizset of the learning objective
		quizSet, err := s.QuizSetRepo.GetQuizSetByLoID(ctx, s.DB, database.Text(req.LmId))
		if err != nil {
			return nil, status.Errorf(codes.Internal, fmt.Errorf("s.QuizSetRepo.GetQuizSetByLoID: %v", err).Error())
		}

		quizExternalIDs = quizSet.QuizExternalIDs
		originalQuizSetID = quizSet.ID
		shouldCreate = true
	} else {
		baseFlashcardProgression, err := s.FlashcardProgressionRepo.GetByStudySetID(ctx, s.DB, database.Text(req.StudySetId))
		if err != nil {
			return nil, status.Errorf(codes.Internal, fmt.Errorf("s.FlashcardProgressionRepo.GetByStudySetID: %v", err).Error())
		}

		if !baseFlashcardProgression.CompletedAt.Time.IsZero() {
			if len(baseFlashcardProgression.SkippedQuestionIDs.Elements) == 0 {
				return nil, status.Error(codes.InvalidArgument, "length of skippedQuestionIDs must be greater than 0")
			}

			shouldCreate = true
			quizExternalIDs = baseFlashcardProgression.SkippedQuestionIDs
			originalQuizSetID = baseFlashcardProgression.OriginalQuizSetID
		}
	}

	if shouldCreate {
		studyPlanID := database.Text(req.StudyPlanId)
		if req.StudyPlanId == "" {
			studyPlanID = pgtype.Text{Status: pgtype.Null}
		}
		var err error
		if flashcardProgression, err = s.toFlashcardProgressionEnt(req, originalQuizSetID, studyPlanID, quizExternalIDs); err != nil {
			return nil, status.Error(codes.Internal, fmt.Errorf("s.toFlashcardProgressionEnt: %v", err).Error())
		}

		if !req.KeepOrder {
			// shuffle quiz external ids
			flashcardProgression.Shuffle()
		}
		studySetID, err := s.FlashcardProgressionRepo.Create(ctx, s.DB, flashcardProgression)
		if err != nil {
			return nil, status.Errorf(codes.Internal, fmt.Errorf("s.FlashcardProgressionRepo.Create: %v", err).Error())
		}

		flashcardProgression.StudySetID = studySetID
	} else {
		if err := flashcardProgression.StudySetID.Set(req.StudySetId); err != nil {
			return nil, status.Errorf(codes.Internal, fmt.Errorf("flashcardProgression.StudySetID.Set: %w", err).Error())
		}
	}

	// Get our flashcard Progression with paging
	pbQuizzes, pagingFlashcardProgression, err := s.getFlashcardProgressionWithPaging(ctx, req.Paging, flashcardProgression.StudySetID, database.Text(req.StudentId))
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Errorf("s.getFlashcardProgressionWithPaging: %v", err).Error())
	}

	resp := &sspb.CreateFlashCardStudyResponse{
		NextPage: &cpb.Paging{
			Limit: req.Paging.Limit,
			Offset: &cpb.Paging_OffsetInteger{
				OffsetInteger: req.Paging.GetOffsetInteger() + int64(req.Paging.Limit),
			},
		},
		StudySetId:    pagingFlashcardProgression.StudySetID.String,
		Items:         pbQuizzes,
		StudyingIndex: pagingFlashcardProgression.StudyingIndex.Int,
	}
	return resp, nil
}

func (s *FlashcardService) validateCreateFlashcardStudyRequest(req *sspb.CreateFlashCardStudyRequest) error {
	if req.StudentId == "" {
		return fmt.Errorf("req must have student id")
	}
	if req.LmId == "" {
		return fmt.Errorf("req must have learning material id")
	}

	if req.Paging == nil {
		return fmt.Errorf("req must have paging field")
	}

	if req.Paging.GetOffsetInteger() <= 0 {
		return fmt.Errorf("offset must be positive")
	}

	if req.Paging.Limit <= 0 {
		req.Paging.Limit = 100
	}
	return nil
}

func (s *FlashcardService) toFlashcardProgressionEnt(
	req *sspb.CreateFlashCardStudyRequest,
	originalQuizSetID, studyPlanID pgtype.Text, quizExternalIDs pgtype.TextArray,
) (*entities.FlashcardProgression, error) {
	e := &entities.FlashcardProgression{}
	database.AllNullEntity(e)
	now := timeutil.Now()
	if err := multierr.Combine(
		e.OriginalQuizSetID.Set(originalQuizSetID),
		e.StudySetID.Set(idutil.ULIDNow()),
		e.OriginalStudySetID.Set(req.StudySetId),
		e.StudentID.Set(req.StudentId),
		e.StudyPlanID.Set(studyPlanID),
		e.LearningMaterialID.Set(req.LmId),
		e.QuizExternalIDs.Set(quizExternalIDs.Elements),
		e.StudyingIndex.Set(nil),
		e.SkippedQuestionIDs.Set(nil),
		e.RememberedQuestionIDs.Set(nil),
		e.CreatedAt.Set(now),
		e.UpdatedAt.Set(now),
		e.CompletedAt.Set(nil),
		e.DeletedAt.Set(nil),
	); err != nil {
		return nil, fmt.Errorf("set data: %w", err)
	}
	return e, nil
}

/*
	GetFlashcardProgressionWithPaging

1. get flashcard by study_set_id, student_id, from, to
2. get quizzes by externalIDs from flashcardProgression
3. convert quizzes to flashcardQuizzes
*/
func (s *FlashcardService) getFlashcardProgressionWithPaging(
	ctx context.Context, paging *cpb.Paging, studySetID, studentID pgtype.Text,
) ([]*sspb.FlashcardQuizzes, *entities.FlashcardProgression, error) {
	offset := paging.GetOffsetInteger()
	limit := paging.Limit
	from := database.Int8(offset)
	to := database.Int8(offset + int64(limit) - 1)

	pagingFlashcardProgression, err := s.FlashcardProgressionRepo.Get(ctx, s.DB, &repositories.GetFlashcardProgressionArgs{
		StudySetID:      studySetID,
		StudentID:       studentID,
		LoID:            pgtype.Text{Status: pgtype.Null},
		StudyPlanItemID: pgtype.Text{Status: pgtype.Null},
		LmID:            pgtype.Text{Status: pgtype.Null},
		StudyPlanID:     pgtype.Text{Status: pgtype.Null},
		From:            from,
		To:              to,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("s.FlashcardProgressionRepo.Get: %w", err)
	}
	skippedQuestionIDsMap := make(map[string]int64)
	for i, externalID := range pagingFlashcardProgression.SkippedQuestionIDs.Elements {
		skippedQuestionIDsMap[externalID.String] = int64(i)
	}
	rememberedQuestionIDsMap := make(map[string]int64)
	for i, externalID := range pagingFlashcardProgression.RememberedQuestionIDs.Elements {
		rememberedQuestionIDsMap[externalID.String] = int64(i)
	}

	// Get list of quiz item - entities.Quizzes
	quizzes, err := s.QuizRepo.GetByExternalIDsAndLmID(ctx, s.DB, pagingFlashcardProgression.QuizExternalIDs, pagingFlashcardProgression.LearningMaterialID)
	if err != nil {
		return nil, nil, fmt.Errorf("s.QuizRepo.GetByExternalIDsAndLmID: %w", err)
	}
	// Convert from entities.Quizzes to []pb.FlashcardProgression
	pbQuizzes, err := s.toListFlashcardQuizzes(pagingFlashcardProgression.LoID.String, quizzes, skippedQuestionIDsMap, rememberedQuestionIDsMap)
	if err != nil {
		return nil, nil, fmt.Errorf("s.toListFlashcardQuizzes: %w", err)
	}
	return pbQuizzes, pagingFlashcardProgression, nil
}

func (s *FlashcardService) toListFlashcardQuizzes(
	loID string, quizzes entities.Quizzes,
	skippedQuestionIDsMap, rememberedQuestionIDsMap map[string]int64,
) ([]*sspb.FlashcardQuizzes, error) {
	pbQuizzes := []*sspb.FlashcardQuizzes{}
	for _, quiz := range quizzes {
		pbQuiz, err := toQuizPb(loID, quiz)
		if err != nil {
			return nil, err
		}
		flashcardQuiz := &sspb.FlashcardQuizzes{
			Item:   pbQuiz,
			Status: sspb.FlashcardQuizStudyStatus_FLASHCARD_QUIZ_STUDY_STATUS_NONE,
		}
		if _, ok := skippedQuestionIDsMap[quiz.ExternalID.String]; ok {
			flashcardQuiz.Status = sspb.FlashcardQuizStudyStatus_FLASHCARD_QUIZ_STUDY_STATUS_SKIPPED
		}
		if _, ok := rememberedQuestionIDsMap[quiz.ExternalID.String]; ok {
			flashcardQuiz.Status = sspb.FlashcardQuizStudyStatus_FLASHCARD_QUIZ_STUDY_STATUS_REMEMBERED
		}
		pbQuizzes = append(pbQuizzes, flashcardQuiz)
	}
	return pbQuizzes, nil
}

func (s *FlashcardService) GetLastestProgress(ctx context.Context, req *sspb.GetLastestProgressRequest) (*sspb.GetLastestProgressResponse, error) {
	userGroup := interceptors.UserGroupFromContext(ctx)
	if userGroup == constants.RoleStudent {
		userID := interceptors.UserIDFromContext(ctx)
		req.StudyPlanItemIdentity.StudentId = wrapperspb.String(userID)
	}
	if err := s.validateGetLastestProgress(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	var studySetID string

	flashcardProgression, err := s.FlashcardProgressionRepo.GetLastFlashcardProgressionV2(ctx, s.DB,
		&repositories.StudyPlanItemIdentity{
			StudentID:          database.Text(req.StudyPlanItemIdentity.GetStudentId().GetValue()),
			StudyPlanID:        database.Text(req.StudyPlanItemIdentity.StudyPlanId),
			LearningMaterialID: database.Text(req.StudyPlanItemIdentity.LearningMaterialId),
		}, database.Bool(req.IsCompleted))
	if err != nil && err != pgx.ErrNoRows {
		return nil, status.Error(codes.Internal, fmt.Errorf("c.FlashcardProgressionRepo.GetLastFlashcardProgression: %v", err).Error())
	}

	if flashcardProgression != nil {
		studySetID = flashcardProgression.StudySetID.String
	}

	return &sspb.GetLastestProgressResponse{
		StudySetId: wrapperspb.String(studySetID),
	}, nil
}

func (c *FlashcardService) validateGetLastestProgress(req *sspb.GetLastestProgressRequest) error {
	if req.StudyPlanItemIdentity.GetStudentId() == nil && req.StudyPlanItemIdentity.GetStudentId().Value == "" {
		return errors.New("req must have student id")
	}
	if req.StudyPlanItemIdentity.LearningMaterialId == "" {
		return errors.New("req must have learning material id")
	}

	if req.StudyPlanItemIdentity.StudyPlanId == "" {
		return errors.New("req must have study plan id")
	}

	return nil
}

func (s *FlashcardService) FinishFlashCardStudy(ctx context.Context, req *sspb.FinishFlashCardStudyRequest) (*sspb.FinishFlashCardStudyResponse, error) {
	if err := s.validateFinishFlashCardStudyRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Errorf("s.ValidateFinishFlashCardStudyRequest: %w", err).Error())
	}

	flashcardProgress, err := s.FlashcardProgressionRepo.Get(ctx, s.DB, &repositories.GetFlashcardProgressionArgs{
		StudySetID:  database.Text(req.StudySetId),
		StudyPlanID: database.Text(req.StudyPlanItemIdentity.StudyPlanId),
		LmID:        database.Text(req.StudyPlanItemIdentity.LearningMaterialId),
		StudentID:   database.Text(req.StudyPlanItemIdentity.StudentId.Value),
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "s.FlashcardProgressionRepo.Get: %v", err)
	}

	if err := database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		if !req.IsRestart && flashcardProgress.CompletedAt.Time.IsZero() {
			if err := s.FlashcardProgressionRepo.UpdateCompletedAt(ctx, tx, database.Text(req.StudySetId)); err != nil {
				return fmt.Errorf("s.FlashcardProgressionRepo.UpdateCompletedAt: %v", err)
			}
		}

		if req.IsRestart {
			if err := s.FlashcardProgressionRepo.DeleteByStudySetID(ctx, tx, database.Text(req.StudySetId)); err != nil {
				return fmt.Errorf("s.FlashcardProgressionRepo.DeleteByStudySetID: %v", err)
			}
		}

		return nil
	}); err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("database.ExecInTx: %w", err).Error())
	}

	return &sspb.FinishFlashCardStudyResponse{}, nil
}

func (s *FlashcardService) validateFinishFlashCardStudyRequest(req *sspb.FinishFlashCardStudyRequest) error {
	if req.StudySetId == "" {
		return fmt.Errorf("StudySetId must not be empty")
	}

	if req.StudyPlanItemIdentity == nil {
		return fmt.Errorf("StudyPlanItemIdentity must not be empty")
	}

	if req.StudyPlanItemIdentity.StudyPlanId == "" {
		return fmt.Errorf("StudyPlanId must not be empty")
	}

	if req.StudyPlanItemIdentity.LearningMaterialId == "" {
		return fmt.Errorf("LearningMaterialId must not be empty")
	}

	if req.StudyPlanItemIdentity.StudentId == nil || req.StudyPlanItemIdentity.StudentId.Value == "" {
		return fmt.Errorf("StudentId must not be empty")
	}
	return nil
}
