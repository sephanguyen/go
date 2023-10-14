package services

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/eureka/services/question"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	"github.com/manabie-com/backend/internal/golibs/speeches"
	"github.com/manabie-com/backend/internal/golibs/timeutil"
	yasuo_entities "github.com/manabie-com/backend/internal/yasuo/entities"
	"github.com/manabie-com/backend/internal/yasuo/utils"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"
	ypb "github.com/manabie-com/backend/pkg/manabuf/yasuo/v1"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
	"golang.org/x/exp/slices"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type QuizService struct {
	sspb.UnimplementedQuizServer
	DB database.Ext

	QuizSetRepo interface {
		GetQuizSetByLoID(context.Context, database.QueryExecer, pgtype.Text) (*entities.QuizSet, error)
		RetrieveByLoIDs(ctx context.Context, db database.QueryExecer, loIDs pgtype.TextArray) (entities.QuizSets, error)
		GetQuizSetsContainQuiz(ctx context.Context, db database.QueryExecer, quizID pgtype.Text) (entities.QuizSets, error)
		BulkUpsert(ctx context.Context, db database.QueryExecer, data []*entities.QuizSet) ([]*entities.QuizSet, error)
	}

	ShuffledQuizSetRepo interface {
		Create(context.Context, database.QueryExecer, *entities.ShuffledQuizSet) (pgtype.Text, error)
		Get(context.Context, database.QueryExecer, pgtype.Text, pgtype.Int8, pgtype.Int8) (*entities.ShuffledQuizSet, error)
		Retrieve(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) ([]*entities.ShuffledQuizSet, error)
		GetExternalIDsFromSubmissionHistory(ctx context.Context, db database.QueryExecer, shuffleQuizSetID pgtype.Text, isAccepted bool) (pgtype.TextArray, error)
		GetByStudyPlanItemIdentities(ctx context.Context, db database.QueryExecer, identities []*repositories.StudyPlanItemIdentity) (entities.ShuffledQuizSets, error)
		ListExternalIDsFromSubmissionHistory(ctx context.Context, db database.QueryExecer, shuffleQuizSetIDs pgtype.TextArray, isAccepted bool) (map[string][]string, error)
		GetCorrectnessInfo(ctx context.Context, db database.QueryExecer, shuffleQuizSetID pgtype.Text, quizID pgtype.Text) (*entities.CorrectnessInfo, error)
		UpdateTotalCorrectnessAndSubmissionHistory(ctx context.Context, db database.QueryExecer, e *entities.ShuffledQuizSet) error
		UpsertLOSubmission(ctx context.Context, db database.QueryExecer, shuffleQuizSetID pgtype.Text) (*entities.LOSubmissionAnswerKey, error)
		UpsertFlashCardSubmission(ctx context.Context, db database.QueryExecer, shuffleQuizSetID pgtype.Text) (*entities.FlashCardSubmissionAnswerKey, error)
	}

	QuizRepo interface {
		GetByExternalIDsAndLOID(context.Context, database.QueryExecer, pgtype.TextArray, pgtype.Text) (entities.Quizzes, error)
		GetByExternalIDs(context.Context, database.QueryExecer, pgtype.TextArray, pgtype.Text) (entities.Quizzes, error)
		Upsert(ctx context.Context, db database.QueryExecer, data []*entities.Quiz) ([]*entities.Quiz, error)
		GetQuizByExternalID(ctx context.Context, db database.QueryExecer, externalID pgtype.Text) (*entities.Quiz, error)
		Search(ctx context.Context, db database.QueryExecer, filter repositories.QuizFilter) (entities.Quizzes, error)
		Create(ctx context.Context, db database.QueryExecer, quiz *entities.Quiz) error
		DeleteByExternalID(ctx context.Context, db database.QueryExecer, id pgtype.Text, schoolID pgtype.Int4) error
		GetByExternalID(ctx context.Context, db database.QueryExecer, id pgtype.Text, schoolID pgtype.Int4) (*entities.Quiz, error)
		GetByExternalIDsAndLmID(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray, lmID pgtype.Text) (entities.Quizzes, error)
	}

	QuestionGroupRepo interface {
		GetQuestionGroupsByIDs(ctx context.Context, db database.QueryExecer, ids ...string) (entities.QuestionGroups, error)
	}

	LearningTimeCalculatorSvc interface {
		Calculate([]*entities.StudentEventLog) (learningTime time.Duration, completedAt *time.Time, err error)
	}

	StudentEventLogRepo interface {
		RetrieveStudentEventLogsByStudyPlanIdentities(context.Context, database.QueryExecer, []*repositories.StudyPlanItemIdentity) ([]*entities.StudentEventLog, error)
	}

	SpeechesRepo interface {
		UpsertSpeeches(ctx context.Context, db database.QueryExecer, data []*yasuo_entities.Speeches) ([]*yasuo_entities.Speeches, error)
		RetrieveSpeeches(ctx context.Context, db database.QueryExecer, sentences pgtype.TextArray, settings pgtype.JSONBArray) ([]*yasuo_entities.Speeches, error)
		RetrieveAllSpeaches(ctx context.Context, db database.QueryExecer, limit, offset pgtype.Int8) ([]*yasuo_entities.Speeches, error)
	}

	StudentsLearningObjectivesCompletenessRepo interface {
		UpsertHighestQuizScore(ctx context.Context, db database.QueryExecer, loID pgtype.Text, studentID pgtype.Text, newScore pgtype.Float4) error
		UpsertFirstQuizCompleteness(ctx context.Context, db database.QueryExecer, loID pgtype.Text, studentID pgtype.Text, firstQuizScore pgtype.Float4) error
	}

	LOSubmissionAnswerRepo interface {
		Upsert(ctx context.Context, db database.QueryExecer, e *entities.LOSubmissionAnswer) error
		List(ctx context.Context, db database.QueryExecer, filter *repositories.LOSubmissionAnswerFilter) ([]*entities.LOSubmissionAnswer, error)
		BulkUpsert(ctx context.Context, db database.QueryExecer, items entities.LOSubmissionAnswers) error
	}

	FlashCardSubmissionAnswerRepo interface {
		Upsert(ctx context.Context, db database.QueryExecer, e *entities.FlashCardSubmissionAnswer) error
	}

	BobMediaModifier    BobMediaModifierServiceClient
	YasuoUploadReader   YasuoUploadReaderServiceClient
	YasuoUploadModifier YasuoUploadModifierServiceClient
}

type BobMediaModifierServiceClient interface {
	GenerateAudioFile(ctx context.Context, in *bpb.GenerateAudioFileRequest, opts ...grpc.CallOption) (*bpb.GenerateAudioFileResponse, error)
}

type YasuoUploadModifierServiceClient interface {
	BulkUploadHtmlContent(ctx context.Context, req *ypb.BulkUploadHtmlContentRequest, opts ...grpc.CallOption) (*ypb.BulkUploadHtmlContentResponse, error)
	UploadHtmlContent(ctx context.Context, in *ypb.UploadHtmlContentRequest, opts ...grpc.CallOption) (*ypb.UploadHtmlContentResponse, error)
}

type YasuoUploadReaderServiceClient interface {
	RetrieveUploadInfo(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*ypb.RetrieveUploadInfoResponse, error)
}

func NewQuizService(
	db database.Ext,
	bobMediaModifier BobMediaModifierServiceClient,
	yasuoUploadReader YasuoUploadReaderServiceClient,
	yasuoUploadModifier YasuoUploadModifierServiceClient,
) sspb.QuizServer {
	return &QuizService{
		DB:                        db,
		QuizSetRepo:               new(repositories.QuizSetRepo),
		QuizRepo:                  new(repositories.QuizRepo),
		ShuffledQuizSetRepo:       new(repositories.ShuffledQuizSetRepo),
		QuestionGroupRepo:         new(repositories.QuestionGroupRepo),
		StudentEventLogRepo:       new(repositories.StudentEventLogRepo),
		SpeechesRepo:              new(repositories.SpeechesRepository),
		BobMediaModifier:          bobMediaModifier,
		YasuoUploadModifier:       yasuoUploadModifier,
		YasuoUploadReader:         yasuoUploadReader,
		LearningTimeCalculatorSvc: &LearningTimeCalculator{},
		StudentsLearningObjectivesCompletenessRepo: new(repositories.StudentsLearningObjectivesCompletenessRepo),
		LOSubmissionAnswerRepo:                     new(repositories.LOSubmissionAnswerRepo),
		FlashCardSubmissionAnswerRepo:              new(repositories.FlashCardSubmissionAnswerRepo),
	}
}

func (s *QuizService) validateCreateQuizTestV2Request(req *sspb.CreateQuizTestV2Request) error {
	if req.StudyPlanItemIdentity.StudentId == nil {
		return fmt.Errorf("req must have student id")
	}
	if req.StudyPlanItemIdentity.LearningMaterialId == "" {
		return fmt.Errorf("req must have learning material id")
	}

	if req.StudyPlanItemIdentity.StudyPlanId == "" {
		return fmt.Errorf("req must have study plan id")
	}

	if req.Paging == nil {
		return fmt.Errorf("req must have paging field")
	}

	if req.Paging.GetOffsetInteger() <= 0 {
		return fmt.Errorf(("offset must be positive"))
	}

	if req.Paging.Limit <= 0 {
		return fmt.Errorf("limit must be positive")
	}
	return nil
}

func (s *QuizService) CreateQuizTestV2(ctx context.Context, req *sspb.CreateQuizTestV2Request) (*sspb.CreateQuizTestV2Response, error) {
	err := s.validateCreateQuizTestV2Request(req)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Get shuffledQuizset
	// if this is the firs time call, the req's SetID will be empty.
	// it means we need to create new shuffled quiz set, then response the new shuffled quiz set id for the later call
	shuffledQuizSet := &entities.ShuffledQuizSet{}
	database.AllNullEntity(shuffledQuizSet)
	if req.ShuffleQuizSetId.GetValue() == "" {
		// get quizset of the learning objective
		quizSet, err := s.QuizSetRepo.GetQuizSetByLoID(ctx, s.DB, database.Text(req.StudyPlanItemIdentity.LearningMaterialId))
		if err != nil {
			if err.Error() == pgx.ErrNoRows.Error() {
				return nil, status.Error(codes.NotFound, "Doesn't have any quiz set")
			}
			return nil, status.Errorf(codes.Internal, fmt.Errorf("CreateQuizTest.QuizSetRepo.GetQuizSetByLoID: %v", err).Error())
		}
		// instance new shuffledQuizSet
		seed := time.Now().UTC().UnixNano()
		now := timeutil.Now()
		sessionID := database.Text(req.SessionId)

		if req.SessionId == "" {
			sessionID.Set(nil)
		}
		if req.StudyPlanItemIdentity.StudentId != nil {
			shuffledQuizSet.StudentID.Set(req.StudyPlanItemIdentity.StudentId.GetValue())
		}
		err = multierr.Combine(
			shuffledQuizSet.ID.Set(idutil.ULIDNow()),
			shuffledQuizSet.LearningMaterialID.Set(req.StudyPlanItemIdentity.LearningMaterialId),
			shuffledQuizSet.StudyPlanID.Set(req.StudyPlanItemIdentity.StudyPlanId),
			shuffledQuizSet.OriginalQuizSetID.Set(quizSet.ID),
			shuffledQuizSet.QuizExternalIDs.Set(quizSet.QuizExternalIDs.Elements),
			shuffledQuizSet.Status.Set(quizSet.Status),
			shuffledQuizSet.RandomSeed.Set(strconv.FormatInt(seed, 10)),
			shuffledQuizSet.CreatedAt.Set(now),
			shuffledQuizSet.UpdatedAt.Set(now),
			shuffledQuizSet.DeletedAt.Set(nil),
			shuffledQuizSet.TotalCorrectness.Set(0),
			shuffledQuizSet.SubmissionHistory.Set(database.JSONB("[]")),
			shuffledQuizSet.SessionID.Set(sessionID),
			shuffledQuizSet.OriginalShuffleQuizSetID.Set(nil),
			shuffledQuizSet.QuestionHierarchy.Set(quizSet.QuestionHierarchy.Elements),
		)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, fmt.Errorf("unable to set value: %w", err).Error())
		}
		qms := &QuizModifierService{
			DB:       s.DB,
			QuizRepo: s.QuizRepo,
		}
		qms.generateShuffledQuizSetRandomSeed(ctx, shuffledQuizSet, req.StudyPlanItemIdentity.LearningMaterialId)
		if !req.KeepOrder {
			// shuffle quiz external ids
			shuffledQuizSet.ShuffleQuiz()
		}
		shuffledQuizSet.ID, err = s.ShuffledQuizSetRepo.Create(ctx, s.DB, shuffledQuizSet)
		if err != nil {
			return nil, status.Errorf(codes.Internal, fmt.Errorf("CreateQuizTest.ShuffledQuizSetRepo.Create: %v", err).Error())
		}
	} else {
		err := shuffledQuizSet.ID.Set(req.ShuffleQuizSetId.GetValue())
		if err != nil {
			return nil, status.Errorf(codes.Internal, err.Error())
		}
	}

	// Get our shuffledQuizset with paging
	offset := req.Paging.GetOffsetInteger()
	limit := req.Paging.Limit
	from := database.Int8(offset)
	to := database.Int8(offset + int64(limit) - 1)
	pagingShuffledQuizSet, err := s.ShuffledQuizSetRepo.Get(ctx, s.DB, shuffledQuizSet.ID, from, to)
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Errorf("CreateQuizTest.ShuffledQuizSetRepo.Get: %v", err).Error())
	}

	externalIDMap := make(map[string]int64)
	for i, externalID := range pagingShuffledQuizSet.QuizExternalIDs.Elements {
		externalIDMap[externalID.String] = int64(i)
	}

	// Get list of quiz item - entities.Quizzes
	quizzes, err := s.QuizRepo.GetByExternalIDsAndLmID(ctx, s.DB, pagingShuffledQuizSet.QuizExternalIDs, database.Text(req.StudyPlanItemIdentity.LearningMaterialId))
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Errorf("CreateQuizTest.QuizRepo.GetByExternalIDsAndLmID: %v", err).Error())
	}

	// Shuffled quiz's options
	seed, err := strconv.ParseInt(pagingShuffledQuizSet.RandomSeed.String, 10, 64)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	zapLogger := ctxzap.Extract(ctx).Sugar()
	if err = quizzes.ShuffleOptions(seed, offset, externalIDMap); err != nil {
		zapLogger.Errorf("got ERROR: quizzes.ShuffleOptions: request: %v: error: %v\n", req, err)
	}

	// Convert from entites.Quizzes to []bpb.Quiz
	pbQuizzes, err := toListQuizpb(req.StudyPlanItemIdentity.LearningMaterialId, quizzes)
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Errorf("CreateQuizTest.ToListQuizpb: %v", err).Error())
	}

	// get list question group
	qr, err := getQuestionGroupByQuiz(ctx, s.QuestionGroupRepo, s.DB, quizzes)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	respQuestionGroups, err := entities.QuestionGroupsToQuestionGroupProtoBufMess(qr)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	resp := &sspb.CreateQuizTestV2Response{
		NextPage: &cpb.Paging{
			Limit: req.Paging.Limit,
			Offset: &cpb.Paging_OffsetInteger{
				OffsetInteger: req.Paging.GetOffsetInteger() + int64(req.Paging.Limit),
			},
		},
		Quizzes:          pbQuizzes,
		ShuffleQuizSetId: shuffledQuizSet.ID.String,
		QuestionGroups:   respQuestionGroups,
	}

	return resp, nil
}

func (s *QuizService) CreateRetryQuizTestV2(ctx context.Context, req *sspb.CreateRetryQuizTestV2Request) (*sspb.CreateRetryQuizTestV2Response, error) {
	// currently we're not handle `keep_order` field
	if err := s.validateCreateRetryQuizTestV2(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// if this is the first time call, the req's RetryShuffleQuizId will be empty.
	// it means we need to create new shuffled quiz set, then response the new retry shuffled quiz set id for the later call
	shuffledQuizSetID := req.GetRetryShuffleQuizId().GetValue()
	if shuffledQuizSetID == "" {
		if shuffledQuizSet, err := s.createNewShuffledQuizSet(ctx, req); err != nil {
			return nil, err
		} else {
			shuffledQuizSetID = shuffledQuizSet.ID.String
		}
	}
	// Get our shuffledQuizset with paging
	offset := req.Paging.GetOffsetInteger()
	limit := req.Paging.Limit
	from := database.Int8(offset)
	to := database.Int8(offset + int64(limit) - 1)
	pagingShuffledQuizSet, err := s.ShuffledQuizSetRepo.Get(ctx, s.DB, database.Text(shuffledQuizSetID), from, to)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("CreateRetryQuizTest.ShuffledQuizSetRepo.Get: %v", err).Error())
	}

	externalIDMap := make(map[string]int64)
	for i, externalID := range pagingShuffledQuizSet.QuizExternalIDs.Elements {
		externalIDMap[externalID.String] = int64(i)
	}

	// Get list of quiz item - entities.Quizzes
	quizzes, err := s.QuizRepo.GetByExternalIDs(ctx, s.DB, pagingShuffledQuizSet.QuizExternalIDs, database.Text(req.StudyPlanItemIdentity.LearningMaterialId))
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("CreateRetryQuizTest.QuizRepo.GetByExternalIDs: %v", err).Error())
	}

	// Shuffled quiz's options
	seed, err := strconv.ParseInt(pagingShuffledQuizSet.RandomSeed.String, 10, 64)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("CreateRetryQuizTest.strconv.ParseInt: %v", err).Error())
	}

	zapLogger := ctxzap.Extract(ctx).Sugar()
	if err = quizzes.ShuffleOptions(seed, offset, externalIDMap); err != nil {
		zapLogger.Errorf("got ERROR: quizzes.ShuffleOptions: request: %v: error: %v\n", req, err)
	}

	// Convert from entites.Quizzes to []bpb.Quiz
	pbQuizzes, err := toListQuizpb(req.StudyPlanItemIdentity.LearningMaterialId, quizzes)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("CreateRetryQuizTest.ToListQuizpb: %v", err).Error())
	}

	// get list question group
	qr, err := getQuestionGroupByQuiz(ctx, s.QuestionGroupRepo, s.DB, quizzes)
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Errorf("getQuestionGroupByQuiz: %w", err).Error())
	}

	questionGroups, err := entities.QuestionGroupsToQuestionGroupProtoBufMess(qr)
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Errorf("QuestionGroupsToQuestionGroupProtoBufMess: %w", err).Error())
	}

	resp := &sspb.CreateRetryQuizTestV2Response{
		NextPage: &cpb.Paging{
			Limit: req.Paging.Limit,
			Offset: &cpb.Paging_OffsetInteger{
				OffsetInteger: req.Paging.GetOffsetInteger() + int64(req.Paging.Limit),
			},
		},
		Quizzes:          pbQuizzes,
		ShuffleQuizSetId: shuffledQuizSetID,
		QuestionGroups:   questionGroups,
	}

	return resp, nil
}

func (s *QuizService) validateCreateRetryQuizTestV2(req *sspb.CreateRetryQuizTestV2Request) error {
	if err := validStudyPlanItemIdentity(req.StudyPlanItemIdentity); err != nil {
		return err
	}

	if req.SessionId == "" {
		return fmt.Errorf("req must have session id")
	}
	if req.GetShuffleQuizSetId() == nil || req.GetShuffleQuizSetId().GetValue() == "" {
		return fmt.Errorf("req must have shuffle quizset id")
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

func (s *QuizService) createNewShuffledQuizSet(ctx context.Context, req *sspb.CreateRetryQuizTestV2Request) (*entities.ShuffledQuizSet, error) {
	shuffledQuizSet := &entities.ShuffledQuizSet{}
	database.AllNullEntity(shuffledQuizSet)
	// get quizset of the learning objective
	quizSet, err := s.QuizSetRepo.GetQuizSetByLoID(ctx, s.DB, database.Text(req.StudyPlanItemIdentity.LearningMaterialId))
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("CreateRetryQuizTest.QuizSetRepo.GetQuizSetByLoID: %w", err).Error())
	}
	shuffledQuizzes, err := s.ShuffledQuizSetRepo.Retrieve(ctx, s.DB, database.TextArray([]string{req.GetShuffleQuizSetId().GetValue()}))
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("CreateRetryQuizTest.ShuffledQuizSetRepo.Retrieve: %w", err).Error())
	}
	if len(shuffledQuizzes) == 0 {
		return nil, status.Error(codes.NotFound, fmt.Errorf("not found any shuffle quiz").Error())
	}
	correctQuizIDs, err := s.ShuffledQuizSetRepo.GetExternalIDsFromSubmissionHistory(ctx, s.DB, database.Text(req.GetShuffleQuizSetId().GetValue()), true)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("CreateRetryQuizTest.ShuffledQuizSetRepo.GetCorrectQuizIDs: %w", err).Error())
	}

	retryQuizIDs := getRetryQuizIDs(quizSet.QuizExternalIDs, correctQuizIDs)
	if len(retryQuizIDs) == 0 {
		return nil, status.Error(codes.FailedPrecondition, fmt.Errorf("CreateRetryQuizTest: all quizzes's answers accepted").Error())
	}

	currentQuestionHierarchy := entities.QuestionHierarchy{}
	if err := currentQuestionHierarchy.UnmarshalJSONBArray(quizSet.QuestionHierarchy); err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Errorf("CreateRetryQuizTestV2: unable to unmarshal: %w", err).Error())
	}

	retryQuestionHierarchy := currentQuestionHierarchy.ExcludeQuestionIDs(database.FromTextArray(correctQuizIDs))

	now := timeutil.Now()
	sessionID := database.Text(req.SessionId)

	err = multierr.Combine(
		shuffledQuizSet.ID.Set(idutil.ULIDNow()),
		shuffledQuizSet.StudentID.Set(req.StudyPlanItemIdentity.StudentId.GetValue()),
		shuffledQuizSet.LearningMaterialID.Set(req.StudyPlanItemIdentity.LearningMaterialId),
		shuffledQuizSet.StudyPlanID.Set(req.StudyPlanItemIdentity.StudyPlanId),
		shuffledQuizSet.OriginalQuizSetID.Set(quizSet.ID),
		shuffledQuizSet.QuizExternalIDs.Set(retryQuizIDs),
		shuffledQuizSet.Status.Set(shuffledQuizzes[0].Status),
		shuffledQuizSet.RandomSeed.Set(shuffledQuizzes[0].RandomSeed),
		shuffledQuizSet.CreatedAt.Set(now),
		shuffledQuizSet.UpdatedAt.Set(now),
		shuffledQuizSet.DeletedAt.Set(nil),
		shuffledQuizSet.TotalCorrectness.Set(shuffledQuizzes[0].TotalCorrectness),
		shuffledQuizSet.SubmissionHistory.Set(shuffledQuizzes[0].SubmissionHistory),
		shuffledQuizSet.SessionID.Set(sessionID),
		shuffledQuizSet.OriginalShuffleQuizSetID.Set(shuffledQuizzes[0].ID),
		shuffledQuizSet.QuestionHierarchy.Set(retryQuestionHierarchy),
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Errorf("CreateRetryQuizTest.Combine %w", err).Error())
	}
	qms := &QuizModifierService{
		DB:       s.DB,
		QuizRepo: s.QuizRepo,
	}
	qms.generateShuffledQuizSetRandomSeed(ctx, shuffledQuizSet, req.StudyPlanItemIdentity.LearningMaterialId)
	shuffledQuizSet.ID, err = s.ShuffledQuizSetRepo.Create(ctx, s.DB, shuffledQuizSet)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("CreateRetryQuizTest.ShuffledQuizSetRepo.Create: %v", err).Error())
	}
	return shuffledQuizSet, nil
}

// RetrieveQuizTestsV2 return list of quiz test info
func (s *QuizService) RetrieveQuizTestsV2(ctx context.Context, req *sspb.RetrieveQuizTestV2Request) (*sspb.RetrieveQuizTestV2Response, error) {
	if err := s.validateRetrieveQuizTestsV2(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// list shuffledQuizTests by study_plan_item_identities
	rIdentities := make([]*repositories.StudyPlanItemIdentity, len(req.StudyPlanItemIdentities))
	for i, v := range req.StudyPlanItemIdentities {
		rIdentities[i] = &repositories.StudyPlanItemIdentity{
			StudentID:          database.Text(v.StudentId.GetValue()),
			StudyPlanID:        database.Text(v.StudyPlanId),
			LearningMaterialID: database.Text(v.LearningMaterialId),
		}
	}
	quizTests, err := s.ShuffledQuizSetRepo.GetByStudyPlanItemIdentities(ctx, s.DB, rIdentities)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("s.ShuffledQuizSetRepo.GetByStudyPlanItemIdentities: %v", err).Error())
	}

	// list studentEventLogs by study_plan_item_identities
	// group logs by studyPlanItemID and SessionID
	studentEventLogsMap := make(map[repositories.StudyPlanItemIdentity]map[string][]*entities.StudentEventLog)
	for _, identity := range rIdentities {
		studentEventLogsMap[*identity] = make(map[string][]*entities.StudentEventLog)
	}

	studentEventLogs, err := s.retrieveStudentEventLogsConcurrently(ctx, rIdentities)
	if err != nil {
		return nil, err
	}

	for _, studentEventLog := range studentEventLogs {
		payload := make(map[string]interface{})
		if err := studentEventLog.Payload.AssignTo(&payload); err != nil {
			return nil, status.Error(codes.Internal, fmt.Errorf("log.Payload.AssignTo: %v", err).Error())
		}
		sessionID, ok := payload["session_id"].(string)
		if !ok {
			continue
		}

		key := repositories.StudyPlanItemIdentity{
			StudentID:          studentEventLog.StudentID,
			StudyPlanID:        studentEventLog.StudyPlanID,
			LearningMaterialID: studentEventLog.LearningMaterialID,
		}

		studentEventLogsMap[key][sessionID] = append(studentEventLogsMap[key][sessionID], studentEventLog)
	}

	// list map shuffledQuizSetID and externalQuizIDs -> externalIDsFromSubmissionHistoriesMap
	mapQuizTests := make(map[repositories.StudyPlanItemIdentity][]*entities.ShuffledQuizSet)
	var originalShuffledQuizSetIDs []string
	originalShuffledQuizSetIDMap := make(map[string]struct{})
	for _, quizTest := range quizTests {
		key := repositories.StudyPlanItemIdentity{
			StudentID:          quizTest.StudentID,
			StudyPlanID:        quizTest.StudyPlanID,
			LearningMaterialID: quizTest.LearningMaterialID,
		}
		mapQuizTests[key] = append(mapQuizTests[key], quizTest)

		if quizTest.OriginalShuffleQuizSetID.Status != pgtype.Present {
			continue
		}

		if _, ok := originalShuffledQuizSetIDMap[quizTest.OriginalShuffleQuizSetID.String]; !ok {
			originalShuffledQuizSetIDMap[quizTest.OriginalShuffleQuizSetID.String] = struct{}{}
			originalShuffledQuizSetIDs = append(originalShuffledQuizSetIDs, quizTest.OriginalShuffleQuizSetID.String)
		}
	}

	externalIDsFromSubmissionHistoriesMap := make(map[string][]string)
	if len(originalShuffledQuizSetIDs) != 0 {
		externalIDsFromSubmissionHistoriesMap, err = s.ShuffledQuizSetRepo.ListExternalIDsFromSubmissionHistory(ctx, s.DB, database.TextArray(originalShuffledQuizSetIDs), false)
		if err != nil {
			return nil, status.Error(codes.Internal, fmt.Errorf("RetrieveQuizTestsV2.ShuffledQuizSetRepo.ListExternalIDsFromSubmissionHistory: %w", err).Error())
		}
	}

	// calculate learning time, score and crown
	items := make([]*sspb.RetrieveQuizTestV2ResponseItem, 0)
	var totalAttempt int32
	highestScore := &cpb.HighestQuizScore{}

	for identity, tests := range mapQuizTests {
		testInfos := []*cpb.QuizTestInfo{}
		for _, test := range tests {
			logs := studentEventLogsMap[identity][test.SessionID.String]
			learningTime, completedAt, err := s.LearningTimeCalculatorSvc.Calculate(logs)
			if err != nil {
				return nil, status.Error(codes.Internal, fmt.Errorf("RetrieveQuizTestsV2.LearningTimeCalculatorSvc.Calculate: %v", err).Error())
			}
			// omit if not suitable with request's condition
			if req.GetIsCompleted() && completedAt == nil {
				continue
			}

			createdAt := timestamppb.New(test.CreatedAt.Time)
			var completedAtPb *timestamppb.Timestamp
			if completedAt != nil {
				completedAtPb = timestamppb.New(*completedAt)
			}
			var isRetry bool
			var totalQuiz int32
			totalQuiz = int32(len(test.QuizExternalIDs.Elements))
			if test.OriginalShuffleQuizSetID.Status == pgtype.Present {
				// only the origin attempt, not retry mode, because the default auto plus in below, so we have to sub
				totalAttempt--
				isRetry = true

				// because we only save the incorrect external_ids + new (external_ids which recently add from admin) to field quiz_external_ids
				// so when we use retry mode, it will handle missing the external_ids which done before,
				// we have to get all with distinct external_ids
				externalQuizIDs := externalIDsFromSubmissionHistoriesMap[test.OriginalShuffleQuizSetID.String]

				for _, e := range test.QuizExternalIDs.Elements {
					if e.Status == pgtype.Present {
						externalQuizIDs = append(externalQuizIDs, e.String)
					}
				}
				totalQuiz = int32(len(golibs.GetUniqueElementStringArray(externalQuizIDs)))
			}

			infor := &cpb.QuizTestInfo{
				SetId:             test.ID.String,
				TotalCorrectness:  test.TotalCorrectness.Int,
				TotalQuiz:         totalQuiz, // int32(len(test.QuizExternalIDs.Elements)),
				CreatedAt:         createdAt,
				TotalLearningTime: int64(learningTime.Seconds()),
				CompletedAt:       completedAtPb,
				IsRetry:           isRetry,
			}
			totalAttempt++
			getMaxQuizScore(highestScore, infor.GetTotalCorrectness(), infor.GetTotalQuiz())
			testInfos = append(testInfos, infor)
		}

		// sort the item decs completed_at
		sortQuizTestInfo(testInfos)

		items = append(items, &sspb.RetrieveQuizTestV2ResponseItem{
			StudyPlanItemIdentity: &sspb.StudyPlanItemIdentity{
				StudentId:          wrapperspb.String(identity.StudentID.String),
				StudyPlanId:        identity.StudyPlanID.String,
				LearningMaterialId: identity.LearningMaterialID.String,
			},
			QuizTest: &cpb.QuizTests{
				Items: testInfos,
			},
		})
	}

	resp := &sspb.RetrieveQuizTestV2Response{
		Items:         items,
		HighestCrown:  sspb.AchievementCrown(getCrown(highestScore.GetCorrectQuestion(), highestScore.GetTotalQuestion())),
		HighestScore:  highestScore,
		TotalAttempts: totalAttempt,
	}

	return resp, nil
}

// fetch the logs chunk by chunk to avoid Seq Scan
// because of too large amount of records will be over the effective_cache_size
// Postgres will choose Seq Scan rather than Index Scan
func (s *QuizService) retrieveStudentEventLogsConcurrently(ctx context.Context, studyPlanItemIdentities []*repositories.StudyPlanItemIdentity) ([]*entities.StudentEventLog, error) {
	return retrieveStudentEventLogsConcurrentlyByStudyPlanItemIdentities(ctx, s.DB, studyPlanItemIdentities, s.StudentEventLogRepo)
}

func (s *QuizService) validateRetrieveQuizTestsV2(req *sspb.RetrieveQuizTestV2Request) error {
	if len(req.StudyPlanItemIdentities) == 0 {
		return fmt.Errorf("req must have Study Plan Item Identity")
	}

	for i, identity := range req.StudyPlanItemIdentities {
		if err := validStudyPlanItemIdentity(identity); err != nil {
			return fmt.Errorf("StudyPlanItemIdentities[%d] is error %w", i, err)
		}
	}

	return nil
}

func validStudyPlanItemIdentity(identity *sspb.StudyPlanItemIdentity) error {
	if identity.StudentId == nil {
		return fmt.Errorf("req must have student id")
	}
	if identity.LearningMaterialId == "" {
		return fmt.Errorf("req must have learning material id")
	}
	if identity.StudyPlanId == "" {
		return fmt.Errorf("req must have study plan id")
	}

	return nil
}

type QuestionGroupRepo interface {
	GetQuestionGroupsByIDs(ctx context.Context, db database.QueryExecer, ids ...string) (entities.QuestionGroups, error)
}

func getQuestionGroupByQuiz(ctx context.Context, repo QuestionGroupRepo, db database.Ext, quizzes entities.Quizzes) (entities.QuestionGroups, error) {
	questionGrIDs := make([]string, 0)
	for _, quiz := range quizzes {
		if quiz.QuestionGroupID.Status == pgtype.Present && len(quiz.QuestionGroupID.String) != 0 {
			if !sliceutils.Contains(questionGrIDs, quiz.QuestionGroupID.String) {
				questionGrIDs = append(questionGrIDs, quiz.QuestionGroupID.String)
			}
		}
	}
	qr, err := repo.GetQuestionGroupsByIDs(ctx, db, questionGrIDs...)
	if err != nil {
		return nil, fmt.Errorf("QuestionGroupRepo.GetQuestionGroupsByIDs: %v", err)
	}

	return qr, nil
}

func validateUpsertFlashcardContent(req *cpb.QuizCore) error {
	if req == nil {
		return status.Error(codes.InvalidArgument, "missing Quiz")
	}

	if req.ExternalId == "" {
		return status.Error(codes.InvalidArgument, "missing ExternalId")
	}

	if req.Question == nil {
		return status.Error(codes.InvalidArgument, "missing Question")
	}

	if req.Explanation == nil {
		return status.Error(codes.InvalidArgument, "missing Explanation")
	}

	if len(req.Options) == 0 {
		return status.Error(codes.InvalidArgument, "missing Options")
	}
	return nil
}
func (s *QuizService) quizReqPbToEnt(req *cpb.QuizCore, q *entities.Quiz, info *ypb.RetrieveUploadInfoResponse, flashcardID string) (contentChange map[string]string, err error) {
	endpoint := info.GetEndpoint()
	bucket := info.GetBucket()
	contentChange = make(map[string]string)
	q.ExternalID = database.Text(req.ExternalId)
	q.Country = database.Text(req.Info.Country.String())
	q.SchoolID = database.Int4(req.Info.SchoolId)
	if q.LoIDs.Status != pgtype.Present {
		q.LoIDs = database.TextArray([]string{flashcardID})
	}
	q.Kind = database.Text(req.Kind.String())
	q.TaggedLOs = database.TextArray(req.TaggedLos)
	q.DifficultLevel = database.Int4(req.DifficultyLevel)
	if req.Point != nil {
		q.Point = database.Int4(req.Point.Value)
	} else {
		q.Point = database.Int4(1) // default value
	}
	q.QuestionTagIds = database.TextArray(req.QuestionTagIds)
	q.Status = database.Text(epb.QuizStatus_QUIZ_STATUS_APPROVED.String())
	q.CreatedAt = database.Timestamptz(time.Now())
	q.UpdatedAt = database.Timestamptz(time.Now())

	url, _ := generateUploadURL(endpoint, bucket, req.Question.Rendered)
	question, err := q.GetQuestion()
	if err != nil {
		return nil, fmt.Errorf("err GetQuestion: %w", err)
	}

	configs := make([]string, 0)
	for _, each := range req.Attribute.Configs {
		configs = append(configs, each.String())
	}

	if question.RenderedURL != url {
		q.Question = database.JSONB(&entities.QuizQuestion{
			Raw:         req.Question.Raw,
			RenderedURL: url,
			Attribute: entities.QuizItemAttribute{
				ImgLink: req.Attribute.ImgLink,
				Configs: configs,
			},
		})

		contentChange[url] = req.Question.Rendered
	}

	url, _ = generateUploadURL(endpoint, bucket, req.Explanation.Rendered)

	e, err := q.GetExplaination()
	if err != nil {
		return nil, fmt.Errorf("err GetExplaination: %w", err)
	}

	if e.RenderedURL != url {
		q.Explanation = database.JSONB(&entities.RichText{
			Raw:         req.Explanation.Raw,
			RenderedURL: url,
		})

		contentChange[url] = req.Explanation.Rendered
	}

	originOption, _ := q.GetOptions()
	quizOptions := make([]*entities.QuizOption, 0, len(req.Options))
	for i, o := range req.Options {
		url, _ = generateUploadURL(endpoint, bucket, o.Content.Rendered)
		if err != nil {
			return nil, fmt.Errorf("err UploadHtmlContent Option: %w", err)
		}

		content := entities.RichText{
			Raw:         o.Content.Raw,
			RenderedURL: url,
		}

		isChange := true
		if i < len(originOption) {
			opt := originOption[i]
			isChange = opt != nil && opt.Content.RenderedURL != url
		}

		if isChange {
			contentChange[url] = o.Content.Rendered
		}

		configs = []string{}
		for _, each := range o.Attribute.Configs {
			configs = append(configs, each.String())
		}

		quizOptions = append(quizOptions, &entities.QuizOption{
			Content:     content,
			Correctness: o.Correctness,
			Configs:     quizCfg2ArrayString(o.Configs),
			Label:       o.Label,
			Key:         o.Key,
			Attribute: entities.QuizItemAttribute{
				Configs: configs,
			},
		})
	}

	q.Options = database.JSONB(quizOptions)
	return contentChange, nil
}

// AssignLosToQuiz for old version quizzes which dont have column lo ids, we need to migrate assign lo ids to old quizzes
func (s *QuizService) AssignLosToQuiz(ctx context.Context, quiz *entities.Quiz) error {
	quizSets, err := s.QuizSetRepo.GetQuizSetsContainQuiz(ctx, s.DB, quiz.ExternalID)
	if err != nil {
		return err
	}
	loID := make([]string, 0, len(quizSets))
	duplicateLOMap := make(map[string]bool)
	for _, set := range quizSets {
		if _, ok := duplicateLOMap[set.LoID.String]; !ok {
			loID = append(loID, set.LoID.String)
		}
		duplicateLOMap[set.LoID.String] = true
	}

	err = quiz.LoIDs.Set(loID)
	if err != nil {
		return err
	}
	return nil
}

// nolint
func (s *QuizService) generateAudioFilesHandler(ctx context.Context, tx pgx.Tx, createdQuizzes []*entities.Quiz) error {
	userID := interceptors.UserIDFromContext(ctx)

	mdctx, err := interceptors.GetOutgoingContext(ctx)
	if err != nil {
		return fmt.Errorf("QuizModifierService.generateAudioFilesHandler.GetOutgoingContext: %w", err)
	}

	speechesReq := new(bpb.GenerateAudioFileRequest)
	var sentences []string
	var configs []interface{}
	for _, quiz := range createdQuizzes {
		q, err := quiz.GetQuestionV2()
		if err != nil {
			return err
		}
		if lang := speeches.GetLanguage(q.Attribute.Configs); lang != "" && q.GetText() != "" {

			if slices.Contains(speeches.WhiteList, lang) {
				sentences = append(sentences, q.GetText())
				configs = append(configs, &repositories.SpeechConfig{
					Language: lang,
				})
			} else {
				q.Attribute.AudioLink = ""
				quiz.Question = database.JSONB(q)
			}
		}
		o, err := quiz.GetOptions()
		if err != nil {
			return err
		}
		for _, each := range o {
			if lang := speeches.GetLanguage(each.Attribute.Configs); lang != "" && each.GetText() != "" {
				if slices.Contains(speeches.WhiteList, lang) {
					sentences = append(sentences, each.GetText())
					configs = append(configs, &repositories.SpeechConfig{
						Language: lang,
					})
				} else {
					each.Attribute.AudioLink = ""
				}
			}
		}
		quiz.Options = database.JSONB(o)
	}
	if len(sentences) > 0 {
		spces, err := s.SpeechesRepo.RetrieveSpeeches(ctx, tx, database.TextArray(sentences), database.JSONBArray(configs))
		if err != nil {
			return fmt.Errorf("SpeechesRepo.RetrieveSpeeches: %w", err)
		}

		type info struct {
			Language string `json:"language,omitempty"`
			Text     string `json:"text,omitempty"`
		}
		speechMap := make(map[info]*yasuo_entities.Speeches)
		for _, speech := range spces {
			in := info{
				Text: speech.Sentence.String,
			}
			speech.Settings.AssignTo(&in)
			speechMap[in] = speech
		}
		for _, quiz := range createdQuizzes {
			q, err := quiz.GetQuestionV2()
			if err != nil {
				return err
			}
			lang := speeches.GetLanguage(q.Attribute.Configs)
			if speech, ok := speechMap[info{
				Language: lang,
				Text:     q.GetText(),
			}]; !ok {
				if slices.Contains(speeches.WhiteList, lang) {
					speechesReq.Options = append(speechesReq.Options, &bpb.AudioOptionRequest{
						Text: q.GetText(),
						Configs: &bpb.AudioConfig{
							Language: lang,
						},
						QuizId: quiz.ID.String,
						Type:   bpb.AudioOptionType_FLASHCARD_AUDIO_TYPE_TERM,
					})
				} else {
					q.Attribute.AudioLink = ""
				}

			} else {
				q.Attribute.AudioLink = speech.Link.String
				quiz.Question = database.JSONB(q)
			}
			o, err := quiz.GetOptions()
			if err != nil {
				return err
			}
			for _, each := range o {
				lang := speeches.GetLanguage(each.Attribute.Configs)
				if speech, ok := speechMap[info{
					Language: lang,
					Text:     each.GetText(),
				}]; !ok {
					if slices.Contains(speeches.WhiteList, lang) {
						speechesReq.Options = append(speechesReq.Options, &bpb.AudioOptionRequest{
							Text: each.GetText(),
							Configs: &bpb.AudioConfig{
								Language: lang,
							},
							QuizId: quiz.ID.String,
							Type:   bpb.AudioOptionType_FLASHCARD_AUDIO_TYPE_DEFINITION,
						})
					} else {
						each.Attribute.AudioLink = ""
					}
				} else {
					each.Attribute.AudioLink = speech.Link.String
				}
			}
			if err := quiz.Options.Set(o); err != nil {
				return err
			}
		}
	}

	if len(speechesReq.Options) > 0 {
		resp, err := s.BobMediaModifier.GenerateAudioFile(mdctx, speechesReq)
		if err != nil {
			return fmt.Errorf("MediaModifierService.GenerateAudioFile: %w", err)
		}

		upsertSpeechesReq := make([]*yasuo_entities.Speeches, 0)
		for _, each := range createdQuizzes {
			for _, option := range resp.Options {
				if each.ID.String == option.QuizId {
					entity := new(yasuo_entities.Speeches)
					database.AllNullEntity(entity)
					if err := multierr.Combine(
						entity.SpeechID.Set(idutil.ULIDNow()),
						entity.Sentence.Set(option.Text),
						entity.Link.Set(option.Link),
						entity.Settings.Set(&repositories.SpeechConfig{
							Language: option.Configs.Language,
						}),
						entity.CreatedBy.Set(userID),
						entity.UpdatedBy.Set(userID),
						entity.Type.Set(option.Type.String()),
						entity.QuizID.Set(option.QuizId),
					); err != nil {
						return err
					}

					upsertSpeechesReq = append(upsertSpeechesReq, entity)
				}
			}
		}

		createdSpeeches, err := s.SpeechesRepo.UpsertSpeeches(ctx, tx, upsertSpeechesReq)
		if err != nil {
			return fmt.Errorf("SpeechesRepo.UpsertSpeeches: %w", err)
		}

		for _, each := range createdQuizzes {
			r, err := each.GetQuestionV2()
			if err != nil {
				return err
			}

			o, err := each.GetOptions()
			if err != nil {
				return err
			}

			changeOptionIdx := []int{}
			for i, option := range o {
				if !slices.Contains(option.Attribute.Configs, cpb.QuizItemAttributeConfig_FLASHCARD_LANGUAGE_CONFIG_NONE.String()) {
					changeOptionIdx = append(changeOptionIdx, i)
				}
			}

			for _, speech := range createdSpeeches {
				if each.ID == speech.QuizID {
					switch speech.Type.String {
					case bpb.AudioOptionType_FLASHCARD_AUDIO_TYPE_TERM.String():
						r.Attribute = entities.QuizItemAttribute{
							AudioLink: speech.Link.String,
							ImgLink:   r.Attribute.ImgLink,
							Configs:   r.Attribute.Configs,
						}
						if err := each.Question.Set(r); err != nil {
							return fmt.Errorf("Question.Set %w", err)
						}
					case bpb.AudioOptionType_FLASHCARD_AUDIO_TYPE_DEFINITION.String():
						for z, idx := range changeOptionIdx {
							if o[idx].Content.GetText() == speech.Sentence.String {
								o[idx].Attribute = entities.QuizItemAttribute{
									AudioLink: speech.Link.String,
									Configs:   o[idx].Attribute.Configs,
								}
								changeOptionIdx = slices.Delete(changeOptionIdx, z, z+1)
							}
						}
					}
				}
			}
			if err := each.Options.Set(o); err != nil {
				return err
			}
		}
	}

	if _, err := s.QuizRepo.Upsert(ctx, tx, createdQuizzes); err != nil {
		return fmt.Errorf("QuizRepo.Upsert: %w", err)
	}

	return nil
}

// nolint
func (s *QuizService) UpsertFlashcardContent(ctx context.Context, req *sspb.UpsertFlashcardContentRequest) (*sspb.UpsertFlashcardContentResponse, error) {
	type QuizEntityInfo struct {
		ContentChange map[string]string
		QuizEntity    *entities.Quiz
		Quiz          *cpb.QuizCore
	}

	var (
		uresp   *ypb.RetrieveUploadInfoResponse
		quizzes entities.Quizzes
		sets    entities.QuizSets
	)
	quizEntityInfos := make([]*QuizEntityInfo, 0)

	marker := make(map[string]bool)
	userID := interceptors.UserIDFromContext(ctx)

	mdCtx, err := interceptors.GetOutgoingContext(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, fmt.Errorf("interceptors.GetOutgoingContext: %w", err).Error())
	}

	externalIDs := make([]string, 0, len(req.Quizzes))
	for _, q := range req.Quizzes {
		if err := validateUpsertFlashcardContent(q); err != nil {
			return nil, err
		}

		externalIDs = append(externalIDs, q.ExternalId)
	}
	if req.GetFlashcardId() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "lo_id can not accept empty")
	}
	eg, _ := errgroup.WithContext(ctx)
	eg.Go(func() error {
		uresp, err = s.YasuoUploadReader.RetrieveUploadInfo(mdCtx, &emptypb.Empty{})
		return err
	})

	eg.Go(func() error {
		quizzes, err = s.QuizRepo.GetByExternalIDsAndLmID(ctx, s.DB, database.TextArray(externalIDs), database.Text(req.GetFlashcardId()))
		if err != nil {
			return status.Error(codes.Internal, err.Error())
		}
		return nil
	})

	eg.Go(func() error {
		sets, err = s.QuizSetRepo.RetrieveByLoIDs(ctx, s.DB, database.TextArray([]string{req.GetFlashcardId()}))
		if err != nil {
			return status.Errorf(codes.Internal, fmt.Errorf("QuizSetRepo.RetrieveByLoIDs: %w", err).Error())
		}
		return nil
	})

	if err := eg.Wait(); err != nil {
		return nil, err
	}
	quizMap := make(map[string]*entities.Quiz)
	for _, quiz := range quizzes {
		quizMap[quiz.ExternalID.String] = quiz
	}
	for _, q := range req.Quizzes {
		externalID := q.ExternalId
		quiz, ok := quizMap[externalID]
		if !ok {
			quiz = &entities.Quiz{}
			database.AllNullEntity(quiz)
			quiz.ID.Set(idutil.ULIDNow())
		} else {
			if len(quiz.LoIDs.Elements) == 0 {
				// migrate from old version quizzes which dont have column lo_ids
				// assign lo_ids to this old version quizzes
				if err := s.AssignLosToQuiz(ctx, quiz); err != nil {
					return nil, status.Error(codes.Internal, "can not assign los to Quiz")
				}
			}
			if !slices.Contains(database.FromTextArray(quiz.LoIDs), req.GetFlashcardId()) {
				// not update case
				// this case is other Lo create quiz which is already existed
				return nil, status.Error(codes.Internal, fmt.Sprintf("err quiz is already existed in LOs %v", quiz.LoIDs))
			}
		}

		contentChange, err := s.quizReqPbToEnt(q, quiz, uresp, req.GetFlashcardId())
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		quiz.CreatedBy.Set(database.Text(userID))
		quizEntityInfos = append(quizEntityInfos, &QuizEntityInfo{
			ContentChange: contentChange,
			QuizEntity:    quiz,
			Quiz:          q,
		})
	}

	var quizsets []*entities.QuizSet

	var preSet *entities.QuizSet
	var quizExternalIDs []string
	if len(sets) > 0 {
		preSet = sets[len(sets)-1]
		quizExternalIDs = database.FromTextArray(preSet.QuizExternalIDs)
	}
	var contents []string
	var urls []string
	for _, info := range quizEntityInfos {
		//upload content change
		for k, c := range info.ContentChange {
			url := k
			content := c
			contents = append(contents, content)
			urls = append(urls, url)
		}

		if err := multierr.Combine(
			info.QuizEntity.Status.Set(epb.QuizStatus_QUIZ_STATUS_APPROVED.String()),
			info.QuizEntity.ApprovedBy.Set(userID),
		); err != nil {
			return nil, status.Errorf(codes.InvalidArgument, err.Error())
		}

		quizzes = append(quizzes, info.QuizEntity)

		if preSet != nil && slices.Contains(database.FromTextArray(preSet.QuizExternalIDs), info.Quiz.ExternalId) {
			continue
		}

		if preSet != nil {
			preSet.DeletedAt.Set(time.Now())
			if ok := marker[preSet.ID.String]; !ok {
				quizsets = append(quizsets, preSet)
				marker[preSet.ID.String] = true
			}
		}

		quizExternalIDs = append(quizExternalIDs, info.Quiz.ExternalId)

		set := &entities.QuizSet{}

		database.AllNullEntity(set)
		now := time.Now()
		set.ID.Set(idutil.ULIDNow())
		set.LoID.Set(req.GetFlashcardId())
		set.Status.Set(epb.QuizSetStatus_QUIZSET_STATUS_APPROVED.String())
		set.CreatedAt.Set(now)
		set.UpdatedAt.Set(now)
		set.QuizExternalIDs.Set(quizExternalIDs)

		quizsets = append(quizsets, set)
		marker[set.ID.String] = true
		preSet = set
	}

	contentResp, err := s.YasuoUploadModifier.BulkUploadHtmlContent(mdCtx, &ypb.BulkUploadHtmlContentRequest{
		Contents: contents,
	})
	if err != nil {
		return nil, err
	}
	if !sliceutils.UnorderedEqual(urls, contentResp.GetUrls()) {
		return nil, status.Errorf(codes.Internal, "url return does not match, expect %v but got %v", urls, contentResp.GetUrls())
	}
	if err := database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {

		if len(quizsets) > 0 {
			if _, err := s.QuizSetRepo.BulkUpsert(ctx, tx, quizsets); err != nil {
				return fmt.Errorf("QuizSetRepo.BulkUpsert: %w", err)
			}
		}
		createdQuizzes, err := s.QuizRepo.Upsert(ctx, tx, quizzes)
		if err != nil {
			return fmt.Errorf("QuizRepo.Upsert: %w", err)
		}

		if req.Kind == cpb.QuizType_QUIZ_TYPE_POW {
			if err := s.generateAudioFilesHandler(ctx, tx, createdQuizzes); err != nil {
				return fmt.Errorf("s.generateAudioFilesHandler: %w", err)
			}
		}

		return nil
	}); err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	return &sspb.UpsertFlashcardContentResponse{}, nil
}

func (s *QuizService) CheckQuizCorrectness(ctx context.Context, req *sspb.CheckQuizCorrectnessRequest) (*sspb.CheckQuizCorrectnessResponse, error) {
	if err := validateCheckQuizCorrectnessRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	quiz, err := s.QuizRepo.GetQuizByExternalID(ctx, s.DB, database.Text(req.QuizId))
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("QuizRepo.GetByExternalQuizID: %w", err).Error())
	}
	correctnessInfo, err := s.ShuffledQuizSetRepo.GetCorrectnessInfo(ctx, s.DB, database.Text(req.ShuffledQuizSetId), database.Text(req.QuizId))
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("ShuffledQuizSetRepo.GetCorrectnessInfo: %w", err).Error())
	}

	correctnessQuiz := &CorrectnessQuiz{
		Quiz:            quiz,
		CorrectnessInfo: correctnessInfo,
		Answers:         req.Answer,
	}
	answer, err := correctnessQuiz.Check()
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("CorrectnessQuiz.Check: %w", err).Error())
	}

	err = database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		switch req.LmType {
		case sspb.LearningMaterialType_LEARNING_MATERIAL_LEARNING_OBJECTIVE:
			loKey, err := s.ShuffledQuizSetRepo.UpsertLOSubmission(ctx, tx, database.Text(req.ShuffledQuizSetId))
			if err != nil {
				return fmt.Errorf("ShuffledQuizSetRepo.UpsertLOSubmission: %w", err)
			}

			loAnswer, err := toLOSubmissionAnswerEntityByLoKey(loKey, answer)
			if err != nil {
				return fmt.Errorf("toLOSubmissionAnswerEntityByLoKey: %w", err)
			}

			if correctnessInfo.OriginalShuffleQuizSetID.Status == pgtype.Null {
				if err = s.LOSubmissionAnswerRepo.Upsert(ctx, tx, loAnswer); err != nil {
					return fmt.Errorf("LOSubmissionAnswerRepo.Upsert: %w", err)
				}
			} else { // Retry case
				loAnswers, err := s.LOSubmissionAnswerRepo.List(ctx, tx, &repositories.LOSubmissionAnswerFilter{
					SubmissionID:      pgtype.Text{Status: pgtype.Null},
					ShuffledQuizSetID: database.Text(correctnessInfo.OriginalShuffleQuizSetID.String),
				})
				if err != nil {
					return fmt.Errorf("LOSubmissionAnswerRepo.List: %w", err)
				}

				for i := 0; i < len(loAnswers); i++ {
					if loAnswers[i].QuizID.String == loAnswer.QuizID.String {
						loAnswers[i] = loAnswer // replace old answer by QuizID
					}
				}

				if err = s.LOSubmissionAnswerRepo.BulkUpsert(ctx, tx, loAnswers); err != nil {
					return fmt.Errorf("LOSubmissionAnswerRepo.BulkUpsert: %w", err)
				}
			}
		case sspb.LearningMaterialType_LEARNING_MATERIAL_FLASH_CARD:
			fcKey, err := s.ShuffledQuizSetRepo.UpsertFlashCardSubmission(ctx, tx, database.Text(req.ShuffledQuizSetId))
			if err != nil {
				return fmt.Errorf("ShuffledQuizSetRepo.UpsertFlashCardSubmission: %w", err)
			}

			fcAnswer, err := toFlashCardSubmissionAnswerEntityByFcKey(fcKey, answer)
			if err != nil {
				return fmt.Errorf("toFlashCardSubmissionAnswerEntityByFcKey: %w", err)
			}

			if err = s.FlashCardSubmissionAnswerRepo.Upsert(ctx, tx, fcAnswer); err != nil {
				return fmt.Errorf("FlashCardSubmissionAnswerRepo.Upsert: %w", err)
			}
		}

		totalCorrectness := correctnessInfo.TotalCorrectness.Int
		if answer.IsAccepted {
			totalCorrectness++
		}
		shuffledQuizSet := &entities.ShuffledQuizSet{}
		database.AllNullEntity(shuffledQuizSet)
		if err = multierr.Combine(
			shuffledQuizSet.ID.Set(req.ShuffledQuizSetId),
			shuffledQuizSet.TotalCorrectness.Set(totalCorrectness),
			shuffledQuizSet.SubmissionHistory.Set(answer),
			shuffledQuizSet.UpdatedAt.Set(time.Now()),
		); err != nil {
			return fmt.Errorf("unable to setup shuffledQuizSet: %w", err)
		}
		if err = s.ShuffledQuizSetRepo.UpdateTotalCorrectnessAndSubmissionHistory(ctx, tx, shuffledQuizSet); err != nil {
			return fmt.Errorf("ShuffledQuizSetRepo.UpdateTotalCorrectnessAndSubmissionHistory: %w", err)
		}

		isRetry := correctnessInfo.OriginalShuffleQuizSetID.Status == pgtype.Present
		isFinished := correctnessInfo.TotalSubmissionHistory.Int == correctnessInfo.TotalQuizExternalIDs.Int

		if isRetry || isFinished {
			score := float32(math.Floor(float64(correctnessInfo.TotalCorrectness.Int) / float64(correctnessInfo.TotalSubmissionHistory.Int) * 100))
			if err = s.StudentsLearningObjectivesCompletenessRepo.UpsertFirstQuizCompleteness(ctx, tx, correctnessInfo.LoID, correctnessInfo.StudentID, database.Float4(score)); err != nil {
				return fmt.Errorf("CheckQuizCorrectness.StudentsLearningObjectivesCompletenessRepo.UpsertFirstQuizCompleteness: %w", err)
			}

			if err = s.StudentsLearningObjectivesCompletenessRepo.UpsertHighestQuizScore(ctx, tx, correctnessInfo.LoID, correctnessInfo.StudentID, database.Float4(score)); err != nil {
				return fmt.Errorf("CheckQuizCorrectness.StudentsLearningObjectivesCompletenessRepo.UpsertHighestQuizScore: %w", err)
			}
		}

		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &sspb.CheckQuizCorrectnessResponse{
		Correctness:  answer.Correctness,
		IsCorrectAll: answer.IsAllCorrect,
		Result: &sspb.CheckQuizCorrectnessResponse_OrderingResult{
			OrderingResult: &cpb.OrderingResult{
				SubmittedKeys: answer.SubmittedKeys,
				CorrectKeys:   answer.CorrectKeys,
			},
		},
	}, nil
}

func validateCheckQuizCorrectnessRequest(req *sspb.CheckQuizCorrectnessRequest) error {
	if req.ShuffledQuizSetId == "" {
		return fmt.Errorf("req must have ShuffledQuizSetId")
	}
	if req.QuizId == "" {
		return fmt.Errorf("req must have QuizId")
	}
	if len(req.Answer) == 0 {
		return fmt.Errorf("req must have Answer")
	}

	return nil
}

type CorrectnessQuiz struct {
	*entities.Quiz
	*entities.CorrectnessInfo
	Answers []*sspb.Answer
}

func (q *CorrectnessQuiz) Check() (*entities.QuizAnswer, error) {
	switch q.Kind.String {
	case cpb.QuizType_QUIZ_TYPE_MCQ.String(), cpb.QuizType_QUIZ_TYPE_MAQ.String(), cpb.QuizType_QUIZ_TYPE_MIQ.String():
		return CheckMultipleChoice(q)
	case cpb.QuizType_QUIZ_TYPE_FIB.String(), cpb.QuizType_QUIZ_TYPE_POW.String(), cpb.QuizType_QUIZ_TYPE_TAD.String():
		return CheckFillInBlank(q)
	case cpb.QuizType_QUIZ_TYPE_ORD.String():
		return CheckOrder(q)
	}
	return nil, fmt.Errorf("quiz type is not supported")
}

func CheckMultipleChoice(q *CorrectnessQuiz) (*entities.QuizAnswer, error) {
	if _, IsSelectedIndex := q.Answers[0].Format.(*sspb.Answer_SelectedIndex); !IsSelectedIndex {
		return nil, fmt.Errorf("your answer is not the selected index type")
	}

	now := time.Now()

	// options store the important information of the correctness answers
	options, err := q.GetOptions()
	if err != nil {
		return nil, fmt.Errorf("CheckMultipleChoice.Quiz.GetOptions: %w", err)
	}
	if len(options) < len(q.Answers) {
		return nil, fmt.Errorf("the number of answers cannot be higher than the number of options")
	}

	// Re-simulate answers order by random seed
	if q.Kind.String != cpb.QuizType_QUIZ_TYPE_MIQ.String() { // If quiz type is manual input quiz type, we keep first option is true, second option is false
		seed, err := strconv.ParseInt(q.RandomSeed.String, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("CheckMultipleChoice.strconv.ParseInt: %w", err)
		}
		r := rand.New(rand.NewSource(seed + int64(q.QuizIndex.Int)))
		r.Shuffle(len(options), func(i, j int) { options[i], options[j] = options[j], options[i] })
	}

	// Get correctIndex from the Re-simulate answers order OR the options (QUIZ_TYPE_MIQ)
	correctIndex := make([]uint32, 0)
	for i, option := range options {
		if option.Correctness {
			correctIndex = append(correctIndex, uint32(i+1))
		}
	}

	// Get the order of user selection and check correctness
	selectedIndex := make([]uint32, 0)
	correctness := make([]bool, 0, len(q.Answers))
	var IsExistCorrectAnswer, IsExistWrongAnswer bool
	for _, answer := range q.Answers {
		idx := answer.GetSelectedIndex()
		selectedIndex = append(selectedIndex, idx)
		// Check correctness by comparing correctIndex & selectedIndex
		correctness = append(correctness, options[idx-1].Correctness)
		if options[idx-1].Correctness {
			IsExistCorrectAnswer = true
		} else {
			IsExistWrongAnswer = true
		}
	}

	var isAccepted bool
	// Partial config is true
	if utils.IsContain(options[0].Configs, entities.QuizOptionConfigPartialCredit) {
		// Only 1 correct answer is required
		if IsExistCorrectAnswer {
			isAccepted = true
		}
	} else {
		// All answers are correct
		if !IsExistWrongAnswer {
			isAccepted = true
		}
	}

	// Get gained point for the quiz
	var point uint32
	if isAccepted {
		point = uint32(q.Point.Int)
	}

	return &entities.QuizAnswer{
		QuizID:        q.ExternalID.String,
		QuizType:      q.Kind.String,
		SelectedIndex: selectedIndex,
		CorrectIndex:  correctIndex,
		Correctness:   correctness,
		IsAccepted:    isAccepted,
		IsAllCorrect:  isAccepted,
		SubmittedAt:   now,
		Point:         point,
	}, nil
}

func CheckFillInBlank(q *CorrectnessQuiz) (*entities.QuizAnswer, error) {
	if _, IsFilledText := q.Answers[0].Format.(*sspb.Answer_FilledText); !IsFilledText {
		return nil, fmt.Errorf("your answer is not the filled text type")
	}

	now := time.Now()

	// options store the important information of the correctness answers
	options, err := q.GetOptionsWithAlternatives()
	if err != nil {
		return nil, fmt.Errorf("CheckMultipleChoice.Quiz.GetOptionsWithAlternatives: %w", err)
	}
	if len(options) < len(q.Answers) {
		return nil, fmt.Errorf("the number of answers cannot be higher than the number of options")
	}

	// Get correctText from the options
	correctText := make([]string, 0)
	for _, option := range options {
		content := strings.TrimSpace(option.GetText())
		correctText = append(correctText, content)
	}

	// Get user input and check correctness
	filledText := make([]string, 0)
	correctness := make([]bool, 0, len(q.Answers))
	var IsExistCorrectAnswer, IsExistWrongAnswer bool
	for i, answer := range q.Answers {
		content := answer.GetFilledText()
		filledText = append(filledText, content)
		// Check correctness by comparing correctIndex & selectedIndex
		isCorrect := options[i].IsCorrect(content)
		correctness = append(correctness, isCorrect)
		if isCorrect {
			IsExistCorrectAnswer = true
		} else {
			IsExistWrongAnswer = true
		}
	}

	var isAccepted bool
	// Partial config is true
	if utils.IsContain(options[0].AlternativeOptions[0].Configs, entities.QuizOptionConfigPartialCredit) {
		// Only 1 correct answer is required
		if IsExistCorrectAnswer {
			isAccepted = true
		}
	} else {
		// All answers are correct
		if !IsExistWrongAnswer {
			isAccepted = true
		}
	}

	// Get gained point for the quiz
	var point uint32
	if isAccepted {
		point = uint32(q.Point.Int)
	}

	return &entities.QuizAnswer{
		QuizID:       q.ExternalID.String,
		QuizType:     q.Kind.String,
		FilledText:   filledText,
		CorrectText:  correctText,
		Correctness:  correctness,
		IsAccepted:   isAccepted,
		IsAllCorrect: isAccepted,
		SubmittedAt:  now,
		Point:        point,
	}, nil
}

func CheckOrder(q *CorrectnessQuiz) (*entities.QuizAnswer, error) {
	answersEnt, err := (&question.Service{}).CheckQuestionsCorrectness([]*entities.Quiz{q.Quiz}, question.WithCheckQuizCorrectnessRequest([]*sspb.CheckQuizCorrectnessRequest{
		{
			QuizId: q.ExternalID.String,
			Answer: q.Answers,
		},
	}))
	if err != nil {
		return nil, fmt.Errorf("questionSrv.CheckQuestionsCorrectness: %w", err)
	}
	if len(answersEnt) == 0 {
		return nil, fmt.Errorf("questionSrv.CheckQuestionsCorrectness: could not check correctness for quiz %s", q.ID.String)
	}
	answer := answersEnt[0]
	return answer, nil
}

func toLOSubmissionAnswerEntityByLoKey(loKey *entities.LOSubmissionAnswerKey, answerEnt *entities.QuizAnswer) (*entities.LOSubmissionAnswer, error) {
	e := &entities.LOSubmissionAnswer{}
	database.AllNullEntity(e)

	now := time.Now()
	err := multierr.Combine(
		e.StudentID.Set(loKey.StudentID),
		e.QuizID.Set(answerEnt.QuizID),
		e.SubmissionID.Set(loKey.SubmissionID),
		e.StudyPlanID.Set(loKey.StudyPlanID),
		e.LearningMaterialID.Set(loKey.LearningMaterialID),
		e.ShuffledQuizSetID.Set(loKey.ShuffledQuizSetID),
		e.StudentTextAnswer.Set(answerEnt.FilledText),
		e.CorrectTextAnswer.Set(answerEnt.CorrectText),
		e.StudentIndexAnswer.Set(answerEnt.SelectedIndex),
		e.CorrectIndexAnswer.Set(answerEnt.CorrectIndex),
		e.IsCorrect.Set(answerEnt.Correctness),
		e.Point.Set(answerEnt.Point),
		e.CreatedAt.Set(now),
		e.UpdatedAt.Set(now),
		e.SubmittedKeysAnswer.Set(answerEnt.SubmittedKeys),
		e.CorrectKeysAnswer.Set(answerEnt.CorrectKeys),
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
		return nil, fmt.Errorf("fail to setup lo submission answer: %w", err)
	}

	return e, nil
}

func toFlashCardSubmissionAnswerEntityByFcKey(loKey *entities.FlashCardSubmissionAnswerKey, answerEnt *entities.QuizAnswer) (*entities.FlashCardSubmissionAnswer, error) {
	e := &entities.FlashCardSubmissionAnswer{}
	database.AllNullEntity(e)

	now := time.Now()
	err := multierr.Combine(
		e.StudentID.Set(loKey.StudentID),
		e.QuizID.Set(answerEnt.QuizID),
		e.SubmissionID.Set(loKey.SubmissionID),
		e.StudyPlanID.Set(loKey.StudyPlanID),
		e.LearningMaterialID.Set(loKey.LearningMaterialID),
		e.ShuffledQuizSetID.Set(loKey.ShuffledQuizSetID),
		e.StudentTextAnswer.Set(answerEnt.FilledText),
		e.CorrectTextAnswer.Set(answerEnt.CorrectText),
		e.StudentIndexAnswer.Set(answerEnt.SelectedIndex),
		e.CorrectIndexAnswer.Set(answerEnt.CorrectIndex),
		e.IsCorrect.Set(answerEnt.Correctness),
		e.Point.Set(answerEnt.Point),
		e.CreatedAt.Set(now),
		e.UpdatedAt.Set(now),
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
		return nil, fmt.Errorf("fail to setup fc submission answer: %w", err)
	}

	return e, nil
}

func (s *QuizService) RegenerateSpeechesAudioLink(ctx context.Context) error {
	limit := int64(10)
	offset := int64(0)

	updateSpeechesStmt := `
	UPDATE quizzes q
	SET question = replace(question::text, s.old_link, s.new_link)::jsonb,
	options = replace(options::text, old_link, new_link)::jsonb
	FROM UNNEST($1::TEXT[], $2::TEXT[]) s(old_link, new_link)
	`

	type info struct {
		QuizID   string
		Text     string
		Language string
	}

	getKey := func(i *info) string {
		b, _ := json.Marshal(i)
		return string(b)
	}

	for ; ; offset += limit {
		speechesList, err := s.SpeechesRepo.RetrieveAllSpeaches(ctx, s.DB, database.Int8(limit), database.Int8(offset))
		if err != nil {
			return fmt.Errorf("SpeechesRepo.RetrieveAllSpeaches at offset %d: %w", offset, err)
		}
		if len(speechesList) == 0 {
			break
		}
		options := make([]*bpb.AudioOptionRequest, 0, len(speechesList))

		speechesMap := make(map[string]*yasuo_entities.Speeches)

		for _, speech := range speechesList {
			var configs bpb.AudioConfig
			if err := speech.Settings.AssignTo(&configs); err != nil {
				return fmt.Errorf("speech.Settings.AssignToat offset %d: %w", offset, err)
			}
			req := &bpb.AudioOptionRequest{
				Text:    speech.Sentence.String,
				Configs: &configs,
				QuizId:  speech.QuizID.String,
			}
			speechesMap[getKey(&info{
				QuizID:   speech.QuizID.String,
				Text:     speech.Sentence.String,
				Language: configs.Language,
			})] = speech
			options = append(options, req)
		}

		resp, err := s.BobMediaModifier.GenerateAudioFile(ctx, &bpb.GenerateAudioFileRequest{
			Options: options,
		})
		if err != nil {
			return fmt.Errorf("BobMediaModifier.GenerateAudioFile at offset %d: %w", offset, err)
		}

		oldList := make([]string, 0, len(resp.Options))
		newList := make([]string, 0, len(resp.Options))

		for _, option := range resp.Options {
			if speeches, ok := speechesMap[getKey(&info{
				QuizID:   option.QuizId,
				Text:     option.Text,
				Language: option.Configs.Language,
			})]; ok {
				oldList = append(oldList, speeches.Link.String)
				newList = append(newList, option.Link)

				if err := speeches.Link.Set(option.Link); err != nil {
					return fmt.Errorf("speeches.Link.Set at offset %d: %w", offset, err)
				}
			}
		}

		if _, err := s.SpeechesRepo.UpsertSpeeches(ctx, s.DB, speechesList); err != nil {
			return fmt.Errorf("speechesRepo.UpsertSpeeches at offset %d: %w", offset, err)
		}
		if _, err := s.DB.Exec(ctx, updateSpeechesStmt, database.TextArray(oldList), database.TextArray(newList)); err != nil {
			return fmt.Errorf("update speeches for quizzes at offset %d: %w", offset, err)
		}
	}
	return nil
}
