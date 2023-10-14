package services

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	"github.com/manabie-com/backend/internal/eureka/entities"
	db "github.com/manabie-com/backend/internal/eureka/golibs/database"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type StudentSubmissionService struct {
	DB          database.Ext
	StudentRepo interface {
		FindStudentsByCourseID(context.Context, database.QueryExecer, pgtype.Text) (*pgtype.TextArray, error)
		FindStudentsByClassIDs(context.Context, database.QueryExecer, pgtype.TextArray) (*pgtype.TextArray, error)
		FindStudentsByCourseLocation(ctx context.Context, db database.QueryExecer, courseID pgtype.Text, locationID pgtype.TextArray) (*pgtype.TextArray, error)
		FindStudentsByLocation(ctx context.Context, db database.QueryExecer, locationIDs pgtype.TextArray) (*pgtype.TextArray, error)
	}
	StudyPlanRepo interface {
		FindByIDs(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) ([]*entities.StudyPlan, error)
	}
	StudentSubmissionRepo interface {
		ListV3(ctx context.Context, db database.QueryExecer, filter *repositories.StudentSubmissionFilter) ([]*repositories.StudentSubmissionInfo, error)
		ListV4(ctx context.Context, db database.QueryExecer, filter *repositories.StudentSubmissionFilter) ([]*repositories.StudentSubmissionInfo, error)
	}
	QuizRepo interface {
		GetByExternalIDs(context.Context, database.QueryExecer, pgtype.TextArray, pgtype.Text) (entities.Quizzes, error)
	}
	ShuffledQuizSetRepo interface {
		GetLoID(context.Context, database.QueryExecer, pgtype.Text) (pgtype.Text, error)
		GetRelatedLearningMaterial(context.Context, database.QueryExecer, pgtype.Text) (*entities.LearningMaterial, error)
		GetSubmissionHistory(context.Context, database.QueryExecer, pgtype.Text, pgtype.Int4, pgtype.Int4) (map[pgtype.Text]pgtype.JSONB, []pgtype.Text, error)
		GetSeed(context.Context, database.QueryExecer, pgtype.Text) (pgtype.Text, error)
		GetQuizIdx(context.Context, database.QueryExecer, pgtype.Text, pgtype.Text) (pgtype.Int4, error)
	}
	QuestionGroupRepo interface {
		GetQuestionGroupsByIDs(ctx context.Context, db database.QueryExecer, ids ...string) (entities.QuestionGroups, error)
	}
	LOSubmissionAnswerRepo interface {
		ListSubmissionAnswers(ctx context.Context, db database.QueryExecer, setID pgtype.Text, limit, offset pgtype.Int8) ([]*entities.LOSubmissionAnswer, []pgtype.Text, error)
	}
	FlashCardSubmissionAnswerRepo interface {
		ListSubmissionAnswers(ctx context.Context, db database.QueryExecer, setID pgtype.Text, limit, offset pgtype.Int8) ([]*entities.FlashCardSubmissionAnswer, []pgtype.Text, error)
	}
}

func NewStudentSubmissionService(db database.Ext) sspb.StudentSubmissionServiceServer {
	return &StudentSubmissionService{
		DB:                            db,
		StudentRepo:                   &repositories.StudentRepo{},
		StudyPlanRepo:                 &repositories.StudyPlanRepo{},
		StudentSubmissionRepo:         &repositories.StudentSubmissionRepo{},
		QuizRepo:                      &repositories.QuizRepo{},
		ShuffledQuizSetRepo:           &repositories.ShuffledQuizSetRepo{},
		QuestionGroupRepo:             &repositories.QuestionGroupRepo{},
		LOSubmissionAnswerRepo:        &repositories.LOSubmissionAnswerRepo{},
		FlashCardSubmissionAnswerRepo: &repositories.FlashCardSubmissionAnswerRepo{},
	}
}

func (s *StudentSubmissionService) ListSubmissionsV3(ctx context.Context, req *sspb.ListSubmissionsV3Request) (*sspb.ListSubmissionsV3Response, error) {
	if req.GetPaging() == nil {
		return nil, status.Error(codes.InvalidArgument, "empty paging")
	}

	if req.Paging.Limit < 1 || req.Paging.Limit > 100 {
		return nil, status.Error(codes.InvalidArgument, "limit must in range 1 to 100")
	}
	filter := &repositories.StudentSubmissionFilter{
		Limit: uint(req.Paging.Limit),
	}

	if err := multierr.Combine(
		filter.OffsetID.Set(nil),
		filter.CreatedAt.Set(nil),
		filter.StudentIDs.Set(nil),
		filter.Statuses.Set(nil),
		filter.StartDate.Set(nil),
		filter.EndDate.Set(nil),
		filter.AssignmentName.Set(nil),
		filter.CourseID.Set(nil),
		filter.ClassIDs.Set(nil),
		filter.LocationIDs.Set(req.LocationIds),
	); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Errorf("unable to set value: %w", err).Error())
	}

	if offset := req.Paging.GetOffsetCombined().GetOffsetString(); offset != "" {
		filter.OffsetID.Set(offset)
	}
	if createdAt := req.Paging.GetOffsetCombined().GetOffsetTime(); createdAt.IsValid() {
		filter.CreatedAt.Set(createdAt.AsTime())
	}

	if len(req.Statuses) > 0 {
		ss := make([]string, 0, len(req.Statuses))
		for _, s := range req.Statuses {
			ss = append(ss, s.String())
		}

		filter.Statuses.Set(ss)
	}

	if req.Start.IsValid() {
		filter.StartDate.Set(req.Start.AsTime())
	}
	if req.End.IsValid() {
		filter.EndDate.Set(req.End.AsTime())
	}

	if req.CourseId.GetValue() != "" {
		filter.CourseID.Set(req.CourseId.Value)

		if len(req.ClassIds) > 0 {
			filter.ClassIDs.Set(req.ClassIds)
		}
	}

	if req.SearchText.GetValue() != "" {
		if req.SearchType == sspb.SearchType_SEARCH_TYPE_ASSIGNMENT_NAME {
			filter.AssignmentName.Set("%" + req.SearchText.Value + "%")
		}
	}

	ss, err := s.StudentSubmissionRepo.ListV3(ctx, s.DB, filter)

	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Errorf("database.ExecInTx: %w", err).Error())
	}
	items := toStudentSubmissionsPbV3(ss)
	var next *cpb.Paging
	if len(items) == int(req.Paging.Limit) {
		next = &cpb.Paging{
			Limit: req.Paging.Limit,
			Offset: &cpb.Paging_OffsetCombined{
				OffsetCombined: &cpb.Paging_Combined{
					OffsetString: items[len(items)-1].SubmissionId,
					OffsetTime:   items[len(items)-1].UpdatedAt,
				},
			},
		}
	}

	return &sspb.ListSubmissionsV3Response{
		NextPage: next,
		Items:    items,
	}, nil
}

func (s *StudentSubmissionService) ListSubmissionsV4(ctx context.Context, req *sspb.ListSubmissionsV4Request) (*sspb.ListSubmissionsV4Response, error) {
	if req.GetPaging() == nil {
		return nil, status.Error(codes.InvalidArgument, "empty paging")
	}

	if req.Paging.Limit < 1 || req.Paging.Limit > 100 {
		return nil, status.Error(codes.InvalidArgument, "limit must in range 1 to 100")
	}
	filter := &repositories.StudentSubmissionFilter{
		Limit: uint(req.Paging.Limit),
	}

	if err := multierr.Combine(
		filter.OffsetID.Set(nil),
		filter.CreatedAt.Set(nil),
		filter.StudentIDs.Set(nil),
		filter.Statuses.Set(nil),
		filter.StartDate.Set(nil),
		filter.EndDate.Set(nil),
		filter.AssignmentName.Set(nil),
		filter.CourseID.Set(nil),
		filter.ClassIDs.Set(nil),
		filter.LocationIDs.Set(req.LocationIds),
		filter.StudentName.Set(nil),
	); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Errorf("unable to set value: %w", err).Error())
	}

	if offset := req.Paging.GetOffsetCombined().GetOffsetString(); offset != "" {
		_ = filter.OffsetID.Set(offset)
	}
	if createdAt := req.Paging.GetOffsetCombined().GetOffsetTime(); createdAt.IsValid() {
		_ = filter.CreatedAt.Set(createdAt.AsTime())
	}

	if len(req.Statuses) > 0 {
		ss := make([]string, 0, len(req.Statuses))
		for _, s := range req.Statuses {
			ss = append(ss, s.String())
		}

		_ = filter.Statuses.Set(ss)
	}

	if req.Start.IsValid() {
		_ = filter.StartDate.Set(req.Start.AsTime())
	}
	if req.End.IsValid() {
		_ = filter.EndDate.Set(req.End.AsTime())
	}

	if req.CourseId.GetValue() != "" {
		_ = filter.CourseID.Set(req.CourseId.Value)

		if len(req.ClassIds) > 0 {
			_ = filter.ClassIDs.Set(req.ClassIds)
		}
	}

	if req.SearchText.GetValue() != "" {
		cleanValue := db.ReplaceSpecialChars(req.SearchText.Value)
		if req.SearchType == sspb.SearchType_SEARCH_TYPE_ASSIGNMENT_NAME {
			_ = filter.AssignmentName.Set(cleanValue)
		}
	}

	if req.StudentName.GetValue() != "" {
		cleanValue := db.ReplaceSpecialChars(req.StudentName.Value)
		_ = filter.StudentName.Set(cleanValue)
	}

	ss, err := s.StudentSubmissionRepo.ListV4(ctx, s.DB, filter)
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Errorf("database.ExecInTx: %w", err).Error())
	}
	items := toStudentSubmissionsPbV3(ss)
	var next *cpb.Paging
	if len(items) == int(req.Paging.Limit) {
		next = &cpb.Paging{
			Limit: req.Paging.Limit,
			Offset: &cpb.Paging_OffsetCombined{
				OffsetCombined: &cpb.Paging_Combined{
					OffsetString: items[len(items)-1].SubmissionId,
					OffsetTime:   items[len(items)-1].UpdatedAt,
				},
			},
		}
	}

	return &sspb.ListSubmissionsV4Response{
		NextPage: next,
		Items:    items,
	}, nil
}
func toStudentSubmissionsPbV3(ss []*repositories.StudentSubmissionInfo) []*sspb.StudentSubmission {
	results := make([]*sspb.StudentSubmission, 0, len(ss))

	for _, info := range ss {
		i := &sspb.StudentSubmission{
			SubmissionId: info.ID.String,
			StudyPlanItemIdentity: &sspb.StudyPlanItemIdentity{
				LearningMaterialId: info.LearningMaterialID.String,
				StudyPlanId:        info.StudyPlanID.String,
				StudentId:          wrapperspb.String(info.StudentID.String),
			},

			CreatedAt:          timestamppb.New(info.CreatedAt.Time),
			UpdatedAt:          timestamppb.New(info.UpdatedAt.Time),
			Status:             sspb.SubmissionStatus(pb.SubmissionStatus_value[info.Status.String]),
			Note:               info.Note.String,
			Duration:           info.Duration.Int,
			CorrectScore:       wrapperspb.Float(info.CorrectScore.Float),
			TotalScore:         wrapperspb.Float(info.TotalScore.Float),
			UnderstandingLevel: sspb.SubmissionUnderstandingLevel(pb.SubmissionUnderstandingLevel_value[info.UnderstandingLevel.String]),

			StartDate: timestamppb.New(info.StartDate.Time),
			EndDate:   timestamppb.New(info.EndDate.Time),
			CourseId:  info.CourseID.String,
		}

		if info.CompleteDate.Status == pgtype.Present {
			i.CompleteDate = timestamppb.New(info.CompleteDate.Time)
		}

		if info.SubmissionGradeID.Status == pgtype.Present {
			i.SubmissionGradeId = wrapperspb.String(info.SubmissionGradeID.String)
		}

		info.SubmissionContent.AssignTo(&i.SubmissionContent)

		results = append(results, i)
	}
	return results
}

func (s *StudentSubmissionService) validateRetrieveSubmissionHistoryRequest(req *sspb.RetrieveSubmissionHistoryRequest) error {
	if req.SetId == "" {
		return fmt.Errorf("SetId is missing")
	}
	return nil
}

func (s *StudentSubmissionService) RetrieveSubmissionHistory(ctx context.Context, req *sspb.RetrieveSubmissionHistoryRequest) (*sspb.RetrieveSubmissionHistoryResponse, error) {
	if err := s.validateRetrieveSubmissionHistoryRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Errorf("s.validateRetrieveSubmissionHistoryRequest: %w", err).Error())
	}

	if req.Paging == nil {
		req.Paging = &cpb.Paging{
			Limit:  100,
			Offset: &cpb.Paging_OffsetInteger{OffsetInteger: 0},
		}
	}
	limit := req.Paging.Limit
	offset := req.Paging.GetOffsetInteger()

	lm, err := s.ShuffledQuizSetRepo.GetRelatedLearningMaterial(ctx, s.DB, database.Text(req.SetId))
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("s.ShuffledQuizSetRepo.GetRelatedLearningMaterial: %v", err).Error())
	}
	var answerLogs []*cpb.AnswerLog
	var listQuizIDs []pgtype.Text

	switch new(sspb.LearningMaterialType).FromString(lm.Type.String) {
	case sspb.LearningMaterialType_LEARNING_MATERIAL_LEARNING_OBJECTIVE:
		if answerLogs, listQuizIDs, err = s.getLOAnswerLogsAndQuizIDs(ctx, database.Text(req.SetId), database.Int8(int64(limit)), database.Int8(int64(offset))); err != nil {
			return nil, status.Error(codes.Internal, fmt.Errorf("s.getLOAnswerLogsAndQuizIDs: %w", err).Error())
		}
	case sspb.LearningMaterialType_LEARNING_MATERIAL_FLASH_CARD:
		if answerLogs, listQuizIDs, err = s.getFlashCardAnswerLogsAndQuizIDs(ctx, database.Text(req.SetId), database.Int8(int64(limit)), database.Int8(int64(offset))); err != nil {
			return nil, status.Error(codes.Internal, fmt.Errorf("s.getFlashCardAnswerLogsAndQuizIDs: %w", err).Error())
		}
	default:
		return nil, status.Error(codes.InvalidArgument, "This learning material type is not supported")
	}

	quizIDs := sliceutils.Map(listQuizIDs, func(id pgtype.Text) string { return id.String })

	quizMap := make(map[string]*entities.Quiz)
	if len(answerLogs) != 0 {
		quizzesWrap, err := s.QuizRepo.GetByExternalIDs(ctx, s.DB, database.TextArray(quizIDs), lm.ID)
		if err != nil {
			return nil, status.Error(codes.Internal, fmt.Errorf("s.QuizRepo.GetByExternalIDs: %v", err).Error())
		}

		if len(quizzesWrap) == 0 {
			return nil, status.Error(codes.Internal, fmt.Errorf("s.QuizRepo.GetByExternalIDs: cannot find quiz %v", quizIDs).Error())
		}

		for _, quiz := range quizzesWrap {
			quizMap[quiz.ExternalID.String] = quiz
		}
	}

	for _, log := range answerLogs {
		quiz := quizMap[log.QuizId]
		core, err := toQuizCore(quiz)
		if err != nil {
			return nil, status.Error(codes.Internal, fmt.Errorf("toQuizCore: %v", err).Error())
		}
		if core.Kind == cpb.QuizType_QUIZ_TYPE_MCQ || core.Kind == cpb.QuizType_QUIZ_TYPE_MAQ {
			// only shuffle options of multiple choice quiz
			seedStr, err := s.ShuffledQuizSetRepo.GetSeed(ctx, s.DB, database.Text(req.SetId))
			if err != nil {
				return nil, status.Error(codes.Internal, fmt.Errorf("s.ShuffledQuizSetRepo.GetSeed: %v", err).Error())
			}
			seed, err := strconv.ParseInt(seedStr.String, 10, 64)
			if err != nil {
				return nil, status.Error(codes.Internal, fmt.Errorf("strconv.ParseInt: %v", err).Error())
			}
			idx, err := s.ShuffledQuizSetRepo.GetQuizIdx(ctx, s.DB, database.Text(req.SetId), quiz.ExternalID)
			if err != nil || idx.Int == 0 {
				return nil, status.Error(
					codes.Internal,
					fmt.Errorf(
						"s.ShuffledQuizSetRepo.GetQuizIdx: could not found quiz id %s in shuffled quiz set %s : %v",
						quiz.ExternalID.String,
						req.SetId,
						err,
					).Error(),
				)
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
					return nil, status.Error(codes.Internal, fmt.Errorf("could not convert options jsonb to QuizOption struct: %v", err).Error())
				}

				for _, opt := range optionsEnt {
					correctText = append(correctText, strings.TrimSpace(opt.GetText()))
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
	for _, quiz := range quizMap {
		if quiz.QuestionGroupID.Status == pgtype.Present && len(quiz.QuestionGroupID.String) != 0 {
			questionGrIDs = append(questionGrIDs, quiz.QuestionGroupID.String)
		}
	}

	var questionGroup entities.QuestionGroups
	if len(questionGrIDs) != 0 {
		questionGroup, err = s.QuestionGroupRepo.GetQuestionGroupsByIDs(ctx, s.DB, questionGrIDs...)
		if err != nil {
			return nil, status.Error(codes.Internal, fmt.Errorf("s.QuestionGroupRepo.GetQuestionGroupsByIDs: %w", err).Error())
		}
	}

	respQuestionGroups, err := entities.QuestionGroupsToQuestionGroupProtoBufMess(questionGroup)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("entities.QuestionGroupsToQuestionGroupProtoBufMess: %w", err).Error())
	}

	resp := &sspb.RetrieveSubmissionHistoryResponse{
		Logs: answerLogs,
		NextPage: &cpb.Paging{
			Limit:  limit,
			Offset: &cpb.Paging_OffsetInteger{OffsetInteger: int64(limit) + offset},
		},
		QuestionGroups: respQuestionGroups,
	}
	return resp, nil
}

func (s *StudentSubmissionService) getLOAnswerLogsAndQuizIDs(ctx context.Context, setID pgtype.Text, limit, offset pgtype.Int8) ([]*cpb.AnswerLog, []pgtype.Text, error) {
	answerLogs := make([]*cpb.AnswerLog, 0)

	answers, quizIDs, err := s.LOSubmissionAnswerRepo.ListSubmissionAnswers(ctx, s.DB, setID, limit, offset)
	if err != nil {
		return nil, nil, fmt.Errorf("s.LOSubmissionAnswerRepo.ListSubmissionAnswers: %w", err)
	}

	quizLoSubmissionAnswerMap := make(map[pgtype.Text]*entities.LOSubmissionAnswer)
	for _, answer := range answers {
		quizLoSubmissionAnswerMap[answer.QuizID] = answer
	}

	for _, quizID := range quizIDs {
		e := quizLoSubmissionAnswerMap[quizID]
		if e == nil {
			// incomplete quiz
			e = &entities.LOSubmissionAnswer{
				QuizID: quizID,
			}
		}
		answerLogs = append(answerLogs, loSubmissionAnswerToAnswerLog(e))
	}
	return answerLogs, quizIDs, nil
}

func loSubmissionAnswerToAnswerLog(e *entities.LOSubmissionAnswer) *cpb.AnswerLog {
	return &cpb.AnswerLog{
		QuizId:        e.QuizID.String,
		SelectedIndex: sliceutils.Map(database.Int4ArrayToInt32Array(e.StudentIndexAnswer), func(idx int32) uint32 { return uint32(idx) }),
		CorrectIndex:  sliceutils.Map(database.Int4ArrayToInt32Array(e.CorrectIndexAnswer), func(idx int32) uint32 { return uint32(idx) }),
		FilledText:    database.FromTextArray(e.StudentTextAnswer),
		CorrectText:   database.FromTextArray(e.CorrectTextAnswer),
		Correctness:   database.FromBoolArray(e.IsCorrect),
		IsAccepted:    e.IsAccepted.Bool,
		SubmittedAt:   timestamppb.New(e.CreatedAt.Time),
	}
}

func (s *StudentSubmissionService) getFlashCardAnswerLogsAndQuizIDs(ctx context.Context, setID pgtype.Text, limit, offset pgtype.Int8) ([]*cpb.AnswerLog, []pgtype.Text, error) {
	answerLogs := make([]*cpb.AnswerLog, 0)

	answers, quizIDs, err := s.FlashCardSubmissionAnswerRepo.ListSubmissionAnswers(ctx, s.DB, setID, limit, offset)
	if err != nil {
		return nil, nil, fmt.Errorf("s.FlashCardSubmissionAnswerRepo.ListSubmissionAnswers: %w", err)
	}

	quizLoSubmissionAnswerMap := make(map[pgtype.Text]*entities.FlashCardSubmissionAnswer)
	for _, answer := range answers {
		quizLoSubmissionAnswerMap[answer.QuizID] = answer
	}

	for _, quizID := range quizIDs {
		e := quizLoSubmissionAnswerMap[quizID]
		if e == nil {
			// incomplete quiz
			e = &entities.FlashCardSubmissionAnswer{
				QuizID: quizID,
			}
		}
		answerLogs = append(answerLogs, flashCardSubmissionAnswerToAnswerLog(e))
	}
	return answerLogs, quizIDs, nil
}

func flashCardSubmissionAnswerToAnswerLog(e *entities.FlashCardSubmissionAnswer) *cpb.AnswerLog {
	return &cpb.AnswerLog{
		QuizId:        e.QuizID.String,
		SelectedIndex: sliceutils.Map(database.Int4ArrayToInt32Array(e.StudentIndexAnswer), func(idx int32) uint32 { return uint32(idx) }),
		CorrectIndex:  sliceutils.Map(database.Int4ArrayToInt32Array(e.CorrectIndexAnswer), func(idx int32) uint32 { return uint32(idx) }),
		FilledText:    database.FromTextArray(e.StudentTextAnswer),
		CorrectText:   database.FromTextArray(e.CorrectTextAnswer),
		Correctness:   database.FromBoolArray(e.IsCorrect),
		IsAccepted:    e.IsAccepted.Bool,
		SubmittedAt:   timestamppb.New(e.CreatedAt.Time),
	}
}
