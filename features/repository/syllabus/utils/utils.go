package utils

import (
	"fmt"
	"reflect"
)

// make generic, easy to compare
func FormatName(id string) string {
	return fmt.Sprintf("name_%s", id)
}

func Reverse(s any) {
	n := reflect.ValueOf(s).Len()
	swap := reflect.Swapper(s)
	for start, end := 0, n-1; start < end; start, end = start+1, end-1 {
		swap(start, end)
	}
}
