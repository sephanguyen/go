package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"cloud.google.com/go/storage"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	mock_services "github.com/manabie-com/backend/mock/notification/services"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/api/option"
)

type transportResult struct {
	res *http.Response
	err error
}

type mockTransport struct {
	gotReq  *http.Request
	gotBody []byte
	results []transportResult
}

func (t *mockTransport) addResult(res *http.Response, err error) {
	t.results = append(t.results, transportResult{res, err})
}

func (t *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	t.gotReq = req
	t.gotBody = nil
	if req.Body != nil {
		bytes, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		t.gotBody = bytes
	}
	if len(t.results) == 0 {
		return nil, fmt.Errorf("error handling request")
	}
	result := t.results[0]
	t.results = t.results[1:]
	return result.res, result.err
}

func (t *mockTransport) gotJSONBody() map[string]interface{} {
	m := map[string]interface{}{}
	if err := json.Unmarshal(t.gotBody, &m); err != nil {
		panic(err)
	}
	return m
}

func mockClient(t *testing.T, m *mockTransport) *storage.Client {
	client, err := storage.NewClient(context.Background(), option.WithHTTPClient(&http.Client{Transport: m}))
	if err != nil {
		t.Fatal(err)
	}
	return client
}
func bodyReader(s string) io.ReadCloser {
	return io.NopCloser(strings.NewReader(s))
}
func Test_generateUploadURL(t *testing.T) {
	endpoint := "https:://example.com"
	bucket := "manabie"
	content := "sample content"

	url, fileName := generateUploadURL(endpoint, bucket, content)

	expectedResult := endpoint + "/" + bucket + "/"

	assert.Equal(t, strings.Contains(url, expectedResult), true)
	assert.Equal(t, len(fileName) > 0, true)
}
func Test_uploadToCloudStorage(t *testing.T) {
	bucket := mock.Anything
	path := mock.Anything
	data := mock.Anything
	contentType := mock.Anything
	ctx := context.Background()
	doWrite := func(mt *mockTransport) *storage.Writer {
		client := mockClient(t, mt)
		wc := client.Bucket(bucket).Object(path).If(storage.Conditions{DoesNotExist: true}).NewWriter(ctx)
		wc.ContentType = contentType

		// We can't check that the Write fails, since it depends on the write to the
		// underling mockTransport failing which is racy.
		wc.Write([]byte(data))
		return wc
	}
	t.Run("happy case", func(t *testing.T) {
		mt := &mockTransport{}
		mt.addResult(&http.Response{StatusCode: 200, Body: bodyReader("{}")}, nil)
		wc := doWrite(mt)
		err := uploadToCloudStorage(wc, data, contentType)
		assert.Nil(t, err)
	})

	t.Run("error case", func(t *testing.T) {
		wc := doWrite(&mockTransport{})
		err := uploadToCloudStorage(wc, data, contentType)
		assert.NotNil(t, err)
	})
}

func Test_uploadToS3(t *testing.T) {
	bucket := mock.Anything
	path := mock.Anything
	data := mock.Anything
	contentType := mock.Anything

	t.Run("happy case", func(t *testing.T) {
		uploader := mock_services.Uploader{}
		ctx := context.Background()
		uploader.On("UploadWithContext", ctx, &s3manager.UploadInput{
			Bucket:      aws.String(bucket),
			Key:         aws.String(path),
			Body:        strings.NewReader(data),
			ACL:         aws.String("public-read"),
			ContentType: aws.String(contentType),
		}).Once().Return(&s3manager.UploadOutput{}, nil)

		err := uploadToS3(ctx, &uploader, data, bucket, path, contentType)
		assert.Nil(t, err)
	})

	t.Run("error case", func(t *testing.T) {
		uploader := mock_services.Uploader{}
		ctx := context.Background()
		expectedErr := fmt.Errorf("UploadWithContext: %w", errors.New("some thing"))
		uploader.On("UploadWithContext", ctx, &s3manager.UploadInput{
			Bucket:      aws.String(bucket),
			Key:         aws.String(path),
			Body:        strings.NewReader(data),
			ACL:         aws.String("public-read"),
			ContentType: aws.String(contentType),
		}).Once().Return(&s3manager.UploadOutput{}, expectedErr)

		err := uploadToS3(ctx, &uploader, data, bucket, path, contentType)
		assert.ErrorIs(t, err, expectedErr)
	})
}
