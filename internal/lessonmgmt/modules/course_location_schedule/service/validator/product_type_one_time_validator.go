package validator

type ProductTypeOneTimeValidator struct {
	Freq          *int
	TotalNoLesson *int
}

func (v *ProductTypeOneTimeValidator) Validation() string {
	if v.TotalNoLesson == nil {
		return "total_no_lessons is required"
	}
	if v.Freq != nil {
		return "Frequency should be blank"
	}
	return ""
}
