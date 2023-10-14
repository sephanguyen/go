package validator

import (
	"testing"

	"github.com/manabie-com/backend/internal/lessonmgmt/modules/course_location_schedule/domain"

	"github.com/stretchr/testify/assert"
)

func TestProductTypeValidator_GetValidator(t *testing.T) {
	t.Parallel()

	t.Run("return validator for ProductTypeOneTimeValidator when produt type is one time", func(t *testing.T) {
		totalNoLessonCorrect := 1
		var freqCorrect *int = nil

		v, err := GetValidator(&ParamValidationProductTypeSchedule{ProductType: domain.OneTime, TotalNoLesson: &totalNoLessonCorrect, Freq: freqCorrect})
		assert.Nil(t, err)
		assert.Equal(t, v.Validation(), "")
		assert.IsType(t, v, &ProductTypeOneTimeValidator{})

		var totalNoLessoNotCorrect *int = nil
		v, err = GetValidator(&ParamValidationProductTypeSchedule{ProductType: domain.OneTime, TotalNoLesson: totalNoLessoNotCorrect, Freq: freqCorrect})
		assert.Nil(t, err)
		assert.Equal(t, v.Validation(), "total_no_lessons is required")

		freqNotCorrect := 2
		v, err = GetValidator(&ParamValidationProductTypeSchedule{ProductType: domain.OneTime, TotalNoLesson: &totalNoLessonCorrect, Freq: &freqNotCorrect})
		assert.Nil(t, err)
		assert.Equal(t, v.Validation(), "Frequency should be blank")
	})

	t.Run("return validator for ProductTypeOtherValidator when produt type is frequency", func(t *testing.T) {
		var freqCorrect *int = nil
		var totalNoLessonCorrect *int = nil

		v, err := GetValidator(&ParamValidationProductTypeSchedule{ProductType: domain.Frequency, TotalNoLesson: totalNoLessonCorrect, Freq: freqCorrect})
		assert.Nil(t, err)
		assert.Equal(t, v.Validation(), "")
		assert.IsType(t, v, &ProductTypeOtherValidator{})

		totalNoLessoNotCorrect := 3
		v, err = GetValidator(&ParamValidationProductTypeSchedule{ProductType: domain.Frequency, TotalNoLesson: &totalNoLessoNotCorrect, Freq: freqCorrect})
		assert.Nil(t, err)
		assert.Equal(t, v.Validation(), "total_no_lessons should be blank")

		freqNotCorrect := 2
		v, err = GetValidator(&ParamValidationProductTypeSchedule{ProductType: domain.Frequency, TotalNoLesson: totalNoLessonCorrect, Freq: &freqNotCorrect})
		assert.Nil(t, err)
		assert.Equal(t, v.Validation(), "Frequency should be blank")
	})

	t.Run("return validator for ProductTypeOtherValidator when produt type is slot base", func(t *testing.T) {
		var freqCorrect *int = nil
		var totalNoLessonCorrect *int = nil

		v, err := GetValidator(&ParamValidationProductTypeSchedule{ProductType: domain.SlotBase, TotalNoLesson: totalNoLessonCorrect, Freq: freqCorrect})
		assert.Nil(t, err)
		assert.Equal(t, v.Validation(), "")
		assert.IsType(t, v, &ProductTypeOtherValidator{})

		totalNoLessoNotCorrect := 3
		v, err = GetValidator(&ParamValidationProductTypeSchedule{ProductType: domain.SlotBase, TotalNoLesson: &totalNoLessoNotCorrect, Freq: freqCorrect})
		assert.Nil(t, err)
		assert.Equal(t, v.Validation(), "total_no_lessons should be blank")

		freqNotCorrect := 2
		v, err = GetValidator(&ParamValidationProductTypeSchedule{ProductType: domain.SlotBase, TotalNoLesson: totalNoLessonCorrect, Freq: &freqNotCorrect})
		assert.Nil(t, err)
		assert.Equal(t, v.Validation(), "Frequency should be blank")
	})

	t.Run("return validator for ProductTypeScheduledValidatior when produt type is Scheduled", func(t *testing.T) {
		freqCorrect := 2
		var totalNoLessonCorrect *int = nil

		v, err := GetValidator(&ParamValidationProductTypeSchedule{ProductType: domain.Scheduled, TotalNoLesson: totalNoLessonCorrect, Freq: &freqCorrect})
		assert.Nil(t, err)
		assert.Equal(t, v.Validation(), "")
		assert.IsType(t, v, &ProductTypeScheduledValidatior{})

		totalNoLessoNotCorrect := 3
		v, err = GetValidator(&ParamValidationProductTypeSchedule{ProductType: domain.Scheduled, TotalNoLesson: &totalNoLessoNotCorrect, Freq: &freqCorrect})
		assert.Nil(t, err)
		assert.Equal(t, v.Validation(), "total_no_lessons should be blank")

		var freqNotCorrect *int = nil
		v, err = GetValidator(&ParamValidationProductTypeSchedule{ProductType: domain.Scheduled, TotalNoLesson: totalNoLessonCorrect, Freq: freqNotCorrect})
		assert.Nil(t, err)
		assert.Equal(t, v.Validation(), "Frequency is required")

		freqNotCorrectMinValue := 0
		v, err = GetValidator(&ParamValidationProductTypeSchedule{ProductType: domain.Scheduled, TotalNoLesson: totalNoLessonCorrect, Freq: &freqNotCorrectMinValue})
		assert.Nil(t, err)
		assert.Equal(t, v.Validation(), "Frequency have min value is 1 - Max value is 7")

		freqNotCorrectMaxValue := 8
		v, err = GetValidator(&ParamValidationProductTypeSchedule{ProductType: domain.Scheduled, TotalNoLesson: totalNoLessonCorrect, Freq: &freqNotCorrectMaxValue})
		assert.Nil(t, err)
		assert.Equal(t, v.Validation(), "Frequency have min value is 1 - Max value is 7")
	})
}
