package stresstest

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func (s *Suite) GetLocationsByHasura(ctx context.Context, jwt string, locationIDs []string) ([]string, error) {
	// TODO: get query from deployments/helm/manabie-all-in-one/charts/bob/files/hasura/metadata/query_collections.yaml later
	payload := struct {
		Query     string `json:"query"`
		Variables struct {
			LocationIDs []string `json:"location_ids"`
		} `json:"variables"`
	}{
		Query: "query LocationListByIds($location_ids: [String!] = []) {\n  locations(where: {location_id: {_in: $location_ids}}) {\n    name\n    location_id\n  }\n}\n",
		Variables: struct {
			LocationIDs []string `json:"location_ids"`
		}{
			LocationIDs: locationIDs,
		},
	}
	body, err := json.Marshal(&payload)
	if err != nil {
		return nil, fmt.Errorf("could not Marshal payload to get centers by hasura query")
	}
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
			Locations []struct {
				Name       string `json:"name"`
				LocationId string `json:"location_id"`
			} `json:"locations"`
		} `json:"data"`
	}{}
	if err = json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("decode response body: %w", err)
	}

	ids := make([]string, 0, len(res.Data.Locations))
	for _, location := range res.Data.Locations {
		ids = append(ids, location.LocationId)
	}

	return ids, nil
}
