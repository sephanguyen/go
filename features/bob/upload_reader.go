package bob

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	pb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"

	"github.com/jackc/fake"
	"google.golang.org/protobuf/types/known/durationpb"
)

func (s *suite) aFileInformationToGeneratePutObjectUrl(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.Request = &pb.PresignedPutObjectRequest{
		PrefixName:    "test#01",
		FileExtension: "pdf",
		Expiry:        durationpb.New(3 * time.Minute),
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) aFileInformationToGenerateResumableUploadUrl(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.Request = &pb.ResumableUploadURLRequest{
		PrefixName:    "test#01",
		Expiry:        durationpb.New(3 * time.Minute),
		FileExtension: "pdf",
		AllowOrigin:   "*",
		ContentType:   "application/pdf",
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) generatePresignUrlToPutObject(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.Response, stepState.ResponseErr = pb.NewUploadServiceClient(s.Conn).GeneratePresignedPutObjectURL(contextWithToken(s, ctx), stepState.Request.(*pb.PresignedPutObjectRequest))
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) generateResumableUploadUrl(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.Response, stepState.ResponseErr = pb.NewUploadServiceClient(s.Conn).GenerateResumableUploadURL(contextWithToken(s, ctx), stepState.Request.(*pb.ResumableUploadURLRequest))
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) returnPresignPutObjectUrl(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	prefixUrl := "http://minio-infras.emulator.svc.cluster.local:9000/manabie/user-upload/test%2301"
	res := stepState.Response.(*pb.PresignedPutObjectResponse)
	if !strings.HasPrefix(res.DownloadUrl, prefixUrl) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expect download url has prefix string: %s, got: %s", prefixUrl, res.DownloadUrl)
	}

	if !strings.HasPrefix(res.PresignedUrl, prefixUrl) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expect presign url has prefix string: %s, got: %s", prefixUrl, res.PresignedUrl)
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) returnResumableUploadUrl(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	prefixUrl := "http://minio-infras.emulator.svc.cluster.local:9000/manabie/user-upload/test%2301"
	res := stepState.Response.(*pb.ResumableUploadURLResponse)
	if !strings.HasPrefix(res.DownloadUrl, prefixUrl) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expect download url has prefix string: %s, got: %s", prefixUrl, res.DownloadUrl)
	}

	if !strings.HasPrefix(res.ResumableUploadUrl, prefixUrl) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expect presign url has prefix string: %s, got: %s", prefixUrl, res.ResumableUploadUrl)
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) theFileCanBeUploadedUsingTheReturnedUrl(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	fileData := []string{}
	for i := 0; i < 1000; i++ {
		fileData = append(fileData, fake.Words())
	}
	buf := &bytes.Buffer{}
	gob.NewEncoder(buf).Encode(fileData)
	fileByte := buf.Bytes()
	r := bytes.NewReader(fileByte)
	presignUrl := stepState.Response.(*pb.PresignedPutObjectResponse)
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodPut, presignUrl.PresignedUrl, r)
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

	if !(res.StatusCode >= 200 && res.StatusCode < 300) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expect status code 2xx, got %d", res.StatusCode)
	}

	// Download file
	resp, err := http.Get(presignUrl.DownloadUrl)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}
	defer resp.Body.Close()
	var file bytes.Buffer
	_, err = io.Copy(&file, resp.Body)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}

	// Check size file
	if binary.Size(fileByte) != binary.Size(file.Bytes()) {
		return StepStateToContext(ctx, stepState), errors.New("file's size changed")
	}

	fileDataDownload := []string{}
	gob.NewDecoder(&file).Decode(&fileDataDownload)
	for i, str := range fileData {
		if fileDataDownload[i] != str {
			return StepStateToContext(ctx, stepState), errors.New("file broken")
		}
	}

	return StepStateToContext(ctx, stepState), nil
}
