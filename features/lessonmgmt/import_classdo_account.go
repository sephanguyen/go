package lessonmgmt

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/classdo/domain"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"

	"github.com/jackc/pgtype"
)

func (s *Suite) userImportsClassDoAccountsWithData(ctx context.Context, condition string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	header := fmt.Sprintf("%s,%s,%s,%s",
		domain.ClassDoIDLabel,
		domain.ClassDoEmailLabel,
		domain.ClassDoAPIKeyLabel,
		domain.ClassDoActionLabel,
	)
	mockIDs := []string{
		idutil.ULIDNow(),
		idutil.ULIDNow(),
		idutil.ULIDNow(),
		idutil.ULIDNow(),
	}
	req := &lpb.ImportClassDoAccountRequest{}

	switch condition {
	case "valid":
		stepState.ValidCsvRows = mockIDs

		req = &lpb.ImportClassDoAccountRequest{
			Payload: []byte(fmt.Sprintf(`%s
					%s
					%s
					%s
					%s`,
				header,
				","+mockIDs[0]+"@email.com,APIKEY"+mockIDs[0]+",Upsert",
				","+mockIDs[1]+"@email.com,APIKEY"+mockIDs[1]+",upsert",
				","+mockIDs[2]+"@email.com,APIKEY"+mockIDs[2]+",uPsert ",
				","+mockIDs[3]+"@email.com,APIKEY"+mockIDs[3]+",Upsert",
			)),
		}
	case "invalid":
		stepState.InvalidCsvRows = mockIDs

		req = &lpb.ImportClassDoAccountRequest{
			Payload: []byte(fmt.Sprintf(`%s
					%s
					%s
					%s
					%s`,
				header,
				","+mockIDs[0]+"@email.com,APIKEY"+mockIDs[0]+",Upsert",
				",,APIKEY"+mockIDs[1]+",Upsert",
				","+mockIDs[2]+"@email.com,,Upsert",
				","+mockIDs[3]+"@email.com,APIKEY"+mockIDs[3]+",Upsert",
			)),
		}
	}

	return s.importClassDoAccount(StepStateToContext(ctx, stepState), req)
}

func (s *Suite) userImportsClassDoAccountsWithDeleteAction(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if len(stepState.ValidCsvRows) == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("no valid rows in the stepState are found")
	}

	emails := make([]string, 0, len(stepState.ValidCsvRows))
	for _, accountID := range stepState.ValidCsvRows {
		emails = append(emails, accountID+"@email.com")
	}
	classDoAccMap, err := s.getClassDoAccountsIDsByEmail(ctx, emails)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("error when retrieving ClassDo account IDs by email: %w", err)
	}

	payloadString := fmt.Sprintf("%s,%s,%s,%s",
		domain.ClassDoIDLabel,
		domain.ClassDoEmailLabel,
		domain.ClassDoAPIKeyLabel,
		domain.ClassDoActionLabel,
	)
	for classDoEmail, classDoID := range classDoAccMap {
		payloadString += fmt.Sprintf(`
			%s`, classDoID+","+classDoEmail+",,delete")
	}

	req := &lpb.ImportClassDoAccountRequest{
		Payload: []byte(payloadString),
	}

	return s.importClassDoAccount(StepStateToContext(ctx, stepState), req)
}

func (s *Suite) importClassDoAccount(ctx context.Context, req *lpb.ImportClassDoAccountRequest) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.Response, stepState.ResponseErr = lpb.NewClassDoAccountServiceClient(s.LessonMgmtConn).
		ImportClassDoAccount(contextWithToken(s, ctx), req)
	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) classDoAccountsAreInTheDatabase(ctx context.Context, dataCondition string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	switch dataCondition {
	case "existing":
		if len(stepState.ValidCsvRows) == 0 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("no valid rows in the stepState are found")
		}
		emails := make([]string, 0, len(stepState.ValidCsvRows))
		for _, id := range stepState.ValidCsvRows {
			emails = append(emails, id+"@email.com")
		}

		foundEmails, err := s.getClassDoAccountsEmailsByEmail(ctx, emails)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("error when retrieving ClassDo accounts by email: %w", err)
		}

		isMatching := sliceutils.UnorderedEqual(emails, foundEmails)
		if !isMatching {
			return StepStateToContext(ctx, stepState), fmt.Errorf("emails found (%v) does not match with the expected emails (%v) (@email.com is just added, check the IDs)", foundEmails, emails)
		}

	case "not existing":
		if len(stepState.InvalidCsvRows) == 0 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("no invalid rows in the stepState are found")
		}
		emails := make([]string, 0, len(stepState.ValidCsvRows))

		foundEmails, err := s.getClassDoAccountsEmailsByEmail(ctx, emails)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("error when retrieving ClassDo accounts by email: %w", err)
		}

		if len(foundEmails) > 0 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("there are emails found (%v) that should not be present", foundEmails)
		}

	case "removed":
		if len(stepState.ValidCsvRows) == 0 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("no valid rows in the stepState are found")
		}
		emails := make([]string, 0, len(stepState.ValidCsvRows))
		for _, id := range stepState.ValidCsvRows {
			emails = append(emails, id+"@email.com")
		}

		foundEmails, err := s.getDeletedClassDoAccountsEmailsByEmail(ctx, emails)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("error when retrieving ClassDo accounts by email: %w", err)
		}

		isMatching := sliceutils.UnorderedEqual(emails, foundEmails)
		if !isMatching {
			return StepStateToContext(ctx, stepState), fmt.Errorf("emails found (%v) does not match with the expected emails (%v) (@email.com is just added, check the IDs)", foundEmails, emails)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) getClassDoAccountsEmailsByEmail(ctx context.Context, emails []string) ([]string, error) {
	query := `SELECT classdo_email 
				FROM classdo_account 
				WHERE classdo_email = ANY($1) 
				AND deleted_at IS NULL`
	rows, err := s.LessonmgmtDB.Query(ctx, query, emails)
	if err != nil {
		return nil, fmt.Errorf("db.Query: %w", err)
	}
	defer rows.Close()

	var email pgtype.Text
	classDoEmails := make([]string, 0, len(emails))
	for rows.Next() {
		if err := rows.Scan(&email); err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}

		classDoEmails = append(classDoEmails, email.String)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}

	return classDoEmails, nil
}

func (s *Suite) getClassDoAccountsIDsByEmail(ctx context.Context, emails []string) (map[string]string, error) {
	query := `SELECT classdo_id,  classdo_email
				FROM classdo_account 
				WHERE classdo_email = ANY($1) 
				AND deleted_at IS NULL`
	rows, err := s.LessonmgmtDB.Query(ctx, query, emails)
	if err != nil {
		return nil, fmt.Errorf("db.Query: %w", err)
	}
	defer rows.Close()

	var id, email pgtype.Text
	classDoAccounts := make(map[string]string, len(emails))
	for rows.Next() {
		if err := rows.Scan(&id, &email); err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}

		classDoAccounts[email.String] = id.String
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}

	return classDoAccounts, nil
}

func (s *Suite) getDeletedClassDoAccountsEmailsByEmail(ctx context.Context, emails []string) ([]string, error) {
	query := `SELECT classdo_email 
				FROM classdo_account 
				WHERE classdo_email = ANY($1) 
				AND deleted_at IS NOT NULL`
	rows, err := s.LessonmgmtDB.Query(ctx, query, emails)
	if err != nil {
		return nil, fmt.Errorf("db.Query: %w", err)
	}
	defer rows.Close()

	var email pgtype.Text
	classDoEmails := make([]string, 0, len(emails))
	for rows.Next() {
		if err := rows.Scan(&email); err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}

		classDoEmails = append(classDoEmails, email.String)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}

	return classDoEmails, nil
}
