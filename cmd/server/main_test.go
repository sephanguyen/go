package main

import (
	"context"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/manabie-com/backend/cmd/custom_lint/sqlclosecheck"
	_ "github.com/manabie-com/backend/internal/golibs/automaxprocs"
	"golang.org/x/tools/go/analysis/analysistest"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// TestMain is the entrypoint for test-mode binaries
// so that integration tests can have code-coverage profiles
func TestCoverage(t *testing.T) {
	// remove non-cobra command flags
	stripTestArgs()

	mainCtx, cancel := context.WithCancel(context.Background())

	// start kill server
	killServer := newKillServer(":19999", cancel)
	go killServer.Start()
	defer killServer.server.Shutdown(context.Background())

	go func() {
		makeRootCmd()
		if err := rootCmd.ExecuteContext(mainCtx); err != nil {
			fmt.Println("rootCmd.ExecuteContext Error:", err)
		}
	}()

	<-mainCtx.Done()

	fmt.Println("TestCoverage finished")
}

// stripTestArgs removes -test. args so cobra doesn't break
func stripTestArgs() {
	newArgs := []string{}
	for _, arg := range os.Args {
		if !strings.HasPrefix(arg, "-test.") {
			newArgs = append(newArgs, arg)
		}
	}
	os.Args = newArgs
}

// killServer is an HTTP server that kills the process once it receives a request
// this is needed to generate code coverage after running integration tests
type killServer struct {
	server http.Server
	cancel context.CancelFunc
}

func newKillServer(addr string, cancel context.CancelFunc) *killServer {
	return &killServer{
		server: http.Server{
			Addr: addr,
		},
		cancel: cancel,
	}
}

func (s *killServer) Start() {
	s.server.Handler = s

	fmt.Println("Started KillServer")
	err := s.server.ListenAndServe()

	if err == http.ErrServerClosed {
		fmt.Println("KillServer Closed")
	} else {
		fmt.Println("KillServer Error:", err)
	}

	// wait for http server
}

func (s *killServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)

	s.cancel()
}

func TestSqlCloseCheck(t *testing.T) {
	testdata, _ := filepath.Abs("../custom_lint/testdata")

	checker := sqlclosecheck.NewAnalyzer()
	analysistest.Run(t, testdata, checker, "rows")
}
