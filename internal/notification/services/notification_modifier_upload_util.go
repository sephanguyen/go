package services

//nolint
import (
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/manabie-com/backend/internal/golibs/idutil"
)

type Uploader interface {
	UploadWithContext(ctx aws.Context, input *s3manager.UploadInput, opts ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error)
}

func (svc *NotificationModifierService) UploadHTMLContent(ctx context.Context, content string) (string, error) {
	url, fileName := generateUploadURL(svc.StorageConfig.Endpoint, svc.StorageConfig.Bucket, content)
	bucket := svc.StorageConfig.Bucket
	if svc.Env != "local" {
		client, err := storage.NewClient(ctx)
		if err != nil {
			return "", fmt.Errorf("err storage.NewClient: %w", err)
		}
		wc := client.Bucket(bucket).Object(fileName[1:]).NewWriter(ctx)
		err = uploadToCloudStorage(wc, content, "text/html; charset=utf-8")
		if err != nil {
			return "", fmt.Errorf("err uploadToCloudStorage: %w", err)
		}
	} else {
		err := uploadToS3(ctx, svc.Uploader, content, bucket, fileName, "text/html; charset=UTF-8")
		if err != nil {
			return "", fmt.Errorf("err uploadToS3: %w", err)
		}
	}

	return url, nil
}

// nolint
func generateUploadURL(endpoint, bucket, content string) (generatedUrl, fileName string) {
	h := md5.New()
	_, err := io.WriteString(h, content)

	generatedFileName := ""
	if err != nil {
		generatedFileName = fmt.Sprintf("%s.html", idutil.ULIDNow())
	} else {
		generatedFileName = fmt.Sprintf("%x.html", h.Sum(nil))
	}

	fileName = "/content/" + generatedFileName
	generatedUrl = endpoint + "/" + bucket + fileName

	return
}

func uploadToS3(ctx context.Context, uploader Uploader, data, bucket, path, contentType string) error {
	// Upload the file to S3.
	_, err := uploader.UploadWithContext(ctx, &s3manager.UploadInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(path),
		Body:        strings.NewReader(data),
		ACL:         aws.String("public-read"),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return fmt.Errorf("UploadWithContext: %w", err)
	}

	return nil
}

func uploadToCloudStorage(wc *storage.Writer, data, contentType string) error {
	wc.ContentType = contentType
	if _, err := io.Copy(wc, strings.NewReader(data)); err != nil {
		return fmt.Errorf("io.Copy: %v", err)
	}

	if err := wc.Close(); err != nil {
		return fmt.Errorf("Writer.Close: %v", err)
	}
	return nil
}
