package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/elastic"
	tomcons "github.com/manabie-com/backend/internal/tom/constants"
	domain "github.com/manabie-com/backend/internal/tom/domain/support"

	"github.com/elastic/go-elasticsearch/v7/esapi"
)

func (s *SearchRepo) _bulkUpsertV2(ctx context.Context, cl elastic.SearchFactory, docs []domain.SearchConversationDoc) (int, error) {
	data := make(map[string]elastic.Doc)

	var (
		totalSuccess int
		err          error
	)
	for idx, value := range docs {
		data[value.ConversationID] = elastic.NewDoc(docs[idx])
	}
	totalSuccess, err = cl.BulkIndexWithResourcePath(ctx, data, tomcons.ESConversationIndexName)
	if err != nil {
		return totalSuccess, err
	}
	if len(docs) != totalSuccess {
		return totalSuccess, fmt.Errorf("did not fully index document: want %d, has %d", len(docs), totalSuccess)
	}
	return totalSuccess, nil
}

type SearchRepo struct {
	version int
}

func (s *SearchRepo) V2() {
	s.version = 2
}

func (s *SearchRepo) BulkUpsert(ctx context.Context, cl elastic.SearchFactory, docs []domain.SearchConversationDoc) (int, error) {
	if s.version == 2 {
		return s._bulkUpsertV2(ctx, cl, docs)
	}

	data := make(map[string][]byte)

	var (
		totalSuccess int
		err          error
	)
	for _, value := range docs {
		valueMarshal, err := json.Marshal(value)
		if err != nil {
			return totalSuccess, fmt.Errorf("unable to marshal, createConversationDocuments: %w", err)
		}
		data[value.ConversationID] = valueMarshal
	}
	totalSuccess, err = cl.BulkIndex(ctx, data, "index", tomcons.ESConversationIndexName)
	if err != nil {
		return totalSuccess, err
	}
	if len(docs) != totalSuccess {
		return totalSuccess, fmt.Errorf("did not fully index document: want %d, has %d", len(docs), totalSuccess)
	}
	return totalSuccess, nil
}

type ConversationUpsertPayload struct {
	Doc struct {
		IsReplied   *bool                     `json:"is_replied,omitempty"`
		LastMessage *domain.SearchLastMessage `json:"last_message,omitempty"`
		CourseIDs   *JSONArr                  `json:"course_ids,omitempty"`
		UserIDs     *JSONArr                  `json:"user_ids,omitempty"`
	} `json:"doc"`
}

// wrap a slide and used as pointer
// if null, fields is ignore, if has value, even if the value is empty, it still include the field to upsert
// we can have "some_arr": null/[] instead of not included at all
type JSONArr struct {
	internal []string
}

func (j JSONArr) MarshalJSON() ([]byte, error) {
	return json.Marshal(j.internal)
}

func (s *SearchRepo) UpsertFields(ctx context.Context, cl elastic.SearchFactory, item domain.UpdateItems) error {
	doc := ConversationUpsertPayload{}

	if item.Replied.NotNull {
		isReply := item.Replied.Bool
		doc.Doc.IsReplied = &isReply
	}
	if item.LatestMessageUpdateTime != nil {
		doc.Doc.LastMessage = &domain.SearchLastMessage{
			UpdatedAt: *item.LatestMessageUpdateTime,
		}
	}
	if item.Courses.NotNull {
		doc.Doc.CourseIDs = &JSONArr{item.Courses.StrArr}
	}
	if item.UserIDs.NotNull {
		doc.Doc.UserIDs = &JSONArr{item.UserIDs.StrArr}
	}
	bs, err := json.Marshal(doc)
	if err != nil {
		return err
	}
	res, err := cl.UpdateCtx(ctx, tomcons.ESConversationIndexName, item.ConversationID, strings.NewReader(string(bs)))
	if err != nil {
		return err
	}
	defer res.Body.Close()
	return elastic.CheckResponse(res)
}

func (s *SearchRepo) BulkUpsertFields(ctx context.Context, cl elastic.SearchFactory, items []domain.UpdateItems) error {
	data := map[string][]byte{}
	for _, item := range items {
		doc := ConversationUpsertPayload{}

		if item.Replied.NotNull {
			isReply := item.Replied.Bool
			doc.Doc.IsReplied = &isReply
		}
		if item.LatestMessageUpdateTime != nil {
			doc.Doc.LastMessage = &domain.SearchLastMessage{
				UpdatedAt: *item.LatestMessageUpdateTime,
			}
		}
		if item.Courses.NotNull {
			doc.Doc.CourseIDs = &JSONArr{item.Courses.StrArr}
		}
		if item.UserIDs.NotNull {
			doc.Doc.UserIDs = &JSONArr{item.UserIDs.StrArr}
		}
		bs, err := json.Marshal(doc)
		if err != nil {
			return err
		}
		data[item.ConversationID] = bs
	}
	_, err := cl.BulkIndex(ctx, data, "update", tomcons.ESConversationIndexName)
	return err
}

func (s *SearchRepo) Search(ctx context.Context, client elastic.SearchFactory, filter domain.ConversationFilter) (ret []domain.SearchConversationDoc, err error) {
	boolQ := elastic.NewBoolQuery()
	if filter.ConversationName.NotNull {
		// not sure if language is in domain or in repo
		boolQ.Must(
			elastic.NewMultiMatchQuery(filter.ConversationName.Str, "conversation_name.english", "conversation_name.japanese").Type("most_fields").
				Operator("and"),
		)
	}
	if filter.AccessPaths.NotNull {
		boolQ.Filter(elastic.NewTermsQuery("access_paths", filter.AccessPaths.ToInterfaces()...))
	}
	if filter.RepliedStatus.NotNull {
		boolQ.Filter(elastic.NewTermQuery("is_replied", filter.RepliedStatus.Bool))
	}
	// version 2 uses DLS, which already filters by resource_path using DLS policy
	if filter.School.NotNull && s.version != 2 {
		boolQ.Filter(elastic.NewTermsQuery("owner", filter.School.ToInterfaces()...))
	}
	if filter.JoinStatus.NotNull {
		switch filter.JoinStatus.Bool {
		case true:
			boolQ.Must(elastic.NewTermsQuery("user_ids.keyword", filter.UserID))
		case false:
			boolQ.MustNot(elastic.NewTermsQuery("user_ids.keyword", filter.UserID))
		}
	}
	if filter.Courses.NotNull {
		boolQ.Filter(elastic.NewTermsQuery("course_ids.keyword", filter.Courses.ToInterfaces()...))
	}
	if filter.ConversationTypes.NotNull {
		boolQ.Filter(elastic.NewTermsQuery("conversation_type", filter.ConversationTypes.ToInterfaces()...))
	}
	source := elastic.NewSearchSource().Query(boolQ)
	if filter.Limit.NotNull {
		source.Size(int(filter.Limit.I64))
	}
	for _, item := range filter.SortBy {
		switch item.Key {
		case domain.SortKey_LatestMsgTime:
			source.Sort("last_message.updated_at", item.Asc)
		case domain.SortKey_ConversationID:
			source.Sort("conversation_id", item.Asc)
		}
	}
	if filter.OffsetTime.NotNull {
		source.SearchAfter(filter.OffsetTime.I64)
	}
	if filter.OffsetConverstionID.NotNull {
		source.SearchAfter(filter.OffsetConverstionID.Str)
	}

	var (
		res *esapi.Response
	)
	switch s.version {
	case 2:
		res, err = elastic.DoSearchFromSourceUsingJwtToken(ctx, client, tomcons.ESConversationIndexName, source)
	default:
		res, err = elastic.DoSearchFromSource(ctx, client, tomcons.ESConversationIndexName, source)
	}
	if err != nil {
		return nil, err
	}
	err = elastic.ParseSearchResponse(res.Body, func(hit *elastic.SearchHit) error {
		var t domain.SearchConversationDoc
		err = json.Unmarshal(hit.Source, &t)
		if err != nil {
			return fmt.Errorf("json.Unmarshal: %w", err)
		}
		ret = append(ret, t)
		return nil
	})

	return
}

func (s *SearchRepo) SearchV2(ctx context.Context, client elastic.SearchFactory, filter domain.ConversationFilter) (ret []domain.SearchConversationDoc, err error) {
	boolQ := elastic.NewBoolQuery()
	if filter.ConversationName.NotNull {
		// not sure if language is in domain or in repo
		boolQ.Must(
			elastic.NewMultiMatchQuery(filter.ConversationName.Str, "conversation_name.english", "conversation_name.japanese").Type("most_fields").
				Operator("and"),
		)
	}
	if len(filter.LocationConfigs) > 0 {
		boolQ2 := elastic.NewBoolQuery()
		// AND (
		//	(conversation_type='student' AND access_paths='loc1')
		//		OR
		//	(conversation_type='parent' AND access_paths='loc2')
		// )
		for _, config := range filter.LocationConfigs {
			boolQ2.Should(
				elastic.NewBoolQuery().Must(
					elastic.NewTermsQuery("conversation_type", config.ConversationType.Str),
					elastic.NewTermsQuery("access_paths", config.AccessPaths.ToInterfaces()...),
				),
			)
		}
		boolQ.Must(boolQ2)
	}
	if filter.AccessPaths.NotNull {
		boolQ.Filter(elastic.NewTermsQuery("access_paths", filter.AccessPaths.ToInterfaces()...))
	}
	if filter.RepliedStatus.NotNull {
		boolQ.Filter(elastic.NewTermQuery("is_replied", filter.RepliedStatus.Bool))
	}
	// version 2 uses DLS, which already filters by resource_path using DLS policy
	if filter.School.NotNull && s.version != 2 {
		boolQ.Filter(elastic.NewTermsQuery("owner", filter.School.ToInterfaces()...))
	}
	if filter.JoinStatus.NotNull {
		switch filter.JoinStatus.Bool {
		case true:
			boolQ.Must(elastic.NewTermsQuery("user_ids.keyword", filter.UserID))
		case false:
			boolQ.MustNot(elastic.NewTermsQuery("user_ids.keyword", filter.UserID))
		}
	}
	if filter.Courses.NotNull {
		boolQ.Filter(elastic.NewTermsQuery("course_ids.keyword", filter.Courses.ToInterfaces()...))
	}
	if filter.ConversationTypes.NotNull {
		boolQ.Filter(elastic.NewTermsQuery("conversation_type", filter.ConversationTypes.ToInterfaces()...))
	}
	source := elastic.NewSearchSource().Query(boolQ)
	if filter.Limit.NotNull {
		source.Size(int(filter.Limit.I64))
	}
	for _, item := range filter.SortBy {
		switch item.Key {
		case domain.SortKey_LatestMsgTime:
			source.Sort("last_message.updated_at", item.Asc)
		case domain.SortKey_ConversationID:
			source.Sort("conversation_id", item.Asc)
		}
	}
	if filter.OffsetTime.NotNull {
		source.SearchAfter(filter.OffsetTime.I64)
	}
	if filter.OffsetConverstionID.NotNull {
		source.SearchAfter(filter.OffsetConverstionID.Str)
	}

	var (
		res *esapi.Response
	)
	switch s.version {
	case 2:
		res, err = elastic.DoSearchFromSourceUsingJwtToken(ctx, client, tomcons.ESConversationIndexName, source)
	default:
		res, err = elastic.DoSearchFromSource(ctx, client, tomcons.ESConversationIndexName, source)
	}
	if err != nil {
		return nil, err
	}
	err = elastic.ParseSearchResponse(res.Body, func(hit *elastic.SearchHit) error {
		var t domain.SearchConversationDoc
		err = json.Unmarshal(hit.Source, &t)
		if err != nil {
			return fmt.Errorf("json.Unmarshal: %w", err)
		}
		ret = append(ret, t)
		return nil
	})

	return
}
