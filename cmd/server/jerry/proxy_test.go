package jerry

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/manabie-com/backend/internal/golibs/bootstrap"
	"github.com/manabie-com/backend/internal/jerry/configurations"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestExtractQueryName(t *testing.T) {
	type tcase struct {
		full string
		name string
	}
	tcases := []tcase{
		{
			full: "\n query helloworld () some query",
			name: "helloworld",
		},
		{
			full: "\t mutation helloworld{}",
			name: "helloworld",
		},
		{
			full: "somestring",
		},
	}
	for _, c := range tcases {
		found := c.name != ""
		gotname, _, isfound := extractQueryName(c.full)
		assert.Equal(t, found, isfound)
		assert.Equal(t, c.name, gotname)
	}
}

type mockServer struct {
	port string
	s    *httptest.Server
}

func TestCache(t *testing.T) {
	req, err := http.NewRequest("POST", "localhost", nil)
	assert.NoError(t, err)
	req.Header.Set("authorization", `Bearer eyJhbGciOiJSUzI1NiIsImtpZCI6ImM2NzQ3NWMwM2NjNjQwMjY0YjBhOTRkMTQ0N2YzYjU3OTBiMmFiZDQifQ.eyJpc3MiOiJtYW5hYmllIiwic3ViIjoiMDFGUUdUQVlCNThDN1A0WUNBRjFHNUM1NTEiLCJhdWQiOiJtYW5hYmllLXN0YWciLCJleHAiOjE2NjkzNjk2ODMsImlhdCI6MTY2OTM2NjA3OSwianRpIjoiMDFHSlBaSFFWOTkzRFFGNVhLMUFHSDg4R0giLCJodHRwczovL2hhc3VyYS5pby9qd3QvY2xhaW1zIjp7IngtaGFzdXJhLWFsbG93ZWQtcm9sZXMiOlsiVVNFUl9HUk9VUF9TQ0hPT0xfQURNSU4iXSwieC1oYXN1cmEtZGVmYXVsdC1yb2xlIjoiVVNFUl9HUk9VUF9TQ0hPT0xfQURNSU4iLCJ4LWhhc3VyYS11c2VyLWlkIjoiMDFGUUdUQVlCNThDN1A0WUNBRjFHNUM1NTEiLCJ4LWhhc3VyYS1zY2hvb2wtaWRzIjoiey0yMTQ3NDgzNjQ4fSIsIngtaGFzdXJhLXVzZXItZ3JvdXAiOiJVU0VSX0dST1VQX1NDSE9PTF9BRE1JTiIsIngtaGFzdXJhLXJlc291cmNlLXBhdGgiOiItMjE0NzQ4MzY0OCJ9LCJtYW5hYmllIjp7ImFsbG93ZWRfcm9sZXMiOlsiVVNFUl9HUk9VUF9TQ0hPT0xfQURNSU4iXSwiZGVmYXVsdF9yb2xlIjoiVVNFUl9HUk9VUF9TQ0hPT0xfQURNSU4iLCJ1c2VyX2lkIjoiMDFGUUdUQVlCNThDN1A0WUNBRjFHNUM1NTEiLCJzY2hvb2xfaWRzIjpbIi0yMTQ3NDgzNjQ4Il0sInVzZXJfZ3JvdXAiOiJVU0VSX0dST1VQX1NDSE9PTF9BRE1JTiIsInJlc291cmNlX3BhdGgiOiItMjE0NzQ4MzY0OCJ9LCJyZXNvdXJjZV9wYXRoIjoiLTIxNDc0ODM2NDgiLCJ1c2VyX2dyb3VwIjoiVVNFUl9HUk9VUF9TQ0hPT0xfQURNSU4ifQ.ifYkefrU9kmSugxWn4D1Az4yvSwnyKdTzsGMF7IGqfQgS5M-inXqlKdn8uez-9JmWJfBmQgi0Wiy8b1B_cKLU-AnRVoDhLSDD1Z6ODL8MyXBSON341HXEjeBTZRIh1B-8nC42OkiWGC5y-Vb95ooWx5tQbfOrr-forDm9oHq_ZvnmqIv-1F5QP8szi8_w7OynXbQTiWWUKWKLHUaExVH1S1RlNt4yB_-jMzzMPA3NvkutDaQ8IqM0tV10845HWCm_i5XciafFRuGmW87e5sCrBTyBdo4yjQr42iRiNOdS59d_ydIsg12hHd6j2YNP403XlAgG9pD1UJI-ITKmtCDEw`)
	s := server{logger: zap.NewNop()}
	sessionVar, err := s.extractSessionVarFromReq(req)
	assert.NoError(t, err)
	assert.Equal(t, "-2147483648", sessionVar.ResourcePath)
	assert.Equal(t, "01FQGTAYB58C7P4YCAF1G5C551", sessionVar.UserID)
}

func TestProxy(t *testing.T) {
	servers := initMockHttpServers(t, 2)
	c := configurations.Config{}
	s1, s2 := servers[0], servers[1]
	c.HasuraRoutingRules = append(c.HasuraRoutingRules, configurations.RoutingRule{
		MatchedExacts: []string{"/1/match-1", "/1/match-2"},
		RewriteUri:    fmt.Sprintf("/rewritten-1"),
		ForwardHost:   "localhost",
		ForwardPort:   s1.port,
	})
	c.HasuraRoutingRules = append(c.HasuraRoutingRules, configurations.RoutingRule{
		MatchedPrefixes: []string{"/2/match-1", "/2/match-2"},
		RewriteUri:      fmt.Sprintf("/rewritten-2"),
		ForwardHost:     "localhost",
		ForwardPort:     s2.port,
	})
	engine := gin.New()
	l, err := net.Listen("tcp", ":0")
	assert.NoError(t, err)
	proxyUrl := fmt.Sprintf("http://localhost:%d", l.Addr().(*net.TCPAddr).Port)
	gserver := &server{}
	rsc := bootstrap.NewResources().WithLogger(zap.NewNop())
	assert.NoError(t, gserver.InitDependencies(c, rsc))
	engine.Any("/*proxyPath", gserver.proxy)
	go engine.RunListener(l)
	testProxy(t, proxyUrl, &s1, "/1/match-1", "/rewritten-1")
	testProxy(t, proxyUrl, &s1, "/1/match-2", "/rewritten-1")
	testProxy(t, proxyUrl, &s2, "/2/match-1/someapi", "/rewritten-2/someapi")
	testProxy(t, proxyUrl, &s2, "/2/match-2/someapi", "/rewritten-2/someapi")
}

// call original endpoint, check if proxied endpoint receive
// expect server behind proxy returning the full url
func testProxy(t *testing.T, proxyUrl string, s *mockServer, originalReq, proxyReq string) {
	resp, err := http.Get(proxyUrl + originalReq)
	assert.NoError(t, err)
	raw, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Equal(t, "localhost:"+s.port+proxyReq, string(raw))
}

func initMockHttpServers(t *testing.T, num int) (ret []mockServer) {
	for i := 0; i < num; i++ {
		l, err := net.Listen("tcp", ":0")
		assert.NoError(t, err)
		port := l.Addr().(*net.TCPAddr).Port
		ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fullPath := fmt.Sprintf("localhost:%d", port) + r.URL.Path
			fmt.Fprintf(w, fullPath)
		}))
		ts.Listener = l
		ts.Start()
		ret = append(ret, mockServer{port: fmt.Sprintf("%d", port), s: ts})
	}
	return ret
}
