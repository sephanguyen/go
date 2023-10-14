package eurekav2

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/idutil"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v2"
)

func (s *suite) GetBookHierarchyFlattenByLmID(ctx context.Context, validity string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx = contextWithToken(s, ctx)
	learningMaterialID := stepState.LearningMaterialIDs[0]
	if validity == "not-existing" {
		learningMaterialID = idutil.ULIDNow()
	}
	stepState.Response, stepState.ResponseErr = epb.
		NewBookServiceClient(s.EurekaConn).
		GetBookHierarchyFlattenByLearningMaterialID(ctx, &epb.GetBookHierarchyFlattenByLearningMaterialIDRequest{
			LearningMaterialId: learningMaterialID,
		})

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) CheckBookHierarchyFlattenByLmID(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	bookHierarchyFlatten := stepState.Response.(*epb.GetBookHierarchyFlattenByLearningMaterialIDResponse).BookHierarchyFlatten

	bookID := stepState.BookID
	chapterID := stepState.ChapterIDs[0]
	topicID := stepState.TopicIDs[0]
	learningMaterialID := stepState.LearningMaterialIDs[0]

	expectedHierarchyFlatten := strings.Join([]string{bookID, chapterID, topicID, learningMaterialID}, "#")
	actualHierarchyFlatten := strings.Join([]string{bookHierarchyFlatten.BookId, bookHierarchyFlatten.ChapterId, bookHierarchyFlatten.TopicId, bookHierarchyFlatten.LearningMaterialId}, "#")

	if actualHierarchyFlatten != expectedHierarchyFlatten {
		return StepStateToContext(ctx, stepState), fmt.Errorf("CheckBookHierarchyFlattenByLmID: response is not equal:\nactual: %v\n\nexpected: %v", actualHierarchyFlatten, expectedHierarchyFlatten)
	}

	return StepStateToContext(ctx, stepState), nil
}
