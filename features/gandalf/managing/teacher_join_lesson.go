package managing

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/features/bob"
	"github.com/manabie-com/backend/internal/golibs/constants"
	pb_bob "github.com/manabie-com/backend/pkg/genproto/bob"
)

func (s *suite) teacherJoinLesson(ctx context.Context) (context.Context, error) {
	stepState := GandalfStepStateFromContext(ctx)
	ctx, err := s.bobSuite.TeacherJoinLesson(ctx)
	if err != nil {
		return GandalfStepStateToContext(ctx, stepState), err

	}
	stepState.YasuoStepState.CurrentUserGroup = "USER_GROUP_TEACHER"
	return GandalfStepStateToContext(ctx, stepState), nil
}

func (s *suite) aTeacherFromSameSchoolWithValidLesson(ctx context.Context) (context.Context, error) {
	stepState := GandalfStepStateFromContext(ctx)
	ctx, err := s.bobSuite.ATeacherFromSameSchoolWithValidLesson(ctx)
	if err != nil {
		return GandalfStepStateToContext(ctx, stepState), err

	}

	pbLiveLessons := make([]*pb_bob.EvtLesson_Lesson, 0, 1)
	pbLiveLessons = append(pbLiveLessons, &pb_bob.EvtLesson_Lesson{
		LessonId: bob.StepStateFromContext(ctx).CurrentLessonID,
	})

	msg := &pb_bob.EvtLesson{
		Message: &pb_bob.EvtLesson_CreateLessons_{
			CreateLessons: &pb_bob.EvtLesson_CreateLessons{
				Lessons: pbLiveLessons,
			},
		},
	}

	data, err := msg.Marshal()
	if err != nil {
		return ctx, err
	}

	_, err = s.jsm.PublishContext(ctx, constants.SubjectLessonCreated, data)
	if err != nil {
		return ctx, fmt.Errorf("s.jsm.PublishContext: %w", err)
	}

	return ctx, nil
}
