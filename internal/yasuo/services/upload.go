package services

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type Uploader interface {
	UploadWithContext(ctx aws.Context, input *s3manager.UploadInput, opts ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error)
	UploadWithIterator(ctx aws.Context, iter s3manager.BatchUploadIterator, opts ...func(*s3manager.Uploader)) error
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

type UploadObject struct {
	Data        string
	Bucket      string
	Path        string
	ContentType string
}

func bulkUploadToS3(ctx context.Context, uploader Uploader, objs []*UploadObject) error {
	// Upload the file to S3.
	objects := []s3manager.BatchUploadObject{}
	for _, obj := range objs {
		objects = append(objects, s3manager.BatchUploadObject{
			Object: &s3manager.UploadInput{
				Bucket:      aws.String(obj.Bucket),
				Key:         aws.String(obj.Path),
				Body:        strings.NewReader(obj.Data),
				ACL:         aws.String("public-read"),
				ContentType: aws.String(obj.ContentType),
			},
		})
	}
	if err := uploader.UploadWithIterator(ctx, &s3manager.UploadObjectsIterator{
		Objects: objects,
	}); err != nil {
		return fmt.Errorf("UploadWithContext: %w", err)
	}

	return nil
}

type UploadPayloadObject struct {
	Data        []byte
	Bucket      string
	Path        string
	ContentType string
}

func bulkUploadPayloadToS3(ctx context.Context, uploader Uploader, objs []*UploadPayloadObject) error {
	// Upload the file to S3.
	objects := []s3manager.BatchUploadObject{}
	for _, obj := range objs {
		objects = append(objects, s3manager.BatchUploadObject{
			Object: &s3manager.UploadInput{
				Bucket:      aws.String(obj.Bucket),
				Key:         aws.String(obj.Path),
				Body:        bytes.NewReader(obj.Data),
				ACL:         aws.String("public-read"),
				ContentType: aws.String(obj.ContentType),
			},
		})
	}
	if err := uploader.UploadWithIterator(ctx, &s3manager.UploadObjectsIterator{
		Objects: objects,
	}); err != nil {
		return fmt.Errorf("UploadWithContext: %w", err)
	}

	return nil
}
