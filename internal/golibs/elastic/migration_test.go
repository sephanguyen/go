package elastic

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpsertFieldDefinition(t *testing.T) {
	t.Parallel()
	type TestCase struct {
		name           string
		mockDefinition string
		mockResponse   string
		err            error
	}
	cases := []TestCase{
		{
			name:           "mapping already exist",
			mockDefinition: `dummy definition`,
			mockResponse:   `{"conversations":{"mappings":{"resource_path":{"full_name":"resource_path","mapping":{"resource_path":{"type":"keyword"}}}}}}`,
			err:            nil,
		},
		{
			name:           "creating mapping has error",
			mockDefinition: `dummy definition`,
			mockResponse:   `{"error":"dummy"}`,
			err:            fmt.Errorf("creating index for field %s with definition %s has error %v", "resource_path", "dummy definition", "dummy"),
		},
		{
			name:           "success",
			mockDefinition: `dummy definition`,
			mockResponse:   `{"acknowledged" : true}`,
			err:            nil,
		},
	}
	for _, tcase := range cases {
		t.Run(tcase.name, func(t *testing.T) {
			mockElas, close := NewMockSearchFactory(tcase.mockResponse)
			defer close()
			err := mockElas.UpsertFieldDefinition(context.Background(), "conversations", "resource_path", tcase.mockDefinition)
			assert.Equal(t, tcase.err, err)
		})
	}
}
