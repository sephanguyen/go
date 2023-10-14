package bob

import (
	"context"

	pb "github.com/manabie-com/backend/pkg/genproto/bob"
)

func (s *suite) aStudentInLiveLessonBackground(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	s.aTeacherAndAClassWithSomeStudents(ctx)
	s.aListOfCoursesAreExistedInDBOf(ctx, "above teacher")
	s.aStudentWithValidLesson(ctx)
	s.studentJoinLesson(ctx)
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) AStudentInLiveLessonBackground(ctx context.Context) (context.Context, error) {
	return s.aStudentInLiveLessonBackground(ctx)
}
func (s *suite) studentLeaveLesson(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.Response, stepState.ResponseErr = pb.NewClassClient(s.Conn).LeaveLesson(contextWithToken(s, ctx), &pb.LeaveLessonRequest{
		LessonId: stepState.CurrentLessonID,
		UserId:   stepState.CurrentUserID,
	})
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) StudentLeaveLesson(ctx context.Context) (context.Context, error) {
	return s.studentLeaveLesson(ctx)
}
func (s *suite) studentLeaveLessonForOtherStudent(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.Response, stepState.ResponseErr = pb.NewClassClient(s.Conn).LeaveLesson(contextWithToken(s, ctx), &pb.LeaveLessonRequest{
		LessonId: stepState.CurrentLessonID,
		UserId:   stepState.StudentIds[0],
	})
	return StepStateToContext(ctx, stepState), nil
}
