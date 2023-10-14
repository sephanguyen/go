package mastermgmt

import (
	"context"

	"github.com/manabie-com/backend/features/helper"

	"google.golang.org/grpc/metadata"
)

// import (
// 	"bytes"
// 	"context"
// 	"encoding/json"
// 	"fmt"
// 	"io"
// 	"io/ioutil"
// 	"math/rand"
// 	"net/http"
// 	"time"

// 	"github.com/manabie-com/backend/features/helper"
// 	golibs_auth "github.com/manabie-com/backend/internal/golibs/auth"
// 	"github.com/manabie-com/backend/internal/golibs/constants"
// 	"github.com/manabie-com/backend/internal/golibs/database"
// 	"github.com/manabie-com/backend/internal/golibs/idutil"
// 	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
// 	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
// 	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
// 	"github.com/manabie-com/backend/internal/usermgmt/pkg/utils"
// 	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
// 	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

// 	"firebase.google.com/go/auth"
// 	"github.com/jackc/pgx/v4"
// 	"github.com/pkg/errors"
// 	"go.uber.org/multierr"
// 	"google.golang.org/grpc/metadata"
// )

// type userOption func(u *entity.User)

// func withID(id string) userOption {
// 	return func(u *entity.User) {
// 		_ = u.ID.Set(id)
// 	}
// }

// func withRole(group string) userOption {
// 	return func(u *entity.User) {
// 		_ = u.Group.Set(group)
// 	}
// }

// type ImportCSVErrors interface {
// 	GetRowNumber() int32
// 	GetError() string
// }

func contextWithValidVersion(ctx context.Context) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "pkg", "com.manabie.liz", "version", "1.0.0")
}

// func (s *suite) signedAsAccount(ctx context.Context, group string) (context.Context, error) {
// 	stepState := StepStateFromContext(ctx)

// 	if group == "unauthenticated" {
// 		stepState.AuthToken = "random-token"
// 		return StepStateToContext(ctx, stepState), nil
// 	}

// 	if group == "admin" {
// 		return s.aSignedInAdmin(ctx)
// 	}

// 	if group == "student" {
// 		return s.aSignedInStudent(ctx)
// 	}

// 	id := idutil.ULIDNow()
// 	var (
// 		userGroup string
// 		err       error
// 	)

// 	switch group {
// 	case "teacher":
// 		userGroup = constant.UserGroupTeacher
// 	case "school admin":
// 		userGroup = constant.UserGroupSchoolAdmin
// 	case "parent":
// 		userGroup = constant.UserGroupParent
// 	case "organization manager":
// 		userGroup = constant.UserGroupOrganizationManager
// 	}

// 	stepState.CurrentUserID = id
// 	stepState.CurrentUserGroup = userGroup

// 	ctx, err = s.aValidUser(StepStateToContext(ctx, stepState), withID(id), withRole(userGroup))
// 	if err != nil {
// 		return StepStateToContext(ctx, stepState), err
// 	}
// 	stepState.AuthToken, err = s.generateExchangeToken(id, userGroup)
// 	if err != nil {
// 		return StepStateToContext(ctx, stepState), err
// 	}

// 	return StepStateToContext(ctx, stepState), nil
// }

// func (s *suite) aSignedInAdmin(ctx context.Context) (context.Context, error) {
// 	stepState := StepStateFromContext(ctx)

// 	id := idutil.ULIDNow()
// 	var err error

// 	ctx, err = s.aValidUser(StepStateToContext(ctx, stepState), withID(id), withRole(constant.UserGroupAdmin))
// 	if err != nil {
// 		return StepStateToContext(ctx, stepState), fmt.Errorf("aValidUser: %w", err)
// 	}

// 	stepState.AuthToken, err = s.generateExchangeToken(id, constant.UserGroupAdmin)
// 	if err != nil {
// 		return StepStateToContext(ctx, stepState), err
// 	}

// 	stepState.CurrentUserID = id
// 	stepState.CurrentUserGroup = constant.UserGroupAdmin

// 	return ctx, nil
// }

// func generateValidAuthenticationToken(sub, userGroup string) (string, error) {
// 	return generateAuthenticationToken(sub, "templates/"+userGroup+".template")
// }

// func generateAuthenticationToken(sub string, template string) (string, error) {
// 	resp, err := http.Get("http://" + firebaseAddr + "/token?template=" + template + "&UserID=" + sub)
// 	if err != nil {
// 		return "", fmt.Errorf("aValidAuthenticationToken: cannot generate new user token, err: %v", err)
// 	}

// 	b, err := ioutil.ReadAll(resp.Body)
// 	if err != nil {
// 		return "", fmt.Errorf("aValidAuthenticationToken: cannot read token from response, err: %v", err)
// 	}
// 	resp.Body.Close()

// 	return string(b), nil
// }

// func (s *suite) aValidUser(ctx context.Context, opts ...userOption) (context.Context, error) {
// 	stepState := StepStateFromContext(ctx)

// 	schoolID := int64(stepState.CurrentSchoolID)
// 	if schoolID == 0 {
// 		schoolID = constants.ManabieSchool
// 	}
// 	ctx = golibs_auth.InjectFakeJwtToken(ctx, fmt.Sprint(schoolID))

// 	num := rand.Int()

// 	u := &entity.User{}
// 	database.AllNullEntity(u)
// 	firstName := fmt.Sprintf("valid-user-first-name-%d", num)
// 	lastName := fmt.Sprintf("valid-user-last-name-%d", num)

// 	err := multierr.Combine(
// 		u.FullName.Set(helper.CombineFirstNameAndLastNameToFullName(firstName, lastName)),
// 		u.FirstName.Set(firstName),
// 		u.LastName.Set(lastName),
// 		u.PhoneNumber.Set(fmt.Sprintf("+848%d", num)),
// 		u.Email.Set(fmt.Sprintf("valid-user-%d@email.com", num)),
// 		u.Country.Set(cpb.Country_COUNTRY_VN.String()),
// 		u.Group.Set(constant.UserGroupAdmin),
// 		u.Avatar.Set(fmt.Sprintf("http://valid-user-%d", num)),
// 		u.ResourcePath.Set(database.Text(fmt.Sprint(schoolID))),
// 	)
// 	if err != nil {
// 		return StepStateToContext(ctx, stepState), err
// 	}

// 	for _, opt := range opts {
// 		opt(u)
// 	}

// 	err = database.ExecInTx(ctx, s.BobDBTrace, func(ctx context.Context, tx pgx.Tx) error {
// 		userRepo := repository.UserRepo{}
// 		err := userRepo.Create(ctx, tx, u)
// 		if err != nil {
// 			return fmt.Errorf("cannot create user: %w", err)
// 		}

// 		switch u.Group.String {
// 		case constant.UserGroupTeacher:
// 			teacherRepo := repository.TeacherRepo{}
// 			t := &entity.Teacher{}
// 			database.AllNullEntity(t)
// 			t.ID = u.ID
// 			err := t.SchoolIDs.Set([]int64{schoolID})
// 			if err != nil {
// 				return err
// 			}

// 			err = teacherRepo.CreateMultiple(ctx, tx, []*entity.Teacher{t})
// 			if err != nil {
// 				return fmt.Errorf("cannot create teacher: %w", err)
// 			}
// 		case constant.UserGroupSchoolAdmin:
// 			schoolAdminRepo := repository.SchoolAdminRepo{}
// 			schoolAdminAccount := &entity.SchoolAdmin{}
// 			database.AllNullEntity(schoolAdminAccount)
// 			err := multierr.Combine(
// 				schoolAdminAccount.SchoolAdminID.Set(u.ID.String),
// 				schoolAdminAccount.SchoolID.Set(schoolID),
// 			)
// 			if err != nil {
// 				return fmt.Errorf("cannot create school admin: %w", err)
// 			}
// 			err = schoolAdminRepo.CreateMultiple(ctx, tx, []*entity.SchoolAdmin{schoolAdminAccount})
// 			if err != nil {
// 				return err
// 			}
// 		case constant.UserGroupParent:
// 			parentRepo := repository.ParentRepo{}
// 			parentEnt := &entity.Parent{}
// 			database.AllNullEntity(parentEnt)
// 			err := multierr.Combine(
// 				parentEnt.ID.Set(u.ID.String),
// 				parentEnt.SchoolID.Set(schoolID),
// 				parentEnt.ResourcePath.Set(fmt.Sprint(schoolID)),
// 			)
// 			if err != nil {
// 				return err
// 			}
// 			err = parentRepo.CreateMultiple(ctx, tx, []*entity.Parent{parentEnt})
// 			if err != nil {
// 				return fmt.Errorf("cannot create parent: %w", err)
// 			}
// 		}

// 		return nil
// 	})
// 	if err != nil {
// 		return StepStateToContext(ctx, stepState), err
// 	}

// 	uGroup := &entity.UserGroup{}
// 	database.AllNullEntity(uGroup)

// 	err = multierr.Combine(
// 		uGroup.GroupID.Set(u.Group.String),
// 		uGroup.UserID.Set(u.ID.String),
// 		uGroup.IsOrigin.Set(true),
// 		uGroup.Status.Set("USER_GROUP_STATUS_ACTIVE"),
// 		uGroup.ResourcePath.Set(database.Text(fmt.Sprint(schoolID))),
// 	)
// 	if err != nil {
// 		return StepStateToContext(ctx, stepState), err
// 	}

// 	userGroupRepo := &repository.UserGroupRepo{}
// 	err = userGroupRepo.Upsert(ctx, s.BobDBTrace, uGroup)
// 	if err != nil {
// 		return StepStateToContext(ctx, stepState), fmt.Errorf("userGroupRepo.Upsert: %w %s", err, u.Group.String)
// 	}

// 	return StepStateToContext(ctx, stepState), nil
// }

// func (s *suite) aSignedInStudent(ctx context.Context) (context.Context, error) {
// 	stepState := StepStateFromContext(ctx)

// 	id := idutil.ULIDNow()
// 	var err error
// 	stepState.AuthToken, err = generateValidAuthenticationToken(id, "phone")
// 	if err != nil {
// 		return StepStateToContext(ctx, stepState), err
// 	}
// 	stepState.CurrentUserID = id
// 	stepState.CurrentUserGroup = constant.UserGroupStudent

// 	return s.aValidStudentInDB(StepStateToContext(ctx, stepState), id)
// }

// func (s *suite) aValidStudentInDB(ctx context.Context, id string) (context.Context, error) {
// 	stepState := StepStateFromContext(ctx)

// 	studentRepo := repository.StudentRepo{}
// 	now := time.Now()
// 	num := rand.Int()
// 	student := &entity.Student{}
// 	database.AllNullEntity(student)
// 	database.AllNullEntity(&student.User)
// 	firstName := fmt.Sprintf("valid-user-first-name-%d", num)
// 	lastName := fmt.Sprintf("valid-user-last-name-%d", num)
// 	err := multierr.Combine(
// 		student.User.ID.Set(id),
// 		student.User.Email.Set(fmt.Sprintf("valid-user-%d@email.com", num)),
// 		student.User.PhoneNumber.Set(fmt.Sprintf("+848%d", num)),
// 		student.User.GivenName.Set(""),
// 		student.User.Avatar.Set(""),
// 		student.User.IsTester.Set(false),
// 		student.User.FacebookID.Set(id),
// 		student.User.PhoneVerified.Set(false),
// 		student.User.EmailVerified.Set(false),
// 		student.User.DeletedAt.Set(nil),
// 		student.User.LastLoginDate.Set(nil),
// 		student.User.FullName.Set(helper.CombineFirstNameAndLastNameToFullName(firstName, lastName)),
// 		student.User.FirstName.Set(firstName),
// 		student.User.LastName.Set(lastName),
// 		student.User.Country.Set(cpb.Country_COUNTRY_VN.String()),
// 		student.User.Group.Set(entity.UserGroupStudent),
// 		student.User.Birthday.Set(now),
// 		student.User.Gender.Set(pb.Gender_MALE),

// 		student.ID.Set(id),
// 		student.CurrentGrade.Set(12),
// 		student.OnTrial.Set(true),
// 		student.TotalQuestionLimit.Set(10),
// 		student.SchoolID.Set(constants.ManabieSchool),
// 		student.CreatedAt.Set(now),
// 		student.UpdatedAt.Set(now),
// 		student.BillingDate.Set(now),
// 		student.EnrollmentStatus.Set("STUDENT_ENROLLMENT_STATUS_ENROLLED"),
// 	)
// 	if err != nil {
// 		return StepStateToContext(ctx, stepState), err
// 	}
// 	err = studentRepo.Create(ctx, s.BobDBTrace, student)
// 	if err != nil {
// 		return StepStateToContext(ctx, stepState), err
// 	}
// 	return StepStateToContext(ctx, stepState), nil
// }

func contextWithToken(s *suite, ctx context.Context) context.Context {
	stepState := StepStateFromContext(ctx)
	return helper.GRPCContext(ctx, "token", stepState.AuthToken)
}

// func (s *suite) loginFirebaseAccount(ctx context.Context, email string, password string) error {
// 	url := fmt.Sprintf("https://identitytoolkit.googleapis.com/v1/accounts:signInWithPassword?key=%s", s.Cfg.FirebaseAPIKey)

// 	loginInfo := struct {
// 		Email             string `json:"email"`
// 		Password          string `json:"password"`
// 		ReturnSecureToken bool   `json:"returnSecureToken"`
// 	}{
// 		Email:             email,
// 		Password:          password,
// 		ReturnSecureToken: true,
// 	}
// 	body, err := json.Marshal(&loginInfo)
// 	if err != nil {
// 		return err
// 	}

// 	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
// 	if err != nil {
// 		return err
// 	}

// 	resp, err := (&http.Client{}).Do(req)
// 	if err != nil {
// 		return err
// 	}

// 	defer resp.Body.Close()

// 	if resp.StatusCode == http.StatusOK {
// 		return nil
// 	}

// 	data, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		return errors.Wrap(err, "failed to login and failed to decode error")
// 	}
// 	return errors.New("failed to login " + string(data))
// }

// func (s *suite) loginIdentityPlatform(ctx context.Context, tenantID string, email string, password string) error {
// 	url := fmt.Sprintf("https://identitytoolkit.googleapis.com/v1/accounts:signInWithPassword?key=%s", s.Cfg.FirebaseAPIKey)

// 	loginInfo := struct {
// 		TenantID          string `json:"tenantId"`
// 		Email             string `json:"email"`
// 		Password          string `json:"password"`
// 		ReturnSecureToken bool   `json:"returnSecureToken"`
// 	}{
// 		TenantID:          tenantID,
// 		Email:             email,
// 		Password:          password,
// 		ReturnSecureToken: true,
// 	}
// 	body, err := json.Marshal(&loginInfo)
// 	if err != nil {
// 		return err
// 	}

// 	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
// 	if err != nil {
// 		return err
// 	}

// 	resp, err := (&http.Client{}).Do(req)
// 	if err != nil {
// 		return err
// 	}

// 	defer resp.Body.Close()

// 	if resp.StatusCode == http.StatusOK {
// 		return nil
// 	}

// 	data, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		return errors.Wrap(err, "failed to login and failed to decode error")
// 	}
// 	return errors.New("failed to login " + string(data))
// }

func (s *suite) signedCtx(ctx context.Context) context.Context {
	stepState := StepStateFromContext(ctx)
	return metadata.AppendToOutgoingContext(contextWithValidVersion(ctx), "token", stepState.AuthToken)
}

// func (s *suite) createFirebaseUser(ctx context.Context, schoolID int32, user entity.User) error {
// 	userToImport := (&auth.UserToImport{}).
// 		UID(user.ID.String).
// 		Email(user.Email.String).
// 		CustomClaims(utils.CustomUserClaims(user.Group.String, user.ID.String, int64(schoolID)))
// 	_, err := s.FirebaseClient.ImportUsers(ctx, []*auth.UserToImport{userToImport})
// 	if err != nil {
// 		return errors.Wrap(err, "ImportUsers()")
// 	}

// 	userToUpdate := (&auth.UserToUpdate{}).Password(user.UserAdditionalInfo.Password)
// 	_, err = s.FirebaseClient.UpdateUser(ctx, user.ID.String, userToUpdate)
// 	if err != nil {
// 		return errors.Wrap(err, "overrideUserPassword()")
// 	}
// 	return nil
// }
