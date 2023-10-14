package common

import (
	"context"
	"fmt"
	"strconv"

	"github.com/manabie-com/backend/features/conversationmgmt/common/helpers"
	"github.com/manabie-com/backend/internal/bob/constants"
	golibs_auth "github.com/manabie-com/backend/internal/golibs/auth"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/location/domain"
	location_repo "github.com/manabie-com/backend/internal/mastermgmt/modules/location/infrastructure/repo"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
)

func (s *ConversationMgmtSuite) signedAsAccount(ctx context.Context, group string, orgID int64, grantedLocationIDs []string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	var (
		userGroup string
		err       error
	)

	switch group {
	case unauthenticatedType:
		stepState.AuthToken = "random-token"
		return StepStateToContext(ctx, stepState), nil
	case helpers.StaffGrantedRoleSchoolAdmin:
		stepState.CurrentGrandtedRoles = []string{constant.RoleSchoolAdmin}
		return s.aSignedInStaff(ctx, stepState.CurrentGrandtedRoles, orgID, grantedLocationIDs)
	case helpers.StaffGrantedRoleTeacher:
		stepState.CurrentGrandtedRoles = []string{constant.RoleTeacher}
		return s.aSignedInStaff(ctx, stepState.CurrentGrandtedRoles, orgID, grantedLocationIDs)
	case helpers.StaffGrantedRoleTeacherLead:
		stepState.CurrentGrandtedRoles = []string{constant.RoleTeacherLead}
		return s.aSignedInStaff(ctx, stepState.CurrentGrandtedRoles, orgID, grantedLocationIDs)
	case helpers.StaffGrantedRoleHQStaff:
		stepState.CurrentGrandtedRoles = []string{constant.RoleHQStaff}
		return s.aSignedInStaff(ctx, stepState.CurrentGrandtedRoles, orgID, grantedLocationIDs)
	case helpers.StaffGrantedRoleCentreLead:
		stepState.CurrentGrandtedRoles = []string{constant.RoleCentreLead}
		return s.aSignedInStaff(ctx, stepState.CurrentGrandtedRoles, orgID, grantedLocationIDs)
	case helpers.StaffGrantedRoleCentreManager:
		stepState.CurrentGrandtedRoles = []string{constant.RoleCentreManager}
		return s.aSignedInStaff(ctx, stepState.CurrentGrandtedRoles, orgID, grantedLocationIDs)
	case helpers.StaffGrantedRoleCentreStaff:
		stepState.CurrentGrandtedRoles = []string{constant.RoleCentreStaff}
		return s.aSignedInStaff(ctx, stepState.CurrentGrandtedRoles, orgID, grantedLocationIDs)
	case studentType:
		userGroup = constant.UserGroupStudent
	case teacherType:
		userGroup = constant.UserGroupTeacher
	case schoolAdminType:
		userGroup = constant.UserGroupSchoolAdmin
	case parentType:
		userGroup = constant.UserGroupParent
	case organizationType:
		userGroup = constant.UserGroupOrganizationManager
	}

	id := idutil.ULIDNow()
	stepState.CurrentUserID = id
	stepState.CurrentUserGroup = userGroup

	if ctx, err = s.aValidUser(StepStateToContext(ctx, stepState), withID(id), withRole(userGroup)); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	token, err := s.generateExchangeToken(id, userGroup, orgID)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.generateExchangeToken: %v", err)
	}
	stepState.AuthToken = token

	return StepStateToContext(ctx, stepState), nil
}

func (s *ConversationMgmtSuite) aValidUser(ctx context.Context, opts ...userOption) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	orgID := stepState.CurrentOrganicationID
	if orgID == 0 {
		orgID = constants.ManabieSchool
	}
	ctx = golibs_auth.InjectFakeJwtToken(ctx, fmt.Sprint(orgID))

	user, err := newUserEntity()
	if err != nil {
		return StepStateToContext(ctx, stepState), errors.Wrap(err, "newUserEntity")
	}

	for _, opt := range opts {
		opt(user)
	}

	err = database.ExecInTx(ctx, s.BobDBConn, func(ctx context.Context, tx pgx.Tx) error {
		userRepo := repository.UserRepo{}
		err := userRepo.Create(ctx, tx, user)
		if err != nil {
			return fmt.Errorf("cannot create user: %w", err)
		}

		switch user.Group.String {
		case constant.UserGroupTeacher:
			teacherRepo := repository.TeacherRepo{}
			t := &entity.Teacher{}
			database.AllNullEntity(t)
			t.ID = user.ID
			err := multierr.Combine(
				t.SchoolIDs.Set([]int32{orgID}),
				t.ResourcePath.Set(fmt.Sprint(orgID)),
			)
			if err != nil {
				return err
			}

			err = teacherRepo.CreateMultiple(ctx, tx, []*entity.Teacher{t})
			if err != nil {
				return fmt.Errorf("cannot create teacher: %w", err)
			}
		case constant.UserGroupSchoolAdmin:
			schoolAdminRepo := repository.SchoolAdminRepo{}
			schoolAdminAccount := &entity.SchoolAdmin{}
			database.AllNullEntity(schoolAdminAccount)
			err := multierr.Combine(
				schoolAdminAccount.SchoolAdminID.Set(user.ID.String),
				schoolAdminAccount.SchoolID.Set(orgID),
				schoolAdminAccount.ResourcePath.Set(database.Text(fmt.Sprint(orgID))),
			)
			if err != nil {
				return fmt.Errorf("cannot create school admin: %w", err)
			}
			err = schoolAdminRepo.CreateMultiple(ctx, tx, []*entity.SchoolAdmin{schoolAdminAccount})
			if err != nil {
				return err
			}
		case constant.UserGroupParent:
			parentRepo := repository.ParentRepo{}
			parentEnt := &entity.Parent{}
			database.AllNullEntity(parentEnt)
			err := multierr.Combine(
				parentEnt.ID.Set(user.ID.String),
				parentEnt.SchoolID.Set(orgID),
				parentEnt.ResourcePath.Set(fmt.Sprint(orgID)),
			)
			if err != nil {
				return err
			}
			err = parentRepo.CreateMultiple(ctx, tx, []*entity.Parent{parentEnt})
			if err != nil {
				return fmt.Errorf("cannot create parent: %w", err)
			}
		case constant.UserGroupStudent:
			studentRepo := repository.StudentRepo{}
			student, err := newStudentEntity()
			if err != nil {
				return err
			}
			err = multierr.Combine(
				student.ID.Set(user.ID.String),
				student.SchoolID.Set(orgID),
				student.ResourcePath.Set(fmt.Sprint(orgID)),
			)
			if err != nil {
				return err
			}
			err = studentRepo.CreateMultiple(ctx, tx, []*entity.LegacyStudent{student})
			if err != nil {
				return fmt.Errorf("cannot create student: %w", err)
			}
		}

		return nil
	})
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	uGroup := &entity.UserGroup{}
	database.AllNullEntity(uGroup)

	err = multierr.Combine(
		uGroup.GroupID.Set(user.Group.String),
		uGroup.UserID.Set(user.ID.String),
		uGroup.IsOrigin.Set(true),
		uGroup.Status.Set("USER_GROUP_STATUS_ACTIVE"),
		uGroup.ResourcePath.Set(database.Text(fmt.Sprint(orgID))),
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	userGroupRepo := &repository.UserGroupRepo{}
	err = userGroupRepo.Upsert(ctx, s.BobDBConn, uGroup)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("userGroupRepo.Upsert: %w %s", err, user.Group.String)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *ConversationMgmtSuite) aSignedInStaff(ctx context.Context, roles []string, orgID int64, grantedLocationIDs []string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	id := idutil.ULIDNow()
	ctx, err := s.aValidUser(StepStateToContext(ctx, stepState), withID(id), withRole(constant.UserGroupSchoolAdmin))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("aValidUser: %w", err)
	}

	var token string
	err = doRetry(func() (bool, error) {
		token, err = s.generateExchangeToken(id, constant.UserGroupSchoolAdmin, orgID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return true, fmt.Errorf("retrying %+v", err)
			}
			return false, fmt.Errorf("s.generateExchangeToken: %v", err)
		}
		return false, nil
	})
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.AuthToken = token
	ctx = StepStateToContext(ctx, stepState)

	userGroupID, err := s.createUserGroupWithRoleNames(ctx, roles, grantedLocationIDs, orgID)
	if err != nil {
		return ctx, err
	}

	if err := assignUserGroupToUser(ctx, s.BobDBConn, id, []string{userGroupID}, strconv.Itoa(int(orgID))); err != nil {
		return ctx, err
	}

	stepState.CurrentUserID = id
	return StepStateToContext(ctx, stepState), nil
}

func (s *ConversationMgmtSuite) createUserGroupWithRoleNames(ctx context.Context, roleNames []string, grantedLocationIDs []string, resourcePath int64) (string, error) {
	req := &upb.CreateUserGroupRequest{
		UserGroupName: fmt.Sprintf("user-group_%s", idutil.ULIDNow()),
	}

	stmt := "SELECT role_id FROM role WHERE deleted_at IS NULL AND role_name = ANY($1) LIMIT $2"
	rows, err := s.BobDBConn.Query(ctx, stmt, roleNames, len(roleNames))
	if err != nil {
		return "", err
	}
	defer rows.Close()

	var roleIDs []string
	for rows.Next() {
		roleID := ""
		if err := rows.Scan(&roleID); err != nil {
			return "", fmt.Errorf("rows.Err: %w", err)
		}
		roleIDs = append(roleIDs, roleID)
	}
	if err := rows.Err(); err != nil {
		return "", fmt.Errorf("rows.Err: %w", err)
	}

	if err != nil {
		return "", fmt.Errorf("s.getOrgLocationID: %w", err)
	}

	for _, roleID := range roleIDs {
		req.RoleWithLocations = append(
			req.RoleWithLocations,
			&upb.RoleWithLocations{
				RoleId:      roleID,
				LocationIds: grantedLocationIDs,
			},
		)
	}

	locationRepo := &location_repo.LocationRepo{}
	userGroupV2Repo := repository.UserGroupV2Repo{}
	grantedRoleRepo := repository.GrantedRoleRepo{}

	var userGroup *entity.UserGroupV2
	if err = database.ExecInTx(ctx, s.BobPostgresDBConn, func(ctx context.Context, tx pgx.Tx) error {
		orgLocation, err := locationRepo.GetLocationOrg(ctx, tx, fmt.Sprint(resourcePath))
		if err != nil {
			return fmt.Errorf("locationRepo.GetLocationOrg: %w", err)
		}

		// convert payload to entity
		if userGroup, err = userGroupPayloadToUserGroupEnt(req, fmt.Sprint(resourcePath), orgLocation); err != nil {
			return fmt.Errorf("s.UserGroupPayloadToUserGroupEnts: %w", err)
		}

		// create user group first
		if err = userGroupV2Repo.Create(ctx, tx, userGroup); err != nil {
			return fmt.Errorf("userGroupV2Repo.Create: %w", err)
		}

		var grantedRole *entity.GrantedRole
		for _, roleWithLocations := range req.RoleWithLocations {
			// convert payload to entity
			if grantedRole, err = roleWithLocationsPayloadToGrantedRole(roleWithLocations, userGroup.UserGroupID.String, fmt.Sprint(resourcePath)); err != nil {
				return fmt.Errorf("s.RoleWithLocationsPayloadToGrantedRole: %w", err)
			}
			// create granted_role
			if err = grantedRoleRepo.Create(ctx, tx, grantedRole); err != nil {
				return fmt.Errorf("grantedRoleRepo.Create: %w", err)
			}

			// link granted_role to access path(by location ids)
			if err = grantedRoleRepo.LinkGrantedRoleToAccessPath(ctx, tx, grantedRole, roleWithLocations.LocationIds); err != nil {
				return fmt.Errorf("grantedRoleRepo.LinkGrantedRoleToAccessPath: %w", err)
			}
		}
		return nil
	}); err != nil {
		return "", fmt.Errorf("database.ExecInTx: %w", err)
	}

	return userGroup.UserGroupID.String, nil
}

func userGroupPayloadToUserGroupEnt(payload *upb.CreateUserGroupRequest, resourcePath string, orgLocation *domain.Location) (*entity.UserGroupV2, error) {
	userGroup := &entity.UserGroupV2{}
	database.AllNullEntity(userGroup)
	if err := multierr.Combine(
		userGroup.UserGroupID.Set(idutil.ULIDNow()),
		userGroup.UserGroupName.Set(payload.UserGroupName),
		userGroup.ResourcePath.Set(resourcePath),
		userGroup.OrgLocationID.Set(orgLocation.LocationID),
		userGroup.IsSystem.Set(false),
	); err != nil {
		return nil, fmt.Errorf("set user group fail: %w", err)
	}

	return userGroup, nil
}

func roleWithLocationsPayloadToGrantedRole(payload *upb.RoleWithLocations, userGroupID string, resourcePath string) (*entity.GrantedRole, error) {
	grantedRole := &entity.GrantedRole{}
	database.AllNullEntity(grantedRole)
	if err := multierr.Combine(
		grantedRole.GrantedRoleID.Set(idutil.ULIDNow()),
		grantedRole.UserGroupID.Set(userGroupID),
		grantedRole.RoleID.Set(payload.RoleId),
		grantedRole.ResourcePath.Set(resourcePath),
	); err != nil {
		return nil, fmt.Errorf("set granted role fail: %w", err)
	}

	return grantedRole, nil
}

func (s *ConversationMgmtSuite) GenerateIndividualIDs(ctx context.Context, isGeneric bool, numStudentNew int) ([]string, error) {
	ctx, err := s.CreatesNumberOfStudentsWithParentsInfo(ctx, fmt.Sprint(numStudentNew), "1")
	if err != nil {
		return nil, fmt.Errorf("failed CreatesNumberOfStudentsWithParentsInfo: %v", err)
	}
	stepState := StepStateFromContext(ctx)
	userIDs := []string{}

	if isGeneric {
		// only use newly added user id
		for idx, student := range stepState.Students {
			if idx >= len(stepState.Students)-numStudentNew {
				userIDs = append(userIDs, student.ID)
				for _, parent := range student.Parents {
					userIDs = append(userIDs, parent.ID)
				}
			}
		}
	} else {
		// case for Nats requests
		// Only add new students for individual list
		for idx, student := range stepState.Students {
			if idx >= len(stepState.Students)-numStudentNew {
				userIDs = append(userIDs, student.ID)
			}
		}
	}

	return userIDs, nil
}
