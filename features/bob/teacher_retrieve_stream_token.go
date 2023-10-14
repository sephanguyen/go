package bob

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	entities_bob "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/whiteboard"
	"github.com/manabie-com/backend/internal/yasuo/constant"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"

	"go.uber.org/multierr"
)

func (s *suite) generateValidLesson(courseID string, teacherID string) (*entities_bob.Lesson, error) {
	var e entities_bob.Lesson
	database.AllNullEntity(&e)
	err := multierr.Combine(e.LessonID.Set(s.newID()), e.CourseID.Set(courseID), e.TeacherID.Set(teacherID), e.CreatedAt.Set(time.Now()), e.UpdatedAt.Set(time.Now()), e.DeletedAt.Set(nil), e.RoomID.Set(s.newID()), e.LessonGroupID.Set(s.newID()), e.LessonType.Set("LESSON_TYPE_ONLINE"), e.StreamLearnerCounter.Set(database.Int4(0)), e.LearnerIds.Set(database.JSONB([]byte("{}"))))

	if err := e.Normalize(); err != nil {
		return nil, fmt.Errorf("lesson.Normalize err: %s", err)
	}

	return &e, err
}
func (s *suite) aRandomNumber(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.Random = strconv.Itoa(rand.Int())
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) aTeacherWithValidLesson(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	s.createAClassWithSchoolIdIsAndExpiredAt(ctx, "2150-12-12 23:59:59")
	ctx, err := s.signedAsAccountV2(ctx, "staff granted role teacher")
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	s.aRandomNumber(ctx)
	s.aListOfCoursesAreExistedInDBOf(ctx, "manabie")

	courseID := stepState.courseIds[0]

	_, err = s.DB.Exec(ctx, "UPDATE courses SET school_id = 2 WHERE course_id = $1", courseID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}

	lesson, err := s.generateValidLesson(courseID, stepState.CurrentTeacherID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}
	room, err := whiteboard.New(&s.Cfg.Whiteboard).CreateRoom(ctx, &whiteboard.CreateRoomRequest{
		Name:     lesson.LessonID.String,
		IsRecord: false,
	})
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}
	lesson.RoomID = database.Text(room.UUID)
	lesson.CenterID = database.Text(constants.ManabieOrgLocation)

	_, err = database.Insert(ctx, lesson, s.DB.Exec)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}
	stepState.CurrentLessonID = lesson.LessonID.String
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) teacherRetrieveStreamToken(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.Response, stepState.ResponseErr = pb.NewClassClient(s.Conn).TeacherRetrieveStreamToken(contextWithToken(s, ctx),
		&pb.TeacherRetrieveStreamTokenRequest{
			LessonId: stepState.CurrentLessonID,
		})
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) aTeacherWithInvalidLesson(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	s.createAClassWithSchoolIdIsAndExpiredAt(ctx, "2150-12-12 23:59:59")
	ctx, err := s.signedAsAccountV2(ctx, "teacher")
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	s.aRandomNumber(ctx)
	s.aListOfCoursesAreExistedInDBOf(ctx, "manabie")

	courseID := stepState.courseIds[0]

	_, err = s.DB.Exec(ctx, "UPDATE courses SET school_id = 2 WHERE course_id = $1", courseID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	fakeID := s.newID()
	lesson, err := s.generateValidLesson(courseID, fakeID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	lesson.CenterID = database.Text(constants.ManabieOrgLocation)
	_, err = database.Insert(ctx, lesson, s.DB.Exec)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}
	stepState.CurrentLessonID = lesson.LessonID.String
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) aTeacherFromSameSchoolWithValidLesson(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.CurrentSchoolID = constant.ManabieSchool
	s.createAClassWithSchoolIdIsAndExpiredAt(ctx, "2150-12-12 23:59:59")
	ctx, err := s.signedAsAccountV2(ctx, "teacher")
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	s.aListOfCoursesAreExistedInDBOf(ctx, "above teacher")
	var courseID string

	err = s.DB.QueryRow(ctx, "SELECT course_id FROM courses WHERE school_id = $1 AND deleted_at IS NULL and end_date >= NOW() LIMIT 1", stepState.CurrentSchoolID).Scan(&courseID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}

	fakeID := s.newID()
	lesson, err := s.generateValidLesson(courseID, fakeID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	lesson.CenterID = database.Text(constants.ManabieOrgLocation)
	_, err = database.Insert(ctx, lesson, s.DB.Exec)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}
	stepState.CurrentLessonID = lesson.LessonID.String
	return StepStateToContext(ctx, stepState), nil

}
func (s *suite) ATeacherFromSameSchoolWithValidLesson(ctx context.Context) (context.Context, error) {
	return s.aTeacherFromSameSchoolWithValidLesson(ctx)
}
