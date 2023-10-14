package sql

import (
	"fmt"

	"github.com/manabie-com/backend/cmd/samena/log"
	"github.com/manabie-com/backend/internal/golibs/execwrapper"

	"github.com/spf13/cobra"
)

func NewCmdCloudSQLProxy() *cobra.Command {
	longHelper := `
Run cloud_sql_proxy command to connect to staging/uat/production Cloud SQL instance.
This command only initiates the proxy connection. To actually connect and execute SQL statements,
a conjunction "samena psql" can be run in parallel (i.e. in a separate terminal), or use your own GUI.

By default, the command generates the token using the currently logged in GCloud account, determined by:
  gcloud config get-value account

If you want to login as another email (usually to impersonate a service account), use the --username/-u flag.
Note that you need to specify the full email of the service account here, while for "samena psql",
the domain suffix ".gserviceaccount.com" must be removed for the same --username/-u flag.
`
	example := `  # Run cloud_sql_proxy for a specific env/org (default is stag-manabie)
  samena proxy
  samena proxy jprep prod
  samena proxy manabie stag common

  # Run as another user
  samena proxy -u dev@manabie.com
  samena proxy -u stag-draft@staging-manabie-online.iam.gserviceaccount.com`

	cmd := &cobra.Command{
		Use:     "proxy [PARTNER [ENVIRONMENT [INSTANCE]]]",
		Short:   "Run `cloud_sql_proxy`",
		Long:    longHelper,
		Example: example,
		RunE: func(cmd *cobra.Command, args []string) error {
			p, e, i := getTargetEnv(args)
			config, err := getCloudSQLConfig(p, e, i)
			if err != nil {
				return fmt.Errorf("failed to get Cloud SQL config: %s", err)
			}

			proxyArgs := []string{
				"-instances", config.instancesFQN(),
				"-enable_iam_login",
			}

			if p == "manabie" || p == "tokyo" {
				log.Info("Creating proxy to \"%s.%s.%s\"", e, p, i)
			} else {
				log.Info("Creating proxy to \"%s.%s\"", e, p)
			}
			log.Info("Connection string: %s", config.connectionString)
			log.Info("Port: %d", config.port)

			dbUser, err := cmd.Flags().GetString(usernameFlag)
			if err != nil {
				return err
			}
			useAtlantis, err := cmd.Flags().GetBool(atlantisFlag)
			if err != nil {
				return err
			}
			if useAtlantis {
				if dbUser != "" {
					return fmt.Errorf("--%s cannot be used together with --%s", atlantisFlag, usernameFlag)
				}
				dbUser = "atlantis@student-coach-e1e95.iam.gserviceaccount.com"
			}

			if dbUser != "" {
				log.Info("Using token from account: %s", dbUser)
				token, err := execwrapper.GCloudPrintAccessTokenOf(dbUser) // token is sensitive, do not print
				if err != nil {
					return fmt.Errorf("failed to get token of service account: %s", err)
				}
				proxyArgs = append(proxyArgs, fmt.Sprintf("--token=%q", token))
			}
			return execwrapper.CloudSQLProxyCommand(proxyArgs...)
		},
		Args:              cobra.MaximumNArgs(3),
		ValidArgsFunction: argsCompletion,
		SilenceUsage:      true,
	}
	cmd.Flags().StringP(usernameFlag, "u", "", `GCloud account to generate token for "cloud_sql_proxy -token"
If a service account is specified, you must have "Service Account User" and "Service Account Token Creator" permission`)
	cmd.Flags().BoolP(atlantisFlag, "a", false, `Use atlantis service account (similar to "-u atlantis@student-coach-e1e95.iam.gserviceaccount.com" option)`)
	return cmd
}

func getTargetEnv(args []string) (string, string, string) {
	defaultPartner := "manabie" //nolint:goconst
	defaultEnv := "stag"
	defaultInstance := "common" //nolint:goconst
	if len(args) == 0 {
		return defaultPartner, defaultEnv, defaultInstance
	}
	if len(args) == 1 {
		return args[0], defaultEnv, defaultInstance
	}
	if len(args) == 2 {
		return args[0], args[1], defaultInstance
	}
	return args[0], args[1], args[2]
}

type cloudSQLconfig struct {
	connectionString string
	port             int
}

func (c *cloudSQLconfig) instancesFQN() string {
	return fmt.Sprintf("%s=tcp:%d", c.connectionString, c.port)
}

func getCloudSQLConfig(partner, environment, instance string) (*cloudSQLconfig, error) {
	var config *cloudSQLconfig
	switch environment {
	case "stag":
		switch partner {
		case "manabie":
			switch instance {
			case "common":
				config = &cloudSQLconfig{
					connectionString: "staging-manabie-online:asia-southeast1:manabie-common-88e1ee71",
					port:             10011,
				}
			case "lms":
				config = &cloudSQLconfig{
					connectionString: "staging-manabie-online:asia-southeast1:manabie-lms-de12e08e",
					port:             10012,
				}
			case "auth":
				config = &cloudSQLconfig{
					connectionString: "staging-manabie-online:asia-southeast1:manabie-auth-f2dc7988",
					port:             10013,
				}
			default:
				return nil, fmt.Errorf("invalid instance %q", instance)
			}
		case "jprep": //nolint:goconst
			log.Warn("stag.jprep is using the same database instance with uat.jprep")
			config = &cloudSQLconfig{
				connectionString: "staging-manabie-online:asia-southeast1:jprep-uat",
				port:             10002,
			}
		}
	case "uat":
		switch partner {
		case "manabie":
			log.Warn("uat.manabie is using the same database instance with stag.manabie")
			switch instance {
			case "common":
				config = &cloudSQLconfig{
					connectionString: "staging-manabie-online:asia-southeast1:manabie-common-88e1ee71",
					port:             10011,
				}
			case "lms":
				config = &cloudSQLconfig{
					connectionString: "staging-manabie-online:asia-southeast1:manabie-lms-de12e08e",
					port:             10012,
				}
			default:
				return nil, fmt.Errorf("invalid instance %q", instance)
			}
		case "jprep":
			config = &cloudSQLconfig{
				connectionString: "staging-manabie-online:asia-southeast1:jprep-uat",
				port:             10002,
			}
		}
	case "dorp":
		switch partner {
		case "jprep":
			config = &cloudSQLconfig{
				connectionString: "student-coach-e1e95:asia-northeast1:clone-jprep-6a98",
				port:             20002,
			}
		case "synersia":
			config = &cloudSQLconfig{
				connectionString: "synersia:asia-northeast1:clone-synersia-228d",
				port:             20003,
			}
		case "renseikai":
			config = &cloudSQLconfig{
				connectionString: "production-renseikai:asia-northeast1:clone-renseikai-83fc",
				port:             20004,
			}
		case "ga", "aic":
			config = &cloudSQLconfig{
				connectionString: "student-coach-e1e95:asia-northeast1:clone-jp-partners-b04fbb69",
				port:             20005,
			}
		case "tokyo":
			switch instance {
			case "common":
				config = &cloudSQLconfig{
					connectionString: "student-coach-e1e95:asia-northeast1:clone-prod-tokyo",
					port:             20006,
				}
			case "lms":
				config = &cloudSQLconfig{
					connectionString: "student-coach-e1e95:asia-northeast1:clone-prod-tokyo-lms-b2dc4508",
					port:             20007,
				}
			case "dwh":
				config = &cloudSQLconfig{
					connectionString: "student-coach-e1e95:asia-northeast1:preprod-tokyo-data-warehouse",
					port:             20008,
				}
			}
		}
	case "prod":
		switch partner {
		case "jprep":
			config = &cloudSQLconfig{
				connectionString: "student-coach-e1e95:asia-northeast1:prod-jprep-d995522c",
				port:             30002,
			}
		case "synersia":
			config = &cloudSQLconfig{
				connectionString: "synersia:asia-northeast1:synersia-228d",
				port:             30003,
			}
		case "renseikai":
			config = &cloudSQLconfig{
				connectionString: "production-renseikai:asia-northeast1:renseikai-83fc",
				port:             30004,
			}
		case "ga", "aic":
			config = &cloudSQLconfig{
				connectionString: "student-coach-e1e95:asia-northeast1:jp-partners-b04fbb69",
				port:             30005,
			}
		case "tokyo":
			switch instance {
			case "common":
				config = &cloudSQLconfig{
					connectionString: "student-coach-e1e95:asia-northeast1:prod-tokyo",
					port:             30006,
				}
			case "lms":
				config = &cloudSQLconfig{
					connectionString: "student-coach-e1e95:asia-northeast1:prod-tokyo-lms-b2dc4508",
					port:             30007,
				}
			}
		}
	default:
		return nil, fmt.Errorf("invalid environment %q", environment)
	}

	if config == nil {
		return nil, fmt.Errorf("invalid environment.partner combination: \"%s.%s\"", environment, partner)
	}

	return config, nil
}
