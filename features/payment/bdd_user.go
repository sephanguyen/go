package payment

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/features/common"
	"github.com/manabie-com/backend/features/communication/common/helpers"
	"github.com/manabie-com/backend/features/payment/entities"
	"github.com/manabie-com/backend/features/usermgmt"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/try"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/location/infrastructure/repo"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"

	"github.com/jackc/pgtype"
)

// signedAsAccountWithLocations user root account of ManabieSchool to sign in a user on multiple locations with specific user group
// Make sure user is synced to fatima from, if not, insert user in fatima
func (s *suite) signedAsAccountWithLocations(ctx context.Context, userGroup string, locationIDs []string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if len(locationIDs) == 0 {
		return StepStateToContext(ctx, stepState), errors.New("length of locations is empty")
	}

	ctx = common.ValidContext(ctx, constants.ManabieSchool, s.RootAccount[constants.ManabieSchool].UserID, s.RootAccount[constants.ManabieSchool].Token)
	roleWithLocation := usermgmt.RoleWithLocation{
		LocationIDs: locationIDs,
	}
	stepState.CurrentSchoolID = constants.ManabieSchool
	switch userGroup {
	case UserGroupSchoolAdmin:
		roleWithLocation.RoleName = constant.RoleSchoolAdmin
	case UserGroupHQStaff:
		roleWithLocation.RoleName = constant.RoleHQStaff
	case UserGroupCentreLead:
		roleWithLocation.RoleName = constant.RoleCentreLead
	case UserGroupCentreManager:
		roleWithLocation.RoleName = constant.RoleCentreManager
	case UserGroupCentreStaff:
		roleWithLocation.RoleName = constant.RoleCentreStaff
	case UserGroupTeacher:
		roleWithLocation.RoleName = constant.RoleTeacher
	case UserGroupTeacherLead:
		roleWithLocation.RoleName = constant.RoleTeacherLead
	case UserGroupStudent:
		roleWithLocation.RoleName = constant.RoleStudent
	case UserGroupParent:
		roleWithLocation.RoleName = constant.UserGroupParent
	case UserGroupOrganizationManager:
		roleWithLocation.RoleName = constant.UserGroupOrganizationManager
	default:
		return StepStateToContext(ctx, stepState), errors.New("user group is invalid")
	}

	authInfo, err := usermgmt.SignIn(ctx, s.BobDBTrace, s.AuthPostgresDB, s.ShamirConn, s.Cfg.JWTApplicant, s.FirebaseAddress, s.UserMgmtConn, roleWithLocation, locationIDs)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.CurrentUserID = authInfo.UserID
	stepState.AuthToken = authInfo.Token
	stepState.LocationID = constants.ManabieOrgLocation
	stepState.CurrentUserGroup = userGroup
	stepState.LocationIDs = locationIDs

	ctx = common.ValidContext(ctx, constants.ManabieSchool, authInfo.UserID, authInfo.Token)

	err = try.Do(func(attempt int) (bool, error) {
		time.Sleep(1 * time.Second)
		err = s.getAdmin(ctx, stepState.CurrentUserID)
		if err == nil {
			return false, nil
		}
		retry := attempt <= 5
		if retry {
			return true, nil
		}
		err = s.insertAdmin(ctx, stepState.CurrentUserID, fmt.Sprintf("name-user-id-%s", authInfo.UserID))
		if err != nil {
			return false, fmt.Errorf("error when user info have not been synced from bob to fatima: %s", err.Error())
		}
		return false, nil
	})
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

// createLocations use root account to insert some location types, locations in bob (synced to fatima)
// Store locationIDs in stepState
func (s *suite) createLocations(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx = common.ValidContext(ctx, constants.ManabieSchool, s.RootAccount[constants.ManabieSchool].UserID, s.RootAccount[constants.ManabieSchool].Token)
	stepState.CurrentSchoolID = constants.ManabieSchool

	ctx, err := s.insertLocationTypesInDB(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	ctx, err = s.insertLocationsInDB(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	// Exists stepState.LocationIDs,  stepState.LocationTypeIDs
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) insertLocationTypesInDB(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	now := time.Now()
	randomNum := idutil.ULIDNow()

	//  [0]: branch, [1]: center
	locationTypes := []repo.LocationType{
		{
			LocationTypeID:       pgtype.Text{String: fmt.Sprintf("location-type-id-2-%s", randomNum), Status: pgtype.Present},
			Name:                 pgtype.Text{String: fmt.Sprintf("branch-%s", randomNum), Status: pgtype.Present},
			DisplayName:          pgtype.Text{String: fmt.Sprintf("display-branch-%s ", randomNum), Status: pgtype.Present},
			ParentName:           pgtype.Text{String: "Manabie Org", Status: pgtype.Present},
			ParentLocationTypeID: pgtype.Text{String: helpers.ManabieOrgLocationType, Status: pgtype.Present},
			IsArchived:           pgtype.Bool{Bool: false, Status: pgtype.Present},
			Level:                pgtype.Int4{Int: 1, Status: pgtype.Present},
		},
		{
			LocationTypeID:       pgtype.Text{String: fmt.Sprintf("location-type-id-3-%s", randomNum), Status: pgtype.Present},
			Name:                 pgtype.Text{String: fmt.Sprintf("center-%s", randomNum), Status: pgtype.Present},
			DisplayName:          pgtype.Text{String: fmt.Sprintf("display-center-%s ", randomNum), Status: pgtype.Present},
			ParentName:           pgtype.Text{String: fmt.Sprintf("branch-%s", randomNum), Status: pgtype.Present},
			ParentLocationTypeID: pgtype.Text{String: fmt.Sprintf("location-type-id-2-%s", randomNum), Status: pgtype.Present},
			IsArchived:           pgtype.Bool{Bool: false, Status: pgtype.Present},
			Level:                pgtype.Int4{Int: 3, Status: pgtype.Present},
		},
	}
	locationTypeIDs := make([]string, 0)
	for _, lt := range locationTypes {
		stmt := `INSERT INTO location_types (location_type_id, name, display_name, parent_name, parent_location_type_id, is_archived, level, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) ON CONFLICT DO NOTHING `
		_, err := s.BobDBTrace.Exec(ctx, stmt,
			lt.LocationTypeID,
			lt.Name,
			lt.DisplayName,
			lt.ParentName,
			lt.ParentLocationTypeID,
			lt.IsArchived,
			lt.Level,
			now,
			now,
		)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("cannot insert location types with `id:%s`, %v", lt.LocationTypeID.String, err)
		}
		locationTypeIDs = append(locationTypeIDs, lt.LocationTypeID.String)
	}

	stepState.Random = randomNum
	stepState.LocationTypeIDs = locationTypeIDs

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) insertLocationsInDB(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var (
		err       error
		randomNum string

		// [0]: branch, [1]: center
		locationTypeIDs      = stepState.LocationTypeIDs
		branchLocationTypeID = locationTypeIDs[0]
		centerLocationTypeID = locationTypeIDs[1]
		branchLocationID     = fmt.Sprintf("branch-location-id-2-%s", randomNum)
		centerLocationID1    = fmt.Sprintf("center-location-id-3-%s", randomNum)
		centerLocationID2    = fmt.Sprintf("center-location-id-4-%s", randomNum)
	)
	if stepState.Random != "" {
		randomNum = stepState.Random
	} else {
		randomNum = idutil.ULIDNow()
	}
	now := time.Now()
	locations := []repo.Location{
		{
			LocationID:       pgtype.Text{String: branchLocationID, Status: pgtype.Present},
			Name:             pgtype.Text{String: fmt.Sprintf("Payment-Branch-2-%s", randomNum), Status: pgtype.Present},
			LocationType:     pgtype.Text{String: branchLocationTypeID, Status: pgtype.Present},
			ParentLocationID: pgtype.Text{String: constants.ManabieOrgLocation, Status: pgtype.Present},
			IsArchived:       pgtype.Bool{Bool: false, Status: pgtype.Present},
			ResourcePath:     pgtype.Text{String: fmt.Sprintf("%d", constants.ManabieSchool), Status: pgtype.Present},
			AccessPath:       pgtype.Text{String: fmt.Sprintf("%s/%s", constants.ManabieOrgLocation, branchLocationID), Status: pgtype.Present},
		},
		{
			LocationID:       pgtype.Text{String: centerLocationID1, Status: pgtype.Present},
			LocationType:     pgtype.Text{String: centerLocationTypeID, Status: pgtype.Present},
			Name:             pgtype.Text{String: fmt.Sprintf("Payment-Center-3-%s", randomNum), Status: pgtype.Present},
			ParentLocationID: pgtype.Text{String: branchLocationID, Status: pgtype.Present},
			IsArchived:       pgtype.Bool{Bool: false, Status: pgtype.Present},
			ResourcePath:     pgtype.Text{String: fmt.Sprintf("%d", constants.ManabieSchool), Status: pgtype.Present},
			AccessPath:       pgtype.Text{String: fmt.Sprintf("%s/%s/%s", constants.ManabieOrgLocation, branchLocationID, centerLocationID1), Status: pgtype.Present},
		},
		{
			LocationID:       pgtype.Text{String: centerLocationID2, Status: pgtype.Present},
			LocationType:     pgtype.Text{String: centerLocationTypeID, Status: pgtype.Present},
			Name:             pgtype.Text{String: fmt.Sprintf("Payment-Center-4-%s", randomNum), Status: pgtype.Present},
			ParentLocationID: pgtype.Text{String: branchLocationID, Status: pgtype.Present},
			IsArchived:       pgtype.Bool{Bool: false, Status: pgtype.Present},
			ResourcePath:     pgtype.Text{String: fmt.Sprintf("%d", constants.ManabieSchool), Status: pgtype.Present},
			AccessPath:       pgtype.Text{String: fmt.Sprintf("%s/%s/%s", constants.ManabieOrgLocation, branchLocationID, centerLocationID2), Status: pgtype.Present},
		},
	}
	locationIDs := make([]string, 0)
	for _, l := range locations {
		stmt := `INSERT INTO locations (location_id, name, location_type, parent_location_id, is_archived, resource_path,access_path, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) ON CONFLICT DO NOTHING`
		_, err = s.BobDBTrace.Exec(ctx, stmt,
			l.LocationID,
			l.Name,
			l.LocationType,
			l.ParentLocationID,
			l.IsArchived,
			l.ResourcePath,
			l.AccessPath,
			now,
			now,
		)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("cannot insert locations with `id:%s`, %v", l.LocationID.String, err)
		}
		locationIDs = append(locationIDs, l.LocationID.String)
	}

	stepState.LocationIDs = locationIDs
	stepState.LowestLevelLocationIDs = []string{centerLocationID1, centerLocationID2}

	time.Sleep(3000)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) getAdmin(ctx context.Context, userID string) error {
	user := &entities.User{}
	userFieldNames, userFieldValues := user.FieldMap()
	stmt :=
		`
		SELECT %s
		FROM 
			%s
		WHERE 
			user_id = $1
		`
	stmt = fmt.Sprintf(
		stmt,
		strings.Join(userFieldNames, ","),
		user.TableName(),
	)
	row := s.FatimaDBTrace.QueryRow(ctx, stmt, userID)
	err := row.Scan(userFieldValues...)
	if err != nil {
		return err
	}
	return nil
}

func (s *suite) insertAdmin(ctx context.Context, userID string, name string) error {
	adminName := database.Text(fmt.Sprintf("Admin for create order %s", name))
	stmt := `INSERT INTO users
		(user_id, name, user_group, country, updated_at, created_at)
		VALUES ($1, $2, $3, $4, now(), now());`
	_, err := s.FatimaDBTrace.Exec(ctx, stmt, userID, adminName, cpb.UserGroup_USER_GROUP_SCHOOL_ADMIN.String(), "COUNTRY_VN")
	return err
}
