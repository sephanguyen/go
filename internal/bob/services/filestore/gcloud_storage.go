package filestore

import (
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path/filepath"
	"time"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/configs"

	"cloud.google.com/go/storage"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
	ggcreds "google.golang.org/api/iamcredentials/v1"
	"google.golang.org/api/iterator"
)

var _ FileStore = new(GoogleCloudStorage)

type GoogleCloudStorage struct {
	serviceAccountEmail string
	conf                *configs.StorageConfig
	psa                 *ggcreds.ProjectsServiceAccountsService
	cli                 *http.Client
}

func NewGoogleCloudStorage(serviceAccountEmail string, conf *configs.StorageConfig) (*GoogleCloudStorage, error) {
	iamcredentialsService, err := ggcreds.NewService(context.Background())
	if err != nil {
		return nil, fmt.Errorf("could not init IAM credential client: %w", err)
	}

	s := &GoogleCloudStorage{
		serviceAccountEmail: serviceAccountEmail,
		conf:                conf,
		psa:                 ggcreds.NewProjectsServiceAccountsService(iamcredentialsService),
		cli:                 http.DefaultClient,
	}

	return s, nil
}

func NewGoogleCloudStorageWithoutInitIAMCredential(serviceAccountEmail string, conf *configs.StorageConfig) (*GoogleCloudStorage, error) {
	s := &GoogleCloudStorage{
		serviceAccountEmail: serviceAccountEmail,
		conf:                conf,
		cli:                 http.DefaultClient,
	}

	return s, nil
}

// GenerateResumableObjectURL will return A resumable upload allows
// you to resume data transfer operations to Cloud Storage after
// a communication failure has interrupted the flow of data.
//
// Resumable uploads work by sending multiple requests, each of
// which contains a portion of the object you're uploading.
//
// See more at: https://cloud.google.com/storage/docs/resumable-uploads
func (g *GoogleCloudStorage) GenerateResumableObjectURL(ctx context.Context, objectName string, expiry time.Duration, allowOrigin, contentType string) (*url.URL, error) {
	if len(contentType) == 0 {
		contentType = golibs.GetContentType(objectName)
	}

	if allowOrigin == "" {
		allowOrigin = "*"
	}

	u, err := g.signedURL(
		ctx,
		objectName,
		withMethod(http.MethodPost),
		withContentType(contentType),
		withExtensionHeaders(
			map[string][]string{
				"x-goog-resumable":      {"start"},
				"X-Upload-Content-Type": {contentType},
			},
		),
		withExpires(expiry),
		isResumable(),
	)
	if err != nil {
		return nil, err
	}

	return g.initResumableUploadSession(http.MethodPost, u.String(), contentType, allowOrigin)
}

// GenerateResumableObjectURLWithPrivateKey is simulate GenerateResumableObjectURL
// but using privateKey to sign url.
func (g *GoogleCloudStorage) GenerateResumableObjectURLWithPrivateKey(ctx context.Context, objectName string, expiry time.Duration, allowOrigin, contentType string, privateKey []byte) (*url.URL, error) {
	if len(contentType) == 0 {
		contentType = golibs.GetContentType(objectName)
	}

	if allowOrigin == "" {
		allowOrigin = "*"
	}

	u, err := g.signedURL(
		ctx,
		objectName,
		withMethod(http.MethodPost),
		withContentType(contentType),
		withExtensionHeaders(
			map[string][]string{
				"x-goog-resumable":      {"start"},
				"X-Upload-Content-Type": {contentType},
			},
		),
		withExpires(expiry),
		isResumable(),
		withPrivateKey(privateKey),
	)
	if err != nil {
		return nil, err
	}

	return g.initResumableUploadSession(http.MethodPost, u.String(), contentType, allowOrigin)
}

func (g *GoogleCloudStorage) GeneratePresignedPutObjectURL(ctx context.Context, objectName string, expiry time.Duration) (*url.URL, error) {
	return g.signedURL(
		ctx,
		objectName,
		withMethod(http.MethodPut),
		withExpires(expiry),
		withExtensionHeaders(
			map[string][]string{
				"Origin": {"*"},
			},
		),
	)
}

func (g *GoogleCloudStorage) GenerateGetObjectURL(ctx context.Context, objectName string, fileName string, expiry time.Duration) (*url.URL, error) {
	ctxzap.Extract(ctx).Debug("FileStore.signedGCSURL",
		zap.String("objectName", objectName),
		zap.String("bucket", g.conf.Bucket),
		zap.String("fileName", fileName),
	)

	gOpts := storage.SignedURLOptions{
		GoogleAccessID: g.serviceAccountEmail,
		SignBytes: func(b []byte) ([]byte, error) {
			name := fmt.Sprintf("projects/-/serviceAccounts/%s", g.serviceAccountEmail)
			call := g.psa.SignBlob(name, &ggcreds.SignBlobRequest{
				Payload: base64.StdEncoding.EncodeToString(b),
			})

			resp, err := call.Do()
			if err != nil {
				return nil, fmt.Errorf("SignBytes: error when calling SignBlob: %w", err)
			}

			if resp.HTTPStatusCode != 200 {
				return nil, fmt.Errorf("response contained unexpected status code: %d", resp.HTTPStatusCode)
			}

			decoded, err := base64.StdEncoding.DecodeString(resp.SignedBlob)
			if err != nil {
				return nil, err
			}

			return decoded, nil
		},

		Method:          http.MethodGet,
		Expires:         time.Now().Add(expiry),
		Scheme:          storage.SigningSchemeV4,
		QueryParameters: url.Values{},
	}

	gOpts.QueryParameters.Set("response-content-disposition", "attachment; filename=\""+fileName+"\"")

	u, err := storage.SignedURL(g.conf.Bucket, objectName, &gOpts)
	if err != nil {
		return nil, fmt.Errorf("could not sign url: %w", err)
	}

	signedURL, err := url.Parse(u)
	if err != nil {
		return nil, fmt.Errorf("invalid presinged url: %v", err)
	}

	return signedURL, nil
}

// Deprecated
func (g *GoogleCloudStorage) GeneratePublicObjectURL(objectName string) string {
	dir, filename := filepath.Dir(objectName), filepath.Base(objectName)
	if dir == "." {
		filename = url.PathEscape(filename)
	} else {
		filename = fmt.Sprintf("%s/%s", dir, url.PathEscape(filename))
	}
	return fmt.Sprintf("%s/%s/%s", g.conf.Endpoint, g.conf.Bucket, filename)
}

func (g *GoogleCloudStorage) GetObjectInfo(ctx context.Context, bucketName, objectName string) (*StorageObject, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("could new client: %v", err)
	}
	defer client.Close()

	rc, err := client.Bucket(bucketName).Object(objectName).Attrs(ctx)
	if err != nil {
		return nil, fmt.Errorf("could get info object: %v", err)
	}
	return &StorageObject{
		Size: rc.Size,
	}, nil
}

func (g *GoogleCloudStorage) GetObjectsWithPrefix(ctx context.Context, bucketName, prefix, delim string) ([]*StorageObject, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return []*StorageObject{}, fmt.Errorf("storage.NewClient: %v", err)
	}
	defer client.Close()

	it := client.Bucket(bucketName).Objects(ctx, &storage.Query{
		Prefix:    prefix,
		Delimiter: delim,
	})
	st := []*StorageObject{}
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return []*StorageObject{}, fmt.Errorf("Bucket(%q).Objects(): %v", bucketName, err)
		}
		st = append(st, &StorageObject{
			Name: attrs.Name,
		})
	}
	return st, nil
}

func (g *GoogleCloudStorage) DeleteObject(ctx context.Context, bucketName, objectName string) error {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("storage.NewClient: %v", err)
	}
	defer client.Close()

	o := client.Bucket(bucketName).Object(objectName)

	// Optional: set a generation-match precondition to avoid potential race
	// conditions and data corruptions. The request to upload is aborted if the
	// object's generation number does not match your precondition.
	attrs, err := o.Attrs(ctx)
	if err != nil {
		return fmt.Errorf("object.Attrs: %v", err)
	}
	o = o.If(storage.Conditions{GenerationMatch: attrs.Generation})

	if err := o.Delete(ctx); err != nil {
		return fmt.Errorf("Object(%q).Delete: %v", bucketName, err)
	}
	return nil
}

// private method

type signedURLOption func(*storage.SignedURLOptions) *zap.Field

// withContentType will set contentType for signed url
func withContentType(contentType string) signedURLOption {
	return func(s *storage.SignedURLOptions) *zap.Field {
		s.ContentType = contentType
		z := zap.String("contentType", contentType)
		return &z
	}
}

// withExtensionHeaders will set headers for signed url
func withExtensionHeaders(h http.Header) signedURLOption {
	return func(s *storage.SignedURLOptions) *zap.Field {
		headers := golibs.HeaderToArray(h)
		s.Headers = headers
		z := zap.Strings("extensionHeaders", headers)
		return &z
	}
}

// withExpires will set Expires for signed url.
// If not set, default 10 minutes.
func withExpires(exp time.Duration) signedURLOption {
	return func(s *storage.SignedURLOptions) *zap.Field {
		s.Expires = time.Now().Add(exp)
		z := zap.String("expires", exp.String())
		return &z
	}
}

// withMethod will set method for signed url.
// If not set, default GET method.
func withMethod(method string) signedURLOption {
	return func(s *storage.SignedURLOptions) *zap.Field {
		s.Method = method
		z := zap.String("method", method)
		return &z
	}
}

// isResumable will set query parameters 'uploadType' is 'resumable'
// for signed url.
func isResumable() signedURLOption {
	return func(s *storage.SignedURLOptions) *zap.Field {
		s.QueryParameters.Set("uploadType", "resumable")
		z := zap.String("uploadType", "resumable")
		return &z
	}
}

// withPrivateKey will set privateKey for signed url,
// if privateKey not set, default use a service account's
// system-managed private key.
func withPrivateKey(privateKey []byte) signedURLOption {
	return func(s *storage.SignedURLOptions) *zap.Field {
		s.PrivateKey = privateKey
		s.SignBytes = nil
		return nil
	}
}

// signedURL returns a Google Cloud Storage URL for the specified object. signed URLs allow
// the users access to a restricted resource for a limited time by designated method without
// having a Google account or signing in.
func (g *GoogleCloudStorage) signedURL(ctx context.Context, fileName string, opts ...signedURLOption) (*url.URL, error) {
	ctxzap.Extract(ctx).Debug("FileStore.signedGCSURL",
		zap.String("fileName", fileName),
		zap.String("bucketName", g.conf.Bucket),
	)

	gOpts := storage.SignedURLOptions{
		GoogleAccessID: g.serviceAccountEmail,
		SignBytes: func(b []byte) ([]byte, error) {
			name := fmt.Sprintf("projects/-/serviceAccounts/%s", g.serviceAccountEmail)
			call := g.psa.SignBlob(name, &ggcreds.SignBlobRequest{
				Payload: base64.StdEncoding.EncodeToString(b),
			})

			resp, err := call.Do()
			if err != nil {
				return nil, fmt.Errorf("SignBytes: error when calling SignBlob: %w", err)
			}

			if resp.HTTPStatusCode != 200 {
				return nil, fmt.Errorf("response contained unexpected status code: %d", resp.HTTPStatusCode)
			}

			decoded, err := base64.StdEncoding.DecodeString(resp.SignedBlob)
			if err != nil {
				return nil, err
			}

			return decoded, nil
		},

		Method:          http.MethodGet,
		Expires:         time.Now().Add(10 * time.Minute),
		Scheme:          storage.SigningSchemeV4,
		QueryParameters: url.Values{},
	}
	gOpts.QueryParameters.Set("name", fileName)

	// set options
	zapFields := make([]zap.Field, 0, len(opts))
	for _, opt := range opts {
		if z := opt(&gOpts); z != nil {
			zapFields = append(zapFields, *z)
		}
	}
	ctxzap.Extract(ctx).Debug("FileStore.signedGCSURL.options", zapFields...)

	u, err := storage.SignedURL(g.conf.Bucket, fileName, &gOpts)
	if err != nil {
		return nil, fmt.Errorf("could not sign url: %w", err)
	}

	signedURL, err := url.Parse(u)
	if err != nil {
		return nil, fmt.Errorf("invalid presinged url: %v", err)
	}

	return signedURL, nil
}

// initResumableUploadSession will  a session URI, which you use in
// subsequent requests to upload the actual data.
// Google Cloud Storage doc: https://cloud.google.com/storage/docs/performing-resumable-uploads#initiate-session
func (g *GoogleCloudStorage) initResumableUploadSession(method, u, contentType, allowOrigin string) (*url.URL, error) {
	req, err := http.NewRequest(method, u, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("could not init resumable upload session: %v", err)
	}
	req.Header.Set("X-Upload-Content-Type", contentType)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("x-goog-resumable", "start")
	req.Header.Set("Origin", allowOrigin)

	res, err := g.cli.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could not init resumable upload session: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusCreated {
		bodyBytes, _ := ioutil.ReadAll(res.Body)
		bodyString := string(bodyBytes)
		return nil, fmt.Errorf("could not init resumable upload session: status code:%d, header:%s, body:%s", res.StatusCode, res.Header, bodyString)
	}
	sessionURI := res.Header.Get("Location")

	return url.Parse(sessionURI)
}

// Move object function for GCS
func (g *GoogleCloudStorage) MoveObject(ctx context.Context, srcObjectName, destObjetName string) error {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return &Error{
			ErrorCode: UnknownError,
			Err:       err,
		}
	}
	defer client.Close()

	bucketName := g.conf.Bucket
	src := client.Bucket(bucketName).Object(srcObjectName)
	dst := client.Bucket(bucketName).Object(destObjetName)

	// Check source object is exists or not.
	if _, err := src.Attrs(ctx); err != nil {
		return &Error{
			ErrorCode: FileNotFoundError,
			Err:       fmt.Errorf("Object(%s/%s).Attrs: %v", bucketName, srcObjectName, err),
		}
	}

	// Copy the object to the new bucket.
	if _, err := dst.CopierFrom(src).Run(ctx); err != nil {
		return &Error{
			ErrorCode: UnknownError,
			Err:       fmt.Errorf("Object(%s/%s).CopierFrom(%s/%s).Run: %v", bucketName, destObjetName, bucketName, srcObjectName, err),
		}
	}

	// Delete the old object.
	if err := src.Delete(ctx); err != nil {
		return &Error{
			ErrorCode: UnknownError,
			Err:       fmt.Errorf("Object(%s/%s).Delete: %v", bucketName, srcObjectName, err),
		}
	}

	return nil
}
