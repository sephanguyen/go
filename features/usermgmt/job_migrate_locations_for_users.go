package usermgmt

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/cmd/server/usermgmt"
	"github.com/manabie-com/backend/internal/entryexitmgmt/constant"
	"github.com/manabie-com/backend/internal/golibs/auth"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/location/domain"
	location_repo "github.com/manabie-com/backend/internal/mastermgmt/modules/location/infrastructure/repo"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/configurations"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
)

const (
	staff   = "staff"
	student = "student"
	parent  = "parent"
)

func (s *suite) existedUserWithoutLocation(ctx context.Context, userType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.ResourcePath == "" {
		stepState.ResourcePath = fmt.Sprint(constants.ManabieSchool)
	}

	stepState.UserIDs = []string{}
	orgID := OrgIDFromCtx(ctx)
	ctx = s.signedIn(ctx, orgID, StaffRoleSchoolAdmin)

	switch userType {
	case staff:
		for i := 0; i < amountUserToTest; i++ {
			roleWithLocationTeacher := RoleWithLocation{
				RoleName:    constant.RoleTeacher,
				LocationIDs: []string{constants.ManabieOrgLocation},
			}
			resp, err := CreateStaff(ctx, s.BobDBTrace, s.UserMgmtConn, nil, []RoleWithLocation{roleWithLocationTeacher}, getChildrenLocation(orgID))
			if err != nil {
				return nil, errors.Wrap(err, "CreateStaff")
			}
			stepState.UserIDs = append(stepState.UserIDs, resp.Staff.StaffId)
		}
	case student:
		for i := 0; i < amountUserToTest; i++ {
			studentRepo := &repository.StudentRepo{}
			newStudentID := idutil.ULIDNow()

			stepState.UserIDs = append(stepState.UserIDs, newStudentID)

			studentEntityWithFullNameOnly, err := studentEntityWithFullNameOnly(newStudentID)
			if err != nil {
				return StepStateToContext(ctx, stepState), err
			}
			err = studentRepo.Create(ctx, s.BobDBTrace, studentEntityWithFullNameOnly)
			if err != nil {
				return StepStateToContext(ctx, stepState), err
			}
		}
	case parent:
		studentParents := []*entity.StudentParent{}
		userAccessPathEnt := &entity.UserAccessPath{}
		userAccessPathEnt2 := &entity.UserAccessPath{}
		database.AllNullEntity(userAccessPathEnt)
		database.AllNullEntity(userAccessPathEnt2)

		userAccessPathRepo := &repository.UserAccessPathRepo{}
		studentParentRepo := &repository.StudentParentRepo{}

		// Create new student and then assign location to that student
		userID := fmt.Sprintf("userWithoutLocation-%v-", parent) + idutil.ULIDNow()
		if _, err := s.aValidStudentInDB(ctx, userID); err != nil {
			return nil, errors.Wrap(err, "s.aValidStudentInDB")
		}

		// stepState.ExistingLocations = {0:manabie, 1:jprep, 2:manabie}
		if err := multierr.Combine(
			userAccessPathEnt.UserID.Set(userID),
			userAccessPathEnt.LocationID.Set(stepState.ExistingLocations[0].LocationID.String),
			userAccessPathEnt2.UserID.Set(userID),
			userAccessPathEnt2.LocationID.Set(stepState.ExistingLocations[1].LocationID.String),
		); err != nil {
			return nil, err
		}

		if err := userAccessPathRepo.Upsert(ctx, s.BobDBTrace, []*entity.UserAccessPath{userAccessPathEnt, userAccessPathEnt2}); err != nil {
			return nil, errors.Wrap(err, "userAccessPathRepo.Upsert")
		}

		// Create new parent and connect relationship with student
		for i := 0; i < amountUserToTest; i++ {
			studentParent := &entity.StudentParent{}
			database.AllNullEntity(studentParent)

			parentID := fmt.Sprintf("userWithoutLocation-parent-%v-", i) + idutil.ULIDNow()

			// Create new parent
			_, err := aValidParentInDB(ctx, s.BobDBTrace, parentID)
			if err != nil {
				return nil, errors.Wrap(err, "s.aValidParentInDB")
			}

			if err := multierr.Combine(
				studentParent.ParentID.Set(parentID),
				studentParent.StudentID.Set(userID),
				studentParent.Relationship.Set(pb.FamilyRelationship_FAMILY_RELATIONSHIP_FATHER),
			); err != nil {
				return nil, err
			}

			studentParents = append(studentParents, studentParent)
			stepState.UserIDs = append(stepState.UserIDs, parentID)
		}

		// insert new relationship between new student and parent we just create
		if err := studentParentRepo.Upsert(ctx, s.BobDBTrace, studentParents); err != nil {
			return nil, errors.Wrap(err, "studentParentRepo.Upsert")
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) runMigrationLocationsForUsers(ctx context.Context, schoolName, locationType, userType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.ResourcePath = fmt.Sprint(SchoolNameWithResourcePath[schoolName])

	var err error
	stepState.LocationIDs = []string{}

	switch locationType {
	case "empty location id":
		err := database.ExecInTx(ctx, s.BobPostgresDB, func(ctx context.Context, tx pgx.Tx) error {
			locations := []*domain.Location{}
			locationRepo := &location_repo.LocationRepo{}

			switch userType {
			case staff:
				// In migration assign location to staff
				// we must assign lowest location of this org to user if without passing any ids
				locations, err = locationRepo.RetrieveLowestLevelLocations(ctx, tx, &location_repo.GetLowestLevelLocationsParams{Offset: 0, Limit: 1000})
				if err != nil {
					return errors.Wrap(err, "locationRepo.RetrieveLowestLevelLocations")
				}

			case student:
				// In migration assign location to staff
				// we must assign location of this org to user if without passing any ids
				location, err := locationRepo.GetLocationOrg(ctx, tx, stepState.ResourcePath)
				if err != nil {
					return errors.Wrap(err, "locationRepo.GetLocationOrg")
				}
				locations = []*domain.Location{location}

			case parent:
				locations = []*domain.Location{
					{LocationID: stepState.ExistingLocations[0].LocationID.String},
					{LocationID: stepState.ExistingLocations[1].LocationID.String},
				}
			}

			for _, location := range locations {
				stepState.LocationIDs = append(stepState.LocationIDs, location.LocationID)
			}

			return nil
		})
		if err != nil {
			return StepStateToContext(ctx, stepState), errors.Wrap(err, "database.ExecInTx get locations")
		}

	case "one location id":
		// stepState.ExistingLocations = {0:manabie, 1:jprep, 2:manabie}
		stepState.LocationIDs = []string{stepState.ExistingLocations[0].LocationID.String}
	}

	usermgmt.RunMigrateLocationsForUsers(
		auth.InjectFakeJwtToken(ctx, fmt.Sprint(stepState.SchoolID)),
		&configurations.Config{
			Common:     s.Cfg.Common,
			PostgresV2: s.Cfg.PostgresV2,
		},
		stepState.ResourcePath,
		strings.Join(stepState.LocationIDs, usermgmt.Separator),
		userType,
	)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) usersMustHaveLocation(ctx context.Context, userType, locationType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx = auth.InjectFakeJwtToken(ctx, stepState.ResourcePath)
	userLocations := map[string]map[string]struct{}{}

	query := `
		SELECT user_id, location_id
		FROM user_access_paths
		WHERE user_id = ANY($1) AND deleted_at IS NULL
		GROUP BY user_id, location_id
	`

	rows, err := s.BobDB.Query(ctx, query, database.TextArray(stepState.UserIDs))
	if err != nil {
		return StepStateToContext(ctx, stepState), errors.Wrap(err, "s.BobDB.Query")
	}

	for rows.Next() {
		userID := ""
		locationID := ""
		if err := rows.Scan(&userID, &locationID); err != nil {
			return StepStateToContext(ctx, stepState), errors.Wrap(err, "rows.Scan")
		}
		if _, ok := userLocations[userID]; !ok {
			userLocations[userID] = make(map[string]struct{})
		}

		userLocations[userID][locationID] = struct{}{}
	}

	if rows.Err() != nil {
		return StepStateToContext(ctx, stepState), errors.Wrap(err, "rows.Err")
	}

	for _, userID := range stepState.UserIDs {
		locationIDs, ok := userLocations[userID]
		if !ok {
			return StepStateToContext(ctx, stepState), fmt.Errorf("locations of user %s was not migrated", userID)
		}
		for _, locationID := range stepState.LocationIDs {
			if _, ok := locationIDs[locationID]; !ok {
				return StepStateToContext(ctx, stepState), fmt.Errorf("user %s missed location %s", userID, locationID)
			}
		}
	}

	return StepStateToContext(ctx, stepState), nil
}
