package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/eureka/configurations"
	entities "github.com/manabie-com/backend/internal/eureka/entities/items_bank"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/learnosity"
	learnosity_entity "github.com/manabie-com/backend/internal/golibs/learnosity/entity"
)

type ItemsBankRepo struct {
	LearnosityConfig *configurations.LearnosityConfig
	HTTP             learnosity.HTTP
	DataAPI          learnosity.DataAPI
}

const (
	LearnosityDomain = "localhost"
)

type ActivityResponse struct {
	Reference string `json:"reference"`
	Data      Data   `json:"data"`
}

// Items can be in both formats, more detail here: https://reference.learnosity.com/data-api/endpoints/itembank_endpoints#getActivities
type Data struct {
	Items []any `json:"items"`
}

func (i *ItemsBankRepo) ArchiveItems(ctx context.Context, itemIDs []string, uploadedQuestionID string) error {
	security := i.generateLearnositySecurity(ctx)

	items := make([]*learnosity_entity.Item, 0, len(itemIDs))
	// set items api must have at least 1 widget in the definition
	for _, itemID := range itemIDs {
		items = append(items, &learnosity_entity.Item{
			Reference: itemID,
			Status:    learnosity_entity.StatusArchived,
			Definition: learnosity_entity.Definition{
				Widgets: []learnosity_entity.Reference{
					{
						Reference: uploadedQuestionID,
					},
				},
			},
		})
	}

	result, err := i.DataAPI.Request(
		ctx,
		i.HTTP,
		learnosity.EndpointDataAPISetItems,
		security,
		learnosity.Request{
			"items": items,
		},
		learnosity.ActionSet,
	)

	if err != nil {
		return fmt.Errorf("error archiving items: %w", err)
	}

	if !result.Meta["status"].(bool) {
		return fmt.Errorf("error archiving items: %v", result.Meta)
	}

	return nil
}

func (i *ItemsBankRepo) MapItemsByActivity(ctx context.Context, organizationID string, newItemIDsByLoID map[string][]string) error {
	loIDs := []string{}
	for loID := range newItemIDsByLoID {
		loIDs = append(loIDs, loID)
	}
	currentItemIDs, err := i.GetCurrentItemIDs(ctx, loIDs)

	if err != nil {
		return err
	}

	activities := []*learnosity_entity.Activity{}
	tags := learnosity_entity.Tags{
		Tenant: []string{organizationID},
	}

	for _, loID := range loIDs {
		var itemIDs []any
		for _, id := range currentItemIDs[loID] {
			itemIDs = append(itemIDs, id)
		}
		for _, id := range newItemIDsByLoID[loID] {
			itemIDs = append(itemIDs, id)
		}
		activities = append(activities, &learnosity_entity.Activity{
			Reference: loID,
			Data: learnosity_entity.ActivityData{
				Items:         itemIDs,
				RenderingType: learnosity_entity.RenderingTypeAssess,
				Config: learnosity_entity.Config{
					Regions: "main",
				},
			},
			Tags: tags,
		})
	}

	security := i.generateLearnositySecurity(ctx)

	result, err := i.DataAPI.Request(
		context.Background(),
		i.HTTP,
		learnosity.EndpointDataAPISetActivities,
		security,
		learnosity.Request{"activities": activities},
		learnosity.ActionSet,
	)
	if err != nil {
		return err
	}

	if !result.Meta["status"].(bool) {
		return fmt.Errorf("error uploading activities: %v", result.Meta)
	}
	return nil
}

func (i *ItemsBankRepo) GetExistedIDs(ctx context.Context, itemIDs []string) (existedIDs []string, err error) {
	security := i.generateLearnositySecurity(ctx)

	result, err := i.DataAPI.Request(
		ctx,
		i.HTTP,
		learnosity.EndpointDataAPIGetItems,
		security,
		learnosity.Request{
			"references": itemIDs,
			"status": []string{
				learnosity_entity.StatusPublished,
				learnosity_entity.StatusUnpublished,
			},
		},
		learnosity.ActionGet,
	)

	if err != nil {
		return nil, err
	}

	if !result.Meta["status"].(bool) {
		return nil, fmt.Errorf("error get items: %v", result.Meta)
	}
	numberOfRecords := result.Meta.Records()
	if numberOfRecords == 0 {
		return []string{}, nil
	}

	var refsResponse []*learnosity_entity.Reference
	err = json.Unmarshal(result.Data, &refsResponse)
	if err != nil {
		return nil, err
	}

	if len(refsResponse) == 0 {
		return nil, fmt.Errorf("number of records != records value in meta")
	}

	existedIds := []string{}
	for _, ref := range refsResponse {
		existedIds = append(existedIds, ref.Reference)
	}

	return existedIds, nil
}

func (i *ItemsBankRepo) UploadContentData(ctx context.Context, organizationID string, items map[string]*entities.ItemsBankItem, questions []*entities.ItemsBankQuestion) ([]string, error) {
	learnosityQuestions := []*learnosity_entity.Question{}
	learnosityFeatures := []*learnosity_entity.Feature{}
	learnosityItems := []*learnosity_entity.Item{}

	mapQuestionRefsByItemID := map[string][]string{}
	for _, question := range questions {
		learnosityQuestion := ToLearnosityQuestion(question)
		learnosityQuestions = append(learnosityQuestions, learnosityQuestion)
		_, exists := mapQuestionRefsByItemID[question.ItemID]
		if !exists {
			mapQuestionRefsByItemID[question.ItemID] = []string{}
		}
		mapQuestionRefsByItemID[question.ItemID] = append(mapQuestionRefsByItemID[question.ItemID], learnosityQuestion.Reference)
	}

	for _, item := range items {
		learnosityFeature := ToLearnosityFeature(item)
		featureRef := ""
		if learnosityFeature != nil {
			featureRef = learnosityFeature.Reference
			learnosityFeatures = append(learnosityFeatures, learnosityFeature)
		}
		learnosityItem, err := ToLearnosityItem(item, organizationID, mapQuestionRefsByItemID[item.ItemID], featureRef)
		if err != nil {
			return nil, err
		}
		learnosityItems = append(learnosityItems, learnosityItem)
	}

	err := i.setQuestions(ctx, learnosityQuestions)
	if err != nil {
		return nil, fmt.Errorf("error uploading questions: %w", err)
	}

	err = i.setFeatures(ctx, learnosityFeatures)
	if err != nil {
		return nil, fmt.Errorf("error uploading features: %w", err)
	}

	err = i.setItems(ctx, learnosityItems)
	if err != nil {
		return nil, fmt.Errorf("error uploading items: %w", err)
	}

	generatedQuestionIDs := []string{}
	for _, question := range learnosityQuestions {
		generatedQuestionIDs = append(generatedQuestionIDs, question.Reference)
	}
	return generatedQuestionIDs, nil
}

func (i *ItemsBankRepo) setQuestions(ctx context.Context, questions []*learnosity_entity.Question) error {
	security := i.generateLearnositySecurity(ctx)
	result, err := i.DataAPI.Request(
		context.Background(),
		i.HTTP,
		learnosity.EndpointDataAPISetQuestions,
		security,
		learnosity.Request{"questions": questions},
		learnosity.ActionSet,
	)
	if err != nil {
		return err
	}

	if !result.Meta["status"].(bool) {
		return fmt.Errorf("error uploading questions: %v", result.Meta)
	}
	return nil
}

func (i *ItemsBankRepo) setFeatures(ctx context.Context, features []*learnosity_entity.Feature) error {
	security := i.generateLearnositySecurity(ctx)
	result, err := i.DataAPI.Request(
		context.Background(),
		i.HTTP,
		learnosity.EndpointDataAPISetFeatures,
		security,
		learnosity.Request{"features": features},
		learnosity.ActionSet,
	)
	if err != nil {
		return err
	}

	if !result.Meta["status"].(bool) {
		return fmt.Errorf("error uploading features: %v", result.Meta)
	}
	return nil
}

func (i *ItemsBankRepo) setItems(ctx context.Context, items []*learnosity_entity.Item) error {
	security := i.generateLearnositySecurity(ctx)
	result, err := i.DataAPI.Request(
		context.Background(),
		i.HTTP,
		learnosity.EndpointDataAPISetItems,
		security,
		learnosity.Request{"items": items},
		learnosity.ActionSet,
	)
	if err != nil {
		return err
	}

	if !result.Meta["status"].(bool) {
		return fmt.Errorf("error uploading items: %v", result.Meta)
	}
	return nil
}

func (i *ItemsBankRepo) generateLearnositySecurity(ctx context.Context) learnosity.Security {
	return learnosity.Security{
		ConsumerKey:    i.LearnosityConfig.ConsumerKey,
		Domain:         LearnosityDomain,
		Timestamp:      learnosity.FormatUTCTime(time.Now()),
		UserID:         interceptors.UserIDFromContext(ctx),
		ConsumerSecret: i.LearnosityConfig.ConsumerSecret,
	}
}

func (i *ItemsBankRepo) GetCurrentItemIDs(ctx context.Context, loIDs []string) (map[string][]string, error) {
	security := i.generateLearnositySecurity(ctx)
	result, err := i.DataAPI.Request(
		ctx,
		i.HTTP,
		learnosity.EndpointDataAPIGetActivities,
		security,
		learnosity.Request{
			"references": loIDs,
		},
		learnosity.ActionGet,
	)

	if err != nil {
		return nil, err
	}

	if !result.Meta["status"].(bool) {
		return nil, fmt.Errorf("error get items: %v", result.Meta)
	}
	numberOfRecords := result.Meta.Records()
	if numberOfRecords == 0 {
		return map[string][]string{}, nil
	}

	var activities []*ActivityResponse
	err = json.Unmarshal(result.Data, &activities)
	if err != nil {
		return nil, err
	}

	if len(activities) == 0 {
		return nil, fmt.Errorf("number of records != records value in meta")
	}

	mappedItemIDs := map[string][]string{}

	for _, activity := range activities {
		resItemIDs := []string{}
		for _, item := range activity.Data.Items {
			if ref, ok := item.(string); ok {
				resItemIDs = append(resItemIDs, ref)
			} else if obj, ok := item.(map[string]interface{}); ok {
				if reference, ok := obj["reference"].(string); ok {
					resItemIDs = append(resItemIDs, reference)
				} else {
					return nil, fmt.Errorf("item reference is not string")
				}
			}
		}

		if len(resItemIDs) != len(activity.Data.Items) {
			return nil, fmt.Errorf("num of items != num of items in activity")
		}

		mappedItemIDs[activity.Reference] = resItemIDs
	}

	return mappedItemIDs, nil
}

func (i *ItemsBankRepo) GetListItems(ctx context.Context, itemIDs []string, next *string, limit uint32) (res *learnosity.Result, err error) {
	security := i.generateLearnositySecurity(ctx)
	result, err := i.DataAPI.Request(
		ctx,
		i.HTTP,
		learnosity.EndpointDataAPIGetItems,
		security,
		learnosity.Request{
			"references": itemIDs,
			"status": []string{
				learnosity_entity.StatusPublished,
				learnosity_entity.StatusUnpublished,
			},
			"next":  next,
			"limit": limit,
		},
		learnosity.ActionGet,
	)

	if err != nil {
		return nil, err
	}

	if !result.Meta["status"].(bool) {
		return nil, fmt.Errorf("error get items: %v", result.Meta)
	}

	return result, nil
}
