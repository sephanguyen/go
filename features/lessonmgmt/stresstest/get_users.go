package stresstest

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func (s *Suite) GetTeachersByHasura(ctx context.Context, jwt string, schoolID int32) ([]string, error) {
	// TODO: get query from deployments/helm/manabie-all-in-one/charts/bob/files/hasura/metadata/query_collections.yaml later
	body := []byte(fmt.Sprintf("{\"query\":\"query TeacherManyReference($limit: Int = 10, $offset: Int = 0, $email: String, $name: String, $school_id: Int! = 0) {\\n  find_teacher_by_school_id(\\n    limit: $limit\\n    offset: $offset\\n    order_by: {created_at: desc}\\n    args: {school_id: $school_id}\\n    where: {_or: [{name: {_ilike: $name}}, {email: {_ilike: $email}}]}\\n  ) {\\n    name\\n    email\\n    user_id\\n    avatar\\n  }\\n}\\n\",\"variables\":{\"limit\":30,\"offset\":0,\"school_id\":%d}}", schoolID))
	resp, err := s.st.QueryHasura(ctx, body, jwt)
	if err != nil {
		return nil, fmt.Errorf("QueryHasura: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("expected status 200 but got %d", resp.StatusCode)
	}

	res := struct {
		Data struct {
			FindTeacherBySchoolID []struct {
				Name   string  `json:"name"`
				Email  *string `json:"email"`
				UserID string  `json:"user_id"`
				Avatar *string `json:"avatar"`
			} `json:"find_teacher_by_school_id"`
		} `json:"data"`
	}{}
	if err = json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("decode response body: %w", err)
	}

	ids := make([]string, 0, len(res.Data.FindTeacherBySchoolID))
	for _, item := range res.Data.FindTeacherBySchoolID {
		ids = append(ids, item.UserID)
	}

	return ids, nil
}
