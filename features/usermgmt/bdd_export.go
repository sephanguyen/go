package usermgmt

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/manabie-com/backend/features/common"
	"github.com/manabie-com/backend/internal/golibs"
	internal_auth_tenant "github.com/manabie-com/backend/internal/golibs/auth/multitenant"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/try"
	location_repo "github.com/manabie-com/backend/internal/mastermgmt/modules/location/infrastructure/repo"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/service"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/multierr"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
)

type RoleWithLocation struct {
	RoleName    string
	LocationIDs []string
}

func SignIn(ctx context.Context, db database.QueryExecer, authDB *pgxpool.Pool, shamirConn *grpc.ClientConn, jwtApplicant string, firebaseAddr string, userMgmtConn *grpc.ClientConn, roleWithLocation RoleWithLocation, locationIDs []string) (common.AuthInfo, error) {
	var (
		userID    string
		userGroup string
	)
	switch roleWithLocation.RoleName {
	case constant.RoleSchoolAdmin,
		constant.RoleHQStaff,
		constant.RoleCentreLead,
		constant.RoleCentreManager,
		constant.RoleCentreStaff:
		resp, err := CreateStaff(ctx, db, userMgmtConn, nil, []RoleWithLocation{roleWithLocation}, locationIDs)
		if err != nil {
			return common.AuthInfo{}, err
		}
		userID = resp.Staff.StaffId
		userGroup = constant.UserGroupSchoolAdmin

	case constant.RoleTeacher, constant.RoleTeacherLead:
		resp, err := CreateStaff(ctx, db, userMgmtConn, nil, []RoleWithLocation{roleWithLocation}, locationIDs)
		if err != nil {
			return common.AuthInfo{}, err
		}
		userID = resp.Staff.StaffId
		userGroup = constant.UserGroupTeacher

	case constant.RoleStudent:
		resp, err := CreateStudent(ctx, userMgmtConn, nil, locationIDs)
		if err != nil {
			return common.AuthInfo{}, err
		}

		userID = resp.StudentProfile.Student.UserProfile.UserId
		userGroup = constant.UserGroupStudent

	case constant.RoleParent:
		resp, err := CreateParent(ctx, userMgmtConn, nil, "")
		if err != nil {
			return common.AuthInfo{}, err
		}

		userID = resp.ParentProfiles[0].Parent.UserProfile.UserId
		userGroup = constant.UserGroupTeacher
	}

	idToken, err := GenFakeIDToken(firebaseAddr, userID, "templates/"+userGroup+".template")
	if err != nil {
		return common.AuthInfo{}, err
	}
	exchangedToken, err := ExchangeToken(ctx, shamirConn, jwtApplicant, userID, idToken, NewAuthUserListener(ctx, authDB))
	if err != nil {
		return common.AuthInfo{}, err
	}

	return common.AuthInfo{UserID: userID, Token: exchangedToken}, nil
}

/*
CreateStaff:
- req: can be empty, if req empty, will use default value to create staff
- roleWithLocations: granted role for staff (incase req empty)
- locationIDs: location for that staff belong to (incase req empty)
*/
func CreateStaff(ctx context.Context, db database.QueryExecer, userConnection *grpc.ClientConn, req *upb.CreateStaffRequest, roleWithLocations []RoleWithLocation, locationIDs []string) (*upb.CreateStaffResponse, error) {
	if req == nil {
		createUserGroupReq, err := createUserGroupRequest(ctx, db, roleWithLocations)
		if err != nil {
			return nil, err
		}

		createUserGroupResp, err := createUserGroup(ctx, userConnection, createUserGroupReq)
		if err != nil {
			return nil, err
		}
		req = createStaffReq([]string{createUserGroupResp.UserGroupId}, locationIDs)
	}

	response, err := upb.NewStaffServiceClient(userConnection).CreateStaff(ctx, req)
	if err != nil {
		return nil, err
	}

	return response, nil
}

/*
CreateUserGroup:
- req: can be empty, if req empty, will use default value to create user_group
- roleWithLocations: granted permissions for user_group (incase req empty)
*/
func CreateUserGroup(ctx context.Context, db database.QueryExecer, userConnection *grpc.ClientConn, req *upb.CreateUserGroupRequest, roleWithLocations []RoleWithLocation) (*upb.CreateUserGroupResponse, error) {
	if req == nil {
		var err error
		req, err = createUserGroupRequest(ctx, db, roleWithLocations)
		if err != nil {
			return nil, err
		}
	}

	resp, err := createUserGroup(ctx, userConnection, req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

/*
Create Student:
- req: can be empty, if req empty, will use default value to create student
- locationIDs: location for that student belong to (incase req empty)
*/
func CreateStudent(ctx context.Context, userConnection *grpc.ClientConn, req *upb.CreateStudentRequest, locationIDs []string) (*upb.CreateStudentResponse, error) {
	if req == nil {
		req = createStudentReq(locationIDs)
	}

	response, err := upb.NewUserModifierServiceClient(userConnection).CreateStudent(ctx, req)
	if err != nil {
		return nil, err
	}

	return response, nil
}

/*
Create parent:
  - req: can be empty.
    if req empty:
  - if studentID is not empty, create parent with the studentID
  - if studentID is empty, create parent with random studentID (newly created)
*/
func CreateParent(ctx context.Context, userConnection *grpc.ClientConn, req *upb.CreateParentsAndAssignToStudentRequest, studentID string) (*upb.CreateParentsAndAssignToStudentResponse, error) {
	if req == nil {
		if studentID == "" {
			resourcePath, err := strconv.ParseInt(golibs.ResourcePathFromCtx(ctx), 10, 32)
			if err != nil {
				return nil, fmt.Errorf("resource path is invalid %w", err)
			}
			student, err := CreateStudent(ctx, userConnection, nil, []string{GetOrgLocation(int(resourcePath))})
			if err != nil {
				return nil, err
			}
			req = createParentReq(student.StudentProfile.Student.UserProfile.UserId)
		} else {
			req = createParentReq(studentID)
		}
	}

	resp, err := upb.NewUserModifierServiceClient(userConnection).CreateParentsAndAssignToStudent(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func GetOrgLocation(orgID int) string {
	switch orgID {
	case constants.ManabieSchool:
		return constants.ManabieOrgLocation
	case constants.JPREPSchool:
		return constants.JPREPOrgLocation
	case constants.TestingSchool:
		return constants.E2EOrgLocation
	case constants.KECDemo:
		return constants.KECDemoOrgLocation
	case constants.ManagaraBase:
		return constants.ManagaraBaseOrgLocation
	case constants.ManagaraHighSchool:
		return constants.ManagaraHSOrgLocation
	default:
		return ""
	}
}

func GetAdminID(orgID int) string {
	return mapOrgAndAdminID[orgID]
}

/*
GenerateAuthToken: generate exchanged token for user to call API
*/

func GenerateAuthToken(ctx context.Context, db *pgxpool.Pool, shamirConn *grpc.ClientConn, tenantManager internal_auth_tenant.TenantManager, jwtApplicant string, apiKey string, orgID int32, userID string) (string, error) {
	organization := &Organization{
		organizationID: strconv.Itoa(int(orgID)),
		schoolID:       orgID,
	}
	user := &User{userID: userID}

	tenantID, err := new(repository.OrganizationRepo).WithDefaultValue("local").GetTenantIDByOrgID(ctx, db, organization.OrganizationID().String())
	if err != nil {
		return "", err
	}

	backwardCompatibleAuthUser := &entity.LegacyUser{
		ID:    database.Text(user.UserID().String()),
		Email: database.Text(user.Email().String()),
		UserAdditionalInfo: entity.UserAdditionalInfo{
			Password: user.Password().String(),
		},
	}
	err = service.CreateUsersInIdentityPlatform(ctx, tenantManager, tenantID, entity.LegacyUsers{backwardCompatibleAuthUser}, int64(organization.SchoolID().Int32()))
	if err != nil {
		return "", err
	}

	identityPlatformLoginResult, err := LoginInAuthPlatform(ctx, apiKey, tenantID, user.Email().String(), user.Password().String())
	if err != nil {
		return "", err
	}

	exchangedToken, err := ExchangeToken(ctx, shamirConn, jwtApplicant, user.UserID().String(), identityPlatformLoginResult.IDToken)
	if err != nil {
		return "", err
	}

	return exchangedToken, nil
}

func GenerateFakeAuthInfo(ctx context.Context, shamirConn *grpc.ClientConn, firebaseAddr, jwtApplicant, userID, userGroup string) (common.AuthInfo, error) {
	idToken, err := GenFakeIDToken(firebaseAddr, userID, "templates/"+userGroup+".template")
	if err != nil {
		return common.AuthInfo{}, err
	}
	exchangedToken, err := ExchangeToken(ctx, shamirConn, jwtApplicant, userID, idToken)
	if err != nil {
		return common.AuthInfo{}, err
	}

	return common.AuthInfo{UserID: userID, Token: exchangedToken}, nil
}

// init credential for admin (bdd user): return exchanged token
func InitRootAccount(ctx context.Context, shamirConn *grpc.ClientConn, firebaseAddr, jwtApplicant string) (map[int]common.AuthInfo, error) {
	err := try.Do(func(attempt int) (bool, error) {
		if shamirConn.GetState() == connectivity.Ready {
			return false, nil
		}

		if attempt < retryTimes {
			time.Sleep(time.Second)
			return true, fmt.Errorf("the shamir service is not READY")
		}

		return false, fmt.Errorf("the shamir service is not READY")
	})
	if err != nil {
		return nil, err
	}
	mapOrgDefaultAdmin := make(map[int]common.AuthInfo)
	for orgID, userID := range mapOrgAndAdminID {
		authInfo, err := GenerateFakeAuthInfo(ctx, shamirConn, firebaseAddr, jwtApplicant, userID, constant.UserGroupSchoolAdmin)
		if err != nil {
			return nil, err
		}
		mapOrgDefaultAdmin[orgID] = authInfo
	}

	return mapOrgDefaultAdmin, nil
}

// seed locations
func PrepareLocations(db database.QueryExecer) ([]*location_repo.Location, error) {
	locationManabie1 := &location_repo.Location{}
	locationManabie2 := &location_repo.Location{}
	locationJPREP := &location_repo.Location{}
	locationE2E := &location_repo.Location{}

	database.AllNullEntity(locationManabie1)
	database.AllNullEntity(locationManabie2)
	database.AllNullEntity(locationJPREP)
	database.AllNullEntity(locationE2E)

	uidLocationManabie1 := newID()
	uidLocationJPREP := newID()
	uidLocationManabie2 := newID()
	uidLocationE2E := newID()

	now := time.Now()
	if err := multierr.Combine(
		locationManabie1.LocationID.Set(uidLocationManabie1),
		locationManabie1.PartnerInternalID.Set(uidLocationManabie1),
		locationManabie1.Name.Set(fmt.Sprintf(stubLocationName, uidLocationManabie1)),
		locationManabie1.CreatedAt.Set(now),
		locationManabie1.UpdatedAt.Set(now),
		locationManabie1.ResourcePath.Set(fmt.Sprint(constants.ManabieSchool)),
		locationManabie1.IsArchived.Set(false),
		locationManabie1.AccessPath.Set(fmt.Sprintf("%s/%s", constants.ManabieOrgLocation, uidLocationManabie1)),
		locationManabie1.ParentLocationID.Set(constants.ManabieOrgLocation),

		locationJPREP.LocationID.Set(uidLocationJPREP),
		locationJPREP.Name.Set(fmt.Sprintf(stubLocationName, uidLocationJPREP)),
		locationJPREP.CreatedAt.Set(now),
		locationJPREP.UpdatedAt.Set(now),
		locationJPREP.ResourcePath.Set(fmt.Sprint(constants.JPREPSchool)),
		locationJPREP.IsArchived.Set(false),
		locationJPREP.AccessPath.Set(fmt.Sprintf("%s/%s", constants.JPREPOrgLocation, uidLocationJPREP)),
		locationJPREP.ParentLocationID.Set(constants.JPREPOrgLocation),

		locationManabie2.LocationID.Set(uidLocationManabie2),
		locationManabie2.Name.Set(fmt.Sprintf(stubLocationName, uidLocationManabie2)),
		locationManabie2.PartnerInternalID.Set(uidLocationManabie2),
		locationManabie2.CreatedAt.Set(now),
		locationManabie2.UpdatedAt.Set(now),
		locationManabie2.ResourcePath.Set(fmt.Sprint(constants.ManabieSchool)),
		locationManabie2.IsArchived.Set(false),
		locationManabie2.AccessPath.Set(fmt.Sprintf("%s/%s", constants.ManabieOrgLocation, uidLocationManabie2)),
		locationManabie2.ParentLocationID.Set(constants.ManabieOrgLocation),

		locationE2E.LocationID.Set(uidLocationE2E),
		locationE2E.Name.Set(fmt.Sprintf(stubLocationName, uidLocationE2E)),
		locationE2E.CreatedAt.Set(now),
		locationE2E.UpdatedAt.Set(now),
		locationE2E.ResourcePath.Set(fmt.Sprint(constants.TestingSchool)),
		locationE2E.IsArchived.Set(false),
		locationE2E.AccessPath.Set(fmt.Sprintf("%s/%s", constants.E2EOrgLocation, uidLocationE2E)),
		locationE2E.ParentLocationID.Set(constants.E2EOrgLocation),
	); err != nil {
		return nil, err
	}

	ctx := context.Background()
	manabieOrgID := constants.ManabieSchool
	ctx = common.ValidContext(ctx, manabieOrgID, rootAccount[manabieOrgID].UserID, rootAccount[manabieOrgID].Token)
	if err := insertLocations(ctx, []*location_repo.Location{locationManabie1, locationManabie2}, db); err != nil {
		return nil, err
	}

	jprepOrgID := constants.JPREPSchool
	ctx = common.ValidContext(ctx, jprepOrgID, rootAccount[jprepOrgID].UserID, rootAccount[jprepOrgID].Token)
	if err := insertLocations(ctx, []*location_repo.Location{locationJPREP}, db); err != nil {
		return nil, err
	}

	e2eOrgID := constants.TestingSchool
	ctx = common.ValidContext(ctx, e2eOrgID, rootAccount[e2eOrgID].UserID, rootAccount[e2eOrgID].Token)
	if err := insertLocations(ctx, []*location_repo.Location{locationE2E}, db); err != nil {
		return nil, err
	}

	return []*location_repo.Location{locationManabie1, locationJPREP, locationManabie2, locationE2E}, nil
}
