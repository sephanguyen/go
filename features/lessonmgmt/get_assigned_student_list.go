package lessonmgmt

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/manabie-com/backend/features/helper"
	entities_bob "github.com/manabie-com/backend/internal/bob/entities"
	bob_repo "github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/assigned_student/domain"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
	"golang.org/x/exp/slices"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type PackageType string

const (
	timezone = `Asia/Ho_Chi_Minh`

	studentCourseRecurringSlotOffsetIDQuery = `WITH offsetQuery as (
		SELECT (at.unique_id || '_' || at.week_start) AS offset_id, row_number() over(ORDER BY at.week_start ASC, at.course_id ASC, at.student_id ASC, (at.unique_id || '_' || at.week_start) ASC) AS row_num
		FROM student_course_recurring_slot_info_fn(null, null, null, null, null, null, '%s', null, false) at 
		WHERE (current_timestamp at time zone '%s')::DATE < at.week_end 
		) SELECT offset_id FROM offsetQuery WHERE row_num = $1 LIMIT 1`

	studentCourseSlotOffsetIDQuery = `WITH offsetQuery AS (
		SELECT at.unique_id AS offset_id, row_number() over(ORDER BY at.student_start_date ASC, at.course_id ASC, at.student_id ASC, at.unique_id ASC) AS row_num
		FROM student_course_slot_info_fn(null, null, null, null, null, null, '%s', null, false) at 
		WHERE (current_timestamp at time zone '%s')::DATE < at.student_end_date 
		) SELECT offset_id FROM offsetQuery WHERE row_num = $1 LIMIT 1`

	studentCourseRecurringSlotTotalQuery = `SELECT COUNT((at.unique_id || '_' || at.week_start))
		FROM student_course_recurring_slot_info_fn(null, null, null, null, null, null, '%s', null, false) at 
		WHERE (current_timestamp at time zone '%s')::DATE < at.week_end `

	studentCourseSlotTotalQuery = `SELECT COUNT(at.unique_id)
		FROM student_course_slot_info_fn(null, null, null, null, null, null, '%s', null, false) at 
		WHERE (current_timestamp at time zone '%s')::DATE < at.student_end_date `

	OneTime   PackageType = "PACKAGE_TYPE_ONE_TIME"
	SlotBased PackageType = "PACKAGE_TYPE_SLOT_BASED"
	Frequency PackageType = "PACKAGE_TYPE_FREQUENCY"
	Scheduled PackageType = "PACKAGE_TYPE_SCHEDULED"
)

func (s *Suite) adminGetAssignedStudentListNoFilter(ctx context.Context, purchaseMethod, pageNumber, pageLimit string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.PurchaseMethod = purchaseMethod

	page, err := strconv.Atoi(pageNumber)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("failed to convert page from string to int: %w", err)
	}
	stepState.PageNumber = page

	limit, err := strconv.Atoi(pageLimit)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("failed to convert limit from string to int: %w", err)
	}
	stepState.PageLimit = limit

	ctx, err = s.getAssignedStudentListOffsetTotalDetails(StepStateToContext(ctx, stepState))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.getAssignedStudentListOffsetTotalDetail %w", err)
	}
	stepState = StepStateFromContext(ctx)
	offset := stepState.Offset

	req := &lpb.GetAssignedStudentListRequest{
		PurchaseMethod: lpb.PurchaseMethod(lpb.PurchaseMethod_value[purchaseMethod]),
		Paging: &cpb.Paging{
			Limit: uint32(limit),
		},
		Timezone: timezone,
	}

	if len(offset) > 0 {
		req.Paging.Offset = &cpb.Paging_OffsetString{OffsetString: offset}
	}

	stepState.Request = req
	stepState.Response, stepState.ResponseErr = lpb.NewAssignedStudentListServiceClient(s.LessonMgmtConn).
		GetAssignedStudentList(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) aListOfPurchasedStudentAreExistedInDB(ctx context.Context, strFromID, strToID string, duration int, purchaseMethod, location string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	school := stepState.CurrentSchoolID
	now := time.Now().Add(1000 * 24 * time.Hour)

	fromID, err := strconv.Atoi((strings.Split(strFromID, "_"))[1])
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("cannot init lesson with outline init param")
	}
	toID, err := strconv.Atoi((strings.Split(strToID, "_"))[1])
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("cannot init lesson with outline init param")
	}

	totalStudents := toID - fromID

	startDate := calculateStartDateToMonday()
	indexLocation, _ := strconv.Atoi(location)
	locationID := stepState.CenterIDs[indexLocation]
	for i := fromID; i < toID; i++ {
		stepState.FilterCenterIDs = append(stepState.FilterCenterIDs, locationID)

		// step 1: create student
		studentID := fmt.Sprintf("asg-std-%d", i)
		studentName := fmt.Sprintf("student name-%d", i)
		ctx, err := s.AValidStudentProfileWithId(ctx, studentID, studentName)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("step 1 failed: cannot create new student: %w", err)
		}
		stepState.FilterStudentIDs = append(stepState.FilterStudentIDs, studentID)

		// step 2: create teacher
		teacherID := idutil.ULIDNow()
		ctx, err = s.CommonSuite.AValidTeacherProfileWithId(ctx, teacherID, school)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("step 2 failed: cannot init teacher%w", err)
		}
		stepState.TeacherIDs = append(stepState.TeacherIDs, teacherID)

		// step 3: create course
		courseID := fmt.Sprintf("asg-course-%d", i)
		courseName := fmt.Sprintf("[AS] Course %d", i)

		ctx, err = s.createCourseWithID(ctx, courseID, courseName, school)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("step 3 failed: cannot init course with id %s: %w", courseID, err)
		}
		stepState.FilterCourseIDs = append(stepState.FilterCourseIDs, courseID)
		stepState.CourseIDs = append(stepState.CourseIDs, courseID)
		stepState.FilterLocationIDs = stepState.LocationIDs
		stepState.FilterGradeIDs = stepState.GradeIDs
		// step 4: create lesson group
		lg := &entities_bob.LessonGroup{}
		database.AllNullEntity(lg)
		lg.CourseID = database.Text(courseID)
		lg.LessonGroupID = database.Text(idutil.ULIDNow())
		lg.CreatedAt = database.Timestamptz(now)
		err = (&bob_repo.LessonGroupRepo{}).Create(ctx, s.BobDB, lg)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("step 4 failed: cannot create lesson group %w", err)
		}

		// step 5: create lesson
		totalLessons := rand.Intn(totalStudents + i)
		for j := 0; j < totalLessons; j++ {
			lesson := &entities_bob.Lesson{}
			database.AllNullEntity(lesson)
			lessonID := idutil.ULIDNow()
			lessonName := "Lesson name management " + lessonID
			ctx, err = s.insertClass(ctx)
			if err != nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf("cannot init class err: %w", err)
			}
			classId := stepState.FilterClassIDs[0]

			duration := time.Duration(j)
			startTime := startDate.Add(duration * time.Hour).Add(3 * 24 * time.Hour)
			endTime := startDate.Add(duration*time.Hour + time.Minute*2).Add(3 * 24 * time.Hour)
			if j > 11 {
				startTime = startDate.Add(-duration * time.Hour)
				endTime = startDate.Add(-duration*time.Hour + time.Minute*2)
			}
			centerID := locationID

			status := cpb.LessonSchedulingStatus_name[1]
			teachingMethod := cpb.LessonTeachingMethod_LESSON_TEACHING_METHOD_GROUP.String()
			stepState.CurrentTeachingMethod = "group"

			err = multierr.Combine(
				lesson.LessonID.Set(lessonID),
				lesson.Name.Set(lessonName),
				lesson.CourseID.Set(courseID),
				lesson.TeacherID.Set(teacherID),
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
				return StepStateToContext(ctx, stepState), fmt.Errorf("lesson multierr.Combine err: %w", err)
			}
			if err := lesson.Normalize(); err != nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf("lesson.Normalize err: %s", err)
			}
			cmdTag, err := database.Insert(ctx, lesson, s.BobDB.Exec)
			if err != nil {
				return StepStateToContext(ctx, stepState), err
			}
			if cmdTag.RowsAffected() != 1 {
				return StepStateToContext(ctx, stepState), fmt.Errorf("step 5 failed: cannot insert lesson")
			}
			stepState.LessonIDs = append(stepState.LessonIDs, lessonID)
			updateResourcePath := "UPDATE lessons SET resource_path = $1 WHERE lesson_id = $2"
			_, err = s.BobDB.Exec(ctx, updateResourcePath, database.Text(fmt.Sprint(school)), database.Text(lessonID))
			if err != nil {
				return StepStateToContext(ctx, stepState), err
			}
			for k := 0; k < 2; k++ {
				sql := `insert into lessons_teachers (lesson_id, teacher_id, created_at) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING`
				_, err = s.BobDB.Exec(ctx, sql, lesson.LessonID, database.Text(teacherID), database.Timestamptz(time.Now()))
				if err != nil {
					return StepStateToContext(ctx, stepState), err
				}
				stepState.FilterTeacherIDs = append(stepState.FilterTeacherIDs, teacherID)
			}

			if j%2 == 0 {
				for k := 0; k < 2; k++ {
					timeAny := database.Timestamptz(time.Now())
					sql := `insert into lessons_courses (lesson_id, course_id, created_at) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING`
					_, err = s.BobDB.Exec(ctx, sql, lesson.LessonID, database.Text(courseID), timeAny)
					if err != nil {
						return StepStateToContext(ctx, stepState), err
					}
				}
			}

			// step 6: assign student to lesson -> lesson_members
			sql := `INSERT INTO lesson_members
				(lesson_id, user_id, course_id, created_at, updated_at, resource_path)
				VALUES ($1, $2, $3, $4, $5, $6)
				ON CONFLICT DO NOTHING`
			_, err = s.BobDB.Exec(ctx, sql,
				database.Text(lessonID),
				database.Text(studentID),
				database.Text(courseID),
				time.Now(), time.Now(),
				database.Text(fmt.Sprint(school)),
			)
			if err != nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf("step 6 failed: cannot init lesson_members: %w", err)
			}
		}

		// step 7: register student_course
		studentPackageID := fmt.Sprintf("asg-std-pckg-%d", i)

		var (
			startTime         time.Time
			endTime           time.Time
			packageType       PackageType
			courseSlot        int64
			courseSlotPerWeek int64
		)

		duration := time.Duration(duration)
		startTime = startDate.Add(duration * time.Hour)
		endTime = startDate.Add(duration*time.Hour*24 + time.Minute)

		if purchaseMethod == string(domain.PurchaseMethodSlot) {
			courseSlot = helper.RandInt(1, int64(totalStudents))
			courseSlotPerWeek = int64(0)
			packageType = SlotBased
		}

		if purchaseMethod == string(domain.PurchaseMethodRecurring) {
			courseSlot = int64(0)
			courseSlotPerWeek = helper.RandInt(1, int64(totalStudents))
			packageType = Frequency
		}

		studentStartDate := startTime.Format(timeLayout)
		studentEndDate := endTime.Format(timeLayout)

		sql := `INSERT INTO student_course
		(student_id, course_id, location_id, student_package_id, student_start_date, student_end_date, course_slot, course_slot_per_week, resource_path, package_type, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5::Date, $6::Date, $7, $8, $9, $10, now(), now())
		ON CONFLICT ON CONSTRAINT student_course_pk 
		DO UPDATE SET student_start_date = $5::Date, student_end_date = $6::Date, course_slot = $7, course_slot_per_week = $8, resource_path = $9, package_type = $10, updated_at = now()`
		_, err = s.BobDB.Exec(ctx, sql,
			database.Text(studentID),
			database.Text(courseID),
			database.Text(locationID),
			database.Text(studentPackageID),
			database.Text(studentStartDate),
			database.Text(studentEndDate),
			database.Int4(int32(courseSlot)),
			database.Int4(int32(courseSlotPerWeek)),
			database.Text(fmt.Sprint(school)),
			database.Text(string(packageType)),
		)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("step 7 failed: cannot init lesson_student_subscriptions: %w", err)
		}
	}

	// cleanup student_course slot columns
	updateSlotNum := `UPDATE student_course 
	SET course_slot = (CASE WHEN course_slot = 0 THEN NULL ELSE course_slot END),
	course_slot_per_week = (CASE WHEN course_slot_per_week = 0 THEN NULL ELSE course_slot_per_week END)
	WHERE package_type = ANY($1)`

	_, err = s.BobDB.Exec(ctx, updateSlotNum, []string{string(OneTime), string(SlotBased), string(Frequency), string(Scheduled)})
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("failed to cleanup slot info: %v", err)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) createCourseWithID(ctx context.Context, id, name string, school int32) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	courseRepo := bob_repo.CourseRepo{}

	timeAny := database.Timestamptz(time.Now())
	course := &entities_bob.Course{}
	database.AllNullEntity(course)
	err := multierr.Combine(
		course.ID.Set(id),
		course.Name.Set(name),
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
	err = courseRepo.Upsert(ctx, s.BobDB, []*entities_bob.Course{course})
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("cannot init course err with school %d: %w", school, err)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) AValidStudentProfileWithId(ctx context.Context, id, name string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.CurrentSchoolID == 0 {
		stepState.CurrentSchoolID = 1
	}
	if name == "" {
		name = fmt.Sprintf("valid-student-%s", id)
	}

	student := &entities_bob.Student{}
	database.AllNullEntity(student)
	database.AllNullEntity(&student.User)
	database.AllNullEntity(&student.User.AppleUser)

	now := time.Now()
	err := multierr.Combine(
		student.ID.Set(id),
		student.LastName.Set(name),
		student.Country.Set(pb.COUNTRY_VN.String()),
		student.PhoneNumber.Set(fmt.Sprintf("phone-number+%s", id)),
		student.Email.Set(fmt.Sprintf("email+%s@example.com", id)),
		student.CurrentGrade.Set(12),
		student.TargetUniversity.Set("TG11DT"),
		student.TotalQuestionLimit.Set(5),
		student.SchoolID.Set(stepState.CurrentSchoolID),
		student.UpdatedAt.Set(now),
		student.CreatedAt.Set(now),
		student.Group.Set(entities_bob.UserGroupStudent),
		student.OnTrial.Set(false),
		student.BillingDate.Set(now.Add(0)),
		student.EnrollmentStatus.Set("STUDENT_ENROLLMENT_STATUS_ENROLLED"),
		student.StudentNote.Set(""),

		student.User.ID.Set(student.ID.String),
		student.User.UpdatedAt.Set(now),
		student.User.CreatedAt.Set(now),
		student.User.DeviceToken.Set(nil),
		student.User.AllowNotification.Set(true),
	)
	if student.User.ResourcePath.Status != pgtype.Present {
		resourcePath := golibs.ResourcePathFromCtx(ctx)
		student.User.ResourcePath.Set(resourcePath)
	}
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("err set entity: %w", err)
	}
	// upsert user
	_, err = database.InsertOnConflictDoNothing(ctx, &student.User, s.BobDB.Exec)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("user not inserted: %w", err)
	}
	// upsert student
	_, err = database.InsertOnConflictDoNothing(ctx, student, s.BobDB.Exec)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("student not inserted: %w", err)
	}
	// upsert user group
	group := &entities_bob.UserGroup{}
	err = multierr.Combine(
		group.UserID.Set(student.ID.String),
		group.GroupID.Set(entities_bob.UserGroupStudent),
		group.IsOrigin.Set(true),
		group.Status.Set(entities_bob.UserGroupStatusActive),
		group.CreatedAt.Set(now),
		group.UpdatedAt.Set(now),
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("err set UserGroup: %w", err)
	}

	_, err = database.InsertOnConflictDoNothing(ctx, group, s.BobDB.Exec)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("err insert UserGroup: %w", err)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) mustReturnCorrectAssignedStudentList(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	expectedPageLimit := stepState.PageLimit
	expectedTotal := stepState.TotalRecords
	expectedPreOffset := stepState.PreOffset
	expectedNextOffset := stepState.NextOffset

	res := stepState.Response.(*lpb.GetAssignedStudentListResponse)

	// compare total
	actualTotal := int(res.TotalItems)

	if expectedTotal == 0 && actualTotal != expectedTotal {
		return StepStateToContext(ctx, stepState), fmt.Errorf("must return empty list\n expected total: %d, actual total : %d", expectedTotal, actualTotal)
	} else if actualTotal != expectedTotal {
		return StepStateToContext(ctx, stepState), fmt.Errorf("incorrect total records returned\n expected total: %d, actual: %d", expectedTotal, actualTotal)
	}

	// compare next offset
	if len(expectedNextOffset) > 0 {
		if len(res.Items) == 0 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expected to return list but got empty")
		}

		actualNextOffset := res.NextPage.GetOffsetString()
		if actualNextOffset != expectedNextOffset {
			return StepStateToContext(ctx, stepState), fmt.Errorf("wrong next offset returned \n expected: %s, actual: %s", expectedNextOffset, actualNextOffset)
		}
	}

	// compare pre offset
	if len(expectedPreOffset) > 0 {
		if len(res.Items) == 0 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expected to return list but got empty")
		}

		actualPreOffset := res.PreviousPage.GetOffsetString()
		if actualPreOffset != expectedPreOffset {
			return StepStateToContext(ctx, stepState), fmt.Errorf("wrong pre offset returned \n expected: %s, actual: %s", expectedPreOffset, actualPreOffset)
		}
	}

	// compare Limit
	if int(res.NextPage.Limit) != expectedPageLimit || int(res.PreviousPage.Limit) != expectedPageLimit {
		return StepStateToContext(ctx, stepState), fmt.Errorf("wrong limit returned")
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) adminGetAssignedStudentListWithFilter(ctx context.Context, purchaseMethod, limit, keyword, students, courses, centers, locations, status, daysGapStartDate, daysGapEndDate string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.PurchaseMethod = purchaseMethod
	filter := &lpb.GetAssignedStudentListRequest_Filter{}
	startDate := stepState.StartDate

	pageLimit, err := strconv.Atoi(limit)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("failed to convert limit from string to int: %w", err)
	}
	stepState.PageLimit = pageLimit

	// start date filter
	if daysGapStartDate != NIL_VALUE {
		convertedDaysGap, err := strconv.Atoi(daysGapStartDate)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("failed to convert days gap of start date from string to int: %w", err)
		}
		filter.StartDate = timestamppb.New(startDate.AddDate(0, 0, convertedDaysGap))
	}

	// end date filter
	if daysGapEndDate != NIL_VALUE {
		convertedDaysGap, err := strconv.Atoi(daysGapEndDate)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("failed to convert days gap of end date from string to int: %w", err)
		}
		filter.EndDate = timestamppb.New(startDate.AddDate(0, 0, convertedDaysGap))
	}

	// students filter
	if len(strings.TrimSpace(students)) > 0 && len(strings.Split(students, ",")) > 0 {
		filter.StudentIds = stepState.FilterStudentIDs[0:len(strings.Split(students, ","))]
	}

	// courses filter
	if len(strings.TrimSpace(courses)) > 0 && len(strings.Split(courses, ",")) > 0 {
		filter.CourseIds = stepState.FilterCourseIDs[0:len(strings.Split(courses, ","))]
	}

	// locations request
	locationIds := make([]string, 0, len(locations))
	if len(strings.TrimSpace(locations)) > 0 {
		locationNames := strings.Split(locations, ",")

		for _, v := range locationNames {
			val, err := strconv.Atoi(v)
			if err != nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf("failed to convert location ID from string to int: %w", err)
			}

			locationIds = append(locationIds, stepState.FilterCenterIDs[val-1])
		}
	}

	// centers filter
	if len(strings.TrimSpace(centers)) > 0 {
		centerNames := strings.Split(centers, ",")

		for _, v := range centerNames {
			val, err := strconv.Atoi(v)
			if err != nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf("failed to convert center ID from string to int: %w", err)
			}

			filter.LocationIds = append(filter.LocationIds, stepState.FilterCenterIDs[val-1])
		}
	}

	// status filter
	if len(strings.TrimSpace(status)) > 0 {
		statusesString := strings.Split(status, ",")
		statuses := make([]domain.AssignedStudentStatus, 0, len(statusesString))
		for _, s := range statusesString {
			statuses = append(statuses, domain.AssignedStudentStatus(s))
			filter.Statuses = append(filter.Statuses, lpb.AssignedStudentStatus(lpb.AssignedStudentStatus_value[s]))
		}
		stepState.FilterAssignedStatus = statuses
	}

	req := &lpb.GetAssignedStudentListRequest{
		Paging: &cpb.Paging{
			Limit: uint32(pageLimit),
		},
		PurchaseMethod: lpb.PurchaseMethod(lpb.PurchaseMethod_value[purchaseMethod]),
		Filter:         filter,
		Keyword:        keyword,
		LocationIds:    locationIds,
	}

	stepState.Request = req
	stepState.Response, stepState.ResponseErr = lpb.NewAssignedStudentListServiceClient(s.LessonMgmtConn).
		GetAssignedStudentList(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) mustReturnCorrectAssignedStudentListWithFilters(ctx context.Context, keyWord, students, coursers, centers, locations, status string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.asgStudentResultCorrectCenter(ctx, centers, locations)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("resulting list does not match with center/s: %w", err)
	}

	ctx, err = s.asgStudentResultCorrectStudent(ctx, students)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("resulting list does not match with student/s: %w", err)
	}

	ctx, err = s.asgStudentResultCorrectCourse(ctx, coursers)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("resulting list does not match with course/s: %w", err)
	}

	ctx, err = s.asgStudentResultCorrectStatus(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("resulting list does not match with status: %w", err)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) asgStudentResultCorrectStudent(ctx context.Context, students string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	rsp := stepState.Response.(*lpb.GetAssignedStudentListResponse)

	if students != "" && len(strings.Split(students, ",")) > 0 {
		studentIds := stepState.FilterStudentIDs[0:len(strings.Split(students, ","))]

		for _, item := range rsp.GetItems() {
			if !slices.Contains(studentIds, item.StudentId) {
				return StepStateToContext(ctx, stepState), fmt.Errorf("student %s not match studentIds filter %s", item.StudentId, studentIds)
			}
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) asgStudentResultCorrectStatus(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	rsp := stepState.Response.(*lpb.GetAssignedStudentListResponse)

	for _, item := range rsp.GetItems() {
		if len(item.Status.String()) > 0 {
			if item.Status.String() != lpb.AssignedStudentStatus_STUDENT_STATUS_JUST_ASSIGNED.String() &&
				item.Status.String() != lpb.AssignedStudentStatus_STUDENT_STATUS_OVER_ASSIGNED.String() &&
				item.Status.String() != lpb.AssignedStudentStatus_STUDENT_STATUS_UNDER_ASSIGNED.String() {
				return StepStateToContext(ctx, stepState), fmt.Errorf("student %s assigned status %s not match '%s', '%s' and '%s'",
					item.StudentId, item.Status,
					lpb.AssignedStudentStatus_STUDENT_STATUS_JUST_ASSIGNED.String(),
					lpb.AssignedStudentStatus_STUDENT_STATUS_OVER_ASSIGNED.String(),
					lpb.AssignedStudentStatus_STUDENT_STATUS_UNDER_ASSIGNED.String())
			}
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) asgStudentResultCorrectCourse(ctx context.Context, coursers string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	rsp := stepState.Response.(*lpb.GetAssignedStudentListResponse)

	if coursers != "" && len(strings.Split(coursers, ",")) > 0 {
		courseIds := stepState.FilterCourseIDs[0:len(strings.Split(coursers, ","))]

		for _, item := range rsp.GetItems() {
			if !slices.Contains(courseIds, item.CourseId) {
				return StepStateToContext(ctx, stepState), fmt.Errorf("student %s courseId %s not match courseIds filter %s", item.StudentId, item.CourseId, courseIds)
			}
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) asgStudentResultCorrectCenter(ctx context.Context, centers, locations string) (context.Context, error) {
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

	rsp := stepState.Response.(*lpb.GetAssignedStudentListResponse)

	if len(filteredLocations) == 0 && len(rsp.GetItems()) > 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("should be empty item")
	} else {
		locationIds := make([]string, 0, len(filteredLocations))
		for _, v := range filteredLocations {
			val, _ := strconv.Atoi(v)
			locationIds = append(locationIds, stepState.FilterCenterIDs[val-1])
		}
		for _, asgStudent := range rsp.GetItems() {
			if !slices.Contains(locationIds, asgStudent.LocationId) {
				return StepStateToContext(ctx, stepState), fmt.Errorf("asgStudent %s center %s not match center filter %s", asgStudent.StudentId, asgStudent.LocationId, centers)
			}
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) getAssignedStudentListOffsetTotalDetails(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	pageNumber := stepState.PageNumber
	pageLimit := stepState.PageLimit
	purchaseMethod := stepState.PurchaseMethod
	var (
		totalQuery  string
		OffsetQuery string
		total       int
		offsetID    string
	)

	if purchaseMethod == string(domain.PurchaseMethodSlot) {
		totalQuery = fmt.Sprintf(studentCourseSlotTotalQuery, timezone, timezone)
		OffsetQuery = fmt.Sprintf(studentCourseSlotOffsetIDQuery, timezone, timezone)
	}

	if purchaseMethod == string(domain.PurchaseMethodRecurring) {
		totalQuery = fmt.Sprintf(studentCourseRecurringSlotTotalQuery, timezone, timezone)
		OffsetQuery = fmt.Sprintf(studentCourseRecurringSlotOffsetIDQuery, timezone, timezone)
	}

	// get total records
	row := s.BobDB.QueryRow(ctx, totalQuery)
	if err := row.Scan(&total); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("failed to get total of records: %w", err)
	}
	stepState.TotalRecords = total
	totalPages := int(math.Ceil(float64(total) / float64(pageLimit)))

	// get offset string
	if pageNumber > 1 {
		rowNum := pageLimit * (pageNumber - 1)

		row = s.BobDB.QueryRow(ctx, OffsetQuery, rowNum)
		if err := row.Scan(&offsetID); err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("failed to get offset string: %w", err)
		}
		stepState.Offset = offsetID
	}

	// get pre-offset string
	if pageNumber > 2 {
		rowNum := pageLimit * (pageNumber - 2)

		row = s.BobDB.QueryRow(ctx, OffsetQuery, rowNum)
		if err := row.Scan(&offsetID); err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("failed to get offset string: %w", err)
		}
		stepState.PreOffset = offsetID
	}

	// get next-offset string
	if pageNumber > totalPages {
		rowNum := pageLimit * pageNumber

		row = s.BobDB.QueryRow(ctx, OffsetQuery, rowNum)
		if err := row.Scan(&offsetID); err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("failed to get offset string: %w", err)
		}
		stepState.NextOffset = offsetID
	}

	return StepStateToContext(ctx, stepState), nil
}

func calculateStartDateToMonday() time.Time {
	loc := LoadLocalLocation()
	currentTimestamp := time.Now().In(loc)

	weekday := int(currentTimestamp.Weekday())

	// move the date to the next monday if date is not monday (1)
	if weekday != 1 {
		modifier := time.Duration((7 + (1 - weekday)) % 7)
		currentTimestamp = currentTimestamp.Add(modifier * 24 * time.Hour)
	}

	startDate := time.Date(currentTimestamp.Year(), currentTimestamp.Month(), currentTimestamp.Day(), 9, 0, 0, 0, loc)

	return startDate
}
