package question_tag

import (
	"context"
	"fmt"
	"math/rand"
	"strings"

	"github.com/manabie-com/backend/features/syllabus/utils"
	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"
)

func (s *Suite) someQuestionTagTypesExistedInDatabase(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	n := rand.Intn(5) + 3
	rowFormat := "\n%s,%s"
	rows := "id,name"
	for i := 0; i < n; i++ {
		id := idutil.ULIDNow()
		name := fmt.Sprintf("name-%s", id)
		rows += fmt.Sprintf(rowFormat, id, name)
		stepState.QuestionTagTypeIDs = append(stepState.QuestionTagTypeIDs, id)
	}

	if _, err := sspb.NewQuestionTagTypeClient(s.EurekaConn).ImportQuestionTagTypes(s.AuthHelper.SignedCtx(ctx, stepState.Token), &sspb.ImportQuestionTagTypesRequest{
		Payload: []byte(rows),
	}); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("s.ImportQuestionTagTypes, err: %w", err)
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) aValidCSVContentWithSomeValidQuestionTags(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	rowFormat := "\n%s,%s,%s"
	rows := "id,name,question_tag_type_id"
	for i := 0; i < len(stepState.QuestionTagTypeIDs); i++ {
		id := idutil.ULIDNow()
		name := fmt.Sprintf("name-%s", id)
		questionTagTypeId := stepState.QuestionTagTypeIDs[i]
		rows += fmt.Sprintf(rowFormat, "", name, questionTagTypeId)
		stepState.QuestionTagNames = append(stepState.QuestionTagNames, name)
	}
	stepState.CSVContent = []byte(rows)
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userUpsertQuestionTag(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	req := &sspb.ImportQuestionTagRequest{
		Payload: stepState.CSVContent,
	}
	stepState.Response, stepState.ResponseErr = sspb.NewQuestionTagClient(s.EurekaConn).ImportQuestionTag(s.AuthHelper.SignedCtx(ctx, stepState.Token), req)
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userCreateQuestionTag(ctx context.Context) (context.Context, error) {
	ctx, _ = s.userUpsertQuestionTag(ctx)
	stepState := utils.StepStateFromContext[StepState](ctx)
	// need to get id after upsert because upsert won't use id field
	rawQuestionTag := entities.QuestionTag{}
	query := fmt.Sprintf("SELECT question_tag_id FROM %s WHERE name = ANY($1::TEXT[]) AND deleted_at IS NULL", rawQuestionTag.TableName())
	rows, err := s.EurekaDB.Query(ctx, query, stepState.QuestionTagNames)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("s.EurekaDB.Query, err: %w", err)
	}

	var tempID string
	for rows.Next() {
		rows.Scan(&tempID)
		stepState.QuestionTagIDs = append(stepState.QuestionTagIDs, tempID)
	}
	defer rows.Close()

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) questionTagMustBeCreated(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	rawQuestionTag := entities.QuestionTag{}
	var result int
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE question_tag_id = ANY($1::TEXT[]) AND deleted_at IS NULL", rawQuestionTag.TableName())
	if err := s.EurekaDB.QueryRow(ctx, query, stepState.QuestionTagIDs).Scan(&result); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("s.EurekaDB.QueryRow, err: %w", err)
	}

	if result != len(stepState.QuestionTagIDs) {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("expected to number of question tag %d, got %d", len(stepState.QuestionTagIDs), result)
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userUpdateQuestionTag(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	rowFormat := "\n%s,%s,%s"
	rows := "id,name,question_tag_type_id"
	for i, id := range stepState.QuestionTagIDs {
		name := fmt.Sprintf("updated-name-%s", id)
		questionTagTypeID := stepState.QuestionTagTypeIDs[i]
		rows += fmt.Sprintf(rowFormat, id, name, questionTagTypeID)
	}
	req := &sspb.ImportQuestionTagRequest{
		Payload: []byte(rows),
	}
	stepState.Response, stepState.ResponseErr = sspb.NewQuestionTagClient(s.EurekaConn).ImportQuestionTag(s.AuthHelper.SignedCtx(ctx, stepState.Token), req)
	stepState.Request = req
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) questionTagMustBeUpdated(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	rawQuestionTag := entities.QuestionTag{}
	query := fmt.Sprintf("SELECT name FROM %s WHERE question_tag_id = ANY($1::TEXT[]) AND deleted_at IS NULL", rawQuestionTag.TableName())
	rows, err := s.EurekaDB.Query(ctx, query, stepState.QuestionTagIDs)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("s.EurekaDB.Query, err: %w", err)
	}
	var tempName string
	for rows.Next() {
		rows.Scan(&tempName)
		if !strings.Contains(tempName, "updated") {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("name of question tag not updated")
		}
	}
	defer rows.Close()

	return utils.StepStateToContext(ctx, stepState), nil
}
