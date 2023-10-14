package lessonrecording

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/agoratokenbuilder"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Recorder manages cloud recording
type Recorder struct {
	*http.Client
	*zap.Logger
	Channel string
	Token   string
	UID     int
	RID     string
	SID     string
	Configs Config
	Message string
}

type Config struct {
	AppID           string
	Cert            string
	CustomerID      string
	CustomerSecret  string
	BucketID        string
	BucketAccessKey string
	BucketSecretKey string
	Endpoint        string
	MaxIdleTime     int
}

const UIDFormat = "%09d"

func getIDByResourcePath(orgMap map[string]string, resourcePath string) string {
	for key, value := range orgMap {
		if value == resourcePath {
			return key
		}
	}
	return ""
}

func NewRecorder(ctx context.Context, cfg Config, logger *zap.Logger, channel string, orgMap map[string]string) (*Recorder, error) {
	num, err := rand.Int(rand.Reader, big.NewInt(999999))
	if err != nil {
		return nil, err
	} // random uint32
	id := getIDByResourcePath(orgMap, golibs.ResourcePathFromCtx(ctx))
	if id == "" {
		return nil, fmt.Errorf("new recorder: resource_path %s do not match any organization in init list %s", golibs.ResourcePathFromCtx(ctx), orgMap)
	}
	uIDStr := fmt.Sprintf("%d%s", int(num.Int64())+1, id)
	uID, err := strconv.Atoi(uIDStr)
	if err != nil {
		return nil, fmt.Errorf("fail when parse uIdStr %s to uId int", uIDStr)
	}
	return &Recorder{
		Channel: channel,
		UID:     uID,
		Configs: cfg,
		Client: &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				Dial: (&net.Dialer{
					Timeout: 5 * time.Second,
				}).Dial,
				MaxIdleConnsPerHost: 5,
			},
		},
		Logger: logger,
	}, nil
}

func GetExistingRecorder(ctx context.Context, cfg Config, logger *zap.Logger, uID int, channel, rID, sID string) *Recorder {
	return &Recorder{
		Channel: channel,
		UID:     uID,
		RID:     rID,
		SID:     sID,
		Configs: cfg,
		Client: &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				Dial: (&net.Dialer{
					Timeout: 5 * time.Second,
				}).Dial,
				MaxIdleConnsPerHost: 5,
			},
		},
		Logger: logger,
	}
}

// Acquire runs the acquire endpoint for Cloud Recording
func (rec *Recorder) Acquire() (string, error) {
	creds, err := GenerateUserCredentials(rec.Configs.AppID, rec.Configs.Cert, rec.Channel, rec.UID)
	if err != nil {
		return "", err
	}

	rec.UID = creds.UID
	rec.Token = creds.Rtc
	newUID := fmt.Sprintf(UIDFormat, rec.UID)
	requestBody := fmt.Sprintf(`
		{
			"cname": "%s",
			"uid": "%s",
			"clientRequest": {
				"resourceExpiredHour": 24
			}
		}
	`, rec.Channel, newUID)
	url := rec.Configs.Endpoint + "/v1/apps/" + rec.Configs.AppID + "/cloud_recording/acquire"
	req, err := http.NewRequest("POST", url,
		bytes.NewBuffer([]byte(requestBody)))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(rec.Configs.CustomerID, rec.Configs.CustomerSecret)

	resp, err := rec.Do(req)
	rec.Logger.Info("Call Acquire Recording API", zap.String("url", url), zap.String("body", requestBody))

	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	var result map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	rec.RID = result["resourceId"]
	if len(rec.RID) == 0 {
		return "", fmt.Errorf(result["message"])
	}
	b, err := json.Marshal(result)
	rec.Logger.Info("Result Acquire Recording API", zap.String("result", string(b)))
	return string(b), err
}

// GenerateUserCredentials generates uid, rtc and rtc token
func GenerateUserCredentials(appID, cert, channel string, uid int) (*UserCredentials, error) {
	expireTimestamp := uint32(time.Now().UTC().Unix() + 21600)
	rtmToken, err := agoratokenbuilder.BuildStreamToken(appID, cert,
		channel, strconv.Itoa(uid), agoratokenbuilder.RolePublisher,
		expireTimestamp)
	if err != nil {
		return nil, status.Error(codes.Internal, "agoratokenbuilder.BuildRTMToken: could not generate RTM token: "+err.Error())
	}
	return &UserCredentials{
		Rtc: rtmToken,
		UID: uid,
	}, nil
}

func parseSliceStringToJSON(s []string, inputName string) (string, error) {
	if len(s) > 0 {
		j, err := json.Marshal(s)
		if err != nil {
			return "", fmt.Errorf("error when json.Marshal %s: %s", inputName, err)
		}
		return string(j), nil
	}
	return "[]", nil
}

// Start starts the recording
func (rec *Recorder) Start(sc *StartCall) (string, error) {
	var requestBody string
	vendor := 6 // option to Google Storage
	var subscribeVideoUIDsJSON, subscribeAudioUIDsJSON, fileNamePrefixJSON string
	var err error

	if subscribeVideoUIDsJSON, err = parseSliceStringToJSON(sc.SubscribeVideoUids, "SubscribeVideoUids"); err != nil {
		return "", err
	}

	if subscribeAudioUIDsJSON, err = parseSliceStringToJSON(sc.SubscribeAudioUids, "SubscribeAudioUids"); err != nil {
		return "", err
	}

	if fileNamePrefixJSON, err = parseSliceStringToJSON(sc.FileNamePrefix, "FileNamePrefix"); err != nil {
		return "", err
	}

	newUID := fmt.Sprintf(UIDFormat, rec.UID)

	requestBody = fmt.Sprintf(`
		{
			"cname": "%s",
			"uid": "%s",
			"clientRequest": {
				"token": "%s",
				"recordingConfig": {
					"maxRecordingHour": 12,
					"maxIdleTime": %d,
					"streamTypes": 2,
					"audioProfile": 1,
					"channelType": 0,
					"videoStreamType": 0,
					"transcodingConfig": %s,
					"subscribeVideoUids": %s,
					"subscribeAudioUids": %s
				},
				"recordingFileConfig": {
					"avFileType": [
						"hls",
						"mp4"
					]
				},
				"storageConfig": {
					"vendor": %d,
					"region": 1,
					"bucket": "%s",
					"accessKey": "%s",
					"secretKey": "%s",
					"fileNamePrefix": %s
				}
			}
		}
	`, rec.Channel, newUID, rec.Token, rec.Configs.MaxIdleTime, sc.TranscodingConfigJSON, subscribeVideoUIDsJSON, subscribeAudioUIDsJSON, vendor, rec.Configs.BucketID,
		rec.Configs.BucketAccessKey, rec.Configs.BucketSecretKey, fileNamePrefixJSON)

	url := rec.Configs.Endpoint + "/v1/apps/" + rec.Configs.AppID + "/cloud_recording/resourceid/" + rec.RID + "/mode/mix/start"

	req, err := http.NewRequest("POST", url,
		bytes.NewBuffer([]byte(requestBody)))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(rec.Configs.CustomerID, rec.Configs.CustomerSecret)

	rec.Logger.Info("Call Start Recording API", zap.String("url", url), zap.String("body", requestBody))
	resp, err := rec.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	rec.SID = result["sid"].(string)
	b, err := json.Marshal(result)
	if len(rec.SID) == 0 {
		return "", fmt.Errorf(string(b))
	}
	rec.Logger.Info("Result Start Recording API", zap.String("result", string(b)))
	return string(b), err
}

func (rec *Recorder) CallStatusAPI() (*Status, error) {
	url := rec.Configs.Endpoint + "/v1/apps/" + rec.Configs.AppID + "/cloud_recording/resourceid/" + rec.RID + "/sid/" + rec.SID + "/mode/mix/query"
	rec.Logger.Info("Call Query Status Recording API", zap.String("url", url))
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(rec.Configs.CustomerID, rec.Configs.CustomerSecret)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	// check is not status 200 and 206
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		return nil, fmt.Errorf("expected status 200 and 206 but got %d when call query status recording", resp.StatusCode)
	}
	defer resp.Body.Close()
	var result Status
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	rec.Logger.Info("Result Query Recording API", zap.Any("result", result))
	return &result, nil
}

// Stop stops the cloud recording
func (rec *Recorder) Stop() (*Status, error) {
	newUID := fmt.Sprintf(UIDFormat, rec.UID)
	requestBody := fmt.Sprintf(`
		{
			"cname": "%s",
			"uid": "%s",
			"clientRequest": {
			}
		}
	`, rec.Channel, newUID)

	url := rec.Configs.Endpoint + "/v1/apps/" + rec.Configs.AppID + "/cloud_recording/resourceid/" + rec.RID + "/sid/" + rec.SID + "/mode/mix/stop"
	rec.Logger.Info("Call Stop Status Recording API", zap.String("url", url), zap.String("body", requestBody))

	req, err := http.NewRequest("POST", url,
		bytes.NewBuffer([]byte(requestBody)))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(rec.Configs.CustomerID, rec.Configs.CustomerSecret)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("expected status 200 but got %d when call stop recording", resp.StatusCode)
	}
	var result Status
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	rec.Logger.Info("Result Stop Recording API", zap.Any("result", result))
	return &result, nil
}
