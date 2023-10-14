package lessonmgmt

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	entities_bob "github.com/manabie-com/backend/internal/bob/entities"
	bob_repo "github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/timeutil"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
	"golang.org/x/exp/slices"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Suite) adminRetrieveLessonManagementOnLessonmgmt(ctx context.Context, lessonTime, limitStr, offset string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.adminRetrieveLessonManagement(ctx, lessonTime, limitStr, offset)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	req := stepState.Request.(*lpb.RetrieveLessonsRequest)
	stepState.Response, stepState.ResponseErr = lpb.NewLessonReaderServiceClient(s.LessonMgmtConn).RetrieveLessonsV2(s.CommonSuite.SignedCtx(ctx), req)
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) aCourseTypeWithIDAndSchoolID(ctx context.Context, courseTypeID string, schoolID int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.CommonSuite.ASignedInWithSchool(contextWithResourcePath(ctx, i32ToStr(schoolID)), "school admin", int32(schoolID))
	if err != nil {
		return ctx, err
	}
	sql := `insert into course_type 
	(course_type_id, name , created_at, updated_at, resource_path)
	VALUES ($1, $2, $3, $4, $5)  ON CONFLICT DO NOTHING;`
	_, err = s.BobDB.Exec(ctx, sql, database.Text(courseTypeID), "bdd-test-course-type-name", time.Now(), time.Now(), database.Text(fmt.Sprint(schoolID)))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("cannot init grade err: %s", err)
	}
	stepState.CurrentCourseTypeID = courseTypeID
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) aListOfLessonsManagementOfSchoolAreExistedInDB(ctx context.Context, schoolID int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.CommonSuite.ASignedInWithSchool(contextWithResourcePath(ctx, i32ToStr(schoolID)), "school admin", int32(schoolID))
	if err != nil {
		return ctx, err
	}

	now := time.Now().Add(1000 * 24 * time.Hour)
	stepState.TimeRandom = time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), 0, 0, time.UTC)

	courseRepo := bob_repo.CourseRepo{}

	var courseIDs []string
	var teacherIDs []string
	school := schoolID
	for j := 0; j < 2; j++ {
		courseID := idutil.ULIDNow()
		timeAny := database.Timestamptz(time.Now())
		courseName := courseID
		course := &entities_bob.Course{}
		database.AllNullEntity(course)
		err = multierr.Combine(
			course.ID.Set(courseID),
			course.Name.Set(courseName),
			course.CreatedAt.Set(timeAny),
			course.UpdatedAt.Set(timeAny),
			course.DeletedAt.Set(nil),
			course.Grade.Set(3),
			course.StartDate.Set(time.Now().Add(2*time.Hour)),
			course.SchoolID.Set(school),
			course.Status.Set("COURSE_STATUS_NONE"),
		)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		err = courseRepo.Create(ctx, s.BobDB, course)

		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("cannot init course err with school %d: %s", school, err)
		}
		courseIDs = append(courseIDs, courseID)

		if len(stepState.CurrentCourseTypeID) > 0 {
			sql := `UPDATE courses set course_type_id = $1 WHERE course_id = $2`
			_, err = s.BobDB.Exec(ctx, sql, database.Text(stepState.CurrentCourseTypeID), database.Text(courseID))
			if err != nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf("cannot init lesson_student_subscriptions err: %s", err)
			}
		}
	}

	// create center
	for i := 0; i < 3; i++ {
		locationID := idutil.ULIDNow()
		err = s.CommonSuite.AddLocationUnderShool(ctx, locationID, int32(school))
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("cannot insert location %w", err)
		}
		stepState.FilterCenterIDs = append(stepState.FilterCenterIDs, locationID)
	}
	stepState.FilterCourseIDs = courseIDs
	stepState.FilterLocationIDs = stepState.LocationIDs
	stepState.CurrentCenterID = stepState.FilterCenterIDs[0]

	for j := 0; j < 2; j++ {
		teacherID := idutil.ULIDNow()
		_, err = s.CommonSuite.AValidTeacherProfileWithId(ctx, teacherID, int32(school))
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("cannot init teacher%s", err)
		}
		teacherIDs = append(teacherIDs, teacherID)
	}
	stepState.TeacherIDs = teacherIDs

	stepState.FilterCourseIDs = []string{courseIDs[0]}

	// create lesson group
	lg := &entities_bob.LessonGroup{}
	database.AllNullEntity(lg)
	lg.CourseID = database.Text(courseIDs[0])
	lg.LessonGroupID = database.Text(idutil.ULIDNow())
	lg.CreatedAt = database.Timestamptz(now)
	err = (&bob_repo.LessonGroupRepo{}).Create(ctx, s.BobDB, lg)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.LessonGroupRepo.Create: %w", err)
	}

	studentNames := []string{
		"student name 1",
		"studentname2",
	}
	for _, studentName := range studentNames {
		if _, err = s.CommonSuite.CreateATotalNumberOfStudentAccounts(ctx, studentName, 1); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}

	stepState.CourseIDs = courseIDs[0:1]
	if _, err = s.CommonSuite.SomeStudentSubscriptions(ctx); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("could not insert student subscription: %w", err)
	}
	stepState.FilterStudentIDs = stepState.StudentIDWithCourseID
	stepState.CurrentTeachingMethod = "group"
	// init class
	stepState.FilterClassIDs = make([]string, 0, 4)
	for i := 0; i < 4; i++ {
		_, err = s.insertClass(ctx)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("cannot init class err: %s", err)
		}
	}
	for i := 0; i < 22; i++ {
		lesson := &entities_bob.Lesson{}
		database.AllNullEntity(lesson)
		lessonID := idutil.ULIDNow()
		lessonName := fmt.Sprintf("Lesson name management %s", lessonID)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("cannot init class err: %s", err)
		}
		classId := stepState.FilterClassIDs[0]

		duration := time.Duration(i)
		startTime := timeutil.Now().Add(duration * time.Hour)
		endTime := timeutil.Now().Add(duration*time.Hour + time.Minute*2)
		if i > 11 {
			startTime = timeutil.Now().Add(-duration * time.Hour)
			endTime = timeutil.Now().Add(-duration*time.Hour + time.Minute*2)
		}
		centerID := stepState.FilterCenterIDs[rand.Intn(3)]

		status := cpb.LessonSchedulingStatus_name[int32(rand.Intn(4))]
		teachingMethod := cpb.LessonTeachingMethod_name[int32(rand.Intn(2))]
		courseId := courseIDs[0]
		if teachingMethod == cpb.LessonTeachingMethod_LESSON_TEACHING_METHOD_INDIVIDUAL.String() {
			courseId = ""
			classId = ""
		}

		err = multierr.Combine(
			lesson.LessonID.Set(lessonID),
			lesson.Name.Set(lessonName),
			lesson.CourseID.Set(courseId),
			lesson.CreatedAt.Set(database.Timestamptz(time.Now())),
			lesson.UpdatedAt.Set(database.Timestamptz(time.Now())),
			lesson.LessonType.Set(cpb.LessonType_LESSON_TYPE_ONLINE.String()),
			lesson.Status.Set(cpb.LessonStatus_LESSON_STATUS_NOT_STARTED.String()),
			lesson.StreamLearnerCounter.Set(database.Int4(0)),
			lesson.LearnerIds.Set(database.JSONB([]byte("{}"))),
			lesson.StartTime.Set(startTime),
			lesson.EndTime.Set(endTime),
			lesson.TeachingMethod.Set(database.Text(teachingMethod)),
			lesson.TeachingMedium.Set(database.Text(cpb.LessonTeachingMedium_LESSON_TEACHING_MEDIUM_OFFLINE.String())),
			lesson.ClassID.Set(classId),
			lesson.CenterID.Set(centerID),
			lesson.LessonGroupID.Set(lg.LessonGroupID),
			lesson.SchedulingStatus.Set(database.Text(status)),
		)

		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("cannot init lesson err: %s", err)
		}

		if i%2 == 1 {
			err = lesson.TeacherID.Set(teacherIDs[0])
			if err != nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf("cannot init lesson err: %s", err)
			}
		}

		if err = lesson.Normalize(); err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("lesson.Normalize err: %s", err)
		}

		cmdTag, e := database.Insert(ctx, lesson, s.LessonmgmtDB.Exec)
		if e != nil {
			return StepStateToContext(ctx, stepState), e
		}
		if cmdTag.RowsAffected() != 1 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("cannot insert lesson")
		}
		stepState.LessonIDs = append(stepState.LessonIDs, lessonID)
		updateResourcePath := "UPDATE lessons SET resource_path = $1 WHERE lesson_id = $2"
		_, err = s.LessonmgmtDB.Exec(ctx, updateResourcePath, database.Text(fmt.Sprint(school)), database.Text(lessonID))
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		if i%2 == 1 {
			for j := 0; j < 2; j++ {
				sql := `insert into lessons_teachers 
		(lesson_id, teacher_id, created_at)
		VALUES ($1, $2, $3)`
				_, err = s.LessonmgmtDB.Exec(ctx, sql, lesson.LessonID, database.Text(teacherIDs[j]), database.Timestamptz(time.Now()))
				if err != nil {
					return StepStateToContext(ctx, stepState), err
				}
				stepState.FilterTeacherIDs = append(stepState.FilterTeacherIDs, teacherIDs[j])
			}
		}

		if i%2 == 0 {
			for j := 0; j < 2; j++ {
				timeAny := database.Timestamptz(time.Now())
				sql := `insert into lessons_courses
		(lesson_id, course_id, created_at)
		VALUES ($1, $2, $3)`
				_, err = s.LessonmgmtDB.Exec(ctx, sql, lesson.LessonID, database.Text(courseIDs[j]), timeAny)
				if err != nil {
					return StepStateToContext(ctx, stepState), err
				}
			}
		}
		// insert lessons_member
		for index := 0; index < 2; index++ {
			sql := `INSERT INTO lesson_members
		(lesson_id, user_id, created_at, updated_at, resource_path)
		VALUES ($1, $2, $3, $4, $5)`
			learnerID := stepState.StudentIds[index]
			_, err = s.LessonmgmtDB.Exec(ctx, sql, database.Text(lessonID), database.Text(learnerID), now, now, database.Text(fmt.Sprint(schoolID)))
			if err != nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf("cannot init lesson_members: %s", err)
			}
			// insert into user_basic_info
			sqlUserBasicInfo := `INSERT INTO user_basic_info
			(user_id, name, first_name, last_name, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6) ON CONFLICT (user_id) DO NOTHING;`
			firstName := "bdd-test-student-first-name-" + learnerID
			lastname := "bdd-test-student-last-name-" + learnerID
			_, err = s.LessonmgmtDB.Exec(ctx, sqlUserBasicInfo, database.Text(learnerID), database.Text(studentNames[index]), database.Text(firstName), database.Text(lastname), now, now)
			if err != nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf("cannot init user_basic_info: %s", err)
			}
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) adminRetrieveLiveLessonManagementWithFilterOnLessonmgmt(
	ctx context.Context,
	lessonTime,
	keyWord,
	dateRange,
	timeRange,
	teachers,
	students,
	coursers,
	centers,
	locations,
	grade,
	dow,
	schedulingStatus,
	classes, gradesV2, courseTypeIDs string,
) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx = s.adminRetrieveLiveLessonManagementWithFilter(ctx, lessonTime, keyWord, dateRange,
		timeRange, teachers, students,
		coursers, centers, locations,
		grade, dow, schedulingStatus, classes, gradesV2, courseTypeIDs)

	req := stepState.Request.(*lpb.RetrieveLessonsRequest)
	stepState.Response, stepState.ResponseErr = lpb.NewLessonReaderServiceClient(s.LessonMgmtConn).RetrieveLessonsV2(s.CommonSuite.SignedCtx(ctx), req)
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) adminRetrieveLiveLessonManagementWithFilter(
	ctx context.Context,
	lessonTime,
	keyWord,
	dateRange,
	timeRange,
	teachers,
	students,
	coursers,
	centers,
	locations,
	grade,
	dow, schedulingStatus,
	classes,
	gradeV2,
	courseTypeIDs string) context.Context {
	stepState := StepStateFromContext(ctx)
	filter := &lpb.RetrieveLessonsFilter{}
	if len(dateRange) > 0 {
		if lessonTime == lpb.LessonTime_LESSON_TIME_PAST.String() {
			stepState.FilterFromDate = time.Now().Add(-24 * time.Hour)
			stepState.FilterToDate = time.Now()
		} else {
			stepState.FilterFromDate = time.Now()
			stepState.FilterToDate = time.Now().Add(24 * time.Hour)
		}
		filter.FromDate = timestamppb.New(stepState.FilterFromDate)
		filter.ToDate = timestamppb.New(stepState.FilterToDate)
	}
	if len(timeRange) > 0 {
		if lessonTime == lpb.LessonTime_LESSON_TIME_PAST.String() {
			stepState.FilterFromTime = time.Now().Add(-3 * time.Hour)
			stepState.FilterToTime = time.Now()
			filter.FromTime = durationpb.New(time.Duration(time.Now().Hour()-3) * time.Hour)
			filter.ToTime = durationpb.New(time.Duration(time.Now().Hour()) * time.Hour)
		} else {
			stepState.FilterFromTime = time.Now()
			stepState.FilterToTime = time.Now().Add(3 * time.Hour)
			filter.FromTime = durationpb.New(time.Duration(time.Now().Hour()) * time.Hour)
			filter.ToTime = durationpb.New(time.Duration(time.Now().Hour()+3) * time.Hour)
		}
	}
	if len(dow) > 0 {
		dows := strings.Split(dow, ",")
		ints := make([]cpb.DateOfWeek, len(dows))
		for i, s := range dows {
			val, _ := strconv.Atoi(s)
			ints[i] = cpb.DateOfWeek(cpb.DateOfWeek_value[cpb.DateOfWeek_name[int32(val)]])
		}
		filter.DateOfWeeks = ints
		filter.TimeZone = "UTC"
	}
	if teachers != "" && len(strings.Split(teachers, ",")) > 0 {
		filter.TeacherIds = stepState.FilterTeacherIDs[0:len(strings.Split(teachers, ","))]
	}
	if students != "" && len(strings.Split(students, ",")) > 0 {
		filterStudentIDs := stepState.FilterStudentIDs[0 : len(strings.Split(students, ","))*2]
		for i := 0; i < len(strings.Split(students, ",")); i++ {
			filter.StudentIds = append(filter.StudentIds, filterStudentIDs[i*2])
		}
	}
	if coursers != "" && len(strings.Split(coursers, ",")) > 0 {
		filter.CourseIds = stepState.FilterCourseIDs[0:len(strings.Split(coursers, ","))]
	}
	if centers != "" {
		val, _ := strconv.Atoi(strings.Split(centers, "-")[1])
		filter.LocationIds = []string{stepState.FilterCenterIDs[val-1]}
	}

	locationIds := make([]string, 0, len(locations))
	if locations != "" {
		locationNames := strings.Split(locations, ",")

		for _, v := range locationNames {
			val, _ := strconv.Atoi(strings.Split(v, "-")[1])
			locationIds = append(locationIds, stepState.FilterCenterIDs[val-1])
		}
	}

	if len(grade) > 0 {
		grades := strings.Split(grade, ",")
		ints := make([]int32, len(grades))
		for i, v := range grades {
			val, _ := strconv.Atoi(v)
			ints[i] = int32(val)
		}
		filter.Grades = ints
	}
	if len(gradeV2) > 0 {
		gradesV2 := strings.Split(gradeV2, ",")

		filter.GradesV2 = gradesV2
	}
	if len(schedulingStatus) > 0 {
		statusesString := strings.Split(schedulingStatus, ",")
		statuses := make([]domain.LessonSchedulingStatus, 0, len(statusesString))
		for _, v := range statusesString {
			statuses = append(statuses, domain.LessonSchedulingStatus(v))
			filter.SchedulingStatus = append(filter.SchedulingStatus, cpb.LessonSchedulingStatus(cpb.LessonSchedulingStatus_value[v]))
		}
		stepState.FilterSchedulingStatuses = statuses
	}

	if classes != "" && len(strings.Split(classes, ",")) > 0 {
		filter.ClassIds = stepState.FilterClassIDs[0:len(strings.Split(classes, ","))]
	}

	if len(courseTypeIDs) > 0 {
		filter.CourseTypeIds = append(filter.CourseTypeIds, courseTypeIDs)
	}

	req := &lpb.RetrieveLessonsRequest{
		Paging: &cpb.Paging{
			Limit: uint32(20),
		},
		LessonTime:  lpb.LessonTime(bpb.LessonTime_value[lessonTime]),
		CurrentTime: timestamppb.Now(),
		Filter:      filter,
		Keyword:     keyWord,
		LocationIds: locationIds,
	}

	stepState.Request = req
	return StepStateToContext(ctx, stepState)
}

func contextWithResourcePath(ctx context.Context, rp string) context.Context {
	claim := interceptors.JWTClaimsFromContext(ctx)
	if claim == nil {
		claim = &interceptors.CustomClaims{
			Manabie: &interceptors.ManabieClaims{},
		}
	}
	claim.Manabie.ResourcePath = rp
	return interceptors.ContextWithJWTClaims(ctx, claim)
}

func (s *Suite) intersectLocations(_centers, _locations string) []string {
	centers := strings.Split(_centers, ",")
	locations := strings.Split(_locations, ",")
	res := make([]string, 0)
	for i := range centers {
		if slices.Contains(locations, centers[i]) {
			res = append(res, centers[i])
		}
	}
	return res
}

func i32ToStr(i int) string {
	return strconv.Itoa(int(i))
}

func (s *Suite) aSignedInWithRandomID(ctx context.Context, role string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.RandSchoolID = rand.Intn(2999) + 1
	stepState.CurrentSchoolID = int32(stepState.RandSchoolID)
	ctx, err := s.CommonSuite.ASignedInWithSchool(contextWithResourcePath(ctx, i32ToStr(stepState.RandSchoolID)), "school admin", stepState.CurrentSchoolID)
	if err != nil {
		return ctx, err
	}
	ctx, err = s.CommonSuite.ARandomNumber(ctx)
	if err != nil {
		return ctx, err
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) aListOfLessonManagementOfSchoolAreExistedInDB(ctx context.Context, lessonTime, strFromID, strToID string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	now := time.Now().Add(1000 * 24 * time.Hour)
	stepState.TimeRandom = time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), 0, 0, time.UTC)
	fromID, err := strconv.Atoi((strings.Split(strFromID, "_"))[1])
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("cannot init lesson with outline init param")
	}
	toID, err := strconv.Atoi((strings.Split(strToID, "_"))[1])
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("cannot init lesson with outline init param")
	}
	courseRepo := bob_repo.CourseRepo{}
	for i := fromID; i < toID; i++ {
		duration := time.Duration(i)
		startTime := timeutil.Now().Add(duration * time.Hour)
		endTime := timeutil.Now().Add(duration*time.Hour + time.Minute*2)
		if lessonTime == lpb.LessonTime_LESSON_TIME_PAST.String() {
			startTime = timeutil.Now().Add(-duration * time.Hour)
			endTime = timeutil.Now().Add(-duration*time.Hour + time.Minute*2)
		}
		courseIDs := []string{}
		teacherIDs := []string{}
		school := stepState.RandSchoolID
		// create center
		centerID := idutil.ULIDNow()
		err = s.CommonSuite.AddLocationUnderShool(ctx, centerID, int32(school))
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("s.CommonSuite.AddLocationUnderShool %v", err)
		}
		stepState.CurrentCenterID = centerID

		for j := 0; j < 2; j++ {
			courseID := fmt.Sprintf("course_%s_%d_%d", stepState.Random, i, j)
			timeAny := database.Timestamptz(time.Now())
			courseName := courseID
			course := &entities_bob.Course{}
			database.AllNullEntity(course)
			err := multierr.Combine(
				course.ID.Set(courseID),
				course.Name.Set(courseName),
				course.CreatedAt.Set(timeAny),
				course.UpdatedAt.Set(timeAny),
				course.DeletedAt.Set(nil),
				course.Grade.Set(3),
				course.StartDate.Set(time.Now().Add(2*time.Hour)),
				course.SchoolID.Set(school),
				course.Status.Set("COURSE_STATUS_NONE"),
			)
			if err != nil {
				return StepStateToContext(ctx, stepState), err
			}
			err = courseRepo.Create(ctx, s.BobDB, course)

			if err != nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf("cannot init course err with school %d: %s", school, err)
			}
			courseIDs = append(courseIDs, courseID)
		}
		for j := 0; j < 2; j++ {
			teacherID := fmt.Sprintf("teacher_%s_%d_%d", stepState.Random, i, j)
			ctx, err := s.CommonSuite.AValidTeacherProfileWithId(ctx, teacherID, int32(school))
			if err != nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf("cannot init teacher%s", err)
			}
			teacherIDs = append(teacherIDs, teacherID)
		}

		// create lesson group
		lg := &entities_bob.LessonGroup{}
		database.AllNullEntity(lg)
		lg.CourseID = database.Text(courseIDs[0])
		lg.LessonGroupID = database.Text(idutil.ULIDNow())
		lg.CreatedAt = database.Timestamptz(now)
		err = (&bob_repo.LessonGroupRepo{}).Create(ctx, s.BobDB, lg)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("s.LessonGroupRepo.Create: %w", err)
		}
		// create lesson
		lesson := &entities_bob.Lesson{}
		database.AllNullEntity(lesson)
		lessonID := fmt.Sprintf("%s_id_%d", stepState.Random, i)
		lessonName := fmt.Sprintf("name_%d", i)

		stepState.FilterCourseIDs = courseIDs
		stepState.FilterLocationIDs = stepState.LocationIDs
		ctx, err = s.insertClass(ctx)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("cannot init class err: %s", err)
		}
		classID := stepState.FilterClassIDs[0]
		err := multierr.Combine(
			lesson.LessonID.Set(lessonID),
			lesson.Name.Set(lessonName),
			lesson.CourseID.Set(courseIDs[0]),
			lesson.TeacherID.Set(teacherIDs[0]),
			lesson.CreatedAt.Set(database.Timestamptz(time.Now())),
			lesson.UpdatedAt.Set(database.Timestamptz(time.Now())),
			lesson.LessonType.Set(cpb.LessonType_LESSON_TYPE_ONLINE.String()),
			lesson.Status.Set(cpb.LessonStatus_LESSON_STATUS_NOT_STARTED.String()),
			lesson.StreamLearnerCounter.Set(database.Int4(0)),
			lesson.LearnerIds.Set(database.JSONB([]byte("{}"))),
			lesson.StartTime.Set(startTime),
			lesson.EndTime.Set(endTime),
			lesson.TeachingMethod.Set(database.Text(cpb.LessonTeachingMethod_LESSON_TEACHING_METHOD_INDIVIDUAL.String())),
			lesson.TeachingMedium.Set(database.Text(cpb.LessonTeachingMedium_LESSON_TEACHING_MEDIUM_OFFLINE.String())),
			lesson.ClassID.Set(classID),
			lesson.CenterID.Set(centerID),
			lesson.LessonGroupID.Set(lg.LessonGroupID),
		)

		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("cannot init lesson err: %s", err)
		}

		if err := lesson.Normalize(); err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("lesson.Normalize err: %s", err)
		}
		cmdTag, err := database.Insert(ctx, lesson, s.BobPostgresDB.Exec)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("cannot insert lesson: %s", err)
		}
		if cmdTag.RowsAffected() != 1 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("cannot insert lesson")
		}
		stepState.LessonIDs = append(stepState.LessonIDs, lessonID)

		updateResourcePath := "UPDATE lessons SET resource_path = $1 WHERE lesson_id = $2"
		_, err = s.BobPostgresDB.Exec(ctx, updateResourcePath, database.Text(fmt.Sprint(school)), database.Text(lessonID))
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("cannot update lessons %w", err)
		}
		for j := 0; j < 2; j++ {
			sql := `insert into lessons_teachers
		(lesson_id, teacher_id, created_at)
		VALUES ($1, $2, $3)`
			_, err = s.BobDB.Exec(ctx, sql, lesson.LessonID, database.Text(teacherIDs[j]), database.Timestamptz(time.Now()))
			if err != nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf("cannot insert lessons_teachers %w", err)
			}
			stepState.FilterTeacherIDs = append(stepState.FilterTeacherIDs, teacherIDs[j])
		}

		if i%2 == 0 {
			for j := 0; j < 2; j++ {
				timeAny := database.Timestamptz(time.Now())
				sql := `insert into lessons_courses
		(lesson_id, course_id, created_at)
		VALUES ($1, $2, $3)`
				_, err = s.BobDB.Exec(ctx, sql, lesson.LessonID, database.Text(courseIDs[j]), timeAny)
				if err != nil {
					return StepStateToContext(ctx, stepState), fmt.Errorf("cannot insert lessons_courses %w", err)
				}
			}
		}
		if ctx, err := s.CommonSuite.CreateStudentAccounts(ctx); err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		stepState.CourseIDs = courseIDs[0:1]
		if _, err := s.CommonSuite.SomeStudentSubscriptions(ctx); err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("could not insert student subscription: %w", err)
		}
		err = s.updateLesson(ctx, lessonID, startTime, endTime, teacherIDs, centerID, stepState.StudentIds, courseIDs[0], classID)
		if err != nil {
			return StepStateToContext(ctx, stepState), stepState.ResponseErr
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) resultCorrectKeyword(ctx context.Context, keyword string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if len(keyword) == 0 {
		return StepStateToContext(ctx, stepState), nil
	}
	rsp := stepState.Response.(*lpb.RetrieveLessonsResponse)
	for _, item := range rsp.GetItems() {
		students, err := retrieveStudentsByLessonID(ctx, s, item.Id)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("retrieveStudentsByLessonID lessonId %s error %s ", item.Id, err)
		}
		existed := false
		for _, student := range students {
			parsedKeyword := strings.ToLower(strings.ReplaceAll(keyword, " ", ""))
			parsedName := strings.ToLower(strings.ReplaceAll(student.Name, " ", ""))
			parsedFullNamePhonetic := strings.ToLower(strings.ReplaceAll(student.FullNamePhonetic, " ", ""))
			if strings.Contains(parsedName, parsedKeyword) || strings.Contains(parsedFullNamePhonetic, parsedKeyword) {
				existed = true
				break
			}
		}
		if !existed {
			return StepStateToContext(ctx, stepState), fmt.Errorf("lesson %s doesn't contain any student with keyword %s", item.Id, keyword)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) resultCorrectDateManagement(ctx context.Context, dateRange string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if len(dateRange) == 0 {
		return StepStateToContext(ctx, stepState), nil
	}
	rsp := stepState.Response.(*lpb.RetrieveLessonsResponse)
	for _, lesson := range rsp.GetItems() {
		if stepState.FilterFromDate.After(lesson.EndTime.AsTime()) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("lesson %s not match from_date filter", lesson.Id)
		}
		if stepState.FilterToDate.Before(lesson.StartTime.AsTime()) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("lesson %s not match to_date filter", lesson.Id)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) resultCorrectTimeManagement(ctx context.Context, timeRange string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if len(timeRange) == 0 {
		return StepStateToContext(ctx, stepState), nil
	}
	rsp := stepState.Response.(*lpb.RetrieveLessonsResponse)
	for _, lesson := range rsp.GetItems() {
		if stepState.FilterFromTime.After(lesson.EndTime.AsTime()) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("lesson %s not match from_time filter", lesson.Id)
		}
		if stepState.FilterToTime.Before(lesson.StartTime.AsTime()) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("lesson %s not match to_time filter", lesson.Id)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) resultCorrectDateOfWeek(ctx context.Context, dow string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if len(dow) == 0 {
		return StepStateToContext(ctx, stepState), nil
	}
	rsp := stepState.Response.(*lpb.RetrieveLessonsResponse)
	for _, lesson := range rsp.GetItems() {
		dows := strings.Split(dow, ",")
		dowInts := make([]int32, len(dows))
		for i, s := range dows {
			val, _ := strconv.Atoi(s)
			dowInts[i] = int32(val)
		}
		if !checkExists(dowInts, int32(lesson.StartTime.AsTime().Weekday())) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("lesson %s not match date of week filter, expect %d, actual %d", lesson.Id, dowInts, int(lesson.StartTime.AsTime().Weekday()))
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) resultCorrectTeacherMedium(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	rsp := stepState.Response.(*lpb.RetrieveLessonsResponse)
	for _, lesson := range rsp.GetItems() {
		// if field is not null, check for inputs, else ignore
		if len(lesson.TeachingMedium.String()) > 0 {
			if lesson.TeachingMedium.String() != cpb.LessonTeachingMedium_LESSON_TEACHING_MEDIUM_HYBRID.String() &&
				lesson.TeachingMedium.String() != cpb.LessonTeachingMedium_LESSON_TEACHING_MEDIUM_OFFLINE.String() &&
				lesson.TeachingMedium.String() != cpb.LessonTeachingMedium_LESSON_TEACHING_MEDIUM_ONLINE.String() {
				return StepStateToContext(ctx, stepState), fmt.Errorf("lesson %s teaching medium %s not match teaching medium '%s', '%s' and '%s'",
					lesson.Id, lesson.TeachingMedium,
					cpb.LessonTeachingMedium_LESSON_TEACHING_MEDIUM_HYBRID.String(),
					cpb.LessonTeachingMedium_LESSON_TEACHING_MEDIUM_OFFLINE.String(),
					cpb.LessonTeachingMedium_LESSON_TEACHING_MEDIUM_ONLINE.String())
			}
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) resultCorrectTeachingMethod(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	rsp := stepState.Response.(*lpb.RetrieveLessonsResponse)
	for _, lesson := range rsp.GetItems() {
		// if field is not null, check for inputs, else ignore
		if len(lesson.TeachingMethod.String()) > 0 {
			if lesson.TeachingMethod.String() != cpb.LessonTeachingMethod_LESSON_TEACHING_METHOD_GROUP.String() &&
				lesson.TeachingMethod.String() != cpb.LessonTeachingMethod_LESSON_TEACHING_METHOD_INDIVIDUAL.String() {
				return StepStateToContext(ctx, stepState), fmt.Errorf("lesson %s teaching_method %s not match teaching_method '%s' and teaching_method '%s'",
					lesson.Id, lesson.TeachingMethod,
					cpb.LessonTeachingMethod_LESSON_TEACHING_METHOD_GROUP.String(),
					cpb.LessonTeachingMethod_LESSON_TEACHING_METHOD_INDIVIDUAL.String())
			}
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) resultCorrectCenter(ctx context.Context, centers, locations string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var filteredLocations []string
	if len(centers) == 0 && len(locations) == 0 {
		return StepStateToContext(ctx, stepState), nil
	}
	switch {
	case len(centers) == 0:
		filteredLocations = strings.Split(locations, ",")
	case len(locations) == 0:
		filteredLocations = strings.Split(centers, ",")
	default:
		filteredLocations = s.intersectLocations(centers, locations)
	}

	rsp := stepState.Response.(*lpb.RetrieveLessonsResponse)
	if len(filteredLocations) == 0 && len(rsp.GetItems()) > 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("should be empty item")
	} else {
		locationIds := make([]string, 0, len(filteredLocations))
		for _, v := range filteredLocations {
			val, _ := strconv.Atoi(strings.Split(v, "-")[1])
			locationIds = append(locationIds, stepState.FilterCenterIDs[val-1])
		}
		for _, lesson := range rsp.GetItems() {
			if !slices.Contains(locationIds, lesson.CenterId) {
				return StepStateToContext(ctx, stepState), fmt.Errorf("lesson %s center %s not match center filter %s", lesson.Id, lesson.CenterId, centers)
			}
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func isContainStatus(a []domain.LessonSchedulingStatus, l *lpb.RetrieveLessonsResponse_Lesson) bool {
	for _, v := range a {
		if l.SchedulingStatus.String() == string(v) {
			return true
		}
	}
	return false
}

func (s *Suite) resultCorrectSchedulingStatus(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	rsp := stepState.Response.(*lpb.RetrieveLessonsResponse)
	if len(stepState.FilterSchedulingStatuses) > 0 {
		for _, lesson := range rsp.GetItems() {
			if !isContainStatus(stepState.FilterSchedulingStatuses, lesson) {
				return StepStateToContext(ctx, stepState), fmt.Errorf("lesson %s scheduling status %s not match scheduling status filter %s", lesson.Id, lesson.SchedulingStatus, stepState.FilterSchedulingStatuses)
			}
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) resultCorrectClass(ctx context.Context, classes string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	rsp := stepState.Response.(*lpb.RetrieveLessonsResponse)
	if classes != "" && len(strings.Split(classes, ",")) > 0 {
		classIds := stepState.FilterClassIDs[0:len(strings.Split(classes, ","))]
		for _, lesson := range rsp.GetItems() {
			if !slices.Contains(classIds, lesson.ClassId) {
				return StepStateToContext(ctx, stepState), fmt.Errorf("lesson %s classId %s not match classIds filter %s", lesson.Id, lesson.ClassId, classIds)
			}
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) lessonmgmtMustReturnCorrectListLessonManagementWith(
	ctx context.Context,
	lessonTime,
	keyWord,
	dateRange,
	timeRange,
	teachers,
	students,
	coursers,
	centers,
	locations,
	grade, dow, schedulingStatus,
	classes string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.resultCorrectKeyword(ctx, keyWord)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("list result lesson not match with keyword: %s", err)
	}
	ctx, err = s.resultCorrectDateManagement(ctx, dateRange)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("list result lesson not match with date filter: %s", err)
	}
	ctx, err = s.resultCorrectTimeManagement(ctx, timeRange)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("list result lesson not match with time filter: %s", err)
	}
	ctx, err = s.resultCorrectDateOfWeek(ctx, dow)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("list result lesson not match with date of week filter: %s", err)
	}
	ctx, err = s.resultCorrectTeacherMedium(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("list result lesson not match with teaching medium: %s", err)
	}
	ctx, err = s.resultCorrectTeachingMethod(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("list result lesson not match with teaching method: %s", err)
	}
	ctx, err = s.resultCorrectCenter(ctx, centers, locations)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("list result lesson not match with center: %s", err)
	}
	ctx, err = s.resultCorrectSchedulingStatus(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("list result lesson not match with scheduling status: %s", err)
	}
	ctx, err = s.resultCorrectClass(ctx, classes)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("list result lesson not match with class: %s", err)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) lessonmgmtMustReturnListLessonManagement(ctx context.Context, total, fromID, toID, limit, next, pre string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	limitInt, err := strconv.Atoi(limit)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	rsp := stepState.Response.(*lpb.RetrieveLessonsResponse)
	if total != NIL_VALUE {
		totalInt, err := strconv.Atoi(total)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("err parse expect value")
		}
		if totalInt == 0 {
			if rsp.TotalLesson != uint32(totalInt) {
				return StepStateToContext(ctx, stepState), fmt.Errorf("must return empty list\n expect total: %d, actual total : %d", totalInt, rsp.TotalLesson)
			}
			return StepStateToContext(ctx, stepState), nil
		} else if rsp.TotalLesson != uint32(totalInt) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("wrong list return\n expect total: %d, actual: %d", totalInt, rsp.TotalLesson)
		}
	}
	if len(rsp.Items) > 0 {
		if rsp.Items[0].Id != (stepState.Random + "_" + fromID) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("wrong first items return \n expect: %s, actual: %s", (stepState.Random + "_" + fromID), rsp.Items[0].Id)
		}

		if rsp.Items[len(rsp.Items)-1].Id != (stepState.Random + "_" + toID) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("wrong items return \n expect: %s, actual: %s", (stepState.Random + "_" + toID), rsp.Items[len(rsp.Items)-1].Id)
		}

		if rsp.NextPage.GetOffsetString() != (stepState.Random + "_" + next) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("wrong next offset return \n expect: %s, actual: %s", (stepState.Random + "_" + next), rsp.NextPage.GetOffsetString())
		}
	}

	if int(rsp.NextPage.Limit) != limitInt || int(rsp.PreviousPage.Limit) != limitInt {
		return StepStateToContext(ctx, stepState), fmt.Errorf("wrong limit return")
	}

	if pre == NIL_VALUE {
		pre = ""
	} else {
		pre = stepState.Random + "_" + pre
	}

	if rsp.PreviousPage.GetOffsetString() != pre {
		return StepStateToContext(ctx, stepState), fmt.Errorf("wrong previous offset return \n expect: %s, actual: %s", pre, rsp.PreviousPage.GetOffsetString())
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) updateLesson(ctx context.Context, lessonID string, startTime, endTime time.Time, teacherIDs []string, centerID string, studentIDs []string, courseID string, classID string) error {
	stepState := StepStateFromContext(ctx)
	updateRequest := &bpb.UpdateLessonRequest{
		LessonId:        lessonID,
		StartTime:       timestamppb.New(startTime),
		EndTime:         timestamppb.New(endTime),
		TeachingMedium:  cpb.LessonTeachingMedium_LESSON_TEACHING_MEDIUM_OFFLINE,
		TeachingMethod:  cpb.LessonTeachingMethod_LESSON_TEACHING_METHOD_INDIVIDUAL,
		CourseId:        courseID,
		TeacherIds:      teacherIDs,
		CenterId:        centerID,
		ClassId:         classID,
		StudentInfoList: []*bpb.UpdateLessonRequest_StudentInfo{},
		SavingOption: &bpb.UpdateLessonRequest_SavingOption{
			Method: bpb.CreateLessonSavingMethod_CREATE_LESSON_SAVING_METHOD_ONE_TIME,
		},
	}
	for _, studentID := range studentIDs {
		updateRequest.StudentInfoList = append(updateRequest.StudentInfoList, &bpb.UpdateLessonRequest_StudentInfo{
			StudentId:        studentID,
			CourseId:         courseID,
			AttendanceStatus: bpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_ABSENT,
			LocationId:       centerID,
		})
	}

	stepState.Request = updateRequest
	stepState.Response, stepState.ResponseErr = bpb.NewLessonManagementServiceClient(s.BobConn).UpdateLesson(s.CommonSuite.SignedCtx(ctx), updateRequest)
	return stepState.ResponseErr
}

func (s *Suite) adminRetrieveLessonManagement(ctx context.Context, lessonTime, limitStr, offset string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	req := &lpb.RetrieveLessonsRequest{
		Paging: &cpb.Paging{
			Limit: uint32(limit),
		},
		LessonTime:  lpb.LessonTime(lpb.LessonTime_value[lessonTime]),
		CurrentTime: timestamppb.Now(),
	}

	if offset != NIL_VALUE {
		offset = stepState.Random + "_" + offset
		req = &lpb.RetrieveLessonsRequest{
			Paging: &cpb.Paging{
				Limit:  uint32(limit),
				Offset: &cpb.Paging_OffsetString{OffsetString: offset},
			},
			LessonTime:  lpb.LessonTime(lpb.LessonTime_value[lessonTime]),
			CurrentTime: timestamppb.Now(),
		}
	}
	stepState.Request = req
	return StepStateToContext(ctx, stepState), nil
}

func retrieveStudentsByLessonID(ctx context.Context, s *Suite, lessonID string) ([]*domain.UserBasicInfo, error) {
	userBasicInfo := &domain.UserBasicInfo{}
	fields, _ := userBasicInfo.FieldMap()
	q := fmt.Sprintf(`SELECT ubi.%s
			FROM lessons l join lesson_members lm on l.lesson_id = lm.lesson_id join user_basic_info ubi on ubi.user_id = lm.user_id
			WHERE l.lesson_id = $1 
			And ubi.deleted_at IS null
			AND ubi.user_id is not null
			AND lm.deleted_at IS NULL
	`, strings.Join(fields, ", ubi."))

	rows, err := s.LessonmgmtDB.Query(ctx, q, lessonID)
	if err != nil {
		return nil, fmt.Errorf("can't query user basic info: %s", err)
	}
	defer rows.Close()
	var results []*domain.UserBasicInfo
	for rows.Next() {
		var (
			userID            pgtype.Text
			name              pgtype.Text
			firstName         pgtype.Text
			lastName          pgtype.Text
			fullNamePhonetic  pgtype.Text
			firstNamePhonetic pgtype.Text
			lastNamePhonetic  pgtype.Text
			email             pgtype.Text
		)
		if err = rows.Scan(&userID, &name, &firstName,
			&lastName, &fullNamePhonetic, &firstNamePhonetic, &lastNamePhonetic,
			&email); err != nil {
			return nil, err
		}
		rs := &domain.UserBasicInfo{
			UserID:            userID.String,
			Name:              name.String,
			FirstName:         firstName.String,
			LastName:          lastName.String,
			FullNamePhonetic:  fullNamePhonetic.String,
			FirstNamePhonetic: firstNamePhonetic.String,
			LastNamePhonetic:  lastNamePhonetic.String,
			Email:             email.String,
		}
		results = append(results, rs)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return results, nil
}
