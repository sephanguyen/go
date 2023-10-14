package entity

import (
	"testing"

	"github.com/jackc/pgtype"

	"github.com/stretchr/testify/assert"
)

type TestCase struct {
	name         string
	req          interface{}
	expectedErr  error
	expectedResp interface{}
}

func TestEntityParent_Len(t *testing.T) {
	t.Parallel()

	testCases := []TestCase{
		{
			name: "happy case",
			req: &Parents{
				{ID: pgtype.Text{String: "123", Status: pgtype.Present}},
			},
			expectedResp: int(1),
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			parentsEnt := testCase.req.(*Parents)

			lengthParents := parentsEnt.Len()

			assert.Equal(t, testCase.expectedResp, lengthParents)

		})
	}
}

func TestEntityParent_Ids(t *testing.T) {
	t.Parallel()

	testCases := []TestCase{
		{
			name: "happy case",
			req: &Parents{
				{ID: pgtype.Text{String: "123", Status: pgtype.Present}},
			},
			expectedResp: &Parents{
				{ID: pgtype.Text{String: "123", Status: pgtype.Present}},
			},
		},
		{
			name: "happy case - IDs empty string",
			req: &Parents{
				{ID: pgtype.Text{String: "", Status: pgtype.Null}},
			},
			expectedResp: &Parents{
				{ID: pgtype.Text{String: "", Status: pgtype.Null}},
			},
		},
		{
			name: "happy case - IDs nil",
			req: &Parents{
				{ID: pgtype.Text{}},
			},
			expectedResp: &Parents{
				{ID: pgtype.Text{}},
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			parentsEnt := testCase.req.(*Parents)
			expectedResp := testCase.expectedResp.(*Parents)

			arrIDs := parentsEnt.Ids()

			assert.Equal(t, len(*expectedResp), len(arrIDs))

			for index, id := range arrIDs {
				assert.Equal(t, (*expectedResp)[index].ID.String, id)
			}
		})
	}
}

func TestEntityParent_Emails(t *testing.T) {
	t.Parallel()

	testCases := []TestCase{
		{
			name: "happy case",
			req: &Parents{
				{LegacyUser: LegacyUser{Email: pgtype.Text{String: "123", Status: pgtype.Present}}},
			},
			expectedResp: &Parents{
				{LegacyUser: LegacyUser{Email: pgtype.Text{String: "123", Status: pgtype.Present}}},
			},
		},
		{
			name: "happy case - emails empty string",
			req: &Parents{
				{LegacyUser: LegacyUser{Email: pgtype.Text{String: "", Status: pgtype.Null}}},
			},
			expectedResp: &Parents{},
		},
		{
			name: "happy case - emails nil",
			req: &Parents{
				{LegacyUser: LegacyUser{Email: pgtype.Text{}}},
			},
			expectedResp: &Parents{},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			parentsEnt := testCase.req.(*Parents)
			expectedResp := testCase.expectedResp.(*Parents)

			arrEmails := parentsEnt.Emails()

			assert.Equal(t, len(*expectedResp), len(arrEmails))

			for index, email := range arrEmails {
				assert.Equal(t, (*expectedResp)[index].LegacyUser.Email.String, email)
			}
		})
	}
}

func TestEntityParent_PhoneNumbers(t *testing.T) {
	t.Parallel()

	testCases := []TestCase{
		{
			name: "happy case",
			req: &Parents{
				{LegacyUser: LegacyUser{PhoneNumber: pgtype.Text{String: "123", Status: pgtype.Present}}},
			},
			expectedResp: &Parents{
				{LegacyUser: LegacyUser{PhoneNumber: pgtype.Text{String: "123", Status: pgtype.Present}}},
			},
		},
		{
			name: "happy case - Phone number empty string",
			req: &Parents{
				{LegacyUser: LegacyUser{PhoneNumber: pgtype.Text{String: "", Status: pgtype.Null}}},
			},
			expectedResp: &Parents{},
		},
		{
			name: "happy case - Phone number nil",
			req: &Parents{
				{LegacyUser: LegacyUser{PhoneNumber: pgtype.Text{}}},
			},
			expectedResp: &Parents{},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			parentsEnt := testCase.req.(*Parents)
			expectedResp := testCase.expectedResp.(*Parents)

			arrPhoneNumbers := parentsEnt.PhoneNumbers()

			assert.Equal(t, len(*expectedResp), len(arrPhoneNumbers))

			for index, phoneNumber := range arrPhoneNumbers {
				assert.Equal(t, (*expectedResp)[index].LegacyUser.PhoneNumber.String, phoneNumber)
			}
		})
	}
}
