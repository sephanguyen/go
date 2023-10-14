package common

import (
	"context"
	"fmt"
	"strconv"

	bob_entities "github.com/manabie-com/backend/internal/bob/entities"
	bob_repository "github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
)

var (
	OrganizationRoles = []string{
		constant.RoleTeacher,
		constant.RoleSchoolAdmin,
		constant.RoleStudent,
		constant.RoleParent,
		constant.RoleHQStaff,
		constant.RoleCentreManager,
		constant.RoleCentreStaff,
		constant.RoleCentreLead,
		constant.RoleTeacherLead,
	}
)

func (s *suite) ASchoolNameCountryCityDistrict(ctx context.Context, arg1, arg2, arg3, arg4 string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	city := &bob_entities.City{
		Name:    database.Text(arg3),
		Country: database.Text(arg2),
	}
	district := &bob_entities.District{
		Name:    database.Text(arg4),
		Country: database.Text(arg2),
		City:    city,
	}
	school := &bob_entities.School{
		Name:           database.Text(arg1 + stepState.Random),
		Country:        database.Text(arg2),
		City:           city,
		District:       district,
		IsSystemSchool: pgtype.Bool{Bool: true, Status: pgtype.Present},
		Point:          pgtype.Point{Status: pgtype.Null},
	}
	stepState.Schools = append(stepState.Schools, school)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) AdminInsertsSchools(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	r := &bob_repository.SchoolRepo{}
	if err := r.Import(ctx, s.BobDB, stepState.Schools); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.CurrentSchoolID = stepState.Schools[len(stepState.Schools)-1].ID.Int
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) NewOrgWithOrgLocation(ctx context.Context) (int32, string, string, error) {
	random := idutil.ULIDNow()
	ctx, err := s.ASchoolNameCountryCityDistrict(ctx, random, random, random, random)
	if err != nil {
		return 0, "", "", err
	}
	ctx, err = s.AdminInsertsSchools(ctx)
	if err != nil {
		return 0, "", "", err
	}

	state := StepStateFromContext(ctx)
	school := state.CurrentSchoolID
	schoolText := strconv.Itoa(int(school))
	resourcePath := strconv.Itoa(int(school))

	_, err = s.BobDBTrace.Exec(ctx, `INSERT INTO organizations(
	organization_id, tenant_id, name, resource_path)
	VALUES ($1, $2, $3, $4)`, schoolText, schoolText, schoolText, schoolText)
	if err != nil {
		return 0, "", "", err
	}
	err = s.GenerateOrganizationAuth(ctx, resourcePath)
	if err != nil {
		return 0, "", "", fmt.Errorf("s.CommonSuite.GenerateOrganizationAuth %s", err)
	}

	// default location
	locationID, locationTypeID, err := s.CreateLocationWithDB(ctx, resourcePath, "org", "", "")
	if err != nil {
		return 0, "", "", err
	}

	err = s.GenerateOrganizationRoleAndPermission(ctx, locationID, resourcePath)
	if err != nil {
		return 0, "", "", fmt.Errorf("s.CommonSuite.GenerateOrganizationRoleAndPermission %s", err)
	}

	return state.CurrentSchoolID, locationID, locationTypeID, err
}

func (s *suite) GenerateOrganizationAuth(ctx context.Context, schoolID string) error {
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
	_, err := s.BobDB.Exec(ctx, stmt, schoolID, schoolID)
	return err
}

func (s *suite) ThisSchoolHasConfigIsIsIs(ctx context.Context, planField, planValue, expiredAtField, expiredAtValue, durationField string, durationValue int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	var expiredAt interface{} = expiredAtValue
	if expiredAtValue == "NULL" {
		var pgExpiredAtValue pgtype.Timestamptz
		_ = pgExpiredAtValue.Set(nil)
		expiredAt = &pgExpiredAtValue
	}

	_, err := s.BobDB.Exec(ctx, `INSERT INTO school_configs VALUES
	($1, $2, 'COUNTRY_VN', $3, $4, now(), now())
	ON CONFLICT  ON CONSTRAINT school_configs_pk 
	DO UPDATE SET plan_id = $2, plan_expired_at = $3, plan_duration = $4;`, stepState.CurrentSchoolID, planValue, expiredAt, durationValue)

	return StepStateToContext(ctx, stepState), err
}

func (s *suite) GenerateOrganizationRoleAndPermission(ctx context.Context, defaultLocationID string, resourcePath string) error {
	ctx2 := interceptors.ContextWithJWTClaims(ctx, &interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{
			ResourcePath: resourcePath,
		},
	})

	mapRoleAndRoleID := make(map[string]string)
	for _, role := range OrganizationRoles {
		mapRoleAndRoleID[role] = idutil.ULIDNow()
	}

	_, err := s.BobDB.Exec(ctx2, `
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
	_, err = s.BobDB.Exec(ctx2, `
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
	_, err = s.BobDB.Exec(ctx2, `
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
	_, err = s.BobDB.Exec(ctx2, `
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
		s.insertPermissionRoleForNotificationEnt(ctx, mapRoleAndRoleID, resourcePath),
		s.insertPermissionRoleForLocationEnt(ctx, mapRoleAndRoleID, resourcePath),
		s.insertPermissionRoleForStudentEnt(ctx, mapRoleAndRoleID, resourcePath),
		s.insertPermissionRoleForParentEnt(ctx, mapRoleAndRoleID, resourcePath),
		s.insertPermissionRoleForStaffEnt(ctx, mapRoleAndRoleID, resourcePath),
		s.insertPermissionRoleForUserGroupEnt(ctx, mapRoleAndRoleID, resourcePath),
		s.insertPermissionRoleForUserEnt(ctx, mapRoleAndRoleID, resourcePath),
		s.insertPermissionRoleForCourseEnt(ctx, mapRoleAndRoleID, resourcePath),
		s.insertPermissionRoleForLessonEnt(ctx, mapRoleAndRoleID, resourcePath),
	)
	if err != nil {
		return fmt.Errorf("err insert permission_role: %v", err)
	}

	return nil
}
