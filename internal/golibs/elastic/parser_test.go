package elastic

import (
	"io"
	"strings"
	"testing"

	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/stretchr/testify/assert"
)

func TestParseUpdateResponse(t *testing.T) {
	errStr := `{"error":{"root_cause":[{"type":"document_missing_exception","reason":"[_doc][not exist]: document missing","index_uuid":"jvutXkG7STu8eHqpcJI3KA","shard":"0","index":"conversations"}],"type":"document_missing_exception","reason":"[_doc][not exist]: document missing","index_uuid":"jvutXkG7STu8eHqpcJI3KA","shard":"0","index":"conversations"},"status":404}`
	err := CheckResponse(&esapi.Response{
		Body:       io.NopCloser(strings.NewReader(errStr)),
		StatusCode: 404,
	})
	assert.Error(t, err)
}
