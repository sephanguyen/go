package k8spoddetail

import (
	"github.com/spf13/cobra"
)

var (
	namespace string
)

// RootCmd for get pod detail command
var RootCmd = &cobra.Command{
	Use:   "poddetail",
	Short: "get detail on each pod of a/all namespace(s)",
	RunE:  getPodDetail,
}

func init() {
	RootCmd.PersistentFlags().StringVar(&namespace, "namespace", "", "namespace that you want to check detail, default is all namespaces")
}
