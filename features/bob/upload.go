package bob

// upload.feature
// Feature: Upload file to s3
//   In order to upload avatar or message chat image
//   As a user
//   I need to perform upload file to s3

//   Scenario: user upload "avatar" image
//     Given a signed in student
//       And one "avatar" image
//     When upload "valid" chunk size and "valid" file size
//     Then url must be contain "bucket/avatar/dir"
//       And bob must store image in s3

//   Scenario: user upload "chat" image
//     Given a signed in student
//       And one "chat" image
//     When upload "valid" chunk size and "valid" file size
//     Then url must be contain "bucket/chat/dir"
//       And bob must store image in s3

//   Scenario: user upload "chat" image
//     Given a signed in student
//     And one "assignment" image
//     When upload "valid" chunk size and "valid" file size
//     Then url must be contain "bucket/assignment/dir"
//     And bob must store image in s3

//   Scenario: user upload "chat" image over chunk size
//     Given a signed in student
//       And one "chat" image
//     When upload "invalid" chunk size and "valid" file size
//     Then returns "InvalidArgument" status code

//   Scenario: user upload "chat" image over file size
//     Given a signed in student
//       And one "chat" image
//     When upload "valid" chunk size and "invalid" file size
//     Then returns "InvalidArgument" status code
// upload.go
// import (
// 	"bytes"
// 	"context"
// 	"crypto/rand"
// 	"errors"
// 	"fmt"
// 	"image"
// 	"image/color"
// 	"image/draw"
// 	"image/jpeg"
// 	"io/ioutil"
// 	"net/http"
// 	"strings"

// 	"code.cloudfoundry.org/bytefmt"

// 	pb "github.com/manabie-com/backend/pkg/genproto/bob"
// )

// func (s *suite) oneImage(ctx context.Context, arg1 string) (context.Context, error) {
// 	stepState := StepStateFromContext(ctx)
// 	if arg1 == "avatar" {
// 		stepState.Request = map[string]interface{}{
// 			"type": pb.UPLOAD_TYPE_AVATAR,
// 		}
// 	} else if arg1 == "chat" {
// 		stepState.Request = map[string]interface{}{
// 			"type": pb.UPLOAD_TYPE_CHAT,
// 		}
// 	} else if arg1 == "assignment" {
// 		stepState.Request = map[string]interface{}{
// 			"type": pb.UPLOAD_TYPE_ASSIGNMENT,
// 		}
// 	}
// 	return StepStateToContext(ctx, stepState), nil
// }
// func createImage(width int, height int, background color.RGBA) *image.RGBA {
// 	rect := image.Rect(0, 0, width, height)
// 	img := image.NewRGBA(rect)
// 	draw.Draw(img, img.Bounds(), &image.Uniform{background}, image.Point{}, draw.Src)
// 	return img
// }
// func (s *suite) uploadChunkSizeAndFileSize(ctx context.Context, chunkSizeCfg, fileSizeCfg string) (context.Context, error) {
// 	stepState := StepStateFromContext(ctx)
// 	req := stepState.Request.(map[string]interface{})

// 	// gRPC server
// 	stream, _ := pb.NewUploadServiceClient(s.Conn).Upload(s.signedCtx(ctx))

// 	var file bytes.Buffer
// 	chunksize, _ := bytefmt.ToBytes("500kb")
// 	filesize, _ := bytefmt.ToBytes("5mb")

// 	if chunkSizeCfg == "invalid" {
// 		chunksize, _ = bytefmt.ToBytes("2mb")
// 	}

// 	if fileSizeCfg == "invalid" {
// 		filesize, _ = bytefmt.ToBytes("12mb")
// 	}

// 	if chunkSizeCfg == "valid" && fileSizeCfg == "valid" { //create an image to check content type is jpeg
// 		img := createImage(200, 200, color.RGBA{0, 0, 0, 0})

// 		var b bytes.Buffer
// 		err := jpeg.Encode(&b, img, nil)
// 		if err != nil {
// 			return StepStateToContext(ctx, stepState), err

// 		}

// 		file.Write(b.Bytes())
// 		filesize = uint64(b.Len())
// 	} else {
// 		//make a file with size
// 		data := make([]byte, filesize)
// 		rand.Read(data)
// 		file.Write(data)
// 	}

// 	req["size"] = filesize

// 	buf := make([]byte, chunksize)

// 	uploadType := req["type"].(pb.UploadType)

// 	for {
// 		n, err := file.Read(buf)
// 		if err != nil {
// 			break
// 		}

// 		streamErr := stream.Send(&pb.UploadRequest{
// 			UploadType: uploadType,
// 			Payload:    buf[:n],
// 			Extension:  "png",
// 		})
// 		if streamErr != nil {
// 			break
// 		}
// 	}

// 	stepState.Response, stepState.ResponseErr = stream.CloseAndRecv()
// 	return StepStateToContext(ctx, stepState), nil
// }
// func (s *suite) urlMustBeContain(ctx context.Context, arg1 string) (context.Context, error) {
// 	stepState := StepStateFromContext(ctx)
// 	if stepState.ResponseErr != nil {
// 		return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected error when upload: %w", stepState.ResponseErr)
// 	}

// 	resp := stepState.Response.(*pb.UploadResponse)
// 	if resp == nil {
// 		return StepStateToContext(ctx, stepState), fmt.Errorf("nil response from upload")
// 	}

// 	urlPattern := s.Cfg.Storage.Endpoint + "/" + s.Cfg.Storage.Bucket

// 	switch arg1 {
// 	case "bucket/chat/dir":
// 		if !strings.Contains(resp.Url, urlPattern+"/chat") {
// 			return StepStateToContext(ctx, stepState), errors.New("url does not match the pattern")
// 		}
// 	case "bucket/avatar/dir":
// 		if !strings.Contains(resp.Url, urlPattern+"/avatar") {
// 			return StepStateToContext(ctx, stepState), errors.New("url does not match the pattern")
// 		}
// 	case "bucket/assignment/dir":
// 		if !strings.Contains(resp.Url, urlPattern+"/assignment") {
// 			return StepStateToContext(ctx, stepState), errors.New("url does not match the pattern")
// 		}
// 	}

// 	if !strings.Contains(resp.Url, ".png") {
// 		return StepStateToContext(ctx, stepState), errors.New("url does not match extension")
// 	}

// 	return StepStateToContext(ctx, stepState), nil
// }
// func (s *suite) bobMustStoreImageInS(ctx context.Context, arg1 int) (context.Context, error) {
// 	stepState := StepStateFromContext(ctx)
// 	uploadResp := stepState.Response.(*pb.UploadResponse)
// 	resp, err := http.Get(uploadResp.GetUrl())
// 	if err != nil {
// 		return StepStateToContext(ctx, stepState), nil
// 	}
// 	defer resp.Body.Close()

// 	body, _ := ioutil.ReadAll(resp.Body)

// 	req := stepState.Request.(map[string]interface{})
// 	if uint64(len(body)) != req["size"].(uint64) {
// 		return StepStateToContext(ctx, stepState), fmt.Errorf("invalid file size, expecting %d, got %d", req["size"].(uint64), len(body))
// 	}

// 	if resp.Header.Get("Content-Type") != "image/jpeg" {
// 		return StepStateToContext(ctx, stepState), errors.New("content type must be image/jpeg")
// 	}

// 	return StepStateToContext(ctx, stepState), err
// }
