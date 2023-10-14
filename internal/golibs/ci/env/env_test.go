package env

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnvVar_ToEnv(t *testing.T) {
	vals := Vars(map[string]string{
		"testenvvar_a": "1",
	})
	list := vals.ToEnv()
	assert.Contains(t, list, "testenvvar_a=1")
	assert.NotContains(t, list, "testenvvar_b=2")
}
