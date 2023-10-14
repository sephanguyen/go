package firebase

import (
	"os"
	"regexp"

	"github.com/spf13/cobra"
)

var emailRe = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

// CreateAccountCmd creates firebase account
var CreateAccountCmd = &cobra.Command{
	Use:   "createAccount [email, email, email, ...]",
	Short: "creates firebase account",
	Args:  verifyCreateAccountArgs,
	RunE:  createAccount,
}

// GenTokenCmd gen JWT firebase token
var GenTokenCmd = &cobra.Command{
	Use:   "genToken [id]",
	Short: "generate JWT token",
	Args:  verifyGenTokenArgs,
	RunE:  genToken,
}

// RootCmd for firebase
var RootCmd = &cobra.Command{
	Use:   "firebase [command]",
	Short: "create or set custom claims if account already exist",
}

func init() {
	RootCmd.AddCommand(
		CreateAccountCmd,
		GenTokenCmd,
	)

	RootCmd.PersistentFlags().StringVar(
		&credentialsFile,
		"credentials",
		os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"),
		"firebase creds file, default from GOOGLE_APPLICATION_CREDENTIALS env",
	)

	RootCmd.PersistentFlags().StringVar(
		&group,
		"group",
		"",
		"groups: admin, schoolAdmin, teacher or student",
	)
	RootCmd.MarkFlagRequired("group")

	RootCmd.PersistentFlags().StringVar(
		&schoolID,
		"school",
		"",
		"Manabie School: -2147483648",
	)
	RootCmd.MarkFlagRequired("school")

	RootCmd.PersistentFlags().StringVar(
		&userID,
		"userID",
		"",
		"thu.vo+e2eadmin@manabie.com or thu.vo+e2eschool@manabie.com",
	)
}
