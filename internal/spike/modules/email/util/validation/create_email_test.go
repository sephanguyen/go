package validation

import (
	"context"
	"fmt"
	"testing"

	spb "github.com/manabie-com/backend/pkg/manabuf/spike/v1"

	"github.com/stretchr/testify/assert"
)

func Test_ValidateSendEmailReq(t *testing.T) {
	type testCase struct {
		Name    string
		Req     *spb.SendEmailRequest
		RespErr error
		Setup   func(ctx context.Context, this *testCase)
	}
	testCases := []*testCase{
		{
			Name: "happy case",
			Req: &spb.SendEmailRequest{
				Subject: "subject",
				Content: &spb.SendEmailRequest_EmailContent{
					PlainText: "content",
					HTML:      "content",
				},
				Recipients: []string{"example@manabie.com"},
			},
			RespErr: nil,
			Setup:   func(ctx context.Context, this *testCase) {},
		},
		{
			Name: "empty subject",
			Req: &spb.SendEmailRequest{
				Subject: "",
				Content: &spb.SendEmailRequest_EmailContent{
					PlainText: "content",
					HTML:      "content",
				},
				Recipients: []string{"example@manabie.com"},
			},
			RespErr: fmt.Errorf(ErrMissingSubject),
			Setup:   func(ctx context.Context, this *testCase) {},
		},
		{
			Name: "nil content",
			Req: &spb.SendEmailRequest{
				Subject:    "subject",
				Recipients: []string{"example@manabie.com"},
			},
			RespErr: fmt.Errorf(ErrMissingContent),
			Setup:   func(ctx context.Context, this *testCase) {},
		},
		{
			Name: "empty one content",
			Req: &spb.SendEmailRequest{
				Subject: "subject",
				Content: &spb.SendEmailRequest_EmailContent{
					PlainText: "content",
					HTML:      "",
				},
				Recipients: []string{"example@manabie.com"},
			},
			RespErr: nil,
			Setup:   func(ctx context.Context, this *testCase) {},
		},
		{
			Name: "empty all content",
			Req: &spb.SendEmailRequest{
				Subject: "subject",
				Content: &spb.SendEmailRequest_EmailContent{
					PlainText: "",
					HTML:      "",
				},
				Recipients: []string{"example@manabie.com"},
			},
			RespErr: fmt.Errorf(ErrMissingContent),
			Setup:   func(ctx context.Context, this *testCase) {},
		},
		{
			Name: "nil recipients",
			Req: &spb.SendEmailRequest{
				Subject: "subject",
				Content: &spb.SendEmailRequest_EmailContent{
					PlainText: "content",
					HTML:      "content",
				},
			},
			RespErr: fmt.Errorf(ErrMissingRecipients),
			Setup:   func(ctx context.Context, this *testCase) {},
		},
		{
			Name: "empty recipients",
			Req: &spb.SendEmailRequest{
				Subject: "subject",
				Content: &spb.SendEmailRequest_EmailContent{
					PlainText: "content",
					HTML:      "content",
				},
				Recipients: []string{},
			},
			RespErr: fmt.Errorf(ErrMissingRecipients),
			Setup:   func(ctx context.Context, this *testCase) {},
		},
	}

	for _, tc := range testCases {
		ctx := context.Background()
		tc.Setup(ctx, tc)
		err := ValidateSendEmailRequiredFields(tc.Req)
		assert.Equal(t, tc.RespErr, err)
	}
}
