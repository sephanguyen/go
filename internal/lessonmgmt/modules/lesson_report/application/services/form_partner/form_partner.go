package form_partner

import (
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/idutil"
)

type FormPartner struct {
	formId string
}

func (l *FormPartner) SetConfigFormId(formId string) {
	l.formId = formId
}

func (l *FormPartner) GetConfigFormId() string {
	return l.formId
}

func InitFormPartner(resourcePath string) *FormPartner {
	return &FormPartner{
		formId: fmt.Sprintf(`%s_%s`, resourcePath, idutil.ULIDNow()),
	}
}
