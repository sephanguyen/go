package s3

// nolint
import (
	"crypto/md5"
	"fmt"
	"io"
)

// nolint
func GenerateUploadURL(endpoint, bucket, content string) (string, error) {
	h := md5.New()
	_, err := io.WriteString(h, content)
	if err != nil {
		return "", err
	}

	fileName := "/content/" + fmt.Sprintf("%x.html", h.Sum(nil))

	return endpoint + "/" + bucket + fileName, nil
}
