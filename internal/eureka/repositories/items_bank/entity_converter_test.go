package repositories

import (
	"fmt"
	"testing"

	entities "github.com/manabie-com/backend/internal/eureka/entities/items_bank"
	learnosity_entity "github.com/manabie-com/backend/internal/golibs/learnosity/entity"
)

func TestItemsBankQuestion_ToLearnosityQuestion(t *testing.T) {
	t.Parallel()

	imageURL := "https://example.com/test.png"

	getContentWithImageURL := func(text string) string {
		return fmt.Sprintf(`<p>%s</p><p><img src="%s" style="max-width:100%%;" /></p>`, text, imageURL)
	}

	testCases := []struct {
		name   string
		input  *entities.ItemsBankQuestion
		output *learnosity_entity.Question
	}{
		{
			name: "Multiple Choice Question",
			input: &entities.ItemsBankQuestion{
				ItemID:               "1",
				Point:                2,
				QuestionType:         entities.QuestionTypeMultipleChoice,
				QuestionContentText:  "What is the capital of France?",
				QuestionContentImage: imageURL,
				Options: []*entities.ItemsBankOption{
					{OptionText: "London", OptionImage: imageURL},
					{OptionText: "Paris"},
					{OptionText: "Berlin"},
					{OptionText: "Madrid", OptionImage: imageURL},
				},
				CorrectOptions:   []bool{false, true, false, false},
				ExplanationText:  "Paris is the capital of France.",
				ExplanationImage: imageURL,
			},
			output: &learnosity_entity.Question{
				Reference: "1",
				Type:      learnosity_entity.MultipleChoice,
				Data: &learnosity_entity.McqQuestionData{
					BaseQuestionData: &learnosity_entity.BaseQuestionData{
						Stimulus: getContentWithImageURL("What is the capital of France?"),
						Type:     learnosity_entity.MultipleChoice,
						Metadata: learnosity_entity.Metadata{
							DistractorRationale: getContentWithImageURL("Paris is the capital of France."),
						},
						FeedbackAttempts: 1,
						InstantFeedback:  true,
					},
					Options: []learnosity_entity.Options{
						{Value: "0", Label: getContentWithImageURL("London")},
						{Value: "1", Label: "Paris"},
						{Value: "2", Label: "Berlin"},
						{Value: "3", Label: getContentWithImageURL("Madrid")},
					},
					ShuffleOptions:    true,
					MultipleResponses: false,
					Validation: learnosity_entity.McqValidation{
						ScoringType: learnosity_entity.ScoringTypeExactMatch,
						ValidResponse: learnosity_entity.McqResponse{
							Value: []string{"1"},
							Score: 2,
						},
					},
				},
			},
		},
		{
			name: "Multiple Answers Question",
			input: &entities.ItemsBankQuestion{
				ItemID:              "2",
				Point:               1,
				QuestionType:        entities.QuestionTypeMultipleAnswers,
				QuestionContentText: "Which of the following are capital cities?",
				Options: []*entities.ItemsBankOption{
					{OptionText: "London", OptionImage: imageURL},
					{OptionText: "Paris"},
					{OptionText: "Berlin"},
					{OptionText: "Madrid"},
				},
				CorrectOptions:  []bool{true, true, false, false},
				ExplanationText: "Paris and London are capital cities.",
			},
			output: &learnosity_entity.Question{
				Reference: "2",
				Type:      learnosity_entity.MultipleChoice,
				Data: &learnosity_entity.McqQuestionData{
					BaseQuestionData: &learnosity_entity.BaseQuestionData{
						Stimulus: "Which of the following are capital cities?",
						Type:     learnosity_entity.MultipleChoice,
						Metadata: learnosity_entity.Metadata{
							DistractorRationale: "Paris and London are capital cities.",
						},
						FeedbackAttempts: 1,
						InstantFeedback:  true,
					},

					Options: []learnosity_entity.Options{
						{Value: "0", Label: getContentWithImageURL("London")},
						{Value: "1", Label: "Paris"},
						{Value: "2", Label: "Berlin"},
						{Value: "3", Label: "Madrid"},
					},
					Validation: learnosity_entity.McqValidation{
						ScoringType: learnosity_entity.ScoringTypeExactMatch,
						ValidResponse: learnosity_entity.McqResponse{
							Value: []string{"0", "1"},
							Score: 1,
						},
					},
					ShuffleOptions:    true,
					MultipleResponses: true,
				},
			},
		},
		{
			name: "Ordering Question",
			input: &entities.ItemsBankQuestion{
				ItemID:              "3",
				Point:               1,
				QuestionType:        entities.QuestionTypeOrdering,
				QuestionContentText: "Sort the weekdays in order.",
				Options: []*entities.ItemsBankOption{
					{OptionText: "Monday", OptionImage: imageURL},
					{OptionText: "Tuesday"},
					{OptionText: "Wednesday"},
					{OptionText: "Thursday"},
					{OptionText: "Friday", OptionImage: imageURL},
					{OptionText: "Saturday"},
					{OptionText: "Sunday"},
				},
				CorrectOptions:  []bool{},
				ExplanationText: "The weekdays are Monday, Tuesday, Wednesday, Thursday, Friday, Saturday, Sunday.",
			},
			output: &learnosity_entity.Question{
				Reference: "3",
				Type:      learnosity_entity.Ordering,
				Data: &learnosity_entity.OrdQuestionData{
					BaseQuestionData: &learnosity_entity.BaseQuestionData{
						Stimulus: "Sort the weekdays in order.",
						Type:     learnosity_entity.Ordering,
						Metadata: learnosity_entity.Metadata{
							DistractorRationale: "The weekdays are Monday, Tuesday, Wednesday, Thursday, Friday, Saturday, Sunday.",
						},
						FeedbackAttempts: 1,
						InstantFeedback:  true,
					},
					List: []string{getContentWithImageURL("Monday"), "Tuesday", "Wednesday", "Thursday", getContentWithImageURL("Friday"), "Saturday", "Sunday"},
					Validation: learnosity_entity.OrdValidation{
						ScoringType: learnosity_entity.ScoringTypeExactMatch,
						ValidResponse: learnosity_entity.OrdResponse{
							Value: []int{0, 1, 2, 3, 4, 5, 6},
							Score: 1,
						},
					},
					ShuffleOptions: true,
				},
			},
		},
		{
			name: "Fill in bank question",
			input: &entities.ItemsBankQuestion{
				ItemID:              "4",
				Point:               1,
				QuestionType:        entities.QuestionTypeFillInTheBlank,
				QuestionContentText: "The capital of France is {{response}}. The capital of Belgium is {{response}}.",
				Options: []*entities.ItemsBankOption{
					{OptionText: "Paris;Brussels", OptionImage: imageURL},
					{OptionText: "Phap;Brussels", OptionImage: imageURL},
				},
				CorrectOptions:  []bool{},
				ExplanationText: "The capital of France is Paris. The capital of Belgium is Brussels.",
			},
			output: &learnosity_entity.Question{
				Reference: "4",
				Type:      learnosity_entity.FillInTheBlank,
				Data: &learnosity_entity.FibQuestionData{
					BaseQuestionData: &learnosity_entity.BaseQuestionData{
						Stimulus: "",
						Type:     learnosity_entity.FillInTheBlank,
						Metadata: learnosity_entity.Metadata{
							DistractorRationale: "The capital of France is Paris. The capital of Belgium is Brussels.",
						},
						FeedbackAttempts: 1,
						InstantFeedback:  true,
					},
					Template: "The capital of France is {{response}}. The capital of Belgium is {{response}}.",
					Validation: learnosity_entity.FibValidation{
						ScoringType: learnosity_entity.ScoringTypeExactMatch,
						ValidResponse: learnosity_entity.FibResponse{
							Value: []string{"Paris", "Brussels"},
							Score: 1,
						},
						AltResponses: []learnosity_entity.FibResponse{
							{
								Value: []string{"Phap", "Brussels"},
								Score: 1,
							},
						},
					},
				},
			},
		},
		{
			name: "Short text question",
			input: &entities.ItemsBankQuestion{
				ItemID:               "5",
				Point:                1,
				QuestionType:         entities.QuestionTypeShortText,
				QuestionContentImage: `https://storage.googleapis.com/stag-manabie-backend/items_bank/abc.jpg`,
				Options: []*entities.ItemsBankOption{
					{OptionText: "Paris", OptionImage: imageURL},
					{OptionText: "Phap", OptionImage: imageURL},
				},
				CorrectOptions:  []bool{},
				ExplanationText: "The capital of France is Paris.",
			},
			output: &learnosity_entity.Question{
				Reference: "5",
				Type:      learnosity_entity.ShortText,
				Data: &learnosity_entity.StqQuestionData{
					BaseQuestionData: &learnosity_entity.BaseQuestionData{
						Stimulus: `<img src="https://storage.googleapis.com/stag-manabie-backend/items_bank/abc.jpg" style="max-width:100%;" />`,
						Type:     learnosity_entity.ShortText,
						Metadata: learnosity_entity.Metadata{
							DistractorRationale: "The capital of France is Paris.",
						},
						FeedbackAttempts: 1,
						InstantFeedback:  true,
					},
					Validation: learnosity_entity.StqValidation{
						ScoringType: learnosity_entity.ScoringTypeExactMatch,
						ValidResponse: learnosity_entity.StqResponse{
							Value: "Paris",
							Score: 1,
						},
						AltResponses: []learnosity_entity.StqResponse{
							{
								Value: "Phap",
								Score: 1,
							},
						},
					},
				},
			},
		},
		{
			name: "Manual Input Question",
			input: &entities.ItemsBankQuestion{
				ItemID:               "1",
				Point:                1,
				QuestionType:         entities.QuestionManualInput,
				QuestionContentText:  "What is the capital of France? You can write anything on your paper and click on explanation button to compare with correct result.",
				QuestionContentImage: imageURL,
				Options:              []*entities.ItemsBankOption{},
				CorrectOptions:       []bool{},
				ExplanationText:      "Paris is the capital of France.",
				ExplanationImage:     "https://example.com/test.png",
			},
			output: &learnosity_entity.Question{
				Reference: "1",
				Type:      learnosity_entity.MultipleChoice,
				Data: &learnosity_entity.McqQuestionData{
					BaseQuestionData: &learnosity_entity.BaseQuestionData{
						Stimulus: getContentWithImageURL("What is the capital of France? You can write anything on your paper and click on explanation button to compare with correct result."),
						Type:     learnosity_entity.MultipleChoice,
						Metadata: learnosity_entity.Metadata{
							MIQHintExplanation: `<p>Paris is the capital of France.</p><p><img src="https://example.com/test.png" style="max-width:100%;" /></p>`,
						},
						FeedbackAttempts: 1,
						InstantFeedback:  true,
					},
					Options: []learnosity_entity.Options{
						{Value: "0", Label: "True"},
						{Value: "1", Label: "False"},
					},
					ShuffleOptions:    false,
					MultipleResponses: false,
					Validation: learnosity_entity.McqValidation{
						ScoringType: learnosity_entity.ScoringTypeExactMatch,
						ValidResponse: learnosity_entity.McqResponse{
							Value: []string{"0"},
							Score: 1,
						},
					},
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			actual := ToLearnosityQuestion(testCase.input)
			expected := testCase.output

			err := verifyBasicQuestionData(actual.Data, expected.Data)
			if err != nil {
				t.Error(err)
			}

			if testCase.input.QuestionType == entities.QuestionTypeMultipleChoice || testCase.input.QuestionType == entities.QuestionTypeMultipleAnswers {
				err = verifyMCQQuestionData(actual, expected)
				if err != nil {
					t.Error(err)
				}
			}

			if testCase.input.QuestionType == entities.QuestionTypeOrdering {
				err = verifyOrderingQuestionData(actual, expected)
				if err != nil {
					t.Error(err)
				}
			}

			if testCase.input.QuestionType == entities.QuestionTypeFillInTheBlank {
				err = verifyFIBQuestionData(actual, expected)
				if err != nil {
					t.Error(err)
				}
			}

			if testCase.input.QuestionType == entities.QuestionTypeShortText {
				err = verifySTQQuestionData(actual, expected)
				if err != nil {
					t.Error(err)
				}
			}
		})
	}
}

func verifySTQQuestionData(actual *learnosity_entity.Question, expected *learnosity_entity.Question) error {
	expectedValidation := expected.Data.(*learnosity_entity.StqQuestionData).Validation
	actualValidation := actual.Data.(*learnosity_entity.StqQuestionData).Validation
	if expectedValidation.ScoringType != actualValidation.ScoringType {
		return fmt.Errorf("Expected Validation.ScoringType: %s, but got: %s", expectedValidation.ScoringType, actualValidation.ScoringType)
	}

	expectedResponse := expectedValidation.ValidResponse.Value
	actualResponse := actualValidation.ValidResponse.Value

	if expectedResponse != actualResponse {
		return fmt.Errorf("Expected Validation.ValidResponse.Value: %s, but got: %s", expectedResponse, actualResponse)
	}

	if len(expectedValidation.AltResponses) != len(actualValidation.AltResponses) {
		return fmt.Errorf("Expected Validation.AltResponses: %v, but got: %v", expectedValidation.AltResponses, actualValidation.AltResponses)
	}

	expectedAltResponses := expectedValidation.AltResponses
	actualAltResponses := actualValidation.AltResponses
	for i, altResponse := range expectedAltResponses {
		if altResponse.Value != actualAltResponses[i].Value {
			return fmt.Errorf("Expected Validation.AltResponses[%d].Value: %s, but got: %s", i, actualAltResponses[i].Value, altResponse.Value)
		}

	}

	if expectedValidation.ValidResponse.Score != actualValidation.ValidResponse.Score {
		return fmt.Errorf("Expected Validation.ValidResponse.Score: %d, but got: %d", actualValidation.ValidResponse.Score, expectedValidation.ValidResponse.Score)
	}
	return nil
}

func verifyFIBQuestionData(actual *learnosity_entity.Question, expected *learnosity_entity.Question) error {
	expectedValidation := expected.Data.(*learnosity_entity.FibQuestionData).Validation
	actualValidation := expected.Data.(*learnosity_entity.FibQuestionData).Validation
	if expectedValidation.ScoringType != actualValidation.ScoringType {
		return fmt.Errorf("Expected Validation.ScoringType: %s, but got: %s", expectedValidation.ScoringType, actualValidation.ScoringType)
	}

	expectedResponse := expectedValidation.ValidResponse.Value
	actualResponse := actualValidation.ValidResponse.Value
	if len(expectedResponse) != len(actualResponse) {
		return fmt.Errorf("Expected Validation.ValidResponse.Value: %v, but got: %v", expectedResponse, actualResponse)
	}

	for i, option := range expectedResponse {
		if option != actualResponse[i] {
			return fmt.Errorf("Expected Validation.ValidResponse.Value[%d]: %s, but got: %s", i, option, actualResponse[i])
		}
	}

	// verify alternative answers
	if len(expectedValidation.AltResponses) != len(actualValidation.AltResponses) {
		return fmt.Errorf("Expected Validation.AltResponses: %v, but got: %v", expectedValidation.AltResponses, actualValidation.AltResponses)
	}

	expectedAltResponses := expectedValidation.AltResponses
	actualAltResponses := actualValidation.AltResponses
	for i, altResponse := range expectedAltResponses {
		if len(altResponse.Value) != len(actualAltResponses[i].Value) {
			//len
			return fmt.Errorf("Expected len Validation.AltResponses[%d].Value: %d, but got: %d", i, len(altResponse.Value), len(actualAltResponses[i].Value))
		}

		expectedAnswers := expectedAltResponses[i].Value
		actualAnswers := actualAltResponses[i].Value
		for j, option := range expectedAnswers {
			if option != actualAnswers[j] {
				return fmt.Errorf("Expected Validation.AltResponses[%d].Value[%d]: %s, but got: %s", i, j, option, actualAnswers[j])
			}
		}
	}

	if expectedValidation.ValidResponse.Score != actualValidation.ValidResponse.Score {
		return fmt.Errorf("Expected Validation.ValidResponse.Score: %d, but got: %d", expectedValidation.ValidResponse.Score, actualValidation.ValidResponse.Score)
	}
	return nil
}

func verifyOrderingQuestionData(actual *learnosity_entity.Question, expected *learnosity_entity.Question) error {
	expectedData := expected.Data.(*learnosity_entity.OrdQuestionData)
	actualData := actual.Data.(*learnosity_entity.OrdQuestionData)
	if len(expectedData.List) != len(actualData.List) {
		return fmt.Errorf("Expected Options: %v, but got: %v", expectedData.List, actualData.List)
	}

	for i, option := range expectedData.List {
		if option != actualData.List[i] {
			return fmt.Errorf("Expected Options[%d].Value: %s, but got: %s", i, option, actualData.List[i])
		}
	}

	if expectedData.ShuffleOptions != actualData.ShuffleOptions {
		return fmt.Errorf("Expected ShuffleOptions: %t, but got: %t", expectedData.ShuffleOptions, actualData.ShuffleOptions)
	}

	expectedValidation := expectedData.Validation
	actualValidation := actualData.Validation
	// Ord validation
	if expectedValidation.ScoringType != actualValidation.ScoringType {
		return fmt.Errorf("Expected Validation.ScoringType: %s, but got: %s", expectedValidation.ScoringType, actualValidation.ScoringType)
	}

	expectedResponse := expectedValidation.ValidResponse.Value
	actualResponse := actualValidation.ValidResponse.Value
	if len(expectedResponse) != len(actualResponse) {
		return fmt.Errorf("Expected Validation.ValidResponse.Value: %v, but got: %v", expectedResponse, actualResponse)
	}

	for i, option := range expectedResponse {
		if option != actualResponse[i] {
			return fmt.Errorf("Expected Validation.ValidResponse.Value[%d]: %d, but got: %d", i, option, actualResponse[i])
		}
	}

	if expectedValidation.ValidResponse.Score != actualValidation.ValidResponse.Score {
		return fmt.Errorf("Expected Validation.ValidResponse.Score: %d, but got: %d", expectedValidation.ValidResponse.Score, actualValidation.ValidResponse.Score)
	}
	return nil
}

func verifyBasicQuestionData(actual learnosity_entity.QuestionDataInterface, expected learnosity_entity.QuestionDataInterface) error {
	expectedData := expected.GetBasicData()
	actualData := actual.GetBasicData()
	if expectedData.Stimulus != actualData.Stimulus {
		return fmt.Errorf("Expected Stimulus: %s, but got: %s", expectedData.Stimulus, actualData.Stimulus)
	}

	if expectedData.Type != actualData.Type {
		return fmt.Errorf("Expected Type: %s, but got: %s", expectedData.Type, actualData.Type)
	}

	if expectedData.Metadata.DistractorRationale != actualData.Metadata.DistractorRationale {
		return fmt.Errorf("Expected Metadata.DistractorRationale: %s, but got: %s", expectedData.Metadata.DistractorRationale, actualData.Metadata.DistractorRationale)
	}

	if expectedData.Metadata.MIQHintExplanation != "" {
		if expectedData.Metadata.MIQHintExplanation != actualData.Metadata.MIQHintExplanation {
			return fmt.Errorf("Expected Metadata.DistractorRationale: %s, but got: %s", expectedData.Metadata.MIQHintExplanation, actualData.Metadata.MIQHintExplanation)
		}
	}

	if expectedData.FeedbackAttempts != actualData.FeedbackAttempts {
		return fmt.Errorf("Expected FeedbackAttempts: %d, but got: %d", expectedData.FeedbackAttempts, actualData.FeedbackAttempts)
	}

	if expectedData.InstantFeedback != actualData.InstantFeedback {
		return fmt.Errorf("Expected InstantFeedback: %t, but got: %t", expectedData.InstantFeedback, actualData.InstantFeedback)
	}
	return nil
}

func verifyMCQQuestionData(actual *learnosity_entity.Question, expected *learnosity_entity.Question) error {
	expectedData := expected.Data.(*learnosity_entity.McqQuestionData)
	actualData := actual.Data.(*learnosity_entity.McqQuestionData)
	if len(expectedData.Options) != len(actualData.Options) {
		return fmt.Errorf("Expected Options: %v, but got: %v", expectedData.Options, actualData.Options)
	}

	for i, option := range expectedData.Options {
		if option.Label != actualData.Options[i].Label {
			return fmt.Errorf("Expected Options[%d].Value: %s, but got: %s", i, option.Label, actualData.Options[i].Label)
		}

		if option.Value != actualData.Options[i].Value {
			return fmt.Errorf("Expected Options[%d].Value: %s, but got: %s", i, option.Value, actualData.Options[i].Value)
		}
	}

	expectedValidation := expectedData.Validation
	actualValidation := actualData.Validation
	if expectedValidation.ScoringType != actualValidation.ScoringType {
		return fmt.Errorf("Expected Validation.ScoringType: %s, but got: %s", expectedValidation.ScoringType, actualValidation.ScoringType)
	}

	if len(expectedValidation.ValidResponse.Value) != len(actualValidation.ValidResponse.Value) {
		return fmt.Errorf("Expected Validation.ValidResponse.Value: %v, but got: %v", expectedValidation.ValidResponse.Value, actualValidation.ValidResponse.Value)
	}

	if expectedValidation.ValidResponse.Score != actualValidation.ValidResponse.Score {
		return fmt.Errorf("Expected Validation.ValidResponse.Score: %d, but got: %d", expectedValidation.ValidResponse.Score, actualValidation.ValidResponse.Score)
	}

	if expectedData.ShuffleOptions != actualData.ShuffleOptions {
		return fmt.Errorf("Expected ShuffleOptions: %t, but got: %t", expectedData.ShuffleOptions, actualData.ShuffleOptions)
	}

	if expectedData.MultipleResponses != actualData.MultipleResponses {
		return fmt.Errorf("Expected MultipleResponses: %t, but got: %t", expectedData.MultipleResponses, actualData.MultipleResponses)
	}
	return nil
}

func TestItemsBankItem_ToLearnosityFeature(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name   string
		input  *entities.ItemsBankItem
		output *learnosity_entity.Feature
	}{
		{
			name: "Feature with description",
			input: &entities.ItemsBankItem{
				ItemID:              "1",
				ItemName:            "What is the capital of France?",
				ItemDescriptionText: "lorem ipsum dolor sit amet consectetur adipiscing elit sed do eiusmod tempor incididunt ut labore et dolore magna aliqua",
			},
			output: &learnosity_entity.Feature{
				Type: "sharedpassage",
				Data: learnosity_entity.Data{
					Heading: "What is the capital of France?",
					Content: "lorem ipsum dolor sit amet consectetur adipiscing elit sed do eiusmod tempor incididunt ut labore et dolore magna aliqua",
					Type:    "sharedpassage",
				},
			},
		},
		{
			name: "Feature with no description",
			input: &entities.ItemsBankItem{
				ItemID:              "1",
				ItemName:            "What is the capital of France?",
				ItemDescriptionText: "",
			},
			output: &learnosity_entity.Feature{
				Type: "sharedpassage",
				Data: learnosity_entity.Data{
					Heading: "What is the capital of France?",
					Content: "",
					Type:    "sharedpassage",
				},
			},
		},
		{
			name: "Empty input content",
			input: &entities.ItemsBankItem{
				ItemID:              "1",
				ItemName:            "",
				ItemDescriptionText: "",
			},
			output: nil,
		},
		{
			name: "Empty input content with spaces",
			input: &entities.ItemsBankItem{
				ItemID:              "1",
				ItemName:            " ",
				ItemDescriptionText: "  ",
			},
			output: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			actual := ToLearnosityFeature(testCase.input)
			expected := testCase.output

			if expected == nil && actual == nil {
				return
			}

			if expected.Type != actual.Type {
				t.Errorf("Expected Type: %s, but got: %s", expected.Type, actual.Type)
			}

			if expected.Data.Heading != actual.Data.Heading {
				t.Errorf("Expected Data.Heading: %s, but got: %s", expected.Data.Heading, actual.Data.Heading)
			}

			if expected.Data.Content != actual.Data.Content {
				t.Errorf("Expected Data.Content: %s, but got: %s", expected.Data.Content, actual.Data.Content)
			}

			if expected.Data.Type != actual.Data.Type {
				t.Errorf("Expected Data.Type: %s, but got: %s", expected.Data.Type, actual.Data.Type)
			}

		})
	}
}

func TestItemsBankItem_ToLearnosityItem(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name             string
		input            *entities.ItemsBankItem
		orgID            string
		questionRefs     []string
		featureReference string
		output           *learnosity_entity.Item
		expectedError    error
	}{
		{
			name:  "Item 1 - happy case",
			orgID: "manabie_org_id",
			questionRefs: []string{
				"question_1", "question_2", "question_3",
			},
			featureReference: "feature_1",
			input: &entities.ItemsBankItem{
				ItemID:              "item_1",
				ItemName:            "What is the capital of France?",
				ItemDescriptionText: "lorem ipsum dolor sit amet consectetur adipiscing elit sed do eiusmod tempor incididunt ut labore et dolore magna aliqua",
			},
			output: &learnosity_entity.Item{
				Status:    "published",
				Reference: "item_1",
				Metadata:  nil,
				Definition: learnosity_entity.Definition{
					Widgets: []learnosity_entity.Reference{
						{
							Reference: "feature_1",
						},
						{
							Reference: "question_1",
						},
						{
							Reference: "question_2",
						},
						{
							Reference: "question_3",
						},
					},
				},
				Features: []learnosity_entity.Reference{
					{
						Reference: "feature_1",
					},
				},
				Questions: []learnosity_entity.Reference{
					{
						Reference: "question_1",
					},
					{
						Reference: "question_2",
					},
					{
						Reference: "question_3",
					},
				},
				Tags: learnosity_entity.Tags{
					Tenant: []string{"manabie_org_id"},
				},
			},
		},
		{
			name:  "Item 1 - without feature",
			orgID: "manabie_org_id",
			questionRefs: []string{
				"question_1", "question_2", "question_3",
			},
			featureReference: "",
			input: &entities.ItemsBankItem{
				ItemID:              "item_1",
				ItemName:            "What is the capital of France?",
				ItemDescriptionText: "lorem ipsum dolor sit amet consectetur adipiscing elit sed do eiusmod tempor incididunt ut labore et dolore magna aliqua",
			},
			output: &learnosity_entity.Item{
				Status:    "published",
				Reference: "item_1",
				Metadata:  nil,
				Definition: learnosity_entity.Definition{
					Widgets: []learnosity_entity.Reference{
						{
							Reference: "question_1",
						},
						{
							Reference: "question_2",
						},
						{
							Reference: "question_3",
						},
					},
				},
				Features: []learnosity_entity.Reference{},
				Questions: []learnosity_entity.Reference{
					{
						Reference: "question_1",
					},
					{
						Reference: "question_2",
					},
					{
						Reference: "question_3",
					},
				},
				Tags: learnosity_entity.Tags{
					Tenant: []string{"manabie_org_id"},
				},
			},
		},
		{
			name:             "Item 1 - empty question refs",
			orgID:            "manabie_org_id",
			questionRefs:     []string{},
			featureReference: "",
			input: &entities.ItemsBankItem{
				ItemID:              "item_1",
				ItemName:            "What is the capital of France?",
				ItemDescriptionText: "lorem ipsum dolor sit amet consectetur adipiscing elit sed do eiusmod tempor incididunt ut labore et dolore magna aliqua",
			},
			output:        nil,
			expectedError: fmt.Errorf("question refs must not be empty"),
		},
		{
			name:  "Item 1 - empty org id",
			orgID: "",
			questionRefs: []string{
				"question_1", "question_2", "question_3",
			},
			featureReference: "",
			input: &entities.ItemsBankItem{
				ItemID:              "item_1",
				ItemName:            "What is the capital of France?",
				ItemDescriptionText: "lorem ipsum dolor sit amet consectetur adipiscing elit sed do eiusmod tempor incididunt ut labore et dolore magna aliqua",
			},
			output:        nil,
			expectedError: fmt.Errorf("organization ID must not be empty"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			actual, err := ToLearnosityItem(testCase.input, testCase.orgID, testCase.questionRefs, testCase.featureReference)

			if testCase.expectedError != nil {
				if err == nil {
					t.Errorf("Expected error: %v, but got nil", testCase.expectedError)
				} else if testCase.expectedError.Error() != err.Error() {
					t.Errorf("Expected error: %v, but got: %v", testCase.expectedError, err)
				}
				return
			}

			expected := testCase.output

			if actual == nil {
				t.Errorf("Expected Item to not be nil")
			}

			if expected.Status != actual.Status {
				t.Errorf("Expected Status: %s, but got: %s", expected.Status, actual.Status)
			}

			if expected.Reference != actual.Reference {
				t.Errorf("Expected Reference: %s, but got: %s", expected.Reference, actual.Reference)
			}

			if expected.Metadata != actual.Metadata {
				t.Errorf("Expected Metadata: %v, but got: %v", expected.Metadata, actual.Metadata)
			}

			if len(expected.Questions) != len(actual.Questions) {
				t.Errorf("Expected len(Questions): %d, but got: %d", len(expected.Questions), len(actual.Questions))
			}

			if len(expected.Features) != len(actual.Features) {
				t.Errorf("Expected len(Features): %d, but got: %d", len(expected.Features), len(actual.Features))
			}

			if len(expected.Definition.Widgets) != len(actual.Definition.Widgets) {
				t.Errorf("Expected len(Definition.Widgets): %d, but got: %d", len(expected.Definition.Widgets), len(actual.Definition.Widgets))
			}

			for i, expectedQuestion := range expected.Questions {
				if expectedQuestion.Reference != actual.Questions[i].Reference {
					t.Errorf("Expected Questions[%d].Reference: %s, but got: %s", i, expectedQuestion.Reference, actual.Questions[i].Reference)
				}
			}

			for i, expectedFeature := range expected.Features {
				if expectedFeature.Reference != actual.Features[i].Reference {
					t.Errorf("Expected Features[%d].Reference: %s, but got: %s", i, expectedFeature.Reference, actual.Features[i].Reference)
				}
			}

			for i, expectedWidget := range expected.Definition.Widgets {
				if expectedWidget.Reference != actual.Definition.Widgets[i].Reference {
					t.Errorf("Expected Definition.Widgets[%d].Reference: %s, but got: %s", i, expectedWidget.Reference, actual.Definition.Widgets[i].Reference)
				}
			}

			if len(expected.Tags.Tenant) != 1 {
				t.Errorf("Expected len(Tags.Tenant): %d, but got: %d", 1, len(expected.Tags.Tenant))
			}
			for i, expectedTenant := range expected.Tags.Tenant {
				if expectedTenant != actual.Tags.Tenant[i] {
					t.Errorf("Expected Tags.Tenant[%d]: %s, but got: %s", i, expectedTenant, actual.Tags.Tenant[i])
				}
			}

		})
	}
}

func TestCheckLatexInString(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "empty",
			input:    ``,
			expected: false,
		},
		{
			name:     "Latex in string - inline",
			input:    ` This is a string with latex \(frac{1}{2}\)`,
			expected: true,
		},
		{
			name:     "Latex in string - block",
			input:    `This is a string with latex \[frac{1}{2}\]`,
			expected: true,
		},
		{
			name:     "No latex in string",
			input:    "This is a string with no latex",
			expected: false,
		},
		{
			name:     "Latex in string - inline and block",
			input:    `This is a string with latex \[frac{1}{2}\] and \(frac{1}{2}\)`,
			expected: true,
		},
		{
			name: "Multiple lines",
			input: `This is a string with no latex 
			And another line with latex \[frac{1}{2}\] and \(frac{1}{2}\)`,
			expected: true,
		},
		{
			name: "Multiple lines - no latex",
			input: `This is a string with latex
			And another line with no latex`,
			expected: false,
		},
		{
			name: "Latex content split across lines",
			input: `This is a string with latex \[
			frac{1}{2}
			\]`,
			expected: true,
		},
		{
			name:     "Latex $$..$$",
			input:    `This is a string with latex $$frac{1}{2}$$`,
			expected: true,
		},
		{
			name: "Latex $$..$$ - multiple lines",
			input: `This is a string with latex $$frac
			{1}{2}$$`,
			expected: true,
		},
		{
			name:     `Invalid format, start with $$ but end with \]`,
			input:    `This is a string with latex $$frac{1}{2}\]`,
			expected: false,
		},
		{
			name:     `Invalid format, start with \[ but end with \)`,
			input:    `This is a string with latex \[frac{1}{2}\)`,
			expected: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			actual := isContainsLatex(testCase.input)

			if actual != testCase.expected {
				t.Errorf("Expected: %v, but got: %v", testCase.expected, actual)
			}
		})
	}
}
