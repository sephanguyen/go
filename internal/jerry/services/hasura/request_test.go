package hasura

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestParseRequest(t *testing.T) {
	t.Parallel()

	e := NewEavesdropper("", zap.NewNop())
	res, err := e.parseRequest(
		[]byte(`{"query":"query GetManyGrantedPermissions($permissions: [String!]!) {\\n  granted_permissions(where: {permission_name: {_in: $permissions}}) {\\n    permission_name\\n    location_id\\n    permission_id\\n    user_id\\n  }\\n}\\n","variables":{"permissions":["payment.order.write"]}}`),
	)
	require.NoError(t, err)
	require.Equal(t, &Request{
		Query:     `query GetManyGrantedPermissions($permissions: [String!]!) {\n  granted_permissions(where: {permission_name: {_in: $permissions}}) {\n    permission_name\n    location_id\n    permission_id\n    user_id\n  }\n}\n`,
		Variables: []byte(`{"permissions":["payment.order.write"]}`),
		queryType: "query",
		queryName: "GetManyGrantedPermissions",
	}, res)
}

func TestGetHasuraJWTFromHeader(t *testing.T) {
	t.Parallel()

	e := NewEavesdropper("", zap.NewNop())
	h := http.Header{}
	h.Set("authorization", `Bearer eyJhbGciOiJSUzI1NiIsImtpZCI6ImM2NzQ3NWMwM2NjNjQwMjY0YjBhOTRkMTQ0N2YzYjU3OTBiMmFiZDQifQ.eyJpc3MiOiJtYW5hYmllIiwic3ViIjoiMDFGUUdUQVlCNThDN1A0WUNBRjFHNUM1NTEiLCJhdWQiOiJtYW5hYmllLXN0YWciLCJleHAiOjE2NjkzNjk2ODMsImlhdCI6MTY2OTM2NjA3OSwianRpIjoiMDFHSlBaSFFWOTkzRFFGNVhLMUFHSDg4R0giLCJodHRwczovL2hhc3VyYS5pby9qd3QvY2xhaW1zIjp7IngtaGFzdXJhLWFsbG93ZWQtcm9sZXMiOlsiVVNFUl9HUk9VUF9TQ0hPT0xfQURNSU4iXSwieC1oYXN1cmEtZGVmYXVsdC1yb2xlIjoiVVNFUl9HUk9VUF9TQ0hPT0xfQURNSU4iLCJ4LWhhc3VyYS11c2VyLWlkIjoiMDFGUUdUQVlCNThDN1A0WUNBRjFHNUM1NTEiLCJ4LWhhc3VyYS1zY2hvb2wtaWRzIjoiey0yMTQ3NDgzNjQ4fSIsIngtaGFzdXJhLXVzZXItZ3JvdXAiOiJVU0VSX0dST1VQX1NDSE9PTF9BRE1JTiIsIngtaGFzdXJhLXJlc291cmNlLXBhdGgiOiItMjE0NzQ4MzY0OCJ9LCJtYW5hYmllIjp7ImFsbG93ZWRfcm9sZXMiOlsiVVNFUl9HUk9VUF9TQ0hPT0xfQURNSU4iXSwiZGVmYXVsdF9yb2xlIjoiVVNFUl9HUk9VUF9TQ0hPT0xfQURNSU4iLCJ1c2VyX2lkIjoiMDFGUUdUQVlCNThDN1A0WUNBRjFHNUM1NTEiLCJzY2hvb2xfaWRzIjpbIi0yMTQ3NDgzNjQ4Il0sInVzZXJfZ3JvdXAiOiJVU0VSX0dST1VQX1NDSE9PTF9BRE1JTiIsInJlc291cmNlX3BhdGgiOiItMjE0NzQ4MzY0OCJ9LCJyZXNvdXJjZV9wYXRoIjoiLTIxNDc0ODM2NDgiLCJ1c2VyX2dyb3VwIjoiVVNFUl9HUk9VUF9TQ0hPT0xfQURNSU4ifQ.ifYkefrU9kmSugxWn4D1Az4yvSwnyKdTzsGMF7IGqfQgS5M-inXqlKdn8uez-9JmWJfBmQgi0Wiy8b1B_cKLU-AnRVoDhLSDD1Z6ODL8MyXBSON341HXEjeBTZRIh1B-8nC42OkiWGC5y-Vb95ooWx5tQbfOrr-forDm9oHq_ZvnmqIv-1F5QP8szi8_w7OynXbQTiWWUKWKLHUaExVH1S1RlNt4yB_-jMzzMPA3NvkutDaQ8IqM0tV10845HWCm_i5XciafFRuGmW87e5sCrBTyBdo4yjQr42iRiNOdS59d_ydIsg12hHd6j2YNP403XlAgG9pD1UJI-ITKmtCDEw`)
	res, err := e.getHasuraJWTFromHeader(h)
	require.NoError(t, err)
	require.Equal(t, &JWT{
		DefaultRole:  "USER_GROUP_SCHOOL_ADMIN",
		UserID:       "01FQGTAYB58C7P4YCAF1G5C551",
		UserGroup:    "USER_GROUP_SCHOOL_ADMIN",
		ResourcePath: "-2147483648",
	}, res)
}
