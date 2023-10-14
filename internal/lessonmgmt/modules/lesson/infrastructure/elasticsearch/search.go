package elasticsearch

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/elastic"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/lessonmgmt/constants"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"

	"github.com/elastic/go-elasticsearch/v7/esapi"
	o_elastic "github.com/olivere/elastic/v7"
)

type SearchRepo struct {
	SearchFactory elastic.SearchFactory
}

func (s *SearchRepo) BulkUpsert(ctx context.Context, lessonDocs domain.LessonSearchs) (int, error) {
	ret := NewLessonSearchsFromEntities(lessonDocs)

	return s.bulkUpsert(ctx, *ret)
}

func (s *SearchRepo) bulkUpsert(ctx context.Context, lessonDocs LessonSearchs) (int, error) {
	ctx, span := interceptors.StartSpan(ctx, "SearchRepo.BulkUpsert")
	defer span.End()
	data := make(map[string]elastic.Doc)
	var (
		totalSuccess int
		err          error
	)
	for idx, value := range lessonDocs {
		data[value.LessonID] = elastic.NewDoc(lessonDocs[idx])
	}
	totalSuccess, err = s.SearchFactory.BulkIndexWithResourcePath(ctx, data, constants.LessonIndexName)
	if err != nil {
		return totalSuccess, fmt.Errorf("failed to bulk index lesson document to elasticsearch: %w", err)
	}
	return totalSuccess, nil
}

const timeFormat = "15:04:05"

func (s *SearchRepo) buildLessonQuery(ctx context.Context, params *domain.ListLessonArgs) (boolQ *o_elastic.BoolQuery, isAsc bool, err error) {
	boolQ = elastic.NewBoolQuery()
	boolQ.MustNot(elastic.NewExistQuery("deleted_at"))
	boolQ.Must(elastic.NewTermQuery("resource_path", params.SchoolID))
	isAsc = false

	errorResult := func(err error) (*o_elastic.BoolQuery, bool, error) {
		return boolQ, isAsc, err
	}
	if params.Compare == ">=" {
		isAsc = true
		boolQ.Must(elastic.NewRangeQuery("start_time").Gte(params.CurrentTime))
	} else {
		boolQ.Must(elastic.NewRangeQuery("start_time").Lt(params.CurrentTime))
	}

	if !params.FromDate.IsZero() {
		boolQ.Must(elastic.NewRangeQuery("end_time").Gte(params.FromDate))
	}

	if !params.ToDate.IsZero() {
		boolQ.Must(elastic.NewRangeQuery("start_time").Lte(params.ToDate))
	}

	if len(params.KeyWord) > 0 {
		boolQKeyWork := elastic.NewBoolQuery()
		boolQKeyWork.Should(elastic.NewMatchPhraseQuery("lesson_members.name", params.KeyWord),
			elastic.NewWildcardQuery("lesson_members.name", "*"+params.KeyWord+"*"))
		boolQ.Must(boolQKeyWork)
	}

	if len(params.Dow) > 0 {
		dow := params.Dow
		for i, item := range dow {
			if item == 0 {
				dow[i] = 7
			}
		}

		scriptQuery, err := elastic.NewScriptQuery("int dateOfWeek = doc['start_time'].value.withZoneSameInstant(ZoneId.of(params.time_zone)).getDayOfWeek().getValue(); if(params.date_of_week.contains(dateOfWeek)) return true; return false;", map[string]interface{}{
			"time_zone":    params.TimeZone,
			"date_of_week": dow,
		})
		if err != nil {
			return errorResult(fmt.Errorf("elastic.NewScriptQuery params.Dow: %w", err))
		}
		boolQ.Must(scriptQuery)
	}

	if params.FromTime != "" {
		t, err := time.Parse(timeFormat, params.FromTime)
		if err != nil {
			return errorResult(fmt.Errorf("params.FromTime cannot parse string to time with format 15:04:05(H:mm:ss)"))
		}
		scriptQuery, err := elastic.NewScriptQuery("ZonedDateTime dateTime = doc['end_time'].value.withZoneSameInstant(ZoneId.of(params['time_zone'])); return dateTime.getHour() * 60 + dateTime.getMinute() >= params['time_number']", map[string]interface{}{
			"time_zone":   params.TimeZone,
			"time_number": t.Hour()*60 + t.Minute(),
		})

		if err != nil {
			return errorResult(fmt.Errorf("elastic.NewScriptQuery params.FromTime: %w", err))
		}
		boolQ.Must(scriptQuery)
	}

	if params.ToTime != "" {
		t, err := time.Parse(timeFormat, params.ToTime)
		if err != nil {
			return errorResult(fmt.Errorf("params.ToTime cannot parse string to time with format 15:04:05(H:mm:ss)"))
		}
		scriptQuery, err := elastic.NewScriptQuery("ZonedDateTime dateTime = doc['end_time'].value.withZoneSameInstant(ZoneId.of(params['time_zone'])); return dateTime.getHour() * 60 + dateTime.getMinute() <= params['time_number']", map[string]interface{}{
			"time_zone":   params.TimeZone,
			"time_number": t.Hour()*60 + t.Minute(),
		})

		if err != nil {
			return errorResult(fmt.Errorf("elastic.NewScriptQuery params.ToTime: %w", err))
		}
		boolQ.Must(scriptQuery)
	}
	if len(params.LocationIDs) > 0 {
		boolQ.Must(elastic.NewTermsQuery("location_id", InterfaceSlice(params.LocationIDs)...))
	}

	if len(params.TeacherIDs) > 0 {
		boolQ.Must(elastic.NewTermsQuery("lesson_teachers.keyword", InterfaceSlice(params.TeacherIDs)...))
	}

	if len(params.StudentIDs) > 0 {
		boolQ.Must(elastic.NewTermsQuery("lesson_members.id", InterfaceSlice(params.StudentIDs)...))
	}

	if len(params.Grades) > 0 {
		boolQ.Must(elastic.NewTermsQuery("lesson_members.current_grade", InterfaceSlice(params.Grades)...))
	}

	if len(params.CourseIDs) > 0 {
		boolQ.Must(elastic.NewTermsQuery("lesson_members.course_id", InterfaceSlice(params.CourseIDs)...))
	}
	return boolQ, isAsc, nil
}

func (s *SearchRepo) search(ctx context.Context, params *domain.ListLessonArgs) (ret LessonSearchs, total uint32, offsetID string, err error) {
	boolQ, isAsc, err := s.buildLessonQuery(ctx, params)
	errorResult := func(err error) (LessonSearchs, uint32, string, error) { return nil, 0, "", err }
	if err != nil {
		return errorResult(fmt.Errorf("createLessonQuery: %w", err))
	}

	source := elastic.NewSearchSource().Query(boolQ)
	source.Size(int(params.Limit))
	source.Sort("start_time", isAsc)
	source.Sort("end_time", isAsc)
	source.Sort("lesson_id", isAsc)
	source.TrackTotalHits(true)
	var (
		lessonOffset       *LessonSearch
		isLessonNotExisted bool
	)
	if len(params.LessonID) > 0 {
		lessonOffset, err = s.GetLessonByID(ctx, params.LessonID)
		if err != nil {
			isLessonNotExisted = true
		} else {
			source.SearchAfter(lessonOffset.StartTime)
			source.SearchAfter(lessonOffset.EndTime)
			source.SearchAfter(lessonOffset.LessonID)
		}
	}

	var (
		res *esapi.Response
	)

	res, err = elastic.DoSearchFromSourceUsingJwtToken(ctx, s.SearchFactory, constants.LessonIndexName, source)
	if err != nil {
		return
	}

	total, err = elastic.ParseSearchWithTotalResponse(res.Body, func(hit *elastic.SearchHit) error {
		var t LessonSearch
		err = json.Unmarshal(hit.Source, &t)
		if err != nil {
			return fmt.Errorf("json.Unmarshal: %w", err)
		}

		ret = append(ret, &t)
		return nil
	})
	if err != nil {
		return errorResult(err)
	}

	if isLessonNotExisted {
		return nil, total, "", nil
	}
	// count previous
	if len(params.LessonID) > 0 {
		prevSource := elastic.NewSearchSource().Query(boolQ)
		prevSource.ClearRescorers().FetchSourceIncludeExclude([]string{"lesson_id"}, nil)
		prevSource.Size(int(params.Limit) + 1)
		prevSource.Sort("start_time", !isAsc)
		prevSource.Sort("end_time", !isAsc)
		prevSource.Sort("lesson_id", !isAsc)
		prevSource.SearchAfter(lessonOffset.StartTime)
		prevSource.SearchAfter(lessonOffset.EndTime)
		prevSource.SearchAfter(lessonOffset.LessonID)

		prevRes, err := elastic.DoSearchFromSourceUsingJwtToken(ctx, s.SearchFactory, constants.LessonIndexName, prevSource)
		if err != nil {
			return nil, total, "", nil
		}

		var prevLessonIDs []string

		err = elastic.ParseSearchResponse(prevRes.Body, func(hit *elastic.SearchHit) error {
			var t LessonSearch
			err = json.Unmarshal(hit.Source, &t)
			if err != nil {
				return fmt.Errorf("json.Unmarshal: %w", err)
			}

			prevLessonIDs = append(prevLessonIDs, t.LessonID)
			return nil
		})
		if err != nil {
			return errorResult(fmt.Errorf("get prev - error ParseSearchResponse: %w", err))
		}

		if len(prevLessonIDs) > int(params.Limit) {
			offsetID = prevLessonIDs[params.Limit-1]
		}
	}
	return
}

func (s *SearchRepo) Search(ctx context.Context, args *domain.ListLessonArgs) (ret []*domain.Lesson, total uint32, offsetID string, err error) {
	list, total, offsetID, err := s.search(ctx, args)
	if err != nil {
		return nil, 0, "", fmt.Errorf("Search: %w", err)
	}
	ret, err = list.ToLessonSearchEntities()
	if err != nil {
		return nil, 0, "", fmt.Errorf("Search-convertESToDomain: %w", err)
	}
	return
}

func InterfaceSlice(slice interface{}) []interface{} {
	s := reflect.ValueOf(slice)
	if s.Kind() != reflect.Slice {
		panic("InterfaceSlice() given a non-slice type")
	}

	// Keep the distinction between nil and empty slice input
	if s.IsNil() {
		return nil
	}

	ret := make([]interface{}, s.Len())
	for i := 0; i < s.Len(); i++ {
		ret[i] = s.Index(i).Interface()
	}

	return ret
}

func (s *SearchRepo) GetLessonByID(ctx context.Context, id string) (t *LessonSearch, err error) {
	source := elastic.NewSearchSource().Query(elastic.NewTermQuery("lesson_id", id))

	res, err := elastic.DoSearchFromSourceUsingJwtToken(ctx, s.SearchFactory, constants.LessonIndexName, source)
	if err != nil {
		return nil, fmt.Errorf("error DoSearchFromSource: %w", err)
	}

	total, err := elastic.ParseSearchWithTotalResponse(res.Body, func(hit *elastic.SearchHit) error {
		err = json.Unmarshal(hit.Source, &t)
		if err != nil {
			return fmt.Errorf("json.Unmarshal: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error ParseSearchResponse: %w", err)
	}
	if total == 0 {
		return nil, fmt.Errorf("error not found lesson ID %w", err)
	}
	return t, nil
}

func (s *SearchRepo) DeleteLessonIndex() error {
	isExistLessonIndex, err := s.SearchFactory.CheckIndexExists(constants.LessonIndexName)
	if err != nil {
		return fmt.Errorf("unable check exist lesson index %w", err)
	}
	if isExistLessonIndex {
		response, err := s.SearchFactory.DeleteIndex(constants.LessonIndexName)
		if err != nil {
			return err
		}
		defer response.Body.Close()
		if response.StatusCode != http.StatusOK {
			return fmt.Errorf("error " + response.String())
		}
		return err
	}
	return nil
}

func (s *SearchRepo) CreateLessonIndex() error {
	idxMap := strings.NewReader(constants.LessonIndexMapping)
	response, err := s.SearchFactory.CreateIndex(constants.LessonIndexName, idxMap)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("error " + response.String())
	}
	return err
}
