// Package features
package yasuo

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/manabie-com/backend/features/helper"
	golibs_auth "github.com/manabie-com/backend/internal/golibs/auth"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/try"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
)

type userOption func(u *entity.LegacyUser)

func withID(id string) userOption {
	return func(u *entity.LegacyUser) {
		_ = u.ID.Set(id)
	}
}

func withRole(group string) userOption {
	return func(u *entity.LegacyUser) {
		_ = u.Group.Set(group)
	}
}

func (s *suite) ASignedInAdmin(ctx context.Context) (context.Context, error) {
	return s.aSignedIn(ctx, "school admin")
}

func (s *suite) aSignedInCurrentUser(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var (
		token string
		err   error
	)
	retryTime := 3

	// need to retry because of kafka replicate need sometime to sync data
	err = try.Do(func(attempt int) (bool, error) {
		token, err = s.generateExchangeToken(stepState.CurrentUserID, stepState.CurrentUserGroup)
		if err == nil {
			return false, nil
		}

		if attempt < retryTime {
			time.Sleep(1 * time.Second)
			return true, err
		}

		return false, err
	})
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("failed to generate exchange token: %v", err)
	}

	stepState.AuthToken = token
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aSignedIn(ctx context.Context, user string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var userGroup string
	switch user {
	case "unauthenticated":
		stepState.AuthToken = "random-token"
		return StepStateToContext(ctx, stepState), nil
	case "teacher":
		userGroup = entity.UserGroupTeacher
	case "school admin":
		userGroup = entity.UserGroupSchoolAdmin
	case "student":
		userGroup = constant.UserGroupStudent
	case "parent":
		userGroup = constant.UserGroupParent
	}

	stepState.CurrentUserID = newID()
	ctx, err := s.aValidToken(ctx, userGroup)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("aValidToken: %w", err)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aValidToken(ctx context.Context, userGroup string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.aValidUserInDB(ctx, withID(stepState.CurrentUserID), withRole(userGroup))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("aValidUserInDB: %w", err)
	}

	if err := checkUserExisted(ctx, s.DB, stepState.CurrentUserID); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("user isn't existed: %w", err)
	}
	token, err := s.generateExchangeToken(stepState.CurrentUserID, userGroup)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("generateExchangeToken: %w", err)
	}

	stepState.AuthToken = token
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aValidUserInDB(ctx context.Context, opts ...userOption) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.CurrentSchoolID == 0 {
		stepState.CurrentSchoolID = constants.ManabieSchool
	}
	schoolID := int64(stepState.CurrentSchoolID)
	ctx = golibs_auth.InjectFakeJwtToken(ctx, fmt.Sprint(schoolID))

	user, err := newUserEntity(fmt.Sprint(schoolID))
	if err != nil {
		return StepStateToContext(ctx, stepState), errors.Wrap(err, "newUserEntity")
	}

	for _, opt := range opts {
		opt(user)
	}

	err = database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		userRepo := repository.UserRepo{}
		if err := userRepo.Create(ctx, tx, user); err != nil {
			return fmt.Errorf("cannot create user: %w", err)
		}

		switch user.Group.String {
		case constant.UserGroupTeacher:
			teacherRepo := repository.TeacherRepo{}
			teacher := &entity.Teacher{}
			database.AllNullEntity(teacher)
			teacher.ID = user.ID
			err := multierr.Combine(
				teacher.SchoolIDs.Set([]int64{schoolID}),
				teacher.ResourcePath.Set(fmt.Sprint(schoolID)),
			)
			if err != nil {
				return err
			}

			err = teacherRepo.CreateMultiple(ctx, tx, []*entity.Teacher{teacher})
			if err != nil {
				return fmt.Errorf("cannot create teacher: %w", err)
			}

		case constant.UserGroupSchoolAdmin:
			schoolAdminRepo := repository.SchoolAdminRepo{}
			schoolAdmin := &entity.SchoolAdmin{}
			database.AllNullEntity(schoolAdmin)

			err := multierr.Combine(
				schoolAdmin.SchoolAdminID.Set(user.ID.String),
				schoolAdmin.SchoolID.Set(schoolID),
				schoolAdmin.ResourcePath.Set(database.Text(fmt.Sprint(schoolID))),
			)
			if err != nil {
				return fmt.Errorf("cannot create school admin: %w", err)
			}
			if err = schoolAdminRepo.CreateMultiple(ctx, tx, []*entity.SchoolAdmin{schoolAdmin}); err != nil {
				return err
			}
		case constant.UserGroupParent:
			parentRepo := repository.ParentRepo{}
			parentEnt := &entity.Parent{}
			database.AllNullEntity(parentEnt)
			err := multierr.Combine(
				parentEnt.ID.Set(user.ID.String),
				parentEnt.SchoolID.Set(schoolID),
				parentEnt.ResourcePath.Set(fmt.Sprint(schoolID)),
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
			student, err := newStudentEntity(fmt.Sprint(schoolID))
			if err != nil {
				return err
			}
			err = multierr.Combine(
				student.ID.Set(user.ID.String),
				student.SchoolID.Set(schoolID),
				student.ResourcePath.Set(fmt.Sprint(schoolID)),
			)
			if err != nil {
				return err
			}
			err = studentRepo.CreateMultiple(ctx, tx, []*entity.LegacyStudent{student})
			if err != nil {
				return fmt.Errorf("cannot create student: %w", err)
			}
		}

		userGroup := &entity.UserGroup{}
		database.AllNullEntity(userGroup)

		err = multierr.Combine(
			userGroup.GroupID.Set(user.Group.String),
			userGroup.UserID.Set(user.ID.String),
			userGroup.IsOrigin.Set(true),
			userGroup.Status.Set(entity.UserGroupStatusActive),
			userGroup.ResourcePath.Set(database.Text(fmt.Sprint(schoolID))),
		)
		if err != nil {
			return err
		}

		userGroupRepo := &repository.UserGroupRepo{}
		if err = userGroupRepo.Upsert(ctx, tx, userGroup); err != nil {
			return fmt.Errorf("userGroupRepo.Upsert: %w %s", err, user.Group.String)
		}

		return nil
	})

	return StepStateToContext(ctx, stepState), err
}

func newID() string {
	return idutil.ULIDNow()
}

func newUserEntity(resourcePath string) (*entity.LegacyUser, error) {
	userID := newID()
	now := time.Now()
	user := new(entity.LegacyUser)
	firstName := fmt.Sprintf("user-first-name-%s", userID)
	lastName := fmt.Sprintf("user-last-name-%s", userID)
	fullName := helper.CombineFirstNameAndLastNameToFullName(firstName, lastName)
	database.AllNullEntity(user)
	database.AllNullEntity(&user.AppleUser)
	if err := multierr.Combine(
		user.ID.Set(userID),
		user.Email.Set(fmt.Sprintf("valid-user-%s@email.com", userID)),
		user.Avatar.Set(fmt.Sprintf("http://valid-user-%s", userID)),
		user.IsTester.Set(false),
		user.FacebookID.Set(userID),
		user.PhoneVerified.Set(false),
		user.AllowNotification.Set(true),
		user.EmailVerified.Set(false),
		user.FullName.Set(fullName),
		user.FirstName.Set(firstName),
		user.LastName.Set(lastName),
		user.Country.Set(cpb.Country_COUNTRY_JP.String()),
		user.Group.Set(entity.UserGroupStudent),
		user.Birthday.Set(now),
		user.Gender.Set(pb.Gender_FEMALE.String()),
		user.ResourcePath.Set(resourcePath),
		user.CreatedAt.Set(now),
		user.UpdatedAt.Set(now),
		user.DeletedAt.Set(nil),
	); err != nil {
		return nil, errors.Wrap(err, "set value user")
	}

	user.UserAdditionalInfo = entity.UserAdditionalInfo{
		CustomClaims: map[string]interface{}{
			"external-info": "example-info",
		},
	}
	return user, nil
}

func newStudentEntity(resourcePath string) (*entity.LegacyStudent, error) {
	now := time.Now()
	student := new(entity.LegacyStudent)
	database.AllNullEntity(student)
	database.AllNullEntity(&student.LegacyUser)
	user, err := newUserEntity(resourcePath)
	if err != nil {
		return nil, errors.Wrap(err, "newUserEntity")
	}
	student.LegacyUser = *user
	schoolID, err := strconv.ParseInt(student.LegacyUser.ResourcePath.String, 10, 32)
	if err != nil {
		return nil, errors.Wrap(err, "strconv.ParseInt")
	}

	if err := multierr.Combine(
		student.ID.Set(student.LegacyUser.ID),
		student.SchoolID.Set(schoolID),
		student.EnrollmentStatus.Set(pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED),
		student.StudentExternalID.Set(student.LegacyUser.ID),
		student.StudentNote.Set("this is the note"),
		student.CurrentGrade.Set(1),
		student.TargetUniversity.Set("HUST"),
		student.TotalQuestionLimit.Set(32),
		student.OnTrial.Set(false),
		student.BillingDate.Set(now),
		student.CreatedAt.Set(student.LegacyUser.CreatedAt),
		student.UpdatedAt.Set(student.LegacyUser.UpdatedAt),
		student.DeletedAt.Set(student.LegacyUser.DeletedAt),
		student.PreviousGrade.Set(12),
		student.ResourcePath.Set(resourcePath),
	); err != nil {
		return nil, errors.Wrap(err, "set value student")
	}

	return student, nil
}

func checkUserExisted(ctx context.Context, db database.QueryExecer, userID string) error {
	retryTimes := 3
	err := try.Do(func(attempt int) (retry bool, err error) {
		if _, err = (&repository.UserRepo{}).FindByIDUnscope(ctx, db, database.Text(userID)); err == nil {
			return false, nil
		}
		if attempt < retryTimes {
			time.Sleep(time.Second)
			return false, nil
		}
		return true, err
	})

	return err
}
