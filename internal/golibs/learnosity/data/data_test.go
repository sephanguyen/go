package data

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/learnosity"
	"github.com/manabie-com/backend/internal/golibs/learnosity/entity"
	"github.com/manabie-com/backend/internal/golibs/learnosity/http"
	mock_learnosity "github.com/manabie-com/backend/mock/golibs/learnosity"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestRequest(t *testing.T) {
	t.Skip()
	dataAPI := &Client{}

	security := learnosity.Security{
		ConsumerKey:    "yis0TYCu7U9V4o7M",
		Domain:         "localhost",
		Timestamp:      learnosity.FormatUTCTime(time.Now()),
		UserID:         "$ANONYMIZED_USER_ID", // 81b44c76-da57-47ce-8433-aa46b6d62a4d
		ConsumerSecret: "74c5fd430cf1242a527f6223aebd42d30464be22",
	}

	dataRequest := learnosity.Request{
		"activity_id": []string{"itemsactivitiesdemo"},
		"session_id": []string{
			"a9f1fc23-d214-42b6-b17e-6721b2329b1e",
			"72dfc04a-165b-480a-9502-49af351d9a47",
			"6241f36e-3fa5-48b7-b90f-2a5a994d424e",
		},
		"user_id": []string{"$ANONYMIZED_USER_ID"},
	}

	results, err := dataAPI.Request(context.Background(), &http.Client{}, learnosity.EndpointDataAPISessionsResponses, security, dataRequest)
	require.NoError(t, err)

	dataArr := make([]*entity.SessionResponse, 0, int(results.Meta["records"].(float64)))
	err = json.Unmarshal(results.Data, &dataArr)
	require.NoError(t, err)

	for _, data := range dataArr {
		fmt.Printf("results: %v\n", data)
	}
}

func TestClient_Request(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	now := time.Now()

	dataAPI := &Client{}
	mockHTTP := &mock_learnosity.HTTP{}

	testCases := []struct {
		Name           string
		Ctx            context.Context
		HTTP           learnosity.HTTP
		Endpoint       learnosity.Endpoint
		Security       learnosity.Security
		Request        learnosity.Request
		Action         learnosity.Action
		Setup          func(ctx context.Context)
		ExpectedResult *learnosity.Result
		ExpectedError  error
	}{
		{
			Name:     "happy case",
			Ctx:      ctx,
			HTTP:     mockHTTP,
			Endpoint: learnosity.EndpointDataAPISessionsStatuses,
			Security: learnosity.Security{
				ConsumerKey:    "consumer_key",
				Domain:         "localhost",
				Timestamp:      learnosity.FormatUTCTime(now),
				UserID:         "user_id",
				ConsumerSecret: "consumer_secret",
			},
			Request: learnosity.Request{
				"limit": 5,
			},
			Action: learnosity.ActionGet,
			Setup: func(ctx context.Context) {
				mockHTTP.On("Request", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
			},
			ExpectedResult: &learnosity.Result{},
			ExpectedError:  errors.New("error"),
		},
		{
			Name:     "empty endpoint",
			Ctx:      ctx,
			HTTP:     mockHTTP,
			Endpoint: "",
			Security: learnosity.Security{
				ConsumerKey:    "consumer_key",
				Domain:         "localhost",
				Timestamp:      learnosity.FormatUTCTime(now),
				UserID:         "user_id",
				ConsumerSecret: "consumer_secret",
			},
			Request: learnosity.Request{
				"limit": 5,
			},
			Action: learnosity.ActionGet,
			Setup: func(ctx context.Context) {
				mockHTTP.On("Request", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
			},
			ExpectedResult: nil,
			ExpectedError:  errors.New(learnosity.ErrNotFoundEndpoint.Error()),
		},
		{
			Name:     "err HTTP.Request",
			Ctx:      ctx,
			HTTP:     mockHTTP,
			Endpoint: learnosity.EndpointDataAPISessionsStatuses,
			Security: learnosity.Security{
				ConsumerKey:    "consumer_key",
				Domain:         "localhost",
				Timestamp:      learnosity.FormatUTCTime(now),
				UserID:         "user_id",
				ConsumerSecret: "consumer_secret",
			},
			Request: learnosity.Request{
				"limit": 5,
			},
			Action: learnosity.ActionGet,
			Setup: func(ctx context.Context) {
				mockHTTP.On("Request", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(errors.New("error"))
			},
			ExpectedResult: &learnosity.Result{},
			ExpectedError:  fmt.Errorf("HTTP.Request: %w", errors.New("error")),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			tc.Setup(tc.Ctx)
			result, err := dataAPI.Request(tc.Ctx, tc.HTTP, tc.Endpoint, tc.Security, tc.Request, tc.Action)
			if err != nil {
				assert.Equal(t, tc.ExpectedError.Error(), err.Error())
			} else {
				assert.Equal(t, tc.ExpectedResult, result)
			}
		})
	}
}

func TestClient_RequestIterator(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	now := time.Now()

	dataAPI := &Client{}
	mockHTTP := &mock_learnosity.HTTP{}

	testCases := []struct {
		Name           string
		Ctx            context.Context
		HTTP           learnosity.HTTP
		Endpoint       learnosity.Endpoint
		Security       learnosity.Security
		Request        learnosity.Request
		Action         learnosity.Action
		Setup          func(ctx context.Context)
		ExpectedResult []learnosity.Result
		ExpectedError  error
	}{
		{
			Name:     "err server returned empty meta",
			Ctx:      ctx,
			HTTP:     mockHTTP,
			Endpoint: learnosity.EndpointDataAPISessionsStatuses,
			Security: learnosity.Security{
				ConsumerKey:    "consumer_key",
				Domain:         "localhost",
				Timestamp:      learnosity.FormatUTCTime(now),
				UserID:         "user_id",
				ConsumerSecret: "consumer_secret",
			},
			Request: learnosity.Request{
				"limit": 5,
			},
			Action: learnosity.ActionGet,
			Setup: func(ctx context.Context) {
				mockHTTP.On("Request", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
			},
			ExpectedError: errors.New("server returned empty meta: map[]"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			tc.Setup(tc.Ctx)
			result, err := dataAPI.RequestIterator(tc.Ctx, tc.HTTP, tc.Endpoint, tc.Security, tc.Request, tc.Action)
			if err != nil {
				assert.Equal(t, tc.ExpectedError.Error(), err.Error())
			} else {
				assert.Equal(t, tc.ExpectedResult, result)
			}
		})
	}
}

func TestClient_SetItems(t *testing.T) {
	t.Skip()
	dataAPI := &Client{}
	security := learnosity.Security{
		ConsumerKey:    "yis0TYCu7U9V4o7M",
		Domain:         "localhost",
		Timestamp:      learnosity.FormatUTCTime(time.Now()),
		UserID:         "81b44c76-da57-47ce-8433-aa46b6d62a4d",
		ConsumerSecret: "74c5fd430cf1242a527f6223aebd42d30464be22",
	}

	reference := uuid.New()
	fmt.Printf("Generated UUID: %s", reference.String())
	items := []*entity.Item{
		{
			Reference: reference.String(),
			Metadata:  nil,
			Definition: entity.Definition{
				Widgets: []entity.Reference{
					{
						Reference: "d3c9f50e-74a3-40f9-9085-957370941ebb",
					},
					{
						Reference: "98c926a5-13d0-4786-9e34-fd1d3f93bab0",
					},
					{
						Reference: "f1f997f8-061c-48f7-9748-6e02549ab09f",
					},
					{
						Reference: "fb842313-05e3-4b3e-be71-31354ec5d150",
					},
					{
						Reference: "a10b369a-c9e5-4dc5-b90f-5a108bb13dad",
					},
				},
			},
			Status: "published",
			Questions: []entity.Reference{
				{
					Reference: "98c926a5-13d0-4786-9e34-fd1d3f93bab0",
				}, {
					Reference: "f1f997f8-061c-48f7-9748-6e02549ab09f",
				}, {
					Reference: "fb842313-05e3-4b3e-be71-31354ec5d150",
				},
				{
					Reference: "a10b369a-c9e5-4dc5-b90f-5a108bb13dad",
				},
			},
			Features: []entity.Reference{
				{
					Reference: "d3c9f50e-74a3-40f9-9085-957370941ebb",
				},
			},

			Tags: entity.Tags{
				Tenant: []string{"tagtest"},
			},
		},
	}
	results, err := dataAPI.Request(context.Background(),
		&http.Client{},
		learnosity.EndpointDataAPISetItems,
		security,
		learnosity.Request{"items": items},
		learnosity.ActionSet)
	require.NoError(t, err)

	assert.True(t, results.Meta["status"].(bool))
}

func TestClient_SetFeatures(t *testing.T) {
	t.Skip()
	security := learnosity.Security{
		ConsumerKey:    "yis0TYCu7U9V4o7M",
		Domain:         "localhost",
		Timestamp:      learnosity.FormatUTCTime(time.Now()),
		UserID:         "81b44c76-da57-47ce-8433-aa46b6d62a4d",
		ConsumerSecret: "74c5fd430cf1242a527f6223aebd42d30464be22",
	}

	id := uuid.New()
	fmt.Printf("Generated UUID: %s", id.String())
	features := []*entity.Feature{
		entity.NewPassageFeature("heading test", "content", id.String(), false),
	}

	dataAPI := &Client{}
	results, err := dataAPI.Request(context.Background(),
		&http.Client{},
		learnosity.EndpointDataAPISetFeatures,
		security,
		learnosity.Request{"features": features},
		learnosity.ActionSet)
	require.NoError(t, err)

	assert.True(t, results.Meta["status"].(bool))
}

func TestClient_SetQuestions(t *testing.T) {
	t.Skip()
	security := learnosity.Security{
		ConsumerKey:    "yis0TYCu7U9V4o7M",
		Domain:         "localhost",
		Timestamp:      learnosity.FormatUTCTime(time.Now()),
		UserID:         "81b44c76-da57-47ce-8433-aa46b6d62a4d",
		ConsumerSecret: "74c5fd430cf1242a527f6223aebd42d30464be22",
	}

	ref1 := uuid.New()
	fmt.Printf("Generated UUID 1 - multiple choice: %s\n", ref1.String())
	mcqValidationQuestion1 := entity.McqValidation{
		ScoringType: entity.ScoringTypeExactMatch,
		ValidResponse: entity.McqResponse{
			Value: []string{"1"},
			Score: 1,
		},
	}

	questions := []*entity.Question{
		{
			Reference: ref1.String(),
			Type:      entity.MultipleChoice,
			Data: &entity.McqQuestionData{
				BaseQuestionData: &entity.BaseQuestionData{
					Type:     entity.MultipleChoice,
					Stimulus: "<p>What is the capital of France?</p>",

					Metadata: entity.Metadata{
						DistractorRationale: "<p>The capital of France is Paris.</p>",
					},
					FeedbackAttempts: 1,
					InstantFeedback:  true,
				},
				Options: []entity.Options{
					{Label: "London", Value: "0"},
					{Label: "Paris", Value: "1"},
					{Label: "Dublin", Value: "2"},
					{Label: "Berlin", Value: "3"},
				},
				Validation: mcqValidationQuestion1,
			},
		},
	}

	dataAPI := &Client{}
	results, err := dataAPI.Request(context.Background(),
		&http.Client{},
		learnosity.EndpointDataAPISetQuestions,
		security,
		learnosity.Request{"questions": questions},
		learnosity.ActionSet)
	require.NoError(t, err)

	assert.True(t, results.Meta["status"].(bool))
}

func TestClient_SetItemsWithFeaturesAndQuestions(t *testing.T) {
	t.Skip()
	security := learnosity.Security{
		ConsumerKey:    "yis0TYCu7U9V4o7M",
		Domain:         "localhost",
		Timestamp:      learnosity.FormatUTCTime(time.Now()),
		UserID:         "81b44c76-da57-47ce-8433-aa46b6d62a4d",
		ConsumerSecret: "74c5fd430cf1242a527f6223aebd42d30464be22",
	}

	questionRef1 := uuid.New()
	fmt.Printf("Generated UUID 1 - multiple choice: %s\n", questionRef1.String())
	questionRef2 := uuid.New()
	fmt.Printf("Generated UUID 2 - multiple answers: %s\n", questionRef2.String())
	questionRef3 := uuid.New()
	fmt.Printf("Generated UUID 3 - ordering: %s\n", questionRef3.String())
	questionRef4 := uuid.New()
	fmt.Printf("Generated UUID 4 - fib: %s\n", questionRef4.String())
	questionRef5 := uuid.New()
	fmt.Printf("Generated UUID 5 - short text: %s\n", questionRef4.String())
	mcqValidationQuestion1 := entity.McqValidation{
		ScoringType: entity.ScoringTypeExactMatch,
		ValidResponse: entity.McqResponse{
			Value: []string{"1"},
			Score: 1,
		},
	}
	maqValidationQuestion2 :=
		entity.McqValidation{
			ScoringType: entity.ScoringTypeExactMatch,
			ValidResponse: entity.McqResponse{
				Value: []string{"0", "1"},
				Score: 1,
			},
		}

	questions := []*entity.Question{
		{
			Reference: questionRef1.String(),
			Type:      entity.MultipleChoice,
			Data: &entity.McqQuestionData{
				BaseQuestionData: &entity.BaseQuestionData{
					Type:     entity.MultipleChoice,
					Stimulus: "<p>What is the capital of France?</p>",
					Metadata: entity.Metadata{
						DistractorRationale: "<p>The capital of France is Paris.</p>",
					},
					FeedbackAttempts: 1,
					InstantFeedback:  true,
				},
				Options: []entity.Options{
					{Label: "London", Value: "0"},
					{Label: "Paris", Value: "1"},
					{Label: "Dublin", Value: "2"},
					{Label: "Berlin", Value: "3"},
				},

				Validation: mcqValidationQuestion1,
			},
		},
		// multiple answers
		{
			Reference: questionRef2.String(),
			Type:      entity.MultipleChoice,
			Data: &entity.McqQuestionData{
				BaseQuestionData: &entity.BaseQuestionData{
					Type:     entity.MultipleChoice,
					Stimulus: "<p>Which of the following are capital cities?</p>",

					Metadata: entity.Metadata{
						DistractorRationale: "<p>Paris and London are capital cities.</p>",
					},
					FeedbackAttempts: 1,
					InstantFeedback:  true,
				},
				MultipleResponses: true,
				Options: []entity.Options{
					{Label: "London", Value: "0"},
					{Label: "Paris", Value: "1"},
					{Label: "Dublin", Value: "2"},
					{Label: "Berlin", Value: "3"},
				},
				Validation: maqValidationQuestion2,
			},
		},
		// ordering question
		{
			Reference: questionRef3.String(),
			Type:      entity.Ordering,
			Data: &entity.OrdQuestionData{
				List: []string{
					"Monday",
					"Tuesday",
					"Wednesday",
					"Thursday",
					"Friday",
					"Saturday",
					"Sunday",
				},
				BaseQuestionData: &entity.BaseQuestionData{
					Type:     entity.Ordering,
					Stimulus: "<p>Sort the weekdays in order.</p>",
					Metadata: entity.Metadata{
						DistractorRationale: "<p>Monday is the first day of the week.</p>",
					},
					FeedbackAttempts: 1,
					InstantFeedback:  true,
				},
				Validation: entity.OrdValidation{
					ScoringType: entity.ScoringTypeExactMatch,
					ValidResponse: entity.OrdResponse{
						Score: 1,
						Value: []int{0, 1, 2, 3, 4, 5, 6},
					},
				},
			},
		},
		// fib question
		{
			Reference: questionRef4.String(),
			Type:      entity.FillInTheBlank,
			Data: &entity.FibQuestionData{
				BaseQuestionData: &entity.BaseQuestionData{
					Type: entity.FillInTheBlank,
					Metadata: entity.Metadata{
						DistractorRationale: "<p>The capital of France is Paris. The capital of Belgium is Brussels.</p>",
					},
					FeedbackAttempts: 1,
					InstantFeedback:  true},
				Template: "<p>The capital of France is {{response}}. The capital of Belgium is {{response}}.</p>",
				Validation: entity.FibValidation{
					ScoringType: entity.ScoringTypeExactMatch,
					ValidResponse: entity.FibResponse{
						Value: []string{"Paris", "Brussels"},
						Score: 1,
					},
					AltResponses: []entity.FibResponse{
						{
							Value: []string{"Phap", "Brussels"},
							Score: 1,
						},
					},
				},
			},
		},
		// short text
		{
			Reference: questionRef5.String(),
			Type:      entity.ShortText,
			Data: &entity.StqQuestionData{
				BaseQuestionData: &entity.BaseQuestionData{
					Stimulus: "<p>What is the capital of France?</p>",
					Type:     entity.ShortText,
					Metadata: entity.Metadata{
						DistractorRationale: "<p>The capital of France is Paris.</p>",
					},
					FeedbackAttempts: 1,
					InstantFeedback:  true,
				},
				Validation: entity.StqValidation{
					ScoringType: entity.ScoringTypeExactMatch,
					ValidResponse: entity.StqResponse{
						Value: "Paris",
						Score: 1,
					},
					AltResponses: []entity.StqResponse{
						{
							Value: "Phap",
							Score: 1,
						},
					},
				},
			},
		},
	}

	dataAPI := &Client{}
	results, err := dataAPI.Request(context.Background(),
		&http.Client{},
		learnosity.EndpointDataAPISetQuestions,
		security,
		learnosity.Request{"questions": questions},
		learnosity.ActionSet)
	require.NoError(t, err)

	assert.True(t, results.Meta["status"].(bool))

	featureId := uuid.New()
	fmt.Printf("Generated UUID - feature_id: %s\n", featureId.String())
	features := []*entity.Feature{
		entity.NewPassageFeature("heading test", "content", featureId.String(), false),
	}

	results, err = dataAPI.Request(context.Background(),
		&http.Client{},
		learnosity.EndpointDataAPISetFeatures,
		security,
		learnosity.Request{"features": features},
		learnosity.ActionSet)
	require.NoError(t, err)

	assert.True(t, results.Meta["status"].(bool))

	item_id := uuid.New()
	fmt.Printf("Generated UUID - item_id: %s\n", item_id.String())
	items := []*entity.Item{
		{
			Reference: item_id.String(),
			Metadata:  nil,
			Definition: entity.Definition{
				Widgets: []entity.Reference{
					{
						Reference: featureId.String(),
					},
					{
						Reference: questionRef1.String(),
					},
					{
						Reference: questionRef2.String(),
					},
					{
						Reference: questionRef3.String(),
					},
					{
						Reference: questionRef4.String(),
					},
					{
						Reference: questionRef5.String(),
					},
				},
			},
			Status: "published",
			Questions: []entity.Reference{
				{
					Reference: questionRef1.String(),
				},
				{
					Reference: questionRef2.String(),
				},
				{
					Reference: questionRef3.String(),
				},
				{
					Reference: questionRef4.String(),
				},
				{
					Reference: questionRef5.String(),
				},
			},
			Features: []entity.Reference{
				{
					Reference: featureId.String(),
				},
			},
			Tags: entity.Tags{
				Tenant: []string{"tagtest"},
			},
		},
	}

	results, err = dataAPI.Request(context.Background(),
		&http.Client{},
		learnosity.EndpointDataAPISetItems,
		security,
		learnosity.Request{"items": items},
		learnosity.ActionSet)

	require.NoError(t, err)
	assert.True(t, results.Meta["status"].(bool))

}

func TestClient_GetItems(t *testing.T) {
	t.Skip()
	security := learnosity.Security{
		ConsumerKey:    "yis0TYCu7U9V4o7M",
		Domain:         "localhost",
		Timestamp:      learnosity.FormatUTCTime(time.Now()),
		UserID:         "81b44c76-da57-47ce-8433-aa46b6d62a4d",
		ConsumerSecret: "74c5fd430cf1242a527f6223aebd42d30464be22",
	}

	references := []string{
		"be417bfc-47c1-4ae5-80ea-4a7458746839",
	}

	dataAPI := &Client{}
	results, err := dataAPI.Request(
		context.Background(),
		&http.Client{},
		learnosity.EndpointDataAPIGetItems,
		security,
		learnosity.Request{
			"references": references,
		},
		learnosity.ActionGet)
	require.NoError(t, err)

	assert.True(t, results.Meta["status"].(bool))

	assert.GreaterOrEqual(t, results.Meta["records"].(float64), 1.0)

	// unmarshal the data to string
	var referencesRes []*entity.Reference
	err = json.Unmarshal(results.Data, &referencesRes)

	assert.NoError(t, err)
	assert.Equal(t, referencesRes[0].Reference, references[0])

}

func TestClient_SetActivities(t *testing.T) {
	t.Skip()
	dataAPI := &Client{}
	security := learnosity.Security{
		ConsumerKey:    "yis0TYCu7U9V4o7M",
		Domain:         "localhost",
		Timestamp:      learnosity.FormatUTCTime(time.Now()),
		UserID:         "81b44c76-da57-47ce-8433-aa46b6d62a4d",
		ConsumerSecret: "74c5fd430cf1242a527f6223aebd42d30464be22",
	}

	reference := uuid.New()
	fmt.Printf("Generated UUID: %s", reference.String())

	activities := []*entity.Activity{
		{
			Reference: "504f1fbe-a82b-4aa9-90de-eb492410138b",
			Data: entity.ActivityData{
				Items:         []any{"1ae84141-3897-416a-b1fd-0c676b7e8982"},
				RenderingType: entity.RenderingTypeAssess,
				Config: entity.Config{
					Regions: "main",
				},
			},
			Tags: entity.Tags{
				Tenant: []string{"tagtest"},
			},
		},
	}

	results, err := dataAPI.Request(context.Background(),
		&http.Client{},
		learnosity.EndpointDataAPISetActivities,
		security,
		learnosity.Request{"activities": activities},
		learnosity.ActionSet)
	require.NoError(t, err)

	assert.True(t, results.Meta["status"].(bool))
}

func TestClient_GetActivities(t *testing.T) {
	t.Skip()
	security := learnosity.Security{
		ConsumerKey:    "yis0TYCu7U9V4o7M",
		Domain:         "localhost",
		Timestamp:      learnosity.FormatUTCTime(time.Now()),
		UserID:         "81b44c76-da57-47ce-8433-aa46b6d62a4d",
		ConsumerSecret: "74c5fd430cf1242a527f6223aebd42d30464be22",
	}

	references := []string{
		"504f1fbe-a82b-4aa9-90de-eb492410138b",
	}

	dataAPI := &Client{}
	results, err := dataAPI.Request(
		context.Background(),
		&http.Client{},
		learnosity.EndpointDataAPIGetActivities,
		security,
		learnosity.Request{
			"references": references,
		},
		learnosity.ActionGet)
	require.NoError(t, err)

	assert.True(t, results.Meta["status"].(bool))

	assert.GreaterOrEqual(t, results.Meta["records"].(float64), 1.0)

	// unmarshal the data to string
	var referencesRes []*entity.Reference
	err = json.Unmarshal(results.Data, &referencesRes)

	assert.NoError(t, err)
	assert.Equal(t, referencesRes[0].Reference, references[0])
}
