package eureka

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
)

func (s *suite) schoolAdminDeleteTopics(ctx context.Context, indexList string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.AuthToken = stepState.SchoolAdminToken
	ctx = contextWithToken(s, ctx)
	topicIds := []string{}
	for _, idx := range strings.Split(indexList, ",") {
		i, err := strconv.Atoi(strings.TrimSpace(idx))
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		topicIds = append(topicIds, stepState.TopicIDs[i])
	}

	if _, err := epb.NewTopicModifierServiceClient(s.Conn).DeleteTopics(ctx, &epb.DeleteTopicsRequest{
		TopicIds: topicIds,
	}); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to delete topics: %w", err)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) someMissingTopicIds(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	for i := 0; i < rand.Intn(5)+3; i++ {
		stepState.TopicIDs = append(stepState.TopicIDs, idutil.ULIDNow())
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userDeleteSomeTopics(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.Response, stepState.ResponseErr = epb.NewTopicModifierServiceClient(s.Conn).DeleteTopics(ctx, &epb.DeleteTopicsRequest{
		TopicIds: stepState.TopicIDs,
	})
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) ourSystemMustDeleteTheTopicsCorrectly(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	e := entities.Topic{}
	query := fmt.Sprintf(`SELECT COUNT(*) FROM %s WHERE topic_id = ANY($1::TEXT[]) AND deleted_at IS NOT NULL`, e.TableName())
	var softDeletedTopicsCount int
	if err := s.DB.QueryRow(ctx, query, database.TextArray(stepState.TopicIDs)).Scan(&softDeletedTopicsCount); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("Failed to query soft deleted topics count: %w", err)
	}
	if softDeletedTopicsCount != len(stepState.TopicIDs) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("Failed to soft deleted topics expected: %v, got: %v", len(stepState.ChapterIDs), softDeletedTopicsCount)
	}
	return StepStateToContext(ctx, stepState), nil
}
