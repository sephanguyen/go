// Code generated by protoc-gen-validate. DO NOT EDIT.
// source: syllabus/v1/assessment_service.proto

package sspb

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"net/mail"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"google.golang.org/protobuf/types/known/anypb"
)

// ensure the imports are used
var (
	_ = bytes.MinRead
	_ = errors.New("")
	_ = fmt.Print
	_ = utf8.UTFMax
	_ = (*regexp.Regexp)(nil)
	_ = (*strings.Reader)(nil)
	_ = net.IPv4len
	_ = time.Duration(0)
	_ = (*url.URL)(nil)
	_ = (*mail.Address)(nil)
	_ = anypb.Any{}
	_ = sort.Sort
)

// Validate checks the field values on GetSignedRequestRequest with the rules
// defined in the proto definition for this message. If any rules are
// violated, the first error encountered is returned, or nil if there are no violations.
func (m *GetSignedRequestRequest) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on GetSignedRequestRequest with the
// rules defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// GetSignedRequestRequestMultiError, or nil if none found.
func (m *GetSignedRequestRequest) ValidateAll() error {
	return m.validate(true)
}

func (m *GetSignedRequestRequest) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	// no validation rules for RequestData

	// no validation rules for Domain

	if len(errors) > 0 {
		return GetSignedRequestRequestMultiError(errors)
	}

	return nil
}

// GetSignedRequestRequestMultiError is an error wrapping multiple validation
// errors returned by GetSignedRequestRequest.ValidateAll() if the designated
// constraints aren't met.
type GetSignedRequestRequestMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m GetSignedRequestRequestMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m GetSignedRequestRequestMultiError) AllErrors() []error { return m }

// GetSignedRequestRequestValidationError is the validation error returned by
// GetSignedRequestRequest.Validate if the designated constraints aren't met.
type GetSignedRequestRequestValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e GetSignedRequestRequestValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e GetSignedRequestRequestValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e GetSignedRequestRequestValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e GetSignedRequestRequestValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e GetSignedRequestRequestValidationError) ErrorName() string {
	return "GetSignedRequestRequestValidationError"
}

// Error satisfies the builtin error interface
func (e GetSignedRequestRequestValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sGetSignedRequestRequest.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = GetSignedRequestRequestValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = GetSignedRequestRequestValidationError{}

// Validate checks the field values on GetSignedRequestResponse with the rules
// defined in the proto definition for this message. If any rules are
// violated, the first error encountered is returned, or nil if there are no violations.
func (m *GetSignedRequestResponse) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on GetSignedRequestResponse with the
// rules defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// GetSignedRequestResponseMultiError, or nil if none found.
func (m *GetSignedRequestResponse) ValidateAll() error {
	return m.validate(true)
}

func (m *GetSignedRequestResponse) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	// no validation rules for SignedRequest

	if len(errors) > 0 {
		return GetSignedRequestResponseMultiError(errors)
	}

	return nil
}

// GetSignedRequestResponseMultiError is an error wrapping multiple validation
// errors returned by GetSignedRequestResponse.ValidateAll() if the designated
// constraints aren't met.
type GetSignedRequestResponseMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m GetSignedRequestResponseMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m GetSignedRequestResponseMultiError) AllErrors() []error { return m }

// GetSignedRequestResponseValidationError is the validation error returned by
// GetSignedRequestResponse.Validate if the designated constraints aren't met.
type GetSignedRequestResponseValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e GetSignedRequestResponseValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e GetSignedRequestResponseValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e GetSignedRequestResponseValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e GetSignedRequestResponseValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e GetSignedRequestResponseValidationError) ErrorName() string {
	return "GetSignedRequestResponseValidationError"
}

// Error satisfies the builtin error interface
func (e GetSignedRequestResponseValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sGetSignedRequestResponse.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = GetSignedRequestResponseValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = GetSignedRequestResponseValidationError{}
