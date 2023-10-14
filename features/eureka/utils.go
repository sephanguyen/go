package eureka

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/manabie-com/backend/features/helper"
	bobEnt "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/bob/repositories"
	consta "github.com/manabie-com/backend/internal/entryexitmgmt/constant"
	eureka "github.com/manabie-com/backend/internal/eureka/constants"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/yasuo/constant"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"go.uber.org/multierr"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func appendContextWithToken(ctx context.Context, token string) context.Context {
	return metadata.AppendToOutgoingContext(contextWithValidVersion(ctx), "token", token)
}

type userOption func(u *bobEnt.User)

func withID(id string) userOption {
	return func(u *bobEnt.User) {
		u.ID = database.Text(id)
	}
}

func withRole(group string) userOption {
	return func(u *bobEnt.User) {
		u.Group = database.Text(group)
	}
}

//nolint:interfacer
func generateExchangeToken(userID, userGroup string, schoolID int64, shamirConn *grpc.ClientConn) (string, error) {
	firebaseToken, err := generateValidAuthenticationToken(userID, userGroup)
	if err != nil {
		return "", err
	}
	token, err := helper.ExchangeToken(firebaseToken, userID, userGroup, applicantID, schoolID, shamirConn)
	if err != nil {
		return "", err
	}
	return token, nil
}

// signInAs generates a new id and token for the input role.
func (s *suite) signedInAs(ctx context.Context, role string) (context.Context, string, string, error) {
	stepState := StepStateFromContext(ctx)
	if role == "" {
		return StepStateToContext(ctx, stepState), "", "", fmt.Errorf("role must be not nil")
	}
	id := idutil.ULIDNow()
	stepState.UserId = id
	ctx = s.setFakeClaimToContext(ctx, stepState.SchoolID, role)
	ctx, err := s.aValidUser(ctx, id, role)
	if err != nil {
		return StepStateToContext(ctx, stepState), "", "", err
	}
	var oldGroup string
	switch role {
	case consta.RoleTeacher:
		oldGroup = "USER_GROUP_TEACHER"
	case consta.RoleStudent:
		oldGroup = "USER_GROUP_STUDENT"
	case consta.RoleSchoolAdmin:
		oldGroup = "USER_GROUP_SCHOOL_ADMIN"
	case consta.RoleParent:
		oldGroup = "USER_GROUP_PARENT"
	default:
		oldGroup = "USER_GROUP_STUDENT"
	}
	authToken, err := s.generateExchangeToken(id, oldGroup)
	if err != nil {
		return StepStateToContext(ctx, stepState), "", "", err
	}
	return StepStateToContext(ctx, stepState), id, authToken, nil
}

func (s *suite) validUser(ctx context.Context, opts ...userOption) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	num := rand.Int()

	u := &bobEnt.User{}
	database.AllNullEntity(u)

	resourcePath := golibs.ResourcePathFromCtx(ctx)
	err := multierr.Combine(
		u.ID.Set(idutil.ULIDNow()),
		u.LastName.Set(fmt.Sprintf("valid-user-%d", num)),
		u.PhoneNumber.Set(fmt.Sprintf("+848%d", num)),
		u.Email.Set(fmt.Sprintf("valid-user-%d@email.com", num)),
		u.Avatar.Set(fmt.Sprintf("http://valid-user-%d", num)),
		u.Country.Set(pb.COUNTRY_VN.String()),
		u.Group.Set(bobEnt.UserGroupStudent),
		u.DeviceToken.Set(nil),
		u.AllowNotification.Set(true),
		u.CreatedAt.Set(time.Now()),
		u.UpdatedAt.Set(time.Now()),
		u.IsTester.Set(nil),
		u.ResourcePath.Set(resourcePath),
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	for _, opt := range opts {
		opt(u)
	}
	cmdTag, err := database.Insert(ctx, u, s.BobDB.Exec)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if cmdTag.RowsAffected() == 0 {
		return StepStateToContext(ctx, stepState), errors.New("cannot insert user for testing")
	}

	var schoolID int64
	if u.Group.String == constant.UserGroupTeacher {
		if stepState.School != nil {
			schoolID = int64(stepState.School.ID.Int)
		}
		if schoolID == 0 {
			schoolID = constants.ManabieSchool
		}
		stepState.SchoolIDInt = int32(schoolID)
		stepState.CurrentSchoolID = stepState.SchoolIDInt

		teacher := &bobEnt.Teacher{}
		database.AllNullEntity(teacher)

		err = multierr.Combine(
			teacher.ID.Set(u.ID.String),
			teacher.SchoolIDs.Set([]int64{schoolID}),
			teacher.UpdatedAt.Set(time.Now()),
			teacher.CreatedAt.Set(time.Now()),
			teacher.ResourcePath.Set(resourcePath),
		)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		_, err = database.Insert(ctx, teacher, s.BobDB.Exec)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}

	if u.Group.String == constant.UserGroupSchoolAdmin {
		schoolID := stepState.SchoolIDInt
		if schoolID == 0 {
			schoolID = constants.ManabieSchool
			stepState.SchoolIDInt = schoolID
			stepState.CurrentSchoolID = stepState.SchoolIDInt
		}
		schoolAdminAccount := &bobEnt.SchoolAdmin{}
		database.AllNullEntity(schoolAdminAccount)
		err := multierr.Combine(
			schoolAdminAccount.SchoolAdminID.Set(u.ID.String),
			schoolAdminAccount.SchoolID.Set(schoolID),
			schoolAdminAccount.UpdatedAt.Set(time.Now()),
			schoolAdminAccount.CreatedAt.Set(time.Now()),
			schoolAdminAccount.ResourcePath.Set(resourcePath),
		)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		_, err = database.Insert(ctx, schoolAdminAccount, s.BobDB.Exec)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}
	if u.Group.String == constant.UserGroupStudent {
		schoolID := stepState.SchoolIDInt
		if schoolID == 0 {
			schoolID = constants.ManabieSchool
			stepState.SchoolIDInt = schoolID
			stepState.CurrentSchoolID = stepState.SchoolIDInt
		}
		student := &bobEnt.Student{}
		database.AllNullEntity(student)
		err := multierr.Combine(
			student.ID.Set(u.ID.String),
			student.EnrollmentStatus.Set("STUDENT_ENROLLMENT_STATUS_ENROLLED"),
			student.StudentNote.Set(""),
			student.SchoolID.Set(schoolID),
			student.UpdatedAt.Set(time.Now()),
			student.CreatedAt.Set(time.Now()),
			student.OnTrial.Set(false),
			student.GivenName.Set(fmt.Sprintf("student-name-%s", u.ID.String)),
			student.BillingDate.Set(time.Now()),
			student.ResourcePath.Set(resourcePath),
		)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		_, err = database.Insert(ctx, student, s.BobDB.Exec)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}

	uGroup := &bobEnt.UserGroup{}
	database.AllNullEntity(uGroup)

	err = multierr.Combine(
		uGroup.GroupID.Set(u.Group.String),
		uGroup.UserID.Set(u.ID.String),
		uGroup.IsOrigin.Set(true),
		uGroup.Status.Set("USER_GROUP_STATUS_ACTIVE"),
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	userGroupRepo := &repositories.UserGroupRepo{}
	err = userGroupRepo.Upsert(ctx, s.BobDBTrace, uGroup)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("userGroupRepo.Upsert: %w %s", err, u.Group.String)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) ExecuteWithRetry(ctx context.Context, process func() error, waitTime time.Duration, retryTime int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var count int
	var err error
	for count <= retryTime {
		err = process()
		if err == nil {
			return StepStateToContext(ctx, stepState), err
		}
		time.Sleep(waitTime)
		count++
	}
	return StepStateToContext(ctx, stepState), err
}

func contains(s []string, target string) bool {
	for _, val := range s {
		if target == val {
			return true
		}
	}
	return false
}

func (s *suite) aSignedInStaff(ctx context.Context, roles []string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	id := idutil.ULIDNow()
	// Because we have not implemented the new roles yet,
	// here we will use admin to assign these roles
	// Please change this when we have implemented the new roles
	ctx, err := s.aValidUser(StepStateToContext(ctx, stepState), id, eureka.RoleSchoolAdmin)

	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("aValidUser: %w", err)
	}

	token, err := s.generateExchangeToken(id, constant.UserGroupSchoolAdmin)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.AuthToken = token

	createUserGroupResp, err := s.createUserGroupWithRoleNames(StepStateToContext(ctx, stepState), roles)
	if err != nil {
		return nil, err
	}

	if err := assignUserGroupToUser(ctx, s.BobDBTrace, id, []string{createUserGroupResp.UserGroupId}); err != nil {
		return nil, err
	}

	stepState.CurrentUserID = id
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) createUserGroupWithRoleNames(ctx context.Context, roleNames []string) (*upb.CreateUserGroupResponse, error) {
	req := &upb.CreateUserGroupRequest{
		UserGroupName: fmt.Sprintf("user-group_%s", idutil.ULIDNow()),
	}
	stmt := "SELECT role_id FROM role WHERE deleted_at IS NULL AND role_name = ANY($1) LIMIT $2"
	rows, err := s.BobDB.Query(ctx, stmt, roleNames, len(roleNames))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roleIDs []string
	for rows.Next() {
		roleID := ""
		if err := rows.Scan(&roleID); err != nil {
			return nil, fmt.Errorf("rows.Err: %w", err)
		}
		roleIDs = append(roleIDs, roleID)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}

	for _, roleID := range roleIDs {
		req.RoleWithLocations = append(
			req.RoleWithLocations,
			&upb.RoleWithLocations{
				RoleId:      roleID,
				LocationIds: []string{constants.ManabieOrgLocation},
			},
		)
	}

	resp, err := upb.NewUserGroupMgmtServiceClient(s.UsermgmtConn).CreateUserGroup(contextWithToken(s, ctx), req)
	if err != nil {
		return nil, fmt.Errorf("createUserGroupWithRoleNames: %w", err)
	}

	return resp, nil
}

func assignUserGroupToUser(ctx context.Context, dbBob database.QueryExecer, userID string, userGroupIDs []string) error {
	userGroupMembers := make([]*entity.UserGroupMember, 0)
	for _, userGroupID := range userGroupIDs {
		userGroupMem := &entity.UserGroupMember{}
		database.AllNullEntity(userGroupMem)
		if err := multierr.Combine(
			userGroupMem.UserID.Set(userID),
			userGroupMem.UserGroupID.Set(userGroupID),
			userGroupMem.ResourcePath.Set(fmt.Sprint(constants.ManabieSchool)),
		); err != nil {
			return err
		}
		userGroupMembers = append(userGroupMembers, userGroupMem)
	}

	if err := (&repository.UserGroupsMemberRepo{}).UpsertBatch(ctx, dbBob, userGroupMembers); err != nil {
		return err
	}
	return nil
}
