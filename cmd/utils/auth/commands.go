package auth

import "github.com/spf13/cobra"

// RootCmd for init
var RootCmd = &cobra.Command{
	Use:   "auth [command]",
	Short: "auth define actions related to auth",
}

var UserCmd = &cobra.Command{
	Use:   "user [command]",
	Short: "user define actions related to user",
}

var ImportUsersFromFirebaseToIdentityPlatform = &cobra.Command{
	Use:   "import-firebase-to-identity []",
	Short: "import users from firebase project to identity platform's tenant",
	RunE: func(cmd *cobra.Command, args []string) error {
		return RunImportUsersFromFirebaseToIdentityPlatform(cmd.Context())
	},
}

var ImportUsersBetweenTenants = &cobra.Command{
	Use:   "import-identity-to-identity []",
	Short: "init conversation document from tom database",
	RunE: func(cmd *cobra.Command, args []string) error {
		return RunImportUsersBetweenTenants(cmd.Context())
	},
}

func init() {
	ImportUsersFromFirebaseToIdentityPlatform.PersistentFlags().StringVar(
		&srcFirebaseCredentialsFile, "src-firebase-credential", "", "firebase app credential file")
	ImportUsersFromFirebaseToIdentityPlatform.PersistentFlags().StringVar(
		&srcFirebaseProjectID, "src-firebase-project-id", "", "firebase app project id")
	ImportUsersFromFirebaseToIdentityPlatform.PersistentFlags().StringVar(
		&destIdentityPlatformCredentialFile, "dest-identity-credential", "", "dest gcloud app credential")
	ImportUsersFromFirebaseToIdentityPlatform.PersistentFlags().StringVar(
		&destIdentityPlatformProjectID, "dest-identity-project-id", "", "dest gcloud project id")
	ImportUsersFromFirebaseToIdentityPlatform.PersistentFlags().StringVar(
		&destIdentityPlatformTenantID, "dest-identity-tenant-id", "", "dest identity platform tenant id")
	ImportUsersFromFirebaseToIdentityPlatform.PersistentFlags().BoolVar(
		&exportReport, "export-report", true, "export csv or not")
	ImportUsersFromFirebaseToIdentityPlatform.PersistentFlags().BoolVar(
		&test, "test", false, "use for testing")

	ImportUsersBetweenTenants.PersistentFlags().StringVar(
		&srcIdentityPlatformCredentialFile, "src-credential", "", "src gcloud app credential")
	ImportUsersBetweenTenants.PersistentFlags().StringVar(
		&srcIdentityPlatformProjectID, "src-project-id", "", "src gcloud app project id")
	ImportUsersBetweenTenants.PersistentFlags().StringVar(
		&srcIdentityPlatformTenantID, "src-tenant-id", "", "src identity platform tenant id")
	ImportUsersBetweenTenants.PersistentFlags().StringVar(
		&destIdentityPlatformCredentialFile, "dest-credential", "", "dest gcloud app credential")
	ImportUsersBetweenTenants.PersistentFlags().StringVar(
		&destIdentityPlatformProjectID, "dest-project-id", "", "dest gcloud app project id")
	ImportUsersBetweenTenants.PersistentFlags().StringVar(
		&destIdentityPlatformTenantID, "dest-tenant-id", "", "dest identity platform tenant id")
	ImportUsersBetweenTenants.PersistentFlags().BoolVar(
		&exportReport, "export-report", true, "export csv or not")

	UserCmd.AddCommand(
		ImportUsersFromFirebaseToIdentityPlatform,
		ImportUsersBetweenTenants,
	)

	RootCmd.AddCommand(
		UserCmd,
	)
}
