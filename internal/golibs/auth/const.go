package auth

import "fmt"

const (
	// FirebaseAndIdentityJwkURL is shared jwk for firebase/identity regardless of environment
	FirebaseAndIdentityJwkURL = "https://www.googleapis.com/service_accounts/v1/jwk/securetoken@system.gserviceaccount.com"
)

func FirebaseIssuerFromProjectID(projectID string) string {
	return fmt.Sprintf("https://securetoken.google.com/%s", projectID)
}
