package sliceutils

import (
	"bytes"
	"encoding/json"
	"reflect"
	"sort"
)

// ContainFunc This method to check if a slice contains an element or not.
// Passing the slice, the element and a predicate to compare.
// It's different from "golang.org/x/exp/slices".Contains
// Please use ContainsFunc instead
func ContainFunc[T any](data []T, c T, f func(T, T) bool) bool {
	for _, v := range data {
		if f(v, c) {
			return true
		}
	}
	return false
}

// ContainsFunc This method to check if a slice contains an element or not.
// Passing the slice, the element and a predicate to compare.
// It's different from "golang.org/x/exp/slices".Contains
// no need c T
func ContainsFunc[T any](data []T, f func(T) bool) bool {
	for _, v := range data {
		if f(v) {
			return true
		}
	}
	return false
}

// Map This method is array.prototype.map() in javascript
// Create a new slice with another data type (T2) from slice T[]
func Map[T any, T2 any](data []T, f func(T) T2) []T2 {
	newArr := make([]T2, len(data))

	for i, e := range data {
		newArr[i] = f(e)
	}

	return newArr
}

// MapSkip This method is array.prototype.map() in javascript
// And check to skip the element if skip() return true
func MapSkip[T any, T2 any](data []T, mapper func(T) T2, skip func(T) bool) []T2 {
	newArr := make([]T2, 0, len(data))

	for _, e := range data {
		if !skip(e) {
			newArr = append(newArr, mapper(e))
		}
	}

	return newArr
}

// Filter This method is array.prototype.filter() in javascript
// Return a new slice which satisfy the condition function
func Filter[T any](data []T, f func(T) bool) []T {
	newArr := make([]T, 0, len(data))

	for _, e := range data {
		if f(e) {
			newArr = append(newArr, e)
		}
	}

	return newArr
}

func Intersect[T comparable](a, b []T) []T {
	aMap := map[T]bool{}
	for _, item := range a {
		aMap[item] = true
	}
	ret := make([]T, 0)
	for _, item := range b {
		if aMap[item] {
			ret = append(ret, item)
		}
	}
	return ret
}

// FilterWithReferenceList This function takes a list (l []T2) and compares each item to the reference list (rl []T1).
// Reference list (rl) is the filter, while list (l) is the values to be filtered
func FilterWithReferenceList[T1, T2 any](rl []T1, l []T2, e func(rl []T1, li T2) bool) []T2 {
	resultList := make([]T2, 0, len(l))

	for _, li := range l {
		if e(rl, li) {
			resultList = append(resultList, li)
		}
	}
	return resultList
}

func UnorderedEqual[T comparable](first, second []T) bool {
	comparedFirst := make([]T, 0)
	comparedFirst = append(comparedFirst, first...)

	comparedSecond := make([]T, 0)
	comparedSecond = append(comparedSecond, second...)

	sortSlice := func(arr []T) {
		sort.SliceStable(arr, func(i, j int) bool {
			a, _ := json.Marshal(arr[i])
			b, _ := json.Marshal(arr[j])

			return bytes.Compare(a, b) == 1
		})
	}
	sortSlice(comparedFirst)
	sortSlice(comparedSecond)
	return reflect.DeepEqual(comparedFirst, comparedSecond)
}

func Contains[T comparable](s []T, e T) bool {
	for _, v := range s {
		if v == e {
			return true
		}
	}
	return false
}

func Remove[T any](data []T, f func(T) bool) []T {
	newArr := make([]T, 0, len(data))

	for _, e := range data {
		if !f(e) {
			newArr = append(newArr, e)
		}
	}

	return newArr
}

func Reduce[T, M any](s []T, f func(M, T) (M, error), initValue M) (M, error) {
	acc := initValue
	for _, v := range s {
		var err error
		acc, err = f(acc, v)
		if err != nil {
			return acc, err
		}
	}
	return acc, nil
}

// MapValuesToSlice
// Extract all values from a map to a slice
func MapValuesToSlice[T1 comparable, T2 any](m map[T1]T2) []T2 {
	s := make([]T2, 0, len(m))
	for _, v := range m {
		s = append(s, v)
	}
	return s
}

// Chunk
// Chunks the slice into multiple slices
// Default chunkSize will be 10
func Chunk[T any](s []T, size int) [][]T {
	if size < 1 {
		size = 10
	}
	if size >= len(s) {
		return [][]T{s}
	}

	var chunks [][]T
	for {
		if len(s) == 0 {
			break
		}
		if len(s) < size {
			size = len(s)
		}
		chunks = append(chunks, s[0:size])
		s = s[size:]
	}
	return chunks
}

// Remove duplicates in a slice of int or string
func RemoveDuplicates[T string | int](sliceList []T) []T {
	allKeys := make(map[T]bool)
	list := []T{}
	for _, item := range sliceList {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
}
