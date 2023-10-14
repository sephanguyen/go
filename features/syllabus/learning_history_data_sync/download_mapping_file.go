// nolint
package learning_history_data_sync

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/features/syllabus/utils"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"
)

func (s *Suite) userDownloadMappingFile(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	stepState.Response, stepState.ResponseErr = sspb.NewLearningHistoryDataSyncServiceClient(s.EurekaConn).DownloadMappingFile(s.AuthHelper.SignedCtx(ctx, stepState.Token), &sspb.DownloadMappingFileRequest{})

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) returnURLCorrectly(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	response := stepState.Response.(*sspb.DownloadMappingFileResponse)
	if response.MappingCourseIdUrl == "" {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("mapping course id file url is empty")
	}
	if response.MappingExamLoIdUrl == "" {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("mapping exam lo id file url is empty")
	}
	if response.MappingQuestionTagUrl == "" {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("mapping question tag file url is empty")
	}
	if response.FailedSyncEmailRecipientsUrl == "" {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("failed sync email recipients file url is empty")
	}

	// fmt.Println("response.MappingCourseIdUrl", response.MappingCourseIdUrl)
	// fmt.Println("response.MappingExamLoIdUrl", response.MappingExamLoIdUrl)
	// fmt.Println("response.MappingQuestionTagUrl", response.MappingQuestionTagUrl)
	// fmt.Println("response.FailedSyncEmailRecipientsUrl", response.FailedSyncEmailRecipientsUrl)

	return utils.StepStateToContext(ctx, stepState), nil
}
