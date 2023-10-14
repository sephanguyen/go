package sql

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/manabie-com/backend/cmd/samena/log"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/execwrapper"
	vr "github.com/manabie-com/backend/internal/golibs/variants"

	"github.com/spf13/cobra"
	"go.mozilla.org/sops/v3/decrypt"
	"gopkg.in/yaml.v3"
)

func NewCmdPsql() *cobra.Command {
	longHelper := `Run psql to connect to production database. Requires the proxy
to run first (with "samena proxy" or "cloud_sql_proxy" command).

By default, the current gcloud account is used to login. The current
account is determined by the following command:
  gcloud config get-value account

When using --username/-u flag for a service account, the ".gserviceaccount.com" domain suffix must be removed.

It will connect to "postgres" database at first. Use "\c" to change to your desired database.
`
	example := `
  # Run psql to stag.manabie (the default)
  samena psql

  # Run psql to uat.jprep
  samena psql jprep uat

  # Run psql to uat.jprep as postgres user
  samena psql -r jprep uat

  # Run psql as another user (or a service account)
  samena psql -u dev@manabie.com
  samena psql -u stag-draft@staging-manabie-online.iam 		# Note that ".gserviceaccount.com" suffix must be removed`

	cmd := &cobra.Command{
		Use:               "psql [PARTNER [ENVIRONMENT [INSTANCE]]]",
		Short:             "Run psql (and connect to the port from `samena proxy` command)",
		Long:              longHelper,
		Example:           example,
		Args:              cobra.MaximumNArgs(3),
		ValidArgsFunction: argsCompletion,
		RunE:              psql,
		SilenceUsage:      true,
	}
	cmd.Flags().BoolP(rootEnableFlag, "r", false, `login as postgres, has no effect if "-u" is specified`)
	cmd.Flags().StringP(usernameFlag, "u", "", `database username ("-U" in psql)`)
	cmd.Flags().BoolP(atlantisFlag, "a", false, `Use atlantis service account (similar to "-u atlantis@student-coach-e1e95.iam" option)`)
	return cmd
}

const (
	rootEnableFlag = "root"
	usernameFlag   = "username"
	atlantisFlag   = "atlantis"
)

func psql(cmd *cobra.Command, args []string) error {
	p, e, i := getTargetEnv(args)
	config, err := getCloudSQLConfig(p, e, i)
	if err != nil {
		return err
	}

	isRoot, err := cmd.Flags().GetBool(rootEnableFlag)
	if err != nil {
		return err
	}

	var dbUser, dbPassword string
	dbUser, err = cmd.Flags().GetString(usernameFlag)
	if err != nil {
		return err
	}

	useAtlantis, err := cmd.Flags().GetBool(atlantisFlag)
	if err != nil {
		return err
	}
	if useAtlantis {
		if isRoot || dbUser != "" {
			return fmt.Errorf("--%s cannot be used together with --%s or --%s", atlantisFlag, usernameFlag, rootEnableFlag)
		}
		dbUser = "atlantis@student-coach-e1e95.iam"
	}

	// if --username is not specified, fallback to --root flag or current logged in gcloud account
	if dbUser == "" {
		if isRoot {
			if e == vr.EnvLocal.String() || e == vr.EnvStaging.String() || e == vr.EnvUAT.String() {
				return fmt.Errorf("-r option can only be used in dorp or prod environment")
			}
			dbUser = "postgres"
			dbPassword, err = getPostgresPassword(p, e, i)
			if err != nil {
				return fmt.Errorf("failed to get password from secret: %s", err)
			}
		} else {
			dbUser, err = execwrapper.GCloudGetAccount()
			if err != nil {
				return fmt.Errorf("failed to get gcloud account: %w", err)
			}
		}
	}
	log.Info("logging in as: %s", dbUser)

	const saDomainSuffix string = ".gserviceaccount.com"
	if strings.HasSuffix(dbUser, saDomainSuffix) {
		dbUserNoSuffix := strings.TrimSuffix(dbUser, saDomainSuffix)
		log.Warn(`you have specified to login as service account %q, but the %q domain suffix should be removed when logging in
try specifying "--username=%s" instead`, dbUser, saDomainSuffix, dbUserNoSuffix)
	}
	return execwrapper.Psql("localhost", strconv.Itoa(config.port), dbUser, dbPassword)
}

func getPostgresPassword(partner, environment, instance string) (string, error) {
	if environment == vr.EnvPreproduction.String() {
		environment = vr.EnvProduction.String()
	}
	if environment != vr.EnvProduction.String() {
		return "", fmt.Errorf("invalid environment %q (must be dorp or prod)", environment)
	}
	var targetSvc string
	switch instance {
	case "common":
		targetSvc = "bob"
	case "lms":
		targetSvc = "eureka"
	default:
		return "", fmt.Errorf("invalid instance %q", instance)
	}
	fp := getMigrationSecretPath(partner, environment, targetSvc)
	content, err := decrypt.File(fp, "")
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %s", err)
	}

	pg := database.MigrationConfig{}
	err = yaml.Unmarshal(content, &pg)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal: %s", err)
	}

	if pg.PostgresMigrate.Database.Password == "" {
		return "", fmt.Errorf("password for postgres migration is empty")
	}

	return pg.PostgresMigrate.Database.Password, nil
}

func getMigrationSecretPath(partner, environment, service string) string {
	return filepath.Join(
		execwrapper.RootDirectory(),
		"deployments/helm/manabie-all-in-one/charts/",
		service, "secrets", partner, environment,
		fmt.Sprintf("%s_migrate.secrets.encrypted.yaml", service),
	)
}

func argsCompletion(_ *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) == 0 {
		return []string{"manabie ", "jprep ", "renseikai ", "synersia ", "aic ", "ga ", "tokyo "}, cobra.ShellCompDirectiveNoFileComp
	}
	if len(args) == 1 {
		if args[0] == "manabie" {
			return []string{"stag ", "uat "}, cobra.ShellCompDirectiveNoFileComp
		}
		return []string{"stag ", "uat ", "dorp ", "prod "}, cobra.ShellCompDirectiveNoFileComp
	}
	if len(args) == 2 {
		switch args[0] {
		case "manabie":
			return []string{"common ", "lms ", "auth "}, cobra.ShellCompDirectiveNoFileComp
		case "tokyo":
			return []string{"common ", "lms ", "dwh "}, cobra.ShellCompDirectiveNoFileComp
		default:
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
	}

	return nil, cobra.ShellCompDirectiveNoFileComp
}
