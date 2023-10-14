package yasuo

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/bob/entities"
	repositories_bob "github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/i18n"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/yasuo/constant"
	bobpb "github.com/manabie-com/backend/pkg/genproto/bob"
	pb "github.com/manabie-com/backend/pkg/genproto/yasuo"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"go.uber.org/multierr"
)

func (s *suite) getExamplePbCourses(ctx context.Context) map[string]*pb.UpsertCoursesRequest_Course {
	return map[string]*pb.UpsertCoursesRequest_Course{
		"valid":           s.genPbCourse(ctx, "valid"+s.Random, "course 1", true),
		"valid2":          s.genPbCourse(ctx, "valid2"+s.Random, "course 2", true),
		"":                s.genPbCourse(ctx, "", "course-invalid-id", true),
		"invalidName":     s.genPbCourse(ctx, "invalidName"+s.Random, "course-invalid-name", true),
		"differentSchool": s.genPbCourse(ctx, "differentSchool"+s.Random, "course-different-school", false),
	}
}

func (s *suite) genPbCourse(ctx context.Context, id, name string, isSameSchool bool) *pb.UpsertCoursesRequest_Course {
	stepState := StepStateFromContext(ctx)
	var schoolID int32

	if isSameSchool {
		schoolID = stepState.CurrentSchoolID
	} else {
		schoolID = rand.Int31()
	}

	r := &pb.UpsertCoursesRequest_Course{
		Id:       id,
		Name:     name,
		Country:  bobpb.COUNTRY_MASTER,
		Subject:  bobpb.SUBJECT_ENGLISH,
		Grade:    "Grade 12",
		SchoolId: schoolID,
		Icon:     "link-icon",
	}
	return r
}

func (s *suite) generateChapter(ctx context.Context, country, subject, bookID string, grade, schoolID int) (*entities.Chapter, error) {
	chapter1 := &entities.Chapter{}
	database.AllNullEntity(chapter1)

	err := multierr.Combine(
		chapter1.ID.Set("book-chapter-"+idutil.ULIDNow()),
		chapter1.Name.Set("book-chapter-name-"+bookID),
		chapter1.Country.Set(country),
		chapter1.Subject.Set(subject),
		chapter1.Grade.Set(grade),
		chapter1.DisplayOrder.Set(1),
		chapter1.SchoolID.Set(schoolID),
		chapter1.UpdatedAt.Set(time.Now()),
		chapter1.CreatedAt.Set(time.Now()),
		chapter1.DeletedAt.Set(nil),
		chapter1.CurrentTopicDisplayOrder.Set(0),
	)
	if err != nil {
		return nil, err
	}
	cmdTag, err := database.Insert(ctx, chapter1, s.EurekaDB.Exec)
	if err != nil {
		return nil, fmt.Errorf("database.Insert chapter1: %w", err)
	}
	if cmdTag.RowsAffected() != 1 {
		if err != nil {
			return nil, fmt.Errorf("database.Insert chapter1: %w", repositories_bob.ErrUnAffected)
		}
	}

	return chapter1, nil
}

func (s *suite) generateBook(ctx context.Context, country, subject string, grade, schoolID int, isDelete bool) (*entities.Book, error) {
	now := time.Now()

	book := &entities.Book{}
	database.AllNullEntity(book)
	err := multierr.Combine(
		book.ID.Set(idutil.ULIDNow()),
		book.Country.Set(country),
		book.SchoolID.Set(schoolID),
		book.Subject.Set(subject),
		book.Grade.Set(grade),
		book.Name.Set("course-book"),
		book.CreatedAt.Set(now),
		book.UpdatedAt.Set(now),
		book.CurrentChapterDisplayOrder.Set(0),
	)
	if isDelete {
		err = multierr.Append(err, book.DeletedAt.Set(now))
	}
	if err != nil {
		return nil, err
	}

	cmdTag, err := database.Insert(ctx, book, s.EurekaDB.Exec)
	if err != nil {
		return nil, fmt.Errorf("database.Insert book: %w", err)
	}
	if cmdTag.RowsAffected() != 1 {
		if err != nil {
			return nil, fmt.Errorf("database.Insert book: %w", repositories_bob.ErrUnAffected)
		}
	}
	_, err = s.generateChapter(ctx, country, subject, book.ID.String, grade, schoolID)
	if err != nil {
		return nil, err
	}

	return book, nil
}

func (s *suite) listOfBooksInOurDB(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	bookIDs := []string{}
	for i := 1; i < 5; i++ {
		schoolID := stepState.CurrentSchoolID
		if schoolID == 0 {
			schoolID = constants.ManabieSchool
		}

		book, err := s.generateBook(ctx, bobpb.COUNTRY_VN.String(), bobpb.SUBJECT_BIOLOGY.String(), 8, int(schoolID), false)
		if err != nil {
			return ctx, err
		}

		bookIDs = append(bookIDs, book.ID.String)

		// generate book is deleted
		_, err = s.generateBook(ctx, bobpb.COUNTRY_VN.String(), bobpb.SUBJECT_MATHS.String(), 10, int(schoolID), true)
		if err != nil {
			return ctx, err
		}
	}

	stepState.CurrentBookIDs = bookIDs

	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) getCourseBasedOnState(ctx context.Context, id string) (e *pb.UpsertCoursesRequest_Course, err error) {
	stepState := StepStateFromContext(ctx)
	courses := s.getExamplePbCourses(ctx)

	course, ok := courses[id]
	if !ok {
		course = s.genPbCourse(ctx, id, "specific courses id: "+id, true)
		if strings.Contains(id, "course-book") {
			if _, err := epb.NewCourseModifierServiceClient(s.EurekaConn).AddBooks(contextWithToken(s, ctx), &epb.AddBooksRequest{
				BookIds:  stepState.CurrentBookIDs,
				CourseId: id,
			}); err != nil {
				return nil, fmt.Errorf("unable to add books: %w", err)
			}
		}
		courses[id] = course
	}

	return course, nil
}

func (s *suite) aUpsertCourseRequestWithData(ctx context.Context, state string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req, ok := stepState.Request.(*pb.UpsertCoursesRequest)
	if !ok { // already have courses request before, add new invalid to req
		req = &pb.UpsertCoursesRequest{
			Courses: []*pb.UpsertCoursesRequest_Course{},
		}
	}

	course, err := s.getCourseBasedOnState(ctx, state)
	if err != nil {
		return ctx, errors.New("cannot get course based on state")
	}
	req.Courses = append(req.Courses, course)

	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) adminUpsertCourses(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.Response, stepState.ResponseErr = pb.NewCourseServiceClient(s.Conn).UpsertCourses(contextWithToken(s, ctx), stepState.Request.(*pb.UpsertCoursesRequest))

	return StepStateToContext(ctx, stepState), stepState.ResponseErr
}

func (s *suite) userUpsertCourses(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.Response, stepState.ResponseErr = pb.NewCourseServiceClient(s.Conn).UpsertCourses(contextWithToken(s, ctx), stepState.Request.(*pb.UpsertCoursesRequest))

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) coursesUpsertedInDBMatchWithRequest(ctx context.Context, state string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	r := &repositories_bob.CourseRepo{}
	req := stepState.Request.(*pb.UpsertCoursesRequest)
	courseIDs := []string{}
	courseData := req.Courses

	for _, course := range req.Courses {
		if course.Id != "" {
			courseIDs = append(courseIDs, course.Id)
		}
	}

	coursesInDB, _ := r.FindByIDs(ctx, s.DBTrace, database.TextArray(courseIDs))

	if len(coursesInDB) == 0 && state == "invalid" {
		return StepStateToContext(ctx, stepState), nil
	}

	for index, courseID := range courseIDs {
		courseInDB, ok := coursesInDB[database.Text(courseID)]
		if !ok {
			return ctx, errors.New("course not found")
		}

		if courseInDB.Name.String != courseData[index].Name {
			return ctx, errors.New("not same name")
		}

		if courseInDB.Icon.String != courseData[index].Icon {
			return ctx, errors.New("not same icon")
		}

		if courseInDB.Subject.String != courseData[index].Subject.String() {
			return ctx, errors.New("not same subject")
		}

		if courseInDB.Country.String != courseData[index].Country.String() {
			return ctx, errors.New("not same country")
		}

		grade, _ := i18n.ConvertStringGradeToInt(courseData[index].Country, courseData[index].Grade)
		if int32(courseInDB.Grade.Int) != int32(grade) {
			return ctx, errors.New("not same grade")
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) courseBookUpsertInDBMatchWithRequest(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	r := &repositories_bob.CourseBookRepo{}
	req := stepState.Request.(*pb.UpsertCoursesRequest)
	courseIDs := []string{}
	bookIDs := stepState.CurrentBookIDs

	for _, course := range req.Courses {
		if course.Id != "" {
			courseIDs = append(courseIDs, course.Id)
		}
	}

	mapBookIDByCourseID, err := r.FindByCourseIDs(ctx, s.EurekaDB, courseIDs)
	if len(bookIDs) == 0 || len(courseIDs) == 0 || err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("not found course in course book upsert")
	}

	for _, courseID := range courseIDs {
		bookList, ok := mapBookIDByCourseID[courseID]
		if !ok {
			return StepStateToContext(ctx, stepState), fmt.Errorf("not found course in course book upsert")
		}

		sort.Strings(bookList)
		sort.Strings(bookIDs)
		if !reflect.DeepEqual(bookList, bookIDs) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("not found book in course book upsert")
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) activityLogsOfEventIs(ctx context.Context, event string, eventResult string) error {
	query := `SELECT COUNT(*) FROM activity_logs WHERE action_type = $1`
	actionType := fmt.Sprintf("/manabie.yasuo.CourseService/%s_%s", event, eventResult)
	var count int64

	row := s.DBTrace.QueryRow(ctx, query, &actionType)
	err := row.Scan(&count)
	if err != nil {
		return err
	}

	if count == 0 {
		return fmt.Errorf("expected number of activity logs not match %s", actionType)
	}
	return nil
}

func (s *suite) anExistedCourseWithID(ctx context.Context, id string) (context.Context, error) {
	ctx, err1 := s.aUpsertCourseRequestWithData(ctx, id)
	ctx, err2 := s.userUpsertCourses(ctx)
	err := multierr.Combine(err1, err2)

	return ctx, err
}

func (s *suite) userSendDuplicateOfExistedCourse(ctx context.Context) (context.Context, error) {
	ctx, err1 := s.anExistedCourseWithID(ctx, "valid")
	ctx, err2 := s.aUpsertCourseRequestWithData(ctx, "valid")
	ctx, err3 := s.aUpsertCourseRequestWithData(ctx, "valid")

	err := multierr.Combine(err1, err2, err3)

	return ctx, err
}
func getCourseReq(course *pb.UpsertCoursesRequest_Course) *pb.UpsertCoursesRequest {
	return &pb.UpsertCoursesRequest{
		Courses: []*pb.UpsertCoursesRequest_Course{course},
	}
}
func getCourseWithMissingFields(fields string) *pb.UpsertCoursesRequest {
	course := &pb.UpsertCoursesRequest_Course{
		Id:           idutil.ULIDNow(),
		Name:         "name",
		Country:      bobpb.COUNTRY_VN,
		Subject:      bobpb.SUBJECT_BIOLOGY,
		DisplayOrder: 1,
		SchoolId:     constant.ManabieSchool,
		BookIds:      []string{"book-id"},
		Grade:        "Lớp 1",
	}
	switch fields {
	case Name:
		course.Name = ""
		return getCourseReq(course)
	case SchoolID:
		course.SchoolId = 0
		return getCourseReq(course)
	case Country:
		course.Country = bobpb.COUNTRY_NONE
		return getCourseReq(course)
	case Subject:
		course.Subject = bobpb.SUBJECT_NONE
		return getCourseReq(course)
	case Grade:
		course.Grade = ""
		return getCourseReq(course)
	case CountryAndGrade:
		course.Country = bobpb.COUNTRY_NONE
		course.Grade = ""
		return getCourseReq(course)
	case Book:
		course.BookIds = nil
		return getCourseReq(course)
	case Chapter:
		course.ChapterIds = nil
		return getCourseReq(course)
	case DisplayOrder:
		course.DisplayOrder = 0
		return getCourseReq(course)
	case All:
		return getCourseReq(&pb.UpsertCoursesRequest_Course{})
	case None:
		return getCourseReq(course)
	default:
		return nil
	}
}

func (s *suite) userUpsertCoursesWithSomeAreMissing(ctx context.Context, fields string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.Response, stepState.ResponseErr = pb.NewCourseServiceClient(s.Conn).UpsertCourses(contextWithToken(s, ctx), getCourseWithMissingFields(fields))
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) yasuoMustStoreCourse(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := stepState.Request.(*pb.UpsertCoursesRequest)
	courseRepo := &repositories_bob.CourseRepo{}
	courseIDs := []string{}

	for _, v := range req.Courses {
		courseIDs = append(courseIDs, v.Id)
	}

	courses, err := courseRepo.FindByIDs(ctx, s.DBTrace, database.TextArray(courseIDs))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("courseRepo.FindByIDs: %w", err)
	}

	if len(courses) != len(courseIDs) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("result course not match expected %d, got %d", len(courseIDs), len(courses))
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theAdminUpsertCourses(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	id := idutil.ULIDNow()
	if ctx, err := s.aUpsertCourseRequestWithData(ctx, fmt.Sprintf("course-book-%s", id)); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if ctx, err := s.adminUpsertCourses(ctx); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) ourSystemHasToStoreUpsertCourseWithBookCorrectly(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if ctx, err := s.coursesUpsertedInDBMatchWithRequest(ctx, "valid"); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if ctx, err := s.courseBookUpsertInDBMatchWithRequest(ctx); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}
