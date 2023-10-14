// nolint
package common

import (
	"bufio"
	"crypto/md5"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/stringutil"
	"github.com/manabie-com/backend/internal/golibs/try"
	"github.com/manabie-com/backend/internal/notification/entities"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
)

func RandRangeIn(low, hi int) int {
	return low + rand.Intn(hi-low) //nolint:gosec
}

func generateUploadURL(endpoint, bucket, content string) (errorurl, fileName string) {
	h := md5.New()
	io.WriteString(h, content)
	fileName = "/content/" + fmt.Sprintf("%x.html", h.Sum(nil))

	return endpoint + "/" + bucket + fileName, fileName
}

type userOption func(u *entity.LegacyUser)

func withID(id string) userOption {
	return func(u *entity.LegacyUser) {
		_ = u.ID.Set(id)
	}
}

func withRole(group string) userOption {
	return func(u *entity.LegacyUser) {
		_ = u.Group.Set(group)
	}
}

func LoadExampleName() ([]string, error) {
	workDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	filenames := []string{"hiragana", "kanji", "katakana", "english"}
	sampleNames := make([]string, 0)

	files := make([]*os.File, 0)
	defer func() {
		for _, f := range files {
			f.Close()
		}
	}()

	for _, languageForm := range filenames {
		// open file
		file, err := os.Open(filepath.Join(workDir, "communication", "common", "samples", languageForm+".txt"))
		if err != nil {
			return nil, err
		}

		// for defer close file
		files = append(files, file)
		slc, err := readIntoSlice(file)
		if err != nil {
			return nil, err
		}

		sampleNames = append(sampleNames, slc...)
	}
	return sampleNames, nil
}

func readIntoSlice(f io.Reader) ([]string, error) {
	sl := make([]string, 0)
	reader := bufio.NewReader(f)
	line := 0
	for {
		bs, _, err := reader.ReadLine()
		if err != nil {
			if err == io.EOF {
				return sl, nil
			}
			return nil, fmt.Errorf("failed to read file at line %d: %s", line, err)
		}
		if len(bs) == 0 {
			return nil, fmt.Errorf("unexpected blank line in seed data file at line %d", line)
		}
		sl = append(sl, string(bs))
		line++
	}
}

func StrToBool(str string) bool {
	return str == "true"
}

func substr(s string, from, length int) string {
	//create array like string view
	wb := []string{}
	wb = strings.Split(s, "")

	//miss nil pointer error
	to := from + length

	if to > len(wb) {
		to = len(wb)
	}

	if from > len(wb) {
		from = len(wb)
	}

	return strings.Join(wb[from:to], "")
}

func GetRandomKeywordFromStrings(strings []string) string {
	idx := RandRangeIn(0, len(strings))
	for i, item := range strings {
		if i == idx {
			if len(item) <= 1 {
				continue
			}
			idx2 := RandRangeIn(1, len(item))
			// subStr := string([]rune(item)[0 : idx2-1])
			subStr := substr(item, 0, idx2-1)
			return subStr
		}
	}
	return ""
}

func CheckTargetGroupFilterNameValues[C entities.GInfoNotificationFilter](expectFilter, actualFilter C) error {
	expectedNames := expectFilter.GetNameValues()
	actualNames := actualFilter.GetNameValues()
	if len(expectedNames) != len(actualNames) {
		return fmt.Errorf("expected Name to have %d elements, got %d elements", len(expectFilter.GetNameValues()), len(actualFilter.GetNameValues()))
	}
	if !stringutil.SliceEqual(expectedNames, actualNames) {
		return fmt.Errorf("unequal elements expect %+v, got %+v", expectedNames, actualNames)
	}
	return nil
}

func Reverse(s any) {
	n := reflect.ValueOf(s).Len()
	swap := reflect.Swapper(s)
	for start, end := 0, n-1; start < end; start, end = start+1, end-1 {
		swap(start, end)
	}
}

func StrToInt(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}

func doRetry(f func() (bool, error)) error {
	return try.Do(func(attempt int) (bool, error) {
		retry, err := f()
		if err != nil {
			if retry {
				time.Sleep(2 * time.Second)
				return attempt < 10, err
			}
			return false, err
		}
		return false, nil
	})
}
