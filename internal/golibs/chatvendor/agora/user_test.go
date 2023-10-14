package agora

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/chatvendor/agora/dto"
	"github.com/manabie-com/backend/internal/golibs/chatvendor/agora/entities"
	abstract_dto "github.com/manabie-com/backend/internal/golibs/chatvendor/dto"

	"github.com/stretchr/testify/assert"
)

func Test_GetUser(t *testing.T) {
	t.Run("happy case", func(t *testing.T) {
		t.Parallel()
		userID := "example-username"
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(dto.GetUserResponse{
				Duration:  1,
				Timestamp: 1,
				Count:     1,
				Action:    "get",
				Path:      "/users",
				URI:       "https://example.com",
				Entities: []entities.User{
					{
						NickName:  "example-nickname",
						UUID:      "example-uuid",
						UserName:  "example-username",
						Type:      "user",
						Activated: true,
						Created:   1,
						Modified:  1,
					},
				},
			})
		}))
		defer ts.Close()

		agoraClient := newAgoraClientForUnitTest(ts.URL)
		user, err := agoraClient.GetUser(&abstract_dto.GetUserRequest{
			VendorUserID: userID,
		})
		assert.Equal(t, nil, err)
		assert.NotNil(t, user)
	})
}

func Test_CreateUser(t *testing.T) {
	t.Run("[real] happy case", func(t *testing.T) {
		t.Parallel()
		createUserID := "example-username"
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			req := &dto.CreateUserRequest{}
			err := json.NewDecoder(r.Body).Decode(&req)
			if err == nil && req.UserID == createUserID {
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(dto.CreateUserResponse{
					Duration:        1,
					Timestamp:       1,
					Application:     "app-test",
					ApplicationName: "app-test-name",
					Organization:    "org-test",
					Action:          "get",
					Path:            "/users",
					URI:             "https://example.com",
					Entities: []entities.User{
						{
							NickName:  "example-nickname",
							UUID:      "example-uuid",
							UserName:  "example-username",
							Type:      "user",
							Activated: true,
							Created:   1,
							Modified:  1,
						},
					},
				})
			} else {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(dto.ErrorResponse{
					Duration:         1,
					Timestamp:        1,
					Exception:        "failed",
					Error:            "bad_request",
					ErrorDescription: "bad_request",
				})
			}
		}))
		defer ts.Close()

		agoraClient := newAgoraClientForUnitTest(ts.URL)
		hash := md5.Sum([]byte(createUserID))
		agoraUserID := strings.ToLower(hex.EncodeToString(hash[:]))
		user, err := agoraClient.CreateUser(&abstract_dto.CreateUserRequest{
			UserID:       createUserID,
			VendorUserID: agoraUserID,
		})
		assert.Equal(t, nil, err)
		assert.NotNil(t, user)
	})
}
