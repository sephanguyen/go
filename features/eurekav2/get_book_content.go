package eurekav2

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/eureka/v2/modules/book/domain"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v2"
)

func (s *suite) GetBookContent(ctx context.Context, validity string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx = contextWithToken(s, ctx)
	bookID := stepState.BookContent.ID
	if validity == "not-existing" {
		bookID = idutil.ULIDNow()
	}
	stepState.Response, stepState.ResponseErr = epb.
		NewBookServiceClient(s.EurekaConn).
		GetBookContent(ctx, &epb.GetBookContentRequest{
			BookId: bookID,
		})

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) ValidateBookContent(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx = contextWithToken(s, ctx)

	bookResp := stepState.Response.(*epb.GetBookContentResponse)
	if bookResp == nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("ValidateBookContent: book response is nil, error: %w", stepState.ResponseErr)
	}
	if !compareBooks(stepState.BookContent, bookResp) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("ValidateBookContent: book response is not equal:\nactual: %v\n\nexpected: %v", bookResp, stepState.BookContent)
	}

	return StepStateToContext(ctx, stepState), nil
}

func compareBooks(expected domain.Book, actual *epb.GetBookContentResponse) bool {
	if expected.ID != actual.Id || expected.Name != actual.Name || len(expected.Chapters) != len(actual.Chapters) {
		return false
	}

	for i := 0; i < len(expected.Chapters); i++ {
		if !compareChapters(expected.Chapters[i], actual.Chapters[i]) {
			return false
		}
	}

	return true
}

func compareChapters(expected domain.Chapter, actual *epb.GetBookContentResponse_Chapter) bool {
	if expected.ID != actual.Id || expected.Name != actual.Name ||
		expected.DisplayOrder != int(actual.DisplayOrder) ||
		len(expected.Topics) != len(actual.Topics) {
		return false
	}

	for i := 0; i < len(expected.Topics); i++ {
		if !compareTopics(expected.Topics[i], actual.Topics[i]) {
			return false
		}
	}

	return true
}

func compareTopics(expected domain.Topic, actual *epb.GetBookContentResponse_Topic) bool {
	if expected.ID != actual.Id || expected.Name != actual.Name ||
		expected.DisplayOrder != int(actual.DisplayOrder) ||
		expected.IconURL != actual.IconUrl ||
		len(expected.LearningMaterials) != len(actual.LearningMaterials) {
		return false
	}

	for i := 0; i < len(expected.LearningMaterials); i++ {
		if !compareLearningMaterials(expected.LearningMaterials[i], actual.LearningMaterials[i]) {
			return false
		}
	}

	return true
}

func compareLearningMaterials(expected domain.LearningMaterial, actual *epb.GetBookContentResponse_LearningMaterial) bool {
	return expected.ID == actual.Id && expected.Name == actual.Name &&
		expected.DisplayOrder == int(actual.DisplayOrder) && actual.Type == expected.Type.GetProtobufType()
}
