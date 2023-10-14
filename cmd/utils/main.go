//go:build (linux && 386) || (darwin && !cgo)

package main

import (
	"os"

	"github.com/manabie-com/backend/cmd/utils/auth"
	"github.com/manabie-com/backend/cmd/utils/coverage"
	dplparser "github.com/manabie-com/backend/cmd/utils/data_pipeline_parser"
	"github.com/manabie-com/backend/cmd/utils/firebase"
	"github.com/manabie-com/backend/cmd/utils/grafana"
	poddetail "github.com/manabie-com/backend/cmd/utils/k8s_pod_detail"
	migrationsdata "github.com/manabie-com/backend/cmd/utils/migrations_data"
	"github.com/manabie-com/backend/cmd/utils/mock"
	rls "github.com/manabie-com/backend/cmd/utils/rls"
	"github.com/manabie-com/backend/cmd/utils/sqlparser"
	syncchart "github.com/manabie-com/backend/cmd/utils/sync_chart"
	"github.com/manabie-com/backend/cmd/utils/tiertest"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"

	"github.com/spf13/cobra"
)

var _ = cpb.Country(0)

func main() {
	rootCmd := &cobra.Command{Use: "utils [command]"}
	syncchart.RootCmd.AddCommand(syncchart.E2eLocalCmd)
	rootCmd.AddCommand(
		firebase.RootCmd,
		mock.RootCmd,
		coverage.RootCmd,
		auth.RootCmd,
		sqlparser.RootCmd,
		dplparser.RootCmd,
		rls.RootCmd,
		poddetail.RootCmd,
		grafana.RootCmd,
		tiertest.RootCmd,
		migrationsdata.RootCmd,
		syncchart.RootCmd,
	)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
