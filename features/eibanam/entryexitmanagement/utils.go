package entryexitmanagement

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/manabie-com/backend/features/helper"
	bob_entities "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	user_entities "github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/yasuo/constant"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/hasura/go-graphql-client"
	"google.golang.org/grpc/metadata"
)

type authInfoKey int

const (
	tokenKey          authInfoKey = iota
	hasuraAdminSecret             = "M@nabie123"
	queryPath                     = "/v1/query"
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

type userOptionInUser func(u *user_entities.LegacyUser)

func withIDInUser(id string) userOptionInUser {
	return func(u *user_entities.LegacyUser) {
		u.ID = database.Text(id)
	}
}

func withRoleInUser(group string) userOptionInUser {
	return func(u *user_entities.LegacyUser) {
		u.Group = database.Text(group)
	}
}

func withResourcePathInUser(resourcePath string) userOptionInUser {
	return func(u *user_entities.LegacyUser) {
		u.ResourcePath = database.Text(resourcePath)
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

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("generateAuthenticationToken: cannot read token from response, err: %v", err)
	}
	resp.Body.Close()

	return string(body), nil
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

func trackTableForHasuraQuery(tableNames ...string) error {
	httpClient := &http.Client{}
	rt := WithHeader(httpClient.Transport)
	rt.Set("content-type", "application/json")
	rt.Set("x-hasura-admin-secret", hasuraAdminSecret)
	httpClient.Transport = rt

	type data struct {
		TableNames []string
		Role       string
	}

	t := template.Must(template.New("").Funcs(template.FuncMap{"separator": separator}).Parse(`
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
	resp, err := httpClient.Post(bobHasuraAdminUrl+queryPath, "application/json", bytes.NewBuffer(reqBytes))
	if err != nil {
		return err
	}
	resp.Body.Close()

	return nil
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

func createSelectPermissionForHasuraQuery(tableNames ...string) error {
	httpClient := &http.Client{}
	rt := WithHeader(httpClient.Transport)
	rt.Set("content-type", "application/json")
	rt.Set("x-hasura-admin-secret", hasuraAdminSecret)
	httpClient.Transport = rt

	type data struct {
		TableNames []string
		Role       string
	}

	t := template.Must(template.New("").Funcs(template.FuncMap{"separator": separator}).Parse(
		`{"type": "bulk", "args": [{{$s := separator ","}} {{$role := .Role}} {{range $tblname := .TableNames}}{{call $s}}
    {"type":"create_select_permission","args":
        {"table":{"name":"{{$tblname}}","schema":"public"},"role":"{{$role}}","permission":
            {
                "columns":"*",
                "computed_fields":[],
                "backend_only":false,
                "filter":{"school_id":{"_in":"X-Hasura-School-Ids"}},
                "limit":null,
                "allow_aggregations":false
            }
        }
    }{{end}}]}
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
	resp, err := httpClient.Post(bobHasuraAdminUrl+queryPath, "application/json", bytes.NewBuffer(reqBytes))
	if err != nil {
		return err
	}
	resp.Body.Close()

	return nil
}

func addQueryToAllowListForHasuraQuery(query string) error {
	httpClient := &http.Client{}
	rt := WithHeader(httpClient.Transport)
	rt.Set("content-type", "application/json")
	rt.Set("x-hasura-admin-secret", hasuraAdminSecret)
	httpClient.Transport = rt

	type data struct {
		CollectionID string
		Query        string
	}

	collectionID := idutil.ULIDNow()
	t := template.Must(template.New("").Parse(`
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
	resp, err := httpClient.Post(bobHasuraAdminUrl+queryPath, "application/json", bytes.NewBuffer(reqBytes))
	if err != nil {
		return err
	}
	resp.Body.Close()

	t = template.Must(template.New("").Parse(`
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
	resp, err = httpClient.Post(bobHasuraAdminUrl+queryPath, "application/json", bytes.NewBuffer(reqBytes))
	if err != nil {
		return err
	}
	resp.Body.Close()

	return nil
}

func (s *suite) createStudentWithResourcePath(ctx context.Context) error {
	ctx = contextWithTokenForGrpcCall(s, ctx)
	// Make create student request
	s.Request = reqWithOnlyStudentInfo(int32(s.getSchoolId()))
	s.Response, s.ResponseErr = upb.NewUserModifierServiceClient(s.userMgmtConn).CreateStudent(ctx, s.Request.(*upb.CreateStudentRequest))

	if s.ResponseErr != nil {
		return s.ResponseErr
	}
	s.stepState.ResponseStack.Push(s.Response)
	s.stepState.RequestStack.Push(s.Request)

	return nil
}

func getContextJWTClaims(ctx context.Context, resourcePath string) context.Context {
	claim := interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{
			ResourcePath: resourcePath,
			DefaultRole:  entities.UserGroupSchoolAdmin,
			UserGroup:    entities.UserGroupSchoolAdmin,
		},
	}
	ctx = interceptors.ContextWithJWTClaims(ctx, &claim)
	return ctx
}
