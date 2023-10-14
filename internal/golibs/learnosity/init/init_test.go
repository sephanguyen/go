package init

import (
	"fmt"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/learnosity"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		Name           string
		Service        learnosity.Service
		Security       learnosity.Security
		Option         []learnosity.Option
		ExpectedOutput *Client
	}{
		{
			Name:    "happy case",
			Service: learnosity.ServiceItems,
			Security: learnosity.Security{
				ConsumerKey:    "consumer_key",
				Domain:         "domain",
				Timestamp:      "20230417-1234",
				UserID:         "user_id",
				ConsumerSecret: "consumer_secret",
			},
			Option: []learnosity.Option{
				learnosity.RequestString("request_string"),
			},
			ExpectedOutput: &Client{
				Service: learnosity.ServiceItems,
				Security: learnosity.Security{
					ConsumerKey:    "consumer_key",
					Domain:         "domain",
					Timestamp:      "20230417-1234",
					UserID:         "user_id",
					ConsumerSecret: "consumer_secret",
				},
				RequestString: "request_string",
				Request:       nil,
				Action:        learnosity.ActionNone,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			output := New(tc.Service, tc.Security, tc.Option...)
			assert.Equal(t, tc.ExpectedOutput, output)
		})
	}
}

func TestClient_Generate(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		Name           string
		Client         *Client
		ExpectedOutput any
		ExpectedError  error
	}{
		{
			Name: "happy case",
			Client: &Client{
				Service: learnosity.ServiceItems,
				Security: learnosity.Security{
					ConsumerKey:    "consumer_key",
					Domain:         "domain",
					Timestamp:      "20230417-1234",
					UserID:         "user_id",
					ConsumerSecret: "consumer_secret",
				},
				RequestString: "request_string",
				Request:       nil,
				Action:        learnosity.ActionNone,
			},
			ExpectedOutput: "{\"request\":\"request_string\",\"security\":{\"consumer_key\":\"consumer_key\",\"domain\":\"domain\",\"signature\":\"a81c0eff8b9049a849ec65cdc174f47b1d0d21f682ac807d45907cd87c98ad34\",\"timestamp\":\"20230417-1234\",\"user_id\":\"user_id\"}}",
			ExpectedError:  nil,
		},
		{
			Name: "validateGenerate: Security.ConsumerKey is empty",
			Client: &Client{
				Service: learnosity.ServiceItems,
				Security: learnosity.Security{
					ConsumerKey:    "",
					Domain:         "domain",
					Timestamp:      "20230417-1234",
					UserID:         "user_id",
					ConsumerSecret: "consumer_secret",
				},
				RequestString: "request_string",
				Request:       nil,
				Action:        learnosity.ActionNone,
			},
			ExpectedOutput: nil,
			ExpectedError:  fmt.Errorf("validateGenerate: validator.ValidationErrors: Key: 'Security.ConsumerKey'"),
		},
		{
			Name: "validateGenerate: Security.Domain is empty",
			Client: &Client{
				Service: learnosity.ServiceItems,
				Security: learnosity.Security{
					ConsumerKey:    "consumer_key",
					Domain:         "",
					Timestamp:      "20230417-1234",
					UserID:         "user_id",
					ConsumerSecret: "consumer_secret",
				},
				RequestString: "request_string",
				Request:       nil,
				Action:        learnosity.ActionNone,
			},
			ExpectedOutput: nil,
			ExpectedError:  fmt.Errorf("validateGenerate: validator.ValidationErrors: Key: 'Security.Domain'"),
		},
		{
			Name: "validateGenerate: Security.Timestamp is empty",
			Client: &Client{
				Service: learnosity.ServiceItems,
				Security: learnosity.Security{
					ConsumerKey:    "consumer_key",
					Domain:         "domain",
					Timestamp:      "",
					UserID:         "user_id",
					ConsumerSecret: "consumer_secret",
				},
				RequestString: "request_string",
				Request:       nil,
				Action:        learnosity.ActionNone,
			},
			ExpectedOutput: nil,
			ExpectedError:  fmt.Errorf("validateGenerate: validator.ValidationErrors: Key: 'Security.Timestamp'"),
		},
		{
			Name: "validateGenerate: Security.UserID is empty",
			Client: &Client{
				Service: learnosity.ServiceItems,
				Security: learnosity.Security{
					ConsumerKey:    "consumer_key",
					Domain:         "domain",
					Timestamp:      "20230417-1234",
					UserID:         "",
					ConsumerSecret: "consumer_secret",
				},
				RequestString: "request_string",
				Request:       nil,
				Action:        learnosity.ActionNone,
			},
			ExpectedOutput: nil,
			ExpectedError:  fmt.Errorf("validateGenerate: validator.ValidationErrors: Key: 'Security.UserID'"),
		},
		{
			Name: "validateGenerate: Security.ConsumerSecret is empty",
			Client: &Client{
				Service: learnosity.ServiceItems,
				Security: learnosity.Security{
					ConsumerKey:    "consumer_key",
					Domain:         "domain",
					Timestamp:      "20230417-1234",
					UserID:         "user_id",
					ConsumerSecret: "",
				},
				RequestString: "request_string",
				Request:       nil,
				Action:        learnosity.ActionNone,
			},
			ExpectedOutput: nil,
			ExpectedError:  fmt.Errorf("validateGenerate: validator.ValidationErrors: Key: 'Security.ConsumerSecret'"),
		},
		{
			Name: "happy case: data service",
			Client: &Client{
				Service: learnosity.ServiceData,
				Security: learnosity.Security{
					ConsumerKey:    "consumer_key",
					Domain:         "domain",
					Timestamp:      "20230417-1234",
					UserID:         "user_id",
					ConsumerSecret: "consumer_secret",
				},
				RequestString: "",
				Request: learnosity.Request{
					"limit": 5,
				},
				Action: "",
			},
			ExpectedOutput: map[string]any{
				"action":   "get",
				"request":  "{\"limit\":5}",
				"security": "{\"consumer_key\":\"consumer_key\",\"domain\":\"domain\",\"signature\":\"aac99967105e0d7c660e8d28df583811eda7b7554c009142a1065eb466e4f7e7\",\"timestamp\":\"20230417-1234\",\"user_id\":\"user_id\"}",
			},
			ExpectedError: nil,
		},
		{
			Name: "happy case: items service",
			Client: &Client{
				Service: learnosity.ServiceItems,
				Security: learnosity.Security{
					ConsumerKey:    "consumer_key",
					Domain:         "domain",
					Timestamp:      "20230417-1234",
					UserID:         "user_id",
					ConsumerSecret: "consumer_secret",
				},
				RequestString: "",
				Request: learnosity.Request{
					"limit": 5,
				},
				Action: "",
			},
			ExpectedOutput: map[string]any{
				"request": learnosity.Request{
					"limit": 5,
				},
				"security": map[string]string{
					"consumer_key": "consumer_key",
					"domain":       "domain",
					"signature":    "b50966ab47cad3708e35c285ce0670344ac9834d0e4e85ce9b99cd415a0a3075",
					"timestamp":    "20230417-1234",
					"user_id":      "user_id",
				},
			},
			ExpectedError: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			output, err := tc.Client.Generate(false)
			assert.Equal(t, tc.ExpectedOutput, output)
			if err != nil {
				assert.Contains(t, err.Error(), tc.ExpectedError.Error())
			}
		})
	}
}
