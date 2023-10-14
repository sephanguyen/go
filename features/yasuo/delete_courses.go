package yasuo

import (
	"context"
	"errors"
	"fmt"
	"time"

	entities_bob "github.com/manabie-com/backend/internal/bob/entities"
	repositories_bob "github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/yasuo/entities"
	pb "github.com/manabie-com/backend/pkg/genproto/yasuo"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
)

func (s *suite) setCourseData(schoolID int32, id, name, country, subject string, courseType string, isDeleted bool) *entities.Course {
	c := &entities.Course{}
	database.AllNullEntity(c)

	_ = multierr.Combine(
		c.ID.Set(id+s.Random),
		c.Name.Set(name+s.Random),
		c.Country.Set(country),
		c.Subject.Set(subject),
		c.DisplayOrder.Set(1),
		c.Grade.Set(12),
		c.DisplayOrder.Set(1),
		c.SchoolID.Set(schoolID),
		c.CreatedAt.Set(time.Now()),
		c.UpdatedAt.Set(time.Now()),
		c.StartDate.Set(nil),
		c.EndDate.Set(nil),
		c.TeacherIDs.Set(nil),
		c.CourseType.Set(courseType),
	)
	if isDeleted {
		_ = c.DeletedAt.Set(time.Now())
	} else {
		_ = c.DeletedAt.Set(nil)
	}

	return c
}

func (s *suite) getExampleCourses(ctx context.Context) (context.Context, []*entities.Course) {
	stepState := StepStateFromContext(ctx)

	if stepState.CurrentSchoolID == 0 {
		stepState.CurrentSchoolID = 1
	}
	courses := []*entities.Course{
		s.setCourseData(stepState.CurrentSchoolID, "course-valid-1", "Course 1 name", "COUNTRY_VN", "SUBJECT_BIOLOGY", "COURSE_TYPE_CONTENT", false),
		s.setCourseData(stepState.CurrentSchoolID, "course-valid-2", "Course 2 name", "COUNTRY_VN", "SUBJECT_MATHS", "COURSE_TYPE_CONTENT", false),
		s.setCourseData(stepState.CurrentSchoolID, "course-deleted", "Course 3 name", "COUNTRY_VN", "SUBJECT_MATHS", "COURSE_TYPE_CONTENT", true),
		s.setCourseData(stepState.CurrentSchoolID, "course-teacher-2", "Course teacher 2 name", "COUNTRY_VN", "SUBJECT_MATHS", "COURSE_TYPE_CONTENT", false),
		s.setCourseData(stepState.CurrentSchoolID, "course-teacher-deleted", "Course teacher 3 name", "COUNTRY_VN", "SUBJECT_MATHS", "COURSE_TYPE_CONTENT", true),
		s.setCourseData(stepState.CurrentSchoolID, "course-1-JP", "Course 1 JP name", "COUNTRY_JP", "SUBJECT_BIOLOGY", "COURSE_TYPE_CONTENT", false),
		s.setCourseData(stepState.CurrentSchoolID, "course-2-JP", "Course 2 JP name", "COUNTRY_JP", "SUBJECT_MATHS", "COURSE_TYPE_CONTENT", false),
		s.setCourseData(stepState.CurrentSchoolID, "course-1-SG", "Course 1 SG name", "COUNTRY_SG", "SUBJECT_BIOLOGY", "COURSE_TYPE_CONTENT", false),
		s.setCourseData(stepState.CurrentSchoolID, "course-2-SG", "Course 2 SG name", "COUNTRY_SG", "SUBJECT_MATHS", "COURSE_TYPE_CONTENT", false),
		s.setCourseData(stepState.CurrentSchoolID, "course-1-ID", "Course 1 ID name", "COUNTRY_ID", "SUBJECT_BIOLOGY", "COURSE_TYPE_CONTENT", false),
		s.setCourseData(stepState.CurrentSchoolID, "course-2-ID", "Course 2 ID name", "COUNTRY_ID", "SUBJECT_MATHS", "COURSE_TYPE_CONTENT", false),
	}

	return StepStateToContext(ctx, stepState), courses
}

func (s *suite) getExampleLiveCourses(ctx context.Context) (context.Context, []*entities.Course) {
	stepState := StepStateFromContext(ctx)

	if stepState.CurrentSchoolID == 0 {
		stepState.CurrentSchoolID = 1
	}
	courses := []*entities.Course{
		s.setCourseData(stepState.CurrentSchoolID, "live-course-valid-1", "Course 1 name", "COUNTRY_VN", "SUBJECT_BIOLOGY", "COURSE_TYPE_LIVE", false),
		s.setCourseData(stepState.CurrentSchoolID, "live-course-valid-2", "Course 2 name", "COUNTRY_VN", "SUBJECT_MATHS", "COURSE_TYPE_LIVE", false),
		s.setCourseData(stepState.CurrentSchoolID, "live-course-deleted", "Course 3 name", "COUNTRY_VN", "SUBJECT_MATHS", "COURSE_TYPE_LIVE", true),
		s.setCourseData(stepState.CurrentSchoolID, "live-course-teacher-2", "Course teacher 2 name", "COUNTRY_VN", "SUBJECT_MATHS", "COURSE_TYPE_LIVE", false),
		s.setCourseData(stepState.CurrentSchoolID, "live-course-teacher-deleted", "Course teacher 3 name", "COUNTRY_VN", "SUBJECT_MATHS", "COURSE_TYPE_LIVE", true),
		s.setCourseData(stepState.CurrentSchoolID, "live-course-1-JP", "Course 1 JP name", "COUNTRY_JP", "SUBJECT_BIOLOGY", "COURSE_TYPE_LIVE", false),
		s.setCourseData(stepState.CurrentSchoolID, "live-course-2-JP", "Course 2 JP name", "COUNTRY_JP", "SUBJECT_MATHS", "COURSE_TYPE_LIVE", false),
		s.setCourseData(stepState.CurrentSchoolID, "live-course-1-SG", "Course 1 SG name", "COUNTRY_SG", "SUBJECT_BIOLOGY", "COURSE_TYPE_LIVE", false),
		s.setCourseData(stepState.CurrentSchoolID, "live-course-2-SG", "Course 2 SG name", "COUNTRY_SG", "SUBJECT_MATHS", "COURSE_TYPE_LIVE", false),
		s.setCourseData(stepState.CurrentSchoolID, "live-course-1-ID", "Course 1 ID name", "COUNTRY_ID", "SUBJECT_BIOLOGY", "COURSE_TYPE_LIVE", false),
		s.setCourseData(stepState.CurrentSchoolID, "live-course-2-ID", "Course 2 ID name", "COUNTRY_ID", "SUBJECT_MATHS", "COURSE_TYPE_LIVE", false),
	}

	return StepStateToContext(ctx, stepState), courses
}
func (s *suite) aDeleteCourseRequestWithID(ctx context.Context, id string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	courseIDs := []string{}

	found := false
	requestCourseID := id + s.Random
	for _, each := range s.Examples.([]*entities.Course) {
		if each.ID.String == requestCourseID {
			courseIDs = append(courseIDs, each.ID.String)
			found = true
			break
		}
	}
	if !found {
		return ctx, errors.New("cannot found find the request id")
	}

	switch stepState.Request.(type) {
	case *pb.DeleteCoursesRequest:
		courseIDs = append(courseIDs, stepState.Request.(*pb.DeleteCoursesRequest).CourseIds...)
	}

	stepState.Request = &pb.DeleteCoursesRequest{CourseIds: courseIDs}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userDeleteCourses(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.Response, stepState.ResponseErr = pb.NewCourseServiceClient(s.Conn).DeleteCourses(s.signedCtx(ctx), stepState.Request.(*pb.DeleteCoursesRequest))

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) coursesIsInDB(ctx context.Context, state string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	r := &repositories_bob.CourseRepo{}

	ids, checkFn := func() ([]string, func(map[pgtype.Text]*entities_bob.Course) bool) {
		switch state {
		case "exist":
			ids := []string{}
			for _, each := range stepState.Request.(*pb.UpsertCoursesRequest).Courses {
				ids = append(ids, each.Id)
			}
			return ids, func(courses map[pgtype.Text]*entities_bob.Course) bool {
				return len(courses) > 0
			}
		case "deleted":
			return stepState.Request.(*pb.DeleteCoursesRequest).CourseIds, func(courses map[pgtype.Text]*entities_bob.Course) bool {
				return len(courses) == 0
			}
		default:
			panic("State is not valid")
		}
	}()

	records, err := r.FindByIDs(ctx, s.DBTrace, database.TextArray(ids))
	if err != nil {
		return ctx, err
	}

	if !checkFn(records) {
		return ctx, errors.New("records count is not match")
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aListOfCoursesAreExistedInDB(ctx context.Context, typeCourse string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, courses := s.getExampleCourses(ctx)
	if typeCourse == "live" {
		ctx, courses = s.getExampleLiveCourses(ctx)
	}

	for _, each := range courses {
		_, err := database.Insert(ctx, each, s.DBTrace.Exec)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}

	s.Examples = courses

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) ALiveCourse(ctx context.Context) (context.Context, error) {
	return s.aLiveCourse(ctx)
}

func (s *suite) aLiveCourse(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	authToken := stepState.AuthToken

	ctx, err1 := s.signedAsAccount(ctx, "school admin")
	ctx, err2 := s.aUpsertLiveCourseRequestWithMissing(ctx, "id")
	ctx, err3 := s.userUpsertLiveCourses(ctx)
	err := multierr.Combine(err1, err2, err3)
	if err != nil {
		return ctx, fmt.Errorf("aLiveLesson %w", err)
	}
	stepState.AuthToken = authToken

	return s.yasuoMustStoreLiveCourse(StepStateToContext(ctx, stepState))
}
