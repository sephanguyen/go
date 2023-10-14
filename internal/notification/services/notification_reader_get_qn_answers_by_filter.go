package services

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/notification/repositories"
	"github.com/manabie-com/backend/internal/notification/services/mappers"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (svc *NotificationReaderService) GetAnswersByFilter(ctx context.Context, req *npb.GetAnswersByFilterRequest) (*npb.GetAnswersByFilterResponse, error) {
	if req.Paging == nil {
		req.Paging = &cpb.Paging{
			Limit:  100,
			Offset: &cpb.Paging_OffsetInteger{OffsetInteger: 0},
		}
	}

	if req.Paging.Limit == 0 {
		req.Paging.Limit = 100
	}

	// Get questions by questionnaire_id
	questionnaireQuestions, err := svc.QuestionnaireRepo.FindQuestionsByQnID(ctx, svc.DB, req.QuestionnaireId)
	if err != nil {
		return nil, status.Error(codes.Internal, "an error occurred when finding questions: "+err.Error())
	}
	if len(questionnaireQuestions) == 0 {
		return nil, status.Error(codes.InvalidArgument, "No questions found with questionnaireId: "+req.QuestionnaireId)
	}

	questionnaireQuestionIDs := make([]string, 0)
	for _, q := range questionnaireQuestions {
		questionnaireQuestionIDs = append(questionnaireQuestionIDs, q.QuestionnaireQuestionID.String)
	}

	// Get responders of questionnaire using found users with search_text and paginate it (join with users table of usermnmgt squad)
	findQuestionnaireRespondersFilter := repositories.NewFindQuestionnaireRespondersFilter()
	findQuestionnaireRespondersFilter.QuestionnaireID = database.Text(req.QuestionnaireId)
	findQuestionnaireRespondersFilter.UserName = database.Text(req.Keyword)

	_ = findQuestionnaireRespondersFilter.Offset.Set(req.Paging.GetOffsetInteger())
	_ = findQuestionnaireRespondersFilter.Limit.Set(req.Paging.Limit)

	totalCount, questionnaireResponders, err := svc.QuestionnaireRepo.FindQuestionnaireResponders(ctx, svc.DB, &findQuestionnaireRespondersFilter)
	if err != nil {
		return nil, status.Error(codes.Internal, "an error occurred when finding questionnaire responders: "+err.Error())
	}

	responderIDs := make([]string, 0)
	targetIDs := make([]string, 0)
	for _, questionnaireResponder := range questionnaireResponders {
		responderIDs = append(responderIDs, questionnaireResponder.UserID.String)
		targetIDs = append(targetIDs, questionnaireResponder.TargetID.String)
	}

	// Get questionnaire user answers using found responders and target answered questionnaire
	findUserAnswersFilter := repositories.NewFindUserAnswersFilter()
	findUserAnswersFilter.QuestionnaireQuestionIDs = database.TextArray(questionnaireQuestionIDs)
	findUserAnswersFilter.UserIDs = database.TextArray(responderIDs)
	findUserAnswersFilter.TargetIDs = database.TextArray(targetIDs)
	userAnswers, err := svc.QuestionnaireRepo.FindUserAnswers(ctx, svc.DB, &findUserAnswersFilter)
	if err != nil {
		return nil, status.Error(codes.Internal, "an error occurred when finding questinnaire user answers: "+err.Error())
	}

	offsetPre := req.Paging.GetOffsetInteger() - int64(req.Paging.Limit)

	if offsetPre < 0 {
		offsetPre = 0
	}

	return &npb.GetAnswersByFilterResponse{
		Questions:   mappers.QNQuestionsToPb(questionnaireQuestions),
		UserAnswers: mappers.QNUserAnswersToPb(questionnaireResponders, userAnswers, questionnaireQuestions),
		TotalItems:  totalCount,
		NextPage: &cpb.Paging{
			Limit:  req.Paging.Limit,
			Offset: &cpb.Paging_OffsetInteger{OffsetInteger: req.Paging.GetOffsetInteger() + int64(len(questionnaireResponders))},
		},
		PreviousPage: &cpb.Paging{
			Limit:  req.Paging.Limit,
			Offset: &cpb.Paging_OffsetInteger{OffsetInteger: offsetPre},
		},
	}, nil
}
