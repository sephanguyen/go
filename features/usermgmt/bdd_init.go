package usermgmt

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/features/common"
	"github.com/manabie-com/backend/internal/golibs"
	internal_auth_tenant "github.com/manabie-com/backend/internal/golibs/auth/multitenant"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var mapOrgAndAdminID = map[int]string{
	constants.ManabieSchool:      "bdd_admin-manabie",
	constants.JPREPSchool:        "bdd_admin-jprep",
	constants.TestingSchool:      "bdd_admin-e2e",
	constants.ManagaraBase:       "bdd_admin-managara-base",
	constants.ManagaraHighSchool: "bdd_admin-managara-hs",
	constants.KECDemo:            "bdd_admin-kec-demo",
}

func InitOrganizationTenantConfig(ctx context.Context, db database.QueryExecer) error {
	stmt :=
		`
		INSERT INTO
			organizations(organization_id, tenant_id, resource_path, created_at, scrypt_signer_key, scrypt_salt_separator, scrypt_rounds, scrypt_memory_cost)
		VALUES
			(
				'-2147483648',
				'manabie-0nl6t',
				'-2147483648',
				now(),
				encode(
						encrypt_iv(
								'mAaX5DSYQLUj3XD60McZ3n6m/AdZxEpfiLYqIFtYf2jlNIVaJ6Esu1sWe5HrsyLO1sTD/pygrtoFsQaFhfuRDg==',
								decode('W4qOy896DmWHg22orCQc2NEM9vQuIVvNuj+TwJDV8J0=', 'base64'),
								decode('2/Ukd9Ue2ci6uRB5g3qPSA==', 'base64'), 'aes-cbc/pad:pkcs'),
						'base64'
					),
				encode(
						encrypt_iv(
								'Bw==',
								decode('W4qOy896DmWHg22orCQc2NEM9vQuIVvNuj+TwJDV8J0=', 'base64'),
								decode('2/Ukd9Ue2ci6uRB5g3qPSA==', 'base64'), 'aes-cbc/pad:pkcs'),
						'base64'
					),
				encode(
						encrypt_iv(
								'8',
								decode('W4qOy896DmWHg22orCQc2NEM9vQuIVvNuj+TwJDV8J0=', 'base64'),
								decode('2/Ukd9Ue2ci6uRB5g3qPSA==', 'base64'), 'aes-cbc/pad:pkcs'),
						'base64'
					),
				encode(
						encrypt_iv(
								'14',
								decode('W4qOy896DmWHg22orCQc2NEM9vQuIVvNuj+TwJDV8J0=', 'base64'),
								decode('2/Ukd9Ue2ci6uRB5g3qPSA==', 'base64'), 'aes-cbc/pad:pkcs'),
						'base64'
					)
			),
			(
				'-2147483647',
				'jprep-eznr7',
				'-2147483647',
				now(),
				encode(
						encrypt_iv(
								'T4IYjTo1Wwfom7XT6Lvm729rloqaPu05xQp+JtGhKc8ypMbWVoovi/XrVPUL9jQFs8pZDHYGbWJb9DBs9LU/Ew==',
								decode('W4qOy896DmWHg22orCQc2NEM9vQuIVvNuj+TwJDV8J0=', 'base64'),
								decode('2/Ukd9Ue2ci6uRB5g3qPSA==', 'base64'), 'aes-cbc/pad:pkcs'),
						'base64'
					),
				encode(
						encrypt_iv(
								'Bw==',
								decode('W4qOy896DmWHg22orCQc2NEM9vQuIVvNuj+TwJDV8J0=', 'base64'),
								decode('2/Ukd9Ue2ci6uRB5g3qPSA==', 'base64'), 'aes-cbc/pad:pkcs'),
						'base64'
					),
				encode(
						encrypt_iv(
								'8',
								decode('W4qOy896DmWHg22orCQc2NEM9vQuIVvNuj+TwJDV8J0=', 'base64'),
								decode('2/Ukd9Ue2ci6uRB5g3qPSA==', 'base64'), 'aes-cbc/pad:pkcs'),
						'base64'
					),
				encode(
						encrypt_iv(
								'14',
								decode('W4qOy896DmWHg22orCQc2NEM9vQuIVvNuj+TwJDV8J0=', 'base64'),
								decode('2/Ukd9Ue2ci6uRB5g3qPSA==', 'base64'), 'aes-cbc/pad:pkcs'),
						'base64'
					)
			),
			(
				'-2147483646',
				'synersia-24rue',
				'-2147483646',
				now(),
				encode(
						encrypt_iv(
								'A2jT33D3K0PtdduzUwtgtQN+g41KIKHGpqFxonQyFiWEePUKAQwV/ngeyFzGIBQmfm48DDZmqds+s30KQQG8Lg==',
								decode('W4qOy896DmWHg22orCQc2NEM9vQuIVvNuj+TwJDV8J0=', 'base64'),
								decode('2/Ukd9Ue2ci6uRB5g3qPSA==', 'base64'), 'aes-cbc/pad:pkcs'),
						'base64'
					),
				encode(
						encrypt_iv(
								'Bw==',
								decode('W4qOy896DmWHg22orCQc2NEM9vQuIVvNuj+TwJDV8J0=', 'base64'),
								decode('2/Ukd9Ue2ci6uRB5g3qPSA==', 'base64'), 'aes-cbc/pad:pkcs'),
						'base64'
					),
				encode(
						encrypt_iv(
								'8',
								decode('W4qOy896DmWHg22orCQc2NEM9vQuIVvNuj+TwJDV8J0=', 'base64'),
								decode('2/Ukd9Ue2ci6uRB5g3qPSA==', 'base64'), 'aes-cbc/pad:pkcs'),
						'base64'
					),
				encode(
						encrypt_iv(
								'14',
								decode('W4qOy896DmWHg22orCQc2NEM9vQuIVvNuj+TwJDV8J0=', 'base64'),
								decode('2/Ukd9Ue2ci6uRB5g3qPSA==', 'base64'), 'aes-cbc/pad:pkcs'),
						'base64'
					)
			),
			(
				'-2147483645',
				'renseikai-yu9y7',
				'-2147483645',
				now(),
				encode(
						encrypt_iv(
								'O2XbUE0OcqAeT4ZZUH7k3CXE9Xjaom6AC2QukKYfLegzbnnfnElijWR8VZpproYbQfnzkaOSOHiIUolGzfnIdw==',
								decode('W4qOy896DmWHg22orCQc2NEM9vQuIVvNuj+TwJDV8J0=', 'base64'),
								decode('2/Ukd9Ue2ci6uRB5g3qPSA==', 'base64'), 'aes-cbc/pad:pkcs'),
						'base64'
					),
				encode(
						encrypt_iv(
								'Bw==',
								decode('W4qOy896DmWHg22orCQc2NEM9vQuIVvNuj+TwJDV8J0=', 'base64'),
								decode('2/Ukd9Ue2ci6uRB5g3qPSA==', 'base64'), 'aes-cbc/pad:pkcs'),
						'base64'
					),
				encode(
						encrypt_iv(
								'8',
								decode('W4qOy896DmWHg22orCQc2NEM9vQuIVvNuj+TwJDV8J0=', 'base64'),
								decode('2/Ukd9Ue2ci6uRB5g3qPSA==', 'base64'), 'aes-cbc/pad:pkcs'),
						'base64'
					),
				encode(
						encrypt_iv(
								'14',
								decode('W4qOy896DmWHg22orCQc2NEM9vQuIVvNuj+TwJDV8J0=', 'base64'),
								decode('2/Ukd9Ue2ci6uRB5g3qPSA==', 'base64'), 'aes-cbc/pad:pkcs'),
						'base64'
					)
			),
			(
				'-2147483644',
				'end-to-end-dopvo',
				'-2147483644',
				now(),
				encode(
						encrypt_iv(
								'HX/mMzhYG4IvVmnLbkuMMqbmCAGnGhVaSWNIKU5YuGhhdpMW+FUFkYK1o9Gmi02b+JxHpgaDGckHKfta+BSAWA==',
								decode('W4qOy896DmWHg22orCQc2NEM9vQuIVvNuj+TwJDV8J0=', 'base64'),
								decode('2/Ukd9Ue2ci6uRB5g3qPSA==', 'base64'), 'aes-cbc/pad:pkcs'),
						'base64'
					),
				encode(
						encrypt_iv(
								'Bw==',
								decode('W4qOy896DmWHg22orCQc2NEM9vQuIVvNuj+TwJDV8J0=', 'base64'),
								decode('2/Ukd9Ue2ci6uRB5g3qPSA==', 'base64'), 'aes-cbc/pad:pkcs'),
						'base64'
					),
				encode(
						encrypt_iv(
								'8',
								decode('W4qOy896DmWHg22orCQc2NEM9vQuIVvNuj+TwJDV8J0=', 'base64'),
								decode('2/Ukd9Ue2ci6uRB5g3qPSA==', 'base64'), 'aes-cbc/pad:pkcs'),
						'base64'
					),
				encode(
						encrypt_iv(
								'14',
								decode('W4qOy896DmWHg22orCQc2NEM9vQuIVvNuj+TwJDV8J0=', 'base64'),
								decode('2/Ukd9Ue2ci6uRB5g3qPSA==', 'base64'), 'aes-cbc/pad:pkcs'),
						'base64'
					)
			),
			(
				'99999999',
				'migrate-users-ge51c',
				'99999999',
				now(),
				encode(
						encrypt_iv(
								'ZCeanYs6TLcqSofj4tJDSOf9UiTZfZKGzVOXQf+1LxcBrGpc9q2HAoWA8VsCX4n7c7v8vJODrcdD2rHM3nZJVQ==',
								decode('W4qOy896DmWHg22orCQc2NEM9vQuIVvNuj+TwJDV8J0=', 'base64'),
								decode('2/Ukd9Ue2ci6uRB5g3qPSA==', 'base64'), 'aes-cbc/pad:pkcs'),
						'base64'
					),
				encode(
						encrypt_iv(
								'Bw==',
								decode('W4qOy896DmWHg22orCQc2NEM9vQuIVvNuj+TwJDV8J0=', 'base64'),
								decode('2/Ukd9Ue2ci6uRB5g3qPSA==', 'base64'), 'aes-cbc/pad:pkcs'),
						'base64'
					),
				encode(
						encrypt_iv(
								'8',
								decode('W4qOy896DmWHg22orCQc2NEM9vQuIVvNuj+TwJDV8J0=', 'base64'),
								decode('2/Ukd9Ue2ci6uRB5g3qPSA==', 'base64'), 'aes-cbc/pad:pkcs'),
						'base64'
					),
				encode(
						encrypt_iv(
								'14',
								decode('W4qOy896DmWHg22orCQc2NEM9vQuIVvNuj+TwJDV8J0=', 'base64'),
								decode('2/Ukd9Ue2ci6uRB5g3qPSA==', 'base64'), 'aes-cbc/pad:pkcs'),
						'base64'
					)
			)
		ON CONFLICT
			ON CONSTRAINT
				organizations__pk
			DO UPDATE
				SET
					tenant_id = excluded.tenant_id,
					scrypt_signer_key = excluded.scrypt_signer_key,
					scrypt_salt_separator = excluded.scrypt_salt_separator,
					scrypt_rounds = excluded.scrypt_rounds,
					scrypt_memory_cost = excluded.scrypt_memory_cost
		;
		`

	_, err := db.Exec(ctx, stmt)
	if err != nil {
		return errors.Wrap(err, "cannot init organization tenant config info: %v")
	}
	return nil
}

type Suite struct {
	*common.Connections
	*common.StepState
	ZapLogger   *zap.Logger
	Cfg         *common.Config
	CommonSuite *common.Suite
	ApplicantID string
}

func NewSuite(usermgmtConn *grpc.ClientConn, shamirConn *grpc.ClientConn, usermgmtdb *database.DBTrace, logger *zap.Logger, externalFirebaseAddr string, appID string) *Suite {
	userMgmtSuite := &Suite{}
	userMgmtSuite.Connections = &common.Connections{}
	userMgmtSuite.UserMgmtConn = usermgmtConn
	userMgmtSuite.ZapLogger = logger
	userMgmtSuite.ShamirConn = shamirConn
	userMgmtSuite.BobDBTrace = usermgmtdb
	userMgmtSuite.ApplicantID = appID
	firebaseAddr = externalFirebaseAddr
	return userMgmtSuite
}

func InitUser(ctx context.Context, db *pgxpool.Pool, usermgmtConnection *grpc.ClientConn, tenantManager internal_auth_tenant.TenantManager, jwtApplicant string, apiKey string) (map[int]common.MapRoleAndAuthInfo, error) {
	mapOrgUser, err := InitStaff(ctx, db, usermgmtConnection, tenantManager, jwtApplicant, apiKey)
	if err != nil {
		return nil, err
	}

	mapOrgStudent, err := InitStudent(ctx, db, usermgmtConnection, tenantManager, jwtApplicant, apiKey)
	if err != nil {
		return nil, err
	}
	for orgID, roleAndAuthInfo := range mapOrgStudent {
		for role, authInfo := range roleAndAuthInfo {
			mapOrgUser[orgID][role] = authInfo
		}
	}

	mapOrgParent, err := InitParent(ctx, db, usermgmtConnection, tenantManager, jwtApplicant, apiKey)
	if err != nil {
		return nil, err
	}
	for orgID, roleAndAuthInfo := range mapOrgParent {
		for role, authInfo := range roleAndAuthInfo {
			mapOrgUser[orgID][role] = authInfo
		}
	}

	return mapOrgUser, nil
}

/*
InitStaff credential:
- create list staffs with granted permission: school admin in org level, teacher in location level, etc
- return exchanged token
*/
func InitStaff(ctx context.Context, db *pgxpool.Pool, usermgmtConnection *grpc.ClientConn, tenantManager internal_auth_tenant.TenantManager, jwtApplicant string, apiKey string) (map[int]common.MapRoleAndAuthInfo, error) {
	mapOrgStaff := make(map[int]common.MapRoleAndAuthInfo)
	for orgID, grantedPermissions := range orgAndGrantedPermission() {
		validContext := common.ValidContext(ctx, orgID, rootAccount[orgID].UserID, rootAccount[orgID].Token)
		mapRoleAndAuthInfo := make(common.MapRoleAndAuthInfo)

		for _, roleWithLocation := range grantedPermissions {
			resp, err := CreateStaff(validContext, db, usermgmtConnection, nil, []RoleWithLocation{roleWithLocation}, roleWithLocation.LocationIDs)
			if err != nil {
				return nil, err
			}

			authInfo, err := GenerateFakeAuthInfo(ctx, connections.ShamirConn, firebaseAddr, jwtApplicant, resp.Staff.StaffId, constant.MapRoleWithLegacyUserGroup[roleWithLocation.RoleName])
			if err != nil {
				return nil, err
			}
			mapRoleAndAuthInfo[roleWithLocation.RoleName] = authInfo
		}
		mapOrgStaff[orgID] = mapRoleAndAuthInfo
	}

	return mapOrgStaff, nil
}

func InitStudent(ctx context.Context, db *pgxpool.Pool, usermgmtConnection *grpc.ClientConn, tenantManager internal_auth_tenant.TenantManager, jwtApplicant string, apiKey string) (map[int]common.MapRoleAndAuthInfo, error) {
	mapOrgStudent := make(map[int]common.MapRoleAndAuthInfo)

	for orgID := range mapOrgAndAdminID {
		if golibs.InArrayString(fmt.Sprint(orgID), []string{fmt.Sprint(constants.ManagaraHighSchool), fmt.Sprint(constants.ManagaraBase)}) {
			continue
		}
		validContext := common.ValidContext(ctx, orgID, rootAccount[orgID].UserID, rootAccount[orgID].Token)
		mapRoleAndAuthInfo := make(common.MapRoleAndAuthInfo)

		resp, err := CreateStudent(validContext, usermgmtConnection, nil, []string{GetOrgLocation(orgID)})
		if err != nil {
			return nil, err
		}

		studentID := resp.StudentProfile.Student.UserProfile.UserId
		authInfo, err := GenerateFakeAuthInfo(ctx, connections.ShamirConn, firebaseAddr, jwtApplicant, studentID, constant.UserGroupStudent)
		if err != nil {
			return nil, err
		}

		mapRoleAndAuthInfo[constant.RoleStudent] = authInfo
		mapOrgStudent[orgID] = mapRoleAndAuthInfo
	}

	return mapOrgStudent, nil
}

func InitParent(ctx context.Context, db *pgxpool.Pool, usermgmtConnection *grpc.ClientConn, tenantManager internal_auth_tenant.TenantManager, jwtApplicant string, apiKey string) (map[int]common.MapRoleAndAuthInfo, error) {
	mapOrgParent := make(map[int]common.MapRoleAndAuthInfo)

	for orgID := range mapOrgAndAdminID {
		if golibs.InArrayString(fmt.Sprint(orgID), []string{fmt.Sprint(constants.ManagaraHighSchool), fmt.Sprint(constants.ManagaraBase)}) {
			continue
		}
		validContext := common.ValidContext(ctx, orgID, rootAccount[orgID].UserID, rootAccount[orgID].Token)
		mapRoleAndAuthInfo := make(common.MapRoleAndAuthInfo)

		studentID := mapOrgUser[orgID][GetRoleFromConstant(student)].UserID
		resp, err := CreateParent(validContext, usermgmtConnection, nil, studentID)
		if err != nil {
			return nil, err
		}

		parentID := resp.ParentProfiles[0].Parent.UserProfile.UserId
		authInfo, err := GenerateFakeAuthInfo(ctx, connections.ShamirConn, firebaseAddr, jwtApplicant, parentID, constant.UserGroupParent)
		if err != nil {
			return nil, err
		}

		mapRoleAndAuthInfo[constant.RoleParent] = authInfo
		mapOrgParent[orgID] = mapRoleAndAuthInfo
	}

	return mapOrgParent, nil
}

func orgAndGrantedPermission() map[int][]RoleWithLocation {
	mapOrgAndGrantedPermission := make(map[int][]RoleWithLocation)
	for orgID := range mapOrgAndAdminID {
		roleWithLocations := []RoleWithLocation{}
		for _, role := range constant.AllowListRoles {
			var locationIDs []string
			switch role {
			case constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreLead, constant.RoleTeacherLead, constant.RoleUsermgmtScheduleJob:
				locationIDs = []string{GetOrgLocation(orgID)}
			case constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher:
				locationIDs = getChildrenLocation(orgID)
			default:
				continue
			}

			// jprep only have 2 roles: school admin and teacher so far
			if orgID == constants.JPREPSchool {
				if !golibs.InArrayString(role, []string{constant.RoleSchoolAdmin, constant.RoleTeacher}) {
					continue
				}
			}
			if orgID == constants.ManagaraBase || orgID == constants.ManagaraHighSchool {
				if !golibs.InArrayString(role, []string{constant.RoleUsermgmtScheduleJob}) {
					continue
				}
			}
			if orgID == constants.KECDemo {
				if !golibs.InArrayString(role, []string{constant.RoleSchoolAdmin}) {
					continue
				}
			}
			roleWithLocations = append(roleWithLocations, RoleWithLocation{
				RoleName:    role,
				LocationIDs: locationIDs,
			})
		}

		mapOrgAndGrantedPermission[orgID] = roleWithLocations
	}

	return mapOrgAndGrantedPermission
}

func prepairManabieBrandAndCenterLocations(ctx context.Context, db *pgxpool.Pool) ([]string, []string, error) {
	locationTypeIDs := []string{"location-type-id-1", "location-type-id-2"}
	locationIDs := []string{"location-id-1", "location-id-2", "location-id-3"}

	return locationTypeIDs, locationIDs, nil
}

// prepareBrandAndCenterTypesInDB
// init-ed from the file deployments/helm/manabie-all-in-one/charts/hephaestus/ksql/local-init-sql/3-local-init-user.sql
//
// [0]: branch, [1]: center
func prepareBrandAndCenterTypesInDB(ctx context.Context, db database.QueryExecer) ([]string, error) {
	locationTypeIDs := []string{"location-type-id-1", "location-type-id-2"}
	count := 0
	stmt := `
			SELECT count(*)
			FROM location_types
			WHERE
				location_type_id = ANY($1) AND
				deleted_at IS NULL AND
				is_archived = FALSE
	`
	err := db.QueryRow(ctx, stmt, database.TextArray(locationTypeIDs)).Scan(&count)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query location types")
	}

	if count != len(locationTypeIDs) {
		return nil, errors.Wrap(err, "location types were not initialized")
	}

	return locationTypeIDs, nil
}

// prepareBrandAndCenterLocations
// init-ed from the file deployments/helm/manabie-all-in-one/charts/hephaestus/ksql/local-init-sql/3-local-init-user.sql
//
// [0]: brand location type id, [1...]: center location type id
func prepareBrandAndCenterLocations(ctx context.Context, db database.QueryExecer) ([]string, error) {
	locationIDs := []string{"location-id-1", "location-id-2", "location-id-3"}
	count := 0

	stmt := `
		SELECT COUNT(*)
		FROM locations
		WHERE
			location_id = ANY($1) AND
			is_archived = FALSE AND
			deleted_at IS NULL
	`

	err := db.QueryRow(ctx, stmt, database.TextArray(locationIDs)).Scan(&count)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query location types")
	}

	if count != len(locationIDs) {
		return nil, errors.Wrap(err, "locations were not initialized")
	}

	return locationIDs, nil
}
