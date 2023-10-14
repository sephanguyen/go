package course_statistical

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/features/syllabus/utils"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/yasuo/constant"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"
)

func (s *Suite) userGetListTagByStudentInCourse(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	ctx = s.AuthHelper.SignedCtx(ctx, stepState.Token)

	stepState.Response, stepState.ResponseErr = sspb.NewStatisticsClient(s.EurekaConn).ListTagByStudentInCourse(ctx, &sspb.ListTagByStudentInCourseRequest{
		CourseId: stepState.CourseID,
	})

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userCreatesTaggedUser(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	for _, student := range stepState.Students {
		userTags := []struct {
			TagID   string
			TagName string
		}{
			{
				TagID:   idutil.ULIDNow(),
				TagName: "tag-name-A",
			},
			{
				TagID:   idutil.ULIDNow(),
				TagName: "tag-name-B",
			},
		}

		for _, tag := range userTags {
			stmtTag := `INSERT INTO public.user_tag (user_tag_id, user_tag_name, user_tag_type, is_archived, user_tag_partner_id, resource_path, created_at, updated_at)
					VALUES ($1, $2, $3, $4, $5, $6, now(), now()) ON CONFLICT DO NOTHING`
			if _, err := s.EurekaDB.Exec(ctx, stmtTag, tag.TagID, tag.TagName, "tag_type", false, idutil.ULIDNow(), fmt.Sprintf("%d", constant.ManabieSchool)); err != nil {
				return nil, err
			}

			stmtTU := `INSERT INTO public.tagged_user (user_id, tag_id, resource_path, created_at, updated_at)
					VALUES ($1, $2, $3, now(), now()) ON CONFLICT DO NOTHING`
			if _, err := s.EurekaDB.Exec(ctx, stmtTU, student.ID, tag.TagID, fmt.Sprintf("%d", constant.ManabieSchool)); err != nil {
				return nil, err
			}

			stepState.TagIDs = append(stepState.TagIDs, tag.TagID)
		}
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) ourSystemMustReturnsListTagsCorrectly(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	resp := stepState.Response.(*sspb.ListTagByStudentInCourseResponse)
	for i, got := range resp.StudentTags {
		if got.TagId != stepState.TagIDs[i] {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("wrong studentTagID expected: %v, got: %v", got.TagId, stepState.TagIDs[i])
		}
	}

	return utils.StepStateToContext(ctx, stepState), nil
}
