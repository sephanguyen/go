package builders

import (
	"github.com/manabie-com/backend/internal/golibs/sendgrid"

	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

// DefaultEmailBuilder features default usage of sending email,
// which including the use of an transactional template, an sender and multiple recipients.
// Each recipients will supposedly use the same dynamic template data.
type EmailV3Builder struct {
	*mail.SGMailV3

	// rootDynamicTemplateData can be simple data as
	// {
	//	"field_1":"value"
	// }
	// or complex data as
	// {
	// 	"field_1": [
	// 		{
	// 			"item_1": "value"
	// 		},
	// 		{
	// 			"item_2": "value"
	// 		},
	// 	]
	// }
	rootDynamicTemplateData map[string]interface{}
}

func (b *EmailV3Builder) BuildEmail() *mail.SGMailV3 {
	return b.SGMailV3
}

// Transactional template is new template engine of SendGrid EmailV3, along with DynamicTemplateData
func (b *EmailV3Builder) WithTransactionalTemplateID(templateID string) *EmailV3Builder {
	b.SGMailV3.SetTemplateID(templateID)
	return b
}

// WithDynamicTemplateData register root dynamic template data for personalizations
func (b *EmailV3Builder) WithDynamicTemplateData(dynamicData map[string]interface{}) *EmailV3Builder {
	b.rootDynamicTemplateData = dynamicData
	return b
}

// WithSender accepts 1 correspondent
func (b *EmailV3Builder) WithSender(sender sendgrid.Correspondent) *EmailV3Builder {
	fromMail := &mail.Email{
		Name:    sender.Name,
		Address: sender.Address,
	}
	b.SGMailV3.SetFrom(fromMail)
	return b
}

func (b *EmailV3Builder) WithRecipients(recipients ...sendgrid.Correspondent) *EmailV3Builder {
	for _, recipient := range recipients {
		p := mail.NewPersonalization()
		p.AddTos([]*mail.Email{
			mail.NewEmail(recipient.Name, recipient.Address),
		}...)

		for _, c := range recipient.CustomArguments {
			for k, v := range c {
				p.SetCustomArg(k, v)
			}
		}

		b.useDynamicTemplateData(recipient, p)

		b.SGMailV3.AddPersonalizations(p)
	}
	return b
}

// useDynamicTemplateData by default will use rootDynamicTemplateData
// if the Correspondent is not defined with it's own dynamic data
func (b *EmailV3Builder) useDynamicTemplateData(recipient sendgrid.Correspondent, p *mail.Personalization) {
	dynamicTemplateData := b.rootDynamicTemplateData
	if len(recipient.DynamicData) > 0 {
		dynamicTemplateData = recipient.DynamicData
	}
	for key, val := range dynamicTemplateData {
		p.SetDynamicTemplateData(key, val)
	}
}
