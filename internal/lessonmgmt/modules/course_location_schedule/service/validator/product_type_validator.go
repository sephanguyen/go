package validator

import (
	"fmt"

	"github.com/manabie-com/backend/internal/lessonmgmt/modules/course_location_schedule/domain"
)

type ProductTypeValidator interface {
	Validation() string
}
type ParamValidationProductTypeSchedule struct {
	ProductType   domain.ProductTypeSchedule
	TotalNoLesson *int
	Freq          *int
}

func GetValidator(p *ParamValidationProductTypeSchedule) (ProductTypeValidator, error) {
	switch p.ProductType {
	case domain.OneTime:
		return &ProductTypeOneTimeValidator{
			Freq: p.Freq, TotalNoLesson: p.TotalNoLesson,
		}, nil
	case domain.Scheduled:
		return &ProductTypeScheduledValidatior{
			Freq: p.Freq, TotalNoLesson: p.TotalNoLesson,
		}, nil
	case domain.Frequency, domain.SlotBase:
		return &ProductTypeOtherValidator{Freq: p.Freq, TotalNoLesson: p.TotalNoLesson}, nil
	}
	return nil, fmt.Errorf("wrong product type schedule passed")
}
