package bob

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"go.uber.org/multierr"

	"github.com/manabie-com/backend/internal/bob/constants"
	"github.com/manabie-com/backend/internal/bob/entities"
	entities_bob "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"
)

var courses = [][]interface{}{{"course-1", "Course 1 name", "COUNTRY_VN", "SUBJECT_BIOLOGY", 12, 1, nil, "COURSE_TYPE_CONTENT", nil, "2020-07-02", "2025-07-02"}, {"course-2", "Course 2 name", "COUNTRY_VN", "SUBJECT_MATHS", 10, 2, nil, "COURSE_TYPE_CONTENT", nil, "2020-07-02", "2025-07-02"}, {"course-3", "Course 3 name", "COUNTRY_VN", "SUBJECT_MATHS", 10, 2, "2000-05-10T17:55:56z", "COURSE_TYPE_CONTENT", nil, "2020-07-02", "2025-07-02"}, {"course-teacher-1", "Course teacher 1 name", "COUNTRY_VN", "SUBJECT_BIOLOGY", 12, 1, nil, "COURSE_TYPE_CONTENT", nil, "2020-07-02", "2025-07-02"}, {"course-teacher-2", "Course teacher 2 name", "COUNTRY_VN", "SUBJECT_MATHS", 10, 2, nil, "COURSE_TYPE_CONTENT", nil, "2020-07-02", "2025-07-02"}, {"course-teacher-3", "Course teacher 3 name", "COUNTRY_VN", "SUBJECT_MATHS", 10, 2, "2000-05-10T17:55:56z", "COURSE_TYPE_CONTENT", nil, "2020-07-02", "2025-07-02"}, {"course-1-JP", "Course 1 JP name", "COUNTRY_JP", "SUBJECT_BIOLOGY", 12, 1, nil, "COURSE_TYPE_CONTENT", nil, "2020-07-02", "2025-07-02"}, {"course-2-JP", "Course 2 JP name", "COUNTRY_JP", "SUBJECT_MATHS", 10, 2, nil, "COURSE_TYPE_CONTENT", nil, "2020-07-02", "2025-07-02"}, {"course-1-SG", "Course 1 SG name", "COUNTRY_SG", "SUBJECT_BIOLOGY", 12, 1, nil, "COURSE_TYPE_CONTENT", nil, "2020-07-02", "2025-07-02"}, {"course-2-SG", "Course 2 SG name", "COUNTRY_SG", "SUBJECT_MATHS", 10, 2, nil, "COURSE_TYPE_CONTENT", nil, "2020-07-02", "2025-07-02"}, {"course-1-ID", "Course 1 ID name", "COUNTRY_ID", "SUBJECT_BIOLOGY", 12, 1, nil, "COURSE_TYPE_CONTENT", nil, "2020-07-02", "2025-07-02"}, {"course-2-ID", "Course 2 ID name", "COUNTRY_ID", "SUBJECT_MATHS", 10, 2, nil, "COURSE_TYPE_CONTENT", nil, "2020-07-02", "2025-07-02"}, {"course-live-1", "Course live 1 name", "COUNTRY_VN", "SUBJECT_BIOLOGY", 12, 1, nil, "COURSE_TYPE_LIVE", "course-live-1-plan", "2020-07-02", "2025-07-02"}, {"course-live-2", "Course live 2 name", "COUNTRY_VN", "SUBJECT_MATHS", 10, 2, nil, "COURSE_TYPE_LIVE", "course-live-2-plan", "2020-07-02", "2025-07-02"}, {"course-live-3", "Course live 3 name", "COUNTRY_VN", "SUBJECT_MATHS", 10, 2, "2000-05-10T17:55:56z", "COURSE_TYPE_LIVE", "course-live-3-plan", "2020-07-02", "2025-07-02"}, {"course-live-teacher-1", "Course live teacher 1 name", "COUNTRY_VN", "SUBJECT_BIOLOGY", 12, 1, nil, "COURSE_TYPE_LIVE", "course-live-teacher-1-plan", "2020-07-02", "2025-07-02"}, {"course-live-teacher-2", "Course live teacher 2 name", "COUNTRY_VN", "SUBJECT_MATHS", 10, 2, nil, "COURSE_TYPE_LIVE", "course-live-teacher-2-plan", "2020-07-02", "2025-07-02"}, {"course-live-teacher-3", "Course live teacher 3 name", "COUNTRY_VN", "SUBJECT_MATHS", 10, 2, "2000-05-10T17:55:56z", "COURSE_TYPE_LIVE", "course-live-teacher-3-plan", "2020-07-02", "2025-07-02"}, {"course-live-teacher-4", "Course live teacher 4 name", "COUNTRY_VN", "SUBJECT_MATHS", 10, 2, nil, "COURSE_TYPE_LIVE", "course-live-teacher-4-plan", "2020-07-02", "2025-07-02"}, {"course-live-teacher-5", "Course live teacher 5 name", "COUNTRY_VN", "SUBJECT_MATHS", 10, 2, nil, "COURSE_TYPE_LIVE", "course-live-teacher-5-plan", "2020-07-02", "2025-07-02"}, {"course-live-teacher-6", "Course live teacher 6 name", "COUNTRY_VN", "SUBJECT_MATHS", 10, 2, nil, "COURSE_TYPE_LIVE", "course-live-teacher-6-plan", "2020-07-02", "2025-07-02"}, {"course-live-teacher-7", "Course live teacher 7 name", "COUNTRY_VN", "SUBJECT_MATHS", 10, 2, nil, "COURSE_TYPE_LIVE", "course-live-teacher-7-plan", "2020-07-02", "2025-07-02"}, {"course-live-dont-have-lesson-1", "Course live dont have lesson 1 name", "COUNTRY_VN", "SUBJECT_MATHS", 10, 2, nil, "COURSE_TYPE_LIVE", "course-live-dont-have-lesson-1-plan", "2020-07-02", "2025-07-02"}, {"course-live-dont-have-lesson-2", "Course live dont have lesson 2 name", "COUNTRY_VN", "SUBJECT_MATHS", 10, 2, nil, "COURSE_TYPE_LIVE", "course-live-dont-have-lesson-1-plan", "2020-07-02", "2025-07-02"}, {"course-live-complete-lesson-1", "Course live complete lesson 1 name", "COUNTRY_VN", "SUBJECT_MATHS", 10, 2, nil, "COURSE_TYPE_LIVE", "course-live-complete-lesson-1-plan", "2020-07-02", "2020-07-03"}, {"course-live-complete-lesson-2", "Course live complete lesson 2 name", "COUNTRY_VN", "SUBJECT_MATHS", 10, 2, nil, "COURSE_TYPE_LIVE", "course-live-complete-lesson-2-plan", "2019-07-02", "2019-07-03"}, {"course-dont-have-chapter-1-JP", "Course dont have chapter 1 JP name", "COUNTRY_JP", "SUBJECT_MATHS", 10, 2, nil, "COURSE_TYPE_CONTENT", nil, "2020-07-02", "2025-07-02"}, {"course-dont-have-chapter-1-VN", "Course dont have chapter 1 VN name", "COUNTRY_VN", "SUBJECT_BIOLOGY", 10, 2, nil, "COURSE_TYPE_CONTENT", nil, "2020-07-02", "2025-07-02"}, {"course-have-chapter-dont-exist-1-VN", "Course have chapter dont exist 1 VN name", "COUNTRY_VN", "SUBJECT_BIOLOGY", 10, 2, nil, "COURSE_TYPE_CONTENT", nil, "2020-07-02", "2025-07-02"}, {"course-have-book-1-JP", "Course have book 1 JP name", "COUNTRY_JP", "SUBJECT_MATHS", 10, 2, nil, "COURSE_TYPE_CONTENT", nil, "2020-07-02", "2025-07-02"}, {"course-teacher-have-book-2-VN", "Course teacher have book 2 VN name", "COUNTRY_VN", "SUBJECT_MATHS", 10, 2, nil, "COURSE_TYPE_CONTENT", nil, "2020-07-02", "2025-07-02"}, {"course-teacher-have-book-4-VN", "Course teacher have book 3 VN name", "COUNTRY_VN", "SUBJECT_MATHS", 10, 2, nil, "COURSE_TYPE_CONTENT", nil, "2020-07-02", "2025-07-02"}, {"course-teacher-have-book-5-VN", "Course teacher have book 4 VN name", "COUNTRY_VN", "SUBJECT_MATHS", 10, 2, nil, "COURSE_TYPE_CONTENT", nil, "2020-07-02", "2025-07-02"}, {"course-teacher-have-book-missing-subject-grade", "Course teacher have book 6 VN name", "COUNTRY_VN", "SUBJECT_MATHS", 10, 2, nil, "COURSE_TYPE_CONTENT", nil, "2020-07-02", "2025-07-02"}}

func (s *suite) studentRetrievesCourses(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	var err error
	var limit int32 = 100

	req := &pb.RetrieveCoursesRequest{
		Countries: []pb.Country{pb.COUNTRY_VN, pb.COUNTRY_JP},
		Limit:     limit,
	}

	resp, err := pb.NewCourseClient(s.Conn).RetrieveCourses(s.signedCtx(ctx), req)
	if err != nil {
		stepState.ResponseErr = err
		return StepStateToContext(ctx, stepState), nil
	}

	coursesInResp := make([]*pb.RetrieveCoursesResponse_Course, 0, resp.Total)
	coursesInResp = append(coursesInResp, resp.Courses...)

	total := resp.Total
	var times int32
	if total%limit == 0 {
		times = total / limit
	} else {
		times = total/limit + 1
	}

	var page int32 = 2
	for times > 1 {
		req = &pb.RetrieveCoursesRequest{
			Countries: []pb.Country{pb.COUNTRY_VN, pb.COUNTRY_JP},
			Limit:     limit,
			Page:      page,
		}
		resp, err = pb.NewCourseClient(s.Conn).RetrieveCourses(s.signedCtx(ctx), req)
		if err != nil {
			stepState.ResponseErr = err
			return StepStateToContext(ctx, stepState), nil
		}
		coursesInResp = append(coursesInResp, resp.Courses...)
		times--
		page++
	}

	stepState.Request = req
	stepState.Response = &pb.RetrieveCoursesResponse{
		Courses: coursesInResp,
		Total:   total,
	}
	stepState.ResponseErr = nil

	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) generateChapter(ctx context.Context, country, subject, bookID string, grade, schoolID int, chapterIDs []string) (*entities_bob.Chapter, error) {
	chapter1 := &entities_bob.Chapter{}
	database.AllNullEntity(chapter1)
	err := multierr.Combine(chapter1.ID.Set("book-chapter-"+s.newID()), chapter1.Name.Set("book-chapter-name-"+bookID), chapter1.Country.Set(country), chapter1.Subject.Set(subject), chapter1.Grade.Set(grade), chapter1.DisplayOrder.Set(1), chapter1.SchoolID.Set(schoolID), chapter1.UpdatedAt.Set(time.Now()), chapter1.CreatedAt.Set(time.Now()), chapter1.DeletedAt.Set(nil), chapter1.CurrentTopicDisplayOrder.Set(0))
	if err != nil {
		return nil, err
	}
	chapterIDs = append(chapterIDs, chapter1.ID.String)
	for _, chapterID := range chapterIDs {
		bookChapter := &entities_bob.BookChapter{}
		database.AllNullEntity(bookChapter)
		err := multierr.Combine(bookChapter.BookID.Set(bookID), bookChapter.ChapterID.Set(chapterID), bookChapter.UpdatedAt.Set(time.Now()), bookChapter.CreatedAt.Set(time.Now()), bookChapter.DeletedAt.Set(nil))
		if err != nil {
			return nil, err
		}
	}
	return chapter1, nil
}
func (s *suite) generateBook(ctx context.Context, courseID, country, subject string, grade, schoolID int, chapterIDs []string) (*entities_bob.Book, error) {
	now := time.Now()
	book := &entities_bob.Book{}
	database.AllNullEntity(book)
	bookName := "book-name-course-id_" + courseID
	err := multierr.Combine(book.Country.Set(country), book.SchoolID.Set(schoolID), book.Subject.Set(subject), book.Grade.Set(grade), book.Name.Set(bookName), book.ID.Set(s.newID()), book.CreatedAt.Set(now), book.UpdatedAt.Set(now), book.CurrentChapterDisplayOrder.Set(0))
	if err != nil {
		return nil, err
	}
	_, err = s.generateChapter(ctx, country, subject, book.ID.String, grade, schoolID, chapterIDs)
	if err != nil {
		return nil, err
	}
	return book, nil
}
func (s *suite) aListOfValidPresetStudyPlan(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err1 := s.aValidPresetStudyPlan(ctx, "course-live-1-plan"+stepState.Random)
	ctx, err2 := s.aValidPresetStudyPlan(ctx, "course-live-2-plan"+stepState.Random)
	ctx, err3 := s.aValidPresetStudyPlan(ctx, "course-live-3-plan"+stepState.Random)
	ctx, err4 := s.aValidPresetStudyPlan(ctx, "course-live-teacher-1-plan"+stepState.Random)
	ctx, err5 := s.aValidPresetStudyPlan(ctx, "course-live-teacher-2-plan"+stepState.Random)
	ctx, err6 := s.aValidPresetStudyPlan(ctx, "course-live-teacher-3-plan"+stepState.Random)
	ctx, err7 := s.aValidPresetStudyPlan(ctx, "course-live-teacher-4-plan"+stepState.Random)
	ctx, err8 := s.aValidPresetStudyPlan(ctx, "course-live-teacher-5-plan"+stepState.Random)
	ctx, err9 := s.aValidPresetStudyPlan(ctx, "course-live-teacher-6-plan"+stepState.Random)
	ctx, err10 := s.aValidPresetStudyPlan(ctx, "course-live-teacher-7-plan"+stepState.Random)
	ctx, err11 := s.aValidPresetStudyPlan(ctx, "course-live-dont-have-lesson-1-plan"+stepState.Random)
	ctx, err12 := s.aValidPresetStudyPlan(ctx, "course-live-complete-lesson-1-plan"+stepState.Random)
	ctx, err13 := s.aValidPresetStudyPlan(ctx, "course-live-complete-lesson-2-plan"+stepState.Random)
	err := multierr.Combine(err1, err2, err3, err4, err5, err6, err7, err8, err9, err10, err11, err12, err13)
	return ctx, err
}
func (s *suite) generateAnAdminToken(ctx context.Context) (context.Context, string) {
	id := s.newID()
	var err error
	ctx, _ = s.aValidUser(ctx, withID(id), withRole(entities_bob.UserGroupAdmin))
	token, err := s.generateExchangeToken(id, entities.UserGroupAdmin)
	if err != nil {
		return ctx, ""
	}
	return ctx, token
}

func (s *suite) InsertAPresetStudyPlan(ctx context.Context, p *pb.PresetStudyPlan) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, _ = s.signedAsAccount(ctx, "staff granted role school admin")

	req := &pb.UpsertPresetStudyPlansRequest{
		PresetStudyPlans: []*pb.PresetStudyPlan{p},
	}

	_, err := pb.NewCourseClient(s.Conn).UpsertPresetStudyPlans(contextWithToken(s, ctx), req)
	return StepStateToContext(ctx, stepState), err
}
func (s *suite) aValidPresetStudyPlan(ctx context.Context, id string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	p := generatePresetStudyPlan()
	p.Id = id
	ctx, err := s.InsertAPresetStudyPlan(ctx, p)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}

	return StepStateToContext(ctx, stepState), err
}
func (s *suite) aListOfLessonAreExistedInDB(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	query := "SELECT course_id, preset_study_plan_id, teacher_ids, name, school_id FROM courses WHERE course_type='COURSE_TYPE_LIVE' AND course_id = $1"

	var listCourse [][]interface{}
	if stepState.Courses == nil {
		listCourse = courses
	} else {
		listCourse = stepState.Courses
	}

	for _, course := range listCourse {
		courseID := pgtype.Text{}
		presetStudyPlanID := pgtype.Text{}
		name := pgtype.Text{}
		schoolID := constants.ManabieSchool
		teacherIDs := pgtype.TextArray{}

		if err := s.DB.QueryRow(ctx, query, course[0]).Scan(&courseID, &presetStudyPlanID, &teacherIDs, &name, &schoolID); err != nil && err != pgx.ErrNoRows {
			return StepStateToContext(ctx, stepState), err
		}
		if courseID.String == "" {
			continue
		}

		topic := &entities_bob.Topic{}
		database.AllNullEntity(topic)
		topic.ID.Set(s.newID())
		topic.SchoolID.Set(schoolID)
		topic.Name.Set(name.String)
		topic.CreatedAt.Set(time.Now())
		topic.UpdatedAt.Set(time.Now())
		topic.Grade.Set(1)
		topic.Subject.Set(pb.SUBJECT_BIOLOGY.String())
		topic.Country.Set(pb.COUNTRY_VN.String())
		topic.TopicType.Set(pb.TOPIC_TYPE_LIVE_LESSON.String())
		topic.TotalLOs.Set(0)
		topic.EssayRequired.Set(false)

		lesson := &entities_bob.Lesson{}
		database.AllNullEntity(lesson)
		lesson.CourseID.Set(courseID.String)
		lesson.LessonID.Set(s.newID())
		if len(teacherIDs.Elements) > 0 {
			lesson.TeacherID = teacherIDs.Elements[0]
		} else {
			lesson.TeacherID.Set("teacher-id")
		}
		lesson.CreatedAt.Set(time.Now())
		lesson.UpdatedAt.Set(time.Now())
		lesson.StreamLearnerCounter.Set(database.Int4(0))
		lesson.LearnerIds.Set(database.JSONB([]byte("{}")))

		preset := &entities_bob.PresetStudyPlanWeekly{}
		database.AllNullEntity(preset)
		preset.ID.Set(s.newID())
		preset.LessonID.Set(lesson.LessonID.String)
		preset.PresetStudyPlanID.Set(presetStudyPlanID.String)
		preset.StartDate.Set(time.Now())
		preset.EndDate.Set(time.Now().Add(time.Hour * 24))
		preset.CreatedAt.Set(time.Now())
		preset.UpdatedAt.Set(time.Now())
		preset.TopicID.Set(topic.ID.String)
		preset.Week.Set(1)

		if strings.Contains(courseID.String, "dont-have-lesson") {
			lesson.DeletedAt.Set(time.Now())
		} else if strings.Contains(courseID.String, "complete-lesson") {
			preset.StartDate.Set(time.Now().Add(-2 * time.Hour))
			preset.EndDate.Set(time.Now().Add(-1 * time.Hour))
		}

		cmdTag, err := database.Insert(ctx, topic, s.DB.Exec)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("topic database.Insert: %w", err)
		}

		if err == nil && cmdTag.RowsAffected() != 1 {
			if err != nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf("topic database.Insert: %w", repositories.ErrUnAffected)
			}
		}

		if err := lesson.Normalize(); err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("lesson.Normalize err: %s", err)
		}

		cmdTag, err = database.Insert(ctx, lesson, s.DB.Exec)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("lesson database.Insert: %w", err)
		}

		if err == nil && cmdTag.RowsAffected() != 1 {
			if err != nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf("lesson database.Insert: %w", repositories.ErrUnAffected)
			}
		}

		cmdTag, err = database.Insert(ctx, preset, s.DB.Exec)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("PresetStudyPlanWeekly database.Insert: %w", err)
		}

		if err == nil && cmdTag.RowsAffected() != 1 {
			if err != nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf("PresetStudyPlanWeekly database.Insert: %w", repositories.ErrUnAffected)
			}
		}
	}

	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) aListOfCoursesAreExistedInDBOf(ctx context.Context, owner string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.aListOfValidPresetStudyPlan(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}
	var teacherIDs pgtype.TextArray
	_ = teacherIDs.Set(nil)

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

			args = append(args, constants.ManabieSchool)
			args = append(args, nil)
		case "above teacher":
			if !strings.Contains(courseID, "teacher") {
				continue
			}

			teacherRepo := &repositories.TeacherRepo{}
			teacher, err := teacherRepo.FindByID(ctx, s.DBPostgres, database.Text(stepState.CurrentTeacherID))
			if err != nil {
				return StepStateToContext(ctx, stepState), err

			}

			_ = teacherIDs.Set([]string{teacher.ID.String})
			args = append(args, constants.ManabieSchool)
			args = append(args, &teacherIDs)
		}

		_, err = s.DBPostgres.Exec(ctx, `INSERT INTO public."courses"
			(course_id, name, country, subject, grade, display_order, deleted_at, course_type, preset_study_plan_id, start_date, end_date, school_id, teacher_ids, updated_at, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13::text[], now(), now())
			ON CONFLICT ON CONSTRAINT courses_pk DO UPDATE SET name = $2, country = $3, subject = $4, grade = $5, display_order = $6,
			deleted_at = $7, course_type = $8, preset_study_plan_id = $9, start_date = $10, end_date = $11, school_id = $12, teacher_ids = $13`, args...)
		if err != nil {
			return StepStateToContext(ctx, stepState), err

		}

		schoolID := args[len(args)-2].(int)

		if strings.Contains(courseID, "have-book") {
			_, err = s.generateBook(ctx, courseID, c[2].(string), c[3].(string), c[4].(int), schoolID, []string{})
			if err != nil {
				return StepStateToContext(ctx, stepState), err

			}
		}

		if strings.Contains(courseID, "have-book-missing-subject-grade") {
			_, err = s.generateBook(ctx, courseID, c[2].(string), "", 0, schoolID, []string{})
			if err != nil {
				return StepStateToContext(ctx, stepState), err

			}
		}

		// select all classes of this school and insert to courses_classes table.
		rows, err := s.DB.Query(ctx, "SELECT class_id FROM classes WHERE school_id = $1", &schoolID)
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
			_, err := s.DB.Exec(ctx, "INSERT INTO courses_classes (course_id, class_id, created_at, updated_at) VALUES ($1, $2, now(), now()) ON CONFLICT DO NOTHING", courseID, classID)
			if err != nil {
				return StepStateToContext(ctx, stepState), err

			}
		}
		stepState.Courses = append(stepState.Courses, args)
		stepState.courseIds = append(stepState.courseIds, courseID)
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) AListOfCoursesAreExistedInDBOf(ctx context.Context, owner string) (context.Context, error) {
	return s.aListOfCoursesAreExistedInDBOf(ctx, owner)
}

func checkCourse(ctx context.Context, c *pb.RetrieveCoursesResponse_Course, course []interface{}) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.Courses == nil {
		return StepStateToContext(ctx, stepState), errors.New("wrong data return")
	}

	if isDeletedRecord(ctx, *c) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("cannot get deleted record")
	}

	if c.Id != course[0] {
		return ctx, fmt.Errorf("unexpected course id, got: %q, want: %q", c.Id, course[0])
	}
	if c.Name != course[1] {
		return ctx, fmt.Errorf("unexpected course name, got: %q, want: %q", c.Name, course[1])
	}
	if c.Country.String() != course[2] {
		return ctx, fmt.Errorf("unexpected course country, got: %q, want: %q", c.Country.String(), course[2])
	}
	if c.Subject.String() != course[3] {
		return ctx, fmt.Errorf("unexpected course subject, got: %q, want: %q", c.Subject.String(), course[3])
	}

	return ctx, nil
}

//nolint:gocyclo
func (s *suite) returnsAListOfCoursesOf(ctx context.Context, owner string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	userID := interceptors.UserIDFromContext(ctx)

	req := stepState.Request.(*pb.RetrieveCoursesRequest)
	resp := stepState.Response.(*pb.RetrieveCoursesResponse)

	teacherRepo := &repositories.TeacherRepo{}
	teacher, err := teacherRepo.FindByID(ctx, s.DB, database.Text(stepState.CurrentTeacherID))
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}
	teacherSchoolID := teacher.SchoolIDs.Elements[0].Int

	var expectedSchoolID, expectedTeacherID string
	switch owner {
	case "manabie":
		expectedSchoolID = fmt.Sprintf("%d", constants.ManabieSchool)
	case "above teacher":
		expectedSchoolID = fmt.Sprintf("%d", teacherSchoolID)
		expectedTeacherID = teacher.ID.String
	default:
		expectedSchoolID = fmt.Sprintf("%d|%d", constants.ManabieSchool, teacherSchoolID)
	}
	query := `SELECT count(*) FROM courses_classes WHERE course_id =$1 AND class_id =ANY($2)`

	queryCourseBooks := `SELECT count(*) FROM courses_books WHERE deleted_at IS NULL AND course_id = $1`
	queryBooks := `SELECT count(b.*) FROM books b JOIN courses_books cb ON cb.book_id = b.book_id WHERE b.deleted_at IS NULL AND cb.course_id = $1`

	//queryLesson := `SELECT count(*) FROM lessons WHERE deleted_at IS NULL AND course_id = $1`

	queryCourse := `SELECT COUNT(*) FROM classes c JOIN class_members cm ON cm.class_id = cl.class_id JOIN courses_classes cl ON c.class_id = cl.class_id WHERE cl.course_id = $1 AND cm.user_id = $2 AND cm.deleted_at IS NULL`

	totalCourses := 0

	for _, c := range resp.Courses {
		var course []interface{}
		for _, e := range stepState.Courses {
			if e[0].(string) == c.Id {
				course = e
			}

		}
		if len(course) == 0 || course[0] == "" {
			continue
		}

		if ctx, err := checkCourse(ctx, c, course); err != nil {
			return ctx, err
		}

		if owner == "above teacher" && c.Teachers[0].UserId != expectedTeacherID {
			return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected teacher id, got: %s, want: %s", c.Teachers[0].UserId, expectedTeacherID)
		}
		var count int
		if err := s.DB.QueryRow(ctx, query, c.Id, c.ClassIds).Scan(&count); err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		if count != len(c.ClassIds) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected class ids, got: %d, want: %d", len(c.ClassIds), count)
		}

		schoolStr := fmt.Sprintf("%d", c.SchoolId)
		if !strings.Contains(expectedSchoolID, schoolStr) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected course's school: got %s, want: %s", schoolStr, expectedSchoolID)
		}

		if req.IsAssigned && req.CourseType == pb.COURSE_TYPE_LIVE {
			var countClass int
			if err := s.DB.QueryRow(ctx, queryCourse, c.Id, userID).Scan(&countClass); err != nil {
				return StepStateToContext(ctx, stepState), err
			}
			if countClass == 0 {
				return StepStateToContext(ctx, stepState), fmt.Errorf("course doesnot assign to users")
			}
		}

		var countCoursesBooks int
		if err := s.DB.QueryRow(ctx, queryCourseBooks, c.Id).Scan(&countCoursesBooks); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		if countCoursesBooks > 1 {
			if len(c.BookIds) != countCoursesBooks {
				return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected course books expect %d, but got %d", countCoursesBooks, len(c.BookIds))
			}

			var countBooks int
			if err := s.DB.QueryRow(ctx, queryBooks, c.Id).Scan(&countBooks); err != nil {
				return StepStateToContext(ctx, stepState), err
			}

			if countBooks != len(c.BookIds) {
				return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected book expect %d, but got %d", countBooks, len(c.BookIds))
			}

		}

		totalCourses++
	}

	return checkSomeThingThatIDonotUnderstand(ctx, owner, totalCourses, req)
}

func checkSomeThingThatIDonotUnderstand(ctx context.Context, owner string, totalCourses int, req *pb.RetrieveCoursesRequest) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var totalManabieCourses, totalTeacherCourses, totalActiveLiveCourses, totalLiveCompleteCourse, totalOngoingLiveCourses int
	for _, c := range stepState.Courses {
		country := c[2].(string)
		if !inArrayCountry(pb.Country(pb.Country_value[country]), req.Countries) {
			continue
		}

		//c[6] is deleted_at
		if c[6] != nil {
			continue
		}

		if req.CourseType == pb.COURSE_TYPE_NONE {
			req.CourseType = pb.COURSE_TYPE_CONTENT
		}

		//c[7] is course_type
		if c[7] != req.CourseType.String() {
			continue
		}

		//c[0] is course_id
		if strings.Contains(c[0].(string), "teacher") {
			totalTeacherCourses++
		} else {
			totalManabieCourses++
		}
		if strings.Contains(c[0].(string), "live") { // c[6] deleted_at
			if strings.Contains(c[0].(string), "complete") {
				totalLiveCompleteCourse++
			} else if strings.Contains(c[0].(string), "dont-have-lesson") {
				totalOngoingLiveCourses++
			} else {
				totalActiveLiveCourses++
			}
		}
	}
	switch owner {
	case "manabie":
		if totalCourses != totalManabieCourses {
			return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected total courses of %s: got: %d, want: %d", owner, totalCourses, totalManabieCourses)
		}
	case "above teacher":
		if totalCourses != totalTeacherCourses {
			return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected total courses of %s: got: %d, want: %d", owner, totalCourses, totalTeacherCourses)
		}
	case "live course":
		if totalCourses != totalActiveLiveCourses {
			return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected total live courses of %s: got: %d, want: %d", owner, totalCourses, totalActiveLiveCourses)
		}
	case "completed live course":
		if totalCourses != totalLiveCompleteCourse {
			return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected total live courses of %s: got: %d, want: %d", owner, totalCourses, totalLiveCompleteCourse)
		}
	case "ongoing live course":
		if totalCourses != totalOngoingLiveCourses {
			return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected total live courses of %s: got: %d, want: %d", owner, totalCourses, totalOngoingLiveCourses)
		}
	default:
		if totalCourses != (totalManabieCourses + totalTeacherCourses) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected total courses of %s: got: %d, want: %d", owner, totalCourses, totalTeacherCourses+totalManabieCourses)
		}

	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentRetrievesAssignedCourses(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	var limit int32 = 100
	req := &pb.RetrieveCoursesRequest{
		Countries: []pb.Country{pb.COUNTRY_VN, pb.COUNTRY_JP},
		Limit:     limit,
	}

	resp, err := pb.NewCourseClient(s.Conn).RetrieveAssignedCourses(s.signedCtx(ctx), req)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}

	coursesInResp := make([]*pb.RetrieveCoursesResponse_Course, 0, resp.Total)
	coursesInResp = append(coursesInResp, resp.Courses...)

	total := resp.Total
	var times int32
	if total%limit == 0 {
		times = total / limit
	} else {
		times = total/limit + 1
	}

	var page int32 = 2
	for times > 1 {
		req = &pb.RetrieveCoursesRequest{
			Countries: []pb.Country{pb.COUNTRY_VN, pb.COUNTRY_JP},
			Limit:     limit,
			Page:      page,
		}
		resp, err = pb.NewCourseClient(s.Conn).RetrieveAssignedCourses(s.signedCtx(ctx), req)
		if err != nil {
			return StepStateToContext(ctx, stepState), err

		}
		coursesInResp = append(coursesInResp, resp.Courses...)
		times--
		page++
	}

	stepState.Request = req
	stepState.Response = &pb.RetrieveCoursesResponse{
		Courses: coursesInResp,
		Total:   total,
	}
	stepState.ResponseErr = nil

	return StepStateToContext(ctx, stepState), nil
}
func inArrayCountry(s pb.Country, arr []pb.Country) bool {
	for _, str := range arr {
		if s == str {
			return true
		}
	}
	return false
}
func isDeletedRecord(ctx context.Context, responseCourse pb.RetrieveCoursesResponse_Course) bool {
	stepState := StepStateFromContext(ctx)
	for _, exampleCourse := range stepState.Courses {
		if exampleCourse[6] != nil && exampleCourse[0] == responseCourse.Id {
			return true
		}
	}
	return false
}
func (s *suite) studentRetrievesAssignedLiveCourses(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	var limit int32 = 100

	req := &pb.RetrieveCoursesRequest{
		Countries:  []pb.Country{pb.COUNTRY_VN, pb.COUNTRY_JP},
		CourseType: pb.COURSE_TYPE_LIVE,
		Limit:      limit,
	}
	resp, err := pb.NewCourseClient(s.Conn).RetrieveAssignedCourses(s.signedCtx(ctx), req)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}

	coursesInResp := make([]*pb.RetrieveCoursesResponse_Course, 0, resp.Total)
	coursesInResp = append(coursesInResp, resp.Courses...)

	total := resp.Total
	var times int32
	if total%limit == 0 {
		times = total / limit
	} else {
		times = total/limit + 1
	}

	var page int32 = 2
	for times > 1 {
		req = &pb.RetrieveCoursesRequest{
			Countries:  []pb.Country{pb.COUNTRY_VN, pb.COUNTRY_JP},
			CourseType: pb.COURSE_TYPE_LIVE,
			Limit:      limit,
			Page:       page,
		}
		resp, err = pb.NewCourseClient(s.Conn).RetrieveAssignedCourses(s.signedCtx(ctx), req)
		if err != nil {
			return StepStateToContext(ctx, stepState), err

		}
		coursesInResp = append(coursesInResp, resp.Courses...)
		times--
		page++
	}

	stepState.Request = req
	stepState.Response = &pb.RetrieveCoursesResponse{
		Courses: coursesInResp,
		Total:   total,
	}
	stepState.ResponseErr = nil
	return StepStateToContext(ctx, stepState), nil

}
func (s *suite) studentRetrievesCoursesWithIds(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := &pb.RetrieveCoursesByIDsRequest{
		Ids: []string{
			"course-1", "course-2", "course-3",
			"course-live-complete-lesson-1", "course-live-dont-have-lesson-1",
			"course-dont-have-chapter-1-JP", "course-dont-have-chapter-1-VN",
		},
	}

	if stepState.Random != "" {
		for i := range req.Ids {
			req.Ids[i] += stepState.Random
		}
	}

	stepState.Request = req

	stepState.Response, stepState.ResponseErr = pb.NewCourseClient(s.Conn).RetrieveCoursesByIDs(s.signedCtx(ctx), req)
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) returnsAListOfCoursesFromRequestedIds(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	//queryLesson := `SELECT count(*) FROM lessons WHERE deleted_at IS NULL AND course_id = $1`

	queryCourseBooks := `SELECT count(*) FROM courses_books WHERE deleted_at IS NULL AND course_id = $1`
	queryBooks := `SELECT count(b.*) FROM books b JOIN courses_books cb ON cb.book_id = b.book_id WHERE b.deleted_at IS NULL AND cb.course_id = $1`

	rsp := stepState.Response.(*pb.RetrieveCoursesResponse)
	courseMap := make(map[string]*pb.RetrieveCoursesResponse_Course)
	for _, course := range rsp.Courses {
		courseMap[course.Id] = course

		var countCoursesBooks int
		if err := s.DB.QueryRow(ctx, queryCourseBooks, course.Id).Scan(&countCoursesBooks); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		if countCoursesBooks > 1 {
			if len(course.BookIds) != countCoursesBooks {
				return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected course books expect %d, but got %d", countCoursesBooks, len(course.BookIds))
			}

			var countBooks int
			if err := s.DB.QueryRow(ctx, queryBooks, course.Id).Scan(&countBooks); err != nil {
				return StepStateToContext(ctx, stepState), err
			}

			if countBooks != len(course.BookIds) {
				return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected book expect %d, but got %d", countBooks, len(course.BookIds))
			}

		}

		if len(course.Chapters) > 0 {
			displayOrder := int32(0)

			for _, v := range course.Chapters {
				if v.DisplayOrder < displayOrder {
					return StepStateToContext(ctx, stepState), fmt.Errorf("display order not match")
				}
				displayOrder = v.DisplayOrder
			}

		}

	}
	req := stepState.Request.(*pb.RetrieveCoursesByIDsRequest)
	for _, id := range req.Ids {
		if _, ok := courseMap[id]; !ok {
			// skip deleted course
			if id == "course-3"+stepState.Random {
				continue
			}
			return StepStateToContext(ctx, stepState), fmt.Errorf("cannot find course with id: %s", id)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) studentRetrievesAssignedCoursesWithRetrieveCourseEndPoint(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	var limit int32 = 100

	req := &pb.RetrieveCoursesRequest{
		Countries:  []pb.Country{pb.COUNTRY_VN, pb.COUNTRY_JP},
		IsAssigned: true,
		Limit:      limit,
	}
	resp, err := pb.NewCourseClient(s.Conn).RetrieveCourses(s.signedCtx(ctx), req)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}

	coursesInResp := make([]*pb.RetrieveCoursesResponse_Course, 0, resp.Total)
	coursesInResp = append(coursesInResp, resp.Courses...)

	total := resp.Total
	var times int32
	if total%limit == 0 {
		times = total / limit
	} else {
		times = total/limit + 1
	}

	var page int32 = 2
	for times > 1 {
		req = &pb.RetrieveCoursesRequest{
			Countries:  []pb.Country{pb.COUNTRY_VN, pb.COUNTRY_JP},
			Limit:      limit,
			IsAssigned: true,
			Page:       page,
		}
		resp, err = pb.NewCourseClient(s.Conn).RetrieveCourses(s.signedCtx(ctx), req)
		if err != nil {
			return StepStateToContext(ctx, stepState), err

		}
		coursesInResp = append(coursesInResp, resp.Courses...)
		times--
		page++
	}

	stepState.Request = req
	stepState.Response = &pb.RetrieveCoursesResponse{
		Courses: coursesInResp,
		Total:   total,
	}
	stepState.ResponseErr = nil
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) studentRetrievesCoursesInCurrentClass(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	var limit int32 = 100

	req := &pb.RetrieveCoursesRequest{
		Countries: []pb.Country{pb.COUNTRY_VN, pb.COUNTRY_JP},
		ClassId:   stepState.CurrentClassID,
		Limit:     limit,
	}

	resp, err := pb.NewCourseClient(s.Conn).RetrieveCourses(s.signedCtx(ctx), req)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}

	coursesInResp := make([]*pb.RetrieveCoursesResponse_Course, 0, resp.Total)
	coursesInResp = append(coursesInResp, resp.Courses...)

	total := resp.Total
	var times int32
	if total%limit == 0 {
		times = total / limit
	} else {
		times = total/limit + 1
	}

	var page int32 = 2
	for times > 1 {
		req = &pb.RetrieveCoursesRequest{
			Countries: []pb.Country{pb.COUNTRY_VN, pb.COUNTRY_JP},
			Limit:     limit,
			ClassId:   stepState.CurrentClassID,
			Page:      page,
		}
		resp, err = pb.NewCourseClient(s.Conn).RetrieveCourses(s.signedCtx(ctx), req)
		if err != nil {
			return StepStateToContext(ctx, stepState), err

		}
		coursesInResp = append(coursesInResp, resp.Courses...)
		times--
		page++
	}

	stepState.Request = req
	stepState.Response = &pb.RetrieveCoursesResponse{
		Courses: coursesInResp,
		Total:   total,
	}
	stepState.ResponseErr = nil

	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) studentRetrievesLiveCoursesWithStatus(ctx context.Context, status string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	var limit int32 = 100

	req := &pb.RetrieveCoursesRequest{
		Countries:    []pb.Country{pb.COUNTRY_VN, pb.COUNTRY_JP},
		CourseStatus: pb.CourseStatus(pb.CourseStatus_value[status]),
		CourseType:   pb.COURSE_TYPE_LIVE,
		Limit:        limit,
		Page:         1,
	}

	resp, err := pb.NewCourseClient(s.Conn).RetrieveCourses(s.signedCtx(ctx), req)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}

	coursesInResp := make([]*pb.RetrieveCoursesResponse_Course, 0, resp.Total)
	coursesInResp = append(coursesInResp, resp.Courses...)

	total := resp.Total
	var times int32
	if total%limit == 0 {
		times = total / limit
	} else {
		times = total/limit + 1
	}

	var page int32 = 2
	for times > 1 {
		req = &pb.RetrieveCoursesRequest{
			Countries:    []pb.Country{pb.COUNTRY_VN, pb.COUNTRY_JP},
			Limit:        limit,
			CourseStatus: pb.CourseStatus(pb.CourseStatus_value[status]),
			CourseType:   pb.COURSE_TYPE_LIVE,
			Page:         page,
		}
		resp, err = pb.NewCourseClient(s.Conn).RetrieveCourses(s.signedCtx(ctx), req)
		if err != nil {
			return StepStateToContext(ctx, stepState), err

		}
		coursesInResp = append(coursesInResp, resp.Courses...)
		times--
		page++
	}

	stepState.Request = req
	stepState.Response = &pb.RetrieveCoursesResponse{
		Courses: coursesInResp,
		Total:   total,
	}

	stepState.ResponseErr = nil
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) returnsAListOfCoursesInCurrentClass(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	rsp := stepState.Response.(*pb.RetrieveCoursesResponse)
	courseIDs := []string{}

	for _, course := range rsp.Courses {
		courseIDs = append(courseIDs, course.Id)
	}
	query := `SELECT COUNT(*) FROM courses_classes cc WHERE cc.course_id = ANY($1) and cc.class_id = $2`
	var count int
	if err := s.DB.QueryRow(ctx, query, &courseIDs, &stepState.CurrentClassID).Scan(&count); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if count != len(courseIDs) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("course not belong to class expected %d, got %d", count, len(courseIDs))
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) studentRetrievesAssignedLiveCoursesBelongToCurrentClass(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	var limit int32 = 100

	req := &pb.RetrieveCoursesRequest{
		Countries:  []pb.Country{pb.COUNTRY_VN, pb.COUNTRY_JP},
		CourseType: pb.COURSE_TYPE_LIVE,
		ClassId:    stepState.CurrentClassID,
		Limit:      limit,
	}

	resp, err := pb.NewCourseClient(s.Conn).RetrieveCourses(s.signedCtx(ctx), req)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}

	coursesInResp := make([]*pb.RetrieveCoursesResponse_Course, 0, resp.Total)
	coursesInResp = append(coursesInResp, resp.Courses...)

	total := resp.Total
	var times int32
	if total%limit == 0 {
		times = total / limit
	} else {
		times = total/limit + 1
	}

	var page int32 = 2
	for times > 1 {
		req = &pb.RetrieveCoursesRequest{
			Countries:  []pb.Country{pb.COUNTRY_VN, pb.COUNTRY_JP},
			Limit:      limit,
			CourseType: pb.COURSE_TYPE_LIVE,
			ClassId:    stepState.CurrentClassID,
			Page:       page,
		}
		resp, err = pb.NewCourseClient(s.Conn).RetrieveCourses(s.signedCtx(ctx), req)
		if err != nil {
			return StepStateToContext(ctx, stepState), err

		}
		coursesInResp = append(coursesInResp, resp.Courses...)
		times--
		page++
	}

	stepState.Request = req
	stepState.Response = &pb.RetrieveCoursesResponse{
		Courses: coursesInResp,
		Total:   total,
	}

	stepState.ResponseErr = nil
	return StepStateToContext(ctx, stepState), nil

}
func (s *suite) returnsEmptyListOfCourse(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	rsp := stepState.Response.(*pb.RetrieveCoursesResponse)
	if rsp.Total != 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expect empty list got %d", rsp.Total)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) anUnauthenticatedUser(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.AuthToken = "randome-token"
	return StepStateToContext(ctx, stepState), nil
}
