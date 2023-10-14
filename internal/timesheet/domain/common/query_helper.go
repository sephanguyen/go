package common

import (
	"fmt"
	"strings"
)

func ConcatQueryValue(values ...string) string {
	rs := strings.Join(values, "','")
	return fmt.Sprintf("'%s'", rs)
}
