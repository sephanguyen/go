package accesscontrol

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hasura/go-graphql-client"
)

type GraphqlAcTemplate4MsgQuery struct {
	AcHasuraTestTemplate4 []struct {
		AcHasuraTestTemplate4ID string `graphql:"ac_test_template_4_id"`
	} `graphql:"ac_test_template_4(where: { ac_test_template_4_id: { _eq: $id} })"`
}

type GraphqlAcTemplate4MsgInsert struct {
	AcHasuraTestTemplate4 struct {
		AffectedRows int `graphql:"affected_rows"`
	} `graphql:"insert_ac_test_template_4(objects: [{ac_test_template_4_id: $id, owners: $owners}])"`
}

type GraphqlAcTemplate4MsgDelete struct {
	AcHasuraTestTemplate4 struct {
		AffectedRows int `graphql:"affected_rows"`
	} `graphql:"delete_ac_test_template_4(where: {ac_test_template_4_id: {_eq: $id}})"`
}

type GraphqlAcTemplate4MsgUpdate struct {
	AcHasuraTestTemplate4 struct {
		AffectedRows int `graphql:"affected_rows"`
	} `graphql:"update_ac_test_template_4(where: {ac_test_template_4_id: {_eq: $id}}, _set: {name: $name})"`
}

func insertAcTemplate4Hasura(ctx context.Context, id string, hasuraURL string, token string, owners string) (int64, error) {
	rawQuery := `mutation ($id: String!, $owners: String!) {
		insert_ac_test_template_4(objects: [{ac_test_template_4_id: $id, owners: $owners}]) {
		  affected_rows
		}
	  }`
	m := GraphqlAcTemplate4MsgInsert{}

	variables := map[string]interface{}{
		"id":     graphql.String(id),
		"owners": graphql.String(owners),
	}

	err := mutateHasura(ctx, hasuraURL, rawQuery, &m, variables, token)
	if err != nil {
		return 0, fmt.Errorf("insertAcTemplate4Hasura(): %v", err)
	}

	return int64(m.AcHasuraTestTemplate4.AffectedRows), nil
}

func deleteAcTemplate4Hasura(ctx context.Context, id string, hasuraURL string, token string) (int64, error) {
	rawQuery := `mutation ($id: String!) {
		delete_ac_test_template_4(where: {ac_test_template_4_id: {_eq: $id}}) {
		  affected_rows
		}
	  }`
	m := GraphqlAcTemplate4MsgDelete{}

	variables := map[string]interface{}{
		"id": graphql.String(id),
	}

	err := mutateHasura(ctx, hasuraURL, rawQuery, &m, variables, token)
	if err != nil {
		return 0, fmt.Errorf("deleteAcTemplate4Hasura(): %v", err)
	}

	return int64(m.AcHasuraTestTemplate4.AffectedRows), nil
}

func updateAcTemplate4Hasura(ctx context.Context, id string, hasuraURL string, token string, name string) (int64, error) {
	rawQuery := `mutation ($id: String!, $name: String!) {
		update_ac_test_template_4(where: {ac_test_template_4_id: {_eq: $id}}, _set: {name: $name}) {
		  affected_rows
		}
	  }`
	m := GraphqlAcTemplate4MsgUpdate{}

	variables := map[string]interface{}{
		"id":   graphql.String(id),
		"name": graphql.String(name),
	}

	err := mutateHasura(ctx, hasuraURL, rawQuery, &m, variables, token)
	if err != nil {
		return 0, fmt.Errorf("updateAcTemplate4Hasura(): %v", err)
	}

	return int64(m.AcHasuraTestTemplate4.AffectedRows), nil
}

func (s *suite) getACTemplate4TotalRecordsInHasura(ctx context.Context, id string, hasuraURL string, token string) (int, error) {
	rawQuery := `query ($id: String!) {
		ac_test_template_4(where: { ac_test_template_4_id: { _eq: $id} }) {
			ac_test_template_4_id
		}
	}`
	query := GraphqlAcTemplate4MsgQuery{}
	variables := map[string]interface{}{
		"id": graphql.String(id),
	}

	err := queryHasura(ctx, hasuraURL, rawQuery, &query, variables, token)
	if err != nil {
		return 0, fmt.Errorf("getTotalRecordsInHasura(): %v", err)
	}

	return len(query.AcHasuraTestTemplate4), nil
}

func (s *suite) hasuraTableAcTestTemplate4WithRecordAndOwner(ctx context.Context, data, owner string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	userGroup := "USER_GROUP_ADMIN"
	ctx, err := s.loginAsUserIdAndGroup(ctx, owner, userGroup)

	token, err := s.getDefaultUserInfo(ctx)
	if err != nil {
		return ctx, err
	}
	_, err = insertAcTemplate4Hasura(ctx, data, stepState.HasuraURL, token, owner)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) loginHasuraWithUser(ctx context.Context, userId string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	userGroup := "USER_GROUP_ADMIN"
	ctx, err := s.loginAsUserIdAndGroup(ctx, userId, userGroup)

	if err != nil {
		return ctx, fmt.Errorf("error when login hasura for template 4 verify %v", err)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userGetDataFromHasuraTableAcTestTemplate4(ctx context.Context, id string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.TotalRecords = 0
	token, err := s.getDefaultUserInfo(ctx)
	if err != nil {
		return ctx, err
	}

	totalRecords, err := s.getACTemplate4TotalRecordsInHasura(ctx, id, stepState.HasuraURL, token)
	if err != nil {
		return ctx, fmt.Errorf("get total records in hasura template 4 error %v", err)
	}

	stepState.TotalRecords = totalRecords

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userInsertDataIntoHasuraTableAcTestTemplate4(ctx context.Context, data string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	userID := stepState.CurrentUserID
	return s.hasuraTableAcTestTemplate4WithRecordAndOwner(ctx, data, userID)
}

func (s *suite) userDataWithNameIntoHasuraAcTestTemplate4(ctx context.Context, command, data, name string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	token, err := s.getDefaultUserInfo(ctx)
	if err != nil {
		return ctx, err
	}

	stepState.RowAffected = 0

	stepState.ErrInfo = nil
	switch command {
	case deleteCommand:
		stepState.RowAffected, err = deleteAcTemplate4Hasura(ctx, data, stepState.HasuraURL, token)
		if err != nil {
			return ctx, err
		}
	case updateCommand:
		stepState.RowAffected, err = updateAcTemplate4Hasura(ctx, data, stepState.HasuraURL, token, name)
		if err != nil {
			return ctx, err
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userInsertDataWithOwnersIntoHasuraAcTestTemplate4(ctx context.Context, data, owner string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	token, err := s.getDefaultUserInfo(ctx)
	if err != nil {
		return ctx, fmt.Errorf("get token error when insert data to template 4 %v", err)
	}

	rowsAffected, err := insertAcTemplate4Hasura(ctx, data, stepState.HasuraURL, token, owner)
	if checkConstrainError(err) {
		stepState.ErrInfo = err
	} else {
		return ctx, err
	}
	if rowsAffected > 0 {
		return ctx, fmt.Errorf("Command should not insert any record")
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) commandReturnRowAffected(ctx context.Context, number string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	affectedNum, _ := strconv.ParseInt(number, 10, 64)
	if stepState.RowAffected != affectedNum {
		return ctx, fmt.Errorf("the row affected not as expected %d/%d", stepState.RowAffected, affectedNum)
	}

	return StepStateToContext(ctx, stepState), nil
}
