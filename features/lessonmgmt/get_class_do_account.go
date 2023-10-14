package lessonmgmt

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/lessonmgmt/modules/classdo/infrastructure/repo"
)

func (s *Suite) userGetsClassDoByUserID(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	userEmail := fmt.Sprintf("%s@email.com", stepState.ValidCsvRows[0])
	userID, err := s.getClassDoUserIDFromUserEmail(ctx, userEmail)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	classDoAccountRepo := repo.ClassDoAccountRepo{}
	classDoAccount, err := classDoAccountRepo.GetClassDoAccountByID(ctx, s.LessonmgmtDB, userID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.ClassDoAccount = classDoAccount.ToClassDoAccountDomain("")

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userGotExpectedClassDoAccount(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	userEmail := fmt.Sprintf("%s@email.com", stepState.ValidCsvRows[0])
	userID, err := s.getClassDoUserIDFromUserEmail(ctx, userEmail)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	classDoAccount := stepState.ClassDoAccount
	if userID != classDoAccount.ClassDoID {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expect user %s but got %s", userID, classDoAccount.ClassDoID)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) getClassDoUserIDFromUserEmail(ctx context.Context, userEmail string) (string, error) {
	var userID string
	query := `SELECT classdo_id 
				FROM classdo_account 
				WHERE classdo_email = $1
				AND deleted_at IS NULL`
	err := s.LessonmgmtDB.QueryRow(ctx, query, userEmail).Scan(&userID)

	if err != nil {
		return "", err
	}
	return userID, nil
}
