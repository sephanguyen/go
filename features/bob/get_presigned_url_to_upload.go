package bob

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"

	"github.com/jackc/fake"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/types/known/durationpb"
)

type fileJson struct {
	Data []string `json:"data"`
}
type presignedUrlUploadContext struct {
	FileJson    fileJson `json:"file_json"`
	StatusCode  int      `json:"status_code"`
	DownloadUrl string   `json:"download_url"`
}

func (s *suite) aSignedInUserHasAExpirationTimeAndAPrefixName(ctx context.Context, arg1, arg2 string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	id := s.newID()
	var err error
	stepState.AuthToken, err = generateValidAuthenticationTokenV1(id)
	s.signedAsAccount(ctx, "school admin")
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	expiry, _ := time.ParseDuration(arg1)
	stepState.Request = &bpb.PresignedPutObjectRequest{
		Expiry:     durationpb.New(expiry),
		PrefixName: arg2,
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) userGetUrlToUploadFile(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := &bpb.PresignedPutObjectRequest{}
	if stepState.Request != nil {
		req = stepState.Request.(*bpb.PresignedPutObjectRequest)
	}
	stepState.Response, stepState.ResponseErr = bpb.NewUploadServiceClient(s.Conn).GeneratePresignedPutObjectURL(contextWithToken(s, ctx), req)

	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) returnAPresignedUrlToUploadFileAndAExpirationTime(ctx context.Context, arg1 string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), stepState.ResponseErr

	}

	res := stepState.Response.(*bpb.PresignedPutObjectResponse)
	expectedExpiry, _ := time.ParseDuration(arg1)
	if expectedExpiry != res.Expiry.AsDuration() {
		return StepStateToContext(ctx, stepState), errors.New(fmt.Sprintf("Expiry expected %v but got %v", expectedExpiry, res.Expiry))
	}

	// check file name in url
	prefixName := stepState.Request.(*bpb.PresignedPutObjectRequest).PrefixName
	if !strings.Contains(res.Name, prefixName) {
		return StepStateToContext(ctx, stepState), errors.New(fmt.Sprintf("File's name not contains %v", prefixName))
	}

	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) userWaitAInterval(ctx context.Context, arg1 string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	interval, _ := strconv.Atoi(arg1)
	time.Sleep(time.Second * time.Duration(interval))
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) uploadAFileViaAPresignedUrl(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	// random data
	var file fileJson
	for i := 0; i < 1000; i++ {
		file.Data = append(file.Data, fake.Words())
	}
	fileByte, _ := json.Marshal(file)
	r := bytes.NewReader(fileByte)

	presignedPutObjectResponse := stepState.Response.(*bpb.PresignedPutObjectResponse)
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodPut, presignedPutObjectResponse.PresignedUrl, r)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}

	res, err := client.Do(req)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}
	defer res.Body.Close()
	var fileBuf bytes.Buffer
	_, err = io.Copy(&fileBuf, res.Body)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}

	stepState.Response = &presignedUrlUploadContext{
		FileJson:    file,
		StatusCode:  res.StatusCode,
		DownloadUrl: presignedPutObjectResponse.DownloadUrl,
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) returnAStatusCode(ctx context.Context, arg1 string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	const successfulCode = "2xx"
	const clientErrorCode = "4xx"
	res := stepState.Response.(*presignedUrlUploadContext)

	var statusCode string
	if res.StatusCode >= 200 && res.StatusCode < 300 {
		statusCode = successfulCode
	} else if res.StatusCode >= 400 && res.StatusCode < 500 {
		statusCode = clientErrorCode
	}

	if arg1 != statusCode {
		return StepStateToContext(ctx, stepState), fmt.Errorf("status Code expected %v but got %v", arg1, res.StatusCode)
	}

	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) fileStorageMustStoreFileIfPresignedUrlNotYetExpired(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctxTest := stepState.Response.(*presignedUrlUploadContext)
	if ctxTest.StatusCode != http.StatusOK {
		// skip this step if url be expired
		return StepStateToContext(ctx, stepState), nil
	}

	// download file
	resp, err := http.Get(ctxTest.DownloadUrl)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}
	defer resp.Body.Close()

	var file bytes.Buffer
	_, err = io.Copy(&file, resp.Body)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}

	actualFile := &fileJson{}
	err = json.Unmarshal(file.Bytes(), actualFile)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}

	// check size
	fileByte, _ := json.Marshal(ctxTest.FileJson)
	if binary.Size(fileByte) != binary.Size(file.Bytes()) {
		return StepStateToContext(ctx, stepState), errors.New("file broken: file's size changed")
	}

	// compare file data
	for i, d := range ctxTest.FileJson.Data {
		if actualFile.Data[i] != d {
			return StepStateToContext(ctx, stepState), errors.New("file broken")
		}
	}

	return StepStateToContext(ctx, stepState), nil
}
