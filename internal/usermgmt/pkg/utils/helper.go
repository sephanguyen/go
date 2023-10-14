package utils

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
)

func CustomUserClaims(userGroup, userID string, schoolID int64) map[string]interface{} {
	if schoolID != 0 {
		return map[string]interface{}{
			"https://hasura.io/jwt/claims": map[string]interface{}{
				"x-hasura-allowed-roles": []string{userGroup},
				"x-hasura-default-role":  userGroup,
				"x-hasura-user-id":       userID,
				"x-hasura-school-id":     strconv.Itoa(int(schoolID)),
			},
		}
	}
	return map[string]interface{}{
		"https://hasura.io/jwt/claims": map[string]interface{}{
			"x-hasura-allowed-roles": []string{userGroup},
			"x-hasura-default-role":  userGroup,
			"x-hasura-user-id":       userID,
		},
	}
}

func ManabieUserCustomClaims(userGroup, userID, resourcePath string) *interceptors.CustomClaims {
	return &interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{
			UserID:       userID,
			ResourcePath: resourcePath,
			UserGroup:    userGroup,
		},
	}
}

// TruncateToDay is used to reset hour, minutes, second of time.Time
func TruncateToDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func IndexOf(haystack []string, needle string) int {
	for i, v := range haystack {
		if v == needle {
			return i
		}
	}
	return -1
}

// SplitWithCapacity splits s by sep and returns a slice of substrings with the given capacity.
// If capacity is less than the number of substrings, it is set to that number.
func SplitWithCapacity(s, sep string, capacity int) []string {
	array := strings.Split(s, sep)
	if capacity < len(array) {
		capacity = len(array)
	}

	res := make([]string, capacity)
	copy(res, array)
	return res
}

func TruncateTimeToStartOfDay(date time.Time) time.Time {
	return time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
}

func CompareStringsRegardlessOrder(firstStrings []string, secondStrings []string) error {
	if len(firstStrings) != len(secondStrings) {
		return fmt.Errorf("length of two string slices are not equal, length of first slice is %v but second slice is %v", len(firstStrings), len(secondStrings))
	}

	for _, firstString := range firstStrings {
		exist := false
		for _, secondString := range secondStrings {
			if firstString == secondString {
				exist = true
			}
		}
		if !exist {
			return fmt.Errorf(`can not find "%s" of first slice in second slice: %s`, firstString, secondStrings)
		}
	}

	return nil
}

func CombineFirstNameAndLastNameToFullName(firstName, lastName string) string {
	return lastName + " " + firstName
}

func CombineFirstNamePhoneticAndLastNamePhoneticToFullName(firstNamePhonetic, lastNamePhonetic string) string {
	return strings.TrimSpace(lastNamePhonetic + " " + firstNamePhonetic)
}

type RoutineHandler = func(start int, end int, wg *sync.WaitGroup, mu *sync.Mutex)

func GroupGoroutines(maxLength int, numGroups int, routineHandler RoutineHandler) {
	wg := sync.WaitGroup{}
	var mu sync.Mutex
	groupSize := maxLength / numGroups
	if groupSize < 1 {
		groupSize = 1
		numGroups = 1
	}
	start := 0

	for i := 0; i < numGroups; i++ {
		wg.Add(1)
		end := start + groupSize
		if i == numGroups-1 {
			end = maxLength
		}
		routineHandler(start, end, &wg, &mu)
		start = end
	}
	wg.Wait()
}

// InArrayInt returns true if value is in slice.
func InArrayInt(value int, slice []int) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}

// MaxTime returns the latest time in a slice of times.
// If the slice is empty, it returns the zero value of time.Time.
func MaxTime(times []time.Time) time.Time {
	if len(times) == 0 {
		return time.Time{}
	}

	sort.Slice(times, func(i, j int) bool {
		return times[i].Before(times[j])
	})

	return times[len(times)-1]
}

// IsFutureDate returns true if the given startTime is later than the current time in the same location.
func IsFutureDate(startTime field.Time) bool {
	now := field.NewTime(time.Now().In(startTime.Time().Location()))
	return startTime.After(&now)
}
