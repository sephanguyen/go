package maputils

import "fmt"

// PrintMapElems Prints maps of pointer elem
// For debugging purpose
func PrintMapElems[T1 comparable, T2 any](m map[T1]T2) {
	for k, v := range m {
		fmt.Printf("[Key: %v,\nValue: %v]\n", k, v)
	}
}
