package validations

import (
	"fmt"
	"strings"
	"testing"

	"github.com/manabie-com/backend/internal/notification/consts"

	"github.com/stretchr/testify/assert"
)

func Test_ValidateCSVHeaders(t *testing.T) {
	t.Parallel()
	allowedHeaders := strings.Split(consts.AllowTagCSVHeaders, "|")
	testCases := []struct {
		Name       string
		CsvHeaders map[string]int
		Err        error
	}{
		{
			Name: "correct headers",
			CsvHeaders: func() map[string]int {
				headers := make(map[string]int)
				for idx, header := range allowedHeaders {
					headers[header] = idx
				}

				return headers
			}(),
			Err: nil,
		},
		{
			Name: "invalid headers",
			CsvHeaders: func() map[string]int {
				headers := make(map[string]int)
				for idx, header := range allowedHeaders {
					headers[header] = idx
				}
				headers["tag_order"] = 5
				return headers
			}(),
			Err: fmt.Errorf("Header \"tag_order\" is not allowed. Only allow %s", strings.ReplaceAll(consts.AllowTagCSVHeaders, "|", ", ")),
		},
		{
			Name: "missing headers",
			CsvHeaders: func() map[string]int {
				headers := make(map[string]int)
				for idx, header := range allowedHeaders {
					if idx == 0 {
						continue
					}
					headers[header] = idx
				}
				return headers
			}(),
			Err: fmt.Errorf("Missing headers \"%s\"", allowedHeaders[0]),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			headers, err := ValidateCSVHeaders(tc.CsvHeaders)
			if tc.Err == nil {
				assert.Nil(t, err)
				assert.Equal(t, allowedHeaders, headers)
			} else {
				assert.Equal(t, tc.Err, err)
			}
		})
	}
}
