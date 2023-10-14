package bob

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/manabie-com/backend/internal/bob/constants"
	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"

	"github.com/lestrrat-go/jwx/jwt"
)

func (s *suite) aAddClassMemberRequest(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.Request = &pb.AddClassMemberRequest{
		ClassId:    0,
		TeacherIds: []string{},
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userAddAClassMember(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.createClassUpsertedSubscribe(StepStateToContext(ctx, stepState))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.createClassUpsertedSubscribe: %w", err)
	}
	stepState.Response, stepState.ResponseErr = pb.NewClassClient(s.Conn).AddClassMember(contextWithToken(s, ctx), stepState.Request.(*pb.AddClassMemberRequest))
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aUserIDOfSchoolIdIsInAddClassMemberRequest(ctx context.Context, userId string, schoolID int) (context.Context, error) {
	if userId == "" {
		return ctx, nil
	}

	stepState := StepStateFromContext(ctx)
	var err error
	authToken := stepState.AuthToken
	currentUserID := stepState.CurrentUserID

	if userId == "studentId" {
		ctx, err = s.aSignedInWithSchool(ctx, "student", schoolID)
	}

	if userId == "teacherId" {
		ctx, err = s.aSignedInWithSchool(ctx, "teacher", schoolID)
	}
	if err != nil {
		return ctx, err
	}

	t, _ := jwt.ParseString(stepState.AuthToken)
	req := stepState.Request.(*pb.AddClassMemberRequest)
	req.TeacherIds = append(req.TeacherIds, t.Subject())

	stepState.CurrentUserID = currentUserID
	stepState.AuthToken = authToken

	ctx = StepStateToContext(ctx, stepState)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aClassCodeInAddClassMemberRequest(ctx context.Context, arg string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if arg == "valid" {
		stepState.Request.(*pb.AddClassMemberRequest).ClassId = stepState.CurrentClassID
	}
	if arg == "wrong" {
		stepState.Request.(*pb.AddClassMemberRequest).ClassId = 0
	}

	t, _ := jwt.ParseString(stepState.AuthToken)
	stepState.CurrentStudentID = t.Subject()

	ctx = StepStateToContext(ctx, stepState)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aUserIDOfSchoolNameInAddClassMemberRequest(ctx context.Context, role, schoolName string) (context.Context, error) {
	schoolID := s.getSchoolIDByName(ctx, schoolName)
	return s.aUserIDOfSchoolIdIsInAddClassMemberRequest(ctx, role, schoolID)
}

func (s *suite) getSchoolIDByName(ctx context.Context, name string) int {
	stepState := StepStateFromContext(ctx)
	var schoolID int
	n := name + stepState.Random
	for _, school := range stepState.Schools {
		if school.Name.String == n {
			schoolID = int(school.ID.Int)
			break
		}
	}
	if schoolID == 0 {
		return constants.ManabieSchool
	}
	return schoolID
}

func (s *suite) createAClassWithSchoolNameAndExpiredAt(ctx context.Context, schoolName, schoolExpiredDate string) (context.Context, error) {
	return s.createAClassWithSchoolIdIsAndExpiredAt(ctx, schoolExpiredDate)
}

func (s *suite) CreateAClassWithSchoolNameAndExpiredAt(ctx context.Context, schoolName, schoolExpiredDate string) (context.Context, error) {
	return s.createAClassWithSchoolNameAndExpiredAt(ctx, schoolName, schoolExpiredDate)
}

// func (s *suite) aSignedInStudent(ctx context.Context) (context.Context, error) {
// 	stepState := StepStateFromContext(ctx)

// 	id := s.newID()
// 	var err error
// 	ctx, err = s.aValidStudentInDB(ctx, id)
// 	if err != nil {
// 		return StepStateToContext(ctx, stepState), err
// 	}
// 	ctx, err = s.aValidUserInEureka(ctx, id, constant.RoleStudent, entities.UserGroupStudent)
// 	if err != nil {
// 		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to create user in eureka: %w", err)
// 	}
// 	stepState.AuthToken, err = s.generateExchangeToken(id, entities.UserGroupStudent)
// 	if err != nil {
// 		return StepStateToContext(ctx, stepState), err
// 	}
// 	stepState.CurrentStudentID = id

// 	return StepStateToContext(ctx, stepState), nil

// }

func (s *suite) aSignedInStudentWithSchool(ctx context.Context, schoolID int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.CurrentStudentID = ""
	id := s.newID()
	ctx, err := s.aValidStudentInDB(ctx, id)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	ctx, err = s.aValidStudentWithSchoolID(ctx, id, schoolID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.AuthToken, err = s.generateExchangeToken(id, entities.UserGroupStudent)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) ASignedInStudent(ctx context.Context) (context.Context, error) {
	return s.aSignedInStudent(ctx)
}
func (s *suite) aSignedInStudentGivenName(ctx context.Context, name string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.CurrentStudentID = ""
	id := s.newID()
	var err error
	ctx, err = s.aValidStudentWithName(ctx, id, name)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.AuthToken, err = s.generateExchangeToken(id, entities.UserGroupStudent)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	return StepStateToContext(ctx, stepState), nil
}

//nolint:errcheck
func (s *suite) aValidTeacherProfileWithId(ctx context.Context, id string, schoolID int32) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	c := entities.Teacher{}
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
	u := entities.User{}
	database.AllNullEntity(&u)
	u.ID = c.ID
	u.LastName.Set(fmt.Sprintf("valid-teacher-%d", num))
	u.PhoneNumber.Set(fmt.Sprintf("+848%d", num))
	u.Email.Set(fmt.Sprintf("valid-teacher-%d@email.com", num))
	u.Avatar.Set(fmt.Sprintf("http://valid-teacher-%d", num))
	u.Country.Set(pb.COUNTRY_VN.String())
	u.Group.Set(entities.UserGroupTeacher)
	u.DeviceToken.Set(nil)
	u.AllowNotification.Set(true)
	u.CreatedAt = c.CreatedAt
	u.UpdatedAt = c.UpdatedAt
	u.IsTester.Set(nil)
	u.FacebookID.Set(nil)
	uG := entities.UserGroup{UserID: c.ID, GroupID: database.Text(pb.USER_GROUP_TEACHER.String()), IsOrigin: database.Bool(true)}
	uG.Status.Set("USER_GROUP_STATUS_ACTIVE")
	uG.CreatedAt = u.CreatedAt
	uG.UpdatedAt = u.UpdatedAt
	staff := entity.Staff{}
	staff.ID = c.ID
	staff.UpdatedAt = u.UpdatedAt
	staff.CreatedAt = u.CreatedAt
	staff.DeletedAt.Set(nil)
	staff.StartDate.Set(nil)
	staff.EndDate.Set(nil)
	staff.AutoCreateTimesheet.Set(false)
	staff.WorkingStatus.Set("AVAILABLE")
	_, err := database.InsertExcept(ctx, &u, []string{"resource_path"}, s.DB.Exec)
	if err != nil {
		return ctx, fmt.Errorf("insert user error %w", err)
	}
	_, err = database.InsertExcept(ctx, &c, []string{"resource_path"}, s.DB.Exec)
	if err != nil {
		return ctx, fmt.Errorf("insert teacher error %w", err)
	}
	_, err = database.InsertExcept(ctx, &staff, []string{"resource_path"}, s.DB.Exec)
	if err != nil {
		return ctx, fmt.Errorf("insert staff error %w", err)
	}
	cmdTag, err := database.InsertExcept(ctx, &uG, []string{"resource_path"}, s.DB.Exec)
	if err != nil {
		return ctx, fmt.Errorf("insert user group error %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return ctx, errors.New("cannot insert teacher for testing")
	}

	err = s.createUserAccessPath(ctx, id)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("createUserAccessPath %v", err)
	}
	return ctx, nil
}

// nolint:goconst
func (s *suite) theStudentIs(ctx context.Context, arg1, arg2 string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	switch arg1 {
	case "name":
		query := "UPDATE users SET name = $1 WHERE user_id = $2"
		if _, err := s.DB.Exec(ctx, query, &arg2, &stepState.CurrentStudentID); err != nil {
			return ctx, err
		}
		return ctx, nil
	case "grade":
		grade, _ := strconv.Atoi(arg2)
		query := "UPDATE students SET current_grade = $1 WHERE student_id = $2"
		if _, err := s.DB.Exec(ctx, query, &grade, &stepState.CurrentStudentID); err != nil {
			return ctx, err
		}
		return ctx, nil
	case "country":
		query := "UPDATE users SET country = $1 WHERE user_id = $2"
		if _, err := s.DB.Exec(ctx, query, &arg2, &stepState.CurrentStudentID); err != nil {
			return ctx, err
		}
		return ctx, nil
	}
	return ctx, nil
}
