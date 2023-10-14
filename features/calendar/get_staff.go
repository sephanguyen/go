package calendar

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/features/helper"
	"github.com/manabie-com/backend/internal/calendar/domain/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	userRepository "github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	userEntity "github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	userConstant "github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	cpb "github.com/manabie-com/backend/pkg/manabuf/calendar/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"golang.org/x/exp/slices"
)

func (s *suite) getListStaffByLocation(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := &cpb.GetStaffsByLocationRequest{
		LocationId: stepState.LocationID,
	}

	stepState.Response, stepState.ResponseErr = cpb.NewUserReaderServiceClient(s.CalendarConn).
		GetStaffsByLocation(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aListCorrectStaffReturned(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	resp := stepState.Response.(*cpb.GetStaffsByLocationResponse)
	// expectedUserIDs := stepState.UserIDs

	// actualUserIDs := sliceutils.FilterWithReferenceList(
	// 	expectedUserIDs,
	// 	resp.Staffs,
	// 	func(expectedUserIDs []string, staff *cpb.GetStaffsByLocationResponse_StaffInfo) bool {
	// 		return slices.Contains(expectedUserIDs, staff.Id)
	// 	},
	// )

	// if len(actualUserIDs) != len(expectedUserIDs) {
	// 	return StepStateToContext(ctx, stepState), fmt.Errorf("the expected user IDs (%v) are not in the returned user IDs", expectedUserIDs)
	// }
	if len(resp.Staffs) == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("there are no staff returned")
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) anEmptyListStaffReturned(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	resp := stepState.Response.(*cpb.GetStaffsByLocationResponse)
	expectedUserIDs := stepState.UserIDs

	actualUserIDs := sliceutils.FilterWithReferenceList(
		expectedUserIDs,
		resp.Staffs,
		func(expectedUserIDs []string, staff *cpb.GetStaffsByLocationResponse_StaffInfo) bool {
			return slices.Contains(expectedUserIDs, staff.Id)
		},
	)

	if len(actualUserIDs) != 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("staff without user groups that are not expected to be returned are found")
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aListOfStaffCreated(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	numberOfStaff := 7
	numberOfValidStaff := 5
	userIDs := make([]string, 0, numberOfStaff)
	locationID := stepState.LocationID
	for i := 0; i < numberOfStaff; i++ {
		userID := idutil.ULIDNow()
		if err := s.aValidUser(ctx, withID(userID), withRole(userConstant.UserGroupAdmin)); err != nil {
			return nil, fmt.Errorf("aValidStaffInDB. s.aValidUser: %w", err)
		}
		userIDs = append(userIDs, userID)
	}

	// insert user_access_path
	query := `INSERT INTO user_access_paths (
		user_id,
		location_id,
		created_at,
		updated_at
	)
	VALUES ($1,$2,now(), now())`
	bUap := &pgx.Batch{}
	for i := 0; i < numberOfStaff; i++ {
		bUap.Queue(query, userIDs[i], locationID)
	}
	batchUcp := s.BobDBTrace.SendBatch(ctx, bUap)
	defer batchUcp.Close()
	for i := 0; i < bUap.Len(); i++ {
		_, err := batchUcp.Exec()
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("batchResults.Exec():%w", err)
		}
	}

	// insert staff
	stmt := `INSERT INTO staff (
		staff_id,
		working_status,
		created_at,
		updated_at
	)
	VALUES ($1, $2, now(), now())`
	b := &pgx.Batch{}
	status := string(constants.Available)
	for i := 0; i < numberOfStaff; i++ {
		if i >= numberOfValidStaff {
			status = string(constants.Resigned)
		}
		b.Queue(stmt, userIDs[i], status)
	}
	batchStaff := s.BobDBTrace.SendBatch(ctx, b)
	defer batchStaff.Close()
	for i := 0; i < b.Len(); i++ {
		cmdTag, err := batchStaff.Exec()
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("batchResults.Exec():%w", err)
		}
		if cmdTag.RowsAffected() == 0 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("failed to create staff")
		}
	}
	stepState.UserIDs = userIDs[:numberOfValidStaff]
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aListOfStaffWithUserGroupCreated(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	numberOfStaff := 7
	numberOfValidStaff := 5
	locationID := stepState.LocationID
	userIDs := make([]string, 0, numberOfStaff)
	for i := 0; i < numberOfStaff; i++ {
		userID := idutil.ULIDNow()
		if err := s.aValidUser(ctx, withID(userID), withRole(userConstant.UserGroupAdmin)); err != nil {
			return nil, fmt.Errorf("aValidStaffInDB. s.aValidUser: %w", err)
		}
		userIDs = append(userIDs, userID)
	}

	// insert user_access_path
	query := `INSERT INTO user_access_paths (
		user_id,
		location_id,
		created_at,
		updated_at
	)
	VALUES ($1,$2,now(), now())`
	bUap := &pgx.Batch{}
	for i := 0; i < numberOfStaff; i++ {
		bUap.Queue(query, userIDs[i], locationID)
	}
	batchUcp := s.BobDBTrace.SendBatch(ctx, bUap)
	defer batchUcp.Close()
	for i := 0; i < bUap.Len(); i++ {
		_, err := batchUcp.Exec()
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("batchResults.Exec():%w", err)
		}
	}

	// insert staff
	stmt := `INSERT INTO staff (
		staff_id,
		working_status,
		created_at,
		updated_at
	)
	VALUES ($1, $2, now(), now())`
	b := &pgx.Batch{}
	status := string(constants.Available)
	for i := 0; i < numberOfStaff; i++ {
		if i >= numberOfValidStaff {
			status = string(constants.Resigned)
		}
		b.Queue(stmt, userIDs[i], status)
	}
	batchStaff := s.BobDBTrace.SendBatch(ctx, b)
	defer batchStaff.Close()
	for i := 0; i < b.Len(); i++ {
		cmdTag, err := batchStaff.Exec()
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("batchResults.Exec():%w", err)
		}
		if cmdTag.RowsAffected() == 0 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("failed to create staff")
		}
	}
	stepState.UserIDs = userIDs[:numberOfValidStaff]
	userGroupID := idutil.ULIDNow()
	userGroupRepo := &userRepository.UserGroupV2Repo{}
	err := userGroupRepo.Create(ctx, s.BobPostgresDB, &userEntity.UserGroupV2{
		UserGroupID: pgtype.Text{
			String: userGroupID,
			Status: pgtype.Present,
		},
		UserGroupName: pgtype.Text{
			String: userGroupID,
			Status: pgtype.Present,
		},
		ResourcePath: pgtype.Text{
			String: stepState.ResourcePath,
			Status: pgtype.Present,
		},
		OrgLocationID: pgtype.Text{
			String: locationID,
			Status: pgtype.Present,
		},
		IsSystem: pgtype.Bool{
			Bool:   false,
			Status: pgtype.Present,
		},
	})
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("cant create user group: %w", err)
	}
	roleID, err := s.getFirstRoleID(ctx, "Teacher")
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	grantedRoleID := idutil.ULIDNow()
	grantedRoleRepo := &userRepository.GrantedRoleRepo{}
	grantedRole := &userEntity.GrantedRole{
		GrantedRoleID: pgtype.Text{
			String: grantedRoleID,
			Status: pgtype.Present,
		},
		RoleID: pgtype.Text{
			String: roleID,
			Status: pgtype.Present,
		},
		UserGroupID: pgtype.Text{
			String: userGroupID,
			Status: pgtype.Present,
		},
		ResourcePath: pgtype.Text{
			String: stepState.ResourcePath,
			Status: pgtype.Present,
		},
	}
	err = grantedRoleRepo.Create(ctx, s.BobPostgresDB, grantedRole)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("cant create granted role: %w", err)
	}
	err = grantedRoleRepo.LinkGrantedRoleToAccessPath(ctx, s.BobPostgresDB, grantedRole, []string{locationID})
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("cant link granted role to access path: %w", err)
	}
	userGroupMemberRepo := &userRepository.UserGroupsMemberRepo{}
	for _, userID := range stepState.UserIDs {
		err = userGroupMemberRepo.UpsertBatch(ctx, s.BobPostgresDB, []*userEntity.UserGroupMember{
			{
				UserID: pgtype.Text{
					String: userID,
					Status: pgtype.Present,
				},
				UserGroupID: pgtype.Text{
					String: userGroupID,
					Status: pgtype.Present,
				},
				CreatedAt: database.Timestamptz(time.Now()),
				UpdatedAt: database.Timestamptz(time.Now()),
				DeletedAt: pgtype.Timestamptz{Status: pgtype.Null},
				ResourcePath: pgtype.Text{
					String: stepState.ResourcePath,
					Status: pgtype.Present,
				},
			},
		})
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("cant insert user to user group: %w", err)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) getFirstRoleID(ctx context.Context, roleName string) (string, error) {
	stepState := StepStateFromContext(ctx)
	query := `SELECT role_id FROM role WHERE role_name = ANY($1) and resource_path = $2 and deleted_at IS NULL LIMIT 1`
	rows, err := s.BobDBTrace.Query(ctx, query, []string{roleName}, stepState.ResourcePath)
	if err != nil {
		return "", fmt.Errorf("generateRoleWithLocation: find Role IDs: %w", err)
	}
	defer rows.Close()

	roleID := ""
	for rows.Next() {
		if err = rows.Scan(&roleID); err != nil {
			return "", err
		}
	}
	if len(roleID) == 0 {
		roleID = idutil.ULIDNow()
		roleRepo := userRepository.RoleRepo{}
		err = roleRepo.Create(ctx, s.BobPostgresDB, &userEntity.Role{
			RoleID: pgtype.Text{
				String: roleID,
				Status: pgtype.Present,
			},
			RoleName: pgtype.Text{
				String: "Teacher",
				Status: pgtype.Present,
			},
			ResourcePath: pgtype.Text{
				String: stepState.ResourcePath,
				Status: pgtype.Present,
			},
			IsSystem: pgtype.Bool{
				Bool:   false,
				Status: pgtype.Present,
			},
		})
		if err != nil {
			return "", err
		}
	}
	return roleID, nil
}
