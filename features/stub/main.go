package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/bob/services"
	"github.com/manabie-com/backend/internal/bob/services/classes"
	"github.com/manabie-com/backend/internal/golibs/brightcove"
	"github.com/manabie-com/backend/internal/golibs/configs"
	"github.com/manabie-com/backend/internal/golibs/whiteboard"

	"github.com/awa/go-iap/appstore"
)

func main() {
	startMockServer(&configs.BrightcoveConfig{
		AccountID:           "account-id",
		ClientID:            "client-id",
		Profile:             "multi-platform-standard-static",
		Secret:              "secret",
		PolicyKey:           "policy_key",
		PolicyKeyWithSearch: "policy_key_with_search",
	})
}

func startMockServer(c *configs.BrightcoveConfig) {
	stub := &stubBrightcoverAuth{
		cfg: c,
	}
	http.HandleFunc("/verifyReceipt", verifyReceipt)
	http.HandleFunc("/api/v1/apiv3/CreateOrder", ghnCreateOrder)
	http.HandleFunc("/api/v1/apiv3/FindAvailableServices", ghnFindAvailableServices)
	http.HandleFunc("/v4/access_token", stub.getOAuthToken)
	http.HandleFunc("/v1/accounts/account-id/videos/video-id/upload-urls/"+url.QueryEscape("manabie.mp4"), uploadUrls)
	http.HandleFunc("/v1/accounts/account-id/videos/video-id/ingest-requests", ingressRequest)
	http.HandleFunc("/v1/accounts/account-id/videos/", createVideo)
	http.HandleFunc("/playback/v1/accounts/account-id/videos/", getVideo)

	http.HandleFunc("/room", verifyWhiteboardToken)
	http.HandleFunc("/v5/rooms", createWhiteboardRoom)
	http.HandleFunc("/v5/tokens/rooms/", generateWhiteboardRoomToken)
	http.HandleFunc("/v2/jobs", createCloudConvertJob)

	http.HandleFunc("/cloud_recording/", handleRecording)
	if err := http.ListenAndServe("0.0.0.0:5889", nil); err != nil {
		log.Fatalf("brightcove mock server failure: %s", err)
	}
}

func verifyWhiteboardToken(w http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(req.Body)
	var createWhiteboardRequest classes.CreateWhiteboardRoomRequest
	err := decoder.Decode(&createWhiteboardRequest)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`
	{
		"code": 200,
		"msg": {
			"room": {
				"id": 10987,
				"name": "111",
				"limit": 0,
				"teamId": 1,
				"adminId": 1,
				"uuid": "5d10677345324c0cb3febd3291e2a607",
				"updatedAt": "2018-08-14T11:19:04.895Z",
				"createdAt": "2018-08-14T11:19:04.895Z"
			},
			"hare": "{\"message\":\"ok\"}",
			"roomToken": "whiteboard token"
		}
	}
	`))
}

func createWhiteboardRoom(w http.ResponseWriter, req *http.Request) {
	response := struct {
		UUID string `json:"uuid"`
	}{
		UUID: strconv.Itoa(rand.Int()),
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func generateWhiteboardRoomToken(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("whiteboard token"))
}

func createCloudConvertJob(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}{
		Data: struct {
			ID string `json:"id"`
		}{
			ID: strconv.Itoa(rand.Int()),
		},
	})
}

func fetchConversionTaskProgress(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)

	taskUUID := strings.TrimPrefix(req.URL.Path, "/v5/services/conversion/tasks/")
	if strings.Contains(taskUUID, "finished") {
		json.NewEncoder(w).Encode(whiteboard.FetchTaskProgressResponse{
			UUID:   taskUUID,
			Status: "Finished",
			Progress: &whiteboard.TaskProgress{
				TotalPageSize:       2,
				ConvertedPercentage: 100,
				ConvertedFileList: []whiteboard.ConvertedFile{
					{
						Width:             1294,
						Height:            920,
						ConversionFileURL: fmt.Sprintf("http://%d", rand.Int()),
					},
					{
						Width:             1024,
						Height:            2048,
						ConversionFileURL: fmt.Sprintf("http://%d", rand.Int()),
					},
				},
			},
		})
	} else {
		json.NewEncoder(w).Encode(whiteboard.FetchTaskProgressResponse{
			UUID:   taskUUID,
			Status: "Waiting",
		})
	}
}

func verifyReceipt(w http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(req.Body)
	var t appstore.IAPRequest
	err := decoder.Decode(&t)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{
    "receipt": {
        "receipt_type": "ProductionSandbox",
        "adam_id": 0,
        "app_item_id": 0,
        "bundle_id": "com.manabie.ios",
        "application_version": "3",
        "download_id": 0,
        "version_external_identifier": 0,
        "receipt_creation_date": "2018-11-13 16:46:31 Etc/GMT",
        "receipt_creation_date_ms": "1542127591000",
        "receipt_creation_date_pst": "2018-11-13 08:46:31 America/Los_Angeles",
        "request_date": "2020-01-07 17:26:01 Etc/GMT",
        "request_date_ms": "1578417961942",
        "request_date_pst": "2020-01-07 09:26:01 America/Los_Angeles",
        "original_purchase_date": "2013-08-01 07:00:00 Etc/GMT",
        "original_purchase_date_ms": "1375340400000",
        "original_purchase_date_pst": "2013-08-01 00:00:00 America/Los_Angeles",
        "original_application_version": "1.0",
        "in_app": [
            {
                "quantity": "1",
                "product_id": "com.manabie.ios.basic",
                "transaction_id": "` + t.ReceiptData + `",
                "original_transaction_id": "1000000472106082",
                "purchase_date": "2018-11-13 16:46:31 Etc/GMT",
                "purchase_date_ms": "` + strconv.Itoa(int(time.Now().Add(30*time.Second).Unix()*1000)) + `",
                "purchase_date_pst": "2018-11-13 08:46:31 America/Los_Angeles",
                "original_purchase_date": "2018-11-13 16:46:31 Etc/GMT",
                "original_purchase_date_ms": "1542127591000",
                "original_purchase_date_pst": "2018-11-13 08:46:31 America/Los_Angeles",
                "is_trial_period": "false"
            }
        ]
    },
    "status": 0,
    "environment": "Sandbox"
}`))
}

func ghnCreateOrder(w http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(req.Body)
	var t services.CreateOrderRequest
	err := decoder.Decode(&t)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	if t.Token == "" ||
		t.PaymentTypeID != 1 ||
		t.Content == "" ||
		t.FromDistrictID == 0 ||
		t.FromWardCode == "" ||
		t.ToDistrictID == 0 ||
		t.ToWardCode == "" ||
		t.ExternalCode == "" ||
		t.ClientContactName == "" ||
		t.ClientContactPhone == "" ||
		t.ClientAddress == "" ||
		t.CustomerName == "" ||
		t.CustomerPhone == "" ||
		t.ShippingAddress == "" ||
		t.CoDAmount == 0 ||
		t.NoteCode == "" ||
		t.ServiceID == 0 ||
		t.Weight != 1 ||
		t.Length != 1 ||
		t.Width != 1 ||
		t.Height != 1 ||
		t.ReturnContactName == "" ||
		t.ReturnContactPhone == "" ||
		t.ReturnAddress == "" ||
		t.ReturnDistrictID == 0 ||
		t.ExternalReturnCode == "" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{
	"code": 0,
	"msg": "[GHN-ERR80] Giá trị truyền vào không hợp lệ",
	"data": null
}`))
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`
{
    "code": 1,
    "msg": "Success",
    "data": {
        "ErrorMessage": "",
        "OrderID": 268916,
        "PaymentTypeID": 4,
        "OrderCode": "` + strconv.Itoa(rand.Int()) + `",
        "ExtraFee": 0,
        "TotalServiceFee": 81400,
        "ExpectedDeliveryTime": "2017-09-22T23:00:00+07:00",
        "ClientHubID": 0,
        "SortCode": "N/A"
    }
}`))
}

func ghnFindAvailableServices(w http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(req.Body)
	var t services.FindAvailableServicesRequest
	err := decoder.Decode(&t)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	if t.Token == "" ||
		t.FromDistrictID == 0 ||
		t.ToDistrictID == 0 {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{
	"code": 0,
	"msg": "[GHN-ERR81] Giá trị truyền vào không hợp lệ",
	"data": null
}`))
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`
{
    "code": 1,
    "msg": "Success",
    "data": [
        {
            "ExpectedDeliveryTime": "2017-10-06T23:00:00+07:00",
            "Extras": [
                {
                    "MaxValue": 0,
                    "Name": "Khai Giá Hàng Hoá",
                    "ServiceFee": 0,
                    "ServiceID": 16
                },
                {
                    "MaxValue": 0,
                    "Name": "SMS báo phát",
                    "ServiceFee": 550,
                    "ServiceID": 53331
                },
                {
                    "MaxValue": 0,
                    "Name": "Phí thu hộ",
                    "ServiceFee": 0,
                    "ServiceID": 100012
                },
                {
                    "MaxValue": 0,
                    "Name": "Gửi hàng tại điểm",
                    "ServiceFee": 0,
                    "ServiceID": 53337
                }
            ],
            "Name": "Nhanh",
            "ServiceFee": 31900,
            "ServiceID": 53319
        },
        {
            "ExpectedDeliveryTime": "2017-10-07T19:00:00+07:00",
            "Extras": [
                {
                    "MaxValue": 0,
                    "Name": "SMS báo phát",
                    "ServiceFee": 550,
                    "ServiceID": 53331
                },
                {
                    "MaxValue": 0,
                    "Name": "Gửi hàng tại điểm",
                    "ServiceFee": 0,
                    "ServiceID": 53337
                },
                {
                    "MaxValue": 0,
                    "Name": "Khai Giá Hàng Hoá",
                    "ServiceFee": 0,
                    "ServiceID": 16
                }
            ],
            "Name": "Chuẩn",
            "ServiceFee": 20900,
            "ServiceID": 53320
        }
    ]
}`))
}

type stubBrightcoverAuth struct {
	cfg *configs.BrightcoveConfig
}

func (s *stubBrightcoverAuth) getOAuthToken(w http.ResponseWriter, req *http.Request) {
	userName, password, ok := req.BasicAuth()
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`missing basic authen`))
		return
	}

	if !(userName == s.cfg.ClientID && password == s.cfg.Secret) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`wrong clientID or password`))
		return
	}

	resp := &brightcove.OAuthResponse{
		AccessToken: "access-token",
		TokenType:   "bearer",
		ExpiresIn:   300,
	}

	data, _ := json.Marshal(resp)
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func createVideo(w http.ResponseWriter, req *http.Request) {
	if req.Header.Get("authorization") != "Bearer access-token" {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`wrong access_token`))
		return
	}

	reqData := new(brightcove.CreateVideoRequest)
	_ = json.NewDecoder(req.Body).Decode(reqData)

	if reqData.Name == "" {
		w.WriteHeader(422)
		w.Write([]byte(`[ {
		  "error_code" : "VALIDATION_ERROR",
		  "message" : "name: REQUIRED_FIELD"
		} ]`))
		return
	}

	if reqData.Name != "manabie.mp4" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("wrong name"))
		return
	}

	resp := &brightcove.CreateVideoResponse{
		ID: "video-id",
	}

	data, _ := json.Marshal(resp)
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func ingressRequest(w http.ResponseWriter, req *http.Request) {
	if req.Header.Get("authorization") != "Bearer access-token" {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`wrong access_token`))
		return
	}

	reqData := new(brightcove.SubmitDynamicIngressRequest)
	err := json.NewDecoder(req.Body).Decode(reqData)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	if reqData.Master.URL != "api-request-url" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`missing api_request_url`))
		return
	}

	if reqData.Profile != "multi-platform-standard-static" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`profile: must be multi-platform-standard-static`))
		return
	}

	resp := &brightcove.SubmitDynamicIngressResponse{
		JobID: "job-id",
	}

	data, _ := json.Marshal(resp)
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func uploadUrls(w http.ResponseWriter, req *http.Request) {
	if req.Header.Get("authorization") != "Bearer access-token" {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`wrong access_token`))
		return
	}

	resp := &brightcove.UploadUrlsResponse{
		SignedURL:     "signed-url",
		APIRequestURL: "api-request-url",
	}

	data, _ := json.Marshal(resp)
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func getVideo(rw http.ResponseWriter, r *http.Request) {
	if r.Header.Get("authorization") != "BCOV-Policy policy_key_with_search" {
		rw.WriteHeader(http.StatusUnauthorized)
		rw.Write([]byte(`[
{
"error_code": "INVALID_POLICY_KEY",
"message": "Request policy key is missing or invalid 1"
}
]`))
		return
	}

	switch r.URL.String() {
	case fmt.Sprintf("/playback/v1/accounts/account-id/videos/video-id"):
		rw.WriteHeader(http.StatusOK)
		rw.Write([]byte(`{
"thumbnail": "https://link/to/some/image.jpg",
"name": "video-name",
"duration": 1234,
"offline_enabled": true,
"id": "video-id"
}`))
	case fmt.Sprintf("/playback/v1/accounts/account-id/videos/invalid-video-id"):
		rw.WriteHeader(http.StatusNotFound)
		rw.Write([]byte(`[
{
  "error_code": "VIDEO_NOT_FOUND",
  "message": "The designated resource was not found."
}
]`))
	case fmt.Sprintf("/playback/v1/accounts/account-id/videos/video_not_playable"):
		rw.WriteHeader(http.StatusForbidden)
		rw.Write([]byte(`[
{
"error_code": "VIDEO_NOT_PLAYABLE",
"message": "The policy key provided does not permit this account or video, or the requested resource is inactive."
}
]`))
	default:
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte(`unexpected video ID for testing, must be "video-id" or "invalid-video-id"`))
	}
}
