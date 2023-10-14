package usermanagement

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"text/template"

	"github.com/manabie-com/backend/features/helper"
	bob_entities "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/yasuo/constant"

	"github.com/hasura/go-graphql-client"
	"google.golang.org/grpc/metadata"
)

type authInfoKey int

const (
	tokenKey authInfoKey = iota
)

type userOption func(u *bob_entities.User)

func withID(id string) userOption {
	return func(u *bob_entities.User) {
		u.ID = database.Text(id)
	}
}

func withRole(group string) userOption {
	return func(u *bob_entities.User) {
		u.Group = database.Text(group)
	}
}

func contextWithValidVersion(ctx context.Context) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "pkg", "com.manabie.liz", "version", "1.0.0")
}

func contextWithTokenForGrpcCall(s *suite, ctx context.Context) context.Context {
	authToken := s.UserGroupCredentials[s.UserGroupInContext].AuthToken
	return metadata.AppendToOutgoingContext(contextWithValidVersion(ctx), "token", authToken)
}

func contextWithToken(s *suite, ctx context.Context) context.Context {
	authToken := s.UserGroupCredentials[s.UserGroupInContext].AuthToken
	return context.WithValue(ctx, tokenKey, authToken)
}

func tokenFromContext(ctx context.Context) string {
	v := ctx.Value(tokenKey)
	s, _ := v.(string)
	return s
}

func generateExchangeToken(userID, userGroup string, schoolID int64) (string, error) {
	firebaseToken, err := generateValidAuthenticationToken(userID, userGroup)
	if err != nil {
		return "", err
	}
	token, err := helper.ExchangeToken(firebaseToken, userID, userGroup, applicantID, schoolID, shamirConn)
	if err != nil {
		return "", err
	}
	return token, nil
}

func generateValidAuthenticationToken(userID, userGroup string) (string, error) {
	return generateAuthenticationToken(userID, "templates/"+userGroup+".template")
}

func generateAuthenticationToken(userID string, template string) (string, error) {
	resp, err := http.Get("http://" + firebaseAddr + "/token?template=" + template + "&UserID=" + userID)
	if err != nil {
		return "", fmt.Errorf("generateAuthenticationToken:cannot generate new user token, err: %v", err)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("generateAuthenticationToken: cannot read token from response, err: %v", err)
	}
	resp.Body.Close()

	return string(b), nil
}

func loginFirebaseAccount(ctx context.Context, apiKey, email, password string) (string, error) {
	url := googleIdentityToolkitUrl + "/v1/accounts:signInWithPassword?key=" + apiKey

	loginInfo := struct {
		Email             string `json:"email"`
		Password          string `json:"password"`
		ReturnSecureToken bool   `json:"returnSecureToken"`
	}{
		Email:             email,
		Password:          password,
		ReturnSecureToken: true,
	}
	body, err := json.Marshal(&loginInfo)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}

	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errors.New("failed to login")
	}

	respBodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var idTokenResp struct {
		IdToken string `json:"idToken"`
	}
	err = json.Unmarshal(respBodyBytes, &idTokenResp)
	if err != nil {
		return "", err
	}

	return idTokenResp.IdToken, nil
}

func separator(s string) func() string {
	i := -1
	return func() string {
		i++
		if i == 0 {
			return ""
		}
		return s
	}
}

func trackTableForHasuraQuery(queryPath string, tableNames ...string) error {
	httpClient := &http.Client{}
	rt := WithHeader(httpClient.Transport)
	rt.Set("content-type", "application/json")
	rt.Set("x-hasura-admin-secret", "M@nabie123")
	httpClient.Transport = rt

	type data struct {
		TableNames []string
		Role       string
	}

	t := template.Must(template.New("").Funcs(template.FuncMap{"separator": separator}).Parse(
		`
{"type": "bulk", "args": [{{$s := separator ","}} {{range $tblname := .TableNames}}{{call $s}}
    {"type": "add_existing_table_or_view", "args": {"name": "{{$tblname}}", "schema": "public"}}{{end}}
]}
`))
	output := strings.Builder{}
	err := t.Execute(&output, data{
		TableNames: tableNames,
	})
	if err != nil {
		return err
	}
	reqBytes := []byte(output.String())
	resp, err := httpClient.Post(queryPath, "application/json", bytes.NewBuffer(reqBytes))
	if err != nil {
		return err
	}
	resp.Body.Close()

	return nil
}

func createSelectPermissionForHasuraQuery(queryPath string, tableNames ...string) error {
	httpClient := &http.Client{}
	rt := WithHeader(httpClient.Transport)
	rt.Set("content-type", "application/json")
	rt.Set("x-hasura-admin-secret", "M@nabie123")
	httpClient.Transport = rt

	type data struct {
		TableNames []string
		Role       string
	}

	t := template.Must(template.New("").Funcs(template.FuncMap{"separator": separator}).Parse(
		`
{"type": "bulk", "args": [{{$s := separator ","}} {{$role := .Role}} {{range $tblname := .TableNames}}{{call $s}}
    {"type":"create_select_permission","args":
        {"table":{"name":"{{$tblname}}","schema":"public"},"role":"{{$role}}","permission":
            {
                "columns":"*",
                "computed_fields":[],
                "backend_only":false,
                "filter":{},
                "limit":null,
                "allow_aggregations":false
            }
        }
    }{{end}}
]}
`))
	output := strings.Builder{}
	err := t.Execute(&output, data{
		TableNames: tableNames,
		Role:       constant.UserGroupSchoolAdmin,
	})
	if err != nil {
		return err
	}
	reqBytes := []byte(output.String())
	resp, err := httpClient.Post(queryPath, "application/json", bytes.NewBuffer(reqBytes))
	if err != nil {
		return err
	}
	resp.Body.Close()

	return nil
}

func addQueryToAllowListForHasuraQuery(queryPath string, query string) error {
	httpClient := &http.Client{}
	rt := WithHeader(httpClient.Transport)
	rt.Set("content-type", "application/json")
	rt.Set("x-hasura-admin-secret", "M@nabie123")
	httpClient.Transport = rt

	type data struct {
		CollectionID string
		Query        string
	}

	collectionID := idutil.ULIDNow()
	t := template.Must(template.New("").Parse(
		`
{
    "type" : "create_query_collection",
    "args": {
         "name": "{{ .CollectionID}}",
         "comment": "an optional comment",
         "definition": {
             "queries": [
                 {"name": "query_1", "query": "{{ .Query}}"}
              ]
         }
     }
}
`))
	output := strings.Builder{}
	err := t.Execute(&output, data{
		CollectionID: collectionID,
		Query:        query,
	})
	if err != nil {
		return err
	}
	reqBytes := []byte(output.String())
	resp, err := httpClient.Post(queryPath, "application/json", bytes.NewBuffer(reqBytes))
	if err != nil {
		return err
	}
	resp.Body.Close()

	t = template.Must(template.New("").Parse(
		`
{
	"type" : "add_collection_to_allowlist",
	"args": {
		 "collection": "{{ .CollectionID}}"
	 }
}
`))
	output = strings.Builder{}
	err = t.Execute(&output, data{
		CollectionID: collectionID,
		Query:        query,
	})
	if err != nil {
		return err
	}
	reqBytes = []byte(output.String())
	resp, err = httpClient.Post(queryPath, "application/json", bytes.NewBuffer(reqBytes))
	if err != nil {
		return err
	}
	resp.Body.Close()

	return nil
}

func queryHasura(ctx context.Context, query interface{}, variables map[string]interface{}, graphqlPath string) error {
	httpClient := &http.Client{}
	rt := WithHeader(httpClient.Transport)
	rt.Set("content-type", "application/json")
	if tokenFromContext(ctx) != "" {
		rt.Set("Authorization", "Bearer "+tokenFromContext(ctx))
	}
	httpClient.Transport = rt
	client := graphql.NewClient(graphqlPath, httpClient)
	err := client.Query(ctx, query, variables)
	if err != nil {
		return err
	}
	return nil
}

type withHeader struct {
	http.Header
	rt http.RoundTripper
}

func WithHeader(rt http.RoundTripper) withHeader {
	if rt == nil {
		/* #nosec */
		rt = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	return withHeader{Header: make(http.Header), rt: rt}
}

func (h withHeader) RoundTrip(req *http.Request) (*http.Response, error) {
	for k, v := range h.Header {
		req.Header[k] = v
	}

	return h.rt.RoundTrip(req)
}

func strSliceEq(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
