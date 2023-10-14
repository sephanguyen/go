package firebase

import (
	"context"
	"fmt"
	"log"
	"strings"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"github.com/spf13/cobra"
	"google.golang.org/api/option"

	"github.com/manabie-com/backend/internal/bob/constants"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	bobproto "github.com/manabie-com/backend/pkg/genproto/bob"
)

var (
	group           string
	credentialsFile string
	validGroup      = map[string]bobproto.UserGroup{
		"schoolAdmin": bobproto.USER_GROUP_SCHOOL_ADMIN,
		"admin":       bobproto.USER_GROUP_ADMIN,
		"teacher":     bobproto.USER_GROUP_TEACHER,
		"student":     bobproto.USER_GROUP_STUDENT,
	}
	schoolID string
	userID   string
)

func verifyCreateAccountArgs(cmd *cobra.Command, args []string) error {
	if len(credentialsFile) == 0 {
		return fmt.Errorf("missing credentials file")
	}

	if isTestingAccount() {
		return nil
	}

	if err := cobra.MinimumNArgs(1)(cmd, args); err != nil {
		return err
	}

	for _, e := range args {
		if !emailRe.Match([]byte(e)) {
			return fmt.Errorf("invalid email: %s", e)
		}
	}

	if _, ok := validGroup[group]; !ok {
		return fmt.Errorf("invalid group")
	}
	return nil
}

func isTestingAccount() bool {
	if strings.Contains(userID, "thu.vo+e2eschool") || userID == "thu.vo+e2eadmin@manabie.com" {
		return true
	}

	return false
}

func createAccount(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client := firebaseAuthClient(ctx)

	group = validGroup[group].String()

	if isTestingAccount() {
		if userID == "thu.vo+e2eadmin@manabie.com" {
			group = bobproto.USER_GROUP_ADMIN.String()
		} else {
			group = bobproto.USER_GROUP_SCHOOL_ADMIN.String()
		}
		args = []string{userID}
	}

	for _, email := range args {
		u, err := client.GetUserByEmail(ctx, email)
		if auth.IsUserNotFound(err) {
			id := idutil.ULIDNow()

			if isTestingAccount() {
				id = email
			}

			uc := (&auth.UserToCreate{}).
				UID(id).
				Email(email).
				Password("M@nabie123")

			_, err = client.CreateUser(ctx, uc)
			if err != nil {
				return fmt.Errorf("CreateUser error: %w", err)
			}

			err = client.SetCustomUserClaims(ctx, id, claims(id, group))
			if err != nil {
				return fmt.Errorf("SetCustomUserClaims error: %w", err)
			}

			fmt.Println(email, id, group)
			continue
		} else if err != nil {
			return fmt.Errorf("GetUserByEmail error: %w", err)
		}

		og := getOriginalGroups(u.CustomClaims)
		if !golibs.InArrayString(group, og) {
			var err error
			err = client.SetCustomUserClaims(ctx, u.UID, claimSchools(u.UID, group, schoolID, og...))
			if group != "schoolAdmin" && group != "USER_GROUP_SCHOOL_ADMIN" {
				err = client.SetCustomUserClaims(ctx, u.UID, claims(u.UID, group, og...))
			} else {
				err = client.SetCustomUserClaims(ctx, u.UID, claimSchools(u.UID, group, schoolID, og...))
			}

			if err != nil {
				return fmt.Errorf("SetCustomUserClaims error: %w", err)
			}

			fmt.Println(email, u.UID, group, schoolID)
		}
	}
	return nil
}
func claimSchools(uid, newGroup, schoolID string, originalGroups ...string) map[string]interface{} {
	if schoolID == "" {
		schoolID = fmt.Sprintf("%d", constants.ManabieSchool)
	}

	schoolIDs := fmt.Sprintf("`{%s}`", schoolID)

	return map[string]interface{}{
		"https://hasura.io/jwt/claims": map[string]interface{}{
			"x-hasura-allowed-roles": append(originalGroups, newGroup),
			"x-hasura-default-role":  newGroup,
			"x-hasura-user-id":       uid,
			"x-hasura-school-ids":    schoolIDs,
		},
	}
}

func claims(uid, newGroup string, originalGroups ...string) map[string]interface{} {
	return map[string]interface{}{
		"https://hasura.io/jwt/claims": map[string]interface{}{
			"x-hasura-allowed-roles": append(originalGroups, newGroup),
			"x-hasura-default-role":  newGroup,
			"x-hasura-user-id":       uid,
		},
	}
}

func getOriginalGroups(claims map[string]interface{}) []string {
	hc, ok := claims["https://hasura.io/jwt/claims"].(map[string]interface{})
	if !ok {
		return []string{}
	}

	groups, ok := hc["x-hasura-allowed-roles"].([]string)
	if !ok {
		return []string{}
	}

	return groups
}

func firebaseAuthClient(ctx context.Context) *auth.Client {
	app, err := firebase.NewApp(ctx, nil, option.WithCredentialsFile(credentialsFile))
	if err != nil {
		log.Println("error initializing app:", err)
	}

	client, err := app.Auth(ctx)
	if err != nil {
		log.Println("error getting Auth client:", err)
	}

	return client
}
