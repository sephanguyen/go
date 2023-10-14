package repositories

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/elastic"
	"github.com/manabie-com/backend/internal/golibs/types"
	tomcons "github.com/manabie-com/backend/internal/tom/constants"
	domain "github.com/manabie-com/backend/internal/tom/domain/support"
	mock_elastic "github.com/manabie-com/backend/mock/golibs/elastic"
	tpb "github.com/manabie-com/backend/pkg/manabuf/tom/v1"

	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type TestCase struct {
	name       string
	input      interface{}
	expect     interface{}
	err        error
	resp       interface{}
	statusCode int
}

func Test_SearchV2(t *testing.T) {
	t.Parallel()
	mockElastic := &mock_elastic.SearchFactory{}
	searcRepo := &SearchRepo{
		version: 2,
	}
	lastestMsg, err := time.Parse(time.RFC3339, "2021-11-05T09:31:42+07:00")
	assert.NoError(t, err)

	cases := []TestCase{
		{
			name: "all fields",
			input: []domain.SearchConversationDoc{
				{
					ConversationID:           "someid",
					ConversationNameEnglish:  "John",
					ConversationNameJapanese: "ドラえもん",
					CourseIDs:                []string{"course 1"},
					UserIDs:                  []string{"user-1"},
					LastMessage: domain.SearchLastMessage{
						UpdatedAt: lastestMsg,
					},
					IsReplied:        true,
					Owner:            "school 1",
					ConversationType: tpb.ConversationType_CONVERSATION_PARENT.String(),
				},
			},
		},
		{
			name: "multiple update items",
			input: []domain.SearchConversationDoc{
				{
					ConversationID:           "conv-1",
					ConversationNameEnglish:  "John",
					ConversationNameJapanese: "ドラえもん",
					CourseIDs:                []string{"course 1"},
					UserIDs:                  []string{"user-1"},
					LastMessage: domain.SearchLastMessage{
						UpdatedAt: lastestMsg,
					},
					IsReplied:        true,
					Owner:            "school 1",
					ConversationType: tpb.ConversationType_CONVERSATION_PARENT.String(),
				},
				{
					ConversationID:           "conv-2",
					ConversationNameEnglish:  "John 2",
					ConversationNameJapanese: "ドラえもん2",
					CourseIDs:                []string{"course 2"},
					UserIDs:                  []string{"user 2"},
					LastMessage: domain.SearchLastMessage{
						UpdatedAt: lastestMsg,
					},
					IsReplied:        false,
					Owner:            "school 2",
					ConversationType: tpb.ConversationType_CONVERSATION_STUDENT.String(),
				},
			},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			mockElastic.On("BulkIndexWithResourcePath", mock.Anything, mock.MatchedBy(func(datas map[string]elastic.Doc) bool {
				for idx := range datas {
					elastic.AssertDocIsValid(t, datas[idx])
				}
				return true
			}), tomcons.ESConversationIndexName).Once().Return(len(c.input.([]domain.SearchConversationDoc)), nil)
			_, err := searcRepo.BulkUpsert(context.Background(), mockElastic, c.input.([]domain.SearchConversationDoc))
			if err != nil {
				assert.ErrorIs(t, err, c.err)
			}
		})
	}
}

func Test_Search_BulkUpsert(t *testing.T) {
	t.Parallel()
	mockElastic := &mock_elastic.SearchFactory{}
	searcRepo := &SearchRepo{}
	lastestMsg, err := time.Parse(time.RFC3339, "2021-11-05T09:31:42+07:00")
	assert.NoError(t, err)

	cases := []TestCase{
		{
			name: "all fields",
			input: []domain.SearchConversationDoc{
				{
					ConversationID:           "someid",
					ConversationNameEnglish:  "John",
					ConversationNameJapanese: "ドラえもん",
					CourseIDs:                []string{"course 1"},
					UserIDs:                  []string{"user-1"},
					LastMessage: domain.SearchLastMessage{
						UpdatedAt: lastestMsg,
					},
					IsReplied:        true,
					Owner:            "school 1",
					ConversationType: tpb.ConversationType_CONVERSATION_PARENT.String(),
					AccessPath:       []string{"location1/location2", "location3/location4"},
				},
			},
			expect: map[string]string{
				"someid": `{"conversation_id":"someid","conversation_name.english":"John","conversation_name.japanese":"ドラえもん","course_ids":["course 1"],"user_ids":["user-1"],"last_message":{"updated_at":"2021-11-05T09:31:42+07:00"},"is_replied":true,"owner":"school 1","conversation_type":"CONVERSATION_PARENT","access_paths":["location1/location2","location3/location4"]}`,
			},
		},
		{
			name: "multiple update items",
			input: []domain.SearchConversationDoc{
				{
					ConversationID:           "conv-1",
					ConversationNameEnglish:  "John",
					ConversationNameJapanese: "ドラえもん",
					CourseIDs:                []string{"course 1"},
					UserIDs:                  []string{"user-1"},
					LastMessage: domain.SearchLastMessage{
						UpdatedAt: lastestMsg,
					},
					IsReplied:        true,
					Owner:            "school 1",
					ConversationType: tpb.ConversationType_CONVERSATION_PARENT.String(),
				},
				{
					ConversationID:           "conv-2",
					ConversationNameEnglish:  "John 2",
					ConversationNameJapanese: "ドラえもん2",
					CourseIDs:                []string{"course 2"},
					UserIDs:                  []string{"user 2"},
					LastMessage: domain.SearchLastMessage{
						UpdatedAt: lastestMsg,
					},
					IsReplied:        false,
					Owner:            "school 2",
					ConversationType: tpb.ConversationType_CONVERSATION_STUDENT.String(),
				},
			},
			expect: map[string]string{
				"conv-1": `{"conversation_id":"conv-1","conversation_name.english":"John","conversation_name.japanese":"ドラえもん","course_ids":["course 1"],"user_ids":["user-1"],"last_message":{"updated_at":"2021-11-05T09:31:42+07:00"},"is_replied":true,"owner":"school 1","conversation_type":"CONVERSATION_PARENT","access_paths":null}`,
				"conv-2": `{"conversation_id":"conv-2","conversation_name.english":"John 2","conversation_name.japanese":"ドラえもん2","course_ids":["course 2"],"user_ids":["user 2"],"last_message":{"updated_at":"2021-11-05T09:31:42+07:00"},"is_replied":false,"owner":"school 2","conversation_type":"CONVERSATION_STUDENT","access_paths":null}`,
			},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			mockElastic.On("BulkIndex", mock.Anything, mock.MatchedBy(func(datas map[string][]byte) bool {
				for idx, data := range datas {
					expect := c.expect.(map[string]string)[idx]
					equal := assertJsonStringEqual(t, string(data), expect)
					if !equal {
						return false
					}
				}
				return true
			}), "index", tomcons.ESConversationIndexName).Once().Return(len(c.expect.(map[string]string)), nil)
			_, err := searcRepo.BulkUpsert(context.Background(), mockElastic, c.input.([]domain.SearchConversationDoc))
			if err != nil {
				assert.ErrorIs(t, err, c.err)
			}
		})
	}
}

func Test_Search_UpsertFields(t *testing.T) {
	t.Parallel()
	searcRepo := &SearchRepo{}
	lastestMsg, err := time.Parse(time.RFC3339, "2021-11-05T09:31:42+07:00")
	assert.NoError(t, err)
	errJson := `{"error":{"root_cause":[{"type":"document_missing_exception","reason":"[_doc][not exist]: document missing","index_uuid":"jvutXkG7STu8eHqpcJI3KA","shard":"0","index":"conversations"}],"type":"document_missing_exception","reason":"[_doc][not exist]: document missing","index_uuid":"jvutXkG7STu8eHqpcJI3KA","shard":"0","index":"conversations"},"status":404}`
	successJson := `{"_shards":{"total":0,"successful":0,"failed":0},"_index":"test","_type":"_doc","_id":"1","_version":2,"_primary_term":1,"_seq_no":1,"result":"noop"}`

	errGen := func() error {
		expectErr := elastic.CheckResponse(&esapi.Response{
			StatusCode: 404,
			Body:       io.NopCloser(strings.NewReader(errJson)),
		})
		return expectErr
	}

	cases := []TestCase{
		{
			name: "document missing error",
			input: domain.UpdateItems{
				ConversationID:          "someid",
				Courses:                 types.NewStrArr([]string{"courses"}),
				UserIDs:                 types.NewStrArr([]string{"user 1"}),
				Replied:                 types.NewBool(true),
				LatestMessageUpdateTime: &lastestMsg,
			},
			expect:     `{"doc":{"is_replied":true,"last_message":{"updated_at":"2021-11-05T09:31:42+07:00"},"course_ids":["courses"],"user_ids":["user 1"]}}`,
			resp:       errJson,
			err:        errGen(),
			statusCode: 404,
		},
		{
			name: "all fields",
			input: domain.UpdateItems{
				ConversationID:          "someid",
				Courses:                 types.NewStrArr([]string{"courses"}),
				UserIDs:                 types.NewStrArr([]string{"user 1"}),
				Replied:                 types.NewBool(true),
				LatestMessageUpdateTime: &lastestMsg,
			},
			expect:     `{"doc":{"is_replied":true,"last_message":{"updated_at":"2021-11-05T09:31:42+07:00"},"course_ids":["courses"],"user_ids":["user 1"]}}`,
			resp:       successJson,
			statusCode: 200,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			mockElastic := &mock_elastic.SearchFactory{}
			mockElastic.On("UpdateCtx", mock.Anything, mock.Anything, mock.Anything, mock.MatchedBy(func(r *strings.Reader) bool {
				data, err := ioutil.ReadAll(r)
				assert.NoError(t, err)
				expect := c.expect.(string)
				equal := assertJsonStringEqual(t, string(data), expect)
				return equal
			})).Once().Return(&esapi.Response{
				StatusCode: c.statusCode,
				Body:       io.NopCloser(strings.NewReader(c.resp.(string))),
			}, nil)
			err := searcRepo.UpsertFields(context.Background(), mockElastic, c.input.(domain.UpdateItems))
			if err != nil {
				assert.Equal(t, err, c.err)
			}
		})
	}
}

func Test_Search_BulkUpsertFields(t *testing.T) {
	t.Parallel()
	mockElastic := &mock_elastic.SearchFactory{}
	searcRepo := &SearchRepo{}
	lastestMsg, err := time.Parse(time.RFC3339, "2021-11-05T09:31:42+07:00")
	assert.NoError(t, err)

	cases := []TestCase{
		{
			name: "all fields",
			input: []domain.UpdateItems{
				{
					ConversationID:          "someid",
					Courses:                 types.NewStrArr([]string{"courses"}),
					UserIDs:                 types.NewStrArr([]string{"user 1"}),
					Replied:                 types.NewBool(true),
					LatestMessageUpdateTime: &lastestMsg,
				},
			},
			expect: map[string]string{
				"someid": `{"doc":{"is_replied":true,"last_message":{"updated_at":"2021-11-05T09:31:42+07:00"},"course_ids":["courses"],"user_ids":["user 1"]}}`,
			},
		},
		{
			name: "multiple update items",
			input: []domain.UpdateItems{
				{
					ConversationID: "conv-1",
				},
				{
					ConversationID:          "conv-2",
					LatestMessageUpdateTime: &lastestMsg,
				},
			},
			expect: map[string]string{
				"conv-1": `{"doc":{}}`,
				"conv-2": `{"doc":{"last_message":{"updated_at":"2021-11-05T09:31:42+07:00"}}}`,
			},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			mockElastic.On("BulkIndex", mock.Anything, mock.MatchedBy(func(datas map[string][]byte) bool {
				for idx, data := range datas {
					expect := c.expect.(map[string]string)[idx]
					equal := assertJsonStringEqual(t, string(data), expect)
					if !equal {
						return false
					}
				}
				return true
			}), "update", tomcons.ESConversationIndexName).Once().Return(0, nil)
			err := searcRepo.BulkUpsertFields(context.Background(), mockElastic, c.input.([]domain.UpdateItems))
			if err != nil {
				assert.ErrorIs(t, err, c.err)
			}
		})
	}
}

func getfilecontent(name string) (string, error) {
	f, err := os.Open(name)
	if err != nil {
		return "", err
	}
	defer f.Close()
	bs, err := ioutil.ReadAll(f)
	if err != nil {
		return "", err
	}
	return string(bs), nil
}

func TestSearchError(t *testing.T) {
	t.Parallel()
	repo := SearchRepo{}
	//mocking response parser
	mockResp, err := getfilecontent("./elastic/error.json")
	assert.NoError(t, err)
	cl, close := elastic.NewMockSearchFactory(mockResp)
	assert.NoError(t, err)
	defer close()
	_, err = repo.Search(context.Background(), cl, domain.ConversationFilter{})
	respErr := errors.Unwrap(err).(elastic.ResponseErr)
	assert.Equal(t, respErr.Errtype, "search_phase_execution_exception")
	assert.Equal(t, respErr.Reason, "all shards failed")
}
func TestSearchResponseParser(t *testing.T) {
	t.Parallel()
	repo := SearchRepo{}

	//mocking response parser
	mockResp, err := getfilecontent("./elastic/conversation.json")
	assert.NoError(t, err)
	cl, close := elastic.NewMockSearchFactory(mockResp)
	assert.NoError(t, err)
	defer close()
	doc, err := repo.Search(context.Background(), cl, domain.ConversationFilter{})
	assert.NoError(t, err)
	for _, item := range doc {
		fields := item.GetFields()
		for idx, field := range fields {
			assert.NotEmpty(t, field, "field %d is empty", idx)
		}
	}
}

type searchTestCase struct {
	name        string
	input       domain.ConversationFilter
	expectQuery string
}

func TestSearchConversationQuery(t *testing.T) {
	t.Parallel()
	cases := []searchTestCase{
		{
			name: "reply status",
			input: domain.ConversationFilter{
				RepliedStatus: types.NewBool(true),
			},
			expectQuery: `
		{"query":{"bool":{"filter":{"term":{"is_replied":true}}}}}
			`,
		},
		{
			name: "join status",
			input: domain.ConversationFilter{
				JoinStatus: types.NewBool(false),
			},
			expectQuery: `
{"query":{"bool":{"must_not":{"terms":{"user_ids.keyword":[""]}}}}}
				`,
		},
		{
			name: "search by name",
			input: domain.ConversationFilter{
				ConversationName: types.NewStr("some name"),
			},
			expectQuery: `
{
    "query": {
        "bool": {
            "must": {
                "multi_match": {
                    "fields": [
                        "conversation_name.english",
                        "conversation_name.japanese"
                    ],
					"operator": "and",
                    "query": "some name",
                    "type": "most_fields"
                }
            }
        }
    }
}
			`,
		},
		{
			name: "full filter",
			input: domain.ConversationFilter{
				UserID:            "some id",
				RepliedStatus:     types.NewBool(false),
				JoinStatus:        types.NewBool(true),
				ConversationName:  types.NewStr("some name"),
				School:            types.NewStrArr([]string{"school 1"}),
				ConversationTypes: types.NewStrArr([]string{"student conv"}),
				Courses:           types.NewStrArr([]string{"course 1"}),
				AccessPaths:       types.NewStrArr([]string{"orgloc/loc1/loc3", "orgloc/loc1/loc2"}),
			},
			expectQuery: `
{
    "query": {
        "bool": {
            "filter": [
				{
					"terms":  {
						"access_paths":["orgloc/loc1/loc3","orgloc/loc1/loc2"]
					}
				},
                {
                    "term": {
                        "is_replied": false
                    }
                },
                {
                    "terms": {
                        "owner": [
                            "school 1"
                        ]
                    }
                },
                {
                    "terms": {
                        "course_ids.keyword": [
                            "course 1"
                        ]
                    }
                },
                {
                    "terms": {
                        "conversation_type": [
                            "student conv"
                        ]
                    }
                }
            ],
            "must": [
                {
                    "multi_match": {
                        "fields": [
                            "conversation_name.english",
                            "conversation_name.japanese"
                        ],
						"operator": "and",
                        "query": "some name",
                        "type": "most_fields"
                    }
                },
                {
                    "terms": {
                        "user_ids.keyword": [
                            "some id"
                        ]
                    }
                }
            ]
        }
    }
}
`,
		},
		{
			name: "paging",
			input: domain.ConversationFilter{
				SortBy: []domain.ConversationSortItem{
					{
						Key: domain.SortKey_LatestMsgTime,
					},
					{
						Key: domain.SortKey_ConversationID,
					},
				},
				OffsetTime:          types.NewInt64(strTimeToInt64("2021-10-28T06:49:33.651862Z")),
				OffsetConverstionID: types.NewStr("some id"),
				Limit:               types.NewInt64(100),
			},
			expectQuery: `
{
    "query": {
        "bool": {}
    },
    "search_after": [
        1635403773,
        "some id"
    ],
    "size": 100,
    "sort": [
        {
            "last_message.updated_at": {
                "order": "desc"
            }
        },
        {
            "conversation_id": {
                "order": "desc"
            }
        }
    ]
}
`,
		},
	}
	for _, testcase := range cases {
		input := testcase.input
		expectQuery := testcase.expectQuery
		t.Run(testcase.name, func(t *testing.T) {
			repo := SearchRepo{}
			mockClient := &mock_elastic.SearchFactory{}
			mockClient.On("Search", mock.Anything, tomcons.ESConversationIndexName, mock.MatchedBy(func(r *strings.Reader) bool {
				givenQuery, err := ioutil.ReadAll(r)
				assert.NoError(t, err)
				return assertJsonStringEqual(t, string(givenQuery), expectQuery)
			})).Once().Return(mockEsapiResponse(), nil)

			_, err := repo.Search(context.Background(), mockClient, input)
			assert.NoError(t, err)
		})
	}
}
func assertJsonStringEqual(t *testing.T, has, want string) bool {
	buffer := new(bytes.Buffer)
	if err := json.Compact(buffer, []byte(want)); err != nil {
		assert.NoError(t, err, "compacting json %s", want)
	}

	str := strings.ReplaceAll(string(has), "\\\"", "")
	return bytes.Equal(buffer.Bytes(), []byte(str))
}

func mockEsapiResponse() *esapi.Response {
	str, err := getfilecontent("./elastic/conversation.json")
	if err != nil {
		panic(err)
	}
	return &esapi.Response{
		Body: io.NopCloser(strings.NewReader(str)),
	}
}

func strTimeToInt64(str string) int64 {
	time, err := time.Parse(time.RFC3339, str)
	if err != nil {
		panic(err)
	}
	return time.Unix()
}

func TestSearchConversationQuery2(t *testing.T) {
	t.Parallel()
	cases := []searchTestCase{
		{
			name: "full filter",
			input: domain.ConversationFilter{
				UserID:            "some id",
				RepliedStatus:     types.NewBool(false),
				JoinStatus:        types.NewBool(true),
				ConversationName:  types.NewStr("some name"),
				School:            types.NewStrArr([]string{"school 1"}),
				ConversationTypes: types.NewStrArr([]string{"student conv"}),
				Courses:           types.NewStrArr([]string{"course 1"}),
				AccessPaths:       types.NewStrArr([]string{"orgloc/loc1/loc3", "orgloc/loc1/loc2"}),
				LocationConfigs: []domain.LocationConfigFilter{
					{
						ConversationType: types.NewStr("student conv"),
						AccessPaths:      types.NewStrArr([]string{"orgloc/loc1/loc3", "orgloc/loc1/loc2"}),
					},
					{
						ConversationType: types.NewStr("parent conv"),
						AccessPaths:      types.NewStrArr([]string{"orgloc/loc1/loc3", "orgloc/loc1/loc2"}),
					},
				},
			},
			expectQuery: `
{
    "query": {
        "bool": {
            "filter": [
				{
					"terms":  {
						"access_paths":["orgloc/loc1/loc3","orgloc/loc1/loc2"]
					}
				},
                {
                    "term": {
                        "is_replied": false
                    }
                },
                {
                    "terms": {
                        "owner": [
                            "school 1"
                        ]
                    }
                },
                {
                    "terms": {
                        "course_ids.keyword": [
                            "course 1"
                        ]
                    }
                },
                {
                    "terms": {
                        "conversation_type": [
                            "student conv"
                        ]
                    }
                }
            ],
            "must": [
                {
                    "multi_match": {
                        "fields": [
                            "conversation_name.english",
                            "conversation_name.japanese"
                        ],
						"operator": "and",
                        "query": "some name",
                        "type": "most_fields"
                    }
                },
				{
					"bool": {
					  "should": [
						{
						  "bool": {
							"must": [
							  {
								"terms": {
								  "conversation_type": [
									"student conv"
								  ]
								}
							  },
							  {
								"terms": {
								  "access_paths": [
									"orgloc/loc1/loc3",
									"orgloc/loc1/loc2"
								  ]
								}
							  }
							]
						  }
						},
						{
						  "bool": {
							"must": [
							  {
								"terms": {
								  "conversation_type": [
									"parent conv"
								  ]
								}
							  },
							  {
								"terms": {
								  "access_paths": [
									"orgloc/loc1/loc3",
									"orgloc/loc1/loc2"
								  ]
								}
							  }
							]
						  }
						}
					  ]
					}
				  },
                {
                    "terms": {
                        "user_ids.keyword": [
                            "some id"
                        ]
                    }
                }
            ]
        }
    }
}
`,
		},
	}
	for _, testcase := range cases {
		input := testcase.input
		expectQuery := testcase.expectQuery
		t.Run(testcase.name, func(t *testing.T) {
			repo := SearchRepo{}
			mockClient := &mock_elastic.SearchFactory{}
			mockClient.On("Search", mock.Anything, tomcons.ESConversationIndexName, mock.MatchedBy(func(r *strings.Reader) bool {
				givenQuery, err := ioutil.ReadAll(r)
				assert.NoError(t, err)
				return assertJsonStringEqual(t, string(givenQuery), expectQuery)
			})).Once().Return(mockEsapiResponse(), nil)

			_, err := repo.SearchV2(context.Background(), mockClient, input)
			assert.NoError(t, err)
		})
	}
}
