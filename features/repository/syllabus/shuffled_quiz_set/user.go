package shuffled_quiz_set

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/manabie-com/backend/features/repository/syllabus/utils"
	"github.com/manabie-com/backend/internal/bob/constants"
	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/yasuo/constant"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
)

func (s *Suite) aValidUserInDB(ctx context.Context, dbConn *database.DBTrace, id, newgroup, oldGroup string) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	num := rand.Int()
	var now pgtype.Timestamptz
	now.Set(time.Now())
	u := entities.User{}
	database.AllNullEntity(&u)
	u.ID = database.Text(id)
	u.LastName.Set(fmt.Sprintf("valid-user-%d", num))
	u.PhoneNumber.Set(fmt.Sprintf("+848%d", num))
	u.Email.Set(fmt.Sprintf("valid-user-%d@email.com", num))
	u.Avatar.Set(fmt.Sprintf("http://valid-user-%d", num))
	u.Country.Set(cpb.Country_COUNTRY_VN.String())
	u.Group.Set(oldGroup)
	u.DeviceToken.Set(nil)
	u.AllowNotification.Set(true)
	u.CreatedAt = now
	u.UpdatedAt = now
	u.IsTester.Set(nil)
	u.FacebookID.Set(nil)
	u.ResourcePath.Set(fmt.Sprint(constants.ManabieSchool))

	gr := &entities.Group{}
	database.AllNullEntity(gr)
	gr.ID.Set(oldGroup)
	gr.Name.Set(oldGroup)
	gr.UpdatedAt.Set(time.Now())
	gr.CreatedAt.Set(time.Now())
	fieldNames, _ := gr.FieldMap()
	placeHolders := database.GeneratePlaceholders(len(fieldNames))
	stmt := fmt.Sprintf("INSERT INTO groups (%s) VALUES(%s) ON CONFLICT DO NOTHING", strings.Join(fieldNames, ","), placeHolders)
	if _, err := dbConn.Exec(ctx, stmt, database.GetScanFields(gr, fieldNames)...); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("insert group error: %v", err)
	}
	ctx = s.setFakeClaimToContext(context.Background(), u.ResourcePath.String, oldGroup)

	ugroup := &entity.UserGroupV2{}
	database.AllNullEntity(ugroup)
	ugroup.UserGroupID.Set(idutil.ULIDNow())
	ugroup.UserGroupName.Set("name")
	ugroup.UpdatedAt.Set(time.Now())
	ugroup.CreatedAt.Set(time.Now())
	ugroup.ResourcePath.Set(fmt.Sprint(constants.ManabieSchool))

	ugMember := &entity.UserGroupMember{}
	database.AllNullEntity(ugMember)
	ugMember.UserID.Set(u.ID)
	ugMember.UserGroupID.Set(ugroup.UserGroupID.String)
	ugMember.CreatedAt.Set(time.Now())
	ugMember.UpdatedAt.Set(time.Now())
	ugMember.ResourcePath.Set(fmt.Sprint(constants.ManabieSchool))

	uG := entities.UserGroup{
		UserID:   u.ID,
		GroupID:  database.Text(oldGroup),
		IsOrigin: database.Bool(true),
	}
	uG.Status.Set("USER_GROUP_STATUS_ACTIVE")
	uG.CreatedAt = u.CreatedAt
	uG.UpdatedAt = u.UpdatedAt

	role := &entity.Role{}

	database.AllNullEntity(role)
	roleID := idutil.ULIDNow()
	stepState.RoleIDs = append(stepState.RoleIDs, roleID)
	role.RoleID.Set(roleID)
	role.RoleName.Set(newgroup)
	role.CreatedAt.Set(time.Now())
	role.UpdatedAt.Set(time.Now())
	role.ResourcePath.Set(fmt.Sprint(constants.ManabieSchool))

	grantedRole := &entity.GrantedRole{}
	database.AllNullEntity(grantedRole)
	grantedRole.RoleID.Set(role.RoleID.String)
	grantedRole.UserGroupID.Set(ugroup.UserGroupID.String)
	grantedRole.GrantedRoleID.Set(idutil.ULIDNow())
	grantedRole.CreatedAt.Set(time.Now())
	grantedRole.UpdatedAt.Set(time.Now())
	grantedRole.ResourcePath.Set(fmt.Sprint(constants.ManabieSchool))

	if _, err := database.Insert(ctx, &u, dbConn.Exec); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("insert user error: %v", err)
	}

	if _, err := database.Insert(ctx, &uG, dbConn.Exec); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("insert user group error: %v", err)
	}
	if _, err := database.Insert(ctx, ugroup, dbConn.Exec); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("insert user group member error: %v", err)
	}

	if _, err := database.Insert(ctx, ugMember, dbConn.Exec); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("insert user group member error: %v", err)
	}

	if _, err := database.Insert(ctx, role, dbConn.Exec); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("insert user group member error: %v", err)
	}
	if _, err := database.Insert(ctx, grantedRole, dbConn.Exec); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("insert user group member error: %v", err)
	}
	if u.Group.String == entities.UserGroupTeacher {
		teacher := &entities.Teacher{}
		database.AllNullEntity(teacher)

		err := multierr.Combine(
			teacher.ID.Set(u.ID.String),
			teacher.SchoolIDs.Set([]int64{constants.ManabieSchool}),
			teacher.UpdatedAt.Set(time.Now()),
			teacher.CreatedAt.Set(time.Now()),
		)
		if err != nil {
			return utils.StepStateToContext(ctx, stepState), err
		}

		_, err = database.Insert(ctx, teacher, dbConn.Exec)
		if err != nil {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("insert teacher error: %v", err)
		}
	}

	if u.Group.String == constant.UserGroupStudent {
		stepState.StudentID = u.ID.String
		stepState.CurrentStudentID = u.ID.String

		student := &entities.Student{}
		database.AllNullEntity(student)
		err := multierr.Combine(
			student.ID.Set(u.ID.String),
			student.SchoolID.Set(constants.ManabieSchool),
			student.EnrollmentStatus.Set("STUDENT_ENROLLMENT_STATUS_ENROLLED"),
			student.StudentNote.Set("example-student-note"),
			student.ResourcePath.Set(fmt.Sprint(constants.ManabieSchool)),
			student.CurrentGrade.Set(12),
			student.OnTrial.Set(true),
			student.TotalQuestionLimit.Set(10),
			student.BillingDate.Set(now),
			student.UpdatedAt.Set(time.Now()),
			student.CreatedAt.Set(time.Now()),
		)
		if err != nil {
			return utils.StepStateToContext(ctx, stepState), err
		}

		_, err = database.Insert(ctx, student, dbConn.Exec)
		if err != nil {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("insert student error: %v", err)
		}
	}

	if u.Group.String == constant.UserGroupSchoolAdmin {
		schoolAdminAccount := &entities.SchoolAdmin{}
		database.AllNullEntity(schoolAdminAccount)
		err := multierr.Combine(
			schoolAdminAccount.SchoolAdminID.Set(u.ID.String),
			schoolAdminAccount.SchoolID.Set(u.ResourcePath.String),
			schoolAdminAccount.UpdatedAt.Set(time.Now()),
			schoolAdminAccount.CreatedAt.Set(time.Now()),
			schoolAdminAccount.ResourcePath.Set(fmt.Sprint(constants.ManabieSchool)),
		)
		if err != nil {
			return utils.StepStateToContext(ctx, stepState), err
		}

		_, err = database.Insert(ctx, schoolAdminAccount, dbConn.Exec)
		if err != nil {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("aValidUser insert school error: %w", err)
		}
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) setFakeClaimToContext(ctx context.Context, resourcePath string, userGroup string) context.Context {
	claims := &interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{
			ResourcePath: resourcePath,
			UserGroup:    userGroup,
		},
	}
	return interceptors.ContextWithJWTClaims(ctx, claims)
}
