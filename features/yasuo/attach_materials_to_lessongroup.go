package yasuo

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/try"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"
	yasuoPb "github.com/manabie-com/backend/pkg/genproto/yasuo"
	yasuoPbV1 "github.com/manabie-com/backend/pkg/manabuf/yasuo/v1"
)

func (s *suite) AValidCourse(ctx context.Context) (context.Context, error) {
	return s.aValidCourse(ctx)
}

func (s *suite) aValidCourse(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.CurrentCourseID = idutil.ULIDNow()

	stepState.LessonGroupID = idutil.ULIDNow()

	stepState.MaterialIds = []string{"materialId-1", "materialId-2"}

	if stepState.CurrentSchoolID == 0 {
		stepState.CurrentSchoolID = constants.ManabieSchool
	}
	schoolID := stepState.CurrentSchoolID

	_, err := yasuoPb.NewCourseServiceClient(s.Conn).UpsertCourses(s.signedCtx(ctx), &yasuoPb.UpsertCoursesRequest{
		Courses: []*yasuoPb.UpsertCoursesRequest_Course{
			{
				Id:       stepState.CurrentCourseID,
				Name:     "course",
				Country:  1,
				Subject:  pb.SUBJECT_BIOLOGY,
				SchoolId: schoolID,
				Grade:    "Grade 7",
			},
		},
	})
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) attachMaterialsToLessonGroup(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	_, err := yasuoPbV1.NewCourseModifierServiceClient(s.Conn).AttachMaterialsToCourse(s.signedCtx(ctx),
		&yasuoPbV1.AttachMaterialsToCourseRequest{
			LessonGroupId: stepState.LessonGroupID,
			CourseId:      stepState.CurrentCourseID,
			MaterialIds:   stepState.MaterialIds,
		})
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) bobMustAttachMaterialToLessonGroup(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	count := 0
	query := "SELECT count(*) FROM lesson_groups WHERE lesson_group_id = $1 AND course_id = $2 AND media_ids IN ($3)"
	if err := try.Do(func(attempt int) (retry bool, err error) {
		err = s.DBTrace.QueryRow(ctx, query, stepState.LessonGroupID, stepState.CurrentCourseID, stepState.MaterialIds).Scan(&count)
		if err != nil {
			return true, err
		}
		if count != 1 {
			// time.Sleep(250 * time.Millisecond) //TODO: dont know why we should sleep here
			return false, fmt.Errorf("Bob does not attach materials to lesson group correctly")
		}

		return attempt < 5, err
	}); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}
