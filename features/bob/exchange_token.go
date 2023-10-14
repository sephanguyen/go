package bob

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/manabie-com/backend/features/common"
	"github.com/manabie-com/backend/features/usermgmt"
	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/auth"
	internal_auth_tenant "github.com/manabie-com/backend/internal/golibs/auth/multitenant"
	internal_auth_user "github.com/manabie-com/backend/internal/golibs/auth/user"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/try"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	oldPb "github.com/manabie-com/backend/pkg/genproto/bob"
	pb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/square/go-jose/v3/jwt"
	"go.uber.org/multierr"
)

func (s *suite) aUserExchangeToken(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(contextWithValidVersion(ctx))
	var err error
	stepState.AuthToken, err = generateValidAuthenticationTokenV1(stepState.CurrentUserID)
	if err != nil {
		return StepStateToContext(ctx, stepState), nil
	}

	retryCtx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	err = usermgmt.TryUntilSuccess(retryCtx, 200*time.Millisecond, func(ctx context.Context) (bool, error) {
		stepState.Response, stepState.ResponseErr = pb.NewUserModifierServiceClient(s.Conn).
			ExchangeToken(ctx, &pb.ExchangeTokenRequest{
				Token: stepState.AuthToken,
			})
		if stepState.ResponseErr == nil {
			return false, nil
		}
		return true, stepState.ResponseErr
	})
	return StepStateToContext(ctx, stepState), errors.Wrap(err, "usermgmt.TryUntilSuccess")
	/*_ = try.Do(func(attempt int) (bool, error) {
		stepState.Response, stepState.ResponseErr = pb.NewUserModifierServiceClient(s.Conn).
			ExchangeToken(ctx, &pb.ExchangeTokenRequest{
				Token: stepState.AuthToken,
			})
		if stepState.ResponseErr == nil {
			return false, nil
		}
		if attempt < 10 {
			time.Sleep(time.Millisecond * 200)
			return true, stepState.ResponseErr
		}
		return false, nil
	})
	return StepStateToContext(ctx, stepState), nil*/
}

func (s *suite) ourSystemNeedToDoReturnValidToken(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return ctx, fmt.Errorf("unexpected error when exchange token: %w", stepState.ResponseErr)
	}

	resp := stepState.Response.(*pb.ExchangeTokenResponse)

	return ctx, compareToken(s.Cfg.JWTApplicant, stepState.AuthToken, resp.Token, entities.UserGroupStudent, []int64{})
}

func compareToken(applicant, original, exchanged, userGroup string, schoolIds []int64) error {
	respClaims, err := getUnsafeClaims(exchanged)
	if err != nil {
		return fmt.Errorf("resp getUnsafeClaims %w", err)
	}
	originClaims, err := getUnsafeClaims(original)
	if err != nil {
		return fmt.Errorf("origin getUnsafeClaims %w", err)
	}

	var stringSchoolIDs []string
	if len(schoolIds) > 0 {
		stringSchoolIDs = golibs.ToArrayStringFromArrayInt64(schoolIds)
	}

	// Check jwt claims
	switch {
	case len(respClaims.Audience) < 1:
		return errors.New("jwt audience is empty")
	case respClaims.Audience[0] != applicant:
		return fmt.Errorf("expecting jwt aud: %s, got %s", applicant, respClaims.Audience[0])
	case respClaims.Subject != originClaims.Subject:
		return fmt.Errorf("expected jwt issuer %s, got %s", originClaims.Subject, respClaims.Subject)
	case !respClaims.Expiry.Time().After(originClaims.Expiry.Time()):
		return fmt.Errorf("expecting respClaims expire (%s) after originalClaims (%s)", respClaims.Expiry.Time().String(), originClaims.Expiry.Time().String())
	}

	// Check manabie claims
	switch {
	case respClaims.Manabie.UserID != originClaims.Subject:
		return fmt.Errorf("expecting manabie user_id: %s, got %s", originClaims.Subject, respClaims.Manabie.UserID)
	case respClaims.Manabie.UserGroup != userGroup:
		return fmt.Errorf("expecting manabie user group %s, got %s", userGroup, respClaims.Manabie.UserGroup)
	case respClaims.Manabie.DefaultRole != userGroup:
		return fmt.Errorf("expecting manabie default role %s, got %s", userGroup, respClaims.Manabie.DefaultRole)
	case len(respClaims.Manabie.AllowedRoles) < 1:
		return errors.New("manabie allowed roles is empty")
	case respClaims.Manabie.AllowedRoles[0] != userGroup:
		return fmt.Errorf("expecting manabie first allowed role %s, got %s", userGroup, respClaims.Manabie.AllowedRoles[0])
	case len(respClaims.Manabie.SchoolIDs) != len(stringSchoolIDs):
		return errors.New("length of actual manabie school ids and expected school ids is different")
	case !golibs.EqualStringArray(stringSchoolIDs, respClaims.Manabie.SchoolIDs):
		return fmt.Errorf("expected %s manabie school ids got %s", stringSchoolIDs, respClaims.Manabie.SchoolIDs)
	}

	// Check hasura claims
	switch {
	case respClaims.Hasura.UserID != originClaims.Subject:
		return fmt.Errorf("expected hasura user id: %s, got %s", originClaims.Subject, respClaims.Hasura.UserID)
	case respClaims.Hasura.UserGroup != userGroup:
		return fmt.Errorf("expecting hasura user group %s, got %s", userGroup, respClaims.Manabie.UserGroup)
	case respClaims.Hasura.DefaultRole != userGroup:
		return fmt.Errorf("expecting hasura default role: %s, got %s", userGroup, respClaims.Hasura.DefaultRole)
	case len(respClaims.Hasura.AllowedRoles) < 1:
		return errors.New("hasura allowed roles is empty")
	case respClaims.Hasura.AllowedRoles[0] != userGroup:
		return fmt.Errorf("expecting first allowed role %s, got %s", userGroup, respClaims.Manabie.AllowedRoles[0])
	case fmt.Sprintf("{%s}", strings.Join(stringSchoolIDs, ",")) != respClaims.Hasura.SchoolIDs:
		return fmt.Errorf("expected %s hasura school ids got %s", fmt.Sprintf("{%s}", strings.Join(stringSchoolIDs, ",")), respClaims.Manabie.SchoolIDs)
	}

	return nil
}

func getUnsafeClaims(t string) (*interceptors.CustomClaims, error) {
	tok, err := jwt.ParseSigned(t)
	if err != nil {
		return nil, fmt.Errorf("unexpected error when parse token: %w", err)
	}
	claims := &interceptors.CustomClaims{}
	if err := tok.UnsafeClaimsWithoutVerification(claims); err != nil {
		return nil, fmt.Errorf("unexpected error when reading claims: %w", err)
	}
	return claims, nil
}

func (s *suite) ourSystemNeedToDoReturnValidSchoolAdminToken(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return ctx, fmt.Errorf("unexpected error when exchange token: %w", stepState.ResponseErr)
	}

	resp := stepState.Response.(*pb.ExchangeTokenResponse)
	schoolIDs := []int64{int64(stepState.CurrentSchoolID)}
	return ctx, compareToken(s.Cfg.JWTApplicant, stepState.AuthToken, resp.Token, entities.UserGroupSchoolAdmin, schoolIDs)
}
func (s *suite) anSchoolAdminProfileInDB(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx = common.ValidContext(ctx, constants.ManabieSchool, s.RootAccount[constants.ManabieSchool].UserID, s.RootAccount[constants.ManabieSchool].Token)

	stepState.CurrentSchoolID = constants.ManabieSchool
	return s.aSignedInSchoolAdminWithSchoolID(ctx, entities.UserGroupSchoolAdmin, int(stepState.CurrentSchoolID))
}

func (s *suite) ourSystemNeedToDoReturnValidExchangedToken(ctx context.Context, userGroupString string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var userGroup string
	switch userGroupString {
	case "school admin":
		userGroup = entities.UserGroupSchoolAdmin
	case "teacher":
		userGroup = entities.UserGroupTeacher
	}

	if stepState.ResponseErr != nil {
		return ctx, fmt.Errorf("unexpected error when exchange token: %w", stepState.ResponseErr)
	}

	resp := stepState.Response.(*pb.ExchangeTokenResponse)
	schoolIDs := []int64{int64(stepState.CurrentSchoolID)}

	// fmt.Println("stepState.AuthToken", stepState.AuthToken)
	// s.ZapLogger.Sugar().Info("stepState.AuthToken: ", stepState.AuthToken)
	// fmt.Println("resp.Token", resp.Token)
	// s.ZapLogger.Sugar().Info("resp.Token: ", resp.Token)

	return ctx, compareToken(s.Cfg.JWTApplicant, stepState.AuthToken, resp.Token, userGroup, schoolIDs)

}

// ////////////////////////////

const (
	LegacyStudentUserType     = "legacy student"
	LegacyParentUserType      = "legacy parent"
	LegacyTeacherUserType     = "legacy teacher"
	LegacySchoolAdminUserType = "legacy school admin"

	StudentUserType        = "student"
	ParentUserType         = "parent"
	TeacherUserType        = "teacher"
	SchoolAdminUserType    = "school admin"
	StudentWithKidsType    = "student with kids type"
	StudentWithAPlusType   = "student with a+ type"
	StudentWithInvalidType = "student with invalid type"
)

const ResourcePathFieldName = "resource_path"

func aValidUserEntity(userType string, userID string, schoolID int32, createdAt time.Time, updatedAt time.Time) (*entities.User, error) {
	num := rand.Int()

	user := &entities.User{}
	database.AllNullEntity(user)

	var userGroup string
	switch userType {
	case StudentUserType:
		userGroup = entities.UserGroupStudent
	case ParentUserType:
		userGroup = entities.UserGroupParent
	case TeacherUserType:
		userGroup = entities.UserGroupTeacher
	case SchoolAdminUserType:
		userGroup = entities.UserGroupSchoolAdmin
	}

	userType = strings.ReplaceAll(userType, " ", "-")

	err := multierr.Combine(
		user.ID.Set(userID),
		user.LastName.Set(fmt.Sprintf("valid-%s-%d", userType, num)),
		user.PhoneNumber.Set(fmt.Sprintf("+848%d", num)),
		user.Email.Set(fmt.Sprintf("valid-%s-%d@email.com", userType, num)),
		user.Avatar.Set(fmt.Sprintf("http://valid-%s-%d", userType, num)),
		user.Country.Set(oldPb.COUNTRY_VN.String()),
		user.Group.Set(userGroup),
		user.DeviceToken.Set(nil),
		user.AllowNotification.Set(true),
		user.CreatedAt.Set(createdAt),
		user.UpdatedAt.Set(updatedAt),
		user.IsTester.Set(nil),
		user.FacebookID.Set(nil),
		user.ResourcePath.Set(fmt.Sprint(schoolID)),
	)

	if err != nil {
		return nil, err
	}

	return user, nil
}

func aValidUserGroupEntity(userType string, userID string, createdAt time.Time, updatedAt time.Time) (*entities.UserGroup, error) {
	userGroup := &entities.UserGroup{}
	database.AllNullEntity(userGroup)

	var userGroupID string
	switch userType {
	case StudentUserType:
		userGroupID = oldPb.USER_GROUP_STUDENT.String()
	case ParentUserType:
		userGroupID = oldPb.USER_GROUP_PARENT.String()
	case TeacherUserType:
		userGroupID = oldPb.USER_GROUP_TEACHER.String()
	case SchoolAdminUserType:
		userGroupID = oldPb.USER_GROUP_SCHOOL_ADMIN.String()
	}

	err := multierr.Combine(
		userGroup.UserID.Set(userID),
		userGroup.Status.Set("USER_GROUP_STATUS_ACTIVE"),
		userGroup.IsOrigin.Set(true),
		userGroup.GroupID.Set(userGroupID),
		userGroup.CreatedAt.Set(createdAt),
		userGroup.UpdatedAt.Set(updatedAt),
	)
	if err != nil {
		return nil, err
	}

	return userGroup, nil
}

func aValidStudentEntity(id string, schoolID int32, createdAt time.Time, updatedAt time.Time) (*entities.Student, error) {
	studentEnt := &entities.Student{}
	database.AllNullEntity(&studentEnt.User)
	database.AllNullEntity(studentEnt)

	err := multierr.Combine(
		studentEnt.ID.Set(id),
		studentEnt.SchoolID.Set(schoolID),
		studentEnt.Group.Set(oldPb.USER_GROUP_STUDENT.String()),
		studentEnt.UpdatedAt.Set(updatedAt),
		studentEnt.CreatedAt.Set(createdAt),
		studentEnt.OnTrial.Set(false),
		studentEnt.BillingDate.Set(createdAt),
		studentEnt.EnrollmentStatus.Set("STUDENT_ENROLLMENT_STATUS_ENROLLED"),
		studentEnt.StudentNote.Set("example-student-note"),
		studentEnt.ResourcePath.Set(fmt.Sprint(schoolID)),
	)
	if err != nil {
		return nil, err
	}
	return studentEnt, nil
}

func aValidParentEntity(id string, schoolID int32, createdAt time.Time, updatedAt time.Time) (*entities.Parent, error) {
	parentEnt := &entities.Parent{}
	database.AllNullEntity(&parentEnt.User)
	database.AllNullEntity(parentEnt)

	err := multierr.Combine(
		parentEnt.ID.Set(id),
		parentEnt.SchoolID.Set(schoolID),
		parentEnt.Group.Set(oldPb.USER_GROUP_PARENT.String()),
		parentEnt.UpdatedAt.Set(createdAt),
		parentEnt.CreatedAt.Set(updatedAt),
		parentEnt.ResourcePath.Set(fmt.Sprint(schoolID)),
	)
	if err != nil {
		return nil, err
	}
	return parentEnt, nil
}

func aValidTeacherEntity(id string, schoolIDs []int32, createdAt time.Time, updatedAt time.Time) (*entities.Teacher, error) {
	teacherEnt := &entities.Teacher{}
	database.AllNullEntity(&teacherEnt.User)
	database.AllNullEntity(teacherEnt)

	err := multierr.Combine(
		teacherEnt.ID.Set(id),
		teacherEnt.SchoolIDs.Set(schoolIDs),
		teacherEnt.CreatedAt.Set(createdAt),
		teacherEnt.UpdatedAt.Set(updatedAt),
	)
	if err != nil {
		return nil, err
	}
	return teacherEnt, nil
}

func aValidSchoolAdminEntity(id string, schoolID int32, createdAt time.Time, updatedAt time.Time) (*entities.SchoolAdmin, error) {
	schoolAdminEnt := &entities.SchoolAdmin{}
	database.AllNullEntity(schoolAdminEnt)

	err := multierr.Combine(
		schoolAdminEnt.SchoolAdminID.Set(id),
		schoolAdminEnt.SchoolID.Set(schoolID),
		schoolAdminEnt.CreatedAt.Set(createdAt),
		schoolAdminEnt.UpdatedAt.Set(updatedAt),
		schoolAdminEnt.ResourcePath.Set(fmt.Sprint(schoolID)),
	)
	if err != nil {
		return nil, err
	}
	return schoolAdminEnt, nil
}

func (s *suite) aValidLegacyStudentProfileWithId(ctx context.Context, id string, schoolID int32) (context.Context, error) {
	now := time.Now()

	userEnt, err := aValidUserEntity(StudentUserType, id, schoolID, now, now)
	if err != nil {
		return nil, err
	}

	studentEnt, err := aValidStudentEntity(id, schoolID, now, now)
	if err != nil {
		return nil, err
	}

	userGroupEnt, err := aValidUserGroupEntity(StudentUserType, id, now, now)
	if err != nil {
		return nil, err
	}

	if cmdTag, err := database.Insert(ctx, userEnt, s.DB.Exec); err != nil {
		return ctx, err
	} else if cmdTag.RowsAffected() == 0 {
		return ctx, errors.New("cannot insert student for testing")
	}

	if cmdTag, err := database.Insert(ctx, studentEnt, s.DB.Exec); err != nil {
		return ctx, err
	} else if cmdTag.RowsAffected() == 0 {
		return ctx, errors.New("cannot insert student for testing")
	}

	if cmdTag, err := database.Insert(ctx, userGroupEnt, s.DB.Exec); err != nil {
		return ctx, err
	} else if cmdTag.RowsAffected() == 0 {
		return ctx, errors.New("cannot insert student for testing")
	}

	return ctx, nil
}

func (s *suite) aValidStudentProfileWithId(ctx context.Context, id string, schoolID int32) (context.Context, error) {
	now := time.Now()
	userEnt, err := aValidUserEntity(StudentUserType, id, schoolID, now, now)
	if err != nil {
		return nil, err
	}

	studentEnt, err := aValidStudentEntity(id, schoolID, now, now)
	if err != nil {
		return nil, err
	}

	userGroupEnt, err := aValidUserGroupEntity(StudentUserType, id, now, now)
	if err != nil {
		return nil, err
	}

	if cmdTag, err := database.InsertExcept(ctx, userEnt, []string{ResourcePathFieldName}, s.DB.Exec); err != nil {
		return ctx, fmt.Errorf("database.InsertExcept user %w with ID %s", err, userEnt.ID.String)
	} else if cmdTag.RowsAffected() == 0 {
		return ctx, errors.New("cannot insert student for testing")
	}

	if cmdTag, err := database.InsertExcept(ctx, studentEnt, []string{ResourcePathFieldName}, s.DB.Exec); err != nil {
		return ctx, fmt.Errorf("database.InsertExcept student %w", err)
	} else if cmdTag.RowsAffected() == 0 {
		return ctx, errors.New("cannot insert student for testing")
	}

	if cmdTag, err := database.InsertExcept(ctx, userGroupEnt, []string{ResourcePathFieldName}, s.DB.Exec); err != nil {
		return ctx, fmt.Errorf("database.InsertExcept user group %w", err)
	} else if cmdTag.RowsAffected() == 0 {
		return ctx, errors.New("cannot insert student for testing")
	}

	return ctx, nil
}

func (s *suite) aValidLegacyParentProfileWithId(ctx context.Context, id string, schoolID int32) (context.Context, error) {
	now := time.Now()

	userEnt, err := aValidUserEntity(ParentUserType, id, schoolID, now, now)
	if err != nil {
		return nil, err
	}

	parentEnt, err := aValidParentEntity(id, schoolID, now, now)
	if err != nil {
		return nil, err
	}

	userGroupEnt, err := aValidUserGroupEntity(ParentUserType, id, now, now)
	if err != nil {
		return nil, err
	}

	if cmdTag, err := database.Insert(ctx, userEnt, s.DB.Exec); err != nil {
		return ctx, err
	} else if cmdTag.RowsAffected() == 0 {
		return ctx, errors.New("cannot insert user for testing")
	}

	if cmdTag, err := database.Insert(ctx, parentEnt, s.DB.Exec); err != nil {
		return ctx, err
	} else if cmdTag.RowsAffected() == 0 {
		return ctx, errors.New("cannot insert parent for testing")
	}

	if cmdTag, err := database.Insert(ctx, userGroupEnt, s.DB.Exec); err != nil {
		return ctx, err
	} else if cmdTag.RowsAffected() == 0 {
		return ctx, errors.New("cannot insert user group for testing")
	}

	return ctx, nil
}

func (s *suite) aValidParentProfileWithId(ctx context.Context, id string, schoolID int32) (context.Context, error) {
	now := time.Now()

	userEnt, err := aValidUserEntity(ParentUserType, id, schoolID, now, now)
	if err != nil {
		return nil, err
	}

	parentEnt, err := aValidParentEntity(id, schoolID, now, now)
	if err != nil {
		return nil, err
	}

	userGroupEnt, err := aValidUserGroupEntity(ParentUserType, id, now, now)
	if err != nil {
		return nil, err
	}

	if cmdTag, err := database.InsertExcept(ctx, userEnt, []string{ResourcePathFieldName}, s.DB.Exec); err != nil {
		return ctx, err
	} else if cmdTag.RowsAffected() == 0 {
		return ctx, errors.New("cannot insert user for testing")
	}

	if cmdTag, err := database.InsertExcept(ctx, parentEnt, []string{ResourcePathFieldName}, s.DB.Exec); err != nil {
		return ctx, err
	} else if cmdTag.RowsAffected() == 0 {
		return ctx, errors.New("cannot insert parent for testing")
	}

	if cmdTag, err := database.InsertExcept(ctx, userGroupEnt, []string{ResourcePathFieldName}, s.DB.Exec); err != nil {
		return ctx, err
	} else if cmdTag.RowsAffected() == 0 {
		return ctx, errors.New("cannot insert user group for testing")
	}

	return ctx, nil
}

func (s *suite) aValidLegacyTeacherProfileWithId(ctx context.Context, id string, schoolID int32) (context.Context, error) {
	now := time.Now()

	userEnt, err := aValidUserEntity(TeacherUserType, id, schoolID, now, now)
	if err != nil {
		return nil, err
	}

	teacherEnt, err := aValidTeacherEntity(id, []int32{schoolID}, now, now)
	if err != nil {
		return nil, err
	}

	userGroupEnt, err := aValidUserGroupEntity(TeacherUserType, id, now, now)
	if err != nil {
		return nil, err
	}

	if cmdTag, err := database.Insert(ctx, userEnt, s.DB.Exec); err != nil {
		return ctx, err
	} else if cmdTag.RowsAffected() == 0 {
		return ctx, errors.New("cannot insert user for testing")
	}

	if cmdTag, err := database.Insert(ctx, teacherEnt, s.DB.Exec); err != nil {
		return ctx, err
	} else if cmdTag.RowsAffected() == 0 {
		return ctx, errors.New("cannot insert teacher for testing")
	}

	if cmdTag, err := database.Insert(ctx, userGroupEnt, s.DB.Exec); err != nil {
		return ctx, err
	} else if cmdTag.RowsAffected() == 0 {
		return ctx, errors.New("cannot insert user group for testing")
	}

	return ctx, nil
}

func (s *suite) aValidLegacySchoolAdminProfileWithId(ctx context.Context, id string, schoolID int32) (context.Context, error) {
	now := time.Now()

	userEnt, err := aValidUserEntity(SchoolAdminUserType, id, schoolID, now, now)
	if err != nil {
		return nil, err
	}

	schoolAdminEnt, err := aValidSchoolAdminEntity(id, schoolID, now, now)
	if err != nil {
		return nil, err
	}

	userGroupEnt, err := aValidUserGroupEntity(SchoolAdminUserType, id, now, now)
	if err != nil {
		return nil, err
	}

	if cmdTag, err := database.Insert(ctx, userEnt, s.DB.Exec); err != nil {
		return ctx, fmt.Errorf("database.Insert %w", err)
	} else if cmdTag.RowsAffected() == 0 {
		return ctx, errors.New("cannot insert user for testing")
	}

	if cmdTag, err := database.Insert(ctx, schoolAdminEnt, s.DB.Exec); err != nil {
		return ctx, fmt.Errorf("database.Insert %w", err)
	} else if cmdTag.RowsAffected() == 0 {
		return ctx, errors.New("cannot insert school admin for testing")
	}

	if cmdTag, err := database.Insert(ctx, userGroupEnt, s.DB.Exec); err != nil {
		return ctx, fmt.Errorf("database.Insert %w", err)
	} else if cmdTag.RowsAffected() == 0 {
		return ctx, errors.New("cannot insert user group for testing")
	}

	return ctx, nil
}

func InjectFakeJwtToken(ctx context.Context, resourcePath string) context.Context {
	claim := &interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{
			UserGroup:    cpb.UserGroup_USER_GROUP_SCHOOL_ADMIN.String(),
			ResourcePath: resourcePath,
		},
	}
	return interceptors.ContextWithJWTClaims(ctx, claim)
}

func (s *suite) loginFirebaseOrIdentityPlatform(ctx context.Context, firebaseAPIKey string, tenantID string, email string, password string) (string, error) {
	url := fmt.Sprintf("https://identitytoolkit.googleapis.com/v1/accounts:signInWithPassword?key=%s", firebaseAPIKey)

	loginInfo := struct {
		TenantID          string `json:"tenantId,omitempty"`
		Email             string `json:"email"`
		Password          string `json:"password"`
		ReturnSecureToken bool   `json:"returnSecureToken"`
	}{
		TenantID:          tenantID,
		Email:             email,
		Password:          password,
		ReturnSecureToken: true,
	}
	body, err := json.Marshal(&loginInfo)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}

	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Wrap(err, "failed to login and failed to decode error")
	}

	if resp.StatusCode == http.StatusOK {
		type result struct {
			IdToken string `json:"idToken"`
		}

		r := &result{}
		if err := json.Unmarshal(data, &r); err != nil {
			return "", errors.Wrap(err, "failed to login and failed to decode error")
		}
		return r.IdToken, nil
	}

	return "", errors.New("failed to login " + string(data))
}

func (s *suite) aValidUserInOurSystem(ctx context.Context, userType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	switch userType {
	case LegacySchoolAdminUserType:
		ctx, err := s.aValidLegacySchoolAdminProfileWithId(InjectFakeJwtToken(ctx, strconv.Itoa(int(stepState.CurrentSchoolID))), stepState.CurrentUserID, stepState.CurrentSchoolID)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("aValidLegacySchoolAdminProfileWithId %w", err)
		}
	case SchoolAdminUserType:
		ctx, err := s.aValidSchoolAdminProfileWithId(InjectFakeJwtToken(ctx, strconv.Itoa(int(stepState.CurrentSchoolID))), stepState.CurrentUserID, cpb.UserGroup_USER_GROUP_SCHOOL_ADMIN.String(), int(stepState.CurrentSchoolID))
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("aValidSchoolAdminProfileWithId %w", err)
		}
	case LegacyTeacherUserType:
		ctx, err := s.aValidLegacyTeacherProfileWithId(InjectFakeJwtToken(ctx, strconv.Itoa(int(stepState.CurrentSchoolID))), stepState.CurrentUserID, stepState.CurrentSchoolID)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("aValidLegacyTeacherProfileWithId %w", err)
		}
	case TeacherUserType:
		ctx, err := s.aValidTeacherProfileWithId(InjectFakeJwtToken(ctx, strconv.Itoa(int(stepState.CurrentSchoolID))), stepState.CurrentUserID, stepState.CurrentSchoolID)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("aValidTeacherProfileWithId %w", err)
		}
	case LegacyStudentUserType:
		ctx, err := s.aValidLegacyStudentProfileWithId(InjectFakeJwtToken(ctx, strconv.Itoa(int(stepState.CurrentSchoolID))), stepState.CurrentUserID, stepState.CurrentSchoolID)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("aValidLegacyStudentProfileWithId %w", err)
		}
	case StudentUserType, StudentWithKidsType, StudentWithAPlusType, StudentWithInvalidType:
		ctx, err := s.aValidStudentProfileWithId(InjectFakeJwtToken(ctx, strconv.Itoa(int(stepState.CurrentSchoolID))), stepState.CurrentUserID, stepState.CurrentSchoolID)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("aValidStudentProfileWithId %w", err)
		}
	case LegacyParentUserType:
		ctx, err := s.aValidLegacyParentProfileWithId(InjectFakeJwtToken(ctx, strconv.Itoa(int(stepState.CurrentSchoolID))), stepState.CurrentUserID, stepState.CurrentSchoolID)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("aValidLegacyParentProfileWithId %w", err)
		}
	case ParentUserType:
		ctx, err := s.aValidParentProfileWithId(InjectFakeJwtToken(ctx, strconv.Itoa(int(stepState.CurrentSchoolID))), stepState.CurrentUserID, stepState.CurrentSchoolID)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("aValidParentProfileWithId %w", err)
		}
	default:
		return StepStateToContext(ctx, stepState), fmt.Errorf("this type of user is not supported for now: %w", stepState.ResponseErr)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) signInAuthPlatform(ctx context.Context, authPlatform string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx = InjectFakeJwtToken(ctx, strconv.Itoa(int(stepState.CurrentSchoolID)))
	user, err := (&repository.UserRepo{}).Get(ctx, s.DB, database.Text(stepState.CurrentUserID))
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	var tenantClient internal_auth_tenant.TenantClient
	var tenantID string

	switch authPlatform {
	case "firebase":
		tenantClient = s.FirebaseAuthClient
	case "identity":
		tenantID = auth.LocalTenants[int(stepState.CurrentSchoolID)]

		tenantClient, err = s.TenantManager.TenantClient(ctx, tenantID)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	default:
		return StepStateToContext(ctx, stepState), errors.New("this type of auth platform is not supported")
	}

	// Hash password for user
	pwd := fmt.Sprintf("password%v", stepState.CurrentUserID)
	pwdSalt := stepState.CurrentUserID
	hashedPwd, err := auth.HashedPassword(tenantClient.GetHashConfig(), []byte(pwd), []byte(pwdSalt))
	if err != nil {
		s.ZapLogger.Fatal(fmt.Sprintf("err: %v\n", err))
	}
	user.UserAdditionalInfo.PasswordSalt = []byte(pwdSalt)
	user.UserAdditionalInfo.PasswordHash = hashedPwd

	// skip phone number because firebase/identity will validate based on its rule
	// and our test phone number format will not satisfy
	_ = user.PhoneNumber.Set("")

	// Import user with hashed password
	result, err := tenantClient.ImportUsers(ctx, internal_auth_user.Users{user}, tenantClient.GetHashConfig())
	if err != nil {
		return StepStateToContext(ctx, stepState), errors.Wrap(err, "ImportUsers")
	}
	if len(result.UsersFailedToImport) > 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("failed to import user to tenant, user: %+v", result.UsersFailedToImport[0])
	}

	// Login as a user
	idToken, err := s.loginFirebaseOrIdentityPlatform(ctx, s.GCPApp.ProjectConfig.Client.ApiKey, tenantID, user.Email.String, pwd)
	if err != nil {
		return StepStateToContext(ctx, stepState), errors.Wrap(err, "loginFirebaseOrIdentityPlatform")
	}
	stepState.AuthToken = idToken

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) inLoggedIn(ctx context.Context, userType string, authPlatform string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.CurrentUserID = idutil.ULIDNow()
	stepState.CurrentSchoolID = constants.ManabieSchool

	ctx, err := s.aValidUserInOurSystem(ctx, userType)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	ctx, err = s.signInAuthPlatform(ctx, authPlatform)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

type Account struct {
	Username string
	Password string
}

var mapKeycloakAccounts = map[string]Account{
	"legacy student":       {"student-legacy-integration-testing", "123456"},
	"legacy teacher":       {"teacher-legacy-integration-testing", "123456"},
	"legacy school admin":  {"school-admin-legacy-integration-testing", "123456"},
	"student":              {"student-integration-testing", "123456"},
	"teacher":              {"teacher-integration-testing", "123456"},
	"school admin":         {"school-admin-integration-testing", "123456"},
	StudentWithKidsType:    {"student-kids-type-integration-testing", "123456"},
	StudentWithAPlusType:   {"student-aplus-type-integration-testing", "123456"},
	StudentWithInvalidType: {"student-invalid-student-division-integration-testing", "123456"},
}

func (s *suite) inKeycloakLoggedIn(ctx context.Context, userType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	switch userType {
	case LegacySchoolAdminUserType, LegacyTeacherUserType, LegacyStudentUserType, SchoolAdminUserType, TeacherUserType, StudentUserType, StudentWithKidsType, StudentWithAPlusType, StudentWithInvalidType:
		break
	default:
		return StepStateToContext(ctx, stepState), fmt.Errorf(`"%s" is not supported`, userType)
	}

	ctx = InjectFakeJwtToken(ctx, strconv.Itoa(constants.JPREPSchool))

	account, found := mapKeycloakAccounts[userType]
	if !found {
		return StepStateToContext(ctx, stepState), fmt.Errorf(`"%s" is not supported`, userType)
	}

	idToken, err := s.KeycloakClient.VerifyEmailPassword(ctx, account.Username, account.Password)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.AuthToken = idToken

	claims, err := getUnsafeClaims(stepState.AuthToken)
	if err != nil {
		return StepStateToContext(ctx, stepState), errors.Wrap(err, "getUnsafeClaims")
	}

	stepState.CurrentSchoolID = constants.JPREPSchool
	stepState.CurrentUserID = claims.Subject

	// Because jprep users in db have fixed id so clear before starting test
	err = database.ExecInTx(ctx, s.DBPostgres, func(ctx context.Context, tx pgx.Tx) error {
		_, err = tx.Exec(ctx, `DELETE FROM teachers WHERE teacher_id = $1::text`, stepState.CurrentUserID)
		if err != nil {
			return fmt.Errorf("delete teach error %w", err)
		}
		_, err := tx.Exec(ctx, `DELETE FROM school_admins WHERE school_admin_id = $1::text`, stepState.CurrentUserID)
		if err != nil {
			return fmt.Errorf("delete school_admins error %w", err)
		}
		_, err = tx.Exec(ctx, `DELETE FROM staff WHERE staff_id = $1::text`, stepState.CurrentUserID)
		if err != nil {
			return fmt.Errorf("delete staff error %w", err)
		}
		_, err = tx.Exec(ctx, `DELETE FROM students WHERE student_id = $1::text`, stepState.CurrentUserID)
		if err != nil {
			return fmt.Errorf("delete students error %w", err)
		}
		_, err = tx.Exec(ctx, `DELETE FROM users_groups WHERE user_id = $1::text`, stepState.CurrentUserID)
		if err != nil {
			return fmt.Errorf("delete users_groups error %w", err)
		}
		_, err = tx.Exec(ctx, `DELETE FROM user_access_paths WHERE user_id = $1::text`, stepState.CurrentUserID)
		if err != nil {
			return fmt.Errorf("delete user_access_paths error %w", err)
		}
		_, err = tx.Exec(ctx, `DELETE FROM users WHERE user_id = $1::text`, stepState.CurrentUserID)
		if err != nil {
			return fmt.Errorf("delete users error %w", err)
		}
		return nil
	})

	if err != nil {
		return StepStateToContext(ctx, stepState), errors.Wrap(err, "ExecInTx")
	}

	ctx, err = s.aValidUserInOurSystem(ctx, userType)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("aValidUserInOurSystem %w", err)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) usesIdTokenToExchangesTokenWithOurSystem(ctx context.Context, userType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	_ = try.Do(func(attempt int) (bool, error) {
		stepState.Response, stepState.ResponseErr = pb.NewUserModifierServiceClient(s.Conn).
			ExchangeToken(ctx, &pb.ExchangeTokenRequest{
				Token: stepState.AuthToken,
			})
		if stepState.ResponseErr == nil {
			return false, nil
		}
		if attempt < 10 {
			time.Sleep(time.Millisecond * 200)
			return true, stepState.ResponseErr
		}
		return false, nil
	})
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) receivesValidExchangedToken(ctx context.Context, userType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	var userGroup string
	var schoolIDs []int64

	switch userType {
	case LegacySchoolAdminUserType, SchoolAdminUserType:
		userGroup = entities.UserGroupSchoolAdmin
		schoolIDs = []int64{int64(stepState.CurrentSchoolID)}
	case LegacyTeacherUserType, TeacherUserType:
		userGroup = entities.UserGroupTeacher
		schoolIDs = []int64{int64(stepState.CurrentSchoolID)}
	case LegacyStudentUserType, StudentUserType, StudentWithKidsType, StudentWithAPlusType:
		userGroup = entities.UserGroupStudent
		schoolIDs = []int64{int64(stepState.CurrentSchoolID)}
	case LegacyParentUserType, ParentUserType:
		userGroup = entities.UserGroupParent
		schoolIDs = []int64{int64(stepState.CurrentSchoolID)}
	default:
		return ctx, fmt.Errorf("this type of user is not supported for now: %w", stepState.ResponseErr)
	}

	if stepState.ResponseErr != nil {
		return ctx, fmt.Errorf("unexpected error when exchange token: %w", stepState.ResponseErr)
	}
	// s.ZapLogger.Sugar().Info(stepState.Response.(*pb.ExchangeTokenResponse).Token)

	return StepStateToContext(ctx, stepState), compareToken(s.Cfg.JWTApplicant, stepState.AuthToken, stepState.Response.(*pb.ExchangeTokenResponse).Token, userGroup, schoolIDs)
}

func (s *suite) systemInitDefaultValuesForAuthInfoIn(ctx context.Context, env string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	defaultValues := (&repository.OrganizationRepo{}).DefaultOrganizationAuthValues(env)

	queryCtx := interceptors.ContextWithJWTClaims(ctx, &interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{
			ResourcePath: strconv.Itoa(constants.ManabieSchool),
			UserID:       "bdd_admin+manabie",
			UserGroup:    constant.UserGroupSchoolAdmin,
		},
	})
	_, stepState.ResponseErr = (&repository.UserRepoV2{}).GetByAuthInfo(queryCtx, s.DB, defaultValues, "bdd_admin+manabie", "dev-manabie-online", "manabie-0nl6t")

	organizationAuths, err := (&repository.OrganizationRepo{}).WithDefaultValue(env).GetAll(ctx, s.DB, 1000)
	if err != nil {
		return StepStateToContext(ctx, stepState), errors.Wrap(err, ".GetAll()")
	}

	for _, organizationAuth := range organizationAuths {
		if organizationAuth.OrganizationID.Status != pgtype.Present {
			return StepStateToContext(ctx, stepState), errors.New("queried organization id is not valid")
		}
		if organizationAuth.AuthProjectID.String == "" || organizationAuth.AuthProjectID.Status != pgtype.Present {
			return StepStateToContext(ctx, stepState), errors.New("queried auth project is not valid")
		}
		if organizationAuth.AuthTenantID.Status != pgtype.Present {
			return StepStateToContext(ctx, stepState), errors.New("queried auth tenant is not valid")
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theInitializedValuesMustBeValid(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.ResponseErr != nil {
		return ctx, fmt.Errorf("failed to use default values: %w", stepState.ResponseErr)
	}

	return StepStateToContext(ctx, stepState), nil
}
