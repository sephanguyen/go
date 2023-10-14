package builders

import (
	"testing"

	"github.com/manabie-com/backend/internal/golibs/sendgrid"
	"github.com/stretchr/testify/assert"
)

func Test_EmailLegacyBuilder(t *testing.T) {
	t.Parallel()
	senderName := "sender"
	senderAddress := "address"
	customArgs := []map[string]string{
		{
			"organization_id": "1",
		},
	}
	testCases := []struct {
		Name             string
		Sender           sendgrid.Correspondent
		Recipients       []sendgrid.Correspondent
		Subject          string
		SubstitutionData map[string]string
		Content          struct {
			Type  string
			Value string
		}
	}{
		{
			Name: "happy case",
			Sender: sendgrid.Correspondent{
				Name:    senderName,
				Address: senderAddress,
			},
			Recipients: []sendgrid.Correspondent{
				{
					Name:            "recipient-1",
					Address:         "address-1",
					CustomArguments: customArgs,
				},
				{
					Name:            "recipient-2",
					Address:         "address-2",
					CustomArguments: customArgs,
				},
			},
			Subject: "Manabie email",
			Content: struct {
				Type  string
				Value string
			}{
				Type:  "text/html",
				Value: "<p>Email content</p>",
			},
			SubstitutionData: map[string]string{
				"username": "user_name",
				"lastname": "last_name",
			},
		},
		{
			Name: "case recipient with local substitution data",
			Sender: sendgrid.Correspondent{
				Name:    senderName,
				Address: senderAddress,
			},
			Recipients: []sendgrid.Correspondent{
				{
					Name:    "recipient-1",
					Address: "address-1",
					SubstitutionData: map[string]string{
						"field_1": "value_1",
						"field_2": "value_2",
					},
				},
				{
					Name:    "recipient-2",
					Address: "address-2",
				},
			},
			Subject: "Manabie email",
			Content: struct {
				Type  string
				Value string
			}{
				Type:  "text/html",
				Value: "<p>Email content</p>",
			},
			SubstitutionData: map[string]string{
				"username": "user_name",
				"lastname": "last_name",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			b := NewEmailLegacyBuilder()
			email := b.WithSender(tc.Sender).
				WithSubject(tc.Subject).
				WithContents(tc.Content.Type, tc.Content.Value).
				WithSubstitutions(tc.SubstitutionData).
				WithRecipients(tc.Recipients...).
				BuildEmail()

			assert.Equal(t, tc.Subject, email.Subject)
			assert.Equal(t, tc.Sender.Name, email.From.Name)
			assert.Equal(t, tc.Sender.Address, email.From.Address)
			for _, content := range email.Content {
				assert.Equal(t, tc.Content.Type, content.Type)
				assert.Equal(t, tc.Content.Value, content.Value)
			}
			assert.Equal(t, len(tc.Recipients), len(email.Personalizations))
			for i, p := range email.Personalizations {
				assert.Equal(t, tc.Recipients[i].Name, p.To[0].Name)
				assert.Equal(t, tc.Recipients[i].Address, p.To[0].Address)
				if len(tc.Recipients[i].SubstitutionData) > 0 {
					assert.Equal(t, tc.Recipients[i].SubstitutionData, p.Substitutions)
				} else {
					assert.Equal(t, tc.SubstitutionData, p.Substitutions)
				}

				if len(tc.Recipients[i].CustomArguments) > 0 {
					assert.Equal(t, tc.Recipients[i].CustomArguments, customArgs)
				}
			}
		})
	}
}
