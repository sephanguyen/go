package support

import (
	"container/list"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

func SplitAndTrim(s, sep string) []string {
	results := strings.Split(s, sep)
	for i := range results {
		results[i] = strings.TrimSpace(results[i])
	}
	return results
}

func ConvertStringToPointInt(s string) (*int, error) {
	if s == "" {
		return nil, nil
	}
	val, err := strconv.Atoi(s)
	if err != nil {
		return nil, err
	}
	return &val, nil
}

func Compare(arr1 []string, arr2 []string) (bool, int) {
	sort.Strings(arr1)
	sort.Strings(arr2)
	for i, v := range arr1 {
		if v != arr2[i] {
			return false, i
		}
	}
	return true, -1
}

func caculateWeek(from int, to int) []string {
	result := make([]string, 0, 100)
	for i := from; i <= to; i++ {
		result = append(result, fmt.Sprint(i))
	}
	return result
}

func Enqueue(q *list.List, data string) {
	q.PushBack(data)
}

func DeQueue(q *list.List) string {
	frontValue := q.Front()
	q.Remove(frontValue)
	return fmt.Sprint(frontValue.Value)
}

func ConvertAcademicWeeks(str string) ([]string, error) {
	q := list.New()
	result := make([]string, 0, 100)
	strNumber := ""
	lStr := len(str)
	for index, char := range str {
		strV := string(char)
		if strV != "-" && strV != "_" {
			strNumber += strV
		} else {
			Enqueue(q, strNumber)
			Enqueue(q, strV)
			strNumber = ""
		}
		if index == lStr-1 {
			Enqueue(q, strNumber)
		}
	}
	q2 := list.New()
	for q.Len() > 0 {
		v := DeQueue(q)
		if v != "_" && v != "-" {
			Enqueue(q2, v)
		} else if q2.Len() > 0 {
			reNumber := DeQueue(q2)
			if v == "_" {
				result = append(result, fmt.Sprint(reNumber))
			} else {
				nextNumber := DeQueue(q)
				pReNumber, err := ConvertStringToPointInt(fmt.Sprint(reNumber))
				if err != nil {
					return result, nil
				}
				pNextNumber, err := ConvertStringToPointInt(fmt.Sprint(nextNumber))
				if err != nil {
					return result, nil
				}
				result = append(result, caculateWeek(*pReNumber, *pNextNumber)...)
			}
		}
	}
	if q2.Len() > 0 {
		v := DeQueue(q2)
		result = append(result, v)
	}
	if q.Len() > 0 {
		v := DeQueue(q)
		result = append(result, v)
	}
	return result, nil
}
