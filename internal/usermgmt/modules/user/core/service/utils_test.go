package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_countElementOccurrencesInSlice(t *testing.T) {
	type args struct {
		arr []int
	}
	tests := []struct {
		name string
		args func(t *testing.T) args

		res map[int]int
	}{
		{
			name: "count element occurrences in slice",
			args: func(t *testing.T) args {
				return args{
					arr: []int{1, 2, 2, 2},
				}
			},
			res: map[int]int{1: 1, 2: 3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tArgs := tt.args(t)

			res := countElementOccurrencesInSlice(tArgs.arr)
			assert.Equal(t, tt.res, res)
		})
	}
}

func Test_SplitNameToFirstNameAndLastName(t *testing.T) {

	tests := []struct {
		name string
		args func(t *testing.T) string

		firstName string
		lastName  string
	}{
		{
			name: "normal spaces",
			args: func(t *testing.T) string {
				return "Nguyen Minh Thao TEST"
			},
			firstName: "Minh Thao TEST",
			lastName:  "Nguyen",
		},
		{
			name: "special japanese spaces",
			args: func(t *testing.T) string {
				return "Nguyen　Minh　Thao"
			},
			firstName: "Minh　Thao",
			lastName:  "Nguyen",
		},
		{
			name: "no spaces",
			args: func(t *testing.T) string {
				return "NguyenMinhThao"
			},
			firstName: "",
			lastName:  "NguyenMinhThao",
		},
		{
			name: "empty string",
			args: func(t *testing.T) string {
				return ""
			},
			firstName: "",
			lastName:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tArgs := tt.args(t)

			firstName, lastName := SplitNameToFirstNameAndLastName(tArgs)
			assert.Equal(t, tt.firstName, firstName)
			assert.Equal(t, tt.lastName, lastName)
		})
	}
}

func Test_prependBeforeColumn(t *testing.T) {
	type args struct {
		currentHeaders  string
		currentValues   string
		anchorHeader    string
		headerToPrepend string
		valueToPrepend  string
	}
	tests := []struct {
		name           string
		args           args
		expectedHeader string
		expectedValues string
	}{
		{
			name: "add to the middle",
			args: args{
				currentHeaders:  "A,B,C",
				currentValues:   "a,b,c",
				anchorHeader:    "B",
				headerToPrepend: "B1",
				valueToPrepend:  "b1",
			},
			expectedHeader: "A,B1,B,C",
			expectedValues: "a,b1,b,c",
		},
		{
			name: "add to the tail",
			args: args{
				currentHeaders:  "A,B,C",
				currentValues:   "a,b,c",
				anchorHeader:    "C",
				headerToPrepend: "B1",
				valueToPrepend:  "b1",
			},
			expectedHeader: "A,B,B1,C",
			expectedValues: "a,b,b1,c",
		},
		{
			name: "add to the head",
			args: args{
				currentHeaders:  "A,B,C",
				currentValues:   "a,b,c",
				anchorHeader:    "A",
				headerToPrepend: "A1",
				valueToPrepend:  "a1",
			},
			expectedHeader: "A1,A,B,C",
			expectedValues: "a1,a,b,c",
		},
		{
			name: "can not find the correct position to append",
			args: args{
				currentHeaders:  "A,B,C",
				currentValues:   "a,b,c",
				anchorHeader:    "",
				headerToPrepend: "A1",
				valueToPrepend:  "a1",
			},
			expectedHeader: "A,B,C",
			expectedValues: "a,b,c",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			header, values := prependBeforeColumn(tt.args.currentHeaders, tt.args.currentValues, tt.args.anchorHeader, tt.args.headerToPrepend, tt.args.valueToPrepend)
			assert.Equal(t, tt.expectedHeader, header)
			assert.Equal(t, tt.expectedValues, values)

		})
	}
}
