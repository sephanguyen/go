package question_tag_type

import (
	"context"
	"fmt"
	"math/rand"
	"strings"

	"github.com/manabie-com/backend/features/syllabus/utils"
	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"
)

func (s *Suite) userUpsertQuestionTagType(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	importQuestionTagTypesRequest := &sspb.ImportQuestionTagTypesRequest{
		Payload: stepState.CSVContent,
	}
	stepState.Response, stepState.ResponseErr = sspb.NewQuestionTagTypeClient(s.EurekaConn).ImportQuestionTagTypes(s.AuthHelper.SignedCtx(ctx, stepState.Token), importQuestionTagTypesRequest)
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) questionTagTypeMustBeCreated(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	query := `SELECT count(*) FROM question_tag_type WHERE question_tag_type_id = ANY($1::text[])`
	var count int
	if err := s.EurekaDB.QueryRow(ctx, query, stepState.QuestionTagTypeIDs).Scan(&count); err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}
	if count != len(stepState.QuestionTagTypeIDs) {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("expected to number of question tag type %d, got %d", len(stepState.QuestionTagTypeIDs), count)
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userUpdateQuestionTagType(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	rowFormat := "\n%s,%s"
	rows := "id,name"
	for _, id := range stepState.QuestionTagTypeIDs {
		name := fmt.Sprintf("updated-name-%s", id)
		rows += fmt.Sprintf(rowFormat, id, name)
	}
	importQuestionTagTypesRequest := &sspb.ImportQuestionTagTypesRequest{
		Payload: []byte(rows),
	}
	stepState.Response, stepState.ResponseErr = sspb.NewQuestionTagTypeClient(s.EurekaConn).ImportQuestionTagTypes(s.AuthHelper.SignedCtx(ctx, stepState.Token), importQuestionTagTypesRequest)
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) questionTagTypeMustBeUpdated(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	t := &entities.QuestionTagType{}
	ts := &entities.QuestionTagTypes{}
	query := fmt.Sprintf("SELECT %s FROM %s WHERE question_tag_type_id = ANY($1::_Text)", strings.Join(database.GetFieldNames(t), ","), t.TableName())
	if err := database.Select(ctx, s.EurekaDB, query, stepState.QuestionTagTypeIDs).ScanAll(ts); err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}
	items := ts.Get()
	for i, item := range items {
		expectedName := fmt.Sprintf("updated-name-%s", stepState.QuestionTagTypeIDs[i])
		if item.Name.String != expectedName {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unexpected updated name: got %s, want %s", item.Name.String, expectedName)
		}
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) aValidCSVContentWith(ctx context.Context, arg string) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	n := rand.Intn(5) + 3
	rowFormat := "\n%s,%s"
	rows := "id,name"
	for i := 0; i < n; i++ {
		id := idutil.ULIDNow()
		name := fmt.Sprintf("name-%s", id)
		if arg == "no id" {
			id = ""
		}
		rows += fmt.Sprintf(rowFormat, id, name)
		stepState.QuestionTagTypeIDs = append(stepState.QuestionTagTypeIDs, id)
	}
	stepState.CSVContent = []byte(rows)
	return utils.StepStateToContext(ctx, stepState), nil
}
