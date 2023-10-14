package entitiescreator

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/manabie-com/backend/features/common"
	bobEntities "github.com/manabie-com/backend/internal/bob/entities"
	bobRepo "github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	invoiceEntities "github.com/manabie-com/backend/internal/invoicemgmt/entities"
	userRepo "github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	userEntities "github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"

	"go.uber.org/multierr"
)

// CreateUser used userRepo.Create to insert user.
// stepState dependency:
//   - stepState.ResourcePath
//
// stepState assigned:
//   - stepState.CurrentUserID
func (c *EntitiesCreator) CreateUser(ctx context.Context, db database.QueryExecer, userID, userGroup string) InsertEntityFunction {
	return func(stepState *common.StepState) error {
		// Set the user entity according to its bound context.
		user := &bobEntities.User{}
		database.AllNullEntity(user)
		now := time.Now()

		err := multierr.Combine(
			user.ID.Set(userID),
			user.LastName.Set(fmt.Sprintf("invoice-user-name-%v", userID)),
			user.PhoneNumber.Set(fmt.Sprintf("%v-phone-number", userID)),
			user.Email.Set(fmt.Sprintf("%v@manabie.com", userID)),
			user.Country.Set("COUNTRY_VN"),
			user.Group.Set(userGroup),
			user.CreatedAt.Set(now),
			user.UpdatedAt.Set(now),
			user.ResourcePath.Set(stepState.ResourcePath),
		)
		if err != nil {
			return fmt.Errorf("user set: %w", err)
		}

		userRepo := bobRepo.UserRepo{}
		err = userRepo.Create(ctx, db, user)
		if err != nil {
			return fmt.Errorf("userRepo.Create: %w ", err)
		}

		stepState.CurrentUserID = userID

		return nil
	}
}

// CreateStudent used studentRepo.Create to insert student.
// stepState dependency:
//   - stepState.ResourcePath
//
// stepState assigned:
//   - stepState.StudentID
//   - stepState.CurrentUserID
func (c *EntitiesCreator) CreateStudent(ctx context.Context, db database.QueryExecer, studentID string) InsertEntityFunction {
	return func(stepState *common.StepState) error {
		firstName := fmt.Sprintf("valid-user-first-name-%s", studentID)
		lastName := fmt.Sprintf("valid-user-last-name-%s", studentID)

		// Set the student entity according to its bound context.
		student := &userEntities.LegacyStudent{}
		database.AllNullEntity(student)
		database.AllNullEntity(&student.LegacyUser)

		now := time.Now()
		err := multierr.Combine(
			student.LegacyUser.ID.Set(studentID),
			student.LegacyUser.GivenName.Set(firstName),
			student.LegacyUser.LastName.Set(lastName),
			student.LegacyUser.Country.Set(cpb.Country_COUNTRY_VN.String()),
			student.LegacyUser.Group.Set(userEntities.UserGroupStudent),
			student.LegacyUser.ExternalUserID.Set(fmt.Sprintf("external-%s", studentID)),
			student.LegacyUser.FullName.Set(firstName+" "+lastName),
			student.LegacyUser.FirstName.Set(firstName),

			student.ID.Set(studentID),
			student.CurrentGrade.Set(5),
			student.CreatedAt.Set(now),
			student.UpdatedAt.Set(now),
			student.ResourcePath.Set(stepState.ResourcePath),
		)
		if err != nil {
			return fmt.Errorf("student set: %w", err)
		}

		studentRepo := userRepo.StudentRepo{}
		err = studentRepo.Create(ctx, db, student)
		if err != nil {
			return fmt.Errorf("studentRepo.Create: %w", err)
		}

		stepState.StudentID = studentID
		stepState.CurrentUserID = studentID
		stepState.CurrentStudentFirstName = firstName
		stepState.CurrentStudentLastName = lastName

		return nil
	}
}

// CreateLocation used insert location statement to insert location.
// stepState assigned:
//   - stepState.LocationID
func (c *EntitiesCreator) CreateLocation(ctx context.Context, db database.QueryExecer) InsertEntityFunction {
	return func(stepState *common.StepState) error {
		ctx, span := interceptors.StartSpan(ctx, "EntitiesCreator.WithLocation")
		defer span.End()

		// Set the location entity according to its bound context.
		locationID := idutil.ULIDNow()
		location := &bobEntities.Location{}
		database.AllNullEntity(location)

		now := time.Now()
		err := multierr.Combine(
			location.LocationID.Set(locationID),
			location.AccessPath.Set(locationID),
			location.Name.Set(fmt.Sprintf("test-location-%v", locationID)),
			location.LocationType.Set("01FR4M51XJY9E77GSN4QZ1Q8M5"),
			location.PartnerInternalID.Set("1"),
			location.CreatedAt.Set(now),
			location.UpdatedAt.Set(now),
		)
		if err != nil {
			return fmt.Errorf("location set: %w", err)
		}

		stmt := InsertLocationStmt
		args := []interface{}{location.LocationID.String, location.LocationID.String, location.Name.String, location.LocationType.String, location.PartnerInternalID.String}

		if _, err := db.Exec(ctx, stmt, args...); err != nil {
			return fmt.Errorf("error insert new location record: %v", err)
		}

		stepState.LocationID = locationID

		return nil
	}
}

// CreateUserAccessPath used userAccessPathRepo.Upsert to insert user_access_path.
// stepState dependency:
//   - stepState.CurrentUserID
//   - stepState.LocationID
//   - stepState.ResourcePath
func (c *EntitiesCreator) CreateUserAccessPath(ctx context.Context, db database.QueryExecer) InsertEntityFunction {
	return func(stepState *common.StepState) error {
		userAccessPath := &userEntities.UserAccessPath{}
		database.AllNullEntity(userAccessPath)

		now := time.Now()
		err := multierr.Combine(
			userAccessPath.UserID.Set(stepState.CurrentUserID),
			userAccessPath.LocationID.Set(stepState.LocationID),
			userAccessPath.CreatedAt.Set(now),
			userAccessPath.UpdatedAt.Set(now),
			userAccessPath.ResourcePath.Set(stepState.ResourcePath),
		)

		if err != nil {
			return fmt.Errorf("userAccessPath set: %w", err)
		}

		userAccessPathRepo := userRepo.UserAccessPathRepo{}
		err = userAccessPathRepo.Upsert(ctx, db, []*userEntities.UserAccessPath{userAccessPath})
		if err != nil {
			log.Println(stepState.LocationID)
			return fmt.Errorf("userAccessPathRepo.Upsert: %w ", err)
		}

		return nil
	}
}

// CreateUserAccessPath used userAccessPathRepo.Upsert to insert user_access_path.
// stepState dependency:
//   - stepState.StudentID
//   - stepState.LocationID
//   - stepState.ResourcePath

func (c *EntitiesCreator) CreateUserAccessPathForStudent(ctx context.Context, db database.QueryExecer) InsertEntityFunction {
	return func(stepState *common.StepState) error {
		userAccessPath := &userEntities.UserAccessPath{}
		database.AllNullEntity(userAccessPath)

		now := time.Now()
		err := multierr.Combine(
			userAccessPath.UserID.Set(stepState.StudentID),
			userAccessPath.LocationID.Set(stepState.LocationID),
			userAccessPath.CreatedAt.Set(now),
			userAccessPath.UpdatedAt.Set(now),
			userAccessPath.ResourcePath.Set(stepState.ResourcePath),
		)

		if err != nil {
			return fmt.Errorf("userAccessPath set: %w", err)
		}

		userAccessPathRepo := userRepo.UserAccessPathRepo{}
		err = userAccessPathRepo.Upsert(ctx, db, []*userEntities.UserAccessPath{userAccessPath})
		if err != nil {
			return fmt.Errorf("userAccessPathRepo.Upsert: %w ", err)
		}

		return nil
	}
}

// CreateParent used parentRepo.CreateMultiple to insert parent.
// stepState dependency:
//   - stepState.CurrentUserID
//   - stepState.ResourcePath
//
// stepState assigned:
//   - stepState.CurrentParentID
//   - stepState.CurrentUserID
func (c *EntitiesCreator) CreateParent(ctx context.Context, db database.QueryExecer) InsertEntityFunction {
	return func(stepState *common.StepState) error {
		userID := idutil.ULIDNow()
		f := c.CreateUser(ctx, db, userID, bobEntities.UserGroupParent)
		err := f(stepState)
		if err != nil {
			return err
		}

		parentID := stepState.CurrentUserID

		parentRepo := userRepo.ParentRepo{}
		parent := &userEntities.Parent{}
		database.AllNullEntity(parent)

		now := time.Now()
		err = multierr.Combine(
			parent.ID.Set(stepState.CurrentUserID),
			parent.SchoolID.Set(stepState.CurrentSchoolID),
			parent.CreatedAt.Set(now),
			parent.UpdatedAt.Set(now),
			parent.ResourcePath.Set(stepState.ResourcePath),
		)
		if err != nil {
			return fmt.Errorf("parent set: %w", err)
		}

		err = parentRepo.CreateMultiple(ctx, db, []*userEntities.Parent{parent})
		if err != nil {
			return fmt.Errorf("parentRepo.CreateMultiple: %w", err)
		}

		stepState.CurrentParentID = parentID
		stepState.ParentIDs = append(stepState.ParentIDs, parentID)
		stepState.CurrentUserID = userID

		return nil
	}
}

// CreateStudentParent used studentParentRepo.Upsert to insert student_parent.
// stepState dependency:
//   - stepState.StudentID
//   - stepState.stepState.ParentIDs
func (c *EntitiesCreator) CreateStudentParent(ctx context.Context, db database.QueryExecer, relationship string) InsertEntityFunction {
	return func(stepState *common.StepState) error {
		entities := make([]*bobEntities.StudentParent, 0, len(stepState.ParentIDs))

		for _, parentID := range stepState.ParentIDs {
			studentParent := &bobEntities.StudentParent{}
			database.AllNullEntity(studentParent)
			now := time.Now()

			err := multierr.Combine(
				studentParent.StudentID.Set(stepState.StudentID),
				studentParent.ParentID.Set(parentID),
				studentParent.Relationship.Set(relationship),
				studentParent.CreatedAt.Set(now),
				studentParent.UpdatedAt.Set(now),
			)
			if err != nil {
				return fmt.Errorf("studentParent set: %w", err)
			}
			entities = append(entities, studentParent)
		}

		studentParentRepo := bobRepo.StudentParentRepo{}
		if err := studentParentRepo.Upsert(ctx, db, entities); err != nil {
			return fmt.Errorf("studentParentRepo.Upsert: %w", err)
		}

		return nil
	}
}

// CreateSchoolAdmin used schoolAdminRepo.CreateMultiple to insert school_admin.
// stepState dependency:
//   - stepState.ResourcePath
//
// stepState assigned:
//   - stepState.CurrentUserID
func (c *EntitiesCreator) CreateSchoolAdmin(ctx context.Context, db database.QueryExecer) InsertEntityFunction {
	return func(stepState *common.StepState) error {
		userID := idutil.ULIDNow()
		f := c.CreateUser(ctx, db, userID, bobEntities.UserGroupSchoolAdmin)
		err := f(stepState)
		if err != nil {
			return err
		}

		schoolAdminRepo := userRepo.SchoolAdminRepo{}
		schoolAdminAccount := &userEntities.SchoolAdmin{}
		database.AllNullEntity(schoolAdminAccount)
		now := time.Now()

		err = multierr.Combine(
			schoolAdminAccount.SchoolAdminID.Set(stepState.CurrentUserID),
			schoolAdminAccount.SchoolID.Set(stepState.CurrentSchoolID),
			schoolAdminAccount.CreatedAt.Set(now),
			schoolAdminAccount.UpdatedAt.Set(now),
			schoolAdminAccount.ResourcePath.Set(stepState.ResourcePath),
		)
		if err != nil {
			return fmt.Errorf("schoolAdmin set: %w", err)
		}

		err = schoolAdminRepo.CreateMultiple(ctx, db, []*userEntities.SchoolAdmin{schoolAdminAccount})
		if err != nil {
			return fmt.Errorf("schoolAdminRepo.CreateMultiple: %w", err)
		}

		stepState.CurrentUserID = userID

		return nil
	}
}

// CreateTeacher used teacherRepo.CreateMultiple to insert teacher.
// stepState dependency:
//   - stepState.ResourcePath
//
// stepState assigned:
//   - stepState.CurrentUserID
func (c *EntitiesCreator) CreateTeacher(ctx context.Context, db database.QueryExecer) InsertEntityFunction {
	return func(stepState *common.StepState) error {
		userID := idutil.ULIDNow()
		f := c.CreateUser(ctx, db, userID, bobEntities.UserGroupTeacher)
		err := f(stepState)
		if err != nil {
			return err
		}

		teacherRepo := userRepo.TeacherRepo{}
		t := &userEntities.Teacher{}
		database.AllNullEntity(t)

		_ = t.ResourcePath.Set(stepState.ResourcePath)

		t.ID = database.Text(stepState.CurrentUserID)
		err = t.SchoolIDs.Set([]int64{int64(stepState.CurrentSchoolID)})
		if err != nil {
			return fmt.Errorf("teacher set: %w", err)
		}

		err = teacherRepo.CreateMultiple(ctx, db, []*userEntities.Teacher{t})
		if err != nil {
			return fmt.Errorf("teacherRepo.CreateMultiple: %w", err)
		}

		stepState.CurrentUserID = userID

		return nil
	}
}

// CreateUserGroup used userGroupRepo.Upsert to insert user_group.
// stepState dependency:
//   - stepState.ResourcePath
func (c *EntitiesCreator) CreateUserGroup(ctx context.Context, db database.QueryExecer, userID, groupID string) InsertEntityFunction {
	return func(stepState *common.StepState) error {
		uGroup := &userEntities.UserGroup{}
		database.AllNullEntity(uGroup)

		err := multierr.Combine(
			uGroup.GroupID.Set(groupID),
			uGroup.UserID.Set(userID),
			uGroup.IsOrigin.Set(true),
			uGroup.Status.Set("USER_GROUP_STATUS_ACTIVE"),
			uGroup.ResourcePath.Set(stepState.ResourcePath),
		)
		if err != nil {
			return err
		}

		userGroupRepo := &userRepo.UserGroupRepo{}
		err = userGroupRepo.Upsert(ctx, db, uGroup)
		if err != nil {
			return err
		}

		return nil
	}
}

// CreateGrantedRoleAccessPath creates granted role access path
// stepState dependency:
//   - stepState.LocationID
func (c *EntitiesCreator) CreateGrantedRoleAccessPath(ctx context.Context, db database.QueryExecer, role string) InsertEntityFunction {
	return func(stepState *common.StepState) error {
		ctx, span := interceptors.StartSpan(ctx, "EntitiesCreator.CreateGrantedRoleAccessPathOrgLevel")
		defer span.End()

		insertStmt := InsertGrantedRoleAccessPathStmt
		args := []interface{}{stepState.GrantedRoleId, stepState.LocationID}

		if _, err := db.Exec(ctx, insertStmt, args...); err != nil {
			// ignore duplicate record error
			if !strings.Contains(err.Error(), "SQLSTATE 23505") {
				return fmt.Errorf("error insert new granted role access path record: %v", err)
			}
		}

		return nil
	}
}

// CreateUserGroupV2 used userGroupRepo.Create to insert user_group.
// stepState dependency:
//   - stepState.ResourcePath
func (c *EntitiesCreator) CreateUserGroupV2(ctx context.Context, db database.QueryExecer, groupName string) InsertEntityFunction {
	return func(stepState *common.StepState) error {
		userGroupRepo := &userRepo.UserGroupV2Repo{}
		uGroup := &userEntities.UserGroupV2{}
		database.AllNullEntity(uGroup)

		newGroupID := idutil.ULIDNow()
		err := multierr.Combine(
			uGroup.UserGroupID.Set(newGroupID),
			uGroup.UserGroupName.Set(groupName),
			uGroup.IsSystem.Set(false),
			uGroup.ResourcePath.Set(stepState.ResourcePath),
		)
		if err != nil {
			return err
		}

		err = userGroupRepo.Create(ctx, db, uGroup)
		if err != nil {
			return err
		}

		stepState.ExistedUserGroupID = newGroupID

		return nil
	}
}

// CreateUserGroupV2 used userGroupRepo.Create to insert user_group.
// stepState dependency:
//   - stepState.ResourcePath
//   - stepState.ExistedUserGroupID
func (c *EntitiesCreator) CreateGrantedRole(ctx context.Context, db database.QueryExecer, role string) InsertEntityFunction {
	return func(stepState *common.StepState) error {
		grantedRoleRepo := &userRepo.GrantedRoleRepo{}

		roleRepo := &userRepo.RoleRepo{}
		role, err := roleRepo.GetByName(ctx, db, database.Text(role))
		if err != nil {
			return err
		}

		grantedRole := &userEntities.GrantedRole{}
		database.AllNullEntity(grantedRole)

		newGrantedRoleID := idutil.ULIDNow()
		err = multierr.Combine(
			grantedRole.UserGroupID.Set(stepState.ExistedUserGroupID),
			grantedRole.RoleID.Set(role.RoleID),
			grantedRole.GrantedRoleID.Set(newGrantedRoleID),
			grantedRole.ResourcePath.Set(stepState.ResourcePath),
		)
		if err != nil {
			return err
		}

		err = grantedRoleRepo.Create(ctx, db, grantedRole)
		if err != nil {
			return err
		}

		stepState.GrantedRoleId = newGrantedRoleID

		return nil
	}
}

// CreateUserGroupV2 used userGroupRepo.Create to insert user_group.
// stepState dependency:
//   - stepState.ResourcePath
//   - stepState.ExistedUserGroupID
//   - stepState.CurrentUserID
func (c *EntitiesCreator) CreateUserGroupMember(ctx context.Context, db database.QueryExecer) InsertEntityFunction {
	return func(stepState *common.StepState) error {
		userGroupsMemberRepo := &userRepo.UserGroupsMemberRepo{}

		user := &userEntities.LegacyUser{}
		database.AllNullEntity(user)

		err := multierr.Combine(
			user.ID.Set(stepState.CurrentUserID),
			user.ResourcePath.Set(stepState.ResourcePath),
		)
		if err != nil {
			return err
		}

		return userGroupsMemberRepo.AssignWithUserGroup(ctx, db, []*userEntities.LegacyUser{user}, database.Text(stepState.ExistedUserGroupID))
	}
}

// CreatePrefecture used insert prefecture statement to insert prefecture.
// stepState assigned:
//   - stepState.PrefectureID
func (c *EntitiesCreator) CreatePrefecture(ctx context.Context, db database.QueryExecer) InsertEntityFunction {
	return func(stepState *common.StepState) error {
		ctx, span := interceptors.StartSpan(ctx, "EntitiesCreator.WithLocation")
		defer span.End()

		// Set the prefecture entity according to its bound context.
		prefectureID := idutil.ULIDNow()
		prefecture := &userEntities.Prefecture{}
		database.AllNullEntity(prefecture)

		now := time.Now()
		err := multierr.Combine(
			prefecture.ID.Set(prefectureID),
			prefecture.PrefectureCode.Set(fmt.Sprintf("prefecture-code-%s", prefectureID)),
			prefecture.Country.Set(fmt.Sprintf("prefecture-country-%s", prefectureID)),
			prefecture.Name.Set(fmt.Sprintf("prefecture-name-%s", prefectureID)),
			prefecture.CreatedAt.Set(now),
			prefecture.UpdatedAt.Set(now),
		)
		if err != nil {
			return fmt.Errorf("location set: %w", err)
		}

		stmt := InsertPrefectureStmt
		args := []interface{}{prefecture.ID.String, prefecture.PrefectureCode.String, prefecture.Country.String, prefecture.Name.String}

		if _, err := db.Exec(ctx, stmt, args...); err != nil {
			return fmt.Errorf("error insert new location record: %v", err)
		}

		stepState.PrefectureID = prefectureID
		stepState.PrefectureCode = prefecture.PrefectureCode.String

		return nil
	}
}

// CreateUserBasicInfo used insert user basic info statement to insert user basic info.
// stepState dependency:
//   - stepState.UserID
func (c *EntitiesCreator) CreateUserBasicInfo(ctx context.Context, db database.Ext) InsertEntityFunction {
	return func(stepState *common.StepState) error {
		ctx, span := interceptors.StartSpan(ctx, "EntitiesCreator.CreateUserBasicInfo")
		defer span.End()

		gradeID := fmt.Sprintf("grade-%v", idutil.ULIDNow())
		userBasicInfo := &invoiceEntities.UserBasicInfo{}
		database.AllNullEntity(userBasicInfo)

		id := stepState.CurrentUserID
		now := time.Now()
		err := multierr.Combine(
			userBasicInfo.UserID.Set(id),
			userBasicInfo.Name.Set(fmt.Sprintf("name-%s", id)),
			userBasicInfo.FirstName.Set(fmt.Sprintf("name-%s", id)),
			userBasicInfo.LastName.Set(fmt.Sprintf("last-name-%s", id)),
			userBasicInfo.FullNamePhonetic.Set(fmt.Sprintf("full-name-phonetic-%s", id)),
			userBasicInfo.FirstNamePhonetic.Set(fmt.Sprintf("first-name-phonetic-%s", id)),
			userBasicInfo.LastNamePhonetic.Set(fmt.Sprintf("last-name-phonetic-%s", id)),
			userBasicInfo.CurrentGrade.Set(5),
			userBasicInfo.GradeID.Set(gradeID),
			userBasicInfo.CreatedAt.Set(now),
			userBasicInfo.UpdatedAt.Set(now),
			userBasicInfo.ResourcePath.Set(stepState.ResourcePath),
		)
		if err != nil {
			return fmt.Errorf("location set: %w", err)
		}

		stmt := InsertUserBasicInfoStmt
		args := []interface{}{
			userBasicInfo.UserID.String,
			userBasicInfo.Name.String,
			userBasicInfo.FirstName.String,
			userBasicInfo.LastName.String,
			userBasicInfo.FullNamePhonetic.String,
			userBasicInfo.FirstNamePhonetic.String,
			userBasicInfo.LastNamePhonetic.String,
			userBasicInfo.CurrentGrade.Int,
			userBasicInfo.GradeID.String,
		}

		if _, err := db.Exec(ctx, stmt, args...); err != nil {
			return fmt.Errorf("error insert new user basic info record: %v", err)
		}

		return nil
	}
}

func (c *EntitiesCreator) CreateMigrationStudent(ctx context.Context, db database.QueryExecer, studentID string) InsertEntityFunction {
	return func(stepState *common.StepState) error {
		firstName := fmt.Sprintf("valid-user-first-name-%s", studentID)
		lastName := fmt.Sprintf("valid-user-last-name-%s", studentID)

		// Set the student entity according to its bound context.
		student := &bobEntities.Student{}
		database.AllNullEntity(student)
		database.AllNullEntity(&student.User)

		now := time.Now()
		migrationEmail := "@student.kec.gr.jp"
		err := multierr.Combine(
			student.User.ID.Set(studentID),
			student.User.GivenName.Set(firstName),
			student.User.LastName.Set(lastName),
			student.User.Country.Set(cpb.Country_COUNTRY_VN.String()),
			student.User.Group.Set(userEntities.UserGroupStudent),
			student.User.Email.Set(fmt.Sprintf("%v%v", studentID, migrationEmail)),
			student.ID.Set(studentID),
			student.CurrentGrade.Set(5),
			student.CreatedAt.Set(now),
			student.UpdatedAt.Set(now),
			student.ResourcePath.Set(stepState.ResourcePath),
		)
		if err != nil {
			return fmt.Errorf("student set: %w", err)
		}

		studentRepo := bobRepo.StudentRepo{}
		err = studentRepo.Create(ctx, db, student)
		if err != nil {
			return fmt.Errorf("studentRepo.Create: %w", err)
		}

		stepState.StudentID = studentID
		stepState.CurrentUserID = studentID

		return nil
	}
}
