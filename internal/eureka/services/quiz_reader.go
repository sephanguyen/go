package services

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"github.com/jackc/pgtype"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type QuizReaderService struct {
	DB database.Ext

	QuizSetRepo interface {
		GetTotalQuiz(ctx context.Context, db database.QueryExecer, loIDs pgtype.TextArray) (map[string]int32, error)
		GetQuizExternalIDs(ctx context.Context, db database.QueryExecer, loID pgtype.Text, limit pgtype.Int8, offset pgtype.Int8) ([]string, error)
	}
	ShuffledQuizSetRepo interface {
		ListExternalIDsFromSubmissionHistory(ctx context.Context, db database.QueryExecer, shuffleQuizSetIDs pgtype.TextArray, isAccepted bool) (map[string][]string, error)
		GetSeed(context.Context, database.QueryExecer, pgtype.Text) (pgtype.Text, error)
		GetByStudyPlanItems(context.Context, database.QueryExecer, pgtype.TextArray) (entities.ShuffledQuizSets, error)
		GetSubmissionHistory(context.Context, database.QueryExecer, pgtype.Text, pgtype.Int4, pgtype.Int4) (map[pgtype.Text]pgtype.JSONB, []pgtype.Text, error)
		GetLoID(context.Context, database.QueryExecer, pgtype.Text) (pgtype.Text, error)
		GetQuizIdx(context.Context, database.QueryExecer, pgtype.Text, pgtype.Text) (pgtype.Int4, error)
	}
	StudentEventLogRepo interface {
		RetrieveStudentEventLogsByStudyPlanItemIDs(context.Context, database.QueryExecer, pgtype.TextArray) ([]*entities.StudentEventLog, error)
	}
	LearningTimeCalculatorSvc interface {
		Calculate([]*entities.StudentEventLog) (learningTime time.Duration, completedAt *time.Time, err error)
	}
	QuizRepo interface {
		GetByExternalIDs(context.Context, database.QueryExecer, pgtype.TextArray, pgtype.Text) (entities.Quizzes, error)
	}
	QuestionGroupRepo interface {
		GetQuestionGroupsByIDs(ctx context.Context, db database.QueryExecer, ids ...string) (entities.QuestionGroups, error)
	}
}

func NewQuizReaderService(db database.Ext) *QuizReaderService {
	return &QuizReaderService{
		DB:                        db,
		QuizSetRepo:               &repositories.QuizSetRepo{},
		QuizRepo:                  &repositories.QuizRepo{},
		ShuffledQuizSetRepo:       &repositories.ShuffledQuizSetRepo{},
		StudentEventLogRepo:       &repositories.StudentEventLogRepo{},
		LearningTimeCalculatorSvc: &LearningTimeCalculator{},
		QuestionGroupRepo:         &repositories.QuestionGroupRepo{},
	}
}

func (s *QuizReaderService) RetrieveTotalQuizLOs(ctx context.Context, req *epb.RetrieveTotalQuizLOsRequest) (*epb.RetrieveTotalQuizLOsResponse, error) {
	err := s.validateRetrieveTotalQuizLOsRequest(req)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	totalQuizLOsMap, err := s.QuizSetRepo.GetTotalQuiz(ctx, s.DB, database.TextArray(req.LoIds))
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("s.QuizSetRepo.GetTotalQuiz: %w", err).Error())
	}

	losTotalQuiz := make([]*epb.RetrieveTotalQuizLOsResponse_LoWithTotalQuiz, 0, len(totalQuizLOsMap))

	for _, loID := range req.LoIds {
		losTotalQuiz = append(losTotalQuiz, &epb.RetrieveTotalQuizLOsResponse_LoWithTotalQuiz{
			LoId:      loID,
			TotalQuiz: totalQuizLOsMap[loID],
		})
	}

	resp := &epb.RetrieveTotalQuizLOsResponse{
		LosTotalQuiz: losTotalQuiz,
	}
	return resp, nil
}

// RetrieveSubmissionHistory returns history of a shuffled quizset with paging
func (s *QuizReaderService) RetrieveSubmissionHistory(ctx context.Context, req *epb.RetrieveSubmissionHistoryRequest) (*epb.RetrieveSubmissionHistoryResponse, error) {
	if req.Paging == nil {
		req.Paging = &cpb.Paging{
			Limit: 100,
			Offset: &cpb.Paging_OffsetInteger{
				OffsetInteger: 0,
			},
		}
	}
	limit := req.Paging.Limit
	offset := req.Paging.GetOffsetInteger()

	loID, err := s.ShuffledQuizSetRepo.GetLoID(ctx, s.DB, database.Text(req.SetId))
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Errorf("CourseModifierService.ShuffledQuizSetRepo.GetLoID: %v", err).Error())
	}
	submissionHistory, listQuizIDs, err := s.ShuffledQuizSetRepo.GetSubmissionHistory(ctx, s.DB, database.Text(req.SetId), database.Int4(int32(limit)), database.Int4(int32(offset)))
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Errorf("CourseModifierService.RetrieveSubmissionHistory.GetSubmissionHistory: %v", err).Error())
	}

	ansLogs, err := toAnswerLogPb(submissionHistory, listQuizIDs)
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Errorf("CourseModifierService.RetrieveSubmissionHistory.toAnswerLogPb: %v", err).Error())
	}

	quizIDs := make([]string, 0, len(listQuizIDs))
	for i := range listQuizIDs {
		quizIDs = append(quizIDs, listQuizIDs[i].String)
	}

	quizzesMap := make(map[string]*entities.Quiz)
	if len(ansLogs) != 0 {
		quizzesWrap, err := s.QuizRepo.GetByExternalIDs(ctx, s.DB, database.TextArray(quizIDs), loID)
		if err != nil {
			return nil, status.Errorf(codes.Internal, fmt.Errorf("CourseModifierService.RetrieveSubmissionHistory.QuizRepo.GetByExternalIDs: %v", err).Error())
		}

		if len(quizzesWrap) == 0 {
			return nil, status.Errorf(codes.Internal, fmt.Errorf("CourseModifierService.RetrieveSubmissionHistory.GetByExternalIDs: cannot find quiz %v", quizIDs).Error())
		}

		for _, quiz := range quizzesWrap {
			quizzesMap[quiz.ExternalID.String] = quiz
		}
	}

	for _, log := range ansLogs {
		quiz := quizzesMap[log.QuizId]
		core, err := toQuizCore(quiz)
		if err != nil {
			return nil, status.Errorf(codes.Internal, fmt.Errorf("CourseModifierService.RetrieveSubmissionHistory.toQuizCore: %v", err).Error())
		}
		originalOption := make([]*cpb.QuizOption, len(core.Options))
		copy(originalOption, core.Options)
		if core.Kind == cpb.QuizType_QUIZ_TYPE_MCQ ||
			core.Kind == cpb.QuizType_QUIZ_TYPE_MAQ ||
			core.Kind == cpb.QuizType_QUIZ_TYPE_ORD {
			// only shuffle options of multiple choice quiz
			seedStr, err := s.ShuffledQuizSetRepo.GetSeed(ctx, s.DB, database.Text(req.SetId))
			if err != nil {
				return nil, status.Errorf(codes.Internal, fmt.Errorf("CourseModifierService.RetrieveSubmissionHistory.ShuffledQuizSetRepo.GetSeed: %v", err).Error())
			}
			seed, err := strconv.ParseInt(seedStr.String, 10, 64)
			if err != nil {
				return nil, status.Errorf(codes.Internal, fmt.Errorf("CourseModifierService.RetrieveSubmissionHistory.CannotParseSeed: %v", err).Error())
			}
			idx, err := s.ShuffledQuizSetRepo.GetQuizIdx(ctx, s.DB, database.Text(req.SetId), quiz.ExternalID)
			if err != nil || idx.Int == 0 {
				return nil, status.Errorf(
					codes.Internal,
					fmt.Errorf(
						"CourseModifierService.RetrieveSubmissionHistory.GetQuizIdx: could not found quiz id %s in shuffled quiz set %s : %v",
						quiz.ExternalID.String,
						req.SetId,
						err,
					).Error())
			}
			r := rand.New(rand.NewSource(seed + int64(idx.Int)))
			r.Shuffle(len(core.Options), func(i, j int) { core.Options[i], core.Options[j] = core.Options[j], core.Options[i] })
		}
		log.Core = core
		log.QuizType = core.Kind

		if log.SubmittedAt == nil {
			// this is uncompleted quiz
			// need to update correct index and correct filled text

			correctIdx := make([]uint32, 0, len(core.Options))
			correctText := make([]string, 0, len(core.Options))
			switch core.Kind {
			case cpb.QuizType_QUIZ_TYPE_MCQ, cpb.QuizType_QUIZ_TYPE_MAQ, cpb.QuizType_QUIZ_TYPE_MIQ:
				for idx, opt := range core.Options {
					if opt.Correctness {
						correctIdx = append(correctIdx, uint32(idx))
					}
				}
			case cpb.QuizType_QUIZ_TYPE_FIB, cpb.QuizType_QUIZ_TYPE_POW, cpb.QuizType_QUIZ_TYPE_TAD:
				var optionsEnt []*entities.QuizOption
				if err = quiz.Options.AssignTo(&optionsEnt); err != nil {
					return nil, status.Errorf(codes.Internal, fmt.Errorf("could not convert options jsonb to QuizOption struct: %v", err).Error())
				}

				for _, opt := range optionsEnt {
					correctText = append(correctText, strings.TrimSpace(opt.GetText()))
				}
			case cpb.QuizType_QUIZ_TYPE_ORD:
				correctKeys := make([]string, 0, len(core.Options))
				for _, opt := range originalOption {
					correctKeys = append(correctKeys, opt.Key)
				}
				log.Result = &cpb.AnswerLog_OrderingResult{
					OrderingResult: &cpb.OrderingResult{
						CorrectKeys: correctKeys,
					},
				}
			case cpb.QuizType_QUIZ_TYPE_ESQ:
				continue
			default:
				return nil, status.Error(codes.FailedPrecondition, "Not supported quiz type!")
			}

			log.CorrectIndex = correctIdx
			log.CorrectText = correctText
		}
	}

	// get list question group
	questionGrIDs := make([]string, 0)
	for _, quiz := range quizzesMap {
		if quiz.QuestionGroupID.Status == pgtype.Present && len(quiz.QuestionGroupID.String) != 0 {
			questionGrIDs = append(questionGrIDs, quiz.QuestionGroupID.String)
		}
	}

	var questionGroup entities.QuestionGroups
	if len(questionGrIDs) != 0 {
		questionGroup, err = s.QuestionGroupRepo.GetQuestionGroupsByIDs(ctx, s.DB, questionGrIDs...)
		if err != nil {
			return nil, status.Errorf(codes.Internal, fmt.Errorf("QuestionGroupRepo.GetQuestionGroupsByIDs: %v", err).Error())
		}
	}

	respQuestionGroups, err := entities.QuestionGroupsToQuestionGroupProtoBufMess(questionGroup)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	resp := &epb.RetrieveSubmissionHistoryResponse{
		Logs: ansLogs,
		NextPage: &cpb.Paging{
			Limit: limit,
			Offset: &cpb.Paging_OffsetInteger{
				OffsetInteger: int64(limit) + offset,
			},
		},
		QuestionGroups: respQuestionGroups,
	}
	return resp, nil
}

func (s *QuizReaderService) validateRetrieveTotalQuizLOsRequest(req *epb.RetrieveTotalQuizLOsRequest) error {
	if len(req.LoIds) == 0 {
		return errors.New("req must have lo ids")
	}
	return nil
}

// QuizzesPagination will return quizzes with a total number
type QuizzesPagination struct {
	Quizzes []*entities.Quiz
	Total   pgtype.Int8
}

func (s *QuizReaderService) validateRetrieveQuizTets(req *epb.RetrieveQuizTestsRequest) error {
	if len(req.StudyPlanItemId) == 0 {
		return errors.New("req must have Study Plan Item Id")
	}

	return nil
}

// fetch the logs chunk by chunk to avoid Seq Scan
// because of too large amount of records will be over the effective_cache_size
// Postgres will choose Seq Scan rather than Index Scan
func (s *QuizReaderService) retrieveStudentEventLogsConcurrently(ctx context.Context, studyPlanItemId []string) ([]*entities.StudentEventLog, error) {
	studentEventLogs := make([]*entities.StudentEventLog, 0)
	chunkSize := 50
	numberOfRoutines := int(math.Ceil(float64(len(studyPlanItemId)) / float64(chunkSize)))
	var wg sync.WaitGroup
	wg.Add(numberOfRoutines)
	cLogs := make(chan []*entities.StudentEventLog, numberOfRoutines)
	cErrs := make(chan error, numberOfRoutines)
	defer func() {
		close(cLogs)
		close(cErrs)
	}()
	for i := 0; i < len(studyPlanItemId); i += chunkSize {
		go func(i int) {
			defer wg.Done()
			t := int(math.Min(float64(i+chunkSize), float64(len(studyPlanItemId))))
			logs, err := s.StudentEventLogRepo.RetrieveStudentEventLogsByStudyPlanItemIDs(ctx, s.DB, database.TextArray(studyPlanItemId[i:t]))
			if err != nil {
				cErrs <- status.Error(codes.Internal, fmt.Errorf("s.StudentEventLogRepo.RetrieveStudentEventLogsByStudyPlanItemIDs: %v", err).Error())
				return
			}
			cLogs <- logs
		}(i)
	}
	wg.Wait()
	for i := 0; i < numberOfRoutines; i++ {
		select {
		case logs := <-cLogs:
			studentEventLogs = append(studentEventLogs, logs...)

		case err := <-cErrs:
			return nil, err
		}
	}
	sort.Slice(studentEventLogs, func(i, j int) bool {
		return studentEventLogs[i].CreatedAt.Time.Before(studentEventLogs[j].CreatedAt.Time)
	})
	return studentEventLogs, nil
}

// RetrieveQuizTests return list of quiz test infor
func (s *QuizReaderService) RetrieveQuizTests(ctx context.Context, req *epb.RetrieveQuizTestsRequest) (*epb.RetrieveQuizTestsResponse, error) {
	if err := s.validateRetrieveQuizTets(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// list shuffledQuizTests by studyPlanItemIDs
	quizTests, err := s.ShuffledQuizSetRepo.GetByStudyPlanItems(ctx, s.DB, database.TextArray(req.StudyPlanItemId))
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("s.ShuffledQuizSetRepo.GetByStudyPlanItems: %v", err).Error())
	}

	// list studentEventLogs by studyPlanItemIDs
	// group logs by studyPlanItemID and SessionID
	studentEventLogsMap := make(map[string]map[string][]*entities.StudentEventLog)
	for _, studyPlanItemID := range req.StudyPlanItemId {
		studentEventLogsMap[studyPlanItemID] = make(map[string][]*entities.StudentEventLog)
	}

	studentEventLogs, err := s.retrieveStudentEventLogsConcurrently(ctx, req.StudyPlanItemId)
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
		studyPlanItemID, ok := payload["study_plan_item_id"].(string)
		if !ok {
			continue
		}
		studentEventLogsMap[studyPlanItemID][sessionID] = append(studentEventLogsMap[studyPlanItemID][sessionID], studentEventLog)
	}

	// list map shuffledQuizSetID and externalQuizIDs -> externalIDsFromSubmissionHistoriesMap
	mapQuizTests := make(map[string][]*entities.ShuffledQuizSet)
	var originalShuffledQuizSetIDs []string
	originalShuffledQuizSetIDMap := make(map[string]bool)
	for _, quizTest := range quizTests {
		mapQuizTests[quizTest.StudyPlanItemID.String] = append(mapQuizTests[quizTest.StudyPlanItemID.String], quizTest)

		if quizTest.OriginalShuffleQuizSetID.Status != pgtype.Present {
			continue
		}

		if _, ok := originalShuffledQuizSetIDMap[quizTest.OriginalShuffleQuizSetID.String]; !ok {
			originalShuffledQuizSetIDMap[quizTest.OriginalShuffleQuizSetID.String] = true
			originalShuffledQuizSetIDs = append(originalShuffledQuizSetIDs, quizTest.OriginalShuffleQuizSetID.String)
		}
	}

	externalIDsFromSubmissionHistoriesMap := make(map[string][]string)
	if len(originalShuffledQuizSetIDs) != 0 {
		externalIDsFromSubmissionHistoriesMap, err = s.ShuffledQuizSetRepo.ListExternalIDsFromSubmissionHistory(ctx, s.DB, database.TextArray(originalShuffledQuizSetIDs), false)
		if err != nil {
			return nil, status.Error(codes.Internal, fmt.Errorf("RetrieveQuizTests.ShuffledQuizSetRepo.ListExternalIDsFromSubmissionHistory: %w", err).Error())
		}
	}

	// calculate learning time, score and crown
	mItems := make(map[string]*cpb.QuizTests)
	var totalAttempt int32
	highestScore := &cpb.HighestQuizScore{}

	for studyPlanItemID := range mapQuizTests {
		testInfors := []*cpb.QuizTestInfo{}
		for _, test := range mapQuizTests[studyPlanItemID] {
			logs := studentEventLogsMap[studyPlanItemID][test.SessionID.String]
			learningTime, completedAt, err := s.LearningTimeCalculatorSvc.Calculate(logs)
			if err != nil {
				return nil, status.Error(codes.Internal, fmt.Errorf("QuizReaderService.LearningTimeCalculatorSvc.Calculate: %v", err).Error())
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
			testInfors = append(testInfors, infor)
		}
		// sort the item decs completed_at
		sortQuizTestInfo(testInfors)
		quizTestsPb := &cpb.QuizTests{
			Items: testInfors,
		}
		mItems[studyPlanItemID] = quizTestsPb
	}

	resp := &epb.RetrieveQuizTestsResponse{
		Items:         mItems,
		HighestCrown:  getCrown(highestScore.GetCorrectQuestion(), highestScore.GetTotalQuestion()),
		HighestScore:  highestScore,
		TotalAttempts: totalAttempt,
	}

	return resp, nil
}

func getMaxQuizScore(curScore *cpb.HighestQuizScore, totalCorrectness, totalQuiz int32) {
	if totalQuiz == 0 {
		return
	}
	if (curScore.GetTotalQuestion() == 0 && totalQuiz != 0) || float32(curScore.GetCorrectQuestion())/float32(curScore.GetTotalQuestion()) < float32(totalCorrectness)/float32(totalQuiz) {
		curScore.CorrectQuestion = totalCorrectness
		curScore.TotalQuestion = totalQuiz
	}
}

func getCrown(correctQuestion, totalQuestion int32) epb.AchievementCrown {
	score := (float32(correctQuestion) / float32(totalQuestion)) * 100
	switch {
	case score == 100:
		return epb.AchievementCrown_ACHIEVEMENT_CROWN_GOLD
	case score >= 80:
		return epb.AchievementCrown_ACHIEVEMENT_CROWN_SILVER
	case score >= 60:
		return epb.AchievementCrown_ACHIEVEMENT_CROWN_BRONZE
	default:
		return epb.AchievementCrown_ACHIEVEMENT_CROWN_NONE
	}
}

// sortQuizTestInfo will sort by completed_at, if two value both nil, will use created_at
func sortQuizTestInfo(testInfors []*cpb.QuizTestInfo) {
	sort.Slice(testInfors, func(i, j int) bool {
		if testInfors[i].CompletedAt == nil && testInfors[j].CompletedAt == nil {
			return testInfors[i].CreatedAt.AsTime().After(testInfors[j].CreatedAt.AsTime())
		}
		return testInfors[i].CompletedAt.AsTime().After(testInfors[j].CompletedAt.AsTime())
	})
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
			ansLog.SubmittedAt = timestamppb.New(ans.SubmittedAt)

			if len(ans.CorrectKeys) != 0 {
				ansLog.Result = &cpb.AnswerLog_OrderingResult{
					OrderingResult: &cpb.OrderingResult{
						SubmittedKeys: ans.SubmittedKeys,
						CorrectKeys:   ans.CorrectKeys,
					},
				}
			}
		}
		ansLogs = append(ansLogs, ansLog)
	}
	return ansLogs, nil
}

// ListQuizzesOfLO list quizzes of a learning objective
func (s *QuizReaderService) ListQuizzesOfLO(ctx context.Context, req *epb.ListQuizzesOfLORequest) (*epb.ListQuizzesOfLOResponse, error) {
	limit := database.Int8(100)
	offset := database.Int8(0)
	if req.Paging != nil && req.Paging.Limit != 0 {
		_ = limit.Set(req.Paging.Limit)
		_ = offset.Set(req.Paging.GetOffsetInteger())
	}
	quizExternalIDs, err := s.QuizSetRepo.GetQuizExternalIDs(ctx, s.DB, database.Text(req.LoId), limit, offset)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "ListQuizzesOfLO.QuizSetRepo.GetQuizSetByLoID: %v", err.Error())
	}

	quizzes, err := s.QuizRepo.GetByExternalIDs(ctx, s.DB, database.TextArray(quizExternalIDs), database.Text(req.LoId))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "ListQuizzesOfLO.QuizRepo.GetByExternalIDs: %v", err.Error())
	}
	quizzesMap := make(map[string]*entities.Quiz)
	for _, quiz := range quizzes {
		quizzesMap[quiz.ExternalID.String] = quiz
	}

	ansLogs := make([]*cpb.AnswerLog, 0, len(quizzes))
	for _, quiz := range quizzes {
		log := &cpb.AnswerLog{
			QuizId:   quiz.ExternalID.String,
			QuizType: cpb.QuizType(cpb.QuizType_value[quiz.Kind.String]),
		}
		core, err := toQuizCore(quiz)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "ListQuizzesOfLO.toQuizCore: %v", err.Error())
		}
		log.Core = core

		correctIdx := make([]uint32, 0, len(core.Options))
		correctText := make([]string, 0, len(core.Options))
		switch core.Kind {
		case cpb.QuizType_QUIZ_TYPE_MCQ, cpb.QuizType_QUIZ_TYPE_MAQ, cpb.QuizType_QUIZ_TYPE_MIQ:
			for idx, opt := range core.Options {
				if opt.Correctness {
					correctIdx = append(correctIdx, uint32(idx))
				}
			}
		case cpb.QuizType_QUIZ_TYPE_FIB, cpb.QuizType_QUIZ_TYPE_POW, cpb.QuizType_QUIZ_TYPE_TAD:
			// TODO: make this logic to consistent when checkCorrectnessQuizzes later
			// https://github.com/manabie-com/backend/blob/191c2351c1cbc13cfffba972830b70b834631a3b/internal/eureka/services/quiz_modifier.go#L1492
			options, err := quiz.GetOptions()
			if err != nil {
				return nil, status.Errorf(codes.Internal, "Quiz.GetOptions: %v", err.Error())
			}
			for _, opt := range options {
				correctText = append(correctText, opt.GetText())
			}
		case cpb.QuizType_QUIZ_TYPE_ORD:
			correctKeys := make([]string, 0, len(core.Options))
			for _, opt := range core.Options {
				correctKeys = append(correctKeys, opt.Key)
			}
			log.Result = &cpb.AnswerLog_OrderingResult{
				OrderingResult: &cpb.OrderingResult{
					CorrectKeys: correctKeys,
				},
			}
		case cpb.QuizType_QUIZ_TYPE_ESQ:
		default:
			return nil, status.Error(codes.FailedPrecondition, "Not supported quiz type!")
		}

		log.CorrectIndex = correctIdx
		log.CorrectText = correctText

		ansLogs = append(ansLogs, log)
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

	resp := &epb.ListQuizzesOfLOResponse{
		Logs: ansLogs,
		NextPage: &cpb.Paging{
			Limit: uint32(limit.Int),
			Offset: &cpb.Paging_OffsetInteger{
				OffsetInteger: limit.Int + offset.Int,
			},
		},
		QuestionGroups: respQuestionGroups,
	}
	return resp, nil
}
