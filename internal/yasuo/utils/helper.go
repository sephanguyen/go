package utils

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/idutil"
	bobpb "github.com/manabie-com/backend/pkg/genproto/bob"
	pb "github.com/manabie-com/backend/pkg/genproto/yasuo"
)

// PbSubject2String converts a bob's pb.Subject slice to a string slice.
func PbSubject2String(s []bobpb.Subject) []string {
	r := make([]string, 0, len(s))
	for _, v := range s {
		r = append(r, v.String())
	}

	return r
}

func GenerateQuestionURL(env, bucket, questionID string) string {
	randVal := idutil.ULIDNow()
	return fmt.Sprintf("%s/%s/%s/%s.html", bucket, env, questionID, randVal)
}

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

// PbOrderStatus2String converts a yasuo's pb.OrderStatus slice to a string slice.
func PbOrderStatus2String(s []pb.OrderStatus) []string {
	r := make([]string, 0, len(s))
	for _, v := range s {
		r = append(r, v.String())
	}

	return r
}

// Character to use, exclude 0,1,I,O
var (
	charMap = []string{
		"U", "V", "W", "X", "Y", "Z",
		"A", "B", "C", "D", "E", "F", "G", "H", "J",
		"2", "3", "4", "5", "6", "7", "8", "9",
		"K", "L", "M", "N", "P", "Q", "R", "S", "T",
	}
)

const (
	codePrefix = "MO" // Manabie order
	randLen    = 3
	randMin    = 101
	randMax    = 999
)

func getRandValue(seed uint, randMin, randMax, randLen int) int {
	localRand := rand.New(rand.NewSource(int64(seed)))
	rawValue := fmt.Sprintf("%d", localRand.Intn((randMax-randMin)+1)+randMin)
	randValue, _ := strconv.Atoi(rawValue[:randLen])
	return randValue
}

// reserveString does what the name says.
// Reference: https://stackoverflow.com/questions/1752414/how-to-reverse-a-string-in-go
func reserveString(s string) (result string) {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < len(r)/2; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}

func IsContain(arr []string, value string) bool {
	for _, item := range arr {
		if item == value {
			return true
		}
	}
	return false
}

func sliceIndex(length int, f func(i int) bool) int {
	for i := 0; i < length; i++ {
		if f(i) {
			return i
		}
	}

	return -1
}

func n2dec(val string) (string, error) {
	base := len(charMap)

	var decode int64 = 0
	val = reserveString(val)
	for digit := 0; digit < len(val); digit++ {
		char := string(val[digit])
		if !IsContain(charMap, char) {
			return "", errors.New("invalid input")
		}

		index := int64(sliceIndex(len(charMap), func(i int) bool { return charMap[i] == char }))
		decode += index * int64(math.Pow(float64(base), float64(digit)))
	}
	return fmt.Sprintf("%d", decode), nil
}

func dec2n(val int64) string {
	base := len(charMap)
	var result string
	for {
		index := val % int64(base)
		result = charMap[index] + result
		if val = int64(math.Floor(float64(val) / float64(base))); val == 0 {
			break
		}
	}
	return result
}

func EncodeNumber2String(n uint) string {
	randVal := getRandValue(n, randMin, randMax, randLen)

	idTemp := fmt.Sprintf("%03d%09d", randVal, n)
	id, _ := strconv.ParseInt(idTemp, 10, 64)
	return dec2n(id)
}

func DecodeString2Number(encoded string) (uint, error) {
	decoded, err := n2dec(encoded)
	if err != nil {
		return 0, err
	}
	if len(decoded) < randLen+1 {
		return 0, errors.New("invalid encoded")
	}
	id, _ := strconv.ParseInt(decoded[randLen:], 10, 64)
	decodedRandVal, _ := strconv.Atoi(decoded[:randLen])

	randVal := getRandValue(uint(id), randMin, randMax, randLen)
	if randVal != decodedRandVal {
		return 0, errors.New("invalid rand value")
	}

	return uint(id), nil
}

func EncodeOrderID2String(n uint) string {
	return codePrefix + EncodeNumber2String(n)
}

func DecodeString2OrderID(encoded string) (uint, error) {
	encoded = strings.ReplaceAll(encoded, codePrefix, "")
	return DecodeString2Number(encoded)
}

func IsArrayMatch(arrayLength int, match func(i int) bool) bool {
	for i := 0; i < arrayLength; i++ {
		if match(i) {
			return true
		}
	}

	return false
}
