package builders

import (
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

func NewEmailV3Builder() *EmailV3Builder {
	b := &EmailV3Builder{}
	b.SGMailV3 = &mail.SGMailV3{}
	b.rootDynamicTemplateData = make(map[string]interface{})
	return b
}

func NewEmailLegacyBuilder() *EmailLegacyBuilder {
	b := &EmailLegacyBuilder{}
	b.SGMailV3 = &mail.SGMailV3{}
	b.rootSubstitutionData = make(map[string]string)
	return b
}
