package eurekav2

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	pb "github.com/manabie-com/backend/pkg/manabuf/eureka/v2"

	"github.com/jackc/pgtype"
)

func (s *suite) sendAnUpdatePublishStatusLearningMaterialsRequest(ctx context.Context, agr string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx = contextWithToken(s, ctx)
	isPublishedLM := true

	if agr == "false" {
		isPublishedLM = false
	}

	req := &pb.UpdatePublishStatusLearningMaterialsRequest{
		PublishStatuses: []*pb.UpdatePublishStatusLearningMaterialsRequest_PublishStatus{{
			LearningMaterialId: stepState.LearningMaterialIDs[0],
			IsPublished:        isPublishedLM,
		},
		},
	}

	stepState.Response, stepState.ResponseErr = pb.NewLearningMaterialServiceClient(s.EurekaConn).UpdatePublishStatusLearningMaterials(
		ctx, req)

	stepState.LearningMaterialIsPublished = isPublishedLM

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) checkUpdatePublishStatusLearningMaterials(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	rawQuery := `SELECT is_published FROM learning_material WHERE learning_material_id = $1::TEXT AND deleted_at IS NULL`

	var isPublished pgtype.Bool
	err := s.EurekaDB.QueryRow(ctx, rawQuery, database.Text(stepState.LearningMaterialIDs[0])).Scan(&isPublished)

	if err != nil {
		return ctx, fmt.Errorf("unable query LearningMaterial by ID: %w", err)
	}

	if isPublished.Bool != stepState.LearningMaterialIsPublished {
		return ctx, fmt.Errorf("updatePublishLearningMaterial failed: expect IsPublished is %t but got %t", stepState.LearningMaterialIsPublished, isPublished.Bool)
	}

	return StepStateToContext(ctx, stepState), nil
}
