package firebase

import (
	"testing"

	"firebase.google.com/go/v4/auth"
	"github.com/stretchr/testify/assert"
)

func TestNewAuthFromApp(t *testing.T) {
	var authClient auth.Client
	resp := NewAuthFromApp(&authClient)
	assert.NotEmpty(t, resp)
}
