package common

import (
	"context"
	"fmt"
	"strings"
	"time"

	bob_constant "github.com/manabie-com/backend/internal/bob/constants"
	entities_bob "github.com/manabie-com/backend/internal/bob/entities"
	bob_repository "github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
)

var insertCourseQuery = `INSERT INTO public."courses"
(course_id, name, country, subject, grade, display_order, deleted_at, course_type, preset_study_plan_id, start_date, end_date, school_id, teacher_ids, updated_at, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13::text[], now(), now())
ON CONFLICT ON CONSTRAINT courses_pk DO UPDATE SET name = $2, country = $3, subject = $4, grade = $5, display_order = $6,
deleted_at = $7, course_type = $8, preset_study_plan_id = $9, start_date = $10, end_date = $11, school_id = $12, teacher_ids = $13`

func (s *suite) UpsertLiveCourse(ctx context.Context, courseID string, teacherIDs []string, schoolID int32) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.CurrentSchoolID = schoolID
	ctx = StepStateToContext(ctx, stepState)
	return s.upsertLiveCourse(ctx, courseID, teacherIDs)
}

func (s *suite) upsertLiveCourse(ctx context.Context, id string, teacherIDs []string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	var course entities_bob.Course
	database.AllNullEntity(&course)
	err := multierr.Combine(
		course.ID.Set(id),
		course.Name.Set("live-course "+stepState.Random),
		course.CreatedAt.Set(time.Now()),
		course.UpdatedAt.Set(time.Now()),
		course.DeletedAt.Set(nil),
		course.Grade.Set(3),
		course.Subject.Set(pb.SUBJECT_BIOLOGY.String()),
		course.TeacherIDs.Set(teacherIDs),
		course.Country.Set(pb.COUNTRY_VN.String()),
		course.StartDate.Set(time.Now()),
		course.StartDate.Set(time.Now().Add(2*time.Hour)),
		course.SchoolID.Set(stepState.CurrentSchoolID),
		course.Status.Set("COURSE_STATUS_NONE"),
		course.EndDate.Set(time.Now().Add(2*365*24*time.Hour)),
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	_, err = database.Insert(ctx, &course, s.BobDB.Exec)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.CurrentCourseID = course.ID.String

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) AListOfCoursesAreExistedInDBOf(ctx context.Context, owner string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.aListOfValidPresetStudyPlan(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	chapterIDs := make([]string, 0)
	var teacherIDs pgtype.TextArray
	_ = teacherIDs.Set(nil)
	switch owner {
	case "JPREP whitelist":
		{
			for _, c := range JPREPWhitelistedCourses {
				args := golibs.CreateCloneOfArrInterface(c)
				_, err = s.BobDB.Exec(ctx, insertCourseQuery, args...)
				if err != nil {
					return StepStateToContext(ctx, stepState), err
				}
			}
		}
	case "JPREP blacklist":
		{
			{
				for _, c := range JPREPBlacklistedCourses {
					args := golibs.CreateCloneOfArrInterface(c)
					_, err = s.BobDB.Exec(ctx, insertCourseQuery, args...)
					if err != nil {
						return StepStateToContext(ctx, stepState), err
					}
				}
			}
		}
	default:
		{
			for _, c := range courses {
				args := golibs.CreateCloneOfArrInterface(c)
				courseID := c[0].(string) + stepState.Random
				args[0] = courseID
				if args[8] != nil {
					args[8] = args[8].(string) + stepState.Random
				}
				switch owner {
				case "manabie":
					if strings.Contains(courseID, "teacher") {
						continue
					}

					args = append(args, bob_constant.ManabieSchool)
					args = append(args, nil)
				case "above teacher":
					if !strings.Contains(courseID, "teacher") {
						continue
					}

					teacherRepo := &bob_repository.TeacherRepo{}
					teacher, err := teacherRepo.FindByID(ctx, s.BobDB, database.Text(stepState.CurrentTeacherID))
					if err != nil {
						return StepStateToContext(ctx, stepState), err
					}

					_ = teacherIDs.Set([]string{teacher.ID.String})
					args = append(args, int(teacher.SchoolIDs.Elements[0].Int))
					args = append(args, &teacherIDs)
				}

				_, err = s.BobDB.Exec(ctx, `INSERT INTO public."courses"
					(course_id, name, country, subject, grade, display_order, deleted_at, course_type, preset_study_plan_id, start_date, end_date, school_id, teacher_ids, updated_at, created_at)
					VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13::text[], now(), now())
					ON CONFLICT ON CONSTRAINT courses_pk DO UPDATE SET name = $2, country = $3, subject = $4, grade = $5, display_order = $6,
					deleted_at = $7, course_type = $8, preset_study_plan_id = $9, start_date = $10, end_date = $11, school_id = $12, teacher_ids = $13`, args...)
				if err != nil {
					return StepStateToContext(ctx, stepState), err
				}

				schoolID := args[len(args)-2].(int)

				if strings.Contains(courseID, "have-book") {
					err = s.generateBook(ctx, courseID, c[2].(string), c[3].(string), c[4].(int), schoolID, chapterIDs)
					if err != nil {
						return StepStateToContext(ctx, stepState), err
					}
				}

				if strings.Contains(courseID, "have-book-missing-subject-grade") {
					err = s.generateBook(ctx, courseID, c[2].(string), "", 0, schoolID, chapterIDs)
					if err != nil {
						return StepStateToContext(ctx, stepState), err
					}
				}

				// select all classes of this school and insert to courses_classes table.
				rows, err := s.BobDB.Query(ctx, "SELECT class_id FROM classes WHERE school_id = $1", &schoolID)
				if err != nil {
					return StepStateToContext(ctx, stepState), err
				}
				defer rows.Close()

				var classIDs []int32
				for rows.Next() {
					classID := new(pgtype.Int4)
					if err := rows.Scan(classID); err != nil {
						return StepStateToContext(ctx, stepState), err
					}
					classIDs = append(classIDs, classID.Int)
				}
				if err := rows.Err(); err != nil {
					return StepStateToContext(ctx, stepState), err
				}

				for _, classID := range classIDs {
					_, err := s.BobDB.Exec(ctx, "INSERT INTO courses_classes (course_id, class_id, created_at, updated_at) VALUES ($1, $2, now(), now()) ON CONFLICT DO NOTHING", courseID, classID)
					if err != nil {
						return StepStateToContext(ctx, stepState), err
					}
				}
				stepState.CourseIDs = append(stepState.CourseIDs, courseID)
			}
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) SomeCourse(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.CourseIDs = []string{s.newID(), s.newID()}
	for _, id := range stepState.CourseIDs {
		if ctx, err := s.upsertLiveCourse(ctx, id, stepState.TeacherIDs); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) ListCourses(ctx context.Context, limit uint32) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx = contextWithToken(s, ctx)

	stepState.ResponseErr = nil
	stepState.Courses = nil
	req := &bpb.ListCoursesRequest{
		Paging: &cpb.Paging{
			Limit: limit,
		},
		Filter: &cpb.CommonFilter{
			SchoolId: stepState.CurrentSchoolID,
		},
	}
	stepState.Request = req
	for {
		resp, err := bpb.NewCourseReaderServiceClient(s.BobConn).
			ListCourses(ctx, stepState.Request.(*bpb.ListCoursesRequest))
		if err != nil {
			stepState.ResponseErr = err
			return StepStateToContext(ctx, stepState), nil
		}
		if len(resp.Items) < 1 {
			break
		}
		if len(resp.Items) > int(req.Paging.Limit) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected total course: got: %d, want: %d", len(resp.Items), req.Paging.Limit)
		}
		for _, item := range resp.Items {
			stepState.Courses = append(stepState.Courses, item)
		}

		req.Paging = resp.NextPage
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) UserUpsertCourses(ctx context.Context, name string, locationIDs []string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := &mpb.UpsertCoursesRequest{
		Courses: []*mpb.UpsertCoursesRequest_Course{
			{
				Id:           idutil.ULIDNow(),
				Name:         name,
				Country:      cpb.Country_COUNTRY_VN,
				Subject:      cpb.Subject_SUBJECT_NONE,
				Grade:        "G5",
				DisplayOrder: 1,
				SchoolId:     stepState.CurrentSchoolID,
				LocationIds:  locationIDs,
			},
		},
	}

	stepState.Request = req
	stepState.Response, stepState.ResponseErr = mpb.NewMasterDataCourseServiceClient(s.MasterMgmtConn).
		UpsertCourses(contextWithToken(s, ctx), req)
	return StepStateToContext(ctx, stepState), nil
}
