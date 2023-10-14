package utils

import (
	"context"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/manabie-com/backend/features/helper"
	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/eureka/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/yasuo/constant"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
	"google.golang.org/grpc"
)

type AuthHelper struct {
	bobDB    *database.DBTrace
	eurekaDB *database.DBTrace
	fatimaDB *database.DBTrace

	applicantID  string
	firebaseAddr string
	shamirConn   *grpc.ClientConn
}

func NewAuthHelper(bobDB, eurekaDB, fatimaDB *database.DBTrace, applicantID, firebaseAddr string, shamirConn *grpc.ClientConn) *AuthHelper {
	return &AuthHelper{
		bobDB:        bobDB,
		eurekaDB:     eurekaDB,
		fatimaDB:     fatimaDB,
		applicantID:  applicantID,
		firebaseAddr: firebaseAddr,
		shamirConn:   shamirConn,
	}
}
func (a *AuthHelper) SignedCtx(ctx context.Context, authToken string) context.Context {
	return helper.GRPCContext(ctx, "token", authToken)
}

// ASignedInAsRole return new userID, authToken
func (a *AuthHelper) AUserSignedInAsRole(ctx context.Context, userGroup string) (string, string, error) {
	userID, authToken, err := a.aValidToken(ctx, "", userGroup)

	if err != nil {
		return "", "", fmt.Errorf("aValidToken: %w", err)
	}
	return userID, authToken, nil
}
func (a *AuthHelper) SignedInWithUserIDAndRole(ctx context.Context, userID string, userGroup string) (string, error) {
	_, authToken, err := a.aValidToken(ctx, userID, userGroup)

	if err != nil {
		return "", fmt.Errorf("aValidToken: %w", err)
	}
	return authToken, nil
}

func (a *AuthHelper) AValidUser(ctx context.Context, userID, group string) (err error) {
	ID := userID
	var oldGroup string
	switch group {
	case constants.RoleTeacher:
		oldGroup = "USER_GROUP_TEACHER"
	case constants.RoleStudent:
		oldGroup = "USER_GROUP_STUDENT"
	case constants.RoleSchoolAdmin:
		oldGroup = "USER_GROUP_SCHOOL_ADMIN"
	case constants.RoleParent:
		oldGroup = "USER_GROUP_PARENT"
	default:
		oldGroup = "USER_GROUP_PARENT"
	}
	err = aValidUserInDB(ctx, a.bobDB, ID, group, oldGroup)
	if err != nil {
		return
	}
	err = aValidUserInDB(ctx, a.eurekaDB, ID, group, group)
	if err != nil {
		return
	}
	err = aValidUserInDB(ctx, a.fatimaDB, ID, group, group)
	if err != nil {
		return
	}
	return
}

func setFakeClaimToContext(ctx context.Context, resourcePath string, userGroup string) context.Context {
	claims := &interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{
			ResourcePath: resourcePath,
			UserGroup:    userGroup,
		},
	}
	return interceptors.ContextWithJWTClaims(ctx, claims)
}

// TODO: need refactor- add comment
func aValidUserInDB(ctx context.Context, dbTrace *database.DBTrace, userID, newgroup, oldGroup string) error {
	num := rand.Int()
	var now pgtype.Timestamptz
	now.Set(time.Now())
	u := entities.User{}
	database.AllNullEntity(&u)
	u.ID = database.Text(userID)
	u.LastName.Set(fmt.Sprintf("valid-user-%d", num))
	u.PhoneNumber.Set(fmt.Sprintf("+848%d", num))
	u.Email.Set(fmt.Sprintf("valid-user-%d@email.com", num))
	u.Avatar.Set(fmt.Sprintf("http://valid-user-%d", num))
	u.Country.Set(cpb.Country_COUNTRY_VN.String())
	u.Group.Set(oldGroup)
	u.DeviceToken.Set(nil)
	u.AllowNotification.Set(true)
	u.CreatedAt = now
	u.UpdatedAt = now
	u.IsTester.Set(nil)
	u.FacebookID.Set(nil)
	u.ResourcePath.Set(fmt.Sprint(constant.ManabieSchool))

	gr := &entities.Group{}
	database.AllNullEntity(gr)
	gr.ID.Set(oldGroup)
	gr.Name.Set(oldGroup)
	gr.UpdatedAt.Set(time.Now())
	gr.CreatedAt.Set(time.Now())
	fieldNames, _ := gr.FieldMap()
	placeHolders := database.GeneratePlaceholders(len(fieldNames))
	stmt := fmt.Sprintf("INSERT INTO groups (%s) VALUES(%s) ON CONFLICT DO NOTHING", strings.Join(fieldNames, ","), placeHolders)
	if _, err := dbTrace.Exec(ctx, stmt, database.GetScanFields(gr, fieldNames)...); err != nil {
		return fmt.Errorf("insert group error: %v", err)
	}
	ctx = setFakeClaimToContext(context.Background(), u.ResourcePath.String, oldGroup)

	ugroup := &entity.UserGroupV2{}
	database.AllNullEntity(ugroup)
	ugroup.UserGroupID.Set(idutil.ULIDNow())
	ugroup.UserGroupName.Set("name")
	ugroup.UpdatedAt.Set(time.Now())
	ugroup.CreatedAt.Set(time.Now())
	ugroup.ResourcePath.Set(fmt.Sprint(constant.ManabieSchool))

	ugMember := &entity.UserGroupMember{}
	database.AllNullEntity(ugMember)
	ugMember.UserID.Set(u.ID)
	ugMember.UserGroupID.Set(ugroup.UserGroupID.String)
	ugMember.CreatedAt.Set(time.Now())
	ugMember.UpdatedAt.Set(time.Now())
	ugMember.ResourcePath.Set(fmt.Sprint(constant.ManabieSchool))

	uG := entities.UserGroup{
		UserID:   u.ID,
		GroupID:  database.Text(oldGroup),
		IsOrigin: database.Bool(true),
	}
	uG.Status.Set("USER_GROUP_STATUS_ACTIVE")
	uG.CreatedAt = u.CreatedAt
	uG.UpdatedAt = u.UpdatedAt

	role := &entity.Role{}
	database.AllNullEntity(role)
	role.RoleID.Set(idutil.ULIDNow())
	role.RoleName.Set(newgroup)
	role.CreatedAt.Set(time.Now())
	role.UpdatedAt.Set(time.Now())
	role.ResourcePath.Set(fmt.Sprint(constant.ManabieSchool))

	grantedRole := &entity.GrantedRole{}
	database.AllNullEntity(grantedRole)
	grantedRole.RoleID.Set(role.RoleID.String)
	grantedRole.UserGroupID.Set(ugroup.UserGroupID.String)
	grantedRole.GrantedRoleID.Set(idutil.ULIDNow())
	grantedRole.CreatedAt.Set(time.Now())
	grantedRole.UpdatedAt.Set(time.Now())
	grantedRole.ResourcePath.Set(fmt.Sprint(constant.ManabieSchool))

	if _, err := database.InsertOnConflictDoNothing(ctx, &u, dbTrace.Exec); err != nil {
		return fmt.Errorf("insert user error: %v", err)
	}

	if _, err := database.InsertOnConflictDoNothing(ctx, &uG, dbTrace.Exec); err != nil {
		return fmt.Errorf("insert user group error: %v", err)
	}
	if _, err := database.InsertOnConflictDoNothing(ctx, ugroup, dbTrace.Exec); err != nil {
		return fmt.Errorf("insert user group member error: %v", err)
	}

	if _, err := database.InsertOnConflictDoNothing(ctx, ugMember, dbTrace.Exec); err != nil {
		return fmt.Errorf("insert user group member error: %v", err)
	}

	if _, err := database.InsertOnConflictDoNothing(ctx, role, dbTrace.Exec); err != nil {
		return fmt.Errorf("insert user group member error: %v", err)
	}
	if _, err := database.InsertOnConflictDoNothing(ctx, grantedRole, dbTrace.Exec); err != nil {
		return fmt.Errorf("insert user group member error: %v", err)
	}
	if u.Group.String == constant.UserGroupTeacher {
		teacher := &entities.Teacher{}
		database.AllNullEntity(teacher)

		err := multierr.Combine(
			teacher.ID.Set(u.ID.String),
			teacher.SchoolIDs.Set([]int64{constant.ManabieSchool}),
			teacher.UpdatedAt.Set(time.Now()),
			teacher.CreatedAt.Set(time.Now()),
		)
		if err != nil {
			return err
		}

		_, err = database.InsertOnConflictDoNothing(ctx, teacher, dbTrace.Exec)
		if err != nil {
			return fmt.Errorf("insert teacher error: %v", err)
		}
	}

	if u.Group.String == constant.UserGroupStudent {
		student := &entities.Student{}
		database.AllNullEntity(student)
		err := multierr.Combine(
			student.ID.Set(u.ID.String),
			student.SchoolID.Set(constant.ManabieSchool),
			student.EnrollmentStatus.Set("STUDENT_ENROLLMENT_STATUS_ENROLLED"),
			student.StudentNote.Set("example-student-note"),
			student.ResourcePath.Set(fmt.Sprint(constant.ManabieSchool)),
			student.CurrentGrade.Set(12),
			student.OnTrial.Set(true),
			student.TotalQuestionLimit.Set(10),
			student.BillingDate.Set(now),
			student.UpdatedAt.Set(time.Now()),
			student.CreatedAt.Set(time.Now()),
		)
		if err != nil {
			return err
		}

		_, err = database.InsertOnConflictDoNothing(ctx, student, dbTrace.Exec)
		if err != nil {
			return fmt.Errorf("insert student error: %v", err)
		}
	}

	if u.Group.String == constant.UserGroupSchoolAdmin {
		schoolAdminAccount := &entities.SchoolAdmin{}
		database.AllNullEntity(schoolAdminAccount)
		err := multierr.Combine(
			schoolAdminAccount.SchoolAdminID.Set(u.ID.String),
			schoolAdminAccount.SchoolID.Set(u.ResourcePath.String),
			schoolAdminAccount.UpdatedAt.Set(time.Now()),
			schoolAdminAccount.CreatedAt.Set(time.Now()),
			schoolAdminAccount.ResourcePath.Set(fmt.Sprint(constant.ManabieSchool)),
		)
		if err != nil {
			return err
		}

		_, err = database.InsertOnConflictDoNothing(ctx, schoolAdminAccount, dbTrace.Exec)
		if err != nil {
			return fmt.Errorf("aValidUser insert school error: %w", err)
		}
	}
	return nil
}

func (a *AuthHelper) aValidToken(ctx context.Context, uID string, userGroup string) (userID, authToken string, err error) {
	userID = uID
	if userID == "" {
		userID = idutil.ULIDNow()
	}
	var oldGroup string
	switch userGroup {
	case "teacher", "current teacher", "lead teacher":
		userGroup = constants.RoleTeacher
		oldGroup = entities.UserGroupTeacher
	case "student":
		userGroup = constants.RoleStudent
		oldGroup = entities.UserGroupStudent
	case "parent":
		userGroup = constants.RoleParent
		oldGroup = entities.UserGroupParent
	case "school admin", "admin":
		userGroup = constants.RoleSchoolAdmin
		oldGroup = entities.UserGroupSchoolAdmin
	case "hq staff":
		userGroup = constants.RoleHQStaff
		oldGroup = entities.UserGroupParent
		/*
			TODO: we'll change belows roles correctly when user team adds them
			For now, just using "constantsnt.RoleParent" temporary instead
		*/
	case "center manager":
		userGroup = constants.RoleCentreManager
		oldGroup = entities.UserGroupParent
	case "center lead", "center staff":
		userGroup = constants.RoleParent
		oldGroup = entities.UserGroupParent
	default:
		userGroup = constants.RoleStudent
		oldGroup = entities.UserGroupStudent
	}
	err = a.AValidUser(ctx, userID, userGroup)
	if err != nil {
		return "", "", fmt.Errorf("aValidUserInDB: %w", err)
	}
	token, err := a.GenerateExchangeToken(userID, oldGroup)
	if err != nil {
		return "", "", fmt.Errorf("generateExchangeToken: %w", err)
	}
	// switch userGroup {
	// case constants.RoleSchoolAdmin:
	// 	stepState.SchoolAdminToken = stepState.AuthToken
	// case constants.RoleStudent:
	// 	stepState.StudentToken = stepState.AuthToken
	// case constants.RoleTeacher:
	// 	stepState.TeacherToken = stepState.AuthToken
	// }
	return userID, token, nil
}

func (a *AuthHelper) GenerateExchangeToken(userID, userGroup string) (string, error) {
	firebaseToken, err := a.generateValidAuthenticationToken(userID, userGroup)
	if err != nil {
		return "", err
	}
	// ALL test should have one resource_path
	token, err := helper.ExchangeToken(firebaseToken, userID, userGroup, a.applicantID, constant.ManabieSchool, a.shamirConn)
	if err != nil {
		return "", err
	}
	return token, nil
}

func (a *AuthHelper) generateValidAuthenticationToken(sub, userGroup string) (string, error) {
	url := ""
	switch userGroup {
	case "USER_GROUP_TEACHER":
		url = "http://" + a.firebaseAddr + "/token?template=templates/USER_GROUP_TEACHER.template&UserID="
	case "USER_GROUP_SCHOOL_ADMIN":
		url = "http://" + a.firebaseAddr + "/token?template=templates/USER_GROUP_SCHOOL_ADMIN.template&UserID="
	default:
		url = "http://" + a.firebaseAddr + "/token?template=templates/phone.template&UserID="
	}

	resp, err := http.Get(url + sub)
	if err != nil {
		return "", fmt.Errorf("aValidAuthenticationToken:cannot generate new user token, err: %v", err)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("aValidAuthenticationToken: cannot read token from response, err: %v", err)
	}
	resp.Body.Close()
	return string(b), nil
}
