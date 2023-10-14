package services

import (
	"context"
	"crypto/md5" // nolint
	"errors"
	"fmt"
	"io"
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
	"github.com/manabie-com/backend/internal/golibs/timeutil"
	"github.com/manabie-com/backend/internal/golibs/unleashclient"
	yasuo_entities "github.com/manabie-com/backend/internal/yasuo/entities"
	"github.com/manabie-com/backend/internal/yasuo/utils"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
	ypb "github.com/manabie-com/backend/pkg/manabuf/yasuo/v1"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type QuizModifierService struct {
	DB            database.Ext
	UnleashClient unleashclient.ClientInstance
	Env           string
	epb.UnimplementedQuizModifierServiceServer

	StudyPlanRepo interface {
		RetrieveStudyPlanItemInfo(ctx context.Context, db database.QueryExecer, args repositories.StudyPlanItemInfoArgs) ([]*repositories.StudyPlanItemInfo, error)
	}

	QuizRepo interface {
		Search(ctx context.Context,
			db database.QueryExecer, filter repositories.QuizFilter) (entities.Quizzes, error)
		Create(ctx context.Context, db database.QueryExecer, quiz *entities.Quiz) error
		Upsert(ctx context.Context, db database.QueryExecer, data []*entities.Quiz) ([]*entities.Quiz, error)
		DeleteByExternalID(ctx context.Context, db database.QueryExecer, id pgtype.Text, schoolID pgtype.Int4) error
		GetByExternalID(ctx context.Context, db database.QueryExecer, id pgtype.Text, schoolID pgtype.Int4) (*entities.Quiz, error)
		GetByExternalIDsAndLOID(context.Context, database.QueryExecer, pgtype.TextArray, pgtype.Text) (entities.Quizzes, error)
		GetByExternalIDs(context.Context, database.QueryExecer, pgtype.TextArray, pgtype.Text) (entities.Quizzes, error)
		GetByExternalIDsAndLmID(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray, lmID pgtype.Text) (entities.Quizzes, error)
	}
	QuizSetRepo interface {
		Search(ctx context.Context,
			db database.QueryExecer, filter repositories.QuizSetFilter) (entities.QuizSets, error)
		Create(ctx context.Context, db database.QueryExecer, e *entities.QuizSet) error
		Delete(ctx context.Context, db database.QueryExecer, id pgtype.Text) error
		GetQuizSetsContainQuiz(ctx context.Context, db database.QueryExecer, quizID pgtype.Text) (entities.QuizSets, error)
		GetQuizSetByLoID(context.Context, database.QueryExecer, pgtype.Text) (*entities.QuizSet, error)
		GetTotalQuiz(context.Context, database.QueryExecer, pgtype.TextArray) (map[string]int32, error)
		BulkUpsert(ctx context.Context, db database.QueryExecer, data []*entities.QuizSet) ([]*entities.QuizSet, error)
		RetrieveByLoIDs(ctx context.Context, db database.QueryExecer, loIDs pgtype.TextArray) (entities.QuizSets, error)
		GetQuizSetsOfLOContainQuiz(ctx context.Context, db database.QueryExecer, loID pgtype.Text, quizID pgtype.Text) (entities.QuizSets, error)
		GetTotalPointsByQuizSetID(ctx context.Context, db database.QueryExecer, id pgtype.Text) (pgtype.Int4, error)
	}

	YasuoCourseReader interface {
		ValidateUserSchool(ctx context.Context, req *ypb.ValidateUserSchoolRequest, opts ...grpc.CallOption) (*ypb.ValidateUserSchoolResponse, error)
	}

	SpeechesRepo interface {
		UpsertSpeeches(ctx context.Context, db database.QueryExecer, data []*yasuo_entities.Speeches) ([]*yasuo_entities.Speeches, error)
		CheckExistedSpeech(ctx context.Context, db database.QueryExecer, input *repositories.CheckExistedSpeechReq) (bool, *yasuo_entities.Speeches)
	}

	BobMediaModifier interface {
		GenerateAudioFile(ctx context.Context, in *bpb.GenerateAudioFileRequest, opts ...grpc.CallOption) (*bpb.GenerateAudioFileResponse, error)
	}
	YasuoUploadReader interface {
		RetrieveUploadInfo(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*ypb.RetrieveUploadInfoResponse, error)
	}

	YasuoUploadModifier interface {
		UploadHtmlContent(ctx context.Context, in *ypb.UploadHtmlContentRequest, opts ...grpc.CallOption) (*ypb.UploadHtmlContentResponse, error)
	}

	ShuffledQuizSetRepo interface {
		Create(context.Context, database.QueryExecer, *entities.ShuffledQuizSet) (pgtype.Text, error)
		Get(context.Context, database.QueryExecer, pgtype.Text, pgtype.Int8, pgtype.Int8) (*entities.ShuffledQuizSet, error)
		Retrieve(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) ([]*entities.ShuffledQuizSet, error)
		UpdateSubmissionHistory(context.Context, database.QueryExecer, pgtype.Text, pgtype.JSONB) error
		UpdateTotalCorrectness(context.Context, database.QueryExecer, pgtype.Text) error
		GetStudentID(context.Context, database.QueryExecer, pgtype.Text) (pgtype.Text, error)
		GetLoID(context.Context, database.QueryExecer, pgtype.Text) (pgtype.Text, error)
		GetScore(context.Context, database.QueryExecer, pgtype.Text) (pgtype.Int4, pgtype.Int4, error)
		IsFinishedQuizTest(context.Context, database.QueryExecer, pgtype.Text) (pgtype.Bool, error)
		GetExternalIDsFromSubmissionHistory(ctx context.Context, db database.QueryExecer, shuffleQuizSetID pgtype.Text, isAccepted bool) (pgtype.TextArray, error)
		GetSeed(context.Context, database.QueryExecer, pgtype.Text) (pgtype.Text, error)
		GetQuizIdx(context.Context, database.QueryExecer, pgtype.Text, pgtype.Text) (pgtype.Int4, error)
		GetByStudyPlanItems(context.Context, database.QueryExecer, pgtype.TextArray) (entities.ShuffledQuizSets, error)
		GetBySessionID(ctx context.Context, db database.QueryExecer, sessionID pgtype.Text) (*entities.ShuffledQuizSet, error)
		GetShuffledQuizSetIDsByOriginalQuizSetID(ctx context.Context, db database.QueryExecer, originalQuizSetID pgtype.Text) ([]string, error)
	}

	StudentsLearningObjectivesCompletenessRepo interface {
		UpsertHighestQuizScore(ctx context.Context, db database.QueryExecer, loID pgtype.Text, studentID pgtype.Text, newScore pgtype.Float4) error
		UpsertFirstQuizCompleteness(ctx context.Context, db database.QueryExecer, loID pgtype.Text, studentID pgtype.Text, firstQuizScore pgtype.Float4) error
	}

	FlashcardProgressionRepo interface {
		Create(ctx context.Context, db database.QueryExecer, flashcardProgression *entities.FlashcardProgression) (pgtype.Text, error)
		Upsert(ctx context.Context, db database.QueryExecer, cc []*entities.FlashcardProgression) error
		Get(ctx context.Context, db database.QueryExecer, args *repositories.GetFlashcardProgressionArgs) (*entities.FlashcardProgression, error)
		GetByStudySetID(ctx context.Context, db database.QueryExecer, studySetID pgtype.Text) (*entities.FlashcardProgression, error)
		GetByStudySetIDAndStudentID(ctx context.Context, db database.QueryExecer, studentID, studySetID pgtype.Text) (*entities.FlashcardProgression, error)
		UpdateCompletedAt(ctx context.Context, db database.QueryExecer, studySetID pgtype.Text) error
		DeleteByStudySetID(ctx context.Context, db database.QueryExecer, studySetID pgtype.Text) error
	}

	QuestionGroupRepo interface {
		GetByQuestionGroupIDAndLoID(context.Context, database.QueryExecer, pgtype.Text, pgtype.Text) (*entities.QuestionGroup, error)
		GetQuestionGroupsByIDs(ctx context.Context, db database.QueryExecer, ids ...string) (entities.QuestionGroups, error)
	}

	ExamLOSubmissionAnswerRepo interface {
		UpdateAcceptedQuizPointsByQuizID(ctx context.Context, db database.QueryExecer, quizID pgtype.Text, newPoint pgtype.Int4) error
	}

	ExamLOSubmissionRepo interface {
		Get(ctx context.Context, db database.QueryExecer, args *repositories.GetExamLOSubmissionArgs) (*entities.ExamLOSubmission, error)
		GetTotalGradedPoint(ctx context.Context, db database.QueryExecer, submissionID pgtype.Text) (pgtype.Int4, error)
		UpdateExamSubmissionTotalPointsWithResult(ctx context.Context, db database.QueryExecer, submissionID pgtype.Text, newTotalPoints pgtype.Int4, newExamResult pgtype.Text) error
		UpdateExamSubmissionTotalPoints(ctx context.Context, db database.QueryExecer, submissionID pgtype.Text, newTotalPoints pgtype.Int4) error
	}

	ExamLORepo interface {
		Get(ctx context.Context, db database.QueryExecer, learningMaterialID pgtype.Text) (*entities.ExamLO, error)
	}
}

func NewQuizModifierService(
	db database.Ext,
	unleashClient unleashclient.ClientInstance,
	env string,
	yasuoCourseReader ypb.CourseReaderServiceClient,
	bobMediaModifier bpb.MediaModifierServiceClient,
	yasuoUploadReader ypb.UploadReaderServiceClient,
	yasuoUploadModifier ypb.UploadModifierServiceClient,
) *QuizModifierService {
	return &QuizModifierService{
		DB:                       db,
		UnleashClient:            unleashClient,
		Env:                      env,
		QuizRepo:                 &repositories.QuizRepo{},
		QuizSetRepo:              &repositories.QuizSetRepo{},
		YasuoCourseReader:        yasuoCourseReader,
		SpeechesRepo:             &repositories.SpeechesRepository{},
		BobMediaModifier:         bobMediaModifier,
		YasuoUploadReader:        yasuoUploadReader,
		YasuoUploadModifier:      yasuoUploadModifier,
		ShuffledQuizSetRepo:      &repositories.ShuffledQuizSetRepo{},
		FlashcardProgressionRepo: &repositories.FlashcardProgressionRepo{},
		StudentsLearningObjectivesCompletenessRepo: &repositories.StudentsLearningObjectivesCompletenessRepo{},
		StudyPlanRepo:              &repositories.StudyPlanRepo{},
		QuestionGroupRepo:          &repositories.QuestionGroupRepo{},
		ExamLOSubmissionAnswerRepo: &repositories.ExamLOSubmissionAnswerRepo{},
		ExamLOSubmissionRepo:       &repositories.ExamLOSubmissionRepo{},
		ExamLORepo:                 &repositories.ExamLORepo{},
	}
}

// AssignLosToQuiz for old version quizzes which dont have column lo ids, we need to migrate assign lo ids to old quizzes
func (s *QuizModifierService) AssignLosToQuiz(ctx context.Context, quiz *entities.Quiz) error {
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

func inTextArray(ta pgtype.TextArray, s pgtype.Text) bool {
	for _, t := range ta.Elements {
		if t.Status == pgtype.Present && s.Status == pgtype.Present && t.String == s.String {
			return true
		}
	}
	return false
}

func (s *QuizModifierService) quizReqPbToEnt(req *epb.QuizLO, q *entities.Quiz, info *ypb.RetrieveUploadInfoResponse) (contentChange map[string]string, err error) {
	endpoint := info.GetEndpoint()
	bucket := info.GetBucket()
	contentChange = make(map[string]string)
	q.ExternalID = database.Text(req.GetQuiz().GetExternalId())
	q.Country = database.Text(req.GetQuiz().GetInfo().GetCountry().String())
	q.SchoolID = database.Int4(req.GetQuiz().GetInfo().GetSchoolId())
	if q.LoIDs.Status != pgtype.Present {
		q.LoIDs = database.TextArray([]string{req.GetLoId()})
	}
	q.Kind = database.Text(req.GetQuiz().GetKind().String())
	q.TaggedLOs = database.TextArray(req.GetQuiz().GetTaggedLos())
	q.DifficultLevel = database.Int4(req.GetQuiz().GetDifficultyLevel())
	q.LabelType = database.Text(req.Quiz.GetLabelType().String())
	if req.Quiz.Point != nil {
		q.Point = database.Int4(req.GetQuiz().GetPoint().GetValue())
	} else {
		q.Point = database.Int4(1) // default value
	}
	q.QuestionTagIds = database.TextArray(req.GetQuiz().GetQuestionTagIds())
	q.Status = database.Text(epb.QuizStatus_QUIZ_STATUS_APPROVED.String())
	q.CreatedAt = database.Timestamptz(time.Now())
	q.UpdatedAt = database.Timestamptz(time.Now())

	url, _ := generateUploadURL(endpoint, bucket, req.GetQuiz().GetQuestion().GetRendered())
	question, err := q.GetQuestion()
	if err != nil {
		return nil, fmt.Errorf("err GetQuestion: %w", err)
	}

	configs := make([]string, 0)
	for _, each := range req.GetQuiz().GetAttribute().GetConfigs() {
		configs = append(configs, each.String())
	}

	if question.RenderedURL != url {
		q.Question = database.JSONB(&entities.QuizQuestion{
			Raw:         req.GetQuiz().GetQuestion().GetRaw(),
			RenderedURL: url,
			Attribute: entities.QuizItemAttribute{
				ImgLink: req.GetQuiz().GetAttribute().GetImgLink(),
				Configs: configs,
			},
		})

		contentChange[url] = req.GetQuiz().GetQuestion().GetRendered()
	}

	url, _ = generateUploadURL(endpoint, bucket, req.GetQuiz().GetExplanation().GetRendered())

	e, err := q.GetExplaination()
	if err != nil {
		return nil, fmt.Errorf("err GetExplaination: %w", err)
	}

	if e.RenderedURL != url {
		q.Explanation = database.JSONB(&entities.RichText{
			Raw:         req.GetQuiz().GetExplanation().GetRaw(),
			RenderedURL: url,
		})

		contentChange[url] = req.GetQuiz().GetExplanation().GetRendered()
	}

	originOption, _ := q.GetOptions()
	quizOptions := make([]*entities.QuizOption, 0, len(req.GetQuiz().GetOptions()))
	for i, o := range req.GetQuiz().GetOptions() {
		url, _ = generateUploadURL(endpoint, bucket, o.GetContent().GetRendered())
		if err != nil {
			return nil, fmt.Errorf("err UploadHtmlContent Option: %w", err)
		}

		content := entities.RichText{
			Raw:         o.GetContent().GetRaw(),
			RenderedURL: url,
		}

		isChange := true
		if i < len(originOption) {
			opt := originOption[i]
			isChange = opt != nil && opt.Content.RenderedURL != url
		}

		if isChange {
			contentChange[url] = o.GetContent().GetRendered()
		}

		configs = []string{}
		for _, each := range o.GetAttribute().GetConfigs() {
			configs = append(configs, each.String())
		}

		var answerConfig entities.AnswerConfig
		if reqEssayConfig, ok := req.GetQuiz().GetAnswerConfig().(*cpb.QuizCore_Essay); ok {
			answerConfig.Essay = entities.EssayConfig{
				LimitEnabled: reqEssayConfig.Essay.GetLimitEnabled(),
				LimitType:    entities.EssayLimitType(reqEssayConfig.Essay.LimitType.String()),
				Limit:        uint32(reqEssayConfig.Essay.GetLimit()),
			}
		}

		quizOptions = append(quizOptions, &entities.QuizOption{
			Content:     content,
			Correctness: o.Correctness,
			Configs:     quizCfg2ArrayString(o.GetConfigs()),
			Label:       o.GetLabel(),
			Key:         o.GetKey(),
			Attribute: entities.QuizItemAttribute{
				Configs: configs,
			},
			AnswerConfig: answerConfig,
		})
	}

	q.Options = database.JSONB(quizOptions)
	return contentChange, err
}

func quizReqValidate(req *epb.QuizLO) error {
	if req.Quiz == nil {
		return status.Error(codes.InvalidArgument, "missing Quiz")
	}

	if req.Quiz.ExternalId == "" {
		return status.Error(codes.InvalidArgument, "missing ExternalId")
	}

	if req.Quiz.Question == nil {
		return status.Error(codes.InvalidArgument, "missing Question")
	}

	if req.Quiz.Explanation == nil {
		return status.Error(codes.InvalidArgument, "missing Explanation")
	}

	if len(req.Quiz.Options) == 0 {
		return status.Error(codes.InvalidArgument, "missing Options")
	}
	return nil
}

func quizCfg2ArrayString(cfg interface{}) []string {
	switch cfgs := cfg.(type) {
	case []epb.QuizOptionConfig:
		result := make([]string, 0, len(cfgs))
		for _, c := range cfgs {
			result = append(result, c.String())
		}

		return result
	case []cpb.QuizOptionConfig:
		result := make([]string, 0, len(cfgs))
		for _, c := range cfgs {
			result = append(result, c.String())
		}

		return result
	default:
		return nil
	}
}

// nolint
func generateUploadURL(endpoint, bucket, content string) (url, fileName string) {
	h := md5.New()
	io.WriteString(h, content)
	fileName = "/content/" + fmt.Sprintf("%x.html", h.Sum(nil))

	return endpoint + "/" + bucket + fileName, fileName
}

func foundExternalID(a []pgtype.Text, s string) bool {
	var found bool
	for _, e := range a {
		if e.String == s {
			found = true
			break
		}
	}

	return found
}

func removeIndex(s []int, index int) []int {
	return append(s[:index], s[index+1:]...)
}

func (s *QuizModifierService) CreateQuizTest(ctx context.Context, req *epb.CreateQuizTestRequest) (*epb.CreateQuizTestResponse, error) {
	err := s.validateCreateQuizTestRequest(req)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Get shuffledQuizset
	// if this is the firs time call, the req's SetID will be empty.
	// it means we need to create new shuffled quiz set, then response the new shuffled quiz set id for the later call
	shuffledQuizSet := &entities.ShuffledQuizSet{}
	database.AllNullEntity(shuffledQuizSet)
	if req.SetId.GetValue() == "" {
		existedShuffledQuizSet := &entities.ShuffledQuizSet{ID: pgtype.Text{String: ""}}
		if req.GetSessionId() != "" {
			// check this session_id existed any shuffled quiz set or not
			existedShuffledQuizSet, err = s.ShuffledQuizSetRepo.GetBySessionID(ctx, s.DB, database.Text(req.SessionId))
			if err != nil && err.Error() != pgx.ErrNoRows.Error() {
				return nil, status.Errorf(codes.Internal, fmt.Errorf("s.ShuffledQuizSetRepo.GetBySessionID: %v", err).Error())
			}
		}

		if existedShuffledQuizSet != nil && existedShuffledQuizSet.ID.String != "" {
			err := shuffledQuizSet.ID.Set(existedShuffledQuizSet.ID.String)
			if err != nil {
				return nil, status.Errorf(codes.Internal, err.Error())
			}
		} else {
			// get quizset of the learning objective
			quizSet, err := s.QuizSetRepo.GetQuizSetByLoID(ctx, s.DB, database.Text(req.LoId))
			if err != nil {
				if err.Error() == pgx.ErrNoRows.Error() {
					return nil, status.Error(codes.NotFound, "Doesn't have any quiz set")
				}
				return nil, status.Errorf(codes.Internal, fmt.Errorf("CreateQuizTest.QuizSetRepo.GetQuizSetByLoID: %v", err).Error())
			}
			// instance new shuffledQuizSet
			seed := time.Now().UTC().UnixNano()
			now := timeutil.Now()
			studyPlanItemID := database.Text(req.StudyPlanItemId)
			sessionID := database.Text(req.SessionId)
			if len(req.StudyPlanItemId) == 0 {
				_ = studyPlanItemID.Set(nil)
			}
			if req.SessionId == "" {
				_ = sessionID.Set(nil)
			}
			err = multierr.Combine(
				shuffledQuizSet.ID.Set(idutil.ULIDNow()),
				shuffledQuizSet.StudentID.Set(req.StudentId),
				shuffledQuizSet.StudyPlanItemID.Set(studyPlanItemID),
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

			s.generateShuffledQuizSetRandomSeed(ctx, shuffledQuizSet, req.LoId)

			if !req.KeepOrder {
				// shuffle quiz external ids
				shuffledQuizSet.ShuffleQuiz()
			}
			shuffledQuizSet.ID, err = s.ShuffledQuizSetRepo.Create(ctx, s.DB, shuffledQuizSet)
			if err != nil {
				return nil, status.Errorf(codes.Internal, fmt.Errorf("CreateQuizTest.ShuffledQuizSetRepo.Create: %v", err).Error())
			}
		}
	} else {
		err := shuffledQuizSet.ID.Set(req.SetId.GetValue())
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
	quizzes, err := s.QuizRepo.GetByExternalIDsAndLmID(ctx, s.DB, pagingShuffledQuizSet.QuizExternalIDs, database.Text(req.LoId))
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

	// Convert from entities.Quizzes to []bpb.Quiz
	pbQuizzes, err := toListQuizpb(req.LoId, quizzes)
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Errorf("CreateQuizTest.ToListQuizpb: %v", err).Error())
	}

	// get list question group
	qr, err := getQuestionGroupByQuiz(ctx, s.QuestionGroupRepo, s.DB, quizzes)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	questionGroups, err := entities.QuestionGroupsToQuestionGroupProtoBufMess(qr)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	resp := &epb.CreateQuizTestResponse{
		NextPage: &cpb.Paging{
			Limit: req.Paging.Limit,
			Offset: &cpb.Paging_OffsetInteger{
				OffsetInteger: req.Paging.GetOffsetInteger() + int64(req.Paging.Limit),
			},
		},
		Items:          pbQuizzes,
		QuizzesId:      shuffledQuizSet.ID.String,
		QuestionGroups: questionGroups,
	}

	return resp, nil
}

// CheckQuizCorrectness checks if students answers is correct
func (s *QuizModifierService) CheckQuizCorrectness(ctx context.Context, req *epb.CheckQuizCorrectnessRequest) (_ *epb.CheckQuizCorrectnessResponse, err error) {
	if err := s.validateCheckQuizCorrectnessRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	loID, err := s.ShuffledQuizSetRepo.GetLoID(ctx, s.DB, database.Text(req.SetId))
	if err != nil {
		return nil, fmt.Errorf("CheckQuizCorrectness.ShuffledQuizSetRepo.GetLoID: %v", err)
	}

	quizzes, err := s.QuizRepo.GetByExternalIDs(ctx, s.DB, database.TextArray([]string{req.QuizId}), loID)
	if err != nil {
		return nil, fmt.Errorf("CheckQuizCorrectness.QuizRepo.GetByExternalIDs: %v", err)
	}
	quiz := quizzes.Get(0)
	if quiz == nil {
		return nil, fmt.Errorf("CheckQuizCorrectness.Quizzes.Get null quiz from quizzes")
	}

	var answer *entities.QuizAnswer
	result := (&epb.CheckQuizCorrectnessResponse{}).Result
	switch quiz.Kind.String {
	case cpb.QuizType_QUIZ_TYPE_MCQ.String(), cpb.QuizType_QUIZ_TYPE_MAQ.String():
		MCQuiz := &MultipleChoiceQuiz{
			Quiz:                quiz,
			SetID:               req.SetId,
			ShuffledQuizSetRepo: s.ShuffledQuizSetRepo,
		}

		answer, err = MCQuiz.CheckCorrectness(ctx, s.DB, req.Answer)
		if err != nil {
			return nil, status.Error(codes.FailedPrecondition, fmt.Sprintf("CheckQuizCorrectness.MultipleChoice.CheckCorrectness: %v", err))
		}
	case cpb.QuizType_QUIZ_TYPE_MIQ.String():
		manualInputQuiz := &ManualInputQuiz{
			Quiz: quiz,
		}

		answer, err = manualInputQuiz.CheckCorrectness(ctx, s.DB, req.Answer)
		if err != nil {
			return nil, status.Error(codes.FailedPrecondition, fmt.Sprintf("CheckQuizCorrectness.ManualInputQuiz.CheckCorrectness: %v", err))
		}
	case cpb.QuizType_QUIZ_TYPE_FIB.String(), cpb.QuizType_QUIZ_TYPE_POW.String(), cpb.QuizType_QUIZ_TYPE_TAD.String():
		fillInTheBlankQuiz := &FillInTheBlankQuiz{
			Quiz: quiz,
		}

		answer, err = fillInTheBlankQuiz.CheckCorrectness(ctx, s.DB, req.Answer)
		if err != nil {
			return nil, status.Error(codes.FailedPrecondition, fmt.Sprintf("CheckQuizCorrectness.FillInTheBlank.CheckCorrectness: %v", err))
		}
	case cpb.QuizType_QUIZ_TYPE_ESQ.String():
	case cpb.QuizType_QUIZ_TYPE_ORD.String():
		answersEnt, err := (&question.Service{}).CheckQuestionsCorrectness(quizzes, question.WithSubmitQuizAnswersRequest(&epb.SubmitQuizAnswersRequest{
			SetId: req.SetId,
			QuizAnswer: []*epb.QuizAnswer{
				{
					QuizId: req.QuizId,
					Answer: req.Answer,
				},
			},
		}))
		if err != nil {
			return nil, status.Error(codes.FailedPrecondition, fmt.Sprintf("questionSrv.CheckQuestionsCorrectness: %v", err))
		}
		if len(answersEnt) == 0 {
			return nil, status.Error(codes.FailedPrecondition, fmt.Sprintf("questionSrv.CheckQuestionsCorrectness: could not check correctness for quiz %s", req.QuizId))
		}
		answer = answersEnt[0]
		result = &epb.CheckQuizCorrectnessResponse_OrderingResult{
			OrderingResult: &cpb.OrderingResult{
				SubmittedKeys: answer.SubmittedKeys,
				CorrectKeys:   answer.CorrectKeys,
			},
		}
	default:
		return nil, status.Error(codes.FailedPrecondition, "Not supported quiz type!")
	}

	err = database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		if err := s.ShuffledQuizSetRepo.UpdateSubmissionHistory(ctx, tx, database.Text(req.SetId), database.JSONB(answer)); err != nil {
			return fmt.Errorf("CheckQuizCorrectness.ShuffledQuizSetRepo.UpdateSubmissionHistory: %v", err)
		}

		if err := s.ShuffledQuizSetRepo.UpdateTotalCorrectness(ctx, tx, database.Text(req.SetId)); err != nil {
			return fmt.Errorf("CheckQuizCorrectness.ShuffledQuizSetRepo.UpdateTotalCorrectness: %v", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	err = database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		isFinished, err := s.ShuffledQuizSetRepo.IsFinishedQuizTest(ctx, tx, database.Text(req.SetId))
		if err != nil {
			return fmt.Errorf("CheckQuizCorrectness.ShuffledQuizSetRepo.IsFinishedQuizTest: %v", err)
		}

		shuffledQuizSet, err := s.ShuffledQuizSetRepo.Get(ctx, tx, database.Text(req.SetId), database.Int8(1), database.Int8(1))
		if err != nil {
			return fmt.Errorf("CheckQuizCorrectness.ShuffledQuizSetRepo.Get: %v", err)
		}
		isRetry := shuffledQuizSet.OriginalShuffleQuizSetID.Status == pgtype.Present

		if isRetry || isFinished.Bool {
			studentID, err := s.ShuffledQuizSetRepo.GetStudentID(ctx, tx, database.Text(req.SetId))
			if err != nil {
				return fmt.Errorf("CheckQuizCorrectness.ShuffledQuizSetRepo.GetStudentID: %v", err)
			}

			totalCorrectness, totalQuiz, err := s.ShuffledQuizSetRepo.GetScore(ctx, tx, database.Text(req.SetId))
			if err != nil {
				return fmt.Errorf("CheckQuizCorrectness.ShuffledQuizSetRepo.GetScore: %v", err)
			}

			if isRetry {
				externalIDsFromSubmissionHistory, err := s.ShuffledQuizSetRepo.GetExternalIDsFromSubmissionHistory(ctx, tx, shuffledQuizSet.OriginalShuffleQuizSetID, false)
				if err != nil {
					return status.Error(codes.Internal, fmt.Errorf("CheckQuizCorrectness.ShuffledQuizSetRepo.GetExternalIDsFromSubmissionHistory: %w", err).Error())
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
				return fmt.Errorf("CheckQuizCorrectness.StudentsLearningObjectivesCompletenessRepo.UpsertFirstQuizCompleteness: %v", err)
			}

			if err = s.StudentsLearningObjectivesCompletenessRepo.UpsertHighestQuizScore(ctx, tx, loID, studentID, database.Float4(score)); err != nil {
				return fmt.Errorf("CheckQuizCorrectness.StudentsLearningObjectivesCompletenessRepo.UpsertHighestQuizScore: %v", err)
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	resp := &epb.CheckQuizCorrectnessResponse{
		Correctness:  answer.Correctness,
		IsCorrectAll: answer.IsAllCorrect,
		Result:       result,
	}
	return resp, nil
}

func containStrings(s []string, target string) bool {
	for _, val := range s {
		if target == val {
			return true
		}
	}
	return false
}

func (s *QuizModifierService) validateCreateQuizTestRequest(req *epb.CreateQuizTestRequest) error {
	if req.StudentId == "" {
		return fmt.Errorf("req must have student id")
	}
	if req.LoId == "" {
		return fmt.Errorf("req must have learning objective id")
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

func toListQuizpb(loID string, quizzes entities.Quizzes) ([]*cpb.Quiz, error) {
	pbQuizzes := []*cpb.Quiz{}
	for _, quiz := range quizzes {
		pbquiz, err := toQuizPb(loID, quiz)
		if err != nil {
			return nil, err
		}
		pbQuizzes = append(pbQuizzes, pbquiz)
	}
	return pbQuizzes, nil
}

// DeleteQuiz delete the quiz
func (s *QuizModifierService) DeleteQuiz(ctx context.Context, req *epb.DeleteQuizRequest) (*epb.DeleteQuizResponse, error) {
	if err := s.validateDeleteQuizRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err := database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		quizSets, err := s.QuizSetRepo.GetQuizSetsContainQuiz(ctx, tx, database.Text(req.QuizId))
		if err != nil {
			return err
		}
		if _, err := s.removeQuizFromQuizSets(ctx, tx, req.QuizId, quizSets); err != nil {
			return err
		}
		if err := s.QuizRepo.DeleteByExternalID(ctx, tx, database.Text(req.QuizId), database.Int4(req.SchoolId)); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	return &epb.DeleteQuizResponse{}, nil
}

func (s *QuizModifierService) validateDeleteQuizRequest(req *epb.DeleteQuizRequest) error {
	if req.QuizId == "" {
		return fmt.Errorf("req must have quiz id")
	}

	return nil
}

func (s *QuizModifierService) removeQuizFromQuizSets(ctx context.Context, tx database.QueryExecer, quizID string, quizSets entities.QuizSets) (entities.QuizSets, error) {
	for _, quizSet := range quizSets {
		quizExternalIDs := database.FromTextArray(quizSet.QuizExternalIDs)
		for i, id := range quizExternalIDs {
			if id == quizID {
				quizExternalIDs = append(quizExternalIDs[:i], quizExternalIDs[i+1:]...)
				break
			}
		}
		quizSet.QuizExternalIDs.Set(quizExternalIDs)

		questionHierarchy := entities.QuestionHierarchy{}
		if err := questionHierarchy.UnmarshalJSONBArray(quizSet.QuestionHierarchy); err != nil {
			return nil, fmt.Errorf("unable to unmarshal: %w", err)
		}
		questionHierarchy = questionHierarchy.ExcludeQuestionIDs([]string{quizID})

		quizSet.QuestionHierarchy.Set(questionHierarchy)
	}

	for _, quizSet := range quizSets {
		if err := s.QuizSetRepo.Delete(ctx, tx, quizSet.ID); err != nil {
			return nil, fmt.Errorf("QuizSetRepo.Delete:%w", err)
		}
		if err := s.QuizSetRepo.Create(ctx, tx, quizSet); err != nil {
			return nil, fmt.Errorf("QuizSetRepo.Create:%w", err)
		}
	}

	return quizSets, nil
}
func (s *QuizModifierService) CreateFlashCardStudy(ctx context.Context, req *epb.CreateFlashCardStudyRequest) (*epb.CreateFlashCardStudyResponse, error) {
	var (
		quizExternalIDs   pgtype.TextArray
		originalQuizSetID pgtype.Text
		create            bool
	)

	if err := s.validateCreateFlashcardStudyRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	flashcardProgression := &entities.FlashcardProgression{}

	// create flashcard study when study_set_id is empty or old flashcard study have value with field completed_at and length of skipped_questions_ids > 0
	// return old value flashcard study when study_set_id (req) is not empty and isn't complete
	if req.StudySetId == "" {
		// get quizset of the learning objective
		quizSet, err := s.QuizSetRepo.GetQuizSetByLoID(ctx, s.DB, database.Text(req.LoId))
		if err != nil {
			return nil, status.Errorf(codes.Internal, fmt.Errorf("CreateFlashCardStudy.FlashcardProgressionRepo.GetQuizSetByLoID: %v", err).Error())
		}

		quizExternalIDs = quizSet.QuizExternalIDs
		originalQuizSetID = quizSet.ID
		create = true
	} else {
		baseFlashcardProgression, err := s.FlashcardProgressionRepo.GetByStudySetID(ctx, s.DB, database.Text(req.StudySetId))
		if err != nil {
			return nil, status.Errorf(codes.Internal, fmt.Errorf("CreateFlashCardStudy.FlashcardProgressionRepo.GetByStudySetID: %v", err).Error())
		}

		if !baseFlashcardProgression.CompletedAt.Time.IsZero() {
			if len(baseFlashcardProgression.SkippedQuestionIDs.Elements) == 0 {
				return nil, status.Error(codes.InvalidArgument, "length of skippedQuestionIDs must be greater than 0")
			}

			create = true
			quizExternalIDs = baseFlashcardProgression.SkippedQuestionIDs
			originalQuizSetID = baseFlashcardProgression.OriginalQuizSetID
		}
	}

	if create {
		studyPlanItemID := database.Text(req.StudyPlanItemId)
		if len(req.StudyPlanItemId) == 0 {
			studyPlanItemID.Set(nil)
		}
		var err error
		flashcardProgression, err = s.convert2flashcardProgressionEntity(req, originalQuizSetID, studyPlanItemID, quizExternalIDs)
		if err != nil {
			return nil, fmt.Errorf("CreateFlashCardStudy.convert2flashcardProgressionEntity: %v", err)
		}

		if !req.KeepOrder {
			// shuffle quiz external ids
			flashcardProgression.Shuffle()
		}
		studySetID, err := s.FlashcardProgressionRepo.Create(ctx, s.DB, flashcardProgression)
		if err != nil {
			return nil, status.Errorf(codes.Internal, fmt.Errorf("CreateFlashCardStudy.FlashcardProgressionRepo.Create: %v", err).Error())
		}

		flashcardProgression.StudySetID = studySetID
	} else {
		err := flashcardProgression.StudySetID.Set(req.StudySetId)
		if err != nil {
			return nil, err
		}
	}

	// Get our flashcard Progression with paging
	pbQuizzes, pagingFlashcardProgression, err := getFlashcardProgressionWithPaging(ctx,
		s.DB, s.FlashcardProgressionRepo, s.QuizRepo, req.Paging, flashcardProgression.StudySetID, database.Text(req.StudentId))
	if err != nil {
		return nil, err
	}

	resp := &epb.CreateFlashCardStudyResponse{
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

func (s *QuizModifierService) validateCreateFlashcardStudyRequest(req *epb.CreateFlashCardStudyRequest) error {
	if req.StudentId == "" {
		return fmt.Errorf("req must have student id")
	}
	if req.LoId == "" {
		return fmt.Errorf("req must have learning objective id")
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

func (s *QuizModifierService) convert2flashcardProgressionEntity(
	req *epb.CreateFlashCardStudyRequest,
	originalQuizSetID pgtype.Text, studyPlanItemID pgtype.Text, quizExternalIDs pgtype.TextArray,
) (*entities.FlashcardProgression, error) {
	flashcardProgression := &entities.FlashcardProgression{}
	database.AllNullEntity(flashcardProgression)
	now := timeutil.Now()
	if err := multierr.Combine(
		flashcardProgression.OriginalQuizSetID.Set(originalQuizSetID),
		flashcardProgression.StudySetID.Set(idutil.ULIDNow()),
		flashcardProgression.OriginalStudySetID.Set(req.StudySetId),
		flashcardProgression.StudentID.Set(req.StudentId),
		flashcardProgression.StudyPlanItemID.Set(studyPlanItemID),
		flashcardProgression.LoID.Set(req.LoId),
		flashcardProgression.QuizExternalIDs.Set(quizExternalIDs.Elements),
		flashcardProgression.StudyingIndex.Set(nil),
		flashcardProgression.SkippedQuestionIDs.Set(nil),
		flashcardProgression.RememberedQuestionIDs.Set(nil),
		flashcardProgression.CreatedAt.Set(now),
		flashcardProgression.UpdatedAt.Set(now),
		flashcardProgression.CompletedAt.Set(nil),
		flashcardProgression.DeletedAt.Set(nil),
	); err != nil {
		return nil, err
	}
	return flashcardProgression, nil
}

func (s *QuizModifierService) validateCheckQuizCorrectnessRequest(req *epb.CheckQuizCorrectnessRequest) error {
	if req.SetId == "" {
		return fmt.Errorf("req must have SetId")
	}
	if req.QuizId == "" {
		return fmt.Errorf("req must have QuizId")
	}
	if len(req.Answer) == 0 {
		return fmt.Errorf("req must have answer")
	}

	for i := range req.Answer {
		if !isValidAnswerMessage(req.Answer[i]) {
			return fmt.Errorf(fmt.Sprintf("your answer of quiz_id(%s) is must not empty", req.QuizId))
		}
	}

	return nil
}

type MultipleChoiceQuiz struct {
	*entities.Quiz
	SetID string

	ShuffledQuizSetRepo interface {
		GetSeed(context.Context, database.QueryExecer, pgtype.Text) (pgtype.Text, error)
		GetQuizIdx(context.Context, database.QueryExecer, pgtype.Text, pgtype.Text) (pgtype.Int4, error)
	}
}

func (q *MultipleChoiceQuiz) CheckCorrectness(ctx context.Context, db database.QueryExecer, userAnswer []*epb.Answer) (*entities.QuizAnswer, error) {
	options, err := q.GetOptions()
	if err != nil {
		return nil, fmt.Errorf("CheckQuizCorrectness.Quiz.GetOptions: %v", err)
	}

	type correctCfs struct {
		correctness bool
		configs     []string
	}

	crn := make([]bool, 0, len(userAnswer))
	correctness := make([]correctCfs, 0, len(userAnswer))
	// Multiple choice type
	// Shuffle option by seed only for multiple choice question
	// cause when create the quiz test, we shuffle option of every quiz in the quiz test with seed.
	// So in check quiz correctness, we must use the same seed to shuffle options of quiz
	// to make sure the sequence of options will be the same
	seedStr, err := q.ShuffledQuizSetRepo.GetSeed(ctx, db, database.Text(q.SetID))
	if err != nil {
		return nil, fmt.Errorf("CheckQuizCorrectness.ShuffledQuizSetRepo.GetSeed: %v", err)
	}
	seed, err := strconv.ParseInt(seedStr.String, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("CheckQuizCorrectness.CannotParseSeed: %v", err)
	}
	idx, err := q.ShuffledQuizSetRepo.GetQuizIdx(ctx, db, database.Text(q.SetID), database.Text(q.ExternalID.String))
	if err != nil {
		return nil, fmt.Errorf("CheckQuizCorrectness.GetQuizIdx: %v", err)
	}
	r := rand.New(rand.NewSource(seed + int64(idx.Int)))
	r.Shuffle(len(options), func(i, j int) { options[i], options[j] = options[j], options[i] })
	// if quiz type is manual input quiz type, we keep first option is true, second option is false
	selectedIdxs := make([]uint32, 0)
	for i, uAns := range userAnswer {
		if i >= len(options) {
			break
		}
		_, selectedIndexFormat := uAns.Format.(*epb.Answer_SelectedIndex)
		if !selectedIndexFormat {
			return nil, fmt.Errorf("your answer is not the multiple choice type")
		}

		idx := uAns.GetSelectedIndex()
		if idx == 0 || idx > uint32(len(options)) {
			return nil, fmt.Errorf("selected index out of range")
		}
		selectedIdxs = append(selectedIdxs, idx)
		correctness = append(correctness, correctCfs{
			correctness: options[idx-1].Correctness,
			configs:     options[idx-1].Configs,
		})
		crn = append(crn, options[idx-1].Correctness)
	}

	correctIdx := make([]uint32, 0)
	for i, opt := range options {
		if opt.Correctness {
			correctIdx = append(correctIdx, uint32(i+1))
		}
	}

	isAccepted := false
	isPartial := false
	allCorrect := true

	for _, cr := range correctness {
		if utils.IsContain(cr.configs, entities.QuizOptionConfigPartialCredit) {
			if cr.correctness {
				isAccepted = true
				isPartial = true
				break
			} else {
				allCorrect = false
			}
		} else {
			if !cr.correctness {
				allCorrect = false
				break
			}
		}
	}

	if !isPartial {
		isAccepted = (allCorrect && len(selectedIdxs) == len(correctIdx))
		allCorrect = isAccepted
	}

	var point uint32
	if isAccepted {
		point = uint32(q.Point.Int)
	}

	answer := &entities.QuizAnswer{
		QuizID:        q.ExternalID.String,
		QuizType:      q.Kind.String,
		SelectedIndex: selectedIdxs,
		CorrectIndex:  correctIdx,
		Correctness:   crn,
		IsAccepted:    isAccepted,
		IsAllCorrect:  allCorrect,
		SubmittedAt:   time.Now(),
		Point:         point,
	}
	return answer, nil
}

type ManualInputQuiz struct {
	*entities.Quiz
}

func (q *ManualInputQuiz) CheckCorrectness(ctx context.Context, db database.Ext, userAnswer []*epb.Answer) (*entities.QuizAnswer, error) {
	options, err := q.GetOptions()
	if err != nil {
		return nil, fmt.Errorf("CheckQuizCorrectness.Quiz.GetOptions: %v", err)
	}

	type correctCfs struct {
		correctness bool
		configs     []string
	}

	crn := make([]bool, 0, len(userAnswer))
	correctness := make([]correctCfs, 0, len(userAnswer))
	// if quiz type is manual input quiz type, we keep first option is true, second option is false
	selectedIdxs := make([]uint32, 0)
	for i, uAns := range userAnswer {
		if i >= len(options) {
			break
		}
		_, selectedIndexFormat := uAns.Format.(*epb.Answer_SelectedIndex)
		if !selectedIndexFormat {
			return nil, fmt.Errorf("your answer is not the multiple choice type")
		}

		idx := uAns.GetSelectedIndex()
		if idx == 0 || idx > uint32(len(options)) {
			return nil, fmt.Errorf("selected index out of range")
		}
		selectedIdxs = append(selectedIdxs, idx)
		correctness = append(correctness, correctCfs{
			correctness: options[idx-1].Correctness,
			configs:     options[idx-1].Configs,
		})
		crn = append(crn, options[idx-1].Correctness)
	}

	correctIdx := make([]uint32, 0)
	for i, opt := range options {
		if opt.Correctness {
			correctIdx = append(correctIdx, uint32(i+1))
		}
	}

	isAccepted := false
	isPartial := false
	allCorrect := true

	for _, cr := range correctness {
		if utils.IsContain(cr.configs, entities.QuizOptionConfigPartialCredit) {
			if cr.correctness {
				isAccepted = true
				isPartial = true
				break
			} else {
				allCorrect = false
			}
		} else {
			if !cr.correctness {
				allCorrect = false
				break
			}
		}
	}

	if !isPartial {
		isAccepted = (allCorrect && len(selectedIdxs) == len(correctIdx))
		allCorrect = isAccepted
	}

	var point uint32
	if isAccepted {
		point = uint32(q.Point.Int)
	}

	answer := &entities.QuizAnswer{
		QuizID:        q.ExternalID.String,
		QuizType:      q.Kind.String,
		SelectedIndex: selectedIdxs,
		CorrectIndex:  correctIdx,
		Correctness:   crn,
		IsAccepted:    isAccepted,
		IsAllCorrect:  allCorrect,
		SubmittedAt:   time.Now(),
		Point:         point,
	}
	return answer, nil
}

type FillInTheBlankQuiz struct {
	*entities.Quiz
}

func (q *FillInTheBlankQuiz) CheckCorrectness(ctx context.Context, db database.Ext, userAnswer []*epb.Answer) (*entities.QuizAnswer, error) {
	options, err := q.GetOptionsWithAlternatives()
	if err != nil {
		return nil, fmt.Errorf("CheckQuizCorrectness.Quiz.GetOptions: %v", err)
	}

	filledText := []string{}
	correctText := []string{}
	type correctCfs struct {
		correctness bool
		configs     []string
	}
	crn := make([]bool, 0, len(userAnswer))
	correctness := make([]correctCfs, 0, len(userAnswer))
	for i, uAns := range userAnswer {
		if i >= len(options) {
			break
		}
		_, filledTextFormat := uAns.Format.(*epb.Answer_FilledText)
		if !filledTextFormat {
			return nil, fmt.Errorf("your answer is not the fill in the blank type")
		}
		// text := strings.ToLower(strings.TrimSpace(uAns.GetFilledText()))
		// content := options[i].GetText()
		// content = strings.ToLower(strings.TrimSpace(content))
		isCorrect := options[i].IsCorrect(uAns.GetFilledText())
		correctness = append(correctness, correctCfs{
			correctness: isCorrect,
			configs:     options[i].AlternativeOptions[0].Configs,
		})
		crn = append(crn, isCorrect)
	}

	for _, uAns := range userAnswer {
		filledText = append(filledText, uAns.GetFilledText())
	}

	for _, opt := range options {
		content := opt.GetText()
		content = strings.ToLower(strings.TrimSpace(content))
		correctText = append(correctText, content)
	}

	isAccepted := false
	isPartial := false
	allCorrect := true

	for _, cr := range correctness {
		if utils.IsContain(cr.configs, entities.QuizOptionConfigPartialCredit) {
			if cr.correctness {
				isAccepted = true
				isPartial = true
				allCorrect = true
				break
			} else {
				allCorrect = false
			}
		} else {
			if !cr.correctness {
				allCorrect = false
				break
			}
		}
	}

	if !isPartial {
		isAccepted = (allCorrect && len(filledText) == len(correctText))
		allCorrect = isAccepted
	}

	var point uint32
	if isAccepted {
		point = uint32(q.Point.Int)
	}

	answer := &entities.QuizAnswer{
		QuizID:       q.ExternalID.String,
		QuizType:     q.Kind.String,
		FilledText:   filledText,
		CorrectText:  correctText,
		Correctness:  crn,
		IsAccepted:   isAccepted,
		IsAllCorrect: allCorrect,
		SubmittedAt:  time.Now(),
		Point:        point,
	}
	return answer, nil
}

// Deprecated: Do not use.
func (s *QuizModifierService) upsertQuizReqValidate(req *epb.UpsertQuizRequest) error {
	if req.Quiz == nil {
		return status.Error(codes.InvalidArgument, "missing Quiz")
	}

	if req.Quiz.ExternalId == "" {
		return status.Error(codes.InvalidArgument, "missing ExternalId")
	}

	if req.Quiz.Question == nil {
		return status.Error(codes.InvalidArgument, "missing Question")
	}

	if req.Quiz.Explanation == nil {
		return status.Error(codes.InvalidArgument, "missing Explanation")
	}

	if len(req.Quiz.Options) == 0 {
		return status.Error(codes.InvalidArgument, "missing Options")
	}
	return nil
}

// Deprecated: Do not use.
func (s *QuizModifierService) upsertQuizReqPbToEnt(req *epb.UpsertQuizRequest, q *entities.Quiz, info *ypb.RetrieveUploadInfoResponse) (contentChange map[string]string, err error) {
	endpoint := info.GetEndpoint()
	bucket := info.GetBucket()
	contentChange = make(map[string]string)
	q.ExternalID = database.Text(req.Quiz.ExternalId)
	q.Country = database.Text(req.Quiz.Country.String())
	q.SchoolID = database.Int4(req.Quiz.SchoolId)
	if q.LoIDs.Status != pgtype.Present {
		q.LoIDs = database.TextArray([]string{req.LoId})
	}
	q.Kind = database.Text(req.Quiz.Kind.String())
	q.TaggedLOs = database.TextArray(req.Quiz.TaggedLos)
	q.DifficultLevel = database.Int4(req.Quiz.DifficultyLevel)
	q.Status = database.Text(cpb.QuizStatus_QUIZ_STATUS_APPROVED.String())

	url, _ := generateUploadURL(endpoint, bucket, req.Quiz.Question.Rendered)
	question, err := q.GetQuestion()
	if err != nil {
		return nil, fmt.Errorf("err GetQuestion: %w", err)
	}

	if question.RenderedURL != url {
		q.Question = database.JSONB(&entities.RichText{
			Raw:         req.Quiz.Question.Raw,
			RenderedURL: url,
		})

		contentChange[url] = req.Quiz.Question.Rendered
	}

	url, _ = generateUploadURL(endpoint, bucket, req.Quiz.Explanation.Rendered)

	e, err := q.GetExplaination()
	if err != nil {
		return nil, fmt.Errorf("err GetExplaination: %w", err)
	}

	if e.RenderedURL != url {
		q.Explanation = database.JSONB(&entities.RichText{
			Raw:         req.Quiz.Explanation.Raw,
			RenderedURL: url,
		})

		contentChange[url] = req.Quiz.Explanation.Rendered
	}

	originOption, _ := q.GetOptions()
	quizOptions := make([]*entities.QuizOption, 0, len(req.Quiz.Options))
	for i, o := range req.Quiz.Options {
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

		quizOptions = append(quizOptions, &entities.QuizOption{
			Content:     content,
			Correctness: o.Correctness,
			Configs:     quizCfg2ArrayString(o.Configs),
			Label:       o.Label,
			Key:         o.Key,
		})
	}

	q.Options = database.JSONB(quizOptions)
	return contentChange, err
}

func (s *QuizModifierService) getQuizSet(ctx context.Context, db database.QueryExecer, loID string) (*entities.QuizSet, error) {
	set, err := s.QuizSetRepo.Search(ctx, db, repositories.QuizSetFilter{
		ObjectiveIDs: database.TextArray([]string{loID}),
		Status:       database.Text(epb.QuizSetStatus_QUIZSET_STATUS_APPROVED.String()),
		Limit:        1,
	})
	if err != nil {
		return nil, err
	}

	if len(set) == 0 {
		return nil, fmt.Errorf("no quiz_sets found: %w", pgx.ErrNoRows)
	}

	return set[0], nil
}

func (s *QuizModifierService) UpsertQuiz(ctx context.Context, req *epb.UpsertQuizRequest) (*epb.UpsertQuizResponse, error) {
	if err := s.upsertQuizReqValidate(req); err != nil {
		return nil, err
	}
	userID := interceptors.UserIDFromContext(ctx)
	schoolID := req.Quiz.SchoolId

	quiz, err := s.QuizRepo.GetByExternalID(ctx, s.DB, database.Text(req.Quiz.ExternalId), database.Int4(schoolID))
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if quiz == nil {
		quiz = &entities.Quiz{}
		database.AllNullEntity(quiz)
	} else {
		if len(quiz.LoIDs.Elements) == 0 {
			// migrate from old version quizzes which don't have column lo_ids
			// assign lo_ids to this old version quizzes
			if err := s.AssignLosToQuiz(ctx, quiz); err != nil {
				return nil, status.Error(codes.Internal, "can not assign los to Quiz")
			}
		}
		if !inTextArray(quiz.LoIDs, database.Text(req.LoId)) {
			// not update case
			// this case is other Lo create quiz which is already existed
			return nil, status.Error(codes.Internal, fmt.Sprintf("err quiz is already existed in LOs %v", quiz.LoIDs))
		}
	}

	mdCtx, err := interceptors.GetOutgoingContext(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, err.Error())
	}
	uresp, err := s.YasuoUploadReader.RetrieveUploadInfo(mdCtx, &emptypb.Empty{})
	if err != nil {
		return nil, err
	}

	contentChange, err := s.upsertQuizReqPbToEnt(req, quiz, uresp)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	quiz.CreatedBy = database.Text(userID)
	err = database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		set, errGetQuizSet := s.getQuizSet(ctx, tx, req.LoId)
		if errGetQuizSet != nil && !errors.Is(errGetQuizSet, pgx.ErrNoRows) {
			return errGetQuizSet
		}

		if err := s.QuizRepo.DeleteByExternalID(ctx, tx, database.Text(req.Quiz.ExternalId), database.Int4(schoolID)); err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("s.QuizRepo.DeleteByExternalID: %w", err)
		}

		if err := multierr.Combine(
			quiz.Status.Set(cpb.QuizStatus_QUIZ_STATUS_APPROVED.String()), // for now,
			quiz.ApprovedBy.Set(userID),                                   // for now,
		); err != nil {
			return err
		}

		if err = s.QuizRepo.Create(ctx, tx, quiz); err != nil {
			return fmt.Errorf("s.QuizRepo.Create: %w", err)
		}

		// upload content change
		eg, egctx := errgroup.WithContext(mdCtx)
		for k, c := range contentChange {
			url := k
			content := c

			eg.Go(func() error {
				r, err := s.YasuoUploadModifier.UploadHtmlContent(egctx, &ypb.UploadHtmlContentRequest{
					Content: content,
				})
				if err != nil {
					return fmt.Errorf("s.YasuoUploadModifier.UploadHtmlContent: %w", err)
				}

				if r.GetUrl() != url {
					return fmt.Errorf("url return does not match")
				}

				return nil
			})
		}

		if err := eg.Wait(); err != nil {
			return err
		}

		if set != nil && foundExternalID(set.QuizExternalIDs.Elements, req.Quiz.ExternalId) {
			return nil // for now, no need to create new quiz_set
		}

		if set == nil {
			set = &entities.QuizSet{}
			database.AllNullEntity(set)
			set.LoID = database.Text(req.LoId)
			set.Status = database.Text(epb.QuizSetStatus_QUIZSET_STATUS_APPROVED.String())
		} else {
			err := s.QuizSetRepo.Delete(ctx, tx, set.ID)
			if err != nil {
				return fmt.Errorf("quizSetRepo.Delete: %w", err)
			}
		}

		extIDs := []string{}
		set.QuizExternalIDs.AssignTo(&extIDs)
		extIDs = append(extIDs, req.Quiz.ExternalId)

		if err := multierr.Combine(
			set.QuizExternalIDs.Set(extIDs),
			s.QuizSetRepo.Create(ctx, tx, set),
		); err != nil {
			return fmt.Errorf("quizSetRepo.Create: %w", err)
		}

		return err
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	return &epb.UpsertQuizResponse{
		Id: quiz.ID.String,
	}, nil
}

func (s *QuizModifierService) UpsertSingleQuiz(ctx context.Context, req *epb.UpsertSingleQuizRequest) (*epb.UpsertSingleQuizResponse, error) {
	if err := quizReqValidate(req.QuizLo); err != nil {
		return nil, err
	}
	isUpdateExamSubmissionPointsEnabled, err := s.UnleashClient.IsFeatureEnabled("Syllabus_StudyPlanManagement_BackOffice_UpdateLOExamSubmissionPointsOnQuizPointUpdate", s.Env)
	if err != nil {
		return nil, fmt.Errorf("UnleashClient.IsFeatureEnabled: %w", err)
	}

	userID := interceptors.UserIDFromContext(ctx)
	schoolID := req.GetQuizLo().GetQuiz().GetInfo().GetSchoolId()

	quiz, err := s.QuizRepo.GetByExternalID(ctx, s.DB, database.Text(req.GetQuizLo().GetQuiz().GetExternalId()), database.Int4(schoolID))
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if quiz == nil {
		quiz = &entities.Quiz{}
		database.AllNullEntity(quiz)
	} else {
		if len(quiz.LoIDs.Elements) == 0 {
			// migrate from old version quizzes which don't have column lo_ids
			// assign lo_ids to this old version quizzes
			if err := s.AssignLosToQuiz(ctx, quiz); err != nil {
				return nil, status.Error(codes.Internal, "can not assign los to Quiz")
			}
		}
		if !inTextArray(quiz.LoIDs, database.Text(req.GetQuizLo().GetLoId())) {
			// not update case
			// this case is other Lo create quiz which is already existed
			return nil, status.Error(codes.Internal, fmt.Sprintf("err quiz is already existed in LOs %v", quiz.LoIDs))
		}
	}

	mdCtx, err := interceptors.GetOutgoingContext(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, err.Error())
	}
	uresp, err := s.YasuoUploadReader.RetrieveUploadInfo(mdCtx, &emptypb.Empty{})
	if err != nil {
		return nil, err
	}

	// declare this value here before quiz properties get overridden on quizReqPbToEnt
	hasChangeInQuizPoints := quiz.Point.Status == pgtype.Present && req.QuizLo.Quiz.Point.Value != quiz.Point.Int
	contentChange, err := s.quizReqPbToEnt(req.QuizLo, quiz, uresp)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	quiz.CreatedBy = database.Text(userID)

	questionGroupID := req.GetQuizLo().GetQuiz().GetQuestionGroupId().GetValue()

	if questionGroupID != "" {
		pgQuestionGroupID := database.Text(questionGroupID)
		pgLoID := database.Text(req.GetQuizLo().GetLoId())

		_, err := s.QuestionGroupRepo.GetByQuestionGroupIDAndLoID(ctx, s.DB, pgQuestionGroupID, pgLoID)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		quiz.QuestionGroupID = pgQuestionGroupID
	}

	if err := database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		set, err := s.getQuizSet(ctx, tx, req.GetQuizLo().GetLoId())
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return err
		}

		if err := s.QuizRepo.DeleteByExternalID(ctx, tx, database.Text(req.GetQuizLo().GetQuiz().GetExternalId()), database.Int4(schoolID)); err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("s.QuizRepo.DeleteByExternalID: %w", err)
		}

		if err := multierr.Combine(
			quiz.Status.Set(cpb.QuizStatus_QUIZ_STATUS_APPROVED.String()), // for now,
			quiz.ApprovedBy.Set(userID),                                   // for now,
		); err != nil {
			return err
		}

		if err = s.QuizRepo.Create(ctx, tx, quiz); err != nil {
			return fmt.Errorf("s.QuizRepo.Create: %w", err)
		}

		// upload content change
		eg, egctx := errgroup.WithContext(mdCtx)
		for k, c := range contentChange {
			url := k
			content := c

			eg.Go(func() error {
				r, err := s.YasuoUploadModifier.UploadHtmlContent(egctx, &ypb.UploadHtmlContentRequest{
					Content: content,
				})
				if err != nil {
					return fmt.Errorf("s.YasuoUploadModifier.UploadHtmlContent: %w", err)
				}

				if r.GetUrl() != url {
					return fmt.Errorf("url return does not match")
				}

				return nil
			})
		}

		if err := eg.Wait(); err != nil {
			return err
		}

		if set != nil && foundExternalID(set.QuizExternalIDs.Elements, req.GetQuizLo().GetQuiz().GetExternalId()) {
			if !hasChangeInQuizPoints || !isUpdateExamSubmissionPointsEnabled {
				return nil // return early if no change in quiz point, skip creating quiz_set
			}

			shuffledQuizSetIDs, err := s.ShuffledQuizSetRepo.GetShuffledQuizSetIDsByOriginalQuizSetID(ctx, tx, set.ID)
			if err != nil {
				return fmt.Errorf("error get shuffled quiz set by original quiz set id: %w", err)
			}

			// update new point for existing exam submissions by external quiz id
			err = s.ExamLOSubmissionAnswerRepo.UpdateAcceptedQuizPointsByQuizID(ctx, tx, quiz.ExternalID, quiz.Point)
			if err != nil {
				return fmt.Errorf("error update quiz point by quiz id: %w", err)
			}

			// recalculate total points for the quiz set if we changed points
			totalPoints, err := s.QuizSetRepo.GetTotalPointsByQuizSetID(ctx, tx, set.ID)
			if err != nil {
				return fmt.Errorf("QuizSetRepo.GetTotalPointsByQuizSetID: %w", err)
			}

			for _, shuffledQuizSetID := range shuffledQuizSetIDs {
				examLOSubmission, err := s.ExamLOSubmissionRepo.Get(ctx, tx, &repositories.GetExamLOSubmissionArgs{
					SubmissionID:      pgtype.Text{Status: pgtype.Null},
					ShuffledQuizSetID: database.Text(shuffledQuizSetID),
				})
				if err != nil && !errors.Is(err, pgx.ErrNoRows) {
					return fmt.Errorf("ExamLOSubmissionRepo.Get: %w", err)
				}

				// no submissions yet, skip
				if examLOSubmission == nil {
					continue
				}

				examLO, err := s.ExamLORepo.Get(ctx, tx, examLOSubmission.LearningMaterialID)
				if err != nil {
					return fmt.Errorf("ExamLORepo.Get: %w", err)
				}

				// get the student's exam score
				resultTotalGradedPoint, err := s.ExamLOSubmissionRepo.GetTotalGradedPoint(ctx, tx, examLOSubmission.SubmissionID)
				if err != nil {
					return fmt.Errorf("ExamLOSubmissionRepo.GetTotalGradedPoint: %w", err)
				}

				// update score submission result if exam isn't manually graded
				if !examLO.ManualGrading.Bool {
					submissionResult := examLOSubmission.Result
					// if the new grade is >= grade_to_pass, then update submission result to passed
					if examLO.GradeToPass.Status != pgtype.Null && resultTotalGradedPoint.Int >= examLO.GradeToPass.Int {
						submissionResult = database.Text(epb.ExamLOSubmissionResult_EXAM_LO_SUBMISSION_PASSED.String())
					}

					if err = s.ExamLOSubmissionRepo.UpdateExamSubmissionTotalPointsWithResult(ctx, tx, examLOSubmission.SubmissionID, totalPoints, submissionResult); err != nil {
						return fmt.Errorf("ExamLOSubmissionRepo.UpdateExamSubmissionTotalPointsWithResultByShuffledQuizSetID: %w", err)
					}
				} else {
					// update the score if it's manually graded
					if err = s.ExamLOSubmissionRepo.UpdateExamSubmissionTotalPoints(ctx, tx, examLOSubmission.SubmissionID, totalPoints); err != nil {
						return fmt.Errorf("ExamLOSubmissionRepo.UpdateExamSubmissionTotalPointsByShuffledQuizSetID: %w", err)
					}
				}
			}

			return nil // for now, no need to create new quiz_set
		}

		if set == nil {
			set = &entities.QuizSet{}
			database.AllNullEntity(set)
			set.LoID = database.Text(req.GetQuizLo().GetLoId())
			set.Status = database.Text(epb.QuizSetStatus_QUIZSET_STATUS_APPROVED.String())
		} else {
			if err := s.QuizSetRepo.Delete(ctx, tx, set.ID); err != nil {
				return fmt.Errorf("quizSetRepo.Delete: %w", err)
			}
		}

		questionHierarchyItems, extIDs, err := s.flattenQuestionHierarchy(quiz, set)
		if err != nil {
			return fmt.Errorf("error flatten question hierarchy: %w", err)
		}

		if err := multierr.Combine(
			set.QuizExternalIDs.Set(extIDs),
			set.QuestionHierarchy.Set(questionHierarchyItems),
			s.QuizSetRepo.Create(ctx, tx, set),
		); err != nil {
			return fmt.Errorf("quizSetRepo.Create: %w", err)
		}

		return err
	}); err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	return &epb.UpsertSingleQuizResponse{
		Id: quiz.ID.String,
	}, nil
}

func (s *QuizModifierService) flattenQuestionHierarchy(quiz *entities.Quiz, set *entities.QuizSet) (entities.QuestionHierarchy, []string, error) {
	extIDs := []string{}

	questionHierarchy := entities.QuestionHierarchy{}
	if err := questionHierarchy.UnmarshalJSONBArray(set.QuestionHierarchy); err != nil {
		return nil, nil, fmt.Errorf("unable to unmarshal: %w", err)
	}

	if quiz.QuestionGroupID.Status != pgtype.Present {
		set.QuizExternalIDs.AssignTo(&extIDs)
		extIDs = append(extIDs, quiz.ExternalID.String)

		questionHierarchy.AddQuestionID(quiz.ExternalID.String)
	} else {
		foundQuestionGroup := false

		for _, questionHierarchyObj := range questionHierarchy {
			switch questionHierarchyObj.Type {
			case entities.QuestionHierarchyQuestion:
				extIDs = append(extIDs, questionHierarchyObj.ID)
			case entities.QuestionHierarchyQuestionGroup:
				if quiz.QuestionGroupID.String == questionHierarchyObj.ID {
					foundQuestionGroup = true
					questionHierarchyObj.ChildrenIDs = append(questionHierarchyObj.ChildrenIDs, quiz.ExternalID.String)
				}
				extIDs = append(extIDs, questionHierarchyObj.ChildrenIDs...)
			default:
				return nil, nil, fmt.Errorf("Invalid question type")
			}
		}

		if !foundQuestionGroup {
			return nil, nil, fmt.Errorf("Unable to find question group id in question hierarchy")
		}
	}

	return questionHierarchy, extIDs, nil
}

func (s *QuizModifierService) createNewShuffledQuizSet(ctx context.Context, req *epb.CreateRetryQuizTestRequest) (*entities.ShuffledQuizSet, error) {
	shuffledQuizSet := &entities.ShuffledQuizSet{}
	database.AllNullEntity(shuffledQuizSet)
	// get quizset of the learning objective
	quizSet, err := s.QuizSetRepo.GetQuizSetByLoID(ctx, s.DB, database.Text(req.LoId))
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("CreateRetryQuizTest.QuizSetRepo.GetQuizSetByLoID: %w", err).Error())
	}
	shuffledQuizzes, err := s.ShuffledQuizSetRepo.Retrieve(ctx, s.DB, database.TextArray([]string{req.GetSetId().GetValue()}))
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("CreateRetryQuizTest.ShuffledQuizSetRepo.Retrieve: %w", err).Error())
	}
	if len(shuffledQuizzes) == 0 {
		return nil, status.Error(codes.NotFound, fmt.Errorf("not found any shuffle quiz").Error())
	}
	correctQuizIDs, err := s.ShuffledQuizSetRepo.GetExternalIDsFromSubmissionHistory(ctx, s.DB, database.Text(req.GetSetId().GetValue()), true)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("CreateRetryQuizTest.ShuffledQuizSetRepo.GetCorrectQuizIDs: %w", err).Error())
	}
	retryQuizIDs := getRetryQuizIDs(quizSet.QuizExternalIDs, correctQuizIDs)
	if len(retryQuizIDs) == 0 {
		return nil, status.Error(codes.FailedPrecondition, fmt.Errorf("CreateRetryQuizTest: all quizzes's answers accepted").Error())
	}

	now := timeutil.Now()
	studyPlanItemID := database.Text(req.StudyPlanItemId)
	sessionID := database.Text(req.SessionId)

	currentQuestionHierarchy := entities.QuestionHierarchy{}
	if err := currentQuestionHierarchy.UnmarshalJSONBArray(quizSet.QuestionHierarchy); err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Errorf("CreateRetryQuizTestV2: unable to unmarshal: %w", err).Error())
	}

	retryQuestionHierarchy := currentQuestionHierarchy.ExcludeQuestionIDs(database.FromTextArray(correctQuizIDs))

	err = multierr.Combine(
		shuffledQuizSet.ID.Set(idutil.ULIDNow()),
		shuffledQuizSet.StudentID.Set(req.StudentId),
		shuffledQuizSet.StudyPlanItemID.Set(studyPlanItemID),
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
	s.generateShuffledQuizSetRandomSeed(ctx, shuffledQuizSet, req.LoId)
	shuffledQuizSet.ID, err = s.ShuffledQuizSetRepo.Create(ctx, s.DB, shuffledQuizSet)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("CreateRetryQuizTest.ShuffledQuizSetRepo.Create: %v", err).Error())
	}
	return shuffledQuizSet, nil
}

func (s *QuizModifierService) CreateRetryQuizTest(ctx context.Context, req *epb.CreateRetryQuizTestRequest) (*epb.CreateRetryQuizTestResponse, error) {
	// currently we're not handle `keep_order` field
	if err := s.validateCreateRetryQuizTests(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// if this is the first time call, the req's RetryShuffleQuizId will be empty.
	// it means we need to create new shuffled quiz set, then response the new retry shuffled quiz set id for the later call
	shuffledQuizSetID := req.GetRetryShuffleQuizId().GetValue()
	if shuffledQuizSetID == "" {
		// check this session_id existed any shuffled quiz set or not
		var err error
		existedShuffledQuizSet := &entities.ShuffledQuizSet{ID: pgtype.Text{String: ""}}
		if req.GetSessionId() != "" {
			existedShuffledQuizSet, err = s.ShuffledQuizSetRepo.GetBySessionID(ctx, s.DB, database.Text(req.SessionId))
			if err != nil && err.Error() != pgx.ErrNoRows.Error() {
				return nil, status.Errorf(codes.Internal, fmt.Errorf("s.ShuffledQuizSetRepo.GetBySessionID: %v", err).Error())
			}
		}

		if existedShuffledQuizSet != nil && existedShuffledQuizSet.ID.String != "" {
			shuffledQuizSetID = existedShuffledQuizSet.ID.String
		} else {
			shuffledQuizSet, err := s.createNewShuffledQuizSet(ctx, req)
			if err != nil {
				return nil, err
			}
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
	quizzes, err := s.QuizRepo.GetByExternalIDs(ctx, s.DB, pagingShuffledQuizSet.QuizExternalIDs, database.Text(req.LoId))
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

	// Convert from entities.Quizzes to []bpb.Quiz
	pbQuizzes, err := toListQuizpb(req.LoId, quizzes)
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

	resp := &epb.CreateRetryQuizTestResponse{
		NextPage: &cpb.Paging{
			Limit: req.Paging.Limit,
			Offset: &cpb.Paging_OffsetInteger{
				OffsetInteger: req.Paging.GetOffsetInteger() + int64(req.Paging.Limit),
			},
		},
		Items:          pbQuizzes,
		QuizzesId:      shuffledQuizSetID,
		QuestionGroups: questionGroups,
	}

	return resp, nil
}

func (s *QuizModifierService) validateCreateRetryQuizTests(req *epb.CreateRetryQuizTestRequest) error {
	if len(req.StudyPlanItemId) == 0 {
		return fmt.Errorf("req must have Study Plan Item Id")
	}
	if req.SessionId == "" {
		return fmt.Errorf("req must have session id")
	}
	if req.GetSetId() == nil || req.GetSetId().GetValue() == "" {
		return fmt.Errorf("req must have shuffle quizset id")
	}
	if req.Paging == nil {
		return fmt.Errorf("req must have paging field")
	}
	if req.GetLoId() == "" {
		return fmt.Errorf("req must have lo id")
	}
	if req.Paging.GetOffsetInteger() <= 0 {
		return fmt.Errorf("offset must be positive")
	}

	if req.Paging.Limit <= 0 {
		req.Paging.Limit = 100
	}

	return nil
}

func getRetryQuizIDs(fullQuizIDs, correctQuizIDs pgtype.TextArray) []string {
	retryQuizIDs := make([]string, 0)
	mapCorrectQuizIDs := make(map[string]bool)
	for _, e := range correctQuizIDs.Elements {
		mapCorrectQuizIDs[e.String] = true
	}

	for _, e := range fullQuizIDs.Elements {
		if _, ok := mapCorrectQuizIDs[e.String]; !ok {
			retryQuizIDs = append(retryQuizIDs, e.String)
		}
	}
	return retryQuizIDs
}

// RemoveQuizFromLO remove a quiz from learning objective
func (s *QuizModifierService) RemoveQuizFromLO(ctx context.Context, req *epb.RemoveQuizFromLORequest) (*epb.RemoveQuizFromLOResponse, error) {
	if err := s.validateRemoveQuizFromLORequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err := database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		// var newQuizSets entities.QuizSets
		quizSets, err := s.QuizSetRepo.GetQuizSetsOfLOContainQuiz(ctx, tx, database.Text(req.LoId), database.Text(req.QuizId))
		if err != nil {
			return fmt.Errorf("QuizSetRepo.GetQuizSetsOfLOContainQuiz: %w", err)
		}

		if _, err = s.removeQuizFromQuizSets(ctx, tx, req.QuizId, quizSets); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	return &epb.RemoveQuizFromLOResponse{}, nil
}

func (s *QuizModifierService) validateRemoveQuizFromLORequest(req *epb.RemoveQuizFromLORequest) error {
	if req.LoId == "" {
		return fmt.Errorf("req must have LO id")
	}
	if req.QuizId == "" {
		return fmt.Errorf("req must have quiz id")
	}

	return nil
}

// UpdateDisplayOrderOfQuizSet in quiz set of lo, user update display order of quizzes in that list
func (s *QuizModifierService) UpdateDisplayOrderOfQuizSet(ctx context.Context, req *epb.UpdateDisplayOrderOfQuizSetRequest) (*epb.UpdateDisplayOrderOfQuizSetResponse, error) {
	quizSet, err := s.QuizSetRepo.GetQuizSetByLoID(ctx, s.DB, database.Text(req.LoId))
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "UpdateDisplayOrderOfQuizSet.QuizSetRepo.GetQuizSetByLoID: %v", err)
	}
	newQuizExternalIDList := make([]string, len(quizSet.QuizExternalIDs.Elements))
	mapIDIndex := make(map[string]int)

	for i, quizExternalID := range quizSet.QuizExternalIDs.Elements {
		newQuizExternalIDList[i] = quizExternalID.String
		mapIDIndex[newQuizExternalIDList[i]] = i
	}

	for _, quizExternalIDPair := range req.Pairs {
		i, ok := mapIDIndex[quizExternalIDPair.First]
		if !ok {
			return nil, status.Errorf(codes.InvalidArgument, "UpdateDisplayOrderOfQuizSet QuizExternalID %v is not exist in quiz set", quizExternalIDPair.First)
		}
		j, ok := mapIDIndex[quizExternalIDPair.Second]
		if !ok {
			return nil, status.Errorf(codes.InvalidArgument, "UpdateDisplayOrderOfQuizSet QuizExternalID %v is not exist in quiz set", quizExternalIDPair.Second)
		}
		newQuizExternalIDList[i], newQuizExternalIDList[j] = newQuizExternalIDList[j], newQuizExternalIDList[i]
		mapIDIndex[quizExternalIDPair.First], mapIDIndex[quizExternalIDPair.Second] = mapIDIndex[quizExternalIDPair.Second], mapIDIndex[quizExternalIDPair.First]
	}

	if err = database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		if err := s.QuizSetRepo.Delete(ctx, tx, quizSet.ID); err != nil {
			return fmt.Errorf("UpdateDisplayOrderOfQuizSet.QuizSetRepo.Delete: %v", err)
		}
		if err = quizSet.QuizExternalIDs.Set(newQuizExternalIDList); err != nil {
			return fmt.Errorf("failed to set newQuizExternalIDList: %v", err)
		}
		if err = s.QuizSetRepo.Create(ctx, tx, quizSet); err != nil {
			return fmt.Errorf("UpdateDisplayOrderOfQuizSet.QuizSetRepo.Create: %v", err)
		}
		return nil
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "ExecInTx err: %v", err)
	}

	return &epb.UpdateDisplayOrderOfQuizSetResponse{}, nil
}

// AppendQuestionToQuizSetByLoID will remove active quiz set
// which belong to a LO and clone a new record with
// QuizExternalIDs and QuestionHierarchy column be appended by way:
//   - quizExternalIDs is not null: append quizExternalIDs into QuizExternalIDs and QuestionHierarchy
//   - quizExternalIDs is null: append questionGroupIDs into QuestionHierarchy
//
// and return new quiz set id
//
//	TODO: replace duplicated code at func such as UpsertQuiz, UpsertSingleQuiz ... by using this func
func (s *QuizModifierService) AppendQuestionToQuizSetByLoID(ctx context.Context, db database.QueryExecer, loID string, quizExternalIDs []string, questionGroupIDs []string) (string, error) {
	set, err := s.getQuizSet(ctx, db, loID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return "", fmt.Errorf("getQuizSet: %w", err)
	}

	extIDs := make([]string, 0)
	questionHierarchy := make(entities.QuestionHierarchy, 0)
	// delete current active quiz set
	if set != nil {
		if err = s.QuizSetRepo.Delete(ctx, db, set.ID); err != nil {
			return "", fmt.Errorf("quizSetRepo.Delete: %w", err)
		}
		set.QuizExternalIDs.AssignTo(&extIDs)
		set.QuestionHierarchy.AssignTo(&questionHierarchy)
	} else {
		set = &entities.QuizSet{}
		database.AllNullEntity(set)
		// default value for QuizExternalID
		set.QuizExternalIDs.Set([]string{})
	}

	set.LoID = database.Text(loID)
	set.Status = database.Text(epb.QuizSetStatus_QUIZSET_STATUS_APPROVED.String())
	if len(quizExternalIDs) != 0 {
		extIDs = append(extIDs, quizExternalIDs...)
		set.QuizExternalIDs = database.TextArray(extIDs)

		// append quiz id into question_hierarchy column
		questionHierarchy.AddQuestionID(extIDs...)
		set.QuestionHierarchy.Set(questionHierarchy)
	} else {
		// append question group id into question_hierarchy column
		questionHierarchy.AddQuestionGroupID(questionGroupIDs...)
		set.QuestionHierarchy.Set(questionHierarchy)
	}

	if err = s.QuizSetRepo.Create(ctx, db, set); err != nil {
		return "", fmt.Errorf("quizSetRepo.Create: %w", err)
	}

	return set.ID.String, nil
}

func (s *QuizModifierService) generateShuffledQuizSetRandomSeed(ctx context.Context, shuffledQuizSet *entities.ShuffledQuizSet, loID string) {
	zapLogger := ctxzap.Extract(ctx).Sugar()
	quizzes, err := s.QuizRepo.GetByExternalIDsAndLmID(ctx, s.DB, shuffledQuizSet.QuizExternalIDs, database.Text(loID))
	if err != nil {
		zapLogger.Errorf("QuizRepo.GetByExternalIDsAndLmID: %v", err)
		return
	}

	allReSorted, err := shuffledQuizSet.GenerateRandomSeed(quizzes)
	if err != nil {
		zapLogger.Errorf("shuffledQuizSet.GenerateRandomSeed: %v", err)
		return
	}
	if !allReSorted {
		zapLogger.Errorf("shuffledQuizSet.GenerateRandomSeed: there are still some quiz in %v have right order answer", quizzes.GetExternalIDs())
	}
}
