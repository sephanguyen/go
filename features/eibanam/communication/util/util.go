package util

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/manabie-com/backend/features/eibanam/communication/entity"

	"google.golang.org/grpc/metadata"
)

func GenerateFakeAuthenticationToken(firebaseAddr, sub string, userGroup string) (string, error) {
	template := "templates/" + userGroup + ".template"
	resp, err := http.Get("http://" + firebaseAddr + "/token?template=" + template + "&UserID=" + sub)
	if err != nil {
		return "", fmt.Errorf("aValidAuthenticationToken:cannot generate new user token, err: %v", err)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("aValidAuthenticationToken: cannot read token from response, err: %v", err)
	}
	resp.Body.Close()

	return string(b), nil
}
func ContextWithToken(ctx context.Context, token string) context.Context {
	return contextWithToken(ctx, token)
}

func ContextWithTokenAndTimeOut(ctx context.Context, token string) (context.Context, context.CancelFunc) {
	newCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	return contextWithToken(newCtx, token), cancel
}

func contextWithToken(ctx context.Context, token string) context.Context {
	return metadata.AppendToOutgoingContext(contextWithValidVersion(ctx), "token", token)
}

func contextWithValidVersion(ctx context.Context) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "pkg", "com.manabie.liz", "version", "1.0.0")
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
		file, err := os.Open(filepath.Join(workDir, "eibanam", "communication", "samples", languageForm+".txt"))
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

func RandRangeIn(low, hi int) int {
	return low + rand.Intn(hi-low) //nolint:gosec
}

func RandPhoneNumber() string {
	return fmt.Sprintf("+84%d", RandRangeIn(100000000, 999999999))
}

func DeepEqualInt32(a, b []int32) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func DeepEqualString(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func StateFromContext(ctx context.Context) *entity.State {
	state := ctx.Value(entity.StateKey{})
	if state == nil {
		return &entity.State{}
	}
	return state.(*entity.State)
}

func StateToContext(ctx context.Context, state *entity.State) context.Context {
	return context.WithValue(ctx, entity.StateKey{}, state)
}
