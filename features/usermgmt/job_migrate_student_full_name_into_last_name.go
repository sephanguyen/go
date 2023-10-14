package usermgmt

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/cmd/server/usermgmt"
	"github.com/manabie-com/backend/internal/bob/configurations"
	"github.com/manabie-com/backend/internal/golibs/auth"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
)

const (
	randomNumberStudent int = 20
)

func (s *suite) someRandomStudentWithFullNameOnlyInDB(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	schoolID := int64(stepState.CurrentSchoolID)
	if schoolID == 0 {
		schoolID = constants.ManabieSchool
	}

	studentRepo := &repository.StudentRepo{}

	for i := 0; i < randomNumberStudent; i++ {
		newStudentId := idutil.ULIDNow()
		stepState.StudentIds = append(stepState.StudentIds, newStudentId)
		studentEntityWithFullNameOnly, err := studentEntityWithFullNameOnly(newStudentId)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		err = studentRepo.Create(auth.InjectFakeJwtToken(ctx, fmt.Sprint(schoolID)), s.BobDBTrace, studentEntityWithFullNameOnly)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) systemRunJobToMigrateStudentFullNameIntoLastName(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	usermgmt.RunMigrateStudentFullNameToLastNameAndFirstName(ctx, &configurations.Config{
		Common:     s.Cfg.Common,
		PostgresV2: s.Cfg.PostgresV2,
	})

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentFullNameMigratedToLastNameSuccessfully(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	schoolID := int64(stepState.CurrentSchoolID)
	if schoolID == 0 {
		schoolID = constants.ManabieSchool
	}

	ctx = contextWithToken(ctx)
	stmt := `
		SELECT 
			users.name,
			users.last_name,
			users.first_name
		FROM
			users
		WHERE
			users.user_id = ANY($1)
	`
	users := entity.LegacyUsers{}
	err := database.Select(auth.InjectFakeJwtToken(ctx, fmt.Sprint(schoolID)), s.BobDB, stmt, &stepState.StudentIds).ScanAll(&users)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	for _, user := range users {
		fullNameArray := strings.Split(user.FullName.String, " ")
		if len(fullNameArray) > 1 {
			lastName := fullNameArray[0]
			firstName := strings.Join(fullNameArray[1:], " ")
			switch {
			case user.FirstName.String != firstName:
				return StepStateToContext(ctx, stepState), fmt.Errorf("expected first name is %v but actually is %v", firstName, user.FirstName.String)
			case user.LastName.String != lastName:
				return StepStateToContext(ctx, stepState), fmt.Errorf("expected last name is %v but actually is %v", lastName, user.LastName.String)
			}
		} else if user.FullName.String != user.LastName.String {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expected last name is %v but actually is %v", user.FullName.String, user.LastName.String)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}
