package golibs

import (
	"context"
	"fmt"
	"math"
	"mime"
	"net/http"
	"net/url"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type UserInfo struct {
	ResourcePath string
	UserID       string
}

// InArrayString returns true if s is in arr.
func InArrayString(s string, arr []string) bool {
	for _, str := range arr {
		if s == str {
			return true
		}
	}
	return false
}

func WarningIfError(l *zap.Logger, err error, msg string) {
	if err == nil {
		return
	}

	l.Warn(msg, zap.Error(err))
}

// EqualStringArray returns true if arr1 and arr2 are equal.
// It can return wrong result when len(arr1) != len(arr2).
// Deprecated: use golibs.stringutil.SliceEqual instead.
func EqualStringArray(arr1, arr2 []string) bool {
	for i, item1 := range arr1 {
		if item1 != arr2[i] {
			return false
		}
	}
	return true
}

// Uniq returns a new string slice with duplicated items in list removed. Order of items
// is not guaranteed to be preserved.
func Uniq(list []string) []string {
	if list == nil {
		return nil
	}
	out := make([]string, len(list))
	copy(out, list)
	sort.Strings(out)
	uniq := out[:0]
	for _, x := range out {
		if len(uniq) == 0 || uniq[len(uniq)-1] != x {
			uniq = append(uniq, x)
		}
	}
	return uniq
}

// Replace return a new string slice of list with all elements in targets replaced by
// elements in replacements. More specifically, if targets[i] exists in list, it gets
// replaced by replacements[i].
//
// Replace does nothing if len(targets) != len(replacements).
func Replace(list []string, targets []string, replacements []string) []string {
	if len(targets) != len(replacements) {
		return list
	}
	result := make([]string, len(list))
	copy(result, list)
	for i, target := range targets {
		replacement := replacements[i]
		for j, element := range result {
			if element == target {
				result[j] = replacement
			}
		}
	}
	return result
}

// Compare compares arr1 with arr2 and returns the common item slice, the added item from arr2,
// and the removed item from arr1.
//
// Compare returns all nil if either of the two input slices is nil.
func Compare(arr1 []string, arr2 []string) (intersect, added, removed []string) {
	if arr1 == nil || arr2 == nil {
		return nil, nil, nil
	}

	hash := make(map[string]bool)
	for _, item := range arr1 {
		hash[item] = true
	}

	for _, item := range arr2 {
		if hash[item] {
			intersect = append(intersect, item)
		} else {
			added = append(added, item)
		}
	}

	intersect = Uniq(intersect)

	hash2 := make(map[string]bool)
	for _, item := range intersect {
		hash2[item] = true
	}
	for _, item := range arr1 {
		if !hash2[item] {
			removed = append(removed, item)
		}
	}

	added = Uniq(added)
	removed = Uniq(removed)
	return
}

func ToArrayStringPostgres(src []int64) string {
	stringSrc := []string{}
	for _, v := range src {
		stringSrc = append(stringSrc, fmt.Sprint(v))
	}

	return fmt.Sprintf("{%s}", strings.Join(stringSrc, ","))
}

func GetBrightcoveVideoIDFromURL(u string) (string, error) {
	const videoIDParam = "videoId"
	videoURL, err := url.Parse(u)
	if err != nil {
		return "", fmt.Errorf("could not parse brightcove video url %v", err)
	}

	values := videoURL.Query()
	ids, ok := values[videoIDParam]
	if !ok || len(ids) == 0 || len(ids[0]) == 0 {
		return "", fmt.Errorf("could not extract video id from brightcove video url %s", u)
	}

	return ids[0], err
}

func GetUniqueElementStringArray(sArr []string) []string {
	var res []string
	existing := make(map[string]bool)
	for _, s := range sArr {
		if _, ok := existing[s]; !ok {
			res = append(res, s)
			existing[s] = true
		}
	}

	return res
}

func ToArrayStringFromArrayInt64(src []int64) []string {
	result := make([]string, 0, len(src))
	for _, v := range src {
		result = append(result, fmt.Sprint(v))
	}
	return result
}

// ToStringArray is revert funcction of ToArrayStringPostgres but return array string
func ToStringArray(src string) []string {
	src = strings.TrimPrefix(src, "{")
	src = strings.TrimSuffix(src, "}")
	return strings.Split(src, ",")
}

func GetContentType(fileName string) string {
	fileExtension := filepath.Ext(fileName)
	contentType := mime.TypeByExtension(fileExtension)
	if len(contentType) == 0 {
		contentType = "application/octet-stream" //default of google
	}

	return contentType
}

func HeaderToArray(header http.Header) (res []string) {
	for name, values := range header {
		for _, value := range values {
			res = append(res, fmt.Sprintf("%s: %s", name, value))
		}
	}
	return
}

type Stack struct {
	Lock     sync.Mutex
	Elements []interface{}
}

func (s *Stack) IsEmpty() bool {
	return len(s.Elements) == 0
}

func (s *Stack) Push(v interface{}) {
	s.Lock.Lock()
	defer s.Lock.Unlock()

	s.Elements = append(s.Elements, v)
}

func (s *Stack) Pop() (interface{}, error) {
	s.Lock.Lock()
	defer s.Lock.Unlock()

	l := len(s.Elements)
	if l == 0 {
		return 0, fmt.Errorf("empty stack")
	}

	res := s.Elements[l-1]
	s.Elements = s.Elements[:l-1]
	return res, nil
}

func (s *Stack) Peek() (interface{}, error) {
	s.Lock.Lock()
	defer s.Lock.Unlock()

	l := len(s.Elements)
	if l == 0 {
		return 0, fmt.Errorf("empty stack")
	}

	return s.Elements[l-1], nil
}

func (s *Stack) PeekMulti(nItems int) ([]interface{}, error) {
	s.Lock.Lock()
	defer s.Lock.Unlock()

	l := len(s.Elements)
	if l < nItems {
		return nil, fmt.Errorf("not enough items in stack")
	}

	return s.Elements[l-nItems:], nil
}

func CreateCloneOfArrInterface(input []interface{}) []interface{} {
	var result = make([]interface{}, 0, len(input))
	for _, v := range input {
		var clone interface{}
		if v != nil {
			clone = reflect.Indirect(reflect.ValueOf(v)).Interface()
		} else {
			clone = nil
		}
		result = append(result, clone)
	}
	return result
}

func ResourcePathFromCtx(ctx context.Context) string {
	claims := interceptors.JWTClaimsFromContext(ctx)
	resourcePath := ""
	if claims != nil && claims.Manabie != nil {
		resourcePath = claims.Manabie.ResourcePath
	}
	return resourcePath
}

func UserInfoFromCtx(ctx context.Context) *UserInfo {
	claims := interceptors.JWTClaimsFromContext(ctx)
	resourcePath := ""
	userID := ""
	if claims != nil && claims.Manabie != nil {
		resourcePath = claims.Manabie.ResourcePath
		userID = claims.Manabie.UserID
	}
	return &UserInfo{ResourcePath: resourcePath, UserID: userID}
}

func ResourcePathToCtx(ctx context.Context, resourcePath string) context.Context {
	claims := interceptors.JWTClaimsFromContext(ctx)
	if claims == nil {
		claims = &interceptors.CustomClaims{
			Manabie: &interceptors.ManabieClaims{
				ResourcePath: resourcePath,
			},
		}
	}
	if claims.Manabie == nil {
		claims.Manabie = &interceptors.ManabieClaims{
			ResourcePath: resourcePath,
		}
	}
	if claims.Manabie.ResourcePath == "" {
		claims.Manabie.ResourcePath = resourcePath
	}
	ctx = interceptors.ContextWithJWTClaims(ctx, claims)
	return ctx
}

func MaxInt32(a, b int32) int32 {
	return int32(math.Max(float64(a), float64(b)))
}

// TimestamppbToTime will return zero value of time.Time type when input is null
// because when *timestamppb.Timestamp is null, AsTime() func will return 1970-01-01 00:00:00.000
func TimestamppbToTime(t *timestamppb.Timestamp) time.Time {
	if t == nil {
		return time.Time{}
	}

	return t.AsTime()
}

// Function check all elements in 1d-array is not empty (for all elements)
func All(value interface{}) bool {
	slice := reflect.ValueOf(value)
	if slice.Kind() != reflect.Slice {
		return false
	}

	if slice.IsNil() {
		return false
	}

	for i := 0; i < slice.Len(); i++ {
		// retrieve type of interface and check if it is empty
		if reflect.ValueOf(slice.Index(i).Interface()).IsZero() {
			return false
		}
	}
	return true
}

func StringSliceToMap(s []string) map[string]bool {
	res := make(map[string]bool)
	for _, e := range s {
		res[e] = true
	}
	return res
}
