package bob

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/manabie-com/backend/features/common"
	internal_auth "github.com/manabie-com/backend/internal/golibs/auth"
	internal_auth_tenant "github.com/manabie-com/backend/internal/golibs/auth/multitenant"
	internal_auth_user "github.com/manabie-com/backend/internal/golibs/auth/user"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/try"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	"github.com/manabie-com/backend/internal/yasuo/utils"
	pb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/jackc/pgtype"
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *suite) anIdentifyPlatformAccountWithExistedAccountInDB(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	userRepo := &repository.UserRepo{}
	userEnt, err := userRepo.Get(ctx, s.DB, database.Text(stepState.UserID))
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	userEnt.UserAdditionalInfo.Password = fmt.Sprintf("password-%v", idutil.ULIDNow())
	stepState.UserPassword = userEnt.UserAdditionalInfo.Password

	orgRepo := (&repository.OrganizationRepo{}).WithDefaultValue(s.Cfg.Common.Environment)
	tenantID, err := orgRepo.GetTenantIDByOrgID(ctx, s.DB, fmt.Sprint(constants.ManabieSchool))
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	err = s.createUsersInIdentityPlatform(ctx, tenantID, []*entity.LegacyUser{userEnt}, 1)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.TenantID = tenantID
	return StepStateToContext(ctx, stepState), nil
}

// func (s *suite) aValidAuthenticationTokenWithTenant(ctx context.Context) (context.Context, error) {
// 	stepState := StepStateFromContext(ctx)
// 	repo := &repository.UserRepo{}
// 	userEnt, err := repo.Get(ctx, s.DB, database.Text(stepState.UserID))
// 	if err != nil {
// 		return StepStateToContext(ctx, stepState), err
// 	}
// 	err = s.loginIdentityPlatform(ctx, stepState.TenantID, userEnt.GetEmail(), stepState.UserPassword)
// 	if err != nil {
// 		return StepStateToContext(ctx, stepState), err
// 	}

// 	return StepStateToContext(ctx, stepState), nil
// }

func (s *suite) aValidAuthenticationTokenWithTenant(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	repo := &repository.UserRepo{}
	userEnt, err := repo.Get(ctx, s.DB, database.Text(stepState.UserID))
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.AuthToken, err = s.loginIdentityPlatformV1(ctx, stepState.TenantID, userEnt.GetEmail(), stepState.UserPassword)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) loginIdentityPlatformV1(ctx context.Context, tenantID string, email string, password string) (string, error) {
	url := fmt.Sprintf("https://identitytoolkit.googleapis.com/v1/accounts:signInWithPassword?key=%s", s.Cfg.FirebaseAPIKey)

	loginInfo := struct {
		TenantID          string `json:"tenantId"`
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
			IDToken string `json:"idToken"`
		}

		r := &result{}
		if err := json.Unmarshal(data, &r); err != nil {
			return "", errors.Wrap(err, "failed to login and failed to decode error")
		}
		return r.IDToken, nil
	}

	return "", errors.New("failed to login " + string(data))
}

func (s *suite) createUsersInIdentityPlatform(ctx context.Context, tenantID string, users []*entity.LegacyUser, resourcePath int64) error {
	zapLogger := ctxzap.Extract(ctx)

	tenantClient, err := s.TenantManager.TenantClient(ctx, tenantID)
	if err != nil {
		zapLogger.Sugar().Warnw(
			"cannot get tenant client",
			"tenantID", tenantID,
			"err", err.Error(),
		)
		return errors.Wrap(err, "TenantClient")
	}

	err = createUserInAuthPlatform(ctx, tenantClient, users, resourcePath)
	if err != nil {
		zapLogger.Sugar().Warnw(
			"cannot create users on identity platform",
			"err", err.Error(),
		)
		return errors.Wrap(err, "createUserInAuthPlatform")
	}

	return nil
}

func createUserInAuthPlatform(ctx context.Context, authClient internal_auth_tenant.TenantClient, users []*entity.LegacyUser, schoolID int64) error {
	var authUsers internal_auth_user.Users
	for i := range users {
		users[i].CustomClaims = utils.CustomUserClaims(users[i].Group.String, users[i].ID.String, schoolID)

		passwordSalt := []byte(idutil.ULIDNow())

		hashedPwd, err := internal_auth.HashedPassword(authClient.GetHashConfig(), []byte(users[i].Password), passwordSalt)
		if err != nil {
			return errors.Wrap(err, "HashedPassword")
		}

		users[i].PhoneNumber.Status = pgtype.Null
		users[i].PhoneNumber = database.Text("")
		users[i].PasswordSalt = passwordSalt
		users[i].PasswordHash = hashedPwd

		authUsers = append(authUsers, users[i])
	}

	result, err := authClient.ImportUsers(ctx, authUsers, authClient.GetHashConfig())
	if err != nil {
		return errors.Wrapf(err, "ImportUsers")
	}

	if len(result.UsersFailedToImport) > 0 {
		var errs []string
		for _, userFailedToImport := range result.UsersFailedToImport {
			errs = append(errs, fmt.Sprintf("%s - %s", userFailedToImport.User.GetEmail(), userFailedToImport.Err))
		}
		return status.Error(codes.InvalidArgument, fmt.Sprintf("create user in auth platform: %s", strings.Join(errs, ", ")))
	}
	return nil
}

func (s *suite) aClientExchangeCustomToken(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(contextWithValidVersion(ctx))

	_ = try.Do(func(attempt int) (bool, error) {
		stepState.Response, stepState.ResponseErr = pb.NewUserModifierServiceClient(s.Conn).
			ExchangeCustomToken(ctx, &pb.ExchangeCustomTokenRequest{
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
func (s *suite) ourSystemNeedToDoReturnValidCustomToken(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected error when exchange token: %w", stepState.ResponseErr)
	}

	resp := stepState.Response.(*pb.ExchangeCustomTokenResponse)

	if stepState.IsParentGroup {
		return s.comparewithCustomToken(ctx, stepState.AuthToken, resp.CustomToken, entity.UserGroupParent)
	}

	return s.comparewithCustomToken(ctx, stepState.AuthToken, resp.CustomToken, entity.UserGroupStudent)
}
func (s *suite) ourSystemNeedToDoReturnError(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr == nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected error in response")
	}

	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) aNewParentProfileInDB(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if ctx, err := s.signedAsAccountV2(ctx, "staff granted role school admin"); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.signedAsAccountV2 %v", err)
	}

	if ctx, err := s.createUserParent(ctx); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.createUserParent %v", err)
	}

	resp := stepState.Response.(*upb.CreateParentsAndAssignToStudentResponse)
	idStr := resp.ParentProfiles[0].Parent.UserProfile.UserId

	token, err := generateValidAuthenticationToken(idStr, constant.UserGroupParent)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("generateValidAuthenticationToken %v", err)
	}

	stepState.IsParentGroup = true
	stepState.UserID = idStr
	stepState.AuthToken = token
	ctx = common.ValidContext(ctx, constants.ManabieSchool, idStr, stepState.AuthToken)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) createUserParent(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.addParentDataToCreateParentReq(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if ctx, err := s.createNewParents(ctx, schoolAdminType); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) comparewithCustomToken(ctx context.Context, original, customexchanged, userGroup string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	respClaims, err := getUnsafeClaims(customexchanged)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("resp getUnsafeClaims %w", err)
	}

	originClaims, err := getUnsafeClaims(original)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("origin getUnsafeClaims %w", err)
	}

	if originClaims.ID != respClaims.ID {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected issuer %s, got %s", originClaims.Subject, respClaims.Subject)
	}

	query := `
		SELECT group_id
		FROM users_groups AS ug
		WHERE user_id = $1
			AND ug.status = 'USER_GROUP_STATUS_ACTIVE'
	`

	var id pgtype.Text
	err = s.DB.QueryRow(ctx, query, stepState.UserID).Scan(&id)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}

	var idStr string
	_ = id.AssignTo(&idStr)

	if idStr != userGroup {
		return StepStateToContext(ctx, stepState), fmt.Errorf("user group not equal")
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) addParentDataToCreateParentReq(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	parentID := newID()
	profiles := []*upb.CreateParentsAndAssignToStudentRequest_ParentProfile{
		{
			Name:         fmt.Sprintf("user-%v", parentID),
			CountryCode:  cpb.Country_COUNTRY_VN,
			PhoneNumber:  fmt.Sprintf("phone-number-%v", parentID),
			Email:        fmt.Sprintf("%v@example.com", parentID),
			Username:     fmt.Sprintf("username%v", parentID),
			Relationship: upb.FamilyRelationship_FAMILY_RELATIONSHIP_FATHER,
			Password:     fmt.Sprintf("password-%v", parentID),
			UserNameFields: &upb.UserNameFields{
				FirstName:         fmt.Sprintf("first-name-%v", parentID),
				LastName:          fmt.Sprintf("last-name-%v", parentID),
				FirstNamePhonetic: fmt.Sprintf("first-name-phonetic-%v", parentID),
				LastNamePhonetic:  fmt.Sprintf("last-name-phonetic-%v", parentID),
			},
		},
	}
	ctx = s.onlyStudentInfo(ctx)
	ctx, err := s.createNewStudentAccount(ctx, schoolAdminType)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), stepState.ResponseErr
	}
	stepState.Request = &upb.CreateParentsAndAssignToStudentRequest{
		SchoolId:       constants.ManabieSchool,
		StudentId:      stepState.Response.(*upb.CreateStudentResponse).StudentProfile.Student.UserProfile.UserId,
		ParentProfiles: profiles,
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) onlyStudentInfo(ctx context.Context) context.Context {
	stepState := StepStateFromContext(ctx)
	randomID := newID()
	req := &upb.CreateStudentRequest{
		SchoolId: constants.ManabieSchool,
		StudentProfile: &upb.CreateStudentRequest_StudentProfile{
			Email:             fmt.Sprintf("%v@example.com", randomID),
			Password:          fmt.Sprintf("password-%v", randomID),
			Name:              fmt.Sprintf("user-%v", randomID),
			CountryCode:       cpb.Country_COUNTRY_VN,
			EnrollmentStatus:  upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
			PhoneNumber:       fmt.Sprintf("phone-number-%v", randomID),
			StudentExternalId: fmt.Sprintf("student-external-id-%v", randomID),
			StudentNote:       fmt.Sprintf("some random student note %v", randomID),
			Grade:             5,
			Birthday:          timestamppb.New(time.Now().Add(-87600 * time.Hour)),
			Gender:            upb.Gender_MALE,
			LocationIds:       []string{constants.ManabieOrgLocation},
		},
	}
	stepState.Request = req
	return StepStateToContext(ctx, stepState)
}

func (s *suite) createNewStudentAccount(ctx context.Context, account string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.signedAsAccount(ctx, account)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.RequestSentAt = time.Now()
	stepState.Response, stepState.ResponseErr = upb.NewUserModifierServiceClient(s.UsermgmtConn).CreateStudent(contextWithToken(s, ctx), stepState.Request.(*upb.CreateStudentRequest))

	if stepState.ResponseErr == nil {
		stepState.CurrentStudentID = stepState.
			Response.(*upb.CreateStudentResponse).
			GetStudentProfile().
			GetStudent().
			GetUserProfile().
			GetUserId()
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) createNewParents(ctx context.Context, account string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.signedAsAccountV2(ctx, account)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.RequestSentAt = time.Now()
	stepState.Response, stepState.ResponseErr = upb.NewUserModifierServiceClient(s.UsermgmtConn).CreateParentsAndAssignToStudent(contextWithToken(s, ctx), stepState.Request.(*upb.CreateParentsAndAssignToStudentRequest))

	return StepStateToContext(ctx, stepState), nil
}
