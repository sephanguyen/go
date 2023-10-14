package grpc

import (
	"reflect"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/idutil"
)

func TestNewUserProfileWithID(t *testing.T) {
	tmpID := idutil.ULIDNow()
	type args struct {
		id string
	}
	tests := []struct {
		name string
		args func(t *testing.T) args

		expected string
	}{
		{
			name: "happy case: empty id",
			args: func(t *testing.T) args {
				return args{id: ""}
			},
			expected: "",
		},
		{
			name: "happy case: random id",
			args: func(t *testing.T) args {
				return args{id: tmpID}
			},
			expected: tmpID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tArgs := tt.args(t)

			got := NewUserProfileWithID(tArgs.id)

			if !reflect.DeepEqual(got.UserID().RawValue(), tt.expected) {
				t.Errorf("NewUserProfileWithID got = %v, expected: %v", got, tt.expected)
			}
		})
	}
}
