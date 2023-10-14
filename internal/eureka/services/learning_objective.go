package services

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/nats"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type LearningObjectiveService struct {
	sspb.UnimplementedLearningObjectiveServer
	DB  database.Ext
	JSM nats.JetStreamManagement

	TopicRepo interface {
		RetrieveByID(ctx context.Context, db database.QueryExecer, id pgtype.Text, enhancers ...repositories.QueryEnhancer) (*entities.Topic, error)
		UpdateLODisplayOrderCounter(ctx context.Context, db database.QueryExecer, topicID pgtype.Text, number pgtype.Int4) error
	}

	LearningObjectiveRepo interface {
		Insert(ctx context.Context, db database.QueryExecer, e *entities.LearningObjectiveV2) error
		Update(ctx context.Context, db database.QueryExecer, e *entities.LearningObjectiveV2) error
		ListByIDs(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) ([]*entities.LearningObjectiveV2, error)
		ListLOBaseByIDs(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) ([]*entities.LearningObjectiveBaseV2, error)
	}

	QuizRepo interface {
		GetByExternalIDsAndLmID(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray, lmID pgtype.Text) (entities.Quizzes, error)
	}

	ShuffledQuizSetRepo interface {
		GetExternalIDs(ctx context.Context, db database.QueryExecer, shuffleQuizSetID pgtype.Text) (pgtype.TextArray, error)
		GetSeed(context.Context, database.QueryExecer, pgtype.Text) (pgtype.Text, error)
	}

	LOProgressionRepo interface {
		Upsert(ctx context.Context, db database.QueryExecer, e *entities.LOProgression) error
		GetByStudyPlanItemIdentity(ctx context.Context, db database.QueryExecer, arg repositories.StudyPlanItemIdentity, from, to pgtype.Int8) (*entities.LOProgression, error)
	}

	LOProgressionAnswerRepo interface {
		BulkUpsert(ctx context.Context, db database.QueryExecer, items []*entities.LOProgressionAnswer) error
		ListByProgressionAndExternalIDs(ctx context.Context, db database.QueryExecer, progressionID pgtype.Text, externalIDs pgtype.TextArray) (entities.LOProgressionAnswers, error)
	}

	QuestionGroupRepo interface {
		GetQuestionGroupsByIDs(ctx context.Context, db database.QueryExecer, ids ...string) (entities.QuestionGroups, error)
	}
}

func NewLearningObjectiveService(db database.Ext, jsm nats.JetStreamManagement) *LearningObjectiveService {
	return &LearningObjectiveService{
		DB:                      db,
		JSM:                     jsm,
		TopicRepo:               new(repositories.TopicRepo),
		LearningObjectiveRepo:   new(repositories.LearningObjectiveRepoV2),
		QuizRepo:                new(repositories.QuizRepo),
		ShuffledQuizSetRepo:     new(repositories.ShuffledQuizSetRepo),
		LOProgressionRepo:       new(repositories.LOProgressionRepo),
		LOProgressionAnswerRepo: new(repositories.LOProgressionAnswerRepo),
		QuestionGroupRepo:       new(repositories.QuestionGroupRepo),
	}
}
func validateLearningObjectiveReq(req *sspb.LearningObjectiveBase) error {
	if req.Base.TopicId == "" {
		return fmt.Errorf("Topic ID must not be empty")
	}
	return nil
}

func toInsertLearningObjectiveEnt(req *sspb.LearningObjectiveBase) (*entities.LearningObjectiveV2, error) {
	e := &entities.LearningObjectiveV2{}
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
		e.Type.Set(sspb.LearningMaterialType_LEARNING_MATERIAL_LEARNING_OBJECTIVE.String()),
		e.VendorType.Set(req.Base.VendorType.String()),
		e.CreatedAt.Set(now),
		e.UpdatedAt.Set(now),
		e.ManualGrading.Set(req.ManualGrading),
		e.IsPublished.Set(false),
	); err != nil {
		return nil, err
	}
	return e, nil
}

func (s *LearningObjectiveService) InsertLearningObjective(ctx context.Context, req *sspb.InsertLearningObjectiveRequest) (*sspb.InsertLearningObjectiveResponse, error) {
	if err := validateLearningObjectiveReq(req.LearningObjective); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Errorf("validateLearningObjectiveReq: %w", err).Error())
	}
	lo, err := toInsertLearningObjectiveEnt(req.LearningObjective)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Errorf("toLearningObjectiveEnt: %w", err).Error())
	}
	if err := database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		topic, err := s.TopicRepo.RetrieveByID(ctx, tx, database.Text(lo.TopicID.String), repositories.WithUpdateLock())
		if err != nil {
			return fmt.Errorf("s.TopicRepo.RetrieveByID: %w", err)
		}

		if err := lo.DisplayOrder.Set(topic.LODisplayOrderCounter.Int + 1); err != nil {
			return fmt.Errorf("lo.DisplayOrder.Set: %w", err)
		}
		if err := s.LearningObjectiveRepo.Insert(ctx, tx, lo); err != nil {
			return fmt.Errorf("unable to bulk import learning objective: %w", err)
		}
		if err := s.TopicRepo.UpdateLODisplayOrderCounter(ctx, tx, topic.ID, database.Int4(1)); err != nil {
			return fmt.Errorf("unable to update lo display order counter: %w", err)
		}

		return nil
	}); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.ForeignKeyViolation {
			return nil, status.Error(codes.FailedPrecondition, err.Error())
		}
		return nil, status.Errorf(codes.Internal, fmt.Errorf("database.ExecInTx: %w", err).Error())
	}

	return &sspb.InsertLearningObjectiveResponse{
		LearningMaterialId: lo.LearningMaterial.ID.String,
	}, nil
}

func ToBaseLearningObjectiveV2(e *entities.LearningObjectiveV2) *sspb.LearningObjectiveBase {
	return &sspb.LearningObjectiveBase{
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
func validateLearningObjectiveUpdateReq(req *sspb.LearningObjectiveBase) error {
	if req.Base.LearningMaterialId == "" {
		return fmt.Errorf("LearningMaterialId must not be empty")
	}
	return nil
}

func toUpdateLearningObjectiveEnt(req *sspb.LearningObjectiveBase) (*entities.LearningObjectiveV2, error) {
	e := &entities.LearningObjectiveV2{}
	database.AllNullEntity(e)

	now := time.Now()
	if err := multierr.Combine(
		e.ID.Set(req.Base.LearningMaterialId),
		e.Name.Set(req.Base.Name),
		e.UpdatedAt.Set(now),
		e.Video.Set(req.VideoId),
		e.StudyGuide.Set(req.StudyGuide),
		e.VideoScript.Set(req.VideoScript),
	); err != nil {
		return nil, err
	}
	return e, nil
}

func (s *LearningObjectiveService) UpdateLearningObjective(ctx context.Context, req *sspb.UpdateLearningObjectiveRequest) (*sspb.UpdateLearningObjectiveResponse, error) {
	if err := validateLearningObjectiveUpdateReq(req.LearningObjective); err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Errorf("validateLearningObjectiveUpdateReq: %w", err).Error())
	}
	lo, err := toUpdateLearningObjectiveEnt(req.LearningObjective)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Errorf("toLearningObjectiveEnt: %w", err).Error())
	}
	if err := database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		if err := s.LearningObjectiveRepo.Update(ctx, tx, lo); err != nil {
			return fmt.Errorf("s.LearningObjectiveRepo.Update: %w", err)
		}
		return nil
	}); err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Errorf("database.ExecInTx: %w", err).Error())
	}
	return &sspb.UpdateLearningObjectiveResponse{}, nil
}

func (s *LearningObjectiveService) ListLearningObjective(ctx context.Context, req *sspb.ListLearningObjectiveRequest) (*sspb.ListLearningObjectiveResponse, error) {
	ids := req.LearningMaterialIds
	if len(ids) == 0 {
		return nil, status.Error(codes.InvalidArgument, "LearningMaterialIds must not be empty")
	}

	los, err := s.LearningObjectiveRepo.ListLOBaseByIDs(ctx, s.DB, database.TextArray(ids))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, status.Error(codes.NotFound, fmt.Errorf("s.LearningObjectiveRepo.ListLOBaseByIDs: %w", err).Error())
		}
		return nil, status.Errorf(codes.Internal, fmt.Errorf("s.LearningObjectiveRepo.ListLOBaseByIDs: %w", err).Error())
	}

	rspLos := make([]*sspb.LearningObjectiveBase, 0, len(los))
	for _, lo := range los {
		sspbLo := &sspb.LearningObjectiveBase{
			Base: &sspb.LearningMaterialBase{
				LearningMaterialId: lo.LearningMaterial.ID.String,
				TopicId:            lo.TopicID.String,
				Name:               lo.Name.String,
				Type:               lo.Type.String,
				DisplayOrder:       wrapperspb.Int32(int32(lo.DisplayOrder.Int)),
			},
			VideoId:       lo.Video.String,
			StudyGuide:    lo.StudyGuide.String,
			VideoScript:   lo.VideoScript.String,
			ManualGrading: lo.ManualGrading.Bool,
			TotalQuestion: lo.TotalQuestion.Int,
		}
		rspLos = append(rspLos, sspbLo)
	}
	return &sspb.ListLearningObjectiveResponse{
		LearningObjectives: rspLos,
	}, nil
}

func (s *LearningObjectiveService) UpsertLOProgression(ctx context.Context, req *sspb.UpsertLOProgressionRequest) (*sspb.UpsertLOProgressionResponse, error) {
	if err := validateUpsertLOProgressionRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	quizExternalIDs, err := s.ShuffledQuizSetRepo.GetExternalIDs(ctx, s.DB, database.Text(req.ShuffledQuizSetId))
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("s.UpsertLOProgression: %w", err).Error())
	}

	loProgression, err := s.LOProgressionRepo.GetByStudyPlanItemIdentity(ctx, s.DB, repositories.StudyPlanItemIdentity{
		StudentID:          database.Text(req.StudyPlanItemIdentity.StudentId.Value),
		StudyPlanID:        database.Text(req.StudyPlanItemIdentity.StudyPlanId),
		LearningMaterialID: database.Text(req.StudyPlanItemIdentity.LearningMaterialId),
	}, database.Int8(0), database.Int8(0))
	if err != nil && err != pgx.ErrNoRows {
		return nil, status.Error(codes.Internal, fmt.Errorf("s.LOProgressionRepo.GetByStudyPlanItemIdentity: %w", err).Error())
	}

	quizIDsMap, quizIDFromReqMap := make(map[string]bool), make(map[string]*sspb.QuizAnswer)

	now := time.Now()
	loProgressionID := idutil.ULIDNow()
	if loProgression != nil {
		loProgressionID = loProgression.ProgressionID.String

		loProgressionAnswers, err := s.LOProgressionAnswerRepo.ListByProgressionAndExternalIDs(ctx, s.DB, database.Text(loProgressionID), quizExternalIDs)
		if err != nil {
			return nil, status.Error(codes.Internal, fmt.Errorf("s.LOProgressionAnswerRepo.ListByProgressionAndExternalIDs: %w", err).Error())
		}
		for _, loProgressionAnswer := range loProgressionAnswers {
			quizIDsMap[loProgressionAnswer.QuizExternalID.String] = true
		}
	}

	for _, quizAnswer := range req.QuizAnswer {
		quizIDsMap[quizAnswer.QuizId] = true
		quizIDFromReqMap[quizAnswer.QuizId] = quizAnswer
	}

	err = database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		loProgression := &entities.LOProgression{}
		database.AllNullEntity(loProgression)
		if err := multierr.Combine(
			loProgression.ProgressionID.Set(loProgressionID),
			loProgression.ShuffledQuizSetID.Set(req.ShuffledQuizSetId),
			loProgression.LastIndex.Set(int32(req.LastIndex)),
			loProgression.QuizExternalIDs.Set(quizExternalIDs.Elements),
			loProgression.StudentID.Set(req.StudyPlanItemIdentity.StudentId.Value),
			loProgression.StudyPlanID.Set(req.StudyPlanItemIdentity.StudyPlanId),
			loProgression.LearningMaterialID.Set(req.StudyPlanItemIdentity.LearningMaterialId),
			loProgression.SessionID.Set(req.SessionId),
			loProgression.CreatedAt.Set(now),
			loProgression.UpdatedAt.Set(now),
		); err != nil {
			return status.Error(codes.Internal, fmt.Errorf("multierr.Combine: %w", err).Error())
		}
		if err := s.LOProgressionRepo.Upsert(ctx, tx, loProgression); err != nil {
			return status.Error(codes.Internal, fmt.Errorf("s.LOProgressionRepo.Upsert: %w", err).Error())
		}

		loProgressionAnswers := make([]*entities.LOProgressionAnswer, 0, len(req.QuizAnswer))
		for quizID := range quizIDsMap {
			var (
				studentTextAnswers, submittedKeyAnswers []string
				studentIndexAnswers                     []uint32
			)

			if quizAnswer, ok := quizIDFromReqMap[quizID]; ok {
				for _, answer := range quizAnswer.Answer {
					if _, ok := answer.GetFormat().(*sspb.Answer_FilledText); ok {
						studentTextAnswers = append(studentTextAnswers, answer.GetFilledText())
					}

					if _, ok := answer.GetFormat().(*sspb.Answer_SelectedIndex); ok {
						studentIndexAnswers = append(studentIndexAnswers, answer.GetSelectedIndex())
					}

					if _, ok := answer.GetFormat().(*sspb.Answer_SubmittedKey); ok {
						submittedKeyAnswers = append(submittedKeyAnswers, answer.GetSubmittedKey())
					}
				}
			}

			loProgressionAnswer := &entities.LOProgressionAnswer{}
			database.AllNullEntity(loProgressionAnswer)
			if err := multierr.Combine(
				loProgressionAnswer.ProgressionAnswerID.Set(idutil.ULIDNow()),
				loProgressionAnswer.ShuffledQuizSetID.Set(req.ShuffledQuizSetId),
				loProgressionAnswer.QuizExternalID.Set(quizID),
				loProgressionAnswer.ProgressionID.Set(loProgressionID),
				loProgressionAnswer.StudentID.Set(req.StudyPlanItemIdentity.StudentId.Value),
				loProgressionAnswer.StudyPlanID.Set(req.StudyPlanItemIdentity.StudyPlanId),
				loProgressionAnswer.LearningMaterialID.Set(req.StudyPlanItemIdentity.LearningMaterialId),
				loProgressionAnswer.StudentTextAnswers.Set(studentTextAnswers),
				loProgressionAnswer.StudentIndexAnswers.Set(studentIndexAnswers),
				loProgressionAnswer.SubmittedKeysAnswer.Set(submittedKeyAnswers),
				loProgressionAnswer.CreatedAt.Set(now),
				loProgressionAnswer.UpdatedAt.Set(now),
			); err != nil {
				return status.Error(codes.Internal, fmt.Errorf("multierr.Combine: %w", err).Error())
			}
			loProgressionAnswers = append(loProgressionAnswers, loProgressionAnswer)
		}
		if err := s.LOProgressionAnswerRepo.BulkUpsert(ctx, tx, loProgressionAnswers); err != nil {
			return status.Error(codes.Internal, fmt.Errorf("s.LOProgressionAnswerRepo.BulkUpsert: %w", err).Error())
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &sspb.UpsertLOProgressionResponse{
		ProgressionId: loProgressionID,
		UpdatedAt:     timestamppb.New(now),
	}, nil
}

func (s *LearningObjectiveService) validateRetrieveExamLOProgression(identity *sspb.StudyPlanItemIdentity) error {
	if identity.GetStudentId() == nil {
		return fmt.Errorf("req must have student id")
	}
	if identity.GetLearningMaterialId() == "" {
		return fmt.Errorf("req must have learning material id")
	}
	if identity.GetStudyPlanId() == "" {
		return fmt.Errorf("req must have study plan id")
	}
	return nil
}

func (s *LearningObjectiveService) RetrieveLOProgression(ctx context.Context, req *sspb.RetrieveLOProgressionRequest) (*sspb.RetrieveLOProgressionResponse, error) {
	if err := s.validateRetrieveExamLOProgression(req.GetStudyPlanItemIdentity()); err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("Validate: %s", err.Error()))
	}

	// Get our shuffledQuizset with paging
	offset := req.GetPaging().GetOffsetInteger()
	limit := req.GetPaging().GetLimit()
	from := database.Int8(offset)
	to := database.Int8(offset + int64(limit) - 1)

	identity := repositories.StudyPlanItemIdentity{
		StudentID:          database.Text(req.GetStudyPlanItemIdentity().GetStudentId().GetValue()),
		StudyPlanID:        database.Text(req.GetStudyPlanItemIdentity().GetStudyPlanId()),
		LearningMaterialID: database.Text(req.GetStudyPlanItemIdentity().GetLearningMaterialId()),
	}

	// Get lo_progression
	progression, err := s.LOProgressionRepo.GetByStudyPlanItemIdentity(ctx, s.DB, identity, from, to)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, fmt.Errorf("LOProgressionRepo.GetByStudyPlanItemIdentity: %w", err).Error())
	}

	// Get lo_progression_answer
	progressionAnswers, err := s.LOProgressionAnswerRepo.ListByProgressionAndExternalIDs(ctx, s.DB, progression.ProgressionID, progression.QuizExternalIDs)
	if err != nil {
		return nil, fmt.Errorf("LOProgressionAnswerRepo.ListByProgressionAndExternalIDs: %w", err)
	}

	mapProgressionAnswer := make(map[string]*entities.LOProgressionAnswer, 0)
	for _, pa := range progressionAnswers {
		mapProgressionAnswer[pa.QuizExternalID.String] = pa
	}

	// Get quizzes
	quizzes, err := s.QuizRepo.GetByExternalIDsAndLmID(ctx, s.DB, progression.QuizExternalIDs, progression.LearningMaterialID)
	if err != nil {
		return nil, fmt.Errorf("QuizRepo.GetByExternalIDs: %w", err)
	}

	{ // Shuffled quiz's options
		// Get Random Seed
		randomSeed, err := s.ShuffledQuizSetRepo.GetSeed(ctx, s.DB, progression.ShuffledQuizSetID)
		if err != nil {
			return nil, fmt.Errorf("ShuffledQuizSetRepo.GetSeed: %w", err)
		}

		seed, err := strconv.ParseInt(randomSeed.String, 10, 64)
		if err != nil {
			return nil, status.Errorf(codes.Internal, err.Error())
		}

		externalIDMap := make(map[string]int64)
		for i, externalID := range progression.QuizExternalIDs.Elements {
			externalIDMap[externalID.String] = int64(i)
		}

		zapLogger := ctxzap.Extract(ctx).Sugar()
		if err = quizzes.ShuffleOptions(seed, offset, externalIDMap); err != nil {
			zapLogger.Errorf("got ERROR: quizzes.ShuffleOptions: request: %v: error: %v\n", req, err)
		}
	}

	// Get question group
	// get list question group
	eqr, err := getQuestionGroupByQuiz(ctx, s.QuestionGroupRepo, s.DB, quizzes)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	questionGroups, err := entities.QuestionGroupsToQuestionGroupProtoBufMess(eqr)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	items := make([]*sspb.QuizAnswerInfo, 0)

	for _, quiz := range quizzes {
		item := &sspb.QuizAnswerInfo{}

		// Quiz
		if item.Quiz, err = toQuizPb(progression.LearningMaterialID.String, quiz); err != nil {
			return nil, status.Errorf(codes.Internal, err.Error())
		}

		// QuizAnswer
		if answer, ok := mapProgressionAnswer[quiz.ExternalID.String]; ok {
			if item.QuizAnswer, err = progressionAnswer2QuizAnswer(answer); err != nil {
				return nil, status.Errorf(codes.Internal, err.Error())
			}
		}

		items = append(items, item)
	}

	return &sspb.RetrieveLOProgressionResponse{
		OriginalShuffledQuizSetId: progression.ShuffledQuizSetID.String,
		LastIndex:                 uint32(progression.LastIndex.Int),
		SessionId:                 progression.SessionID.String,
		QuestionGroups:            questionGroups,
		Items:                     items,
		UpdatedAt:                 timestamppb.New(progression.UpdatedAt.Time),
		NextPage: &cpb.Paging{
			Limit: req.Paging.Limit,
			Offset: &cpb.Paging_OffsetInteger{
				OffsetInteger: req.Paging.GetOffsetInteger() + int64(req.Paging.Limit),
			},
		},
	}, nil
}

func progressionAnswer2QuizAnswer(e *entities.LOProgressionAnswer) (*sspb.QuizAnswer, error) {
	var (
		selectedIndex = make([]uint32, 0)
		filledText    = make([]string, 0)
		submittedKeys = make([]string, 0)
		answers       = make([]*sspb.Answer, 0)
	)

	if err := multierr.Combine(
		e.StudentIndexAnswers.AssignTo(&selectedIndex),
		e.StudentTextAnswers.AssignTo(&filledText),
		e.SubmittedKeysAnswer.AssignTo(&submittedKeys),
	); err != nil {
		return nil, err
	}

	for _, idx := range selectedIndex {
		answers = append(answers, &sspb.Answer{
			Format: &sspb.Answer_SelectedIndex{
				SelectedIndex: idx,
			},
		})
	}
	for _, text := range filledText {
		answers = append(answers, &sspb.Answer{
			Format: &sspb.Answer_FilledText{
				FilledText: text,
			},
		})
	}
	for _, sk := range submittedKeys {
		answers = append(answers, &sspb.Answer{
			Format: &sspb.Answer_SubmittedKey{
				SubmittedKey: sk,
			},
		})
	}

	return &sspb.QuizAnswer{
		QuizId: e.QuizExternalID.String,
		Answer: answers,
	}, nil
}
