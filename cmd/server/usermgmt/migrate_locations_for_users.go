package usermgmt

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/manabie-com/backend/internal/golibs/auth"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/logger"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/location/domain"
	location_repo "github.com/manabie-com/backend/internal/mastermgmt/modules/location/infrastructure/repo"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/configurations"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
	"go.uber.org/zap"
)

const (
	limitBatchUsers = 1000
	// Because k8s job can determine space
	// so we define role without space
	SchoolAdminRole = "SchoolAdmin"
	student         = "student"
	staff           = "staff"
	parent          = "parent"
)

func RunMigrateLocationsForUsers(ctx context.Context, c *configurations.Config, organizationID, locationIDsSequences, userType string) {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	zLogger = logger.NewZapLogger("debug", c.Common.Environment == "local")
	zLogger.Sugar().Info("-----Migration Assign Location To User-----")
	defer zLogger.Sugar().Sync()

	// Define variable
	locationIDs := ExtractSliceFromSequenceElements(locationIDsSequences, Separator)
	locationRepo := &location_repo.LocationRepo{}

	dbPool, dbcancel, err := database.NewPool(ctx, zLogger, c.PostgresV2.Databases["bob"])
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := dbcancel(); err != nil {
			zLogger.Error("dbcancel() failed", zap.Error(err))
		}
	}()

	org, err := (&repository.OrganizationRepo{}).Find(ctx, dbPool, database.Text(strings.TrimSpace(organizationID)))
	if err != nil {
		zLogger.Fatal(fmt.Sprintf("find organization %s failed: %s", organizationID, err.Error()))
	}

	ctx = auth.InjectFakeJwtToken(ctx, org.OrganizationID.String)

	err = database.ExecInTx(ctx, dbPool, func(ctx context.Context, tx pgx.Tx) error {
		switch userType {
		case staff:
			if len(locationIDs) == 0 {
				var offset int32
				for {
					locations, err := locationRepo.RetrieveLowestLevelLocations(ctx, tx, &location_repo.GetLowestLevelLocationsParams{
						Offset: offset,
						Limit:  limitBatchUsers,
					})
					if err != nil {
						return errors.Wrap(err, "locationRepo.RetrieveLowestLevelLocations")
					}

					if len(locations) == 0 {
						break
					}

					if err := migrateLocationForRoles(ctx, tx, locations, userType, org.OrganizationID.String); err != nil {
						return errors.Wrapf(err, "can not run migration for roles %v", userType)
					}

					offset += limitBatchUsers
				}
			} else {
				locations, err := locationRepo.GetLocationsByLocationIDs(ctx, tx, database.TextArray(locationIDs), false)
				if err != nil {
					return errors.Wrap(err, "locationRepo.GetLocationsByLocationIDs")
				}

				if len(locations) != len(locationIDs) {
					return errors.New("location id must be existed in database")
				}

				if err := migrateLocationForRoles(ctx, tx, locations, userType, org.OrganizationID.String); err != nil {
					return errors.Wrapf(err, "can not run migration for roles %v", userType)
				}
			}
		case student:
			location, err := locationRepo.GetLocationOrg(ctx, tx, org.OrganizationID.String)
			if err != nil {
				return errors.Wrap(err, "locationRepo.GetLocationOrg")
			}

			locations := []*domain.Location{location}

			if err := migrateLocationForRoles(ctx, tx, locations, userType, org.OrganizationID.String); err != nil {
				return errors.Wrapf(err, "can not run migration for roles %v", constant.RoleStudent)
			}
		case parent:
			if err := migrateLocationForRoles(ctx, tx, []*domain.Location{}, userType, org.OrganizationID.String); err != nil {
				return errors.Wrapf(err, "can not run migration for roles %v", constant.RoleParent)
			}

		default:
			return errors.New("must have userType")
		}

		return nil
	})
	if err != nil {
		zLogger.Fatal(fmt.Sprintf("RunMigrateLocationsForUsers failed: %s", err.Error()))
	}

	zLogger.Sugar().Info("-----Done Migration Assign Location To User-----")
}

// Only use for special case parent
func findParentIDWithStudentLocation(ctx context.Context, db database.QueryExecer, lastUserID, orgID string) (entity.UserAccessPaths, error) {
	ctx, span := interceptors.StartSpan(ctx, "RunMigrateLocationsForUsers.findParentIDWithStudentLocation")
	defer span.End()

	var userAccessPaths entity.UserAccessPaths

	query := `
		SELECT  sp.parent_id, uap.location_id
		FROM student_parents sp
		JOIN user_access_paths uap ON uap.user_id = sp.student_id  
		WHERE sp.parent_id > $1
			AND sp.resource_path = $2
			AND sp.student_id IS NOT NULL 
			AND sp.parent_id  IS NOT NULL
			AND sp.deleted_at IS NULL 
			AND uap.deleted_at IS NULL 
		ORDER BY sp.parent_id, uap.location_id

		LIMIT $3
	`

	rows, err := db.Query(ctx, query, lastUserID, orgID, limitBatchUsers)
	if err != nil {
		return nil, errors.Wrap(err, "can not query student_parents")
	}

	for rows.Next() {
		var parentID, locationID string
		userAccessPathEnt := &entity.UserAccessPath{}
		database.AllNullEntity(userAccessPathEnt)

		err := rows.Scan(&parentID, &locationID)
		if err != nil {
			return nil, errors.Wrap(err, "can not scan student_parents")
		}

		if err := multierr.Combine(
			userAccessPathEnt.UserID.Set(parentID),
			userAccessPathEnt.LocationID.Set(locationID),
		); err != nil {
			return nil, err
		}

		userAccessPaths = append(userAccessPaths, userAccessPathEnt)
	}

	if rows.Err() != nil {
		return nil, errors.Wrap(err, "rows error when query student_parents")
	}

	return userAccessPaths, nil
}

func findUserWithoutLocation(ctx context.Context, db database.QueryExecer, userType, lastUserID, orgID string) ([]string, error) {
	ctx, span := interceptors.StartSpan(ctx, "RunMigrateLocationsForUsers.findUserWithoutLocation")
	defer span.End()

	var userIDs []string
	var userGroup string

	switch userType {
	case staff:
		userGroup = `JOIN staff s ON s.staff_id  = u.user_id `

	case student:
		userGroup = `JOIN students s ON s.student_id  = u.user_id `
	}

	query := fmt.Sprintf(`
		SELECT DISTINCT(u.user_id) 
		FROM users u 
		%v
		LEFT JOIN user_access_paths uap ON uap.user_id = u.user_id 
		WHERE u.user_id > $1 
			AND u.resource_path = $2
			AND (
				uap.user_id IS NULL 
				OR 
					( 
						uap.deleted_at IS NOT NULL 
						AND NOT EXISTS (
							SELECT uapChild.user_id
							FROM user_access_paths uapChild
							WHERE  uapChild.user_id = uap.user_id AND uapChild.deleted_at IS NULL
						)
					)
			)
		ORDER BY u.user_id
		LIMIT $3
	`, userGroup)

	rows, err := db.Query(ctx, query, lastUserID, orgID, limitBatchUsers)
	if err != nil {
		return nil, errors.Wrap(err, "can not query user ids")
	}

	for rows.Next() {
		var userID string
		err := rows.Scan(&userID)
		if err != nil {
			return nil, errors.Wrap(err, "can not scan user ids")
		}
		userIDs = append(userIDs, userID)
	}

	if rows.Err() != nil {
		return nil, errors.Wrap(err, "rows error when query user ids")
	}

	return userIDs, nil
}

func migrateLocationForRoles(ctx context.Context, db database.QueryExecer, locations []*domain.Location, userType, orgID string) error {
	var offsetUserID string
	switch userType {
	case parent:
		for {
			userAccessPaths, err := findParentIDWithStudentLocation(ctx, db, offsetUserID, orgID)
			if err != nil {
				return errors.Wrap(err, "findParentIDWithStudentLocation")
			}

			if len(userAccessPaths) == 0 {
				break
			}

			if err := migrateLocationForUsers(ctx, db, userAccessPaths); err != nil {
				return errors.Wrap(err, "migrateLocationForUsers")
			}

			offsetUserID = userAccessPaths[len(userAccessPaths)-1].UserID.String
		}
	default:
		for {
			userIDs, err := findUserWithoutLocation(ctx, db, userType, offsetUserID, orgID)
			if err != nil {
				return errors.Wrap(err, "findUserWithoutLocation")
			}

			if len(userIDs) == 0 {
				break
			}

			userAccessPaths, err := toUserAccessPaths(locations, userIDs)
			if err != nil {
				return errors.Wrap(err, "toUserAccessPaths")
			}

			if err := migrateLocationForUsers(ctx, db, userAccessPaths); err != nil {
				return errors.Wrap(err, "migrateLocationForUsers")
			}

			offsetUserID = userIDs[len(userIDs)-1]
		}
	}

	return nil
}

func migrateLocationForUsers(ctx context.Context, db database.QueryExecer, userAccessPaths []*entity.UserAccessPath) error {
	zLogger.Sugar().Infof(".....Found %d locations and users need to be migrated.....", len(userAccessPaths))
	userAccessPathRepo := new(repository.UserAccessPathRepo)

	if err := userAccessPathRepo.Upsert(ctx, db, userAccessPaths); err != nil {
		return errors.Wrap(err, "userAccessPathRepo.Upsert")
	}

	return nil
}

func toUserAccessPaths(locations []*domain.Location, userIDs []string) ([]*entity.UserAccessPath, error) {
	var userAccessPaths []*entity.UserAccessPath

	for _, userID := range userIDs {
		for _, location := range locations {
			userAccessPathEnt := &entity.UserAccessPath{}
			database.AllNullEntity(userAccessPathEnt)

			if err := multierr.Combine(
				userAccessPathEnt.UserID.Set(userID),
				userAccessPathEnt.LocationID.Set(location.LocationID),
			); err != nil {
				return nil, err
			}

			userAccessPaths = append(userAccessPaths, userAccessPathEnt)
		}
	}

	return userAccessPaths, nil
}
