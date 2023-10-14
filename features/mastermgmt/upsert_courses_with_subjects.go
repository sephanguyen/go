package mastermgmt

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/manabie-com/backend/internal/bob/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/course/domain"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

const (
	AllNew   = "all_new"  // add brand new both courses and subjects
	AddNew   = "add_new"  // just add subjects with existing courses
	Modified = "modified" // both delete and add
	Delete   = "delete"   // delete all subjects
)

func (s *suite) checkUpdatedCoursesAndSubjects(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	courses := stepState.CourseProcessing
	courseIDs := sliceutils.Map(courses, func(c *domain.Course) string {
		return c.CourseID
	})

	actualCourses, err := s.getCoursesWithSubjects(ctx, courseIDs)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	for _, expectedCourse := range courses {
		v, ok := actualCourses[expectedCourse.CourseID]
		if !ok {
			return StepStateToContext(ctx, stepState),
				fmt.Errorf("expected courses is not correct.\nexpected:%+v\nactual:%+v",
					courses,
					actualCourses)
		}
		actualSubjects := v.SubjectIDs
		expectedSubjects := expectedCourse.SubjectIDs

		sort.Strings(expectedSubjects)
		sort.Strings(actualSubjects)
		if strings.Join(expectedSubjects, "|") != strings.Join(actualSubjects, "|") {
			return StepStateToContext(ctx, stepState),
				fmt.Errorf("expected course's subject is not correct.\nexpected:%+v\nactual:%+v",
					strings.Join(expectedSubjects, "|"),
					strings.Join(actualSubjects, "|"))
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) upsertCoursesWithSubjects(ctx context.Context, mod string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	courses, err := s.prepareCourseAndSubjects(ctx, mod)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("can not prepare course and subjects: %v", err.Error())
	}
	stepState.CourseProcessing = courses

	coursePayload := sliceutils.Map(courses, func(c *domain.Course) *mpb.UpsertCoursesRequest_Course {
		return &mpb.UpsertCoursesRequest_Course{
			Name:       c.Name,
			Id:         c.CourseID,
			SubjectIds: c.SubjectIDs,
		}
	})

	stepState.Request = &mpb.UpsertCoursesRequest{
		Courses: coursePayload,
	}
	stepState.Response, stepState.ResponseErr = mpb.NewMasterDataCourseServiceClient(s.MasterMgmtConn).
		UpsertCourses(contextWithToken(s, ctx), stepState.Request.(*mpb.UpsertCoursesRequest))

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) seedSubjects(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	count := 10
	subjectIDs := make([]string, count)
	for j := 0; j < count; j++ {
		timeID := idutil.ULIDNow()
		iStmt := `INSERT INTO subject (
					subject_id,
					name,
					created_at,
					updated_at)
				VALUES ($1, $2, NOW(), NOW())`
		_, err := s.BobDB.Exec(ctx, iStmt, timeID, "subject_"+timeID)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("cannot seed subjects, err: %s", err)
		}
		subjectIDs[j] = timeID
	}
	stepState.SubjectIDs = subjectIDs

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) prepareCourseAndSubjects(ctx context.Context, mod string) ([]*domain.Course, error) {
	var courses []*domain.Course
	courseIDs := s.generateULIDs(4)
	id1, id2, id3, id4 := courseIDs[0], courseIDs[1], courseIDs[2], courseIDs[3]
	subjectIDs := s.SubjectIDs

	courses = []*domain.Course{
		{
			CourseID:   id1,
			Name:       "course_" + id1,
			SubjectIDs: subjectIDs[:2], // first two elem
		},
		{
			CourseID:   id2,
			Name:       "course_" + id2,
			SubjectIDs: subjectIDs[:1], // first elem
		},
	}

	err := s.seedCourseWithSubjects(ctx, courses)
	if err != nil {
		return nil, err
	}

	switch mod {
	case AllNew:
		{
			// all new data
			courses = []*domain.Course{
				{
					CourseID:   id3,
					Name:       "course_" + id3,
					SubjectIDs: subjectIDs[2:4], // index 2,3
				},
				{
					CourseID:   id4,
					Name:       "course_" + id4,
					SubjectIDs: subjectIDs[4:5], // index 4
				},
			}
		}
	case AddNew:
		{
			// just add new subjects
			courses = []*domain.Course{
				{
					CourseID:   id1,
					Name:       "course_" + id1,
					SubjectIDs: subjectIDs[:3], // first three elem
				},
				{
					CourseID:   id2,
					Name:       "course_" + id2,
					SubjectIDs: subjectIDs[:2], // first two elem
				},
			}
		}
	case Modified:
		{
			courses = []*domain.Course{
				{
					CourseID: id1,
					Name:     "course_" + id1,
					SubjectIDs: []string{
						subjectIDs[0], // the old
						subjectIDs[5], // the new
					}, // first two elem
				},
				{
					CourseID: id2,
					Name:     "course_" + id2,
					SubjectIDs: []string{
						subjectIDs[1], // the old
						subjectIDs[5], // add
					}, // first two elem, // first elem
				},
			}
		}
	case Delete:
		{
			courses = []*domain.Course{
				{
					CourseID:   id1,
					Name:       "course_" + id1,
					SubjectIDs: []string{},
				},
				{
					CourseID:   id2,
					Name:       "course_" + id2,
					SubjectIDs: []string{},
				},
			}
		}
	}
	return courses, nil
}

func (s *suite) seedCourseWithSubjects(ctx context.Context, courses []*domain.Course) error {
	err := database.ExecInTx(ctx, s.BobDBTrace, func(ctx context.Context, tx pgx.Tx) error {
		for _, course := range courses {
			stmt := `INSERT INTO courses (course_id, name, school_id, is_archived, created_at, updated_at) 
					VALUES ($1,$2,$3,false,now(),now())`
			_, err := s.BobDBTrace.Exec(ctx,
				stmt,
				course.CourseID,
				course.Name,
				constants.ManabieSchool,
			)
			if err != nil {
				return err
			}

			for _, subject := range course.SubjectIDs {
				// insert subjects
				subjectStmt := `INSERT INTO course_subject (course_id, subject_id, created_at, updated_at) 
								VALUES ($1,$2,now(),now())`
				_, err = s.BobDBTrace.Exec(ctx,
					subjectStmt,
					course.CourseID,
					subject,
				)
				if err != nil {
					return err
				}
			}
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("cannot seed courses, %v", err)
	}
	return nil
}

func (s *suite) getCoursesWithSubjects(ctx context.Context, courseIDs []string) (map[string]*domain.Course, error) {
	courses := make(map[string]*domain.Course)
	stmt :=
		`
		SELECT c.course_id, c.name,
				CASE
					WHEN COUNT(s.subject_id) = 0 THEN ARRAY[]::TEXT[]
					ELSE array_agg(s.subject_id)
				END as subjects
		FROM courses c
		LEFT JOIN course_subject cs ON c.course_id = cs.course_id  AND cs.deleted_at IS NULL
		LEFT JOIN subject s ON cs.subject_id = s.subject_id
		WHERE c.course_id = ANY($1)
		AND c.deleted_at IS NULL
		GROUP BY c.course_id
		`
	rows, err := s.BobDBTrace.Query(
		ctx,
		stmt,
		courseIDs,
	)
	if err != nil {
		return nil, errors.Wrap(err, "query course")
	}

	defer rows.Close()
	for rows.Next() {
		e := &domain.Course{}
		err := rows.Scan(
			&e.CourseID,
			&e.Name,
			&e.SubjectIDs,
		)
		if err != nil {
			return nil, errors.WithMessage(err, "rows.Scan course")
		}
		courses[e.CourseID] = e
	}
	return courses, nil
}
