package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/elastic"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/lessonmgmt/constants"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	mock_elastic "github.com/manabie-com/backend/mock/golibs/elastic"

	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/metadata"
)

func TestSearchRepo_BulkUpsert(t *testing.T) {
	t.Parallel()
	mockElastic := &mock_elastic.SearchFactory{}
	searchRepo := &SearchRepo{
		SearchFactory: mockElastic,
	}
	type testcase struct {
		name  string
		input domain.LessonSearchs
		setup func(context.Context, domain.LessonSearchs, interface{})
	}
	now := time.Now()
	cases := []testcase{
		{
			name: "success",
			input: []*domain.LessonSearch{
				{
					LessonID:       "01FK2TJDPFP8VTVBZKCH0WSX89",
					LocationID:     "location-id",
					TeachingMedium: "LESSON_TEACHING_MEDIUM_ONLINE",
					TeachingMethod: "LESSON_TEACHING_METHOD_INDIVIDUAL",
					LessonMember: []*domain.LessonMemberEs{
						{
							ID:           "01FY8Y74C2R090RPMQ2CXKWNRP",
							Name:         "Name",
							CurrentGrade: 3,
							CourseID:     "01FY8Y6EG95JR427ZR1HEDB5EN",
						},
						{
							ID:           "01FY8Y74C2R090RPMQ2CXKWNRP",
							Name:         "Name",
							CurrentGrade: 4,
							CourseID:     "01FY8Y6EG95JR427ZR1HEDB5EN",
						},
					},
					LessonTeacher: []string{"01FY8Y6EFVH5WSFZBVTNE13FP0", "01FY8Y6EFVH5WSFZBVTNE13FP0"},
					UpdatedAt:     now,
					CreatedAt:     now,
				},
			},
			setup: func(ctx context.Context, ls domain.LessonSearchs, fn interface{}) {
				mockElastic.On("BulkIndexWithResourcePath",
					mock.Anything,
					mock.MatchedBy(fn),
					"lesson").Once().Return(len(ls), nil)
			},
		},
		{
			name:  "failed",
			input: []*domain.LessonSearch{},
			setup: func(ctx context.Context, ls domain.LessonSearchs, fn interface{}) {
				mockElastic.On("BulkIndexWithResourcePath",
					mock.Anything,
					mock.MatchedBy(fn),
					"lesson").Once().Return(0, errors.New("Internal Error"))
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctx := context.Background()
			c.setup(ctx, c.input, func(args map[string]elastic.Doc) bool {
				for idx := range args {
					elastic.AssertDocIsValid(t, args[idx])
				}
				return true
			})
			gotTotal, err := searchRepo.BulkUpsert(context.Background(), c.input)
			if err != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, len(c.input), gotTotal)
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

	//mocking response parser
	mockResp, err := getfilecontent("./elastic/error.json")
	assert.NoError(t, err)
	cl, close := elastic.NewMockSearchFactory(mockResp)
	repo := SearchRepo{
		SearchFactory: cl,
	}

	assert.NoError(t, err)
	defer close()

	now := time.Now().UTC()
	args := &domain.ListLessonArgs{
		CurrentTime: now,
		SchoolID:    "5",
		Compare:     ">=",
		LessonTime:  "future",
		KeyWord:     "",
		Limit:       2,
		LessonID:    "",
	}
	rp := "manabie"
	ctx := interceptors.ContextWithJWTClaims(context.Background(), &interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{ResourcePath: rp},
	})
	ctx = metadata.NewIncomingContext(ctx, metadata.MD{"token": []string{"sometoken"}})
	_, _, _, err = repo.search(ctx, args)
	respErr := errors.Unwrap(err).(elastic.ResponseErr)
	assert.Equal(t, respErr.Errtype, "search_phase_execution_exception")
	assert.Equal(t, respErr.Reason, "all shards failed")
}

func TestSearchResponseParser(t *testing.T) {
	t.Parallel()

	//mocking response parser
	mockResp, err := getfilecontent("./elastic/lesson.json")
	assert.NoError(t, err)
	cl, close := elastic.NewMockSearchFactory(mockResp)
	repo := SearchRepo{
		SearchFactory: cl,
	}
	assert.NoError(t, err)
	defer close()
	now := time.Now().UTC()
	args := &domain.ListLessonArgs{
		CurrentTime: now,
		SchoolID:    "5",
		Compare:     ">=",
		LessonTime:  "future",
		KeyWord:     "",
		Limit:       2,
		LessonID:    "",
	}
	rp := "manabie"
	ctx := interceptors.ContextWithJWTClaims(context.Background(), &interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{ResourcePath: rp},
	})
	ctx = metadata.NewIncomingContext(ctx, metadata.MD{"token": []string{"sometoken"}})
	lessons, _, _, err := repo.Search(ctx, args)
	assert.NoError(t, err)

	getFields := func(l *domain.Lesson) []interface{} {
		return []interface{}{
			&l.LessonID,
			&l.LocationID,
			&l.CreatedAt,
			&l.UpdatedAt,
			&l.StartTime,
			&l.EndTime,
			&l.TeachingMedium,
			&l.TeachingMethod,
			&l.Teachers,
		}
	}
	for _, item := range lessons {
		for idx, field := range getFields(item) {
			assert.NotEmpty(t, field, "field %d is empty", idx)
		}
	}
}

type searchTestCase struct {
	name        string
	input       *domain.ListLessonArgs
	expectQuery string
}

func TestSearchSpecifyLessonQuery(t *testing.T) {
	t.Parallel()
	baseTime := time.Date(2022, 5, 4, 17, 26, 0, 0, time.UTC)
	baseQuery := `{"query":{"bool":{"must":[{"term":{"resource_path":""}},{"range":{"start_time":{"from":"0001-01-01T00:00:00Z","include_lower":true,"include_upper":true,"to":null}}}%s],"must_not":{"exists":{"field":"deleted_at"}}}},"size":0,"sort":[{"start_time":{"order":"asc"}},{"end_time":{"order":"asc"}},{"lesson_id":{"order":"asc"}}],"track_total_hits":true}`
	fmt.Println(baseTime)
	cases := []searchTestCase{
		{
			name: "compare >=",
			input: &domain.ListLessonArgs{
				Compare: ">=",
			},
			expectQuery: "",
		},
		{
			name: "FromDate",
			input: &domain.ListLessonArgs{
				Compare:  ">=",
				FromDate: baseTime,
			},
			expectQuery: `{"range":{"end_time":{"from":"2022-05-04T17:26:00Z","include_lower":true,"include_upper":true,"to":null}}}`,
		},
		{
			name: "ToDate",
			input: &domain.ListLessonArgs{
				Compare: ">=",
				ToDate:  baseTime,
			},
			expectQuery: `{"range":{"start_time":{"from":null,"include_lower":true,"include_upper":true,"to":"2022-05-04T17:26:00Z"}}}`,
		},
		{
			name: "KeyWord",
			input: &domain.ListLessonArgs{
				Compare: ">=",
				KeyWord: "lesson-member-name",
			},
			expectQuery: `{"bool":{"should":[{"match_phrase":{"lesson_members.name":{"query":"lesson-member-name"}}},{"wildcard":{"lesson_members.name":{"value":"*lesson-member-name*"}}}]}}`,
		},
		{
			name: "Dow",
			input: &domain.ListLessonArgs{
				Compare: ">=",
				Dow:     []domain.DateOfWeek{1, 2, 3},
			},
			expectQuery: `{"script":{"script":{"params":{"date_of_week":[1,2,3],"time_zone":""},"source":"int dateOfWeek = doc['start_time'].value.withZoneSameInstant(ZoneId.of(params.time_zone)).getDayOfWeek().getValue(); if(params.date_of_week.contains(dateOfWeek)) return true; return false;"}}}`,
		},
		{
			name: "params.FromTime",
			input: &domain.ListLessonArgs{
				Compare:  ">=",
				FromTime: "15:04:05",
			},
			expectQuery: `{"script":{"script":{"params":{"time_number":904,"time_zone":""},"source":"ZonedDateTime dateTime = doc['end_time'].value.withZoneSameInstant(ZoneId.of(params['time_zone'])); return dateTime.getHour() * 60 + dateTime.getMinute() \u003e= params['time_number']"}}}`,
		},
		{
			name: "params.ToTime",
			input: &domain.ListLessonArgs{
				Compare: ">=",
				ToTime:  "15:04:05",
			},
			expectQuery: `{"script":{"script":{"params":{"time_number":904,"time_zone":""},"source":"ZonedDateTime dateTime = doc['end_time'].value.withZoneSameInstant(ZoneId.of(params['time_zone'])); return dateTime.getHour() * 60 + dateTime.getMinute() \u003c= params['time_number']"}}}`,
		},
		{
			name: "params.LocationIDs",
			input: &domain.ListLessonArgs{
				Compare:     ">=",
				LocationIDs: []string{"center-id"},
			},
			expectQuery: `{"terms":{"location_id":["center-id"]}}`,
		},
		{
			name: "params.Teachers",
			input: &domain.ListLessonArgs{
				Compare:    ">=",
				TeacherIDs: []string{"teacher-id"},
			},
			expectQuery: `{"terms":{"lesson_teachers.keyword":["teacher-id"]}}
				`,
		},
		{
			name: "params.Students",
			input: &domain.ListLessonArgs{
				Compare:    ">=",
				StudentIDs: []string{"student-id"},
			},
			expectQuery: `{"terms":{"lesson_members.id":["student-id"]}}`,
		},
		{
			name: "params.Grades",
			input: &domain.ListLessonArgs{
				Compare: ">=",
				Grades:  []int32{1, 2, 3},
			},
			expectQuery: `{"terms":{"lesson_members.current_grade":[1,2,3]}}`,
		},
		{
			name: "params.Courses",
			input: &domain.ListLessonArgs{
				Compare:   ">=",
				CourseIDs: []string{"course-id"},
			},
			expectQuery: `{"terms":{"lesson_members.course_id":["course-id"]}}`,
		},
	}
	for _, testcase := range cases {
		input := testcase.input
		var expectQuery string
		if len(testcase.expectQuery) > 0 {
			expectQuery = fmt.Sprintf(baseQuery, ","+testcase.expectQuery)
		} else {
			expectQuery = fmt.Sprintf(baseQuery, testcase.expectQuery+"")
		}

		t.Run(testcase.name, func(t *testing.T) {
			mockClient := &mock_elastic.SearchFactory{}
			repo := SearchRepo{
				SearchFactory: mockClient,
			}

			mockClient.On("SearchUsingJwtToken", mock.Anything, constants.LessonIndexName, mock.MatchedBy(func(r *strings.Reader) bool {
				givenQuery, err := ioutil.ReadAll(r)
				assert.NoError(t, err)

				return assertJsonStringEqual(t, string(givenQuery), expectQuery)
			})).Once().Return(mockEsapiResponse(), nil)

			_, _, _, err := repo.Search(context.Background(), input)
			assert.NoError(t, err)
		})
	}
}

func TestSearchLessonQueryWithOffset(t *testing.T) {
	t.Parallel()
	baseTime := time.Date(2022, 5, 4, 17, 26, 0, 0, time.UTC)
	fmt.Println(baseTime)
	arg := &domain.ListLessonArgs{
		CurrentTime: baseTime,
		SchoolID:    "5",
		Compare:     ">=",
		LessonTime:  "future",
		CourseIDs:   []string{"course-id-1", "course-id-2"},
		TeacherIDs:  []string{"teacher-id-1", "teacher-id-2"},
		StudentIDs:  []string{"student-id-1", "student-id-2"},
		KeyWord:     "nguyen van a",
		Limit:       5,
		LessonID:    "lesson-id",
	}

	t.Run("select", func(t *testing.T) {
		mockClient := &mock_elastic.SearchFactory{}
		repo := SearchRepo{
			SearchFactory: mockClient,
		}
		mockClient.On("SearchUsingJwtToken", mock.Anything, constants.LessonIndexName, mock.Anything).
			Run(func(args mock.Arguments) {
				r := args[2].(*strings.Reader)
				givenQuery, err := ioutil.ReadAll(r)
				assert.NoError(t, err)
				expectQuery := `{"query":{"term":{"lesson_id":"lesson-id"}}}`
				a := assertJsonStringEqual(t, string(givenQuery), expectQuery)
				assert.Equal(t, true, a)
			}).Once().Return(mockEsapiResponse(), nil)

		mockClient.On("SearchUsingJwtToken", mock.Anything, constants.LessonIndexName, mock.Anything).
			Run(func(args mock.Arguments) {
				r := args[2].(*strings.Reader)
				givenQuery, err := ioutil.ReadAll(r)
				assert.NoError(t, err)
				expectQuery := `{"query":{"bool":{"must":[{"term":{"resource_path":"5"}},{"range":{"start_time":{"from":"2022-05-04T17:26:00Z","include_lower":true,"include_upper":true,"to":null}}},{"bool":{"should":[{"match_phrase":{"lesson_members.name":{"query":"nguyen van a"}}},{"wildcard":{"lesson_members.name":{"value":"*nguyen van a*"}}}]}},{"terms":{"lesson_teachers.keyword":["teacher-id-1","teacher-id-2"]}},{"terms":{"lesson_members.id":["student-id-1","student-id-2"]}},{"terms":{"lesson_members.course_id":["course-id-1","course-id-2"]}}],"must_not":{"exists":{"field":"deleted_at"}}}},"search_after":["2022-05-03T16:15:29.53051Z","2022-05-03T16:17:29.530511Z","01G26MFE7TGYHRWPXEYP6CGRC4"],"size":5,"sort":[{"start_time":{"order":"asc"}},{"end_time":{"order":"asc"}},{"lesson_id":{"order":"asc"}}],"track_total_hits":true}`

				a := assertJsonStringEqual(t, string(givenQuery), expectQuery)
				assert.Equal(t, true, a)
			}).Once().Return(mockEsapiResponse(), nil)

		mockClient.On("SearchUsingJwtToken", mock.Anything, constants.LessonIndexName, mock.Anything).
			Run(func(args mock.Arguments) {
				r := args[2].(*strings.Reader)
				givenQuery, err := ioutil.ReadAll(r)
				assert.NoError(t, err)
				expectQuery := `{"_source":{"includes":["lesson_id"]},"query":{"bool":{"must":[{"term":{"resource_path":"5"}},{"range":{"start_time":{"from":"2022-05-04T17:26:00Z","include_lower":true,"include_upper":true,"to":null}}},{"bool":{"should":[{"match_phrase":{"lesson_members.name":{"query":"nguyen van a"}}},{"wildcard":{"lesson_members.name":{"value":"*nguyen van a*"}}}]}},{"terms":{"lesson_teachers.keyword":["teacher-id-1","teacher-id-2"]}},{"terms":{"lesson_members.id":["student-id-1","student-id-2"]}},{"terms":{"lesson_members.course_id":["course-id-1","course-id-2"]}}],"must_not":{"exists":{"field":"deleted_at"}}}},"search_after":["2022-05-03T16:15:29.53051Z","2022-05-03T16:17:29.530511Z","01G26MFE7TGYHRWPXEYP6CGRC4"],"size":6,"sort":[{"start_time":{"order":"desc"}},{"end_time":{"order":"desc"}},{"lesson_id":{"order":"desc"}}]}`

				a := assertJsonStringEqual(t, string(givenQuery), expectQuery)
				assert.Equal(t, true, a)
			}).Once().Return(mockEsapiResponse(), nil)

		_, _, _, err := repo.Search(context.Background(), arg)
		assert.NoError(t, err)
	})

	t.Run("select error", func(t *testing.T) {
		mockClient := &mock_elastic.SearchFactory{}
		repo := SearchRepo{
			SearchFactory: mockClient,
		}
		mockClient.On("SearchUsingJwtToken", mock.Anything, constants.LessonIndexName, mock.Anything).
			Run(func(args mock.Arguments) {
				r := args[2].(*strings.Reader)
				givenQuery, err := ioutil.ReadAll(r)
				assert.NoError(t, err)
				expectQuery := `{"query":{"bool":{"must":{"term":{"lesson_id":"lesson-id-1"}}}}}`
				a := assertJsonStringEqual(t, string(givenQuery), expectQuery)
				assert.Equal(t, false, a)
			}).Once().Return(mockEsapiResponse(), nil)

		mockClient.On("SearchUsingJwtToken", mock.Anything, constants.LessonIndexName, mock.Anything).Once().Return(nil, fmt.Errorf("search error"))

		_, _, _, err := repo.Search(context.Background(), arg)
		assert.Error(t, err)
	})
}

func mockEsapiResponse() *esapi.Response {
	str, err := getfilecontent("./elastic/lesson.json")
	if err != nil {
		panic(err)
	}
	return &esapi.Response{
		Body: io.NopCloser(strings.NewReader(str)),
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
