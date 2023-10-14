package bob

import (
	"context"
	"time"

	entities_bob "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"

	"github.com/lestrrat-go/jwx/jwt"
	"go.uber.org/multierr"
)

func (s *suite) aStudentWithValidLesson(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	s.aListOfCoursesAreExistedInDBOf(ctx, "above teacher")

	var courseID string = "course-teacher-1"

	lesson, err := s.generateValidLesson(courseID, stepState.CurrentTeacherID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}
	_, err = database.Insert(ctx, lesson, s.DB.Exec)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}

	t, _ := jwt.ParseString(stepState.AuthToken)

	now := time.Now()
	lessonMember := &entities_bob.LessonMember{}
	database.AllNullEntity(lessonMember)
	err = multierr.Combine(
		lessonMember.LessonID.Set(lesson.LessonID.String),
		lessonMember.UserID.Set(t.Subject()),
		lessonMember.CreatedAt.Set(now),
		lessonMember.UpdatedAt.Set(now),
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}

	_, err = database.Insert(ctx, lessonMember, s.DB.Exec)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}

	stepState.CurrentLessonID = lesson.LessonID.String
	stepState.CurrentCourseID = lesson.CourseID.String
	stepState.CurrentLessonGroupID = lesson.LessonGroupID.String
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aStudentWithValidLessonWhichHasNoRoomID(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	s.aListOfCoursesAreExistedInDBOf(ctx, "above teacher")

	var courseID string = "course-teacher-1"

	lesson, err := s.generateValidLesson(courseID, stepState.CurrentTeacherID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}
	_, err = database.Insert(ctx, lesson, s.DB.Exec)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}

	t, _ := jwt.ParseString(stepState.AuthToken)

	now := time.Now()
	lessonMember := &entities_bob.LessonMember{}
	database.AllNullEntity(lessonMember)
	err = multierr.Combine(
		lessonMember.LessonID.Set(lesson.LessonID.String),
		lessonMember.UserID.Set(t.Subject()),
		lessonMember.CreatedAt.Set(now),
		lessonMember.UpdatedAt.Set(now),
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	_, err = database.Insert(ctx, lessonMember, s.DB.Exec)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}

	stepState.CurrentLessonID = lesson.LessonID.String
	stepState.CurrentCourseID = lesson.CourseID.String
	stepState.CurrentLessonGroupID = lesson.LessonGroupID.String

	// remove room id
	sql := "UPDATE lessons SET room_id = NULL WHERE lesson_id = $2"
	_, err = s.DB.Exec(ctx, sql, &lesson.LessonID)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) AStudentWithValidLesson(ctx context.Context) (context.Context, error) {
	return s.aStudentWithValidLesson(ctx)
}

func (s *suite) studentRetrieveStreamToken(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.Response, stepState.ResponseErr = pb.NewClassClient(s.Conn).StudentRetrieveStreamToken(contextWithToken(s, ctx), &pb.StudentRetrieveStreamTokenRequest{
		LessonId: stepState.CurrentLessonID,
	})
	return StepStateToContext(ctx, stepState), nil
}
