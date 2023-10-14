package staff

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/gcp"
)

const (
	happyCase   = "happy case"
	testCaseLog = "Test case: "
)

type TestCase struct {
	name        string
	ctx         context.Context
	req         interface{}
	expectedErr error
	setup       func(ctx context.Context)
	Options     interface{}
	expectedRes interface{}
}

func mockScryptHash() *gcp.HashConfig {
	return &gcp.HashConfig{
		HashAlgorithm:  "SCRYPT",
		HashRounds:     8,
		HashMemoryCost: 8,
		HashSaltSeparator: gcp.Base64EncodedStr{
			Value:        "salt",
			DecodedBytes: []byte("salt"),
		},
		HashSignerKey: gcp.Base64EncodedStr{
			Value:        "key",
			DecodedBytes: []byte("key"),
		},
	}
}
