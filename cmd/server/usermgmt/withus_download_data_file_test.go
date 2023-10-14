package usermgmt

import (
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/cmd/server/usermgmt/withus"
	"github.com/stretchr/testify/assert"
)

func TestWDataFileNameSuffix(t *testing.T) {
	uploadedDate := time.Now()

	year, month, day := uploadedDate.Date()

	expected := fmt.Sprintf("%d", year) + fmt.Sprintf("%02d", month) + fmt.Sprintf("%02d", day)

	assert.Equal(t, expected, withus.DataFileNameSuffix(uploadedDate))
}

func TestUTCToJST(t *testing.T) {
	triggeredTimeInUTC := time.Date(2023, 01, 01, 19, 0, 0, 0, time.UTC)

	tokyoLocation, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "20230102", withus.DataFileNameSuffix(triggeredTimeInUTC.In(tokyoLocation)))
}
