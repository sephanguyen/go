// Package manabie contains a Firestore Cloud Function.
package manabie

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	firebase "firebase.google.com/go"
)

// AuthEvent is the payload of a Firestore Auth event.
type AuthEvent struct {
	Email    string `json:"email"`
	Metadata struct {
		CreatedAt time.Time `json:"createdAt"`
	} `json:"metadata"`
	UID      string `json:"uid"`
	TenantID string `json:"tenantId"`
}

func claims(uid, tenantID string) map[string]interface{} {
	return map[string]interface{}{
		"hasura": map[string]interface{}{
			"x-hasura-allowed-roles": []string{"user"},
			"x-hasura-default-role":  "user",
			"x-hasura-user-id":       uid,
			"x-hasura-tenant-id":     tenantID,
		},
	}
}

// ClaimTokenOnCreate is triggered by Firestore Auth events.
func ClaimTokenOnCreate(ctx context.Context, data AuthEvent) error {
	log.Println("Function triggered", data)
	tenantID := data.TenantID
	uid := data.UID

	log.Printf("UID: %q", uid)
	log.Printf("TenantID: %q", tenantID)

	// TODO: Change this to your own project ID
	srcFirebaseProjectID := os.Getenv("GOOGLE_PROJECT_ID")
	srcFirebaseTenantID := os.Getenv("GOOGLE_TENANT_ID")

	log.Printf("srcFirebaseProjectID %s", srcFirebaseProjectID)
	log.Printf("srcFirebaseTenantID %s", srcFirebaseTenantID)
	if tenantID != srcFirebaseTenantID {
		return nil
	}

	firebaseConfig := &firebase.Config{
		ProjectID: srcFirebaseProjectID,
	}

	firebaseApp, err := firebase.NewApp(ctx, firebaseConfig)
	if err != nil {
		return fmt.Errorf("firebase.NewApp error: %w", err)
	}
	srcAuth, err := firebaseApp.Auth(ctx)
	if err != nil {
		return fmt.Errorf("firebaseApp.Auth error: %w", err)
	}
	tenantClient, err := srcAuth.TenantManager.AuthForTenant(tenantID)
	if err != nil {
		return fmt.Errorf("auth.NewFirebaseAuthClientFromGCP error: %w", err)
	}

	err = tenantClient.SetCustomUserClaims(ctx, uid, claims(uid, tenantID))
	if err != nil {
		return fmt.Errorf("tenantClient.SetCustomUserClaims error: %w", err)
	}

	return nil
}
