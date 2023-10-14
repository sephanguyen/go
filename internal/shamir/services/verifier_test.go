package services

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/square/go-jose/v3/jwt"
	"github.com/stretchr/testify/assert"
)

type mockVerifier struct {
	verifyFn func(ctx context.Context, idToken string) (*interceptors.CustomClaims, error)
}

func (m *mockVerifier) Verify(ctx context.Context, idToken string) (*interceptors.CustomClaims, error) {
	return m.verifyFn(ctx, idToken)
}

func TestTokenVerifier_Verify(t *testing.T) {
	t.Parallel()
	t.Run("verify error", func(tt *testing.T) {
		tt.Parallel()
		testIDToken := "test token"
		m := &mockVerifier{
			verifyFn: func(ctx context.Context, idToken string) (*interceptors.CustomClaims, error) {
				if idToken != testIDToken {
					tt.Error("unexpected idToken", idToken)
				}
				return nil, fmt.Errorf("test error")
			},
		}
		v := &TokenVerifier{
			vendor:   "jprep",
			verifies: []tokenVerifier{m},
		}

		claims, err := v.Verify(context.Background(), testIDToken)
		assert.Error(tt, err)
		assert.Nil(tt, claims)
	})
}

func Test_generateNewToken(t *testing.T) {
	t.Parallel()
	c := generateNewToken(&interceptors.TokenInfo{
		Applicant:    "test-applicant",
		UserID:       "test-user-id",
		SchoolIds:    []int64{1, 2},
		DefaultRole:  "test-role-1",
		AllowedRoles: []string{"test-role-1", "test-role-2"},
		UserGroup:    "test-role-1",
		ResourcePath: "1",
	},
		&interceptors.CustomClaims{
			Claims: jwt.Claims{
				Subject: "test-subject",
			},
		})

	assert.Equal(t, c.Subject, "test-subject")
	assert.Equal(t, c.Manabie.UserID, "test-user-id")
	assert.Equal(t, c.Manabie.SchoolIDs, golibs.ToArrayStringFromArrayInt64([]int64{1, 2}))
	assert.Equal(t, c.Manabie.DefaultRole, "test-role-1")
	assert.Equal(t, c.Manabie.AllowedRoles, []string{"test-role-1", "test-role-2"})

	assert.Equal(t, c.Manabie.UserID, c.Hasura.UserID)
	assert.Equal(t, fmt.Sprintf("{%s}", strings.Join(c.Manabie.SchoolIDs, ",")), c.Hasura.SchoolIDs)
	assert.Equal(t, c.Manabie.DefaultRole, c.Hasura.DefaultRole)
	assert.Equal(t, c.Manabie.AllowedRoles, c.Hasura.AllowedRoles)
	assert.Equal(t, c.Manabie.UserGroup, "test-role-1")
	assert.Equal(t, c.Manabie.ResourcePath, "1")

	assert.Equal(t, c.Hasura.UserGroup, "test-role-1")
	assert.Equal(t, c.Hasura.ResourcePath, "1")

	assert.Equal(t, c.UserGroup, "test-role-1,test-role-2")
	assert.Equal(t, c.ResourcePath, "1")
}
