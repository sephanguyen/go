package constant

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type TestCase struct {
	name         string
	ctx          context.Context
	req          interface{}
	expectedErr  error
	setup        func(ctx context.Context)
	expectedResp interface{}
}

func TestConstant_PhoneNumberPattern(t *testing.T) {
	t.Parallel()
	_, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	testCases := []TestCase{
		{
			name:         "happy case",
			req:          []string{"01234567", "0123456", "01234567890123456789"},
			expectedResp: true,
		},
		{
			name:         "fail case",
			req:          []string{"abc", "+01234567", "", " ", "01234567a"},
			expectedResp: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Log(testCase.name)
			for _, req := range testCase.req.([]string) {
				result, err := regexp.MatchString(PhoneNumberPattern, req)
				assert.Equal(t, testCase.expectedResp, result)
				assert.Nil(t, err)
			}
		})
	}
}

func TestConstant_EmailPattern(t *testing.T) {
	t.Parallel()
	_, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	testCases := []TestCase{
		{
			name: "happy case",
			req: []string{
				"abc@gmail.com",
				"abc@manabie.com",
				"abc.manabie@manabie.com",
				"0987654321@example.com",
				"example@email.com",
				"example@email.co.jp",
				"example@email.museum",
				"example.first.middle.lastname@email.com",
				"example@subdomain.email.com",
				"example+firstname+lastname@email.com",
				"example.firstname-lastname@email.com",
				"_______@email.com",
				"example@[234.234.234.234]",
			},
			expectedResp: true,
		},
		{
			name:         "fail case",
			req:          []string{"abc", "+01234567", "", " ", "01234567a", "example@234.234.234.234", "abc  @gmail.com"},
			expectedResp: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Log(testCase.name)

			for _, req := range testCase.req.([]string) {
				result, err := regexp.MatchString(EmailPattern, req)

				if err != nil {
					assert.Equal(t, testCase.expectedErr.Error(), err.Error())
				} else {
					assert.Equal(t, testCase.expectedResp, result)
				}
			}

		})
	}
}

func TestConstant_ExtractTextBetweenQuotesPattern(t *testing.T) {
	t.Parallel()
	_, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	testCases := []TestCase{
		{
			name: "happy case",
			req: []string{
				`"random" text`,
				`'random' text`,
			},
			expectedResp: true,
		},
		{
			name: "fail case",
			req: []string{
				`random text`,
				`"random text`,
				`'random text`,
				`random" text`,
				`random' text`,
			},
			expectedResp: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Log(testCase.name)

			for _, req := range testCase.req.([]string) {
				result, err := regexp.MatchString(ExtractTextBetweenQuotesPattern, req)

				if err != nil {
					assert.Equal(t, testCase.expectedErr.Error(), err.Error())
				} else {
					assert.Equal(t, testCase.expectedResp, result)
				}
			}

		})
	}
}

func TestConstant_UsernamePattern(t *testing.T) {
	t.Parallel()
	_, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	testCases := []TestCase{
		{
			name: "happy case",
			req: []string{
				"abc",
				"abc123",
				"123abc",
				"abc123abc",
				"ABC",
				"ABC123",
				"123ABC",
				"ABC123ABC",
				"abcABC",
				"abcABC123",
				"abc123ABC",
				"ABCabc",
			},
			expectedResp: true,
		},
		{
			name: "fail case",
			req: []string{
				"abc@",
				"abc@123",
				"abc@123@",
				"abc@123@abc",
				"abc@ABC",
				"abc@ABC123",
				"abc.123ABC",
				"abc@abcABC",
			},
			expectedResp: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Log(testCase.name)

			for _, req := range testCase.req.([]string) {
				result, err := regexp.MatchString(UsernamePattern, req)

				if err != nil {
					assert.Equal(t, testCase.expectedErr.Error(), err.Error())
				} else {
					assert.Equal(t, testCase.expectedResp, result)
				}
			}

		})
	}
}
