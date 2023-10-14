package accesscontrol

import (
	"context"
	"fmt"
	"strconv"

	"github.com/jackc/pgtype"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
)

func insertAcTestTemplate4Table(ctx context.Context, id string, name string, owners string) (int64, error) {
	stmt := `INSERT INTO public.ac_test_template_4
						(ac_test_template_4_id, "name", owners, created_at, updated_at, deleted_at, resource_path)
						VALUES ($1, $2, $3, now(), now(), NULL,$4) on conflict (ac_test_template_4_id) do nothing;`
	var idText pgtype.Text
	_ = idText.Set(id)
	var nameText pgtype.Text
	_ = nameText.Set(name)
	var ownersText pgtype.Text
	_ = ownersText.Set(owners)
	rows, err := connections.MasterMgmtDB.Exec(ctx, stmt, idText, nameText, ownersText, getTextResourcePath())
	if err != nil {
		return 0, err
	}
	return rows.RowsAffected(), nil
}

func deleteAcTestTemplate4Table(ctx context.Context, id string) (int64, error) {
	stmt := `delete from public.ac_test_template_4 where ac_test_template_4_id = $1`
	var idText pgtype.Text
	_ = idText.Set(id)
	deletedRecords, err := connections.MasterMgmtDB.Exec(ctx, stmt, idText)
	if err != nil {
		return 0, err
	}
	return deletedRecords.RowsAffected(), nil
}

func updateAcTestTemplate4Table(ctx context.Context, id string, name string) (int64, error) {
	stmt := `update public.ac_test_template_4 set updated_at = now(), "name" = $1 where ac_test_template_4_id = $2`
	var idText pgtype.Text
	_ = idText.Set(id)
	var nameText pgtype.Text
	_ = nameText.Set(name)
	updatedRecords, err := connections.MasterMgmtDB.Exec(ctx, stmt, nameText, idText)
	if err != nil {
		return 0, err
	}
	return updatedRecords.RowsAffected(), nil
}

func (s *suite) getAcTestTemplate4(ctx context.Context) ([]string, error) {
	stmt := `select ac_test_template_4_id from ac_test_template_4 att`
	rows, err := connections.MasterMgmtDB.Query(ctx, stmt)
	if err != nil {
		return nil, fmt.Errorf("Get ac_test_template_4 error %v", err)
	}
	defer rows.Close()
	ids := []string{}
	for rows.Next() {
		acTestTemplate4Id := ""
		err := rows.Scan(&acTestTemplate4Id)
		if err != nil {
			return nil, fmt.Errorf("scan id fail. %v", err)
		}
		ids = append(ids, acTestTemplate4Id)
	}
	return ids, nil
}

func (s *suite) loginWithUser(ctx context.Context, user string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	claim := interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{
			ResourcePath: testResourcePath,
			UserID:       user,
		},
	}
	ctx = interceptors.ContextWithJWTClaims(ctx, &claim)
	stepState.CurrentUserID = user
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) tableAcTestTemplate4WithWithRecordAndOwner(ctx context.Context, record, user string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.loginWithUser(ctx, user)
	_, err = insertAcTestTemplate4Table(ctx, record, record, user)
	if err != nil {
		return ctx, err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userInsertDataIntoTableAcTestTemplate4(ctx context.Context, data string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	rowsAffected, err := insertAcTestTemplate4Table(ctx, data, data, stepState.CurrentUserID)
	if err != nil {
		return ctx, err
	}
	if rowsAffected == 0 {
		return ctx, fmt.Errorf("Can't insert data")
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userDataWithName(ctx context.Context, command, id, name string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.ErrInfo = nil
	switch command {
	case "insert":
		rowsAffected, err := insertAcTestTemplate4Table(ctx, id, name, stepState.CurrentUserID)
		if rowsAffected == 0 {
			return ctx, fmt.Errorf("Can't insert data")
		}
		stepState.ErrInfo = err
	case "delete":
		rowAffects, err := deleteAcTestTemplate4Table(ctx, id)
		if err != nil {
			return ctx, err
		}
		if rowAffects == 0 {
			stepState.ErrInfo = fmt.Errorf("User can't delete record %d", rowAffects)
		}
	case "update":
		rowAffects, err := updateAcTestTemplate4Table(ctx, id, name)
		if err != nil {
			return ctx, err
		}
		if rowAffects == 0 {
			stepState.ErrInfo = fmt.Errorf("User can't update record %d", rowAffects)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userGetDataFromTableAcTestTemplate4(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ids, err := s.getAcTestTemplate4(ctx)
	if err != nil {
		return ctx, fmt.Errorf("Error when get ac_test_template_4 %v", err)
	}
	stepState.TotalRecords = len(ids)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userShouldOnlyGetTheirWithSigned(ctx context.Context, totalRecords string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	numRecords, err := strconv.Atoi(totalRecords)
	if err != nil {
		return ctx, err
	}
	if stepState.TotalRecords != numRecords {
		return ctx, fmt.Errorf("Error when compare data from expected/actual : %d/%d", numRecords, stepState.TotalRecords)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) commandReturnFail(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ErrInfo == nil {
		return ctx, fmt.Errorf("Should have error when run command %v", stepState.ErrInfo)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) commandReturnSuccessfully(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ErrInfo != nil {
		return ctx, fmt.Errorf("Should not have error when run command %v", stepState.ErrInfo)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userInsertDataWithOwners(ctx context.Context, data, owners string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.ErrInfo = nil
	rowsAffected, err := insertAcTestTemplate4Table(ctx, data, data, owners)
	if invalidRLS(err) {
		stepState.ErrInfo = err
	} else {
		return ctx, err
	}
	if rowsAffected > 0 {
		return ctx, fmt.Errorf("Command should not insert any record")
	}
	return StepStateToContext(ctx, stepState), nil
}
