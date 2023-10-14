package builders

import (
	"github.com/manabie-com/backend/internal/golibs/sendgrid"

	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type EmailLegacyBuilder struct {
	*mail.SGMailV3

	// SendGrid legacy email template only support substitution data as string.
	// If you have complex data, you should use EmailV3Builder and DynamicTemplateData.
	rootSubstitutionData map[string]string
}

func (b *EmailLegacyBuilder) BuildEmail() *mail.SGMailV3 {
	return b.SGMailV3
}

func (b *EmailLegacyBuilder) WithContents(contentType, value string) *EmailLegacyBuilder {
	b.SGMailV3.AddContent([]*mail.Content{mail.NewContent(contentType, value)}...)
	return b
}

func (b *EmailLegacyBuilder) WithSubject(subject string) *EmailLegacyBuilder {
	b.SGMailV3.Subject = subject
	return b
}

// WithSender accepts 1 correspondent
func (b *EmailLegacyBuilder) WithSender(sender sendgrid.Correspondent) *EmailLegacyBuilder {
	fromMail := &mail.Email{
		Name:    sender.Name,
		Address: sender.Address,
	}
	b.SGMailV3.SetFrom(fromMail)
	return b
}

// WithSubstitutions register root substitution data for personalizations
func (b *EmailLegacyBuilder) WithSubstitutions(substitutions map[string]string) *EmailLegacyBuilder {
	b.rootSubstitutionData = substitutions
	return b
}

func (b *EmailLegacyBuilder) WithRecipients(recipients ...sendgrid.Correspondent) *EmailLegacyBuilder {
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

		b.useSubstitutionData(recipient, p)

		b.SGMailV3.AddPersonalizations(p)
	}
	return b
}

// useSubstitutionData by default will use rootSubstitutionData
// if the Correspondent is not defined with it's own substition data
func (b *EmailLegacyBuilder) useSubstitutionData(recipient sendgrid.Correspondent, p *mail.Personalization) {
	substitutionData := b.rootSubstitutionData
	if len(recipient.SubstitutionData) > 0 {
		substitutionData = recipient.SubstitutionData
	}
	for key, val := range substitutionData {
		p.SetSubstitution(key, val)
	}
}
