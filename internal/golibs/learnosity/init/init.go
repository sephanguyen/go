package init

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/learnosity"

	"github.com/go-playground/validator/v10"
)

// Client represents the creation and signing of init options for all supported APIs.
type Client struct {
	// Service represents the name of the API to sign initialization options for.
	Service learnosity.Service

	// Security represents the public and private security keys required to access Learnosity APIs and data.
	Security learnosity.Security

	// RequestString represents the JSON stringify format which is higher priority than Request.
	RequestString learnosity.RequestString

	// Request represents the correct data format to integrate with any of the Learnosity API services.
	Request learnosity.Request

	// Action represents the action type of your request (get, set, update, etc.).
	Action learnosity.Action
}

var _ learnosity.Init = (*Client)(nil)

// New returns an Init Client from the mandatory and optional parameters.
func New(service learnosity.Service, security learnosity.Security, opts ...learnosity.Option) *Client {
	// Options struct with default values and applies any given options.
	options := learnosity.Options{
		RequestString: "",
		Request:       nil,
		Action:        learnosity.ActionNone,
	}
	for _, option := range opts {
		option.Apply(&options)
	}

	return &Client{service, security, options.RequestString, options.Request, options.Action}
}

// Generate used to generate the data necessary to make a request to one of the Learnosity services.
// If encode is True, the result is a JSON string. Otherwise, it's a map[string]any.
// If the service is Data, encode is ignored.
func (c Client) Generate(encode bool) (any, error) {
	if err := c.validateGenerate(); err != nil {
		return nil, fmt.Errorf("validateGenerate: %w", err)
	}

	output := make(map[string]any)

	switch c.Service {
	case learnosity.ServiceAuthor, learnosity.ServiceItems, learnosity.ServiceReports:
		if err := c.handleBasicServices(output); err != nil {
			return nil, fmt.Errorf("HandleBasicServices: %w", err)
		}
	case learnosity.ServiceData:
		if err := c.handleDataService(output); err != nil {
			return nil, fmt.Errorf("HandleDataService: %w", err)
		}
		return output, nil
	}

	// encode or request passed as string
	if encode || c.RequestString != "" {
		outputStr, err := learnosity.JSONMarshalToString(output)
		if err != nil {
			return nil, fmt.Errorf(learnosity.ErrJSONMarshalToString.Error(), err)
		}
		return outputStr, nil
	}
	return output, nil
}

func (c Client) validateGenerate() error {
	validate := validator.New()

	err := validate.Struct(c.Security)
	if _, ok := err.(validator.ValidationErrors); ok {
		return fmt.Errorf("validator.ValidationErrors: %w", err)
	}

	return nil
}

// handleBasicServices handles logic for the basic services (author, items, reports).
func (c Client) handleBasicServices(output map[string]any) error {
	securityMap, err := c.getSecurityMap()
	if err != nil {
		return fmt.Errorf("GetSecurityMap: %w", err)
	}
	output["security"] = securityMap

	if c.RequestString != "" {
		output["request"] = string(c.RequestString)
	} else if len(c.Request) != 0 {
		output["request"] = c.Request
	}

	return nil
}

// handleDataService handles logic for the data service.
func (c Client) handleDataService(output map[string]any) error {
	// Data service works with Request is more convenient.
	if c.RequestString != "" {
		return fmt.Errorf("RequestString is not supported for Data service. Please use Request instead")
	}

	// If action is not set, default to get.
	if c.Action == learnosity.ActionNone {
		c.Action = learnosity.ActionGet
	}

	securityMap, err := c.getSecurityMap()
	if err != nil {
		return fmt.Errorf("GetSecurityMap: %w", err)
	}

	securityStr, err := learnosity.JSONMarshalToString(securityMap)
	if err != nil {
		return fmt.Errorf(learnosity.ErrJSONMarshalToString.Error(), err)
	}
	output["security"] = securityStr

	if len(c.Request) != 0 {
		requestStr, err := learnosity.JSONMarshalToString(c.Request)
		if err != nil {
			return fmt.Errorf(learnosity.ErrJSONMarshalToString.Error(), err)
		}
		output["request"] = requestStr
	}

	output["action"] = string(c.Action)

	return nil
}

// getSecurityMap returns a map with the keys in order.
func (c Client) getSecurityMap() (map[string]string, error) {
	securityMap := make(map[string]string)

	securityMap["consumer_key"] = c.Security.ConsumerKey
	securityMap["domain"] = c.Security.Domain
	securityMap["timestamp"] = c.Security.Timestamp
	securityMap["user_id"] = c.Security.UserID

	signature, err := c.generateSignature()
	if err != nil {
		return nil, fmt.Errorf("generateSignature: %w", err)
	}
	securityMap["signature"] = signature

	return securityMap, nil
}

// generateSignature returns a signature hash string for the request authentication.
// With the concatenation of the Security parameters in order, separated by underscores.
func (c Client) generateSignature() (string, error) {
	signatureArr := make([]string, 0)

	signatureArr = append(signatureArr,
		c.Security.ConsumerKey,
		c.Security.Domain,
		c.Security.Timestamp,
		c.Security.UserID,
		c.Security.ConsumerSecret,
	)

	// RequestString is higher priority than Request.
	if c.RequestString != "" {
		signatureArr = append(signatureArr, string(c.RequestString))
	} else if len(c.Request) != 0 {
		requestStr, err := learnosity.JSONMarshalToString(c.Request)
		if err != nil {
			return "", fmt.Errorf(learnosity.ErrJSONMarshalToString.Error(), err)
		}
		signatureArr = append(signatureArr, requestStr)
	}

	if c.Action != learnosity.ActionNone {
		signatureArr = append(signatureArr, string(c.Action))
	}

	hash := sha256.New()
	_, err := hash.Write([]byte(strings.Join(signatureArr, "_")))
	if err != nil {
		return "", fmt.Errorf("hash.Write: %w", err)
	}
	hashStr := hex.EncodeToString(hash.Sum(nil))

	return hashStr, nil
}
