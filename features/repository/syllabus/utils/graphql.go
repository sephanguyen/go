package utils

import (
	"bytes"
	"context"
	"crypto/tls"
	"html/template"
	"net/http"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/yasuo/constant"

	"github.com/hasura/go-graphql-client"
)

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

//nolint:deadcode
func TrackTableForHasuraQuery(queryPath, password string, tableNames ...string) error {
	httpClient := &http.Client{}
	rt := WithHeader(httpClient.Transport)
	rt.Set("content-type", "application/json")
	rt.Set("x-hasura-admin-secret", password)
	httpClient.Transport = rt

	type data struct {
		TableNames []string
		Role       string
	}

	t := template.Must(template.New("").Funcs(template.FuncMap{"separator": separator}).Parse(
		`{"type": "bulk", "args": [{{$s := separator ","}} {{range $tblname := .TableNames}}{{call $s}}
    	 {"type": "add_existing_table_or_view", "args": {"name": "{{$tblname}}", "schema": "public"}}{{end}}]}`))
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

//nolint:deadcode
func CreateSelectPermissionForHasuraQuery(queryPath, password string, tableNames ...string) error {
	httpClient := &http.Client{}
	rt := WithHeader(httpClient.Transport)
	rt.Set("content-type", "application/json")
	rt.Set("x-hasura-admin-secret", password)
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
					"filter":{},
					"limit":null,
					"allow_aggregations":false
				}
			}
		 }{{end}}]}`))
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

//nolint:deadcode
func AddQueryToAllowListForHasuraQuery(queryPath, query, password string) error {
	httpClient := &http.Client{}
	rt := WithHeader(httpClient.Transport)
	rt.Set("content-type", "application/json")
	rt.Set("x-hasura-admin-secret", password)
	httpClient.Transport = rt

	type data struct {
		CollectionID string
		Query        string
	}

	collectionID := idutil.ULIDNow()
	t := template.Must(template.New("").Parse(
		`{
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
		}`))
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
	defer resp.Body.Close()

	t = template.Must(template.New("").Parse(
		`{
			"type" : "add_collection_to_allowlist",
			"args": {
				"collection": "{{ .CollectionID}}"
					}
		}`))
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
	defer resp.Body.Close()

	return nil
}

func QueryHasura(ctx context.Context, hasuraAdminURL, password string, query interface{}, variables map[string]interface{}) error {
	httpClient := &http.Client{}
	rt := WithHeader(httpClient.Transport)
	rt.Set("content-type", "application/json")
	rt.Set("x-hasura-admin-secret", password)
	httpClient.Transport = rt
	client := graphql.NewClient(hasuraAdminURL+"/v1/graphql", httpClient)
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

/* #nosec */
//nolint:revive
func WithHeader(rt http.RoundTripper) withHeader {
	if rt == nil {
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
