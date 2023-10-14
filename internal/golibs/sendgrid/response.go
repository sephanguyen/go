package sendgrid

import (
	"encoding/json"
	"fmt"

	"github.com/sendgrid/rest"
	"go.uber.org/multierr"
)

type SGError struct {
	Errors []struct {
		Message string `json:"message"`
		Field   string `json:"field"`
		Help    string `json:"help"`
	} `json:"errors"`
}

var (
	XMsgIDHTTPHeader = "X-Message-Id"
)

func HandleResponse(response *rest.Response) (string, error) {
	// process and return anything that is not 202 code
	if response.StatusCode != 202 {
		retErr := GetErrorMessagesFromResponse(response.Body)

		return "", fmt.Errorf("SendGrid API failed with errors: %s", retErr)
	}

	// get message id from success response
	msgID, err := GetMessageIDFromHeaders(response.Headers)
	if err != nil {
		return "", fmt.Errorf("failed GetMessageIDFromHeaders: %+v", err)
	}

	return msgID, nil
}

func GetErrorMessagesFromResponse(responseBody string) error {
	sgErr := &SGError{}
	msErr := json.Unmarshal([]byte(responseBody), sgErr)
	if msErr != nil {
		return fmt.Errorf("failed Unmarshal response body: %+v", msErr)
	}

	var retErr error
	for _, err := range sgErr.Errors {
		errString := fmt.Sprintf("Error %s", err.Message)
		if err.Field != "" {
			errString += fmt.Sprintf(" at position %s", err.Field)
		}
		retErr = multierr.Append(retErr, fmt.Errorf(errString))
	}

	return retErr
}

func GetMessageIDFromHeaders(header map[string][]string) (string, error) {
	xMsgID, ok := header[XMsgIDHTTPHeader]
	if !ok {
		return "", fmt.Errorf("missing %s in response header", XMsgIDHTTPHeader)
	}
	return xMsgID[0], nil
}
