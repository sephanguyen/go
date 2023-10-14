package eureka

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	bob_entities "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/eureka/configurations"
	"github.com/manabie-com/backend/internal/eureka/constants"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/eureka/services"
	"github.com/manabie-com/backend/internal/golibs/bootstrap"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/Masterminds/log-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var tenantID, url, token string

func init() {
	bootstrap.RegisterJob("fix_student_event_logs_data", fixStudentEventLogsData).
		StringVar(&tenantID, "tenantID", "", "specify tenantID").
		Desc("eureka fix student event logs data").
		StringVar(&url, "url", "", "specify url").
		Desc("url of data csv file").
		StringVar(&token, "token", "", "specify token").
		Desc("token of data csv file")
}

func convertPayloadToRequest(payloadStr string) *epb.CreateStudentEventLogsRequest {
	var parsedLog struct {
		Req        epb.CreateStudentEventLogsRequest `json:"req"`
		Err        string                            `json:"err"`
		AppVersion string                            `json:"app_version"`
	}

	if err := json.Unmarshal([]byte(payloadStr), &parsedLog); err != nil {
		fmt.Println("Error parsing log:", err)
		return nil
	}

	filteredLogs := make([]*epb.StudentEventLog, 0)

	for _, log := range parsedLog.Req.StudentEventLogs {
		if log.Payload.SessionId != "" {
			filteredLogs = append(filteredLogs, log)
		}
	}

	// Update the student event logs with the filtered logs
	parsedLog.Req.StudentEventLogs = filteredLogs

	return &parsedLog.Req
}

func ReadCSVData(url, token string) map[string][]*epb.CreateStudentEventLogsRequest {
	// Create an HTTP client with authentication header
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Error creating HTTP request:", err)
		return nil
	}
	req.Header.Set("Authorization", "token "+token)

	// Send the HTTP request
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending HTTP request:", err)
		return nil
	}
	defer resp.Body.Close()

	// Check if the response was successful
	if resp.StatusCode != http.StatusOK {
		fmt.Println("Error:", resp.Status)
		return nil
	}

	// Read the CSV data
	reader := csv.NewReader(resp.Body)

	// Read the CSV headers
	headers, err := reader.Read()
	if err != nil {
		fmt.Printf("Failed to read CSV headers: %v", err)
		// file.Close()
		return nil
	}

	// Find the indices of the "payload" and "user_id" columns
	payloadIndex := -1
	userIDIndex := -1
	for i, header := range headers {
		switch header {
		case "payload":
			payloadIndex = i
		case "user_id":
			userIDIndex = i
		}
	}

	// Check if the required columns were found
	if payloadIndex == -1 || userIDIndex == -1 {
		fmt.Print("Missing required columns in the CSV file")
		return nil
	}

	// Initialize the arrays to store the parsed data
	mapUserPayload := make(map[string][]string, 0)

	// Read and process the CSV data
	for {
		// Read each row from the CSV file
		row, err := reader.Read()
		if err != nil {
			// Check for end of file
			if err.Error() == "EOF" {
				break
			}
			fmt.Printf("Failed to read CSV row: %v", err)
		}

		// Extract the payload and user_id values from the row
		mapUserPayload[row[userIDIndex]] = append(mapUserPayload[row[userIDIndex]], row[payloadIndex])
	}

	// Print the parsed data
	mapUserRequest := make(map[string][]*epb.CreateStudentEventLogsRequest, 0)
	for userID, payload := range mapUserPayload {
		for _, payloadStr := range payload {
			req := convertPayloadToRequest(payloadStr)
			if req != nil {
				mapUserRequest[userID] = append(mapUserRequest[userID], req)
			}
		}
	}

	return mapUserRequest
}

func fixStudentEventLogsData(_ context.Context, _ configurations.Config, rsc *bootstrap.Resources) error {
	db := rsc.DBWith("eureka")
	start := time.Now()
	defer func() {
		log.Infof("fixing student event logs data completed: %v\n", time.Since(start))
	}()

	log.Info("======= set up services and database\n")

	studentEventLogModifierService := &services.StudentEventLogModifierService{
		DB:                        db,
		StudentEventLogRepo:       &repositories.StudentEventLogRepo{},
		StudyPlanItemRepo:         &repositories.StudyPlanItemRepo{},
		StudentLOCompletenessRepo: &repositories.StudentsLearningObjectivesCompletenessRepo{},
		LearningTimeCalculator: &services.LearningTimeCalculator{
			DB:                           db,
			StudentEventLogRepo:          &repositories.StudentEventLogRepo{},
			StudentLearningTimeDailyRepo: &repositories.StudentLearningTimeDailyRepo{},
			UsermgmtUserReaderService:    &FakeUserReaderServiceClient{},
		},
	}

	// var fileName string
	log.Infof("======= read csv with org %v url %v\n", tenantID, url)

	if tenantID == "" || url == "" {
		return errors.New("missing tenantID or url")
	}

	mapUserReq := ReadCSVData(url, token)
	countTotal := 1
	for userID, reqList := range mapUserReq {
		ctx := injectContext(context.Background(), tenantID, userID)
		countUser := 1
		for _, req := range reqList {
			_, err := studentEventLogModifierService.CreateStudentEventLogs(ctx, req)
			if err != nil {
				log.Infof("======= err: %v\n", err)
				log.Infof("======= err from userID %v request payload %v\n", userID, req)
				return err
			}
			log.Infof("======= CreateStudentEventLogs successful with userID %v count %v\n", userID, countUser)
			log.Infof("======= CreateStudentEventLogs successful total count %v\n", countTotal)
			countUser++
			countTotal++
		}
	}

	return nil
}

func injectContext(ctx context.Context, org, userID string) context.Context {
	ctx = interceptors.ContextWithUserID(ctx, userID)
	ctx = injectFakeJwtToken(ctx, org, userID)
	ctx = metadata.NewIncomingContext(ctx, metadata.MD{
		"version": []string{"1.0.0"},
		"pkg":     []string{"com.manabie.liz"},
		"token":   []string{"token"},
	})

	return ctx
}

type UserReaderServiceClient interface {
	upb.UnimplementedUserReaderServiceServer
	SearchBasicProfile(ctx context.Context, in *upb.SearchBasicProfileRequest, opts ...grpc.CallOption) (*upb.SearchBasicProfileResponse, error)
	RetrieveStudentAssociatedToParentAccount(ctx context.Context, in *upb.RetrieveStudentAssociatedToParentAccountRequest, opts ...grpc.CallOption) (*upb.RetrieveStudentAssociatedToParentAccountResponse, error)
	GetBasicProfile(ctx context.Context, in *upb.GetBasicProfileRequest, opts ...grpc.CallOption) (*upb.GetBasicProfileResponse, error)
}

type FakeUserReaderServiceClient struct{}

//nolint
func (s *FakeUserReaderServiceClient) SearchBasicProfile(_ context.Context, _ *upb.SearchBasicProfileRequest, _ ...grpc.CallOption) (*upb.SearchBasicProfileResponse, error) {
	return &upb.SearchBasicProfileResponse{
		Profiles: []*cpb.BasicProfile{
			{
				Country: cpb.Country_COUNTRY_JP,
			},
		},
	}, nil
}

func injectFakeJwtToken(ctx context.Context, resourcePath, userID string) context.Context {
	claim := &interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{
			DefaultRole:  constants.RoleStudent,
			UserGroup:    bob_entities.UserGroupStudent,
			ResourcePath: resourcePath,
			UserID:       userID,
		},
	}

	return interceptors.ContextWithJWTClaims(ctx, claim)
}
