package errorx

import (
	"fmt"
	"testing"

	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/errcode"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestReturnFirstErr(t *testing.T) {
	t.Parallel()

	t.Run("return first error when have only one error", func(t *testing.T) {
		err1 := errors.New("error 1")

		err := ReturnFirstErr(err1)
		assert.Equal(t, err1, err)
	})

	t.Run("return first error when have multiple error", func(t *testing.T) {
		err1 := errors.New("error 1")
		err2 := errors.New("error 2")
		err3 := errors.New("error 3")

		err := ReturnFirstErr(err1, err2, err3)
		assert.Equal(t, err1, err)
	})
}

func Test_getLastFieldName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"field1.field2.field3", "field3"},
		{"field1", "field1"},
		{"field1.field2.", ""},
		{"", ""},
	}
	for _, test := range tests {
		result := getLastFieldName(test.input)
		if result != test.expected {
			t.Errorf("getLastFieldName(%q) = %q, expected %q", test.input, result, test.expected)
		}
	}
}

func TestPbErrorMessage(t *testing.T) {
	tests := []struct {
		inputErr  errcode.Error
		fieldName string
		code      int32
		index     int32
	}{
		{errcode.Error{FieldName: "fieldName", Err: fmt.Errorf("error message"), Code: 10, Index: 2}, "fieldName", 10, 2},
		{errcode.Error{FieldName: `"fieldName"`, Err: fmt.Errorf("error message"), Code: 10, Index: 2}, "fieldName", 10, 2},
		{errcode.Error{FieldName: "field1.field2", Err: fmt.Errorf("error message"), Code: 10, Index: 2}, "field2", 10, 2},
	}

	for _, test := range tests {
		result := PbErrorMessage(test.inputErr)
		if result.FieldName != test.fieldName {
			t.Errorf("PbErrorMessage(%q).FieldName = %q, expected %q", test.inputErr, result.FieldName, test.fieldName)
		}
		if result.Error != test.inputErr.Error() {
			t.Errorf("PbErrorMessage(%q).Error = %q, expected %q", test.inputErr, result.Error, test.inputErr.Error())
		}
		if result.Code != test.code {
			t.Errorf("PbErrorMessage(%q).Code = %d, expected %d", test.inputErr, result.Code, test.code)
		}
		if result.Index != test.index {
			t.Errorf("PbErrorMessage(%q).Index = %d, expected %d", test.inputErr, result.Index, test.index)
		}
	}
}

func TestExtractFieldName(t *testing.T) {
	type args struct {
		msg string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "with double quotes",
			args: args{
				msg: `"text content"`,
			},
			want: `text content`,
		},
		{
			name: "with single quotes",
			args: args{
				msg: `'text content'`,
			},
			want: `text content`,
		},
		{
			name: "without quote",
			args: args{
				msg: `text content`,
			},
			want: `text content`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ExtractFieldName(tt.args.msg); got != tt.want {
				t.Errorf("ExtractFieldName() = %v, want %v", got, tt.want)
			}
		})
	}
}
