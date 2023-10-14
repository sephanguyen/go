package interceptors

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCustomClaims_DefaultRole(t *testing.T) {
	t.Parallel()
	t.Run("expecting manabie default role over hasura", func(tt *testing.T) {
		tt.Parallel()
		c := &CustomClaims{
			Hasura: &HasuraClaims{
				DefaultRole: "hasura",
			},
			Manabie: &ManabieClaims{
				DefaultRole: "manabie",
			},
		}

		assert.Equal(tt, "manabie", c.DefaultRole())
	})

	t.Run("only when manabie is empty", func(tt *testing.T) {
		tt.Parallel()
		c := &CustomClaims{
			Hasura: &HasuraClaims{
				DefaultRole: "hasura",
			},
		}

		assert.Equal(tt, "hasura", c.DefaultRole())
	})

	t.Run("empty when both is empty", func(tt *testing.T) {
		tt.Parallel()
		c := &CustomClaims{}

		assert.Equal(tt, "", c.DefaultRole())
	})
}

func TestCustomClaims_SchoolIds(t *testing.T) {
	t.Parallel()
	t.Run("expecting manabie have school ids", func(tt *testing.T) {
		tt.Parallel()
		c := &CustomClaims{
			Hasura: &HasuraClaims{
				DefaultRole: "hasura",
				SchoolIDs:   "{}",
			},
			Manabie: &ManabieClaims{
				DefaultRole: "manabie",
				SchoolIDs:   []string{"{school-id}"},
			},
		}

		assert.Equal(tt, "{}", c.Hasura.SchoolIDs)
		assert.Equal(tt, []string{"{school-id}"}, c.Manabie.SchoolIDs)
	})
}

func TestMarshalJSON(t *testing.T) {
	t.Parallel()
	jsonStr := `{
		"iss": "manabie",
		"sub": "af1d4f73-2cc1-402e-92fe-2412e4ca4848",
		"aud": "803wsd1dyl3x5jz22t",
		"exp": 1609297904,
		"iat": 1609297599,
		"jti": "01ETRSS7CR9YARJK42DSWHJ587",
		"student_division": "kids",
		"email": "test@email.com",
		"email_verified": true,
		"firebase": {
		  "sign_in_provider": "email",
		  "identities": {
		    "email": ["test@email.com"]
		  }
		},
		"https://hasura.io/jwt/claims": {
		  "x-hasura-allowed-roles": [
			"USER_GROUP_STUDENT"
		  ],
		  "x-hasura-default-role": "USER_GROUP_STUDENT",
		  "x-hasura-user-id": "af1d4f73-2cc1-402e-92fe-2412e4ca4848"
		},
		"manabie": {
		  "allowed_roles": [
			"USER_GROUP_STUDENT"
		  ],
		  "default_role": "USER_GROUP_STUDENT",
		  "user_id": "af1d4f73-2cc1-402e-92fe-2412e4ca4848"
		}
	  }`

	c := &CustomClaims{}
	assert.NoError(t, json.Unmarshal([]byte(jsonStr), c), "expecting no error for simple Unmarshal")
	assert.Equal(t, "manabie", c.Issuer)
	assert.Equal(t, "af1d4f73-2cc1-402e-92fe-2412e4ca4848", c.Subject)
	assert.Equal(t, "803wsd1dyl3x5jz22t", c.Audience[0])
	assert.Equal(t, "kids", c.StudentDivision)
	assert.Equal(t, "test@email.com", c.FirebaseClaims.Email)
	assert.Equal(t, true, c.FirebaseClaims.EmailVerified)
	assert.Equal(t, "email", c.FirebaseClaims.Identity.SignInProvider)
	assert.Equal(t, "", c.FirebaseClaims.Identity.Tenant)

	b, err := json.Marshal(c)
	assert.NoError(t, err)
	assert.JSONEq(t, jsonStr, string(b))
}
