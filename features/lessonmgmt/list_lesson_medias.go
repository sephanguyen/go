package lessonmgmt

import (
	"context"
	"fmt"
	"sort"

	"github.com/manabie-com/backend/features/helper"
	"github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"
)

func (s *Suite) teacherGetMediaOfLesson(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	resp, err := lpb.NewLessonReaderServiceClient(s.LessonMgmtConn).RetrieveLessonMedias(helper.GRPCContext(ctx, "token", stepState.AuthToken), &lpb.ListLessonMediasRequest{
		LessonId: stepState.CurrentLessonID,
		Paging: &cpb.Paging{
			Limit: 1,
		},
	})
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.MediaItems = append(stepState.MediaItems, resp.Items...)

	nextPage := resp.NextPage
	for len(resp.Items) != 0 {
		resp, err = lpb.NewLessonReaderServiceClient(s.LessonMgmtConn).RetrieveLessonMedias(helper.GRPCContext(ctx, "token", stepState.AuthToken), &lpb.ListLessonMediasRequest{
			LessonId: stepState.CurrentLessonID,
			Paging:   nextPage,
		})
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		nextPage = resp.NextPage
		stepState.MediaItems = append(stepState.MediaItems, resp.Items...)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) theListOfMediaMatchWithResponseMedias(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if len(stepState.MediaItems) != len(stepState.MediaIDs) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expect %v media item but got %v", len(stepState.MediaIDs), len(stepState.MediaItems))
	}

	mediaRepo := repositories.MediaRepo{}
	expectMedia, err := mediaRepo.RetrieveByIDs(ctx, s.BobDBTrace, database.TextArray(stepState.MediaIDs))
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	sort.SliceStable(expectMedia, func(i, j int) bool {
		return expectMedia[i].MediaID.String > expectMedia[j].MediaID.String
	})
	for i, item := range stepState.MediaItems {
		if item.MediaId != expectMedia[i].MediaID.String {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expect %v media id but got %v", stepState.MediaIDs[i], item.MediaId)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}
