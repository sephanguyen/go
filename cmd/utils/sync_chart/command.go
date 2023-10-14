package syncchart

import "github.com/spf13/cobra"

var RootCmd = &cobra.Command{
	Use:   "sync_chart",
	Short: "Sync moved charts after updating chart from old namespace",
	Run:   syncChart,
}

var E2eLocalCmd = &cobra.Command{
	Use:   "e2e_local",
	Short: "Copy config from manabie local to e2e",
	Run:   syncAllChartsManaE2ELocal,
}
