package http

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"

	"github.com/stretchr/testify/assert"
)

func TestParseJSONPayload(t *testing.T) {
	t.Parallel()

	output := struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}{}

	customizedTypeOutput := struct {
		Int16  field.Int16  `json:"int16"`
		Int32  field.Int32  `json:"int32"`
		Int64  field.Int64  `json:"int64"`
		String field.String `json:"string"`
		Time   field.Time   `json:"time"`
		Date   field.Date   `json:"date"`
	}{}

	testCases := []struct {
		name           string
		input          interface{}
		expectedOutput error
		outputType     interface{}
	}{
		{
			name:           "happy case",
			input:          `{"name": "testcase", "age": 1}`,
			expectedOutput: nil,
			outputType:     output,
		},
		{
			name:           "bad case",
			input:          `{"name": "testcase", "age": "1"}`,
			expectedOutput: fmt.Errorf(` 'age' expected type 'int', got unconvertible type 'string', value: '1'`),
			outputType:     output,
		},
		{
			name: "valid input with customized type",
			input: `{
				"int16": 123,
				"int32": 456789,
				"int64": 922337203685477580,
				"time": "2006/01/02",
				"date": "2006/01/02"
			}`,
			expectedOutput: nil,
			outputType:     customizedTypeOutput,
		},
		{
			name: "invalid input with int16 type",
			input: `{
				"int16": ""
			}`,
			expectedOutput: fmt.Errorf("* error decoding 'int16': field Int16 invalid: json: cannot unmarshal string into Go value of type int16, invalid value: \"\""),
			outputType:     customizedTypeOutput,
		},
		{
			name: "invalid input with int32 type",
			input: `{
				"int32": ""
			}`,
			expectedOutput: fmt.Errorf("* error decoding 'int32': field Int32 invalid: json: cannot unmarshal string into Go value of type int32, invalid value: \"\""),
			outputType:     customizedTypeOutput,
		},
		{
			name: "invalid input with int64 type",
			input: `{
				"int64": ""
			}`,
			expectedOutput: fmt.Errorf("* error decoding 'int64': field Int64 invalid: json: cannot unmarshal string into Go value of type int64, invalid value: \"\""),
			outputType:     customizedTypeOutput,
		},
		{
			name: "invalid input with string type",
			input: `{
				"string": 0
			}`,
			expectedOutput: fmt.Errorf(`error decoding 'string': field String invalid: json: cannot unmarshal number into Go value of type string, invalid value: 0`),
			outputType:     customizedTypeOutput,
		},
		{
			name: "invalid input with time type",
			input: `{
				"time": "20/20/20"
			}`,
			expectedOutput: fmt.Errorf(`error decoding 'time': field Time invalid: parsing time "20/20/20" as "2006/01/02": cannot parse "20/20/20" as "2006", invalid value: 20/20/20`),
			outputType:     customizedTypeOutput,
		},
		{
			name: "valid input with date type",
			input: `{
				"date": "20/20/20"
			}`,
			expectedOutput: fmt.Errorf(`error decoding 'date': field Date invalid: parsing time "20/20/20" as "2006/01/02": cannot parse "20/20/20" as "2006", invalid value: 20/20/20`),
			outputType:     customizedTypeOutput,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Log(testCase.name)
			mockReq, _ := http.NewRequest(http.MethodPost, "", bytes.NewReader([]byte(testCase.input.(string))))

			err := ParseJSONPayload(mockReq, testCase.outputType)
			if err != nil {
				assert.True(t, strings.Contains(err.Error(), testCase.expectedOutput.Error()))
			} else {
				assert.Equal(t, testCase.expectedOutput, nil)
			}
		})
	}
}
