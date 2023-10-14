package entities

import (
	"testing"

	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/stretchr/testify/assert"
)

func TestUser_GetName(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		name      string
		givenname string
		lastname  string
		expected  string
	}{
		{"success", "ABCD", "WXYZ", "ABCD WXYZ"},
		{"empty given name", "", "Abcde", "Abcde"},
		{"empty both", "", "", ""},
	}
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			u := &User{GivenName: database.Text(tc.givenname), LastName: database.Text(tc.lastname)}
			actual := u.GetName()
			assert.Exactly(t, tc.expected, actual)
		})
	}
}
