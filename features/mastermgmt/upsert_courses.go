package mastermgmt

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	entities_bob "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/timeutil"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
)

const (
	NIL_VALUE string = "nil"
)

func getPbTeachingMethod(teachingMethodString string) mpb.CourseTeachingMethod {
	var teachingMethodPb mpb.CourseTeachingMethod
	switch teachingMethodString {
	case "group":
		{
			teachingMethodPb = mpb.CourseTeachingMethod_COURSE_TEACHING_METHOD_GROUP
			break
		}
	case "individual":
		{
			teachingMethodPb = mpb.CourseTeachingMethod_COURSE_TEACHING_METHOD_INDIVIDUAL
			break
		}
	case "invalid":
		{
			teachingMethodPb = mpb.CourseTeachingMethod_COURSE_TEACHING_METHOD_NONE
			break
		}
	}
	return teachingMethodPb
}
func (s *suite) getExamplePbCourses(ctx context.Context, location int, teachingMethod string) map[string]*mpb.UpsertCoursesRequest_Course {
	stepState := StepStateFromContext(ctx)
	teachingMethodPb := getPbTeachingMethod(teachingMethod)

	return map[string]*mpb.UpsertCoursesRequest_Course{
		"valid":       s.genPbCourse(ctx, "valid"+stepState.Random, "course 1", location, teachingMethodPb),
		"valid2":      s.genPbCourse(ctx, "valid2"+stepState.Random, "course 2", location, teachingMethodPb),
		"":            s.genPbCourse(ctx, "", "course-invalid-id", location, teachingMethodPb),
		"invalidName": s.genPbCourse(ctx, "invalidName"+stepState.Random, "course-invalid-name", location, teachingMethodPb),
	}
}

func (s *suite) genPbCourse(ctx context.Context, courseID, name string, location int, teachingMethod mpb.CourseTeachingMethod) *mpb.UpsertCoursesRequest_Course {
	stepState := StepStateFromContext(ctx)

	schoolID := stepState.CurrentSchoolID
	r := &mpb.UpsertCoursesRequest_Course{
		Id:             courseID,
		Name:           name,
		SchoolId:       schoolID,
		Icon:           "link-icon",
		LocationIds:    stepState.CenterIDs[0:location],
		TeachingMethod: teachingMethod,
		CourseType:     stepState.CourseTypeIDs[0],
	}
	return r
}

func (s *suite) getCourseBasedOnState(ctx context.Context, courseID string, location int, teachingMethod string) (e *mpb.UpsertCoursesRequest_Course) {
	courses := s.getExamplePbCourses(ctx, location, teachingMethod)

	course, ok := courses[courseID]
	if !ok {
		course = s.genPbCourse(ctx, courseID, "specific courses courseID: "+courseID, location, mpb.CourseTeachingMethod_COURSE_TEACHING_METHOD_GROUP)
		courses[courseID] = course
	}

	return course
}

func (s *suite) userUpsertCoursesDataWithLocationsAndTeachingMethod(ctx context.Context, state string, location int, teachingMethod string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.FilterCourseIDs = make([]string, 0, 1)
	req, ok := stepState.Request.(*mpb.UpsertCoursesRequest)
	if !ok {
		req = &mpb.UpsertCoursesRequest{
			Courses: []*mpb.UpsertCoursesRequest_Course{},
		}
	}

	course := s.getCourseBasedOnState(ctx, state, location, teachingMethod)
	req.Courses = append(req.Courses, course)
	stepState.FilterCourseIDs = append(stepState.FilterCourseIDs, course.Id)
	stepState.Request = req

	stepState.Response, stepState.ResponseErr = mpb.NewMasterDataCourseServiceClient(s.MasterMgmtConn).UpsertCourses(contextWithToken(s, ctx), stepState.Request.(*mpb.UpsertCoursesRequest))
	stepState.CurrentCourseID = course.Id
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) returnAErrorMessage(ctx context.Context, errorMessage string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if errorMessage == NIL_VALUE {
		if stepState.ResponseErr != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expect nil error, but got %v", stepState.ResponseErr)
		}
	} else {
		if stepState.ResponseErr == nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expect %v error, but got nil", errorMessage)
		}
		if !strings.Contains(stepState.ResponseErr.Error(), errorMessage) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expect %v error, but got %v", errorMessage, stepState.ResponseErr)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func checkLocationsOfCourse(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func (s *suite) courseAccessPathsExistInDB(ctx context.Context, location int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := stepState.Request.(*mpb.UpsertCoursesRequest)
	courseAccessPath := repositories.CourseAccessPathRepo{}
	for _, course := range req.Courses {
		if course.Id != "" {
			mapLocationIDsByCourseID, err := courseAccessPath.FindByCourseIDs(ctx, s.BobDBTrace, []string{course.Id})
			if err != nil {
				return ctx, errors.New("courseAccessPath.FindByCourseIDs")
			}
			found := false
			for _, e := range stepState.CenterIDs[0:location] {
				found = checkLocationsOfCourse(mapLocationIDsByCourseID[course.Id], e)
			}
			if !found {
				return StepStateToContext(ctx, stepState), fmt.Errorf("failed to insert course access path row")
			}
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aRandomNumber(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.Random = strconv.Itoa(rand.Int())
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aGenerateSchool(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	random := strconv.Itoa(rand.Int())
	stepState.Random = random

	school := &entities_bob.School{
		Name:           database.Text(fmt.Sprintf("school-%s", random)),
		Country:        database.Text(pb.COUNTRY_VN.String()),
		CityID:         pgtype.Int4{Int: 63, Status: pgtype.Present},
		DistrictID:     pgtype.Int4{Int: 703, Status: pgtype.Present},
		IsSystemSchool: pgtype.Bool{Bool: false, Status: pgtype.Present},
		Point: pgtype.Point{
			P:      pgtype.Vec2{X: 11.2, Y: 10.4},
			Status: pgtype.Present,
		},
	}

	schoolRepo := repositories.SchoolRepo{}
	err := schoolRepo.Create(ctx, s.BobDBTrace, school)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	id := strconv.Itoa(int(school.ID.Int))
	ctx, err = s.generateOrganizationAuth(ctx, id)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("cannot generate organization auth: %v", err)
	}

	stepState.CurrentSchoolID = school.ID.Int
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) locationRemovedHaveState(ctx context.Context, state string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var updated string
	switch state {
	case "deleted":
		updated = "deleted_at = now()"
	case "active":
		updated = "deleted_at = null, is_archived = false"
	case "archived":
		updated = "is_archived = true"
	default:
		return StepStateToContext(ctx, stepState), fmt.Errorf("state is not available")
	}
	query := fmt.Sprintf("update locations set %s where location_id = %s::text", updated, fmt.Sprintf(stepState.CenterIDs[1]))
	return s.runBobQuery(ctx, query)
}

func (s *suite) runBobQuery(ctx context.Context, query string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	_, err := s.BobDBTrace.Exec(ctx, query)
	if err != nil {
		return ctx, fmt.Errorf("cannot run BobQuery: %v", err)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) generateOrganizationAuth(ctx context.Context, schoolID string) (context.Context, error) {
	stmt :=
		`
	INSERT INTO organization_auths
		(organization_id, auth_project_id, auth_tenant_id)
	VALUES
		($1, 'fake_aud', '')
	ON CONFLICT 
		DO NOTHING
	;
	`
	_, err := s.BobDBTrace.Exec(ctx, stmt, schoolID)
	if err != nil {
		return ctx, fmt.Errorf("cannot insert organization auth: %v", err)
	}
	return ctx, nil
}

func (s *suite) someCenters(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.CurrentSchoolID = constants.ManabieSchool
	ctx, err := s.signedAsAccountV2(ctx, "school admin")
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	ctx, err = s.aListOfLocationsInDB(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	return StepStateToContext(ctx, stepState), nil
}

// func (s *suite) aSignedInWithSchool(ctx context.Context, role string, schoolID int) (context.Context, error) {
// 	stepState := StepStateFromContext(ctx)
// 	stepState.CurrentSchoolID = int32(schoolID)
// 	switch role {
// 	case "school admin":
// 		stepState.CurrentSchoolID = int32(schoolID)
// 		return s.aSignedInSchoolAdminWithSchoolID(ctx,
// 			entities_bob.UserGroupSchoolAdmin, schoolID)
// 	case "student":
// 		{
// 			if ctx, err := s.aSignedInStudent(ctx); err != nil {
// 				return StepStateToContext(ctx, stepState), err
// 			}

// 			t, _ := jwt.ParseString(stepState.AuthToken)
// 			return s.aValidStudentWithSchoolID(ctx, t.Subject(), schoolID)
// 		}
// 	case "teacher":
// 		stepState.CurrentSchoolID = int32(schoolID)
// 		return s.aSignedInTeacherWithSchoolID(ctx, schoolID)
// 	case "admin":
// 		return s.aSignedInAdmin(ctx)
// 	case "unauthenticated":
// 		stepState.AuthToken = "random-token"
// 		return StepStateToContext(ctx, stepState), nil
// 	}

// 	return StepStateToContext(ctx, stepState), nil
// }

func (s *suite) aSignedInSchoolAdminWithSchoolID(ctx context.Context, group string, schoolID int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	id := s.newID()
	ctx, err := s.aValidSchoolAdminProfileWithId(ctx, id, group, schoolID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.AuthToken, err = s.generateExchangeToken(id, group)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.CurrentUserID = id
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aValidSchoolAdminProfileWithId(ctx context.Context, id, userGroup string, schoolID int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	c := entities_bob.SchoolAdmin{}
	database.AllNullEntity(&c)

	c.SchoolAdminID.Set(id)
	c.SchoolID.Set(schoolID)
	now := time.Now()
	if err := c.UpdatedAt.Set(now); err != nil {
		return nil, err
	}
	if err := c.CreatedAt.Set(now); err != nil {
		return nil, err
	}

	num := rand.Int()
	u := entities_bob.User{}
	database.AllNullEntity(&u)

	u.ID = c.SchoolAdminID
	u.LastName.Set(fmt.Sprintf("valid-school-admin-%d", num))
	u.PhoneNumber.Set(fmt.Sprintf("+848%d", num))
	u.Email.Set(fmt.Sprintf("valid-school-admin-%d@email.com", num))
	u.Avatar.Set(fmt.Sprintf("http://valid-school-admin-%d", num))
	u.Country.Set(pb.COUNTRY_VN.String())
	if userGroup == "" {
		userGroup = entities_bob.UserGroupSchoolAdmin
	}
	u.Group.Set(userGroup)
	u.DeviceToken.Set(nil)
	u.AllowNotification.Set(true)
	u.CreatedAt = c.CreatedAt
	u.UpdatedAt = c.UpdatedAt
	u.IsTester.Set(nil)
	u.FacebookID.Set(nil)

	userRepo := repositories.UserRepo{}

	err := userRepo.Create(ctx, s.BobDBTrace, &u)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	schoolAdminRepo := repositories.SchoolAdminRepo{}
	err = schoolAdminRepo.CreateMultiple(ctx, s.BobDBTrace, []*entities_bob.SchoolAdmin{&c})

	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	ug := entities_bob.UserGroup{}
	database.AllNullEntity(&ug)

	ug.UserID.Set(id)
	ug.GroupID.Set(userGroup)
	ug.UpdatedAt.Set(now)
	ug.CreatedAt.Set(now)
	ug.IsOrigin.Set(true)
	ug.Status.Set(entities_bob.UserGroupStatusActive)

	userGroupRepo := repositories.UserGroupRepo{}
	err = userGroupRepo.Upsert(ctx, s.BobDBTrace, &ug)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) someCourseTypes(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	randNum := strconv.Itoa(rand.Int())
	listCourseType := []struct {
		ID   string
		name string
	}{
		{ID: "1" + randNum, name: "name 1" + randNum},
		{ID: "2" + randNum, name: "name 2" + randNum},
	}
	for _, c := range listCourseType {
		stmt := `INSERT INTO course_type (course_type_id,name,created_at,updated_at) VALUES($1,$2,now(),now()) 
		ON CONFLICT ON CONSTRAINT course_type__pk DO UPDATE SET deleted_at = null`
		_, err := s.BobDBTrace.Exec(ctx, stmt, c.ID,
			c.name)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("cannot insert course type with `id:%s`, %v", c.ID, err)
		}
		stepState.CourseTypeIDs = append(stepState.CourseTypeIDs, c.ID)
	}
	return StepStateToContext(ctx, stepState), nil
}

func buildAccessPath(rootLocation, rand string, locationPrefixes []string) string {
	rs := rootLocation
	for _, str := range locationPrefixes {
		rs += "/" + str + rand
	}
	return rs
}

func (s *suite) aListOfLocationsInDB(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	randNum := strconv.Itoa(rand.Int())
	listLocation := []struct {
		locationID       string
		name             string
		parentLocationID string
		archived         bool
		expected         bool
		accessPath       string
	}{ // satisfied
		{locationID: "1" + randNum, parentLocationID: s.LocationID, archived: false, expected: true, accessPath: buildAccessPath(s.LocationID, randNum, []string{"1"})},
		{locationID: "2" + randNum, parentLocationID: "1" + randNum, archived: false, expected: true, accessPath: buildAccessPath(s.LocationID, randNum, []string{"1", "2"})},
		{locationID: "3" + randNum, parentLocationID: "2" + randNum, archived: false, expected: true, accessPath: buildAccessPath(s.LocationID, randNum, []string{"1", "2", "3"})},
		{locationID: "7" + randNum, archived: false, expected: true, accessPath: buildAccessPath(s.LocationID, randNum, []string{})},
		// unsatisfied
		{locationID: "4" + randNum, parentLocationID: s.LocationID, archived: true, accessPath: buildAccessPath(s.LocationID, randNum, []string{"4"})},
		{locationID: "5" + randNum, parentLocationID: "4" + randNum, archived: false, expected: false, accessPath: buildAccessPath(s.LocationID, randNum, []string{"4", "5"})},
		{locationID: "6" + randNum, parentLocationID: "5" + randNum, archived: false, expected: false, accessPath: buildAccessPath(s.LocationID, randNum, []string{"4", "5", "6"})},
		{locationID: "8" + randNum, parentLocationID: "7" + randNum, archived: true, expected: false, accessPath: buildAccessPath(s.LocationID, randNum, []string{"7", "8"})},
	}
	for _, l := range listLocation {
		stmt := `INSERT INTO locations (location_id,name,parent_location_id, is_archived, access_path) VALUES($1,$2,$3,$4,$5) 
				ON CONFLICT DO NOTHING`
		_, err := s.BobDBTrace.Exec(ctx, stmt, l.locationID,
			l.name,
			NewNullString(l.parentLocationID),
			l.archived,
			l.accessPath,
		)

		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("cannot insert locations with `id:%s`, %v", l.locationID, err)
		}

		_, err = s.BobDBTrace.Exec(ctx, "update locations set deleted_at = null where location_id = $1", l.locationID)

		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("cannot update locations with `id:%s`, %v", l.locationID, err)
		}

		if l.expected {
			stepState.CenterIDs = append(stepState.CenterIDs, l.locationID)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func NewNullString(s string) sql.NullString {
	if len(s) == 0 {
		return sql.NullString{}
	}
	return sql.NullString{
		String: s,
		Valid:  true,
	}
}

func (s *suite) aValidStudentWithSchoolID(ctx context.Context, id string, schoolID int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	sql := "UPDATE students SET school_id = $1 WHERE student_id = $2"
	_, err := s.BobDBTrace.Exec(ctx, sql, &schoolID, &id)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aSignedInTeacherWithSchoolID(ctx context.Context, schoolID int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	id := s.newID()
	ctx, err := s.aValidTeacherProfileWithId(ctx, id, int32(schoolID))
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.CurrentTeacherID = id
	stepState.CurrentUserID = id

	stepState.AuthToken, err = s.generateExchangeToken(id, entities_bob.UserGroupTeacher)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

//nolint:errcheck
func (s *suite) aValidTeacherProfileWithId(ctx context.Context, id string, schoolID int32) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	c := entities_bob.Teacher{}
	database.AllNullEntity(&c.User)
	database.AllNullEntity(&c)
	c.ID.Set(id)
	var schoolIDs []int32
	if len(stepState.Schools) > 0 {
		schoolIDs = []int32{stepState.Schools[0].ID.Int}
	}
	if schoolID != 0 {
		schoolIDs = append(schoolIDs, schoolID)
	}
	c.SchoolIDs.Set(schoolIDs)
	now := time.Now()
	if err := c.UpdatedAt.Set(now); err != nil {
		return nil, err
	}
	if err := c.CreatedAt.Set(now); err != nil {
		return nil, err
	}
	num := rand.Int()
	u := entities_bob.User{}
	database.AllNullEntity(&u)
	u.ID = c.ID
	u.LastName.Set(fmt.Sprintf("valid-teacher-%d", num))
	u.PhoneNumber.Set(fmt.Sprintf("+848%d", num))
	u.Email.Set(fmt.Sprintf("valid-teacher-%d@email.com", num))
	u.Avatar.Set(fmt.Sprintf("http://valid-teacher-%d", num))
	u.Country.Set(pb.COUNTRY_VN.String())
	u.Group.Set(entities_bob.UserGroupTeacher)
	u.DeviceToken.Set(nil)
	u.AllowNotification.Set(true)
	u.CreatedAt = c.CreatedAt
	u.UpdatedAt = c.UpdatedAt
	u.IsTester.Set(nil)
	u.FacebookID.Set(nil)
	uG := entities_bob.UserGroup{UserID: c.ID, GroupID: database.Text(pb.USER_GROUP_TEACHER.String()), IsOrigin: database.Bool(true)}
	uG.Status.Set("USER_GROUP_STATUS_ACTIVE")
	uG.CreatedAt = u.CreatedAt
	uG.UpdatedAt = u.UpdatedAt
	_, err := database.InsertExcept(ctx, &u, []string{"resource_path"}, s.BobDBTrace.Exec)
	if err != nil {
		return ctx, err
	}
	_, err = database.InsertExcept(ctx, &c, []string{"resource_path"}, s.BobDBTrace.Exec)
	if err != nil {
		return ctx, err
	}
	cmdTag, err := database.InsertExcept(ctx, &uG, []string{"resource_path"}, s.BobDBTrace.Exec)
	if err != nil {
		return ctx, err
	}
	if cmdTag.RowsAffected() == 0 {
		return ctx, errors.New("cannot insert teacher for testing")
	}
	return ctx, nil
}

type fileJson struct {
	Data []string `json:"data"`
}

type presignedUrlUploadContext struct {
	FileJson    fileJson `json:"file_json"`
	StatusCode  int      `json:"status_code"`
	DownloadUrl string   `json:"download_url"`
}

func (s *suite) returnAStatusCode(ctx context.Context, msg string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	const successfulCode = "2xx"
	const clientErrorCode = "4xx"
	res := stepState.Response.(*presignedUrlUploadContext)

	var statusCode string
	if res.StatusCode >= 200 && res.StatusCode < 300 {
		statusCode = successfulCode
	} else if res.StatusCode >= 400 && res.StatusCode < 500 {
		statusCode = clientErrorCode
	}

	if msg != statusCode {
		return StepStateToContext(ctx, stepState), fmt.Errorf("status Code expected %v but got %v", msg, res.StatusCode)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) CreateStudentAccounts(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.StudentIds = []string{s.newID(), s.newID()}

	for _, id := range stepState.StudentIds {
		if ctx, err := s.createStudentWithName(ctx, id, ""); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}

	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) createStudentWithName(ctx context.Context, id, name string) (context.Context, error) {
	num := s.newID()
	stepState := StepStateFromContext(ctx)

	if stepState.CurrentSchoolID == 0 {
		stepState.CurrentSchoolID = constants.ManabieSchool
	}

	if name == "" {
		name = fmt.Sprintf("valid-student-%s", num)
	}

	student := &entities_bob.Student{}
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
	)

	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	err = database.ExecInTx(ctx, s.BobPostgresDBTrace, func(ctx context.Context, tx pgx.Tx) error {
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

func (s *suite) someStudentSubscriptionExistedInDB(ctx context.Context, status string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	filterStudentSubs := make([]string, 0, len(stepState.FilterCourseIDs))
	for _, courseID := range stepState.FilterCourseIDs {
		studentSubscriptionId, subscriptionId := s.newID(), s.newID()
		filterStudentSubs = append(filterStudentSubs, studentSubscriptionId)
		var startAt, endAt time.Time
		switch status {
		case "active":
			startAt = timeutil.Now().Add(-time.Hour)
			endAt = timeutil.Now().Add(time.Hour)
		case "future":
			startAt = timeutil.Now().Add(time.Hour)
			endAt = timeutil.Now().Add(2 * time.Hour)
		case "past":
			startAt = timeutil.Now().Add(-2 * time.Hour)
			endAt = timeutil.Now().Add(-1 * time.Hour)
		default:
			return StepStateToContext(ctx, stepState), fmt.Errorf("status is not available")
		}
		SQL := `insert into lesson_student_subscriptions
		(student_subscription_id, course_id, student_id, subscription_id, start_at, end_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
		_, err := s.BobDBTrace.Exec(ctx, SQL, database.Text(studentSubscriptionId), database.Text(courseID), stepState.StudentIds[0], subscriptionId, startAt, endAt, time.Now(), time.Now())
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("cannot init lesson_student_subscriptions err: %s", err)
		}
	}
	stepState.FilterStudentSubs = filterStudentSubs
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aListStudentSubscriptionAccessPathExistedInDB(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	StudentSubscriptionAccessPathRepo := repositories.StudentSubscriptionAccessPathRepo{}
	ss := make([]*entities_bob.StudentSubscriptionAccessPath, 0, len(stepState.FilterStudentSubs))
	for _, sub := range stepState.FilterStudentSubs {
		for _, location := range stepState.CenterIDs[0:2] {
			item, err := toStudentSubscriptionAccessPathEntity(location, sub)
			if err != nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf("err toStudentSubscriptionAccessPathEntity")
			}
			ss = append(ss, item)
		}
	}
	if err := StudentSubscriptionAccessPathRepo.Upsert(ctx, s.BobDBTrace, ss); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("StudentSubscriptionAccessPathRepo.Upsert")
	}
	return StepStateToContext(ctx, stepState), nil
}

func toStudentSubscriptionAccessPathEntity(locationID, studentSubscriptionID string) (*entities_bob.StudentSubscriptionAccessPath, error) {
	cap := &entities_bob.StudentSubscriptionAccessPath{}
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

func (s *suite) checkCourseInfoInDB(ctx context.Context, expectedTeachingMethod string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	expectedTeachingMethodPb := getPbTeachingMethod(expectedTeachingMethod)
	query :=
		`
		SELECT teaching_method, course_type_id, end_date from courses c
		WHERE c.course_id = $1;
	`
	var (
		actualTeachingMethod pgtype.Text
		actualCourseType     pgtype.Text
		actualEnddate        pgtype.Timestamptz
	)
	err := s.BobDBTrace.QueryRow(ctx, query, database.Text(stepState.CurrentCourseID)).Scan(&actualTeachingMethod, &actualCourseType, &actualEnddate)
	if err != nil {
		return ctx, fmt.Errorf("cannot scan course table: %v", err)
	}
	if expectedTeachingMethodPb.String() != actualTeachingMethod.String {
		if expectedTeachingMethodPb.String() == "" && actualTeachingMethod.String != "" {
			return StepStateToContext(ctx, stepState), fmt.Errorf("update teaching method failed for course with id %s. Expected: %s, actual: %s", stepState.CurrentCourseID, expectedTeachingMethodPb.String(), actualTeachingMethod.String)
		}
	}
	if actualCourseType.String != stepState.CourseTypeIDs[0] {
		return StepStateToContext(ctx, stepState), fmt.Errorf("update course type failed for course with id %s. Expected: %s, actual: %s", stepState.CurrentCourseID, stepState.CourseTypeIDs[0], actualCourseType.String)
	}
	// HACK: end_date for courses to adapt with join lesson logic
	var courseEndData = time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	if actualEnddate.Time.Unix() != courseEndData.Unix() {
		return StepStateToContext(ctx, stepState), fmt.Errorf("upsert course end_date failed for course with id %s. Expected: %s, actual: %s", stepState.CurrentCourseID, courseEndData.Format(time.RFC822), actualEnddate.Time.Format(time.RFC822))
	}
	return StepStateToContext(ctx, stepState), nil
}
