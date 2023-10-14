package validations

import (
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/stringutil"
	"github.com/manabie-com/backend/internal/notification/consts"

	"k8s.io/utils/strings/slices"
)

// nolint
func ValidateCSVHeaders(headers map[string]int) ([]string, error) {
	allowedHeaders := strings.Split(consts.AllowTagCSVHeaders, "|")
	csvHeaders := []string{}
	for header := range headers {
		if !slices.Contains(allowedHeaders, header) {
			return nil, fmt.Errorf("Header \"%s\" is not allowed. Only allow %s", header, strings.ReplaceAll(consts.AllowTagCSVHeaders, "|", ", "))
		}
		csvHeaders = append(csvHeaders, header)
	}
	missingHeaders := stringutil.SliceElementsDiff(allowedHeaders, csvHeaders)
	if len(missingHeaders) > 0 {
		return nil, fmt.Errorf("Missing headers \"%s\"", strings.Join(missingHeaders, ","))
	}
	return allowedHeaders, nil
}
