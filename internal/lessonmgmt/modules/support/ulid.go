package support

import (
	"bytes"
	"crypto/sha1" // #nosec
	"time"

	"github.com/oklog/ulid/v2"
)

func GenerateULIDFromString(s string) ulid.ULID {
	// #nosec
	hasher := sha1.New()
	hasher.Write([]byte(s))

	t := time.Unix(1000000, 0)

	entropy := bytes.NewReader(hasher.Sum(nil))
	ulidID := ulid.MustNew(ulid.Timestamp(t), entropy)

	return ulidID
}
