package generateresumableuploadurl_test

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/manabie-com/backend/examples"
	"github.com/manabie-com/backend/internal/bob/services/filestore"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/configs"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	pb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/durationpb"
)

func Test_GenerateResumableUploadURL(t *testing.T) {
	token, err := os.ReadFile("./token.jwt")
	if err != nil && os.IsNotExist(err) {
		t.Skip("please add JWT token to token.jwt file in this folder")
	}

	conn := examples.SimplifiedDial(examples.Stag, false)
	ctx := examples.AuthorizedContext(context.Background(), string(token))
	resp, err := pb.NewUploadServiceClient(conn).GenerateResumableUploadURL(ctx,
		&pb.ResumableUploadURLRequest{
			PrefixName:    "example_img",
			FileExtension: "jpg",
			Expiry:        durationpb.New(time.Hour),
		})

	if err != nil {
		t.Log(err.Error())
		t.Fatal(err)
	}

	t.Log(resp)
}

func Test_GenerateResumableUploadURLByPrivateKey(t *testing.T) {
	const serviceAccountEmail = "stag-bob@staging-manabie-online.iam.gserviceaccount.com"
	objectName := "image-" + idutil.ULIDNow() + ".png"
	allowOrigin := "http://localhost:3001"           // or you maybe use other allowOrigin if you want
	contentType := golibs.GetContentType(objectName) // or you maybe use other contentType if you want
	log.Println("contentType:", contentType)
	expiry := 10 * time.Minute

	conf := &configs.StorageConfig{
		Endpoint: "https://storage.googleapis.com",
		Region:   "asia",
		Bucket:   "stag-manabie-backend",
	}

	s, err := filestore.NewGoogleCloudStorageWithoutInitIAMCredential(serviceAccountEmail, conf)
	require.NoError(t, err)

	// privateKey used to sign url
	privateKey, err := os.ReadFile("./key.pem")
	require.NoError(t, err)

	u, err := s.GenerateResumableObjectURLWithPrivateKey(
		context.Background(),
		objectName,
		expiry,
		allowOrigin,
		contentType,
		privateKey,
	)
	require.NoError(t, err)

	log.Println("sessionURI:", u.String())
	log.Println("download url:", s.GeneratePublicObjectURL(objectName))
}
