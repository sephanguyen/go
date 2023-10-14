package usermgmt

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/cmd/server/usermgmt"
	"github.com/manabie-com/backend/internal/bob/configurations"
	"github.com/manabie-com/backend/internal/golibs/auth"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"

	"go.uber.org/multierr"
)

const (
	randomNumberStaff int = 20
)

func (s *suite) someRandomUserWithPhoneNumberInDB(ctx context.Context, userType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.CurrentSchoolID == 0 {
		stepState.CurrentSchoolID = constants.ManabieSchool
	}

	ctx = auth.InjectFakeJwtToken(ctx, fmt.Sprint(stepState.CurrentSchoolID))
	staffRepo := &repository.StaffRepo{}
	studentRepo := &repository.StudentRepo{}

	for i := 0; i < randomNumberStaff; i++ {
		var err error
		newUserID := idutil.ULIDNow()
		stepState.UserIDs = append(stepState.UserIDs, newUserID)

		switch userType {
		case staff:
			var newStaff *entity.Staff

			newStaff, err = staffEntity(newUserID)
			if err != nil {
				return StepStateToContext(ctx, stepState), err
			}

			err = staffRepo.Create(ctx, s.BobDBTrace, newStaff)

		case student:
			var newStudent *entity.LegacyStudent

			newStudent, err = studentEntityWithFullNameOnly(newUserID)
			if err != nil {
				return StepStateToContext(ctx, stepState), err
			}

			err = studentRepo.Create(ctx, s.BobDBTrace, newStudent)

		case parent:
			_, err = aValidParentInDB(ctx, s.BobDBTrace, newUserID)
		}

		// Check err of 3 case student, staff, parent
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) systemRunJobToMigrateUserPhoneNumber(ctx context.Context, userType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	usermgmt.RunMigrateUserPhoneNumber(ctx, &configurations.Config{
		Common:     s.Cfg.Common,
		PostgresV2: s.Cfg.PostgresV2,
	}, fmt.Sprint(stepState.CurrentSchoolID), userType)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) phoneNumberOfUserMigratedToUserPhoneNumberSuccessfully(ctx context.Context, userType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx = auth.InjectFakeJwtToken(ctx, fmt.Sprint(stepState.CurrentSchoolID))

	var phoneNumberType, userGroup string
	users := entity.LegacyUsers{}
	userPhoneNumbers := entity.UserPhoneNumbers{}

	switch userType {
	case student:
		phoneNumberType = entity.StudentPhoneNumber
		userGroup = `
			JOIN students s ON s.student_id  = u.user_id 
			JOIN user_phone_number upn ON upn.user_id  = u.user_id and upn."type" = 'STUDENT_PHONE_NUMBER'
		`

	case staff:
		phoneNumberType = entity.StaffPrimaryPhoneNumber
		userGroup = `JOIN staff s ON s.staff_id  = u.user_id `

	case parent:
		phoneNumberType = entity.ParentPrimaryPhoneNumber
		userGroup = `JOIN parents s ON s.parent_id  = u.user_id `

	default:
		return StepStateToContext(ctx, stepState), fmt.Errorf("must have userType before doing anything")
	}

	stmt := fmt.Sprintf(`
		SELECT u.user_id, u.phone_number, u.user_group
		FROM users u
		%v
		WHERE u.phone_number is not null and u.phone_number ~ '^[0-9+]+$' and u.user_id = ANY($1)
		ORDER BY u.user_id
	`, userGroup)

	stmtUserPhoneNumber := `
		SELECT user_id, phone_number, "type"
		FROM user_phone_number
		WHERE "type" = $1 and user_id = ANY($2)
		ORDER BY user_id
	`

	err := multierr.Combine(
		database.Select(ctx, s.BobDB, stmt, &stepState.UserIDs).ScanAll(&users),
		database.Select(ctx, s.BobDB, stmtUserPhoneNumber, phoneNumberType, &stepState.UserIDs).ScanAll(&userPhoneNumbers),
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if len(users) != len(userPhoneNumbers) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected number of %v's phoneNumber is %v but actually is %v", userType, len(users), len(userPhoneNumbers))
	}

	for index, user := range users {
		if user.ID != userPhoneNumbers[index].UserID {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expected UserID is %v but actually is %v", user.ID.String, userPhoneNumbers[index].UserID.String)
		}
		if user.PhoneNumber.String != userPhoneNumbers[index].PhoneNumber.String {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expected phone_number is %v but actually is %v", user.PhoneNumber, userPhoneNumbers[index].PhoneNumber)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}
