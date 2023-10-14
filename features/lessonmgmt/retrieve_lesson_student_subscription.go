package lessonmgmt

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/timeutil"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/class/domain"
	master_repo "github.com/manabie-com/backend/internal/mastermgmt/modules/class/infrastructure/repo"
	location_repo "github.com/manabie-com/backend/internal/mastermgmt/modules/location/infrastructure/repo"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Suite) aListStudentSubscriptionsAreExistedInDB(ctx context.Context, total, schoolID int, _startDate, _endDate string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.CommonSuite.ASignedInWithSchool(contextWithResourcePath(ctx, i32ToStr(schoolID)), "school admin", stepState.CurrentSchoolID)
	if err != nil {
		return ctx, err
	}

	s.clearPreviousTestStudentSubscriptions(ctx, schoolID)
	// init courses
	stepState.FilterCourseIDs = []string{idutil.ULIDNow(), idutil.ULIDNow()}
	stepState.FilterLocationIDs = stepState.LocationIDs

	for _, courseID := range stepState.FilterCourseIDs {
		courseName := "name-" + courseID
		sql := `insert into courses
		(course_id, name, grade, created_at, updated_at, school_id, resource_path)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`
		resourcePath := fmt.Sprint(schoolID)
		_, err := s.BobDB.Exec(ctx, sql, database.Text(courseID), courseName, 12, time.Now(), time.Now(), schoolID, resourcePath)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("cannot init course err: %s and %d and %s ", err, schoolID, resourcePath)
		}
	}

	// init class
	for i := 0; i < 4; i++ {
		ctx, err = s.insertClass(ctx)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("cannot init class err: %s", err)
		}
	}

	startDate, _ := time.Parse(time.RFC3339, _startDate)
	endDate, _ := time.Parse(time.RFC3339, _endDate)

	// create students and student subscriptions
	filterStudentSubs := []string{}
	num := 0
	for i := 1; i <= (total / 2); i++ {
		studentName := "student name -"
		if ctx, err := s.CommonSuite.CreateATotalNumberOfStudentAccounts(ctx, studentName, 1); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		for _, courseID := range stepState.FilterCourseIDs {
			num++
			studentSubscriptionId, subscriptionId := idutil.ULIDNow(), idutil.ULIDNow()
			filterStudentSubs = append(filterStudentSubs, studentSubscriptionId)
			duration := time.Duration(20 - num)
			createdAt := timeutil.Now().Add(duration * time.Hour)
			sql := `insert into lesson_student_subscriptions
			(student_subscription_id, course_id, student_id, subscription_id, start_at, end_at, created_at, updated_at, resource_path)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
			_, err := s.BobDB.Exec(ctx, sql, database.Text(studentSubscriptionId), database.Text(courseID), stepState.StudentIds[0], subscriptionId, startDate, endDate, createdAt, time.Now(), database.Text(fmt.Sprint(schoolID)))
			if err != nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf("cannot init lesson_student_subscriptions err: %s", err)
			}

			sql = `INSERT INTO lesson_student_subscription_access_path
			(student_subscription_id, location_id)
			VALUES($1, $2);`

			_, err = s.BobDB.Exec(ctx, sql, database.Text(studentSubscriptionId), stepState.FilterLocationIDs[0])
			if err != nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf("cannot init lesson_student_subscription_access_path err: %s", err)
			}
		}
		sql := `INSERT INTO class_member
	(class_member_id, class_id, user_id, created_at, updated_at)
	VALUES($1, $2, $3, timezone('utc'::text, now()), timezone('utc'::text, now())) ON CONFLICT DO NOTHING;`

		_, err := s.BobDB.Exec(ctx, sql, database.Text(idutil.ULIDNow()), stepState.FilterClassIDs[0], stepState.StudentIds[0])
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("cannot init class_member err: %s", err)
		}
	}

	stepState.FilterStudentSubs = filterStudentSubs
	return StepStateToContext(ctx, stepState), nil
}
func (s *Suite) studentSubscriptionsOfSchoolAreExistedInDBWithAndUsingUserBasicInfoTable(ctx context.Context, total, schoolID int, _startDate, _endDate string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.CommonSuite.ASignedInWithSchool(contextWithResourcePath(ctx, i32ToStr(schoolID)), "school admin", stepState.CurrentSchoolID)
	if err != nil {
		return ctx, err
	}

	s.clearPreviousTestStudentSubscriptions(ctx, schoolID)
	// init courses
	stepState.FilterCourseIDs = []string{idutil.ULIDNow(), idutil.ULIDNow()}
	stepState.FilterLocationIDs = stepState.LocationIDs

	for _, courseID := range stepState.FilterCourseIDs {
		courseName := "name-" + courseID
		sql := `insert into courses
		(course_id, name, grade, created_at, updated_at, school_id)
		VALUES ($1, $2, $3, $4, $5, $6)`
		_, err := s.BobDB.Exec(ctx, sql, database.Text(courseID), courseName, 12, time.Now(), time.Now(), schoolID)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("cannot init course err: %s", err)
		}
	}

	// init class
	for i := 0; i < 4; i++ {
		ctx, err = s.insertClass(ctx)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("cannot init class err: %s", err)
		}
	}

	startDate, _ := time.Parse(time.RFC3339, _startDate)
	endDate, _ := time.Parse(time.RFC3339, _endDate)

	// create students and student subscriptions
	filterStudentSubs := []string{}
	num := 0
	for i := 1; i <= (total / 2); i++ {
		studentName := "student name -"
		if ctx, err := s.CommonSuite.CreateATotalNumberOfStudentAccountsInUserBasicInfoTable(ctx, studentName, 1); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		for _, courseID := range stepState.FilterCourseIDs {
			num++
			studentSubscriptionId, subscriptionId := idutil.ULIDNow(), idutil.ULIDNow()
			filterStudentSubs = append(filterStudentSubs, studentSubscriptionId)
			duration := time.Duration(20 - num)
			createdAt := timeutil.Now().Add(duration * time.Hour)
			sql := `insert into lesson_student_subscriptions
			(student_subscription_id, course_id, student_id, subscription_id, start_at, end_at, created_at, updated_at, resource_path)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
			_, err := s.BobDB.Exec(ctx, sql, database.Text(studentSubscriptionId), database.Text(courseID), stepState.StudentIds[0], subscriptionId, startDate, endDate, createdAt, time.Now(), database.Text(fmt.Sprint(schoolID)))
			if err != nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf("cannot init course err: %s", err)
			}

			sql = `INSERT INTO lesson_student_subscription_access_path
			(student_subscription_id, location_id)
			VALUES($1, $2);`

			_, err = s.BobDB.Exec(ctx, sql, database.Text(studentSubscriptionId), stepState.FilterLocationIDs[0])
			if err != nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf("cannot init lesson_student_subscription_access_path err: %s", err)
			}
		}
		sql := `INSERT INTO class_member
	(class_member_id, class_id, user_id, created_at, updated_at)
	VALUES($1, $2, $3, timezone('utc'::text, now()), timezone('utc'::text, now())) ON CONFLICT DO NOTHING;`

		_, err := s.BobDB.Exec(ctx, sql, database.Text(idutil.ULIDNow()), stepState.FilterClassIDs[0], stepState.StudentIds[0])
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("cannot init class_member err: %s", err)
		}
	}

	stepState.FilterStudentSubs = filterStudentSubs
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) insertClass(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	courseIDs := stepState.FilterCourseIDs
	locationIDs := stepState.FilterLocationIDs
	fields := []string{"class_id", "name", "course_id", "location_id", "school_id", "created_at", "updated_at"}
	query := fmt.Sprintf("INSERT INTO class (%s) VALUES ($1,$2,$3,$4,$5,$6,$7)",
		strings.Join(fields, ","))
	var courseID, locationID string
	classID := idutil.ULIDNow()
	schoolID := golibs.ResourcePathFromCtx(ctx)
	className := "name"
	if len(courseIDs) > 0 {
		courseID = courseIDs[0]
	}
	if len(locationIDs) > 0 {
		locationID = locationIDs[0]
	}
	now := time.Now()
	stepState.RequestSentAt = now
	_, err := s.BobDB.Exec(ctx, query, classID, className, courseID, locationID, schoolID, now, now)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("%w", err)
	}
	stepState.FilterClassIDs = append(stepState.FilterClassIDs, classID)
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) aListStudentSubscriptionsAreExistedInDBWithEnrollmentStatus(ctx context.Context, total, schoolID int, enrollmentStatus, _startDate, _endDate string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.CommonSuite.ASignedInWithSchool(contextWithResourcePath(ctx, i32ToStr(schoolID)), "school admin", stepState.CurrentSchoolID)
	if err != nil {
		return ctx, err
	}
	s.clearPreviousTestStudentSubscriptions(ctx, schoolID)

	stepState.FilterCourseIDs = []string{idutil.ULIDNow(), idutil.ULIDNow()}
	stepState.FilterLocationIDs = stepState.LocationIDs

	for _, courseID := range stepState.FilterCourseIDs {
		courseName := "name-" + courseID
		sql := `insert into courses
		(course_id, name, grade, created_at, updated_at, school_id)
		VALUES ($1, $2, $3, $4, $5, $6)`
		_, err := s.BobDB.Exec(ctx, sql, database.Text(courseID), courseName, 12, time.Now(), time.Now(), schoolID)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("cannot init course err: %s", err)
		}
	}

	// init class
	ctx, err = s.insertClass(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("cannot init class err: %s", err)
	}

	startDate, _ := time.Parse(time.RFC3339, _startDate)
	endDate, _ := time.Parse(time.RFC3339, _endDate)

	// create students and student subscriptions
	filterStudentSubs := []string{}
	num := 0
	for i := 1; i <= (total / 2); i++ {
		studentId := idutil.ULIDNow()
		studentName := "student name -" + studentId
		if ctx, err := s.createStudentWithNameAndEnrollmentStatus(ctx, studentId, studentName, enrollmentStatus); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		for _, courseID := range stepState.FilterCourseIDs {
			num++
			studentSubscriptionId, subscriptionId := idutil.ULIDNow(), idutil.ULIDNow()
			filterStudentSubs = append(filterStudentSubs, studentSubscriptionId)
			duration := time.Duration(20 - num)
			createdAt := timeutil.Now().Add(duration * time.Hour)
			sql := `insert into lesson_student_subscriptions
			(student_subscription_id, course_id, student_id, subscription_id, start_at, end_at, created_at, updated_at, resource_path)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
			_, err := s.BobDB.Exec(ctx, sql, database.Text(studentSubscriptionId), database.Text(courseID), studentId, subscriptionId, startDate, endDate, createdAt, time.Now(), database.Text(fmt.Sprint(schoolID)))
			if err != nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf("cannot init course err: %s", err)
			}

			sql = `INSERT INTO lesson_student_subscription_access_path
			(student_subscription_id, location_id)
			VALUES($1, $2);`

			_, err = s.BobDB.Exec(ctx, sql, database.Text(studentSubscriptionId), stepState.FilterLocationIDs[0])
			if err != nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf("cannot init lesson_student_subscription_access_path err: %s", err)
			}
		}
		sql := `INSERT INTO class_member
	(class_member_id, class_id, user_id, created_at, updated_at)
	VALUES($1, $2, $3, timezone('utc'::text, now()), timezone('utc'::text, now())) ON CONFLICT DO NOTHING;`

		_, err := s.BobDB.Exec(ctx, sql, database.Text(idutil.ULIDNow()), stepState.FilterClassIDs[0], studentId)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("cannot init class_member err: %s", err)
		}
	}

	stepState.FilterStudentSubs = filterStudentSubs
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) createStudentWithNameAndEnrollmentStatus(ctx context.Context, id, name, enrollmentStatus string) (context.Context, error) {
	num := idutil.ULIDNow()
	stepState := StepStateFromContext(ctx)

	if stepState.CurrentSchoolID == 0 {
		stepState.CurrentSchoolID = 1
	}

	if name == "" {
		name = fmt.Sprintf("valid-student-%s", num)
	}

	student := &entities.Student{}
	database.AllNullEntity(student)
	database.AllNullEntity(&student.User)
	database.AllNullEntity(&student.User.AppleUser)

	err := multierr.Combine(
		student.ID.Set(id),
		student.LastName.Set(name),
		student.Country.Set(pb.COUNTRY_VN.String()),
		student.PhoneNumber.Set(fmt.Sprintf("phone-number+%s", num)),
		student.Email.Set(fmt.Sprintf("email+%s@example.com", num)),
		student.CurrentGrade.Set(12),
		student.TargetUniversity.Set("TG11DT"),
		student.TotalQuestionLimit.Set(5),
		student.SchoolID.Set(stepState.CurrentSchoolID),
		student.EnrollmentStatus.Set(enrollmentStatus),
	)

	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	locationRepo := &location_repo.LocationRepo{}
	locationOrg, err := locationRepo.GetLocationOrg(ctx, s.BobPostgresDB, fmt.Sprint(stepState.CurrentSchoolID))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("locationRepo.GetLocationOrg: %v", err)
	}

	err = database.ExecInTx(ctx, s.BobDB, func(ctx context.Context, tx pgx.Tx) error {
		if err := s.CommonSuite.AddUserAccessPath(ctx, id, locationOrg.LocationID, int64(stepState.CurrentSchoolID), tx); err != nil {
			return errors.Wrap(err, "s.AddUserAccessPath")
		}
		if err := (&repositories.StudentRepo{}).Create(ctx, tx, student); err != nil {
			return errors.Wrap(err, "s.StudentRepo.CreateTx")
		}

		if student.AppleUser.ID.String != "" {
			if err := (&repositories.AppleUserRepo{}).Create(ctx, tx, &student.AppleUser); err != nil {
				return errors.Wrap(err, "s.AppleUserRepo.Create")
			}
		}
		return nil
	})

	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}
func (s *Suite) clearPreviousTestStudentSubscriptions(ctx context.Context, schoolID int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	query := `update lesson_student_subscriptions set deleted_at = now()
	where course_id in(
	select l.course_id 
		from lesson_student_subscriptions l
		join courses c on l.course_id = c.course_id
	where c.school_id = $1)`
	_, err := s.BobDB.Exec(ctx, query, schoolID)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("cannot init clear lesson_student_subscriptions err: %s", err)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) adminRetrieveStudentSubscriptions(ctx context.Context, limit, offset int, _lessonDate, keyword, coursers, grades, classIds, locationIds string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	filter := &bpb.RetrieveStudentSubscriptionFilter{}
	keywordStr := ""
	if keyword != "" {
		keywordStr = keyword
	}
	if coursers != "" && len(strings.Split(coursers, ",")) > 0 {
		filter.CourseId = stepState.FilterCourseIDs[0:len(strings.Split(coursers, ","))]
	}
	if len(grades) > 0 {
		filter.Grade = strings.Split(grades, ",")
	}
	if classIds != "" && len(strings.Split(classIds, ",")) > 0 {
		for _, v := range strings.Split(classIds, ",") {
			index, err := strconv.Atoi(v)
			if err != nil {
				return StepStateToContext(ctx, stepState), err
			}
			filter.ClassId = append(filter.ClassId, stepState.FilterClassIDs[index-1])
		}
	}
	if locationIds != "" && len(strings.Split(locationIds, ",")) > 0 {
		for _, v := range strings.Split(locationIds, ",") {
			index, err := strconv.Atoi(v)
			if err != nil {
				return StepStateToContext(ctx, stepState), err
			}
			filter.LocationId = append(filter.LocationId, stepState.FilterLocationIDs[index-1])
		}
	}
	lessonDate, _ := time.Parse(time.RFC3339, _lessonDate)
	req := &bpb.RetrieveStudentSubscriptionRequest{
		Paging: &cpb.Paging{
			Limit: uint32(limit),
		},
		Keyword:    keywordStr,
		Filter:     filter,
		LessonDate: timestamppb.New(lessonDate),
	}

	if offset > 0 {
		offset := stepState.FilterStudentSubs[offset]
		req = &bpb.RetrieveStudentSubscriptionRequest{
			Paging: &cpb.Paging{
				Limit:  uint32(limit),
				Offset: &cpb.Paging_OffsetString{OffsetString: offset},
			},
			Keyword: keywordStr,
			Filter:  filter,
		}
	}

	stepState.Request = req
	stepState.Response, stepState.ResponseErr = bpb.NewStudentSubscriptionServiceClient(s.BobConn).RetrieveStudentSubscription(s.CommonSuite.SignedCtx(ctx), req)

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) aListStudentSubscriptionAccessPathExistedInDB(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	StudentSubscriptionAccessPathRepo := repositories.StudentSubscriptionAccessPathRepo{}
	ss := []*entities.StudentSubscriptionAccessPath{}
	for _, sub := range stepState.FilterStudentSubs {
		for _, location := range stepState.CenterIDs {
			item, err := toStudentSubscriptionAccessPathEntity(location, sub)
			if err != nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf("err toStudentSubscriptionAccessPathEntity")
			}
			ss = append(ss, item)
		}
	}
	if err := StudentSubscriptionAccessPathRepo.Upsert(ctx, s.BobDB, ss); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("StudentSubscriptionAccessPathRepo.Upsert")
	}
	return StepStateToContext(ctx, stepState), nil
}

func toStudentSubscriptionAccessPathEntity(locationID, studentSubscriptionID string) (*entities.StudentSubscriptionAccessPath, error) {
	cap := &entities.StudentSubscriptionAccessPath{}
	database.AllNullEntity(cap)
	err := multierr.Combine(
		cap.StudentSubscriptionID.Set(studentSubscriptionID),
		cap.LocationID.Set(locationID),
	)
	if err != nil {
		return nil, err
	}
	return cap, nil
}

func (s *Suite) bobMustReturnListStudentSubscriptions(ctx context.Context, total int, fromID, toID string, limit int, next, pre, lessonDate, hasFilter string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := stepState.Request.(*bpb.RetrieveStudentSubscriptionRequest)
	rsp := stepState.Response.(*bpb.RetrieveStudentSubscriptionResponse)
	if total == 0 {
		if rsp.TotalItems != uint32(total) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("must return empty list\n expect total: %d, actual total : %d", total, rsp.TotalItems)
		}
		return StepStateToContext(ctx, stepState), nil
	} else if rsp.TotalItems != uint32(total) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("wrong list return\n expect total: %d, actual: %d", total, rsp.TotalItems)
	}
	if len(rsp.Items) > 0 {
		ctx, err := s.checkClassIDInResponse(ctx, rsp.Items)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		ctx, err = s.checkRangeDateInResponse(ctx, rsp.Items, lessonDate)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}
	if hasFilter != "" {
		ctx, err := s.resultStudentsSubCorrectCourse(ctx, req.Filter.GetCourseId())
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("list result not match with courses: %s", err)
		}

		ctx, err = s.resultStudentsSubCorrectGrade(ctx, req.Filter.GetGrade())
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("list result not match with grade: %s", err)
		}

		ctx, err = s.resultStudentsSubCorrectLocationIDs(ctx, req.Filter.GetLocationId())
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("list result not match with locationIds: %s", err)
		}

		ctx, err = s.resultStudentsSubCorrectClassIDs(ctx, req.Filter.GetClassId())
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("list result not match with classIds: %s", err)
		}
	} else {
		if len(rsp.Items) > 0 {
			if fromID != NIL_VALUE {
				fromIDInt, err := strconv.Atoi(fromID)
				if err != nil {
					return StepStateToContext(ctx, stepState), err
				}

				if rsp.Items[0].Id != stepState.FilterStudentSubs[fromIDInt] {
					return StepStateToContext(ctx, stepState), fmt.Errorf("wrong first items return \n expect: %s, actual: %s", stepState.FilterStudentSubs[fromIDInt], rsp.Items[0].Id)
				}
			}
			if toID != NIL_VALUE {
				toIDInt, err := strconv.Atoi(toID)
				if err != nil {
					return StepStateToContext(ctx, stepState), err
				}
				if rsp.Items[len(rsp.Items)-1].Id != stepState.FilterStudentSubs[toIDInt] {
					return StepStateToContext(ctx, stepState), fmt.Errorf("wrong last items return \n expect: %s, actual: %s", stepState.FilterStudentSubs[toIDInt], rsp.Items[len(rsp.Items)-1].Id)
				}
			}
		}

		if int(rsp.NextPage.Limit) != limit || int(rsp.PreviousPage.Limit) != limit {
			return StepStateToContext(ctx, stepState), fmt.Errorf("wrong limit return")
		}
		if next != NIL_VALUE {
			nextInt, err := strconv.Atoi(next)
			if err != nil {
				return StepStateToContext(ctx, stepState), err
			}
			if rsp.NextPage.GetOffsetString() != stepState.FilterStudentSubs[nextInt] {
				return StepStateToContext(ctx, stepState), fmt.Errorf("wrong next offset return \n expect: %s, actual: %s", stepState.FilterStudentSubs[nextInt], rsp.NextPage.GetOffsetString())
			}
		}

		if pre == NIL_VALUE {
			pre = ""
		} else {
			preInt, err := strconv.Atoi(pre)
			if err != nil {
				return StepStateToContext(ctx, stepState), err
			}
			pre = stepState.FilterStudentSubs[preInt]
		}
		if rsp.PreviousPage.GetOffsetString() != pre {
			return StepStateToContext(ctx, stepState), fmt.Errorf("wrong previous offset return \n expect: %s, actual: %s", pre, rsp.PreviousPage.GetOffsetString())
		}
	}
	ctx, err := s.resultCorrectLocation(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("list result not match with access_path: %s", err)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) resultCorrectLocation(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	rsp := stepState.Response.(*bpb.RetrieveStudentSubscriptionResponse)

	for _, item := range rsp.GetItems() {
		for _, locationID := range item.LocationIds {
			if !checkStringExists(locationID, stepState.CenterIDs) {
				return StepStateToContext(ctx, stepState), fmt.Errorf("location_id %s not match", locationID)
			}
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) resultStudentsSubCorrectCourse(ctx context.Context, courses []string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if len(courses) == 0 {
		return StepStateToContext(ctx, stepState), nil
	}
	rsp := stepState.Response.(*bpb.RetrieveStudentSubscriptionResponse)

	for _, item := range rsp.GetItems() {
		if !checkStringExists(item.CourseId, courses) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("course %s not match with course filter", item.CourseId)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) resultStudentsSubCorrectGrade(ctx context.Context, grades []string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if len(grades) == 0 {
		return StepStateToContext(ctx, stepState), nil
	}
	rsp := stepState.Response.(*bpb.RetrieveStudentSubscriptionResponse)

	for _, item := range rsp.GetItems() {
		if !checkStringExists(item.Grade, grades) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("grade %s not match with grade filter", item.Grade)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) resultStudentsSubCorrectLocationIDs(ctx context.Context, locationIds []string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if len(locationIds) == 0 {
		return StepStateToContext(ctx, stepState), nil
	}

	rsp := stepState.Response.(*bpb.RetrieveStudentSubscriptionResponse)

	repo := repositories.StudentSubscriptionAccessPathRepo{}
	studentSubscriptionIDs, err := repo.FindStudentSubscriptionIDsByLocationIDs(ctx, s.BobDB, locationIds)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("cannot get ClassMemberRepo.FindStudentSubscriptionIDsByLocationIDs: %s", err)
	}

	for _, item := range rsp.GetItems() {
		if !checkStringExists(item.Id, studentSubscriptionIDs) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("student subscription %s in location %s not match with location filter", item.Id, locationIds)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) resultStudentsSubCorrectClassIDs(ctx context.Context, classIds []string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if len(classIds) == 0 {
		return StepStateToContext(ctx, stepState), nil
	}

	rsp := stepState.Response.(*bpb.RetrieveStudentSubscriptionResponse)

	repo := master_repo.ClassMemberRepo{}
	students, err := repo.FindStudentIDWithCourseIDsByClassIDs(ctx, s.BobDB, classIds)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("cannot get ClassMemberRepo.FindStudentIDWithCourseIDsByClassIDs: %s", err)
	}

	for _, item := range rsp.GetItems() {
		if !checkStringExists(item.StudentId, students) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("class %s not match with class filter", item.StudentId)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) checkClassIDInResponse(ctx context.Context, list []*bpb.RetrieveStudentSubscriptionResponse_StudentSubscription) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	studentCourses := make([]*domain.ClassWithCourseStudent, 0, len(list))
	for _, sub := range list {
		sc := &domain.ClassWithCourseStudent{CourseID: sub.CourseId, StudentID: sub.StudentId}
		studentCourses = append(studentCourses, sc)
	}

	repo := master_repo.ClassRepo{}
	result, err := repo.FindByCourseIDsAndStudentIDs(ctx, s.BobDB, studentCourses)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("list result not match with courses: %s", err)
	}

	for _, v := range list {
		isIncorrect := s.checkClassIDIsIncorrectWithCourseAndClass(result, v)
		if isIncorrect && v.ClassId != "" {
			return StepStateToContext(ctx, stepState), fmt.Errorf("classId: %s in item %s", v.ClassId, v.Id)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) checkClassIDIsIncorrectWithCourseAndClass(list []*domain.ClassWithCourseStudent, subStudent *bpb.RetrieveStudentSubscriptionResponse_StudentSubscription) bool {
	for _, v := range list {
		if v.CourseID == subStudent.CourseId && v.StudentID == subStudent.StudentId {
			if v.ClassID == subStudent.ClassId {
				return false
			}
		}
	}

	return true
}

func checkStringExists(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func (s *Suite) checkRangeDateInResponse(ctx context.Context, list []*bpb.RetrieveStudentSubscriptionResponse_StudentSubscription, _lessonDate string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	lessonDate, _ := time.Parse(time.RFC3339, _lessonDate)
	d := time.Date(lessonDate.Year(), lessonDate.Month(), lessonDate.Day(), 0, 0, 0, 0, time.UTC)
	for _, sub := range list {
		startDate := time.Date(sub.StartDate.AsTime().Year(), sub.StartDate.AsTime().Month(), sub.StartDate.AsTime().Day(), 0, 0, 0, 0, time.UTC)
		endDate := time.Date(sub.EndDate.AsTime().Year(), sub.EndDate.AsTime().Month(), sub.EndDate.AsTime().Day(), 0, 0, 0, 0, time.UTC)

		if startDate.After(d) || endDate.Before(d) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("student subscription duration(student_id:%s,course_id:%s) not aligned with lesson date", sub.StudentId, sub.CourseId)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}
