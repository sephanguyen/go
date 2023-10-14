package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/usermgmt/example-modules/error-handling/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/example-modules/error-handling/core/errcode"
	"github.com/manabie-com/backend/internal/usermgmt/example-modules/error-handling/core/service"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type mockUser struct {
	userID string
	email  string
}

func (mockUser mockUser) UserID() field.String {
	return field.NewString(mockUser.userID)
}
func (mockUser mockUser) Email() field.String {
	return field.NewString(mockUser.email)
}

type mockUserRepo struct {
	upsertUsersFn func(ctx context.Context, db database.QueryExecer, users entity.Users) error
	getUserFn     func(ctx context.Context, db database.QueryExecer, userIDs field.Strings) (entity.Users, error)
}

func (mockRepo mockUserRepo) UpsertUsers(ctx context.Context, db database.QueryExecer, users entity.Users) error {
	return mockRepo.upsertUsersFn(ctx, db, users)
}
func (mockRepo mockUserRepo) GetUsers(ctx context.Context, db database.QueryExecer, userIDs field.Strings) (entity.Users, error) {
	return mockRepo.getUserFn(ctx, db, userIDs)
}

// Integration tests must assert responded error and messages like assertions in this example
func TestUserService_GetUser(t *testing.T) {
	type testCase struct {
		name                   string
		initService            func() *UserService
		request                string
		expectedHttpStatusCode int
		initExpectedResponse   func() Response
	}

	existingUser := mockUser{
		userID: "existing-user-id",
		email:  "existing-email",
	}

	testCases := []testCase{
		{
			name: "happy case",
			initService: func() *UserService {
				userRepo := mockUserRepo{
					getUserFn: func(ctx context.Context, db database.QueryExecer, userIDs field.Strings) (entity.Users, error) {
						return entity.Users{existingUser}, nil
					},
				}
				userService := &UserService{
					UserDomainService: service.User{
						UserRepo: userRepo,
					},
				}
				return userService
			},
			request:                `{"user_ids":["existing-user-id"]}`,
			expectedHttpStatusCode: http.StatusOK,
			initExpectedResponse: func() Response {
				return Response{
					Code:    errcode.DomainCodeOK,
					Message: "success",
					Data:    []interface{}{map[string]interface{}{}},
				}
			},
		},
		{
			name: "get users with non-existing user id",
			initService: func() *UserService {
				userRepo := mockUserRepo{
					getUserFn: func(ctx context.Context, db database.QueryExecer, userIDs field.Strings) (entity.Users, error) {
						return entity.Users{existingUser}, nil
					},
				}
				userService := &UserService{
					UserDomainService: service.User{
						UserRepo: userRepo,
					},
				}
				return userService
			},
			request:                `{"user_ids":["existing-user-id","non-existing-user-id"]}`,
			expectedHttpStatusCode: http.StatusNotFound,
			initExpectedResponse: func() Response {
				expectedDomainError := entity.NotFoundError{
					EntityName:        "user",
					Index:             1,
					SearchedFieldName: string(entity.UserFieldUserID),
				}
				return Response{
					Code:    errcode.DomainCodeNotFound,
					Message: expectedDomainError.DomainError(),
					Data:    nil,
				}
			},
		},
	}

	t.Parallel()
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			// Setup test server
			router := gin.Default()
			router.GET("/user", testCase.initService().GetUser)
			setupTestHttpServer(t, ctx, router)

			// Send mock request to test server
			httpRecorder := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/user", bytes.NewReader([]byte(testCase.request)))
			router.ServeHTTP(httpRecorder, req)

			// Assertion
			assert.Equal(t, testCase.expectedHttpStatusCode, httpRecorder.Code)

			responseBody := new(Response)
			assert.NoError(t, json.NewDecoder(httpRecorder.Body).Decode(responseBody))
			assert.Equal(t, *responseBody, testCase.initExpectedResponse())
		})
	}
}

func setupTestHttpServer(t *testing.T, ctx context.Context, router *gin.Engine) func() {
	gin.SetMode(gin.TestMode)

	srv := &http.Server{
		Addr:    ":1234",
		Handler: router,
	}

	result := make(chan error, 1)
	go func() {
		result <- srv.ListenAndServe()
	}()

	select {
	case err := <-result:
		t.Fatal(err)
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	default:
		break
	}

	return func() {
		_ = srv.Shutdown(ctx)
	}
}
