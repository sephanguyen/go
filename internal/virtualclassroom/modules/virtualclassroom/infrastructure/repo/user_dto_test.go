package repo

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

func TestUser_GetUID(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		name     string
		ID       string
		expected string
	}{
		{"success", "test-id", "test-id"},
		{"empty", "", ""},
	}
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			u := &User{ID: database.Text(tc.ID)}
			actual := u.GetUID()
			assert.Exactly(t, tc.expected, actual)
		})
	}
}

func TestUser_GetEmail(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		name     string
		email    string
		expected string
	}{
		{"success", "test-email@test.com", "test-email@test.com"},
		{"upper case", "TEST-EMAIL@TEST.COM", "test-email@test.com"},
	}
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			u := &User{Email: database.Text(tc.email)}
			actual := u.GetEmail()
			assert.Exactly(t, tc.expected, actual)
		})
	}
}
