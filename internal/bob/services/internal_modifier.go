package services

import (
	"context"
	"fmt"
	"math"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type InternalModifierService struct {
	EurekaDBTrace database.Ext
	DB            database.Ext

	QuizRepo interface {
		GetByExternalIDs(context.Context, database.QueryExecer, pgtype.TextArray, pgtype.Text) (entities.Quizzes, error)
		GetOptions(context.Context, database.QueryExecer, pgtype.Text, pgtype.Text) ([]*entities.QuizOption, error)
	}
	StudentsLearningObjectivesCompletenessRepo interface {
		UpsertHighestQuizScore(ctx context.Context, db database.QueryExecer, loID pgtype.Text, studentID pgtype.Text, newScore pgtype.Float4) error
		UpsertFirstQuizCompleteness(ctx context.Context, db database.QueryExecer, loID pgtype.Text, studentID pgtype.Text, firstQuizScore pgtype.Float4) error
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
	}
	StudentRepo interface {
		GetCountryByStudent(ctx context.Context, db database.QueryExecer, studentID pgtype.Text) (string, error)
	}

	StudentLearningTimeDailyRepo interface {
		UpsertTaskAssignment(ctx context.Context, db database.QueryExecer, s *entities.StudentLearningTimeDaily) error
		Upsert(ctx context.Context, db database.QueryExecer, s *entities.StudentLearningTimeDaily) error
		Retrieve(ctx context.Context, db database.QueryExecer, studentID pgtype.Text, from, to *pgtype.Timestamptz, queryEnhancers ...repositories.QueryEnhancer) ([]*entities.StudentLearningTimeDaily, error)
	}
}

func (cm *InternalModifierService) SubmitQuizAnswers(ctx context.Context, req *bpb.SubmitQuizAnswersRequest) (_ *bpb.SubmitQuizAnswersResponse, err error) {
	if err := cm.validateCheckSubmitQuizAnswersRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	loID, err := cm.ShuffledQuizSetRepo.GetLoID(ctx, cm.EurekaDBTrace, database.Text(req.SetId))
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Errorf("SubmitQuizAnswers.ShuffledQuizSetRepo.GetLoID: %v", err).Error())
	}

	quizIDs := make([]string, 0, len(req.QuizAnswer))
	quizAnswersMap := make(map[string][]*bpb.Answer)
	for _, quizAnswer := range req.QuizAnswer {
		quizIDs = append(quizIDs, quizAnswer.QuizId)
		quizAnswersMap[quizAnswer.QuizId] = quizAnswer.Answer
	}
	quizzes, err := cm.QuizRepo.GetByExternalIDs(ctx, cm.EurekaDBTrace, database.TextArray(quizIDs), loID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Errorf("SubmitQuizAnswers.ShuffledQuizSetRepo.GetByExternalIDs: %v", err).Error())
	}

	answersEnt, err := cm.checkCorrectnessQuizzes(ctx, quizzes, quizAnswersMap, req)
	if err != nil {
		return nil, err
	}

	err = database.ExecInTx(ctx, cm.EurekaDBTrace, func(ctx context.Context, tx pgx.Tx) error {
		if err := cm.ShuffledQuizSetRepo.UpdateSubmissionHistory(ctx, tx, database.Text(req.SetId), database.JSONB(answersEnt)); err != nil {
			return fmt.Errorf("SubmitQuizAnswers.ShuffledQuizSetRepo.UpdateSubmissionHistory: %v", err)
		}

		if err := cm.ShuffledQuizSetRepo.UpdateTotalCorrectness(ctx, tx, database.Text(req.SetId)); err != nil {
			return fmt.Errorf("SubmitQuizAnswers.ShuffledQuizSetRepo.UpdateTotalCorrectness: %v", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	err = database.ExecInTx(ctx, cm.EurekaDBTrace, func(ctx context.Context, tx pgx.Tx) error {
		isFinished, err := cm.ShuffledQuizSetRepo.IsFinishedQuizTest(ctx, tx, database.Text(req.SetId))
		if err != nil && err != pgx.ErrNoRows {
			return fmt.Errorf("SubmitQuizAnswers.ShuffledQuizSetRepo.IsFinishedQuizTest: %v", err)
		}

		shuffledQuizSet, err := cm.ShuffledQuizSetRepo.Get(ctx, tx, database.Text(req.SetId), database.Int8(1), database.Int8(1))
		if err != nil {
			return fmt.Errorf("SubmitQuizAnswers.ShuffledQuizSetRepo.Get: %v", err)
		}
		isRetry := shuffledQuizSet.OriginalShuffleQuizSetID.Status == pgtype.Present

		if isRetry || isFinished.Bool {
			studentID, err := cm.ShuffledQuizSetRepo.GetStudentID(ctx, tx, database.Text(req.SetId))
			if err != nil {
				return fmt.Errorf("SubmitQuizAnswers.ShuffledQuizSetRepo.GetStudentID: %v", err)
			}

			totalCorrectness, totalQuiz, err := cm.ShuffledQuizSetRepo.GetScore(ctx, tx, database.Text(req.SetId))
			if err != nil {
				return fmt.Errorf("SubmitQuizAnswers.ShuffledQuizSetRepo.GetScore: %v", err)
			}

			if isRetry {
				externalIDsFromSubmissionHistory, err := cm.ShuffledQuizSetRepo.GetExternalIDsFromSubmissionHistory(ctx, tx, shuffledQuizSet.OriginalShuffleQuizSetID, false)
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
			if err = cm.StudentsLearningObjectivesCompletenessRepo.UpsertFirstQuizCompleteness(ctx, tx, loID, studentID, database.Float4(score)); err != nil {
				return fmt.Errorf("SubmitQuizAnswers.StudentsLearningObjectivesCompletenessRepo.UpsertFirstQuizCompleteness: %v", err)
			}

			if err = cm.StudentsLearningObjectivesCompletenessRepo.UpsertHighestQuizScore(ctx, tx, loID, studentID, database.Float4(score)); err != nil {
				return fmt.Errorf("SubmitQuizAnswers.StudentsLearningObjectivesCompletenessRepo.UpsertHighestQuizScore: %v", err)
			}
		}

		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	resp := &bpb.SubmitQuizAnswersResponse{}

	for _, answerEnt := range answersEnt {
		resp.Logs = append(resp.Logs, &cpb.AnswerLog{
			QuizId:        answerEnt.QuizID,
			QuizType:      cpb.QuizType(cpb.QuizType_value[answerEnt.QuizType]),
			SelectedIndex: answerEnt.SelectedIndex,
			CorrectIndex:  answerEnt.CorrectIndex,
			IsAccepted:    answerEnt.IsAccepted,
			CorrectText:   answerEnt.CorrectText,
			Correctness:   answerEnt.Correctness,
			SubmittedAt:   timestamppb.New(answerEnt.SubmittedAt),
			FilledText:    answerEnt.FilledText,
		})
	}

	return resp, nil
}

func (cm *InternalModifierService) checkCorrectnessQuizzes(
	ctx context.Context, quizzes entities.Quizzes,
	quizAnswersMap map[string][]*bpb.Answer, req *bpb.SubmitQuizAnswersRequest) (_ []*entities.QuizAnswer, err error) {
	answersEnt := make([]*entities.QuizAnswer, 0, len(quizzes))
	for _, quiz := range quizzes {
		var answerEnt *entities.QuizAnswer
		answers := quizAnswersMap[quiz.ExternalID.String]

		switch quiz.Kind.String {
		case cpb.QuizType_QUIZ_TYPE_MCQ.String(), cpb.QuizType_QUIZ_TYPE_MAQ.String():
			MCQuiz := &MultipleChoiceQuiz{
				Quiz:                quiz,
				SetID:               req.SetId,
				ShuffledQuizSetRepo: cm.ShuffledQuizSetRepo,
			}

			answerEnt, err = MCQuiz.CheckCorrectness(ctx, cm.EurekaDBTrace, answers)
			if err != nil {
				return nil, status.Error(codes.FailedPrecondition, fmt.Sprintf("SubmitQuizAnswers.MultipleChoice.CheckCorrectness: %v", err))
			}
		case cpb.QuizType_QUIZ_TYPE_MIQ.String():
			manualInputQuiz := &ManualInputQuiz{
				Quiz: quiz,
			}

			answerEnt, err = manualInputQuiz.CheckCorrectness(ctx, cm.EurekaDBTrace, answers)
			if err != nil {
				return nil, status.Error(codes.FailedPrecondition, fmt.Sprintf("SubmitQuizAnswers.ManualInputQuiz.CheckCorrectness: %v", err))
			}
		case cpb.QuizType_QUIZ_TYPE_FIB.String(), cpb.QuizType_QUIZ_TYPE_POW.String(), cpb.QuizType_QUIZ_TYPE_TAD.String():
			fillInTheBlankQuiz := &FillInTheBlankQuiz{
				Quiz: quiz,
			}

			answerEnt, err = fillInTheBlankQuiz.CheckCorrectness(ctx, cm.EurekaDBTrace, answers)
			if err != nil {
				return nil, status.Error(codes.FailedPrecondition, fmt.Sprintf("SubmitQuizAnswers.FillInTheBlank.CheckCorrectness: %v", err))
			}
		default:
			return nil, status.Error(codes.FailedPrecondition, "Not supported quiz type!")
		}

		answersEnt = append(answersEnt, answerEnt)
	}
	return answersEnt, nil
}

func (cm *InternalModifierService) validateCheckSubmitQuizAnswersRequest(req *bpb.SubmitQuizAnswersRequest) error {
	if req.SetId == "" {
		return fmt.Errorf("req must have SetId")
	}
	for _, quizAnswer := range req.QuizAnswer {
		if quizAnswer.QuizId == "" {
			return fmt.Errorf("req must have QuizId")
		}
		if len(quizAnswer.Answer) == 0 {
			return fmt.Errorf("req must have answer")
		}

		_, isSelectedIndexFormat := quizAnswer.Answer[0].Format.(*bpb.Answer_SelectedIndex)
		_, isFilledTextFormat := quizAnswer.Answer[0].Format.(*bpb.Answer_FilledText)
		if !isSelectedIndexFormat && !isFilledTextFormat {
			return fmt.Errorf(fmt.Sprintf("your answer of quiz_id(%s) is not a multiple choice neither fill in the blank", quizAnswer.QuizId))
		}
	}

	return nil
}
