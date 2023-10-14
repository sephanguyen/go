package accesscontrol

import (
	"bytes"
	"context"
	"crypto/tls"
	"html/template"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"sync"

	"github.com/hasura/go-graphql-client"
	"github.com/manabie-com/backend/features/helper"
	"github.com/manabie-com/backend/internal/golibs/idutil"

	"github.com/cucumber/godog"
)

const jsonFormat = "application/json"
const contentTypeHeaderStr = "content-type"
const hasuraHeader = "x-hasura-admin-secret"
const authorizationHeader = "Authorization"
const graphqlQueryEndpoint = "/v1/query"

var (
	buildRegexpMapOnce sync.Once
	regexpMap          map[string]*regexp.Regexp
)

func initSteps(ctx *godog.ScenarioContext, s *suite) {
	steps := map[string]interface{}{
		// new role MANABIE
		`^account admin hasura$`:                         s.accountAdminHasura,
		`^admin export hasura metadata$`:                 s.adminExportHasuraMetadata,
		`^admin sees the "([^"]*)" existed$`:             s.adminSeesTheExisted,
		`^"([^"]*)" name$`:                               s.name,
		`^admin sees the Manabie role$`:                  s.adminSeesTheManabieRole,
		`^columns inclued all columns from other roles$`: s.columnsIncludedAllColumnsFromOtherRoles,
		`^filters inclued all filters from other roles$`: s.filtersIncludedAllFiltersFromOtherRoles,

		// test AC for testing table
		`^login as userId "([^"]*)" and group "([^"]*)"$`:                             s.loginAsUserIdAndGroup,
		`^table ac_test_template_1 with location "([^"]*)" and permission "([^"]*)"$`: s.tableBWithLocationAndPermission,
		`^user assigned "([^"]*)" and "([^"]*)"$`:                                     s.userAssignedAnd,
		`^user get data from table ac_test_template_1$`:                               s.userGetDataFromTableB,
		`^user should only get "([^"]*)" their assigned$`:                             s.userShouldOnlyGetTheirAssigned,

		`^user "([^"]*)" data "([^"]*)" belong to location "([^"]*)" into table ac_test_template_1$`: s.userDataBelongToLocationIntoTableAcTestTemplate1,
		`^user "([^"]*)" data "([^"]*)" into table ac_test_template_1$`:                              s.userDataIntoTableAcTestTemplate1,
		`^return successfully$`: s.returnSuccess,
		`^return fail$`:         s.returnFail,

		`^login with user "([^"]*)"$`:                                          s.loginWithUser,
		`^table ac_test_template_4 with record "([^"]*)" and owner "([^"]*)"$`: s.tableAcTestTemplate4WithWithRecordAndOwner,
		`^user insert data "([^"]*)" into table ac_test_template_4$`:           s.userInsertDataIntoTableAcTestTemplate4,
		`^user "([^"]*)" data "([^"]*)" with name "([^"]*)"$`:                  s.userDataWithName,
		`^user get data from table ac_test_template_4$`:                        s.userGetDataFromTableAcTestTemplate4,
		`^user should only get "([^"]*)" their with signed$`:                   s.userShouldOnlyGetTheirWithSigned,
		`^command return fail$`:                                                s.commandReturnFail,
		`^command return successfully$`:                                        s.commandReturnSuccessfully,
		`^user insert data "([^"]*)" with owners "([^"]*)"$`:                   s.userInsertDataWithOwners,

		`^add data "([^"]*)" with owner is "([^"]*)" to table ac_test_template_11_4$`:                         s.addDataWithOwnerIsToTableAcTestTemplate11And4,
		`^add data "([^"]*)" with owner is "([^"]*)" with location "([^"]*)" to table ac_test_template_11_4$`: s.addDataWithOwnerIsWithLocationToTableAcTestTemplate11And4,
		`^add permission "([^"]*)" and "([^"]*)" to permission role "([^"]*)"$`:                               s.addPermissionAndToPermissionRole,
		`^add role "([^"]*)" and user group "([^"]*)" to granted role "([^"]*)"$`:                             s.addRoleAndUserGroupToGrantedRole,
		`^add user "([^"]*)" to user group "([^"]*)"$`:                                                        s.addUserToUserGroup,
		`^Assign location "([^"]*)" to granted role "([^"]*)"$`:                                               s.assignLocationToGrantedRole,
		`^command should return the records user is owners "([^"]*)"$`:                                        s.commandShouldReturnTheRecordsUserIsOwners,
		`^create location "([^"]*)" with access path "([^"]*)" with parent "([^"]*)"$`:                        s.createLocationWithAccessPathWithParent,
		`^create permission with name "([^"]*)"$`:                                                             s.createPermissionWithName,
		`^create role with name "([^"]*)"$`:                                                                   s.createRoleWithName,
		`^create user group name "([^"]*)"$`:                                                                  s.createUserGroupName,
		`^user get data "([^"]*)" from table ac_test_template_11_4$`:                                          s.userGetDataFromTableAcTestTemplate11And4,
		`^user update data "([^"]*)" with name "([^"]*)" to table ac_test_template_11_4$`:                     s.userUpdateDataWithNameToTableAcTestTemplate11And4,
		`^user delete data "([^"]*)" from table ac_test_template_11_4$`:                                       s.userDeleteDataFromTableAcTestTemplate11And4,

		`^mastermgmt hasura$`:       s.mastermgmtHasura,
		`^hasura return "([^"]*)"$`: s.hasuraReturn,
		`^table ac_hasura_test_template_(\d+) with location "([^"]*)" and permission "([^"]*)"$`:                      s.tableAcHasuraTestTemplate1WithLocationAndPermission,
		`^user "([^"]*)" data "([^"]*)" belong to location "([^"]*)" into table ac_hasura_test_template_1$`:           s.userDataBelongToLocationIntoTableAcHasuraTestTemplate1,
		`^user "([^"]*)" data "([^"]*)" belong to location "([^"]*)" into table ac_hasura_test_template_1 in hasura$`: s.userDataBelongToLocationIntoTableAcHasuraTestTemplate1InHasura,
		`^user get data "([^"]*)" from table ac_hasura_test_template_1 in hasura$`:                                    s.userGetDataFromTableAcHasuraTestTemplate1InHasura,
		`^user insert data "([^"]*)" belong to location "([^"]*)" into table ac_hasura_test_template_1 in hasura$`:    s.userInsertDataBelongToLocationIntoTableAcHasuraTestTemplate1InHasura,

		`^hasura table ac_test_template_4 with record "([^"]*)" and owner "([^"]*)"$`:        s.hasuraTableAcTestTemplate4WithRecordAndOwner,
		`^login hasura with user "([^"]*)"$`:                                                 s.loginHasuraWithUser,
		`^user get data "([^"]*)" from hasura table ac_test_template_4$`:                     s.userGetDataFromHasuraTableAcTestTemplate4,
		`^user insert data "([^"]*)" into hasura table ac_test_template_4$`:                  s.userInsertDataIntoHasuraTableAcTestTemplate4,
		`^user "([^"]*)" data "([^"]*)" with name "([^"]*)" into hasura ac_test_template_4$`: s.userDataWithNameIntoHasuraAcTestTemplate4,
		`^user insert data "([^"]*)" with owners "([^"]*)" into hasura ac_test_template_4$`:  s.userInsertDataWithOwnersIntoHasuraAcTestTemplate4,
		`^command return "([^"]*)" row affected$`:                                            s.commandReturnRowAffected,
	}

	buildRegexpMapOnce.Do(func() { regexpMap = helper.BuildRegexpMapV2(steps) })
	for k, v := range steps {
		ctx.Step(regexpMap[k], v)
	}
}

type withHeader struct {
	http.Header
	rt http.RoundTripper
}

func (h withHeader) RoundTrip(req *http.Request) (*http.Response, error) {
	for k, v := range h.Header {
		req.Header[k] = v
	}

	return h.rt.RoundTrip(req)
}

func newHeader(rt http.RoundTripper) withHeader {
	if rt == nil {
		/* #nosec */
		rt = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	return withHeader{Header: make(http.Header), rt: rt}
}

func (s *suite) ExportHasuraMetadata(url string, secret string) ([]byte, error) {
	httpClient := &http.Client{}
	rt := newHeader(httpClient.Transport)
	rt.Set(contentTypeHeaderStr, jsonFormat)
	rt.Set(hasuraHeader, secret)
	httpClient.Transport = rt

	type data struct {
		Metadata string
	}
	t := template.Must(template.New("").Parse(
		`
{
    "type" : "export_metadata",
    "args": {}
}
`))

	output := strings.Builder{}
	err := t.Execute(&output, data{})
	if err != nil {
		return nil, err
	}

	reqBytes := []byte(output.String())
	resp, err := httpClient.Post(url+graphqlQueryEndpoint, jsonFormat, bytes.NewBuffer(reqBytes))
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if err = resp.Body.Close(); err != nil {
		return nil, err
	}
	return body, nil
}

func parseBody(resp http.Response) ([]byte, error) {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if err = resp.Body.Close(); err != nil {
		return nil, err
	}
	return body, nil
}

func AddQueryToAllowListForHasuraQuery(hasuraAdminURL, query string) error {
	httpClient := &http.Client{}
	rt := newHeader(httpClient.Transport)
	rt.Set(contentTypeHeaderStr, jsonFormat)
	rt.Set(hasuraHeader, "M@nabie123")
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
	resp, err := httpClient.Post(hasuraAdminURL+graphqlQueryEndpoint, jsonFormat, bytes.NewBuffer(reqBytes))
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
	resp, err = httpClient.Post(hasuraAdminURL+graphqlQueryEndpoint, jsonFormat, bytes.NewBuffer(reqBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = parseBody(*resp)
	return err
}

func QueryHasura(ctx context.Context, hasuraAdminURL string, query interface{}, variables map[string]interface{}, token string) error {
	httpClient := &http.Client{}
	rt := newHeader(httpClient.Transport)
	rt.Set(contentTypeHeaderStr, jsonFormat)
	rt.Set(authorizationHeader, "Bearer "+token)
	httpClient.Transport = rt
	client := graphql.NewClient(hasuraAdminURL+"/v1/graphql", httpClient)
	err := client.Query(ctx, query, variables)
	if err != nil {
		return err
	}
	return nil
}

func MutateHasura(ctx context.Context, hasuraAdminURL string, m interface{}, variables map[string]interface{}, token string) error {
	httpClient := &http.Client{}
	rt := newHeader(httpClient.Transport)
	rt.Set(contentTypeHeaderStr, jsonFormat)
	rt.Set(authorizationHeader, "Bearer "+token)
	httpClient.Transport = rt
	client := graphql.NewClient(hasuraAdminURL+"/v1/graphql", httpClient)

	err := client.Mutate(ctx, m, variables)
	if err != nil {
		return err
	}
	return nil
}

type authInfoKey int

const (
	tokenKey authInfoKey = iota
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
