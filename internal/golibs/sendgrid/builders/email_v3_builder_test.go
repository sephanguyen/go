package builders

import (
	"testing"

	"github.com/manabie-com/backend/internal/golibs/sendgrid"
	"github.com/stretchr/testify/assert"
)

func Test_EmailV3Builder(t *testing.T) {
	t.Parallel()

	templateID := "template-id"
	senderName := "sender"
	senderAddress := "address"
	dynamicTemplateData := make(map[string]interface{})
	dynamicTemplateData["username"] = "John"
	dynamicTemplateData["item_list"] = []struct {
		Name  string
		Count int
	}{
		{
			Name:  "pen",
			Count: 10,
		},
		{
			Name:  "book",
			Count: 1,
		},
	}

	customArgs := []map[string]string{
		{
			"organization_id": "1",
		},
	}

	testCases := []struct {
		Name                string
		Sender              sendgrid.Correspondent
		Recipients          []sendgrid.Correspondent
		DynamicTemplateData map[string]interface{}
		TemplateID          string
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
			DynamicTemplateData: dynamicTemplateData,
			TemplateID:          templateID,
		},
		{
			Name: "case recipient with local dynamic data",
			Sender: sendgrid.Correspondent{
				Name:    senderName,
				Address: senderAddress,
			},
			Recipients: []sendgrid.Correspondent{
				{
					Name:    "recipient-1",
					Address: "address-1",
					DynamicData: map[string]interface{}{
						"username": "Doe",
					},
				},
				{
					Name:    "recipient-2",
					Address: "address-2",
				},
			},
			DynamicTemplateData: dynamicTemplateData,
			TemplateID:          templateID,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			b := NewEmailV3Builder()
			email := b.WithSender(tc.Sender).
				WithTransactionalTemplateID(tc.TemplateID).
				WithDynamicTemplateData(tc.DynamicTemplateData).
				WithRecipients(tc.Recipients...).
				BuildEmail()
			assert.Equal(t, senderName, email.From.Name)
			assert.Equal(t, senderAddress, email.From.Address)
			assert.Equal(t, templateID, email.TemplateID)

			assert.Equal(t, len(tc.Recipients), len(email.Personalizations))
			for i, p := range email.Personalizations {
				assert.Equal(t, tc.Recipients[i].Name, p.To[0].Name)
				assert.Equal(t, tc.Recipients[i].Address, p.To[0].Address)
				if len(tc.Recipients[i].DynamicData) > 0 {
					assert.Equal(t, tc.Recipients[i].DynamicData, p.DynamicTemplateData)
				} else {
					assert.Equal(t, tc.DynamicTemplateData, p.DynamicTemplateData)
				}

				if len(tc.Recipients[i].CustomArguments) > 0 {
					assert.Equal(t, tc.Recipients[i].CustomArguments, customArgs)
				}
			}
		})
	}
}
