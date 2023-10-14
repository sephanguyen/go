package multitenant

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFirebaseIssuerFromProjectID(t *testing.T) {
	actual := FirebaseIssuerFromProjectID("example")
	assert.Equal(t, "https://securetoken.google.com/example", actual)
}
