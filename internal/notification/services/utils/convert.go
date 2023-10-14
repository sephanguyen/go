package utils

import "fmt"

// 0 -> A
// 1 -> B
// ...
// 25 -> Z
func ConvertNumberToUppercaseChar(i int) (string, error) {
	if i < 0 || i > 25 {
		return "", fmt.Errorf("cannot convert a number is less than 0 or great than 25 to one uppercase letter")
	}
	return string(rune('A' - 0 + i)), nil
}
