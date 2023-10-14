package configs

import (
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/multierr"
	"gopkg.in/yaml.v3"
)

type testConf struct {
	A     string    `yaml:"a,omitempty"`
	B     string    `yaml:"b,omitempty"`
	C     string    `yaml:"c,omitempty"`
	Child *testConf `yaml:"child,omitempty"`
}

const (
	ccp = "testdata/common.config.yaml"
	cp  = "testdata/config.yaml"
	sp  = "testdata/secret.yaml"
)

func TestLoadPlaintext(t *testing.T) {
	out := new(testConf)
	err := loadFile(ccp, out)
	require.NoError(t, err)
	assert.Equal(t, "1", out.A)
	assert.Equal(t, "1", out.Child.A)
	assert.Equal(t, "1", out.Child.Child.A)
}

func TestLoadEncrypted(t *testing.T) {
	out := new(testConf)
	err := DecryptFile(sp, out, mockDecrypt)
	require.NoError(t, err)
	assert.Equal(t, "3", out.C)
	assert.Equal(t, "3", out.Child.C)
	assert.Equal(t, "3", out.Child.Child.C)
}

func TestLoadAll(t *testing.T) {
	out, err := loadAll[testConf](ccp, cp, sp, mockDecrypt)
	require.NoError(t, err)
	assertAllFields(t, out)

	_, err = loadAll[testConf]("", "", "", mockDecrypt)
	assert.EqualError(t, err, `cannot load common configuration from "": open : no such file or directory`)

	_, err = loadAll[testConf](ccp, cp, "", mockDecrypt)
	assert.Nil(t, err)
}

// mockDecrypt implements decryptFunc. It is used in tests only.
// The decryption method is base64 encoding/decoding of certain fields in the secret.
func mockDecrypt(path, format string) ([]byte, error) {
	if format != "yaml" {
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
	out := new(testConf)
	if err := loadFile(path, out); err != nil {
		return nil, err
	}
	if out == nil || out.Child == nil || out.Child.Child == nil {
		return nil, fmt.Errorf("failed to custom decrypt: target is nil")
	}
	if err := multierr.Combine(
		base64DecodeField(&out.C),
		base64DecodeField(&out.Child.C),
		base64DecodeField(&out.Child.Child.C),
	); err != nil {
		return nil, err
	}
	return yaml.Marshal(out)
}

func base64DecodeField(in *string) error {
	if in == nil {
		return nil
	}
	out, err := base64.StdEncoding.DecodeString(*in)
	if err != nil {
		return err
	}
	*in = string(out)
	return nil
}

func assertAllFields(t *testing.T, v *testConf) {
	assert.Equal(t, "1", v.A)
	assert.Equal(t, "1", v.Child.A)
	assert.Equal(t, "1", v.Child.Child.A)
	assert.Equal(t, "2", v.B)
	assert.Equal(t, "2", v.Child.B)
	assert.Equal(t, "2", v.Child.Child.B)
	assert.Equal(t, "3", v.C)
	assert.Equal(t, "3", v.Child.C)
	assert.Equal(t, "3", v.Child.Child.C)
}
