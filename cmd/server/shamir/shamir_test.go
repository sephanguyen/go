package shamir

import (
	"testing"

	"github.com/manabie-com/backend/internal/golibs/auth"
	"github.com/manabie-com/backend/internal/golibs/configs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"

	"github.com/stretchr/testify/assert"
)

func TestAppendAdditionalIssuer(t *testing.T) {
	currentIssuer := "https://securetoken.google.com/test-audience"
	currentAudience := "test-audience"
	currentJWK := "https://test-jwk.example.com"

	currentIssuers := []configs.TokenIssuerConfig{
		{
			Issuer:       currentIssuer,
			Audience:     currentAudience,
			JWKSEndpoint: currentJWK,
		},
	}

	additionalProjectID1 := database.Text("additional-project-id-1")
	additionalProjectID2 := database.Text("additional-project-id")

	organizationAuths := []*entity.OrganizationAuth{
		{
			OrganizationID: database.Int4(1),
			AuthProjectID:  additionalProjectID1,
			AuthTenantID:   database.Text("additional-tenant-id"),
		},
		{
			OrganizationID: database.Int4(2),
			AuthProjectID:  additionalProjectID2,
			AuthTenantID:   database.Text(""),
		},
		{
			OrganizationID: database.Int4(3),
			AuthProjectID:  database.Text(currentAudience), //duplicated audience
			AuthTenantID:   database.Text(""),
		},
	}

	finalIssuers := appendAdditionalIssuer(currentIssuers, organizationAuths)

	expectedFinalIssuers := []configs.TokenIssuerConfig{
		{
			Issuer:       currentIssuer,
			Audience:     currentAudience,
			JWKSEndpoint: currentJWK,
		},
		{
			Issuer:       auth.FirebaseIssuerFromProjectID(additionalProjectID1.String),
			Audience:     additionalProjectID1.String,
			JWKSEndpoint: auth.FirebaseAndIdentityJwkURL,
		},
		{
			Issuer:       auth.FirebaseIssuerFromProjectID(additionalProjectID2.String),
			Audience:     additionalProjectID2.String,
			JWKSEndpoint: auth.FirebaseAndIdentityJwkURL,
		},
	}

	assert.Equal(t, finalIssuers, expectedFinalIssuers)
}
