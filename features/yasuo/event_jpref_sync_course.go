package yasuo

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	bob_entities "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/bob/repositories"
	enigma_entites "github.com/manabie-com/backend/internal/enigma/entities"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/yasuo/entities"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"

	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
	"golang.org/x/exp/slices"
	"google.golang.org/protobuf/proto"
)

func (s *suite) toCourseSyncMsg(ctx context.Context, course, actionKind string, total int) (context.Context, []*npb.EventMasterRegistration_Course, error) {
	stepState := StepStateFromContext(ctx)
	// reset courseIds
	stepState.CourseIds = []string{}

	courses := []*npb.EventMasterRegistration_Course{}
	statuses := []cpb.CourseStatus{
		cpb.CourseStatus_COURSE_STATUS_ACTIVE,
		cpb.CourseStatus_COURSE_STATUS_INACTIVE,
	}

	switch course {
	case "new course":
		for i := 0; i < total; i++ {
			courses = append(courses, &npb.EventMasterRegistration_Course{
				ActionKind: npb.ActionKind(npb.ActionKind_value[actionKind]),
				CourseId:   idutil.ULIDNow(),
				CourseName: "course name " + idutil.ULIDNow(),
				Status:     statuses[rand.Intn(len(statuses))],
			})
		}
	case "existed course":
		for i := 0; i < total; i++ {
			course := &entities.Course{}
			now := time.Now()
			database.AllNullEntity(course)
			err := multierr.Combine(
				course.ID.Set(idutil.ULIDNow()),
				course.Name.Set("course name"+idutil.ULIDNow()),
				course.SchoolID.Set(constants.JPREPSchool),
				course.Country.Set(database.Text("COUNTRY_JP")),
				course.DisplayOrder.Set(database.Int4(1)),
				course.Grade.Set(0),
				course.CreatedAt.Set(now),
				course.UpdatedAt.Set(now),
			)
			stepState.CourseIds = append(stepState.CourseIds, course.ID.String)
			if err != nil {
				return StepStateToContext(ctx, stepState), nil, fmt.Errorf("err set course: %w", err)
			}

			_, err = database.Insert(ctx, course, s.DBTrace.Exec)
			if err != nil {
				return StepStateToContext(ctx, stepState), nil, fmt.Errorf("err Insert: %w", err)
			}

			courses = append(courses, &npb.EventMasterRegistration_Course{
				ActionKind: npb.ActionKind(npb.ActionKind_value[actionKind]),
				CourseId:   course.ID.String,
				CourseName: "course name " + idutil.ULIDNow(),
				Status:     statuses[rand.Intn(len(statuses))],
			})
		}
	}

	return StepStateToContext(ctx, stepState), courses, nil
}

func (s *suite) jprepSyncCoursesWithActionAndCourseWithActionToOurSystem(ctx context.Context, numberOfNewCourse, newCourseAction, numberOfExistedCourse, existedCourseAction string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	total, err := strconv.Atoi(numberOfNewCourse)
	if err != nil {
		return ctx, err
	}
	stepState.RequestSentAt = time.Now()

	ctx, newCourses, err := s.toCourseSyncMsg(ctx, "new course", newCourseAction, total)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	total, err = strconv.Atoi(numberOfExistedCourse)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	ctx, existedCourses, err := s.toCourseSyncMsg(ctx, "existed course", existedCourseAction, total)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	courses := append(newCourses, existedCourses...)
	stepState.Request = courses

	signature := idutil.ULIDNow()
	ctx, err = s.createPartnerSyncDataLog(ctx, signature, 0)
	if err != nil {
		return ctx, fmt.Errorf("create partner sync data log error: %w", err)
	}
	ctx, err = s.createLogSyncDataSplit(ctx, string(enigma_entites.KindCourse))
	if err != nil {
		return ctx, fmt.Errorf("create partner sync data log split error: %w", err)
	}

	req := &npb.EventMasterRegistration{
		RawPayload: []byte("{}"),
		Signature:  signature,
		Courses:    courses,
		LogId:      stepState.PartnerSyncDataLogSplitId,
	}
	data, _ := proto.Marshal(req)
	_, err = s.JSM.PublishContext(ctx, constants.SubjectSyncMasterRegistration, data)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("Publish: %w", err)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theseCoursesMustBeStoreInOurSystem(ctx context.Context) (context.Context, error) {
	time.Sleep(time.Second)

	stepState := StepStateFromContext(ctx)

	courseRepo := &repositories.CourseRepo{}
	courseAccessPathRepo := &repositories.CourseAccessPathRepo{}
	courseClassRepo := &repositories.CourseClassRepo{}

	courses := stepState.Request.([]*npb.EventMasterRegistration_Course)
	for _, c := range courses {
		course, err := courseRepo.FindByID(ctx, s.DBTrace, database.Text(c.CourseId))
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return ctx, fmt.Errorf("err find Course: %w", err)
		}
		courseAccessPath, err := courseAccessPathRepo.FindByCourseIDs(ctx, s.DBTrace, []string{c.CourseId})
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return ctx, fmt.Errorf("err find course access path: %w", err)
		}
		switch c.ActionKind {
		case npb.ActionKind_ACTION_KIND_UPSERTED:
			if course == nil {
				return ctx, fmt.Errorf("course does not existed %s", c.CourseId)
			}

			if course.Name.String != c.CourseName {
				return ctx, fmt.Errorf("course name does not match, expected: %s, got: %s", c.CourseName, course.Name.String)
			}

			if course.Country.String != "COUNTRY_JP" {
				return ctx, fmt.Errorf("course name does not match, expected: %s, got: %s", "COUNTRY_JP", course.Country.String)
			}

			if course.Grade.Int != 0 {
				return ctx, fmt.Errorf("course grade does not match, expected: %d, got: %d", 0, course.Grade.Int)
			}

			if course.DisplayOrder.Int != 1 {
				return ctx, fmt.Errorf("course displayOrder does not match, expected: %d, got: %d", 1, course.DisplayOrder.Int)
			}

			if course.Status.String != c.Status.String() {
				return ctx, fmt.Errorf("course status does not match, expected: %v, got: %v", c.Status.String(), course.Status.String)
			}

			courseClasses, err := courseClassRepo.FindByCourseIDs(ctx, s.DBTrace, database.TextArray([]string{c.CourseId}))
			if err != nil {
				return ctx, err
			}

			if len(courseClasses) != 0 {
				return ctx, fmt.Errorf("all course class should not create, found %d", len(courseClasses))
			}

			cap, ok := courseAccessPath[c.CourseId]
			if !ok || !slices.Contains(cap, constants.JPREPOrgLocation) {
				return ctx, fmt.Errorf("course access path was not updated with JREP location")
			}

		case npb.ActionKind_ACTION_KIND_DELETED:
			if course != nil {
				return ctx, fmt.Errorf("course does not deleted")
			}
		}
	}

	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) someCoursesExistedInDB(ctx context.Context) (context.Context, error) {
	total := rand.Intn(10) + 1
	ctx, _, err := s.toCourseSyncMsg(ctx, "existed course", "ACTION_KIND_UPSERTED", total)
	ctx, _, err = s.toCourseSyncMsg(ctx, "existed course", "ACTION_KIND_DELETED", 1)

	return ctx, err
}
func (s *suite) someCoursesMustHaveIcon(ctx context.Context) error {
	stepState := StepStateFromContext(ctx)

	updateIconStmt := `UPDATE courses SET icon = 'default icon' WHERE course_id = ANY($1::_TEXT)`
	commandTag, err := s.DBTrace.DB.Exec(ctx, updateIconStmt, stepState.CourseIds)
	if err != nil {
		return fmt.Errorf("unable update: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("no row affected, unable update")
	}
	return nil
}
func (s *suite) aBookExistedInDB(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	now := time.Now()
	book := &bob_entities.Book{}
	database.AllNullEntity(book)
	if err := multierr.Combine(
		book.ID.Set(idutil.ULIDNow()),
		book.Name.Set(strconv.Itoa(rand.Int())),
		book.Country.Set(cpb.Country_COUNTRY_JP.String()),
		book.SchoolID.Set(constants.ManabieSchool),
		book.Subject.Set(cpb.Subject_SUBJECT_ENGLISH.String()),
		book.Grade.Set(12),
		book.CreatedAt.Set(now),
		book.UpdatedAt.Set(now),
		book.CurrentChapterDisplayOrder.Set(0),
	); err != nil {
		return ctx, err
	}

	cmdTag, err := database.Insert(ctx, book, s.DBTrace.DB.Exec)
	if err != nil {
		return ctx, fmt.Errorf("database.Insert book: %v", err)
	}
	if cmdTag.RowsAffected() != 1 {
		return ctx, fmt.Errorf("unable to insert book")
	}
	stepState.BookId = book.ID.String
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) someCoursesMustHaveBook(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.aBookExistedInDB(ctx)
	if err != nil {
		return ctx, err
	}

	coursebookRepo := &repositories.CourseBookRepo{}
	cc := make([]*bob_entities.CoursesBooks, 0, len(stepState.CourseIds))
	for _, id := range stepState.CourseIds {
		now := time.Now()

		coursebook := &bob_entities.CoursesBooks{}
		database.AllNullEntity(coursebook)
		coursebook.BookID = database.Text(stepState.BookId)
		coursebook.CourseID = database.Text(id)
		coursebook.CreatedAt.Set(now)
		coursebook.UpdatedAt.Set(now)
		cc = append(cc, coursebook)
	}

	if err := coursebookRepo.Upsert(context.Background(), s.DBTrace.DB, cc); err != nil {
		return ctx, fmt.Errorf("coursebookRepo.Upsert: %w", err)
	}

	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) jprefSyncArbitraryNumberNewCoursesAndExistedCoursesWithActionToOurSystem(ctx context.Context, arg1 string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	totalNewCourse := rand.Intn(10) + 1
	ctx, newCourses, err := s.toCourseSyncMsg(ctx, "new course", arg1, totalNewCourse)
	if err != nil {
		return ctx, err
	}
	courses := make([]*npb.EventMasterRegistration_Course, 0)
	for _, id := range stepState.CourseIds {
		courses = append(courses, &npb.EventMasterRegistration_Course{
			ActionKind: npb.ActionKind(npb.ActionKind_value[arg1]),
			CourseId:   id,
			CourseName: "course name " + idutil.ULIDNow(),
		})
	}
	courses = append(courses, newCourses...)
	stepState.Request = courses

	req := &npb.EventMasterRegistration{RawPayload: []byte("{}"), Signature: idutil.ULIDNow(), Courses: courses}
	data, _ := proto.Marshal(req)
	_, err = s.JSM.PublishContext(ctx, constants.SubjectSyncMasterRegistration, data)

	return StepStateToContext(ctx, stepState), err
}

func (s *suite) theseCoursesHaveToSaveCorrectly(ctx context.Context) (context.Context, error) {
	if ctx, err := s.theseCoursesMustBeStoreInOurSystem(ctx); err != nil {
		return ctx, err
	}
	stepState := StepStateFromContext(ctx)
	var totalIconExisted, totalCourseBookExisted int
	stmtCountIcon := `SELECT COUNT(*) FROM courses WHERE course_id = ANY($1::_TEXT) AND icon!=''`
	err := s.DBTrace.DB.QueryRow(context.Background(), stmtCountIcon, stepState.CourseIds).Scan(&totalIconExisted)
	if err != nil {
		return ctx, err
	}
	stmtCountCourseBook := `SELECT COUNT(*) FROM courses_books WHERE course_id = ANY($1::_TEXT)`
	err = s.DBTrace.DB.QueryRow(context.Background(), stmtCountCourseBook, stepState.CourseIds).Scan(&totalCourseBookExisted)
	if err != nil {
		return ctx, err
	}
	if totalIconExisted != len(stepState.CourseIds) {
		return ctx, fmt.Errorf("unexpected total icon in db, expected %d, got %d", len(stepState.CourseIds), totalIconExisted)
	}
	if totalCourseBookExisted != len(stepState.CourseIds) {
		return ctx, fmt.Errorf("unexpected total book in db, expected %d, got %d", len(stepState.CourseIds), totalCourseBookExisted)
	}
	return StepStateToContext(ctx, stepState), nil
}
