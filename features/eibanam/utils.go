package eibanam

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/manabie-com/backend/features/helper"
	"github.com/manabie-com/backend/internal/golibs/idutil"

	"github.com/hasura/go-graphql-client"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type authInfoKey int

const (
	tokenKey authInfoKey = iota
)

func GetExampleHasuraMetadata(path string) ([]byte, error) {
	if len(path) == 0 {
		path = "./eibanam/example_hasura_metadata.json"
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func ContextWithTokenForGrpcCall(ctx context.Context, authToken string) context.Context {
	return metadata.AppendToOutgoingContext(contextWithValidVersion(ctx), "token", authToken)
}

func contextWithValidVersion(ctx context.Context) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "pkg", "com.manabie.liz", "version", "1.0.0")
}

func ContextWithToken(ctx context.Context, authToken string) context.Context {
	return context.WithValue(ctx, tokenKey, authToken)
}

func GenerateExchangeToken(firebaseAddr, userID, userGroup, applicantID string, schoolID int32, shamirConn grpc.ClientConnInterface) (string, error) {
	firebaseToken, err := generateValidAuthenticationToken(firebaseAddr, userID, userGroup)
	if err != nil {
		return "", err
	}
	token, err := helper.ExchangeToken(firebaseToken, userID, userGroup, applicantID, int64(schoolID), shamirConn)
	if err != nil {
		return "", err
	}
	return token, nil
}

func generateValidAuthenticationToken(firebaseAddr, userID, userGroup string) (string, error) {
	return generateAuthenticationToken(firebaseAddr, userID, "templates/"+userGroup+".template")
}

func generateAuthenticationToken(firebaseAddr, userID, template string) (string, error) {
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

func LoginFirebaseAccount(ctx context.Context, googleIdentityToolkitUrl, apiKey, email, password string) (string, error) {
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

func TrackTableForHasuraQuery(hasuraAdminUrl string, tableNames ...string) error {
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
	resp, err := httpClient.Post(hasuraAdminUrl+"/v1/query", "application/json", bytes.NewBuffer(reqBytes))
	if err != nil {
		return err
	}
	resp.Body.Close()

	return nil
}

func CreateSelectPermissionForHasuraQuery(hasuraAdminUrl, role string, tableNames ...string) error {
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
		Role:       role,
	})
	if err != nil {
		return err
	}
	reqBytes := []byte(output.String())
	resp, err := httpClient.Post(hasuraAdminUrl+"/v1/query", "application/json", bytes.NewBuffer(reqBytes))
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

func AddQueryToAllowListForHasuraQuery(hasuraAdminUrl, query string) error {
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
	resp, err := httpClient.Post(hasuraAdminUrl+"/v1/query", "application/json", bytes.NewBuffer(reqBytes))
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
	resp, err = httpClient.Post(hasuraAdminUrl+"/v1/query", "application/json", bytes.NewBuffer(reqBytes))
	if err != nil {
		return err
	}
	resp.Body.Close()

	return nil
}

func ReplaceHasuraMetadata(hasuraAdminUrl string, metadata string) error {
	httpClient := &http.Client{}
	rt := WithHeader(httpClient.Transport)
	rt.Set("content-type", "application/json")
	rt.Set("x-hasura-admin-secret", "M@nabie123")
	httpClient.Transport = rt

	type data struct {
		Metadata string
	}
	t := template.Must(template.New("").Parse(
		`
{
    "type" : "replace_metadata",
    "args": {{ .Metadata}}
}
`))

	output := strings.Builder{}
	err := t.Execute(&output, data{
		Metadata: metadata,
	})
	if err != nil {
		return err
	}
	reqBytes := []byte(output.String())
	resp, err := httpClient.Post(hasuraAdminUrl+"/v1/query", "application/json", bytes.NewBuffer(reqBytes))
	if err != nil {
		return err
	}
	if err = resp.Body.Close(); err != nil {
		return err
	}

	return nil
}

func QueryHasura(ctx context.Context, hasuraAdminUrl string, query interface{}, variables map[string]interface{}) error {
	httpClient := &http.Client{}
	rt := WithHeader(httpClient.Transport)
	rt.Set("content-type", "application/json")
	if tokenFromContext(ctx) != "" {
		rt.Set("Authorization", "Bearer "+tokenFromContext(ctx))
	}
	httpClient.Transport = rt
	client := graphql.NewClient(hasuraAdminUrl+"/v1/graphql", httpClient)
	err := client.Query(ctx, query, variables)
	if err != nil {
		return err
	}
	return nil
}

func QueryRawHasura(ctx context.Context, hasuraAdminUrl string, query interface{}, variables map[string]interface{}) (*json.RawMessage, error) {
	httpClient := &http.Client{}
	rt := WithHeader(httpClient.Transport)
	rt.Set("content-type", "application/json")
	if tokenFromContext(ctx) != "" {
		rt.Set("Authorization", "Bearer "+tokenFromContext(ctx))
	}
	httpClient.Transport = rt
	client := graphql.NewClient(hasuraAdminUrl+"/v1/graphql", httpClient)
	res, err := client.QueryRaw(ctx, query, variables)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func tokenFromContext(ctx context.Context) string {
	v := ctx.Value(tokenKey)
	s, _ := v.(string)
	return s
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

// TryUntilSuccess must be used with context.WithTimeout
func TryUntilSuccess(ctx context.Context, tryInterval time.Duration, tryFn func(ctx context.Context) (bool, error)) error {
	ticker := time.NewTicker(tryInterval)
	defer ticker.Stop()

	errCh := make(chan error, 1)
	go func(ctx context.Context) {
		for {
			select {
			case <-ticker.C:
				retry, err := tryFn(ctx)
				if retry {
					continue
				}
				select {
				case errCh <- err:
				case <-time.After(time.Second):
				}
				return
			case <-ctx.Done():
				select {
				case errCh <- ctx.Err():
				case <-time.After(time.Second):
				}
				return
			}
		}
	}(ctx)

	return <-errCh
}
