// Package question will handle logic for question group,
// ordering question, essay question
package question

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/s3"
	"github.com/manabie-com/backend/internal/golibs/stringutil"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"
	ypb "github.com/manabie-com/backend/pkg/manabuf/yasuo/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type GroupRepo interface {
	Upsert(context.Context, database.QueryExecer, *entities.QuestionGroup) (int64, error)
	FindByID(ctx context.Context, db database.Ext, id string) (*entities.QuestionGroup, error)
	DeleteByID(ctx context.Context, db database.QueryExecer, questionGroupID pgtype.Text) error
}

type QuizSetRepo interface {
	GetQuizSetByLoID(context.Context, database.QueryExecer, pgtype.Text) (*entities.QuizSet, error)
	Delete(context.Context, database.QueryExecer, pgtype.Text) error
	Create(context.Context, database.QueryExecer, *entities.QuizSet) error
}

type QuizRepo interface {
	GetByQuestionGroupID(ctx context.Context, db database.QueryExecer, questionGroupID pgtype.Text) (entities.Quizzes, error)
	DeleteByQuestionGroupID(ctx context.Context, db database.QueryExecer, questionGroupID pgtype.Text) error
}

type QuizModifierService interface {
	AppendQuestionToQuizSetByLoID(ctx context.Context, db database.QueryExecer, loID string, quizExternalIDs []string, questionGroupIDs []string) (string, error)
}

type YasuoUploadReader interface {
	RetrieveUploadInfo(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*ypb.RetrieveUploadInfoResponse, error)
}

type YasuoUploadModifier interface {
	UploadHtmlContent(ctx context.Context, in *ypb.UploadHtmlContentRequest, opts ...grpc.CallOption) (*ypb.UploadHtmlContentResponse, error)
}

type Service struct {
	DB database.Ext
	sspb.UnimplementedQuestionServiceServer
	GroupRepo
	QuizSetRepo
	QuizRepo
	QuizModifierService
	YasuoUploadReader
	YasuoUploadModifier
}

func NewQuestionService(
	db database.Ext,
	qgRepo GroupRepo,
	quizSetRepo QuizSetRepo,
	quizRepo QuizRepo,
	qmsrv QuizModifierService,
	yasuoUploadReader YasuoUploadReader,
	yasuoUploadModifier YasuoUploadModifier,
) *Service {
	return &Service{
		DB:                  db,
		GroupRepo:           qgRepo,
		QuizSetRepo:         quizSetRepo,
		QuizRepo:            quizRepo,
		QuizModifierService: qmsrv,
		YasuoUploadReader:   yasuoUploadReader,
		YasuoUploadModifier: yasuoUploadModifier,
	}
}

func (s *Service) UpsertQuestionGroup(ctx context.Context, req *sspb.UpsertQuestionGroupRequest) (*sspb.UpsertQuestionGroupResponse, error) {
	if req.RichDescription == nil {
		req.RichDescription = &cpb.RichText{}
	}

	var isInserting bool
	if len(req.QuestionGroupId) == 0 {
		isInserting = true
	}

	e, err := s.upsertQuestionGroupRequestToQuestionGroup(req)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	mdCtx, err := interceptors.GetOutgoingContext(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, fmt.Errorf("interceptors.GetOutgoingContext: %w", err).Error())
	}

	isContentChanged := true

	uresp, err := s.YasuoUploadReader.RetrieveUploadInfo(mdCtx, &emptypb.Empty{})
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	endpoint := uresp.GetEndpoint()
	bucket := uresp.GetBucket()

	url, err := s3.GenerateUploadURL(endpoint, bucket, req.RichDescription.Rendered)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if !isInserting {
		questionGroup, err := s.GroupRepo.FindByID(ctx, s.DB, req.GetQuestionGroupId())
		if err != nil {
			return nil, status.Error(codes.Internal, fmt.Errorf("groupRepo.FindByID: %w", err).Error())
		}

		richDescription, err := questionGroup.GetRichDescription()
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		if richDescription.RenderedURL == url {
			isContentChanged = false
		}
	}

	e.RichDescription = database.JSONB(&entities.RichText{
		Raw:         req.RichDescription.Raw,
		RenderedURL: url,
	})

	if err = database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		// upsert question group into db
		n, err := s.GroupRepo.Upsert(ctx, tx, e)
		if err != nil {
			return fmt.Errorf("QuestionGroupRepo.Upsert: %w", err)
		}
		if n == 0 {
			return fmt.Errorf("could not upsert question group id: %s", e.QuestionGroupID.String)
		}

		if isInserting {
			// clone new quiz set
			// just NOT pass quiz id at here, we will a new quiz set
			if _, err = s.QuizModifierService.AppendQuestionToQuizSetByLoID(ctx, tx, req.LearningMaterialId, nil, []string{e.QuestionGroupID.String}); err != nil {
				return fmt.Errorf("QuizModifierService.AppendQuestionToQuizSetByLoID: %w", err)
			}
		}

		if isContentChanged {
			r, err := s.YasuoUploadModifier.UploadHtmlContent(mdCtx, &ypb.UploadHtmlContentRequest{
				Content: req.RichDescription.Rendered,
			})
			if err != nil {
				return fmt.Errorf("s.YasuoUploadModifier.UploadHtmlContent: %w", err)
			}

			if r.GetUrl() != url {
				return fmt.Errorf("url return does not match")
			}
		}

		return nil
	}); err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	return &sspb.UpsertQuestionGroupResponse{
		QuestionGroupId: e.QuestionGroupID.String,
	}, nil
}

func (s *Service) DeleteQuestionGroup(ctx context.Context, req *sspb.DeleteQuestionGroupRequest) (*sspb.DeleteQuestionGroupResponse, error) {
	questionGroupID := req.QuestionGroupId

	if questionGroupID == "" {
		return nil, status.Error(codes.InvalidArgument, fmt.Errorf("invalid question group id").Error())
	}

	questionGroup, err := s.GroupRepo.FindByID(ctx, s.DB, questionGroupID)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("GroupRepo.FindByID: %w", err).Error())
	}
	loID := questionGroup.LearningMaterialID

	quizSet, err := s.QuizSetRepo.GetQuizSetByLoID(ctx, s.DB, loID)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("QuizSetRepo.GetQuizSetByLoID: %w", err).Error())
	}

	quizzes, err := s.QuizRepo.GetByQuestionGroupID(ctx, s.DB, database.Text(questionGroupID))
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("QuizRepo.GetByQuestionGroupID: %w", err).Error())
	}

	quizzesExternalIDs := []string{}

	for _, quiz := range quizzes {
		quizzesExternalIDs = append(quizzesExternalIDs, quiz.ExternalID.String)
	}

	if err = database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		if err := s.GroupRepo.DeleteByID(ctx, tx, database.Text(questionGroupID)); err != nil {
			return fmt.Errorf("QuestionGroupRepo.DeleteByID: %w", err)
		}

		if len(quizzes) != 0 {
			if err := s.QuizRepo.DeleteByQuestionGroupID(ctx, tx, database.Text(questionGroupID)); err != nil {
				return fmt.Errorf("QuizRepo.DeleteByQuestionGroupID: %w", err)
			}
		}

		if err := s.QuizSetRepo.Delete(ctx, tx, quizSet.ID); err != nil {
			return fmt.Errorf("QuizSetRepo.Delete: %w", err)
		}

		var questionHierarchy entities.QuestionHierarchy
		var quizSetExternalIDs []string

		if err := multierr.Combine(
			quizSet.QuizExternalIDs.AssignTo(&quizSetExternalIDs),
			quizSet.QuestionHierarchy.AssignTo(&questionHierarchy),
		); err != nil {
			return err
		}

		questionHierarchy = questionHierarchy.ExcludeQuestionGroupIDs([]string{questionGroupID})
		quizSetExternalIDs = stringutil.SliceElementsDiff(quizSetExternalIDs, quizzesExternalIDs)

		if quizSetExternalIDs == nil {
			quizSetExternalIDs = []string{}
		}

		if err := multierr.Combine(
			quizSet.QuizExternalIDs.Set(quizSetExternalIDs),
			quizSet.QuestionHierarchy.Set(questionHierarchy),
			s.QuizSetRepo.Create(ctx, tx, quizSet),
		); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	return &sspb.DeleteQuestionGroupResponse{}, nil
}

func (s *Service) upsertQuestionGroupRequestToQuestionGroup(req *sspb.UpsertQuestionGroupRequest) (*entities.QuestionGroup, error) {
	e := &entities.QuestionGroup{}
	database.AllNullEntity(e)
	if err := multierr.Combine(
		e.QuestionGroupID.Set(req.QuestionGroupId),
		e.LearningMaterialID.Set(req.LearningMaterialId),
		e.Name.Set(req.Name),
		e.Description.Set(req.Description),
	); err != nil {
		return nil, fmt.Errorf("got error when mapping request to QuestionGroup entity: %w", err)
	}

	if err := e.IsValidToUpsert(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	return e, nil
}

// UpdateDisplayOrderOfQuizSet in quiz set of lo, user update display order of quizzes in that list
func (s *Service) UpdateDisplayOrderOfQuizSetV2(ctx context.Context, req *sspb.UpdateDisplayOrderOfQuizSetV2Request) (*sspb.UpdateDisplayOrderOfQuizSetV2Response, error) {
	quizSet, err := s.QuizSetRepo.GetQuizSetByLoID(ctx, s.DB, database.Text(req.LearningMaterialId))
	if err != nil {
		return nil, status.Errorf(codes.NotFound, fmt.Errorf("UpdateDisplayOrderOfQuizSetV2.QuizSetRepo.GetQuizSetByLoID: %w", err).Error())
	}

	newQuizExternalIDList, newQuestionHierarchy, err := s.validateUpdateDisplayOrderReq(req, quizSet.QuestionHierarchy)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Errorf("unable to validate req: %w", err).Error())
	}

	if err = database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		if err := s.QuizSetRepo.Delete(ctx, tx, quizSet.ID); err != nil {
			return fmt.Errorf("UpdateDisplayOrderOfQuizSetV2.QuizSetRepo.Delete: %w", err)
		}

		if err := multierr.Combine(
			quizSet.QuizExternalIDs.Set(newQuizExternalIDList),
			quizSet.QuestionHierarchy.Set(newQuestionHierarchy),
			s.QuizSetRepo.Create(ctx, tx, quizSet),
		); err != nil {
			return fmt.Errorf("UpdateDisplayOrderOfQuizSetV2.quizSetRepo.Create: %w", err)
		}

		return nil
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "ExecInTx err: %v", err)
	}

	return &sspb.UpdateDisplayOrderOfQuizSetV2Response{}, nil
}

func (s *Service) validateUpdateDisplayOrderReq(req *sspb.UpdateDisplayOrderOfQuizSetV2Request, pgQuestionHierarchy pgtype.JSONBArray) ([]string, entities.QuestionHierarchy, error) {
	if len(req.QuestionHierarchy) != len(pgQuestionHierarchy.Elements) {
		return nil, nil, fmt.Errorf("expected question hierarchy length is %d, got %d", len(pgQuestionHierarchy.Elements), len(req.QuestionHierarchy))
	}

	// create map for current q_hierarchy for quick comparison
	oldQuestionHierarchy := entities.QuestionHierarchy{}
	if err := oldQuestionHierarchy.UnmarshalJSONBArray(pgQuestionHierarchy); err != nil {
		return nil, nil, fmt.Errorf("unable to unmarshal: %w", err)
	}

	newQuestionHierarchy := entities.QuestionHierarchy{}
	newQuestionHierarchy.AppendQuestionHierarchyFromPb(req.QuestionHierarchy)

	if err := newQuestionHierarchy.IsIDDuplicated(); err != nil {
		return nil, nil, fmt.Errorf("id validation failed: %w", err)
	}

	if err := newQuestionHierarchy.IsElementsMatched(oldQuestionHierarchy); err != nil {
		return nil, nil, fmt.Errorf("question hierarchy's elements not matched: %w", err)
	}

	newQuizExternalIDList, err := newQuestionHierarchy.CreateQuizExternalIDs()
	if err != nil {
		return nil, nil, fmt.Errorf("unable to create quiz external ids: %w", err)
	}

	return newQuizExternalIDList, newQuestionHierarchy, nil
}

type checkQuestionsCorrectnessOption func(executor Executor) error

//nolint:revive
func WithCheckQuizCorrectnessRequest(req []*sspb.CheckQuizCorrectnessRequest) checkQuestionsCorrectnessOption {
	return func(executor Executor) (err error) {
		_, err = executor.GetUserAnswerFromCheckQuizCorrectnessRequest(req)
		if err != nil {
			return fmt.Errorf("WithCheckQuizCorrectnessRequest.executor.GetUserAnswerFromCheckQuizCorrectnessRequest: %w", err)
		}
		return nil
	}
}

//nolint:revive
func WithSubmitQuizAnswersRequest(req *epb.SubmitQuizAnswersRequest) checkQuestionsCorrectnessOption {
	return func(executor Executor) (err error) {
		_, err = executor.GetUserAnswerFromSubmitQuizAnswersRequest(req)
		if err != nil {
			return fmt.Errorf("WithSubmitQuizAnswersRequest.executor.GetUserAnswerFromSubmitQuizAnswersRequest: %w", err)
		}
		return nil
	}
}

// CheckQuestionsCorrectness will handle logic check Correctness for
// ordering question, essay question
func (s *Service) CheckQuestionsCorrectness(quizzes entities.Quizzes, opt checkQuestionsCorrectnessOption) ([]*entities.QuizAnswer, error) {
	if opt == nil {
		return nil, fmt.Errorf("need input answer")
	}
	result := make([]*entities.QuizAnswer, 0, len(quizzes))
	for _, quiz := range quizzes {
		var executor Executor
		switch quiz.Kind.String {
		case cpb.QuizType_QUIZ_TYPE_ORD.String():
			executor = NewOrderingQuestion(quiz)
		case cpb.QuizType_QUIZ_TYPE_ESQ.String():
			executor = NewEssayQuestion(quiz)
		default:
			result = append(result, nil)
			continue
		}

		if err := opt(executor); err != nil {
			return nil, fmt.Errorf("run opt: %w", err)
		}
		res, err := executor.CheckCorrectness()
		if err != nil {
			return nil, fmt.Errorf("executor.CheckCorrectness: %w", err)
		}
		result = append(result, res)
	}

	return result, nil
}

type Executor interface {
	GetUserAnswerFromSubmitQuizAnswersRequest(submit *epb.SubmitQuizAnswersRequest) (Executor, error)
	GetUserAnswerFromCheckQuizCorrectnessRequest(submit []*sspb.CheckQuizCorrectnessRequest) (Executor, error)
	CheckCorrectness() (*entities.QuizAnswer, error)
	ResetUserAnswer()
	GetQuizExternalID() string
}

type Input interface {
	GetQuizId() string
}
