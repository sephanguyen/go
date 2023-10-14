package firebase

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

func verifyGenTokenArgs(cmd *cobra.Command, args []string) error {
	if err := cobra.MinimumNArgs(1)(cmd, args); err != nil {
		return err
	}

	if len(credentialsFile) == 0 {
		return fmt.Errorf("missing credentials file")
	}

	if _, ok := validGroup[group]; !ok {
		return fmt.Errorf("invalid group")
	}

	return nil
}

func genToken(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client := firebaseAuthClient(ctx)

	token, err := client.CustomTokenWithClaims(ctx, args[0], claims(args[0], validGroup[group].String()))
	if err != nil {
		return fmt.Errorf("CustomTokenWithClaims err: %w", err)
	}

	fmt.Println(token)
	return nil
}
