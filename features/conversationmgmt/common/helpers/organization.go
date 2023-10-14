package helpers

import (
	"context"
	"fmt"
	"strconv"

	"github.com/manabie-com/backend/features/conversationmgmt/common/entities"
	bob_entities "github.com/manabie-com/backend/internal/bob/entities"
	bob_repository "github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
)

func (helper *ConversationMgmtHelper) genarateSchoolNameCountryCityDistrict(arg1, arg2, arg3, arg4 string) *bob_entities.School {
	city := &bob_entities.City{
		Name:    database.Text("city" + arg3),
		Country: database.Text("country" + arg2),
	}
	district := &bob_entities.District{
		Name:    database.Text("district" + arg4),
		Country: database.Text("country" + arg2),
		City:    city,
	}
	school := &bob_entities.School{
		Name:           database.Text("school-" + arg1),
		Country:        database.Text("country" + arg2),
		City:           city,
		District:       district,
		IsSystemSchool: pgtype.Bool{Bool: true, Status: pgtype.Present},
		Point:          pgtype.Point{Status: pgtype.Null},
	}
	return school
}

func (helper *ConversationMgmtHelper) insertsSchoolsIntoDB(ctx context.Context, schools []*bob_entities.School) (int32, error) {
	r := &bob_repository.SchoolRepo{}
	if err := r.Import(ctx, helper.BobDBConn, schools); err != nil {
		return 0, err
	}
	currentSchooldID := schools[len(schools)-1].ID.Int
	return currentSchooldID, nil
}

func (helper *ConversationMgmtHelper) generateOrganizationRole(ctx context.Context, defaultLocationID string, resourcePath string) error {
	ctx2 := interceptors.ContextWithJWTClaims(ctx, &interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{
			ResourcePath: resourcePath,
		},
	})

	mapRoleAndRoleID := make(map[string]string)
	for _, role := range OrganizationRoles {
		mapRoleAndRoleID[role] = idutil.ULIDNow()
	}

	_, err := helper.BobDBConn.Exec(ctx2, `
		INSERT INTO role 
			(role_id, role_name, created_at, updated_at, resource_path, is_system)
		VALUES 
			($1, 'Teacher', now(), now(), autofillresourcepath(), false),
			($2, 'School Admin', now(), now(), autofillresourcepath(), false),
			($3, 'Student', now(), now(), autofillresourcepath(), false),
			($4, 'Parent', now(), now(), autofillresourcepath(), false),
			($5, 'HQ Staff', now(), now(), autofillresourcepath(), false),
			($6, 'Centre Manager', now(), now(), autofillresourcepath(), false),
			($7, 'Centre Staff', now(), now(), autofillresourcepath(), false),
			($8, 'Centre Lead', now(), now(), autofillresourcepath(), false),
			($9, 'Teacher Lead', now(), now(), autofillresourcepath(), false);
	`, mapRoleAndRoleID[constant.RoleTeacher],
		mapRoleAndRoleID[constant.RoleSchoolAdmin],
		mapRoleAndRoleID[constant.RoleStudent],
		mapRoleAndRoleID[constant.RoleParent],
		mapRoleAndRoleID[constant.RoleHQStaff],
		mapRoleAndRoleID[constant.RoleCentreManager],
		mapRoleAndRoleID[constant.RoleCentreStaff],
		mapRoleAndRoleID[constant.RoleCentreLead],
		mapRoleAndRoleID[constant.RoleTeacherLead])
	if err != nil {
		return fmt.Errorf("err insert role: %v", err)
	}

	// System user group, granted role and granted access path for student/parent
	studentUgID := idutil.ULIDNow()
	parentUgID := idutil.ULIDNow()
	_, err = helper.BobDBConn.Exec(ctx2, `
		INSERT INTO public.user_group
			(user_group_id, user_group_name, created_at, updated_at, deleted_at, resource_path, org_location_id, is_system)
		VALUES
			($1, 'Student', now(), now(), NULL, autofillresourcepath(), $3, true),
			($2, 'Parent', now(), now(), NULL, autofillresourcepath(), $3, true);
	`, studentUgID, parentUgID, defaultLocationID)
	if err != nil {
		return fmt.Errorf("err insert user_group: %v", err)
	}

	studentGrID := idutil.ULIDNow()
	parentGrID := idutil.ULIDNow()
	_, err = helper.BobDBConn.Exec(ctx2, `
		INSERT INTO public.granted_role
			(granted_role_id, user_group_id, role_id, created_at, updated_at, deleted_at, resource_path)
		VALUES
			($1, $2, $3, now(), now(), NULL, autofillresourcepath()),
			($4, $5, $6, now(), now(), NULL, autofillresourcepath());
	`, studentGrID, studentUgID, mapRoleAndRoleID[constant.RoleStudent], parentGrID, parentUgID, mapRoleAndRoleID[constant.RoleParent])
	if err != nil {
		return fmt.Errorf("err insert granted_role: %v", err)
	}

	// Assign org access path for student/parent role
	_, err = helper.BobDBConn.Exec(ctx2, `
		INSERT INTO public.granted_role_access_path
			(granted_role_id, location_id, created_at, updated_at, deleted_at, resource_path)
		VALUES
			($1, $3, now(), now(), NULL, autofillresourcepath()),
			($2, $3, now(), now(), NULL, autofillresourcepath());
	`, studentGrID, parentGrID, defaultLocationID)
	if err != nil {
		return fmt.Errorf("err insert granted_role_access_path: %v", err)
	}

	err = multierr.Combine(
		helper.insertPermissionRoleForNotificationEnt(ctx, mapRoleAndRoleID, resourcePath),
		helper.insertPermissionRoleForLocationEnt(ctx, mapRoleAndRoleID, resourcePath),
		helper.insertPermissionRoleForStudentEnt(ctx, mapRoleAndRoleID, resourcePath),
		helper.insertPermissionRoleForParentEnt(ctx, mapRoleAndRoleID, resourcePath),
		helper.insertPermissionRoleForStaffEnt(ctx, mapRoleAndRoleID, resourcePath),
		helper.insertPermissionRoleForUserGroupEnt(ctx, mapRoleAndRoleID, resourcePath),
		helper.insertPermissionRoleForUserEnt(ctx, mapRoleAndRoleID, resourcePath),
		helper.insertPermissionRoleForCourseEnt(ctx, mapRoleAndRoleID, resourcePath),
	)
	if err != nil {
		return fmt.Errorf("err insert permission_role: %v", err)
	}

	return nil
}

func (helper *ConversationMgmtHelper) generateOrganizationAuth(ctx context.Context, schoolID string) error {
	stmt := `
	INSERT INTO organization_auths
		(organization_id, auth_project_id, auth_tenant_id)
	VALUES
		($1, 'fake_aud', ''),
		($2, 'dev-manabie-online', ''),
		($2, 'dev-manabie-online', 'integration-test-1-909wx')
	ON CONFLICT 
		DO NOTHING
	;
	`
	_, err := helper.BobDBConn.Exec(ctx, stmt, schoolID, schoolID)
	return err
}

func (helper *ConversationMgmtHelper) newOrgWithOrgLocation(ctx context.Context) (int32, *entities.Location, error) {
	random := idutil.ULIDNow()
	school := helper.genarateSchoolNameCountryCityDistrict(random, random, random, random)

	schoolID, err := helper.insertsSchoolsIntoDB(ctx, []*bob_entities.School{school})
	if err != nil {
		return 0, nil, err
	}

	orgText := strconv.Itoa(int(schoolID))
	resourcePath := strconv.Itoa(int(schoolID))
	orgID := schoolID

	_, err = helper.BobDBConn.Exec(ctx, `
		INSERT INTO organizations(
			organization_id, tenant_id, name, resource_path)
		VALUES ($1, $2, $3, $4)
	`, orgText, orgText, orgText, orgText)
	if err != nil {
		return 0, nil, fmt.Errorf("create default location: %v", err)
	}

	// default location
	defaultLocation, err := helper.CreateLocationWithDB(ctx, resourcePath, OrganizationLocationTypeName, "", "")
	if err != nil {
		return 0, nil, err
	}

	err = helper.generateOrganizationAuth(ctx, resourcePath)
	if err != nil {
		return 0, nil, fmt.Errorf("s.CommonSuite.GenerateOrganizationAuth %s", err)
	}

	err = helper.generateOrganizationRole(ctx, defaultLocation.ID, resourcePath)
	if err != nil {
		return 0, nil, fmt.Errorf("s.CommonSuite.GenerateOrganizationRole %s", err)
	}

	return orgID, defaultLocation, err
}

func (helper *ConversationMgmtHelper) CreateNewOrganization(orgType string) (*entities.Organization, error) {
	var (
		defaultLocation     *entities.Location
		descendantLocations = make([]*entities.Location, 0)
		orgID               int32
	)

	switch orgType {
	case OrganizationTypeNew:
		newOrgID, orgDefaultLocation, err := helper.newOrgWithOrgLocation(context.Background())
		if err != nil {
			return nil, err
		}
		orgID = newOrgID
		defaultLocation = orgDefaultLocation

	case OrganizationTypeJPREP:
		orgID = constants.JPREPSchool
		defaultLocation = &entities.Location{
			ID:               JPREPOrgLocation,
			Name:             "JPREP",
			AccessPath:       JPREPOrgLocation,
			ParentLocationID: "",
			TypeLocation:     "org",
			TypeLocationID:   JPREPOrgLocationType,
		}

	default:
		orgID = constants.ManabieSchool
		defaultLocation = &entities.Location{
			ID:               ManabieOrgLocation,
			Name:             "Manabie",
			AccessPath:       ManabieOrgLocation,
			ParentLocationID: "",
			TypeLocation:     "org",
			TypeLocationID:   ManabieOrgLocationType,
		}
	}

	// create some centers (descendants of default location - organization location)
	for idxLoc := 0; idxLoc < NumberOfNewCenterLocationCreated; idxLoc++ {
		descendantLocation, err := helper.CreateLocationWithDB(context.Background(), fmt.Sprint(orgID), CenterLocationTypeName, defaultLocation.ID, defaultLocation.TypeLocationID)
		if err != nil {
			return nil, fmt.Errorf("create descendant locations: %v", err)
		}
		descendantLocations = append(descendantLocations, descendantLocation)
	}

	return &entities.Organization{
		ID:                  orgID,
		DefaultLocation:     defaultLocation,
		DescendantLocations: descendantLocations,
	}, nil
}
