package accesscontrol

import (
	"context"
	"fmt"
	"strings"

	"github.com/hasura/go-graphql-client"
)

type GraphqlAcTemplate1MsgQuery struct {
	AcHasuraTestTemplate1 []struct {
		AcHasuraTestTemplate1ID string `graphql:"ac_hasura_test_template_1_id"`
	} `graphql:"ac_hasura_test_template_1(where: { ac_hasura_test_template_1_id: { _eq: $id} })"`
}

type GraphqlAcTemplate1MsgInsert struct {
	AcHasuraTestTemplate1 struct {
		AffectedRows int `graphql:"affected_rows"`
	} `graphql:"insert_ac_hasura_test_template_1(objects: [{ac_hasura_test_template_1_id: $id, location_id: $location}])"`
}

type GraphqlAcTemplate1MsgDelete struct {
	AcHasuraTestTemplate1 struct {
		AffectedRows int `graphql:"affected_rows"`
	} `graphql:"delete_ac_hasura_test_template_1(where: {ac_hasura_test_template_1_id: {_eq: $id}})"`
}

type GraphqlAcTemplate1MsgUpdate struct {
	AcHasuraTestTemplate1 struct {
		AffectedRows int `graphql:"affected_rows"`
	} `graphql:"update_ac_hasura_test_template_1(where: {ac_hasura_test_template_1_id: {_eq: $id}}, _set: {name: $name})"`
}

func queryHasura(ctx context.Context, hasuraURL string, rawQuery string, query interface{}, variables map[string]interface{}, token string) error {
	if err := AddQueryToAllowListForHasuraQuery(hasuraURL, rawQuery); err != nil {
		return fmt.Errorf("addQueryToAllowListForHasuraQuery(): %v", err)
	}

	err := QueryHasura(ctx, hasuraURL, query, variables, token)
	if err != nil {
		return fmt.Errorf("QueryHasura(): %v", err)
	}
	return nil
}

func mutateHasura(ctx context.Context, hasuraURL string, rawQuery string, query interface{}, variables map[string]interface{}, token string) error {
	if err := AddQueryToAllowListForHasuraQuery(hasuraURL, rawQuery); err != nil {
		return fmt.Errorf("addQueryToAllowListForHasuraQuery(): %v", err)
	}

	err := MutateHasura(ctx, hasuraURL, query, variables, token)
	if err != nil {
		return fmt.Errorf("mutateHasura(): %v", err)
	}
	return nil
}

func checkConstrainError(err error) bool {
	return strings.Contains(err.Error(), "Check constraint violation")
}

func insertAcTemplate1Hasura(ctx context.Context, id string, hasuraURL string, token string, location string) (int64, error) {
	rawQuery := `mutation ($id: String!, $location: String!) {
		insert_ac_hasura_test_template_1(objects: [{ac_hasura_test_template_1_id: $id, location_id: $location}]) {
		  affected_rows
		}
	  }`
	m := GraphqlAcTemplate1MsgInsert{}

	variables := map[string]interface{}{
		"id":       graphql.String(id),
		"location": graphql.String(location),
	}

	err := mutateHasura(ctx, hasuraURL, rawQuery, &m, variables, token)
	if err != nil && !checkConstrainError(err) {
		return 0, fmt.Errorf("insertAcTemplate1Hasura(): %v", err)
	}

	return int64(m.AcHasuraTestTemplate1.AffectedRows), nil
}

func deleteAcTemplate1Hasura(ctx context.Context, id string, hasuraURL string, token string) (int64, error) {
	rawQuery := `mutation ($id: String!) {
		delete_ac_hasura_test_template_1(where: {ac_hasura_test_template_1_id: {_eq: $id}}) {
		  affected_rows
		}
	  }`
	m := GraphqlAcTemplate1MsgDelete{}

	variables := map[string]interface{}{
		"id": graphql.String(id),
	}

	err := mutateHasura(ctx, hasuraURL, rawQuery, &m, variables, token)
	if err != nil {
		return 0, fmt.Errorf("deleteAcTemplate1Hasura(): %v", err)
	}

	return int64(m.AcHasuraTestTemplate1.AffectedRows), nil
}

func updateAcTemplate1Hasura(ctx context.Context, id string, hasuraURL string, token string) (int64, error) {
	rawQuery := `mutation ($id: String!, $name: String!) {
		update_ac_hasura_test_template_1(where: {ac_hasura_test_template_1_id: {_eq: $id}}, _set: {name: $name}) {
		  affected_rows
		}
	  }`
	m := GraphqlAcTemplate1MsgUpdate{}

	variables := map[string]interface{}{
		"id":   graphql.String(id),
		"name": graphql.String("updated_name"),
	}

	err := mutateHasura(ctx, hasuraURL, rawQuery, &m, variables, token)
	if err != nil {
		return 0, fmt.Errorf("updateAcTemplate1Hasura(): %v", err)
	}

	return int64(m.AcHasuraTestTemplate1.AffectedRows), nil
}

func (s *suite) getTotalRecordsInHasura(ctx context.Context, id string, hasuraURL string, token string) (int, error) {
	rawQuery := `query ($id: String!) {
		ac_hasura_test_template_1(where: { ac_hasura_test_template_1_id: { _eq: $id} }) {
			ac_hasura_test_template_1_id
		}
	}`
	query := GraphqlAcTemplate1MsgQuery{}
	variables := map[string]interface{}{
		"id": graphql.String(id),
	}

	err := queryHasura(ctx, hasuraURL, rawQuery, &query, variables, token)
	if err != nil {
		return 0, fmt.Errorf("getTotalRecordsInHasura(): %v", err)
	}

	return len(query.AcHasuraTestTemplate1), nil
}

func (s *suite) mastermgmtHasura(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.HasuraURL = s.Cfg.MastermgmtHasuraAdminURL

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) tableAcHasuraTestTemplate1WithLocationAndPermission(ctx context.Context, arg2, arg3 string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userDataBelongToLocationIntoTableAcHasuraTestTemplate1(ctx context.Context, arg1, arg2, arg3 string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userDataBelongToLocationIntoTableAcHasuraTestTemplate1InHasura(ctx context.Context, command, data, location string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	token, err := s.getDefaultUserInfo(ctx)
	if err != nil {
		return ctx, err
	}

	stepState.RowAffected = 0

	stepState.ErrInfo = nil
	switch command {
	case insertCommand:
		stepState.RowAffected, err = insertAcTemplate1Hasura(ctx, data, stepState.HasuraURL, token, location)
		if err != nil {
			return ctx, err
		}
	case deleteCommand:
		stepState.RowAffected, err = deleteAcTemplate1Hasura(ctx, data, stepState.HasuraURL, token)
		if err != nil {
			return ctx, err
		}
	case updateCommand:
		stepState.RowAffected, err = updateAcTemplate1Hasura(ctx, data, stepState.HasuraURL, token)
		if err != nil {
			return ctx, err
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) getDefaultUserInfo(ctx context.Context) (string, error) {
	stepState := StepStateFromContext(ctx)
	userID := stepState.CurrentUserID
	userGroup := "USER_GROUP_ADMIN"

	_, _ = s.aValidUser(ctx, withID(userID), withRole(userGroup))

	token, err := s.CommonSuite.GenerateExchangeToken(userID, userGroup)
	if err != nil {
		return "", err
	}
	return token, nil
}

func (s *suite) userGetDataFromTableAcHasuraTestTemplate1InHasura(ctx context.Context, data string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	token, err := s.getDefaultUserInfo(ctx)
	if err != nil {
		return ctx, err
	}

	stepState.AuthToken = token
	totalRecords, err := s.getTotalRecordsInHasura(ctx, data, stepState.HasuraURL, token)
	if err != nil {
		return ctx, err
	}

	stepState.TotalRecords = totalRecords
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) hasuraReturn(ctx context.Context, result string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if result == "successfully" && stepState.RowAffected == 0 {
		return ctx, stepState.ErrInfo
	}

	if result == "fail" && stepState.RowAffected > 0 {
		return ctx, fmt.Errorf("row affected should not greater than 0")
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userInsertDataBelongToLocationIntoTableAcHasuraTestTemplate1InHasura(ctx context.Context, data, location string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	token, err := s.getDefaultUserInfo(ctx)
	if err != nil {
		return ctx, err
	}

	insertAcTemplate1Hasura(ctx, data, stepState.HasuraURL, token, location)

	return StepStateToContext(ctx, stepState), nil
}
