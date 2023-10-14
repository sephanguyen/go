package sendgrid

import (
	"errors"
	"fmt"
	"testing"

	"github.com/sendgrid/rest"
	"github.com/stretchr/testify/assert"
	"go.uber.org/multierr"
)

func Test_GetErrorMessagesFromResponse(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		Name      string
		Response  *rest.Response
		ExpectErr error
	}{
		{
			Name: "happy case",
			Response: &rest.Response{
				StatusCode: 400,
				Body:       `{"errors":[{"message":"Substitutions may not be used with dynamic templating","field":"personalizations.0.substitutions","help":"http://sendgrid.com/docs/API_Reference/Web_API_v3/Mail/errors.html#message.personalizations.substitutions"},{"message":"Substitutions may not be used with dynamic templating","field":"personalizations.1.substitutions","help":"http://sendgrid.com/docs/API_Reference/Web_API_v3/Mail/errors.html#message.personalizations.substitutions"}]}`,
				Headers:    map[string][]string{},
			},
			ExpectErr: multierr.Combine(errors.New(`Error Substitutions may not be used with dynamic templating at position personalizations.0.substitutions`), errors.New(`Error Substitutions may not be used with dynamic templating at position personalizations.1.substitutions`)),
		},
		{
			Name: "case unauthorized err",
			Response: &rest.Response{
				StatusCode: 401,
				Body:       `{"errors":[{"message":"Permission denied, wrong credentials","field":null,"help":null}]}`,
				Headers:    map[string][]string{},
			},
			ExpectErr: multierr.Combine(errors.New(`Error Permission denied, wrong credentials`)),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			err := GetErrorMessagesFromResponse(tc.Response.Body)
			assert.Equal(t, tc.ExpectErr, err)
		})
	}
}

func Test_GetMessageIDFromHeaders(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		Name      string
		Response  *rest.Response
		ExpectErr error
		ExpectID  string
	}{
		{
			Name: "happy case",
			Response: &rest.Response{
				StatusCode: 202,
				Body:       "",
				Headers: map[string][]string{
					"X-Message-Id": {
						"xmsgid-1",
					},
				},
			},
			ExpectErr: nil,
			ExpectID:  "xmsgid-1",
		},
		{
			Name: "not found in header",
			Response: &rest.Response{
				StatusCode: 202,
				Body:       "",
				Headers: map[string][]string{
					"Access-Control-Allow-Headers": {
						"Authorization, Content-Type, On-behalf-of, x-sg-elas-acl",
					},
					"Access-Control-Allow-Methods": {
						"POST",
					},
					"Access-Control-Allow-Origin": {
						"https://sendgrid.api-docs.io",
					},
					"Access-Control-Max-Age": {
						"600",
					},
				},
			},
			ExpectErr: fmt.Errorf("missing %s in response header", XMsgIDHTTPHeader),
			ExpectID:  "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			msgID, err := GetMessageIDFromHeaders(tc.Response.Headers)
			if tc.ExpectErr != nil {
				assert.Equal(t, tc.ExpectErr, err)
				assert.Equal(t, "", msgID)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tc.ExpectID, msgID)
			}
		})
	}
}
