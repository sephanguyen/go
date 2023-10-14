package services

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/yasuo/utils"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"github.com/golang/protobuf/ptypes"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// QuizzesPagination will return quizzes with a total number
type QuizzesPagination struct {
	Quizzes []*entities.Quiz
	Total   pgtype.Int8
}

// CourseModifierService implements bob proto ClassReaderServiceServer
type CourseModifierService struct {
	EurekaDBTrace database.Ext
	DB            database.Ext
	Env           string
	JSM           nats.JetStreamManagement

	QuizRepo interface {
		GetOptions(context.Context, database.QueryExecer, pgtype.Text, pgtype.Text) ([]*entities.QuizOption, error)
	}

	QuizSetRepo interface {
		GetQuizSetByLoID(context.Context, database.QueryExecer, pgtype.Text) (*entities.QuizSet, error)
		GetTotalQuiz(context.Context, database.QueryExecer, pgtype.TextArray) (map[string]int32, error)
	}
	ShuffledQuizSetRepo interface {
		GetSeed(context.Context, database.QueryExecer, pgtype.Text) (pgtype.Text, error)
		GetByStudyPlanItems(context.Context, database.QueryExecer, pgtype.TextArray) (entities.ShuffledQuizSets, error)
		GetSubmissionHistory(context.Context, database.QueryExecer, pgtype.Text, pgtype.Int4, pgtype.Int4) (map[pgtype.Text]pgtype.JSONB, []pgtype.Text, error)
		GetQuizIdx(context.Context, database.QueryExecer, pgtype.Text, pgtype.Text) (pgtype.Int4, error)
		Retrieve(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) ([]*entities.ShuffledQuizSet, error)
		ListExternalIDsFromSubmissionHistory(ctx context.Context, db database.QueryExecer, shuffleQuizSetIDs pgtype.TextArray, isAccepted bool) (map[string][]string, error)
	}

	ChapterRepo interface {
		DuplicateChapters(ctx context.Context, db database.QueryExecer, chapterIDs pgtype.TextArray) ([]*entities.CopiedChapter, error)
		FindByBookID(ctx context.Context, db database.QueryExecer, bookID string) ([]*entities.Chapter, error)
	}
	BookChapterRepo interface {
		FindByBookIDs(ctx context.Context, db database.QueryExecer, bookIDs []string) (map[string][]*entities.BookChapter, error)
		Upsert(ctx context.Context, db database.Ext, cc []*entities.BookChapter) error
		RetrieveContentStructuresByLOs(ctx context.Context, db database.QueryExecer, loIDs pgtype.TextArray) (map[string][]repositories.ContentStructure, error)
	}
	TopicRepo interface {
		DuplicateTopics(ctx context.Context, db database.QueryExecer, chapterIDs pgtype.TextArray, newChapterIDs pgtype.TextArray) ([]*entities.CopiedTopic, error)
		UpdateTotalLOs(ctx context.Context, db database.QueryExecer, topicID pgtype.Text) error
		RetrieveByIDs(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) ([]*entities.Topic, error)
		RetrieveByID(ctx context.Context, db database.QueryExecer, id pgtype.Text, enhancers ...repositories.QueryEnhancer) (*entities.Topic, error)
		UpdateLODisplayOrderCounter(ctx context.Context, db database.QueryExecer, topicID pgtype.Text, number pgtype.Int4) error
	}
	LearningObjectiveRepo interface {
		DuplicateLearningObjectives(ctx context.Context, db database.QueryExecer, topicIDs pgtype.TextArray, newTopicIDs pgtype.TextArray) ([]*entities.CopiedLearningObjective, error)
		RetrieveByIDs(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) ([]*entities.LearningObjective, error)
		RetrieveByTopicIDs(ctx context.Context, db database.QueryExecer, topicIds pgtype.TextArray) ([]*entities.LearningObjective, error)
		BulkImport(ctx context.Context, db database.QueryExecer, learningObjectives []*entities.LearningObjective) error
		UpdateDisplayOrders(ctx context.Context, db database.QueryExecer, mDisplayOrder map[pgtype.Text]pgtype.Int2) error
	}
	TopicsLearningObjectivesRepo interface {
		BulkImport(context.Context, database.QueryExecer, []*entities.TopicsLearningObjectives) error
		RetrieveByTopicIDs(ctx context.Context, db database.QueryExecer, topicIDs pgtype.TextArray) ([]*repositories.TopicLearningObjective, error)
		BulkUpdateDisplayOrder(
			ctx context.Context, db database.QueryExecer,
			topicsLearningsObjectives []*entities.TopicsLearningObjectives,
		) error
	}
	StudentEventLogRepo interface {
		RetrieveBySessions(context.Context, database.QueryExecer, pgtype.TextArray) (map[string][]*entities.StudentEventLog, error)
		RetrieveStudentEventLogsByStudyPlanItemIDs(context.Context, database.QueryExecer, pgtype.TextArray) ([]*entities.StudentEventLog, error)
	}
	StudentLearningTimeDailyRepo interface {
		Upsert(ctx context.Context, db database.QueryExecer, s *entities.StudentLearningTimeDaily) error
		Retrieve(ctx context.Context, db database.QueryExecer, studentID pgtype.Text, from, to *pgtype.Timestamptz, queryEnhancers ...repositories.QueryEnhancer) ([]*entities.StudentLearningTimeDaily, error)
	}
	StudentRepo interface {
		GetCountryByStudent(ctx context.Context, db database.QueryExecer, studentID pgtype.Text) (string, error)
	}

	EurekaCourseModifier interface {
		CompleteStudyPlanItem(ctx context.Context, req *epb.CompleteStudyPlanItemRequest, opts ...grpc.CallOption) (*epb.CompleteStudyPlanItemResponse, error)
	}
	EurekaQuizReaderSvc interface {
		RetrieveSubmissionHistory(ctx context.Context, in *epb.RetrieveSubmissionHistoryRequest, opts ...grpc.CallOption) (*epb.RetrieveSubmissionHistoryResponse, error)
	}
	EurekaLearningObjectiveModifier interface {
		UpsertLOs(ctx context.Context, in *epb.UpsertLOsRequest, opts ...grpc.CallOption) (*epb.UpsertLOsResponse, error)
	}
}

// NewCourseModifierService returns new *CourseModifierService
func NewCourseModifierService(
	eurekaDBTrace database.Ext,
	db database.Ext,
	jsm nats.JetStreamManagement,
	eurekaCourseModifierSvc epb.CourseModifierServiceClient,
	eurekaQuizReaderSvc epb.QuizReaderServiceClient,
	eurekaQuizModifierSvc epb.QuizModifierServiceClient,
	eurekaLearningObjectiveModifierSvc epb.LearningObjectiveModifierServiceClient,
	env string,
) *CourseModifierService {
	return &CourseModifierService{
		EurekaDBTrace:                eurekaDBTrace,
		DB:                           db,
		Env:                          env,
		JSM:                          jsm,
		QuizRepo:                     &repositories.QuizRepo{},
		QuizSetRepo:                  &repositories.QuizSetRepo{},
		ShuffledQuizSetRepo:          &repositories.ShuffledQuizSetRepo{},
		BookChapterRepo:              &repositories.BookChapterRepo{},
		ChapterRepo:                  &repositories.ChapterRepo{},
		TopicRepo:                    &repositories.TopicRepo{},
		LearningObjectiveRepo:        &repositories.LearningObjectiveRepo{},
		TopicsLearningObjectivesRepo: &repositories.TopicsLearningObjectivesRepo{},

		StudentEventLogRepo:             &repositories.StudentEventLogRepo{},
		StudentLearningTimeDailyRepo:    &repositories.StudentLearningTimeDailyRepo{},
		StudentRepo:                     &repositories.StudentRepo{},
		EurekaCourseModifier:            eurekaCourseModifierSvc,
		EurekaQuizReaderSvc:             eurekaQuizReaderSvc,
		EurekaLearningObjectiveModifier: eurekaLearningObjectiveModifierSvc,
	}
}

func isTopicsExisted(topicIDs []string, topics []*entities.Topic) bool {
	m := make(map[string]bool)
	for _, topic := range topics {
		m[topic.ID.String] = true
	}

	for _, id := range topicIDs {
		if _, ok := m[id]; !ok {
			return false
		}
	}
	return true
}

func isAutoGenLODisplayOrder(los entities.LearningObjectives) bool {
	for _, lo := range los {
		if lo.DisplayOrder.Int != 0 {
			return false
		}
	}
	return true
}

func (cm *CourseModifierService) updateTotalLOs(ctx context.Context, db database.QueryExecer, topicIDs []string) error {
	if len(topicIDs) == 0 {
		return nil
	}

	for _, topicID := range topicIDs {
		if err := cm.TopicRepo.UpdateTotalLOs(ctx, db, database.Text(topicID)); err != nil {
			return fmt.Errorf("cm.TopicRepo.UpdateTotalLOs: %v", err)
		}
	}
	return nil
}

func toListQuizpb(loID string, quizzes entities.Quizzes) ([]*cpb.Quiz, error) {
	pbQuizzes := []*cpb.Quiz{}
	for _, quiz := range quizzes {
		pbquiz, err := toQuizpb(loID, quiz)
		if err != nil {
			return nil, err
		}
		pbQuizzes = append(pbQuizzes, pbquiz)
	}
	return pbQuizzes, nil
}

func toQuizpb(loID string, quiz *entities.Quiz) (*cpb.Quiz, error) {
	question := &entities.RichText{}
	explanation := &entities.RichText{}
	answers := []*entities.QuizOption{}
	err := json.Unmarshal(quiz.Question.Bytes, question)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(quiz.Explanation.Bytes, explanation)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(quiz.Options.Bytes, &answers)
	if err != nil {
		return nil, err
	}

	answersURL := []string{}
	for _, ans := range answers {
		answersURL = append(answersURL, ans.Content.RenderedURL)
	}

	quizCore, err := toQuizCore(quiz)
	if err != nil {
		return nil, err
	}

	// to quiz pb
	pbquiz := &cpb.Quiz{
		Core:           quizCore,
		LoId:           loID,
		QuestionUrl:    question.RenderedURL,
		ExplanationUrl: explanation.RenderedURL,
		AnswersUrl:     answersURL,
		Status:         cpb.QuizStatus(cpb.QuizStatus_value[quiz.Status.String]),
	}

	return pbquiz, nil
}

func (cm *CourseModifierService) duplicateChapter(ctx context.Context, tx pgx.Tx, chapterIDs []string, bookID string) ([]string, []string, error) {
	copiedChapters, err := cm.ChapterRepo.DuplicateChapters(ctx, tx, database.TextArray(chapterIDs))
	if err != nil {
		return nil, nil, fmt.Errorf("cm.ChapterRepo.DuplicateChapters: %w", err)
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

// RetrieveSubmissionHistory returns history of a shuffled quizset with paging
func (cm *CourseModifierService) RetrieveSubmissionHistory(ctx context.Context, req *bpb.RetrieveSubmissionHistoryRequest) (*bpb.RetrieveSubmissionHistoryResponse, error) {
	mdctx, err := interceptors.GetOutgoingContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, fmt.Errorf("CourseReaderService.RetrieveTopics.GetOutgoingContext: %w", err).Error())
	}
	resp, err := cm.EurekaQuizReaderSvc.RetrieveSubmissionHistory(mdctx, &epb.RetrieveSubmissionHistoryRequest{
		SetId:  req.SetId,
		Paging: req.Paging,
	})
	if err != nil {
		return nil, err
	}
	return &bpb.RetrieveSubmissionHistoryResponse{
		Logs:     resp.Logs,
		NextPage: resp.NextPage,
	}, err
}

func toAnswerLogPb(submissionHistoryQuiz map[pgtype.Text]pgtype.JSONB, orderedQuizList []pgtype.Text) ([]*cpb.AnswerLog, error) {
	ansLogs := []*cpb.AnswerLog{}
	for _, quizID := range orderedQuizList {
		sub, ok := submissionHistoryQuiz[quizID]
		ansLog := &cpb.AnswerLog{}
		ansLog.QuizId = quizID.String
		if ok && sub.Status == pgtype.Present {
			ans := &entities.QuizAnswer{}
			err := sub.AssignTo(ans)
			if err != nil {
				return nil, err
			}

			ansLog.QuizType = cpb.QuizType(cpb.QuizType_value[ans.QuizType])
			ansLog.SelectedIndex = ans.SelectedIndex
			ansLog.CorrectIndex = ans.CorrectIndex
			ansLog.FilledText = ans.FilledText
			ansLog.CorrectText = ans.CorrectText
			ansLog.Correctness = ans.Correctness
			ansLog.IsAccepted = ans.IsAccepted
			ansLog.SubmittedAt, _ = ptypes.TimestampProto(ans.SubmittedAt)
		}
		ansLogs = append(ansLogs, ansLog)
	}
	return ansLogs, nil
}

func toQuizCore(quiz *entities.Quiz) (*cpb.QuizCore, error) {
	updatedAt, _ := ptypes.TimestampProto(quiz.UpdatedAt.Time)
	createdAt, _ := ptypes.TimestampProto(quiz.CreatedAt.Time)
	deletedAt, _ := ptypes.TimestampProto(quiz.DeletedAt.Time)

	pbInfo := &cpb.ContentBasicInfo{
		Id:        quiz.ID.String,
		Country:   cpb.Country(cpb.Country_value[quiz.Country.String]),
		SchoolId:  quiz.SchoolID.Int,
		UpdatedAt: updatedAt,
		CreatedAt: createdAt,
		DeletedAt: deletedAt,
	}
	question, err := quiz.GetQuestionV2()
	if err != nil {
		return nil, err
	}
	explanation := &entities.RichText{}
	err = quiz.Explanation.AssignTo(explanation)
	if err != nil {
		return nil, err
	}
	options, err := quiz.GetOptions()
	if err != nil {
		return nil, err
	}

	optionsPb := []*cpb.QuizOption{}
	for _, opt := range options {
		configOptions := []cpb.QuizOptionConfig{}
		for _, config := range opt.Configs {
			configOptions = append(configOptions, cpb.QuizOptionConfig(cpb.QuizOptionConfig_value[config]))
		}

		attCfgs := make([]cpb.QuizItemAttributeConfig, 0)
		for _, item := range opt.Attribute.Configs {
			attCfgs = append(attCfgs, cpb.QuizItemAttributeConfig(cpb.QuizItemAttributeConfig_value[item]))
		}

		optionsPb = append(optionsPb, &cpb.QuizOption{
			Content:     &cpb.RichText{Raw: opt.Content.Raw, Rendered: opt.Content.RenderedURL},
			Correctness: opt.Correctness,
			Configs:     configOptions,
			Label:       opt.Label,
			Key:         opt.Key,
			Attribute: &cpb.QuizItemAttribute{
				ImgLink:   opt.Attribute.ImgLink,
				AudioLink: opt.Attribute.AudioLink,
				Configs:   attCfgs,
			},
		})
	}

	cfgs := make([]cpb.QuizItemAttributeConfig, 0)
	for _, item := range question.Attribute.Configs {
		cfgs = append(cfgs, cpb.QuizItemAttributeConfig(cpb.QuizItemAttributeConfig_value[item]))
	}

	core := &cpb.QuizCore{
		Info:       pbInfo,
		ExternalId: quiz.ExternalID.String,
		Kind:       cpb.QuizType(cpb.QuizType_value[quiz.Kind.String]),
		Question:   &cpb.RichText{Raw: question.Raw, Rendered: question.RenderedURL},
		Attribute: &cpb.QuizItemAttribute{
			ImgLink:   question.Attribute.ImgLink,
			AudioLink: question.Attribute.AudioLink,
			Configs:   cfgs,
		},
		Explanation:     &cpb.RichText{Raw: explanation.Raw, Rendered: explanation.RenderedURL},
		DifficultyLevel: quiz.DifficultLevel.Int,
		TaggedLos:       database.FromTextArray(quiz.TaggedLOs),
		Options:         optionsPb,
	}

	return core, nil
}

type MultipleChoiceQuiz struct {
	*entities.Quiz
	SetID string

	ShuffledQuizSetRepo interface {
		GetSeed(context.Context, database.QueryExecer, pgtype.Text) (pgtype.Text, error)
		GetQuizIdx(context.Context, database.QueryExecer, pgtype.Text, pgtype.Text) (pgtype.Int4, error)
	}
}

func (q *MultipleChoiceQuiz) CheckCorrectness(ctx context.Context, db database.QueryExecer, userAnswer []*bpb.Answer) (*entities.QuizAnswer, error) {
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
		_, selectedIndexFormat := uAns.Format.(*bpb.Answer_SelectedIndex)
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

	answer := &entities.QuizAnswer{
		QuizID:        q.ExternalID.String,
		QuizType:      q.Kind.String,
		SelectedIndex: selectedIdxs,
		CorrectIndex:  correctIdx,
		Correctness:   crn,
		IsAccepted:    isAccepted,
		IsAllCorrect:  allCorrect,
		SubmittedAt:   time.Now(),
	}
	return answer, nil
}

type ManualInputQuiz struct {
	*entities.Quiz
}

func (q *ManualInputQuiz) CheckCorrectness(ctx context.Context, db database.Ext, userAnswer []*bpb.Answer) (*entities.QuizAnswer, error) {
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
		_, selectedIndexFormat := uAns.Format.(*bpb.Answer_SelectedIndex)
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

	answer := &entities.QuizAnswer{
		QuizID:        q.ExternalID.String,
		QuizType:      q.Kind.String,
		SelectedIndex: selectedIdxs,
		CorrectIndex:  correctIdx,
		Correctness:   crn,
		IsAccepted:    isAccepted,
		IsAllCorrect:  allCorrect,
		SubmittedAt:   time.Now(),
	}
	return answer, nil
}

type FillInTheBlankQuiz struct {
	*entities.Quiz
}

func (q *FillInTheBlankQuiz) CheckCorrectness(ctx context.Context, db database.Ext, userAnswer []*bpb.Answer) (*entities.QuizAnswer, error) {
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
		_, filledTextFormat := uAns.Format.(*bpb.Answer_FilledText)
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

	answer := &entities.QuizAnswer{
		QuizID:       q.ExternalID.String,
		QuizType:     q.Kind.String,
		FilledText:   filledText,
		CorrectText:  correctText,
		Correctness:  crn,
		IsAccepted:   isAccepted,
		IsAllCorrect: allCorrect,
		SubmittedAt:  time.Now(),
	}
	return answer, nil
}

// GetFlashcardProgressionWithPaging
// -- get flashcard by study_set_id, student_id, from, to
// -- get quizzes by externalIDs from flashcardProgression
// -- convert quizzes to flashcardQuizzes
func getFlashcardProgressionWithPaging(
	ctx context.Context, db database.Ext,
	flashcardProgressionRepo interface {
		Get(ctx context.Context, db database.QueryExecer, args *repositories.GetFlashcardProgressionArgs) (*entities.FlashcardProgression, error)
	},
	quizRepo interface {
		GetByExternalIDs(context.Context, database.QueryExecer, pgtype.TextArray, pgtype.Text) (entities.Quizzes, error)
	},
	paging *cpb.Paging,
	studySetID, studentID pgtype.Text,
) ([]*bpb.FlashcardQuizzes, *entities.FlashcardProgression, error) {
	offset := paging.GetOffsetInteger()
	limit := paging.Limit
	from := database.Int8(offset)
	to := database.Int8(offset + int64(limit) - 1)
	pagingFlashcardProgression, err := flashcardProgressionRepo.Get(ctx, db, &repositories.GetFlashcardProgressionArgs{
		StudySetID:      studySetID,
		StudentID:       studentID,
		LoID:            pgtype.Text{Status: pgtype.Null},
		StudyPlanItemID: pgtype.Text{Status: pgtype.Null},
		From:            from,
		To:              to,
	})
	if err != nil {
		return nil, nil, status.Errorf(codes.Internal, "getFlashcardProgressionWithPaging.FlashcardProgressionRepo.Get: %v", err)
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
	quizzes, err := quizRepo.GetByExternalIDs(ctx, db, pagingFlashcardProgression.QuizExternalIDs, database.Text(pagingFlashcardProgression.LoID.String))
	if err != nil {
		return nil, nil, status.Errorf(codes.Internal, "getFlashcardProgressionWithPaging.QuizRepo.GetByExternalIDs: %v", err)
	}

	// Convert from entities.Quizzes to []bpb.FlashcardProgression
	pbQuizzes, err := toListFlashcardQuizzes(pagingFlashcardProgression.LoID.String, quizzes, skippedQuestionIDsMap, rememberedQuestionIDsMap)
	if err != nil {
		return nil, nil, status.Errorf(codes.Internal, "getFlashcardProgressionWithPaging.toListFlashcardQuizzes: %v", err)
	}
	return pbQuizzes, pagingFlashcardProgression, nil
}

func toListFlashcardQuizzes(
	loID string, quizzes entities.Quizzes,
	skippedQuestionIDsMap, rememberedQuestionIDsMap map[string]int64,
) ([]*bpb.FlashcardQuizzes, error) {
	pbQuizzes := []*bpb.FlashcardQuizzes{}
	for _, quiz := range quizzes {
		pbQuiz, err := toQuizpb(loID, quiz)
		if err != nil {
			return nil, err
		}
		flashcardQuiz := &bpb.FlashcardQuizzes{
			Item:   pbQuiz,
			Status: bpb.FlashcardQuizStudyStatus_FLASHCARD_QUIZ_STUDY_STATUS_NONE,
		}
		if _, ok := skippedQuestionIDsMap[quiz.ExternalID.String]; ok {
			flashcardQuiz.Status = bpb.FlashcardQuizStudyStatus_FLASHCARD_QUIZ_STUDY_STATUS_SKIPPED
		}
		if _, ok := rememberedQuestionIDsMap[quiz.ExternalID.String]; ok {
			flashcardQuiz.Status = bpb.FlashcardQuizStudyStatus_FLASHCARD_QUIZ_STUDY_STATUS_REMEMBERED
		}
		pbQuizzes = append(pbQuizzes, flashcardQuiz)
	}
	return pbQuizzes, nil
}
