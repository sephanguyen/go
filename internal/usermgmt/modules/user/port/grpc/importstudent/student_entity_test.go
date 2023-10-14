package importstudent

import (
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"

	"github.com/stretchr/testify/assert"
)

func TestSchoolHistory_StartDate(t *testing.T) {
	testCases := []struct {
		name           string
		startDateAttr  field.Time
		expectedOutput field.Time
	}{
		{
			name:           "Start date attribute is undefined",
			startDateAttr:  field.NewUndefinedTime(),
			expectedOutput: field.NewNullTime(),
		},
		{
			name:           "Start date attribute is undefined",
			startDateAttr:  field.NewNullTime(),
			expectedOutput: field.NewNullTime(),
		},
		{
			name:           "Start date attribute is defined",
			startDateAttr:  field.NewTime(time.Date(2020, 10, 1, 0, 0, 0, 0, time.UTC)),
			expectedOutput: field.NewTime(time.Date(2020, 10, 1, 0, 0, 0, 0, time.UTC)),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sh := SchoolHistory{
				StartDateAttr: tc.startDateAttr,
			}
			if output := sh.StartDate(); output != tc.expectedOutput {
				t.Errorf("Expected StartDate() to return %v, but got %v", tc.expectedOutput, output)
			}
		})
	}
}

func TestSchoolHistory_EndDate(t *testing.T) {
	testCases := []struct {
		name           string
		endDateAttr    field.Time
		expectedOutput field.Time
	}{
		{
			name:           "end date attribute is undefined",
			endDateAttr:    field.NewUndefinedTime(),
			expectedOutput: field.NewNullTime(),
		},
		{
			name:           "end date attribute is undefined",
			endDateAttr:    field.NewNullTime(),
			expectedOutput: field.NewNullTime(),
		},
		{
			name:           "end date attribute is defined",
			endDateAttr:    field.NewTime(time.Date(2020, 10, 1, 0, 0, 0, 0, time.UTC)),
			expectedOutput: field.NewTime(time.Date(2020, 10, 1, 0, 0, 0, 0, time.UTC)),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var sh = SchoolHistory{
				EndDateAttr: tc.endDateAttr,
			}
			if output := sh.EndDate(); output != tc.expectedOutput {
				t.Errorf("Expected EndDate() to return %v, but got %v", tc.expectedOutput, output)
			}
		})
	}
}

func TestSchoolHistory_SchoolCourseID(t *testing.T) {
	tests := []struct {
		name          string
		sh            SchoolHistory
		expectedValue field.String
	}{
		{
			name:          "Should return default SchoolCourseIDAttr value",
			sh:            SchoolHistory{SchoolCourseIDAttr: field.NewUndefinedString()},
			expectedValue: field.NewNullString(),
		},
		{
			name:          "Should return SchoolCourseIDAttr value",
			sh:            SchoolHistory{SchoolCourseIDAttr: field.NewString("test-course-id")},
			expectedValue: field.NewString("test-course-id"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value := tt.sh.SchoolCourseID()
			assert.Equal(t, value, tt.expectedValue)
		})
	}
}
