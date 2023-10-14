package usermgmt

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	bob_entities "github.com/manabie-com/backend/internal/bob/entities"
	bob_repo "github.com/manabie-com/backend/internal/bob/repositories"
	internal_auth_tenant "github.com/manabie-com/backend/internal/golibs/auth/multitenant"
	internal_auth_user "github.com/manabie-com/backend/internal/golibs/auth/user"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/gcp"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/logger"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/configurations"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"

	"github.com/jackc/pgtype"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

const LocalTestMigrationTenant = "migrate-users-ge51c"

var LocalSchoolAndTenantIDMap = map[int]string{
	constants.ManabieSchool:   "manabie-0nl6t",
	constants.RenseikaiSchool: "renseikai-yu9y7",
	constants.SynersiaSchool:  "synersia-24rue",
	constants.TestingSchool:   "end-to-end-dopvo",
}

var stagingSchoolAndTenantIDMap = map[int]string{
	constants.ManabieSchool:   "manabie-p7muf",
	constants.SynersiaSchool:  "synersia-m3hr5",
	constants.RenseikaiSchool: "renseikai-5ayyd",
	constants.TestingSchool:   "end-to-end-school-5xn27",
	constants.GASchool:        "ga-school-rtaas",
	constants.KECSchool:       "kec-school-q148v",
	constants.AICSchool:       "aic-school-xhy07",
	constants.NSGSchool:       "nsg-school-5wkab",
	16091:                     "manabie2-edzop",
	constants.E2ETokyo:        "e2e-tokyo-1lpnq",
}

var uatSchoolAndTenantIDMap = map[int]string{
	constants.ManabieSchool:   "manabie-9h0ng",
	constants.SynersiaSchool:  "synersia-oodzl",
	constants.RenseikaiSchool: "renseikai-zxx25",
	constants.TestingSchool:   "end-to-end-school-5mqoc",
	constants.GASchool:        "ga-school-q3mvu",
	constants.KECSchool:       "kec-school-8qe69",
	constants.AICSchool:       "aic-school-fj80n",
	constants.NSGSchool:       "nsg-school-yevm8",
	16091:                     "manabie2-44xpg",
	constants.E2ETokyo:        "e2e-tokyo-druuc",
}

const (
	prodManabieFirebaseProject   string = "production-manabie-vn"
	prodSynerisaFirebaseProject  string = "synersia"
	prodRenseikaiFirebaseProject string = "production-renseikai"
	prodGAFirebaseProject        string = "production-ga"
	prodKECFirebaseProject       string = "production-kec"
	prodAICFirebaseProject       string = "production-aic"
	prodNSGFirebaseProject       string = "production-nsg"
	prodE2ETokyoFirebaseProject  string = "student-coach-e1e95"
)

var prodManabieSchoolAndTenantIDMap = map[int]string{
	constants.ManabieSchool: "prod-manabie-bj1ok",
	constants.TestingSchool: "prod-manabie-bj1ok",
}

var prodSynersiaSchoolAndTenantIDMap = map[int]string{
	constants.SynersiaSchool: "prod-synersia-trscc",
	constants.TestingSchool:  "prod-synersia-trscc",
}

var prodRenseikaiSchoolAndTenantIDMap = map[int]string{
	constants.RenseikaiSchool: "prod-renseikai-8xr29",
	constants.TestingSchool:   "prod-renseikai-8xr29",
}

/*var testingSchoolProdSchoolAndTenantIDMap = map[int]string{
	constants.TestingSchool: "prod-end-to-end-og7nh",
}*/

var prodGAdSchoolAndTenantIDMap = map[int]string{
	constants.GASchool:      "prod-ga-uq2rq",
	constants.TestingSchool: "prod-ga-uq2rq",
	constants.ManabieSchool: "prod-ga-uq2rq",
}

var prodKECSchoolAndTenantIDMap = map[int]string{
	constants.KECSchool:     "prod-kec-58ww0",
	constants.TestingSchool: "prod-kec-58ww0",
}

var prodAICSchoolAndTenantIDMap = map[int]string{
	constants.AICSchool:     "prod-aic-u3d1m",
	constants.TestingSchool: "prod-aic-u3d1m",
}

var prodNSGSchoolAndTenantIDMap = map[int]string{
	constants.NSGSchool:     "prod-nsg-flbh7",
	constants.TestingSchool: "prod-nsg-flbh7",
}

var prodE2ETokyoSchoolAndTenantIDMap = map[int]string{
	constants.E2ETokyo: "prod-e2e-tokyo-2k4xb",
}

func schoolName(schoolID int) string {
	switch schoolID {
	case constants.ManabieSchool:
		return "manabie"
	case constants.SynersiaSchool:
		return "synersia"
	case constants.RenseikaiSchool:
		return "renseikai"
	case constants.TestingSchool:
		return "end-to-end"
	case constants.GASchool:
		return "ga"
	case constants.KECSchool:
		return "kec"
	case constants.AICSchool:
		return "aic"
	case constants.NSGSchool:
		return "nsg"
	case 16091:
		return "manabie2"
	default:
		return "undefined"
	}
}

func schoolIDAndTenant(env string, firebaseProject string) map[int]string {
	switch env {
	case "local":
		return LocalSchoolAndTenantIDMap
	case "stag":
		return stagingSchoolAndTenantIDMap
	case "uat":
		return uatSchoolAndTenantIDMap
	case "prod":
		switch firebaseProject {
		case prodManabieFirebaseProject:
			return prodManabieSchoolAndTenantIDMap
		case prodSynerisaFirebaseProject:
			return prodSynersiaSchoolAndTenantIDMap
		case prodRenseikaiFirebaseProject:
			return prodRenseikaiSchoolAndTenantIDMap
		case prodGAFirebaseProject:
			return prodGAdSchoolAndTenantIDMap
		case prodKECFirebaseProject:
			return prodKECSchoolAndTenantIDMap
		case prodAICFirebaseProject:
			return prodAICSchoolAndTenantIDMap
		case prodNSGFirebaseProject:
			return prodNSGSchoolAndTenantIDMap
		case prodE2ETokyoFirebaseProject:
			return prodE2ETokyoSchoolAndTenantIDMap
		default:
			return nil
		}
	default:
		return nil
	}
}

func TenantClientMap(ctx context.Context, env string, firebaseProject string, tenantManager internal_auth_tenant.TenantManager) (map[int]internal_auth_tenant.TenantClient, error) {
	m := make(map[int]internal_auth_tenant.TenantClient, len(stagingSchoolAndTenantIDMap))

	schoolAndTenantIDMap := schoolIDAndTenant(env, firebaseProject)

	for schoolID, tenantID := range schoolAndTenantIDMap {
		tenantClient, err := tenantManager.TenantClient(ctx, tenantID)
		if err != nil {
			return nil, errors.Wrap(err, "TenantClient")
		}

		m[schoolID] = tenantClient
	}

	return m, nil
}

type migrationStat struct {
	totalIteratedUsers                    int
	totalUsersInDBHaveInvalidResourcePath int
	totalUsersInAuthButNotInDB            int
	schoolAndSuccessImportCount           map[int]int
	schoolAndFailureImportCount           map[int]int
}

func newMigrationStat() *migrationStat {
	s := &migrationStat{
		schoolAndSuccessImportCount: make(map[int]int),
		schoolAndFailureImportCount: make(map[int]int),
	}
	return s
}

// MigrateUsersFromFirebase migrate user from firebase auth to identity platform
// In local env, this should only execute for testing purpose
// Current target of this migration is stag env
func MigrateUsersFromFirebase(ctx context.Context, c *configurations.Config, organizationID string) error {
	zapLogger := logger.NewZapLogger("debug", c.Common.Environment == "local")
	sugaredLogger := zapLogger.Sugar()
	defer sugaredLogger.Sync()

	dbPool, dbcancel, err := database.NewPool(ctx, zapLogger, c.PostgresV2.Databases["bob"])
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := dbcancel(); err != nil {
			zapLogger.Error("dbcancel() failed", zap.Error(err))
		}
	}()

	secondaryTenantConfigProvider := &repository.TenantConfigRepo{
		QueryExecer:      dbPool,
		ConfigAESKey:     c.IdentityPlatform.ConfigAESKey,
		ConfigAESIv:      c.IdentityPlatform.ConfigAESIv,
		OrganizationRepo: &repository.OrganizationRepo{},
	}

	firebaseProject := c.Common.FirebaseProject
	if firebaseProject == "" {
		firebaseProject = c.Common.GoogleCloudProject
	}

	// Init source firebase auth client
	srcApp, err := gcp.NewApp(ctx, "", firebaseProject)
	if err != nil {
		return errors.Wrap(err, "NewApp: failed to init srcApp")
	}

	srcFirebaseAuthClient, err := internal_auth_tenant.NewFirebaseAuthClientFromGCP(ctx, srcApp)
	if err != nil {
		return errors.Wrap(err, "Auth")
	}

	// For testing
	if c.Common.Environment == "local" {
		// Use tenant in identity platform instead of firebase auth
		srcTenantManager, err := internal_auth_tenant.NewTenantManagerFromGCP(ctx, srcApp, internal_auth_tenant.WithSecondaryTenantConfigProvider(secondaryTenantConfigProvider))
		if err != nil {
			return errors.Wrap(err, "NewTenantManagerFromGCP")
		}
		srcFirebaseAuthClient, err = srcTenantManager.TenantClient(ctx, LocalTestMigrationTenant)
		if err != nil {
			return errors.Wrap(err, "NewTenantManagerFromGCP")
		}
	}

	// Init destination tenant client
	destApp, err := gcp.NewApp(ctx, "", c.Common.IdentityPlatformProject)
	if err != nil {
		return errors.Wrap(err, "NewApp: failed to init destApp")
	}

	destTenantManager, err := internal_auth_tenant.NewTenantManagerFromGCP(ctx, destApp, internal_auth_tenant.WithSecondaryTenantConfigProvider(secondaryTenantConfigProvider))
	if err != nil {
		return errors.Wrap(err, "NewTenantManagerFromGCP")
	}

	schoolAndTenantClient, err := TenantClientMap(ctx, c.Common.Environment, firebaseProject, destTenantManager)
	if err != nil {
		return errors.Wrap(err, "TenantClient")
	}

	for schoolID, tenantClient := range schoolAndTenantClient {
		if fmt.Sprint(schoolID) != organizationID {
			sugaredLogger.Info("skip organizationID", fmt.Sprint(schoolID))
			continue
		}
		sugaredLogger.Infow(
			"start user migration process",
			"srcSchoolName", schoolName(schoolID),
			"destTenantID", tenantClient.TenantID(),
		)
	}
	sugaredLogger.Info("----------------------------------")

	// Count number of users
	stat := newMigrationStat()

	err = srcFirebaseAuthClient.IterateAllUsers(ctx, 1000, func(iteratedFirebaseUsers internal_auth_user.Users) error {
		schoolAndUsers := make(map[int]internal_auth_user.Users)

		iteratedUsers := make(internal_auth_user.Users, 0, len(iteratedFirebaseUsers))
		for _, iteratedFirebaseUser := range iteratedFirebaseUsers {
			if strings.TrimSpace(iteratedFirebaseUser.GetEmail()) == "" {
				continue
			}
			iteratedUsers = append(iteratedUsers, iteratedFirebaseUser)
		}

		stat.totalIteratedUsers += len(iteratedUsers)

		iteratedUsersMap := iteratedUsers.IDAndUserMap()

		/*
			//Maybe school id in user's custom claims can be useful in some cases so keep this code
			for _, iteratedUser := range iteratedUsers {
				hasuraClaims, ok := iteratedUser.GetCustomClaims()["https://hasura.io/jwt/claims"].(map[string]interface{})
				if !ok {
					zlogger.Errorf(`can't detect hasura claims of user with ID: "%v"`, iteratedUser.GetUID())
					continue
				}

				userSchoolIDStr, ok := hasuraClaims["x-hasura-school-id"].(string)
				if !ok {
					zlogger.Errorf(`can't detect hasura school id of user with ID: "%v"`, iteratedUser.GetUID())
					continue
				}

				userSchoolID, err := strconv.Atoi(userSchoolIDStr)
				if err != nil {
					zlogger.Errorf(`can't parse hasura school id of user with ID: "%v"`, iteratedUser.GetUID())
					continue
				}

				_, hasTenantClient := schoolAndTenantClient[userSchoolID]
				if !hasTenantClient {
					continue
				}

				schoolAndUsers[userSchoolID].Append(iteratedUser)
			}*/

		usersInDB, err := getUsersFromAllSchool(ctx, schoolAndTenantClient, dbPool, database.TextArray(iteratedUsers.UserIDs()), organizationID)
		if err != nil {
			return errors.Wrap(err, "(&bob_repo.UserRepo{}).Retrieve")
		}

		for _, userInDB := range usersInDB {
			if strings.TrimSpace(userInDB.ResourcePath.String) == "" {
				stat.totalUsersInDBHaveInvalidResourcePath++
				continue
			}

			splitResourcePath := strings.Split(userInDB.ResourcePath.String, ":")
			if len(splitResourcePath) < 1 {
				stat.totalUsersInDBHaveInvalidResourcePath++
				continue
			}

			schoolID, err := strconv.Atoi(splitResourcePath[0])
			if err != nil {
				stat.totalUsersInDBHaveInvalidResourcePath++
				continue
			}

			_, ok := schoolAndTenantClient[schoolID]
			if !ok {
				continue
			}

			schoolAndUsers[schoolID] = append(schoolAndUsers[schoolID], iteratedUsersMap[userInDB.ID.String])
		}

		// Check if users are in auth platform but not in db
		if len(usersInDB) != len(iteratedUsers) {
			usersInDBMap := make(map[string]*bob_entities.User, len(usersInDB))
			for _, userInDB := range usersInDB {
				usersInDBMap[userInDB.ID.String] = userInDB
			}

			usersNotFound := make(internal_auth_user.Users, 0, len(iteratedUsers))
			for _, iteratedUser := range iteratedUsers {
				_, found := usersInDBMap[iteratedUser.GetUID()]
				if !found {
					usersNotFound = append(usersNotFound, iteratedUser)
				}
			}

			stat.totalUsersInAuthButNotInDB += len(usersNotFound)
			sugaredLogger.Infow(
				"there are users in auth platform but can't find in our db",
				"ids", usersNotFound.UserIDs(),
				"emails", usersNotFound.Emails(),
			)
		}

		for schoolID, users := range schoolAndUsers {
			tenantClient := schoolAndTenantClient[schoolID]
			result, err := tenantClient.ImportUsers(ctx, users, srcFirebaseAuthClient.GetHashConfig())
			if err != nil {
				return errors.Wrap(err, "destTenantClient.ImportUsers")
			}

			if len(result.UsersSuccessImport) > 0 {
				stat.schoolAndSuccessImportCount[schoolID] += len(result.UsersSuccessImport)
				sugaredLogger.Infow(
					"imported user successfully",
					"totalUsers", len(result.UsersSuccessImport),
					"srcSchoolName", schoolName(schoolID),
					"destTenantID", tenantClient.TenantID(),
				)
			}
			if len(result.UsersFailedToImport) > 0 {
				stat.schoolAndFailureImportCount[schoolID] += len(result.UsersFailedToImport)
				sugaredLogger.Errorw(
					"failed to import users",
					"totalUsers", len(result.UsersFailedToImport),
					"srcSchoolName", schoolName(schoolID),
					"destTenantID", tenantClient.TenantID(),
					"userIDs", result.UsersFailedToImport.IDs(),
					"userEmails", result.UsersFailedToImport.Emails(),
				)
			}
		}

		return nil
	})
	if err != nil {
		return errors.Wrap(err, "IterateAllUsers")
	}

	sugaredLogger.Info("----------------------------------")
	for schoolID, tenantClient := range schoolAndTenantClient {
		sugaredLogger.Infow(
			"finished user migration process",
			"totalIteratedUsers", stat.totalIteratedUsers,
			"srcSchoolName", schoolName(schoolID),
			"destTenantID", tenantClient.TenantID(),
			"totalSuccessImporting", stat.schoolAndSuccessImportCount[schoolID],
			"totalFailureImporting", stat.schoolAndFailureImportCount[schoolID],
			"totalUsersInAuthButNotInDB", stat.totalUsersInAuthButNotInDB,
			"totalUsersInDBHaveInvalidResourcePath", stat.totalUsersInDBHaveInvalidResourcePath)
	}

	return nil
}

func getUsersFromAllSchool(ctx context.Context, schoolIDsAndTenantClient map[int]internal_auth_tenant.TenantClient, db database.QueryExecer, userIDs pgtype.TextArray, orgID string) ([]*bob_entities.User, error) {
	usersInDB := make([]*bob_entities.User, 0, 1000)

	for schoolID := range schoolIDsAndTenantClient {
		if fmt.Sprint(schoolID) != orgID {
			continue
		}
		usermgmtUserIDs, err := GetUsermgmtUserIDByOrgID(ctx, db, orgID)
		if err != nil {
			zLogger.Sugar().Errorf("failed to get usermgmt user id: %s", err)
			continue
		}
		if len(usermgmtUserIDs) == 0 {
			zLogger.Sugar().Errorf("cannot find userID on organization: %s", organizationID)
			continue
		}
		internalUserID := usermgmtUserIDs[0]

		ctx = interceptors.ContextWithUserID(ctx, internalUserID)
		ctx = interceptors.ContextWithJWTClaims(ctx, &interceptors.CustomClaims{
			Manabie: &interceptors.ManabieClaims{
				ResourcePath: fmt.Sprint(schoolID),
				UserGroup:    cpb.UserGroup_USER_GROUP_SCHOOL_ADMIN.String(),
				UserID:       internalUserID,
			},
		})

		users, err := (&bob_repo.UserRepo{}).Retrieve(ctx, db, userIDs)
		if err != nil {
			return nil, errors.Wrap(err, "(&bob_repo.UserRepo{}).Retrieve")
		}
		usersInDB = append(usersInDB, users...)
	}
	return removeDuplicate(usersInDB), nil
}

// In env that doesn't enable RLS yet, returned users are not filtered
// by resource path, so they may duplicate
// This prevents env that doesn't enable RLS has duplicated users when query
// This is safe for env that enabled RLS
func removeDuplicate(users []*bob_entities.User) []*bob_entities.User {
	if len(users) < 1 {
		return users
	}
	existedUser := make(map[string]bool)

	uniqueUsers := make([]*bob_entities.User, 0, 1000)
	for _, user := range users {
		if !existedUser[user.GetUID()] {
			existedUser[user.GetUID()] = true
			uniqueUsers = append(uniqueUsers, user)
		}
	}

	return uniqueUsers
}
