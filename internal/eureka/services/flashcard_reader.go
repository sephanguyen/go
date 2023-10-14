package services

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func NewFlashCardReaderService(
	db database.Ext, env string,
) *FlashCardReaderService {
	return &FlashCardReaderService{
		DB:                       db,
		Env:                      env,
		QuizRepo:                 &repositories.QuizRepo{},
		FlashcardProgressionRepo: &repositories.FlashcardProgressionRepo{},
	}
}

type FlashCardReaderService struct {
	pb.UnimplementedCourseReaderServiceServer
	DB  database.Ext
	Env string

	QuizRepo interface {
		GetByExternalIDsAndLOID(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray, loID pgtype.Text) (entities.Quizzes, error)
		GetByExternalIDsAndLmID(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray, lmID pgtype.Text) (entities.Quizzes, error)
	}

	FlashcardProgressionRepo interface {
		Get(ctx context.Context, db database.QueryExecer, args *repositories.GetFlashcardProgressionArgs) (*entities.FlashcardProgression, error)
		GetLastFlashcardProgression(
			ctx context.Context, db database.QueryExecer,
			studyPlanItemID, loID, studentID pgtype.Text, isCompleted pgtype.Bool,
		) (*entities.FlashcardProgression, error)
	}
}

func (c *FlashCardReaderService) RetrieveFlashCardStudyProgress(ctx context.Context, req *pb.RetrieveFlashCardStudyProgressRequest) (*pb.RetrieveFlashCardStudyProgressResponse, error) {
	if err := c.validateRetrieveFlashCardStudyProgressRequest(req); err != nil {
		return nil, err
	}

	pbQuizzes, pagingFlashcardProgression, err := getFlashcardProgressionWithPaging(ctx, c.DB,
		c.FlashcardProgressionRepo, c.QuizRepo, req.Paging,
		database.Text(req.StudySetId), database.Text(req.StudentId))
	if err != nil {
		return nil, err
	}

	resp := &pb.RetrieveFlashCardStudyProgressResponse{
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

func (c *FlashCardReaderService) RetrieveLastFlashcardStudyProgress(
	ctx context.Context, req *pb.RetrieveLastFlashcardStudyProgressRequest,
) (*pb.RetrieveLastFlashcardStudyProgressResponse, error) {
	if err := c.validateRetrieveLastFlashcardStudyProgress(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	var studySetID string
	studyPlanItemID := pgtype.Text{Status: pgtype.Null}
	if req.StudyPlanItemId != "" {
		studyPlanItemID = database.Text(req.StudyPlanItemId)
	}

	flashcardProgression, err := c.FlashcardProgressionRepo.GetLastFlashcardProgression(ctx, c.DB,
		studyPlanItemID, database.Text(req.LoId), database.Text(req.StudentId), database.Bool(req.IsCompleted))
	if err != nil && err != pgx.ErrNoRows {
		return nil, status.Error(codes.Internal, fmt.Errorf("c.FlashcardProgressionRepo.GetLastFlashcardProgression: %v", err).Error())
	}

	if flashcardProgression != nil {
		studySetID = flashcardProgression.StudySetID.String
	}

	return &pb.RetrieveLastFlashcardStudyProgressResponse{
		StudySetId: studySetID,
	}, nil
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
		GetByExternalIDsAndLOID(context.Context, database.QueryExecer, pgtype.TextArray, pgtype.Text) (entities.Quizzes, error)
		GetByExternalIDsAndLmID(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray, lmID pgtype.Text) (entities.Quizzes, error)
	},
	paging *cpb.Paging,
	studySetID, studentID pgtype.Text,
) ([]*pb.FlashcardQuizzes, *entities.FlashcardProgression, error) {
	offset := paging.GetOffsetInteger()
	limit := paging.Limit
	from := database.Int8(offset)
	to := database.Int8(offset + int64(limit) - 1)

	pagingFlashcardProgression, err := flashcardProgressionRepo.Get(ctx, db, &repositories.GetFlashcardProgressionArgs{
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
	quizzes, err := quizRepo.GetByExternalIDsAndLmID(ctx, db, pagingFlashcardProgression.QuizExternalIDs, pagingFlashcardProgression.LoID)
	if err != nil {
		return nil, nil, status.Errorf(codes.Internal, "getFlashcardProgressionWithPaging.QuizRepo.GetByExternalIDsAndLmID: %v", err)
	}
	// Convert from entities.Quizzes to []pb.FlashcardProgression
	pbQuizzes, err := toListFlashcardQuizzes(pagingFlashcardProgression.LoID.String, quizzes, skippedQuestionIDsMap, rememberedQuestionIDsMap)
	if err != nil {
		return nil, nil, status.Errorf(codes.Internal, "getFlashcardProgressionWithPaging.toListFlashcardQuizzes: %v", err)
	}
	return pbQuizzes, pagingFlashcardProgression, nil
}

func toListFlashcardQuizzes(
	loID string, quizzes entities.Quizzes,
	skippedQuestionIDsMap, rememberedQuestionIDsMap map[string]int64,
) ([]*pb.FlashcardQuizzes, error) {
	pbQuizzes := []*pb.FlashcardQuizzes{}
	for _, quiz := range quizzes {
		pbQuiz, err := toQuizPb(loID, quiz)
		if err != nil {
			return nil, err
		}
		flashcardQuiz := &pb.FlashcardQuizzes{
			Item:   pbQuiz,
			Status: pb.FlashcardQuizStudyStatus_FLASHCARD_QUIZ_STUDY_STATUS_NONE,
		}
		if _, ok := skippedQuestionIDsMap[quiz.ExternalID.String]; ok {
			flashcardQuiz.Status = pb.FlashcardQuizStudyStatus_FLASHCARD_QUIZ_STUDY_STATUS_SKIPPED
		}
		if _, ok := rememberedQuestionIDsMap[quiz.ExternalID.String]; ok {
			flashcardQuiz.Status = pb.FlashcardQuizStudyStatus_FLASHCARD_QUIZ_STUDY_STATUS_REMEMBERED
		}
		pbQuizzes = append(pbQuizzes, flashcardQuiz)
	}
	return pbQuizzes, nil
}

func toQuizPb(loID string, quiz *entities.Quiz) (*cpb.Quiz, error) {
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

func toQuizCore(quiz *entities.Quiz) (*cpb.QuizCore, error) {
	updatedAt := timestamppb.New(quiz.UpdatedAt.Time)
	createdAt := timestamppb.New(quiz.CreatedAt.Time)
	deletedAt := timestamppb.New(quiz.DeletedAt.Time)

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
	var point *wrapperspb.Int32Value
	if quiz.Point.Status == pgtype.Present {
		point = wrapperspb.Int32(quiz.Point.Int)
	}
	var questionGroupID *wrapperspb.StringValue
	if quiz.QuestionGroupID.Status == pgtype.Present {
		questionGroupID = wrapperspb.String(quiz.QuestionGroupID.String)
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
		Point:           point,
		QuestionGroupId: questionGroupID,
		QuestionTagIds:  database.FromTextArray(quiz.QuestionTagIds),
	}

	core = setAnswerConfig(quiz, core)

	return core, nil
}

func setAnswerConfig(quizEnt *entities.Quiz, quizCorePb *cpb.QuizCore) *cpb.QuizCore {
	options, _ := quizEnt.GetOptions()

	if quizEnt.Kind.String == cpb.QuizType_QUIZ_TYPE_ESQ.String() && len(options) != 0 {
		quizCorePb.AnswerConfig = &cpb.QuizCore_Essay{
			Essay: &cpb.EssayConfig{
				LimitEnabled: options[0].AnswerConfig.Essay.LimitEnabled,
				LimitType:    cpb.EssayLimitType(cpb.EssayLimitType_value[string(options[0].AnswerConfig.Essay.LimitType)]),
				Limit:        int32(options[0].AnswerConfig.Essay.Limit),
			},
		}
	}

	return quizCorePb
}

func (c *FlashCardReaderService) validateRetrieveFlashCardStudyProgressRequest(req *pb.RetrieveFlashCardStudyProgressRequest) error {
	if req.StudySetId == "" {
		return status.Error(codes.InvalidArgument, "req must have study set id")
	}
	if req.StudentId == "" {
		return status.Error(codes.InvalidArgument, "req must have student id")
	}

	if req.Paging == nil {
		return status.Error(codes.InvalidArgument, "req must have paging field")
	}

	if req.Paging.GetOffsetInteger() <= 0 {
		return status.Error(codes.InvalidArgument, "offset must be positive")
	}

	if req.Paging.Limit <= 0 {
		req.Paging.Limit = 100
	}
	return nil
}

func (c *FlashCardReaderService) validateRetrieveLastFlashcardStudyProgress(req *pb.RetrieveLastFlashcardStudyProgressRequest) error {
	if req.StudentId == "" {
		return errors.New("req must have student id")
	}
	if req.LoId == "" {
		return errors.New("req must have learning objective id")
	}

	return nil
}
