package entity

import (
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"

	"github.com/stretchr/testify/assert"
)

func TestNewEnrollmentStatusHistoryWithStartDate(t *testing.T) {
	type args struct {
		e         DomainEnrollmentStatusHistory
		startDate field.Time
	}
	tests := []struct {
		name string
		args args
		want DomainEnrollmentStatusHistory
	}{
		{
			name: "with start date",
			args: args{
				e:         DefaultDomainEnrollmentStatusHistory{},
				startDate: field.NewTime(time.Now()),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			historyWithStartDate := NewEnrollmentStatusHistoryWithStartDate(tt.args.e, tt.args.startDate)
			assert.Equal(t, tt.args.startDate, historyWithStartDate.StartDate())
		})
	}
}
