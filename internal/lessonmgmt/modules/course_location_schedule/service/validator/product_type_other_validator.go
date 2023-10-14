package validator

type ProductTypeOtherValidator struct {
	Freq          *int
	TotalNoLesson *int
}

func (v *ProductTypeOtherValidator) Validation() string {
	if v.Freq != nil {
		return "Frequency should be blank"
	}
	if v.TotalNoLesson != nil {
		return "total_no_lessons should be blank"
	}
	return ""
}
