package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/manabie-com/backend/internal/golibs/configs"
	"github.com/manabie-com/backend/j4/infras"

	j4 "github.com/manabie-com/j4/pkg/runner"

	"github.com/spf13/cobra"
)

var (
	configPath string
	dataPath   string
	secretPath string
	hostname   string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "j4",
		Short: "run stress test by j4",
		Long:  "run stress test by j4",
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, manabieCfg, err := loadConfigFromFileAndFlags()
			if err != nil {
				return fmt.Errorf("cannot load config: %s", err)
			}
			return Run(cmd.Context(), cfg, manabieCfg)
		},
	}

	// Arg definitions
	rootCmd.PersistentFlags().StringVar(
		&configPath,
		"configPath",
		"",
		"path to configuration file, usually used for configuration",
	)
	if err := rootCmd.MarkPersistentFlagRequired("configPath"); err != nil {
		log.Fatalf("error marking configPath: %s", err)
	}
	rootCmd.PersistentFlags().StringVar(
		&secretPath,
		"secretPath",
		"",
		"path to secret file, usually used for configuration",
	)

	rootCmd.PersistentFlags().StringVar(
		&dataPath,
		"dataPath",
		"",
		"path to the data directory set for RQLite",
	)

	rootCmd.PersistentFlags().StringVar(
		&hostname,
		"hostname",
		"",
		"hostname of the running machine",
	)

	// Run
	if err := rootCmd.ExecuteContext(context.Background()); err != nil {
		log.Printf("[ERROR] error executing command: %s", err)
		time.Sleep(-1)
	}
}

func loadConfigFromFileAndFlags() (*j4.Config, *infras.ManabieJ4Config, error) {
	cfg := &j4.Config{}

	configs.MustLoadConfig(
		context.Background(),
		configPath,
		configPath,
		secretPath,
		cfg,
	)
	manabieCfg := &infras.ManabieJ4Config{}
	configs.MustLoadConfig(
		context.Background(),
		configPath,
		configPath,
		secretPath,
		manabieCfg,
	)

	// If values are specified in command-line, override
	if len(dataPath) > 0 {
		cfg.RQLite.DataPath = dataPath
	}
	if len(hostname) > 0 {
		cfg.Hostname = hostname
	}
	return cfg, manabieCfg, nil
}
