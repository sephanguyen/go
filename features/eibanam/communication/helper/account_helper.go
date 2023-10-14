package helper

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/manabie-com/backend/features/eibanam/communication/entity"
	"github.com/manabie-com/backend/features/eibanam/communication/util"
	"github.com/manabie-com/backend/features/helper"
	bentities "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/auth"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/yasuo/constant"
	legacybpb "github.com/manabie-com/backend/pkg/genproto/bob"
	legacyypb "github.com/manabie-com/backend/pkg/genproto/yasuo"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"
)

func (h *CommunicationHelper) CreateSysAdmin(schoolId int64) (*entity.Admin, error) {
	// new admin
	admin := &entity.Admin{}

	ctx := auth.InjectFakeJwtToken(context.Background(), fmt.Sprint(schoolId))

	// fake info
	admin.Email = idutil.ULIDNow() + "-admin@gamil.com"
	admin.ID = idutil.ULIDNow()
	admin.UserGroup = constant.UserGroupAdmin

	// save database
	err := h.saveNewSysAdmin(ctx, admin)
	if err != nil {
		return nil, err
	}

	// fake token
	firebaseToken, err := util.GenerateFakeAuthenticationToken(h.firebaseAddress, admin.ID, admin.UserGroup)
	if err != nil {
		return nil, fmt.Errorf("failed to generate firebase token: %w", err)
	}

	admin.Token, err = helper.ExchangeToken(firebaseToken, admin.ID, admin.UserGroup, h.applicantID, 1, h.shamirGRPCConn)
	if err != nil {
		return nil, fmt.Errorf("failed to generate exchange token: %w", err)
	}
	return admin, nil
}

func (h *CommunicationHelper) CreateNewTeacher(admin *entity.Admin, schoolID int64) (*entity.Teacher, error) {
	req := &legacyypb.CreateUserRequest{
		UserGroup: legacyypb.USER_GROUP_TEACHER,
		SchoolId:  schoolID,
	}
	id := idutil.ULIDNow()
	num := rand.Int()
	email := fmt.Sprintf("e2euser-%s+%d@gmail.com", id, num)
	user := &legacyypb.CreateUserProfile{
		Name:        fmt.Sprintf("e2euser-%s", id),
		PhoneNumber: fmt.Sprintf("+848%d", num),
		Email:       email,
		Country:     legacybpb.COUNTRY_VN,
		Grade:       1,
	}
	req.Users = []*legacyypb.CreateUserProfile{user}
	svc := legacyypb.NewUserServiceClient(h.yasuoGRPCConn)
	ctx, cancel := util.ContextWithTokenAndTimeOut(context.Background(), admin.Token)
	defer cancel()

	res, err := svc.CreateUser(ctx, req)
	if err != nil {
		return nil, err
	}

	teacherID := res.GetUsers()[0].GetId()
	newPassword := idutil.ULIDNow()
	req2 := &upb.ReissueUserPasswordRequest{
		UserId:      teacherID,
		NewPassword: newPassword,
	}

	res2, err := upb.NewUserModifierServiceClient(h.userMgmtGrpcConn).ReissueUserPassword(ctx, req2)
	if err != nil {
		return nil, err
	}
	if !res2.GetSuccessful() {
		return nil, fmt.Errorf("failed to changed password of teacher")
	}
	newTeacher := res.GetUsers()[0]
	return &entity.Teacher{
		User: &entity.User{
			ID:       newTeacher.Id,
			Group:    cpb.UserGroup_USER_GROUP_TEACHER.String(),
			Email:    newTeacher.Email,
			Name:     newTeacher.Name,
			Password: newPassword,
		},
		SchoolID: schoolID,
	}, nil
}

func (h *CommunicationHelper) CreateSchoolAdmin(sysAdmin *entity.Admin, schoolId int64) (*entity.Admin, error) {
	// create request
	req := &legacyypb.CreateUserRequest{
		UserGroup: legacyypb.USER_GROUP_SCHOOL_ADMIN,
		SchoolId:  schoolId,
	}

	id := idutil.ULIDNow()
	num := rand.Int()
	email := fmt.Sprintf("e2euser-%s+%d@gmail.com", id, num)

	user := &legacyypb.CreateUserProfile{
		Name:        fmt.Sprintf("e2euser-%s", id),
		PhoneNumber: fmt.Sprintf("+848%d", num),
		Email:       email,
		Country:     legacybpb.COUNTRY_VN,
	}

	// call yasuo grpc to add an admin to school
	req.Users = []*legacyypb.CreateUserProfile{user}
	svc := legacyypb.NewUserServiceClient(h.yasuoGRPCConn)

	ctx := auth.InjectFakeJwtToken(context.Background(), fmt.Sprint(schoolId))

	ctx, cancel := util.ContextWithTokenAndTimeOut(ctx, sysAdmin.Token)
	defer cancel()
	res, err := svc.CreateUser(ctx, req)
	if err != nil {
		return nil, err
	}

	newSchoolAdmins := res.GetUsers()
	if len(newSchoolAdmins) == 0 {
		return nil, errors.New("create school admin return no rows withour error message")
	}

	schooladmin := &entity.Admin{
		ID:        newSchoolAdmins[0].Id,
		Email:     newSchoolAdmins[0].Email,
		Name:      newSchoolAdmins[0].Name,
		SchoolIds: newSchoolAdmins[0].SchoolIds,
		UserGroup: newSchoolAdmins[0].UserGroup,
	}
	schooladmin.SchoolIds = []int64{schoolId} // somehow api does not return school id, we set it manually

	h.GenerateSchoolAdminRole(ctx, schooladmin)

	return schooladmin, nil
}

func (h *CommunicationHelper) GenerateSchoolAdminRole(ctx context.Context, schoolAdmin *entity.Admin) error {
	resourcePath := strconv.Itoa(int(schoolAdmin.SchoolIds[0]))
	ctx2 := interceptors.ContextWithJWTClaims(ctx, &interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{
			ResourcePath: resourcePath,
		},
	})

	//Get organization location
	accessPath := ""
	queryOrgLoc := "SELECT access_path FROM locations LIMIT 1;"
	if err := database.Select(ctx2, h.BobDBTrace, queryOrgLoc).ScanFields(&accessPath); err != nil {
		return fmt.Errorf("can not get organization location: %v", err)
	}
	orgLocationID := strings.Split(accessPath, "/")[0]

	//Create user_group
	userGroupID := idutil.ULIDNow()
	stmt := `
		INSERT INTO public.user_group
			(user_group_id, user_group_name, created_at, updated_at, deleted_at, resource_path, org_location_id, is_system)
		VALUES($1, $2, 'now()', 'now()', NULL, autofillresourcepath(), $3, true);
	`
	_, err := h.BobDBTrace.Exec(ctx2, stmt, userGroupID, "user group "+userGroupID, orgLocationID)
	if err != nil {
		return fmt.Errorf("cannot insert user_group: %v", err)
	}

	//Get role "School Admin"
	schoolAdminRoleID := ""
	querySchoolAdminRole := "SELECT role_id FROM role WHERE role_name = 'School Admin'"
	if err := database.Select(ctx2, h.BobDBTrace, querySchoolAdminRole).ScanFields(&schoolAdminRoleID); err != nil {
		return fmt.Errorf("can not get School Admin role: %v", err)
	}

	//Create granted_role
	grantedRoleID := idutil.ULIDNow()
	stmt = `
		INSERT INTO public.granted_role
			(granted_role_id, user_group_id, role_id, created_at, updated_at, deleted_at, resource_path)
		VALUES($1, $2, $3, 'now()', 'now()', NULL, autofillresourcepath());
	`
	_, err = h.BobDBTrace.Exec(ctx2, stmt, grantedRoleID, userGroupID, schoolAdminRoleID)
	if err != nil {
		return fmt.Errorf("cannot insert granted_role: %v", err)
	}

	//Create granted_role_access_path
	stmt = `
		INSERT INTO public.granted_role_access_path
			(granted_role_id, location_id, created_at, updated_at, deleted_at, resource_path)
		VALUES($1, $2, 'now()', 'now()', NULL, autofillresourcepath());
	`
	_, err = h.BobDBTrace.Exec(ctx2, stmt, grantedRoleID, orgLocationID)
	if err != nil {
		return fmt.Errorf("cannot insert granted_role_access_path: %v", err)
	}

	//Create user_group_member
	stmt = `
		INSERT INTO public.user_group_member
			(user_id, user_group_id, created_at, updated_at, deleted_at, resource_path)
		VALUES($1, $2, 'now()', 'now()', NULL, autofillresourcepath());
	`
	_, err = h.BobDBTrace.Exec(ctx2, stmt, schoolAdmin.ID, userGroupID)
	if err != nil {
		return fmt.Errorf("cannot insert user_group_member: %v", err)
	}

	return nil
}

func (h *CommunicationHelper) GenerateSchoolAdminPassword(sysAdmin, admin *entity.Admin) error {
	admin.Password = idutil.ULIDNow()
	req := &upb.ReissueUserPasswordRequest{
		UserId:      admin.ID,
		NewPassword: admin.Password,
	}
	ctx, cancel := util.ContextWithTokenAndTimeOut(context.Background(), sysAdmin.Token)
	defer cancel()

	res2, err := upb.NewUserModifierServiceClient(h.userMgmtGrpcConn).ReissueUserPassword(ctx, req)
	if err != nil {
		return err
	}

	if !res2.GetSuccessful() {
		return fmt.Errorf("failed to changed password of school admin")
	}
	return nil
}

func (h *CommunicationHelper) TeacherLoginCms(teacher *entity.Teacher) error {
	authToken, err := h.LoginWithFirebase(teacher.User.Email, teacher.User.Password)
	if err != nil {
		return err
	}

	tokenRes, err := bpb.NewUserModifierServiceClient(h.bobGRPCConn).ExchangeToken(
		context.Background(), &bpb.ExchangeTokenRequest{Token: authToken})
	if err != nil {
		return err
	}

	teacher.User.Token = tokenRes.GetToken()
	return nil
}

func (h *CommunicationHelper) SchoolAdminLoginToCms(ctx context.Context, admin *entity.Admin) error {
	authToken, err := h.GenerateExchangeTokenCtx(ctx, admin.ID, cpb.UserGroup_USER_GROUP_SCHOOL_ADMIN.String())

	// authToken, err := h.LoginWithFirebase(admin.Email, admin.Password)
	if err != nil {
		return err
	}

	// tokenRes, err := bpb.NewUserModifierServiceClient(h.bobGRPCConn).ExchangeToken(
	// 	context.Background(), &bpb.ExchangeTokenRequest{Token: authToken})
	// if err != nil {
	// 	return err
	// }

	// admin.Token = tokenRes.GetToken()
	admin.Token = authToken
	return nil
}

func (h *CommunicationHelper) LoginWithFirebase(email, password string) (string, error) {
	childCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	url := fmt.Sprintf("https://identitytoolkit.googleapis.com/v1/accounts:signInWithPassword?key=%s", h.firebaseKey)

	loginInfo := struct {
		Email             string `json:"email"`
		Password          string `json:"password"`
		ReturnSecureToken bool   `json:"returnSecureToken"`
	}{
		Email:             email,
		Password:          password,
		ReturnSecureToken: true,
	}

	body, _ := json.Marshal(&loginInfo)

	req, err := http.NewRequestWithContext(childCtx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}

	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errors.New("failed to login")
	}

	tokenContainer := struct {
		Token string `json:"idToken"`
	}{}

	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&tokenContainer)

	if err != nil {
		return "", err
	}

	return tokenContainer.Token, nil
}

func (h *CommunicationHelper) saveNewSysAdmin(ctx context.Context, admin *entity.Admin) error {
	u := &bentities.User{
		ID:        database.Text(admin.ID),
		Group:     database.Text(constant.UserGroupAdmin),
		Email:     database.Text(admin.Email),
		Country:   database.Text(constant.CountryVN),
		LastName:  database.Text(admin.ID),
		CreatedAt: database.Timestamptz(time.Now()),
		UpdatedAt: database.Timestamptz(time.Now()),
	}

	fields := []string{"user_id", "email", "user_group", "country", "name", "updated_at", "created_at"}
	placeHolder := database.GeneratePlaceholders(len(fields))

	insertStatement := fmt.Sprintf(
		`INSERT INTO users (%s) VALUES (%s);`,
		strings.Join(fields, ","), placeHolder)

	_, err := h.bobDBConn.Exec(ctx, insertStatement, database.GetScanFields(u, fields)...)
	if err != nil {
		return err
	}

	ug := &bentities.UserGroup{
		UserID:    database.Text(admin.ID),
		GroupID:   database.Text(constant.UserGroupAdmin),
		Status:    database.Text(bentities.UserGroupStatusActive),
		IsOrigin:  database.Bool(true),
		CreatedAt: database.Timestamptz(time.Now()),
		UpdatedAt: database.Timestamptz(time.Now()),
	}

	fields = []string{"user_id", "group_id", "status", "is_origin", "updated_at", "created_at"}
	placeHolder = database.GeneratePlaceholders(len(fields))
	insertStatement = fmt.Sprintf("INSERT INTO users_groups (%s) VALUES (%s);", strings.Join(fields, ","), placeHolder)
	_, err = h.bobDBConn.Exec(ctx, insertStatement, database.GetScanFields(ug, fields)...)
	return err
}

func (h *CommunicationHelper) LoginLeanerApp(email, password string) (string, error) {
	token, err := h.LoginWithFirebase(email, password)
	if err != nil {
		return "", err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	res, err := bpb.NewUserModifierServiceClient(h.bobGRPCConn).ExchangeToken(
		ctx, &bpb.ExchangeTokenRequest{Token: token})
	if err != nil {
		return "", err
	}
	return res.GetToken(), nil
}
