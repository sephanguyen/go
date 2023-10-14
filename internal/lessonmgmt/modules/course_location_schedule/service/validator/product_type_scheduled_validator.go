package validator

type ProductTypeScheduledValidatior struct {
	Freq          *int
	TotalNoLesson *int
}

func (v *ProductTypeScheduledValidatior) Validation() string {
	if v.Freq == nil {
		return "Frequency is required"
	} else if *v.Freq < 1 || *v.Freq > 7 {
		return "Frequency have min value is 1 - Max value is 7"
	}
	if v.TotalNoLesson != nil {
		return "total_no_lessons should be blank"
	}
	return ""
}
