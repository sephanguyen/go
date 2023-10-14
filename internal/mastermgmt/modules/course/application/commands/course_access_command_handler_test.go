package commands

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/course/domain"
	locationDomain "github.com/manabie-com/backend/internal/mastermgmt/modules/location/domain"
	"github.com/manabie-com/backend/internal/mastermgmt/shared/dto"
	"github.com/manabie-com/backend/internal/mastermgmt/shared/validators"
	mockDatabase "github.com/manabie-com/backend/mock/golibs/database"
	mockCourseRepo "github.com/manabie-com/backend/mock/mastermgmt/modules/course/infrastructure/repo"
	mockLocationRepo "github.com/manabie-com/backend/mock/mastermgmt/modules/location/infrastructure/repo"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCourseAccessPathCommandHandler_CheckCoursesAndLocations(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	db := &mockDatabase.Ext{}
	courseAccessPathRepo := new(mockCourseRepo.MockCourseAccessPathRepo)
	courseRepo := new(mockCourseRepo.MockCourseRepo)
	locationRepo := new(mockLocationRepo.MockLocationRepo)
	caps := fakeManyAccessPaths(3)

	t.Run("Course and location are all existed", func(t *testing.T) {
		// arrange
		csvCAP := sliceutils.Map(caps, func(path domain.CourseAccessPath) *validators.CSVLineValue[domain.CourseAccessPath] {
			return &validators.CSVLineValue[domain.CourseAccessPath]{
				Value: &path,
			}
		})

		locationRepo.On("GetLocationsByLocationIDs",
			mock.Anything,
			db,
			database.TextArray([]string{"location_id_1", "location_id_2", "location_id_3"}),
			false).
			Return([]*locationDomain.Location{
				{
					LocationID: "location_id_1",
				},
				{
					LocationID: "location_id_3",
				},
				{
					LocationID: "location_id_2",
				},
			}, nil).
			Once()
		courseRepo.On("GetByIDs",
			mock.Anything,
			db,
			[]string{"course_id_1", "course_id_2", "course_id_3"}).
			Return([]*domain.Course{
				{
					CourseID: "course_id_1",
				},
				{
					CourseID: "course_id_3",
				},
				{
					CourseID: "course_id_2",
				},
			}, nil).
			Once()

		handler := CourseAccessPathCommandHandler{
			DB:                   db,
			CourseAccessPathRepo: courseAccessPathRepo,
			LocationRepo:         locationRepo,
			CourseRepo:           courseRepo,
		}

		// act
		_ = handler.CheckCoursesAndLocations(ctx, csvCAP)

		// assert
		for _, v := range csvCAP {
			assert.Nil(t, v.Error)
		}
		mock.AssertExpectationsForObjects(t, locationRepo)
		mock.AssertExpectationsForObjects(t, courseRepo)
	})

	t.Run("Some courses are not existed", func(t *testing.T) {
		// arrange
		csvCAP := sliceutils.Map(caps, func(path domain.CourseAccessPath) *validators.CSVLineValue[domain.CourseAccessPath] {
			return &validators.CSVLineValue[domain.CourseAccessPath]{
				Value: &path,
			}
		})
		locationRepo.On("GetLocationsByLocationIDs",
			mock.Anything,
			mock.Anything,
			database.TextArray([]string{"location_id_1", "location_id_2", "location_id_3"}),
			false).
			Return([]*locationDomain.Location{
				{
					LocationID: "location_id_1",
				},
				{
					LocationID: "location_id_3",
				},
				{
					LocationID: "location_id_2",
				},
			}, nil).
			Once()
		courseRepo.On("GetByIDs",
			mock.Anything,
			mock.Anything,
			[]string{"course_id_1", "course_id_2", "course_id_3"}).
			Return([]*domain.Course{
				{
					CourseID: "course_id_1",
				},
				{
					CourseID: "course_id_3",
				},
			}, nil).
			Once()

		handler := CourseAccessPathCommandHandler{
			DB:                   db,
			CourseAccessPathRepo: courseAccessPathRepo,
			LocationRepo:         locationRepo,
			CourseRepo:           courseRepo,
		}

		// act
		_ = handler.CheckCoursesAndLocations(ctx, csvCAP)

		// assert
		for _, v := range csvCAP {
			if v.Value.CourseID == "course_id_2" {
				assert.Equal(t, &dto.UpsertError{
					RowNumber: 3,
					Error:     fmt.Sprintf("course id %s is not exist", "course_id_2"),
				}, v.Error)
			}
		}

		mock.AssertExpectationsForObjects(t, locationRepo)
		mock.AssertExpectationsForObjects(t, courseAccessPathRepo)
		mock.AssertExpectationsForObjects(t, courseRepo)
	})

	t.Run("Some locations are not existed", func(t *testing.T) {
		// arrange
		csvCAP := sliceutils.Map(caps, func(path domain.CourseAccessPath) *validators.CSVLineValue[domain.CourseAccessPath] {
			return &validators.CSVLineValue[domain.CourseAccessPath]{
				Value: &path,
			}
		})
		locationRepo.On("GetLocationsByLocationIDs",
			mock.Anything,
			mock.Anything,
			database.TextArray([]string{"location_id_1", "location_id_2", "location_id_3"}),
			false).
			Return([]*locationDomain.Location{
				{
					LocationID: "location_id_3",
				},
				{
					LocationID: "location_id_1",
				},
			}, nil).
			Once()
		courseRepo.On("GetByIDs",
			mock.Anything,
			mock.Anything,
			[]string{"course_id_1", "course_id_2", "course_id_3"}).
			Return([]*domain.Course{
				{
					CourseID: "course_id_1",
				},
				{
					CourseID: "course_id_3",
				},
				{
					CourseID: "course_id_2",
				},
			}, nil).
			Once()

		handler := CourseAccessPathCommandHandler{
			DB:                   db,
			CourseAccessPathRepo: courseAccessPathRepo,
			LocationRepo:         locationRepo,
			CourseRepo:           courseRepo,
		}

		// act
		_ = handler.CheckCoursesAndLocations(ctx, csvCAP)

		// assert
		for _, v := range csvCAP {
			if v.Value.CourseID == "course_id_2" {
				assert.Equal(t, &dto.UpsertError{
					RowNumber: 3,
					Error:     fmt.Sprintf("location id %s is not exist", "location_id_2"),
				}, v.Error)
			}
		}

		mock.AssertExpectationsForObjects(t, locationRepo)
		mock.AssertExpectationsForObjects(t, courseRepo)
	})

	t.Run("Course and location are more than 200 items", func(t *testing.T) {
		// arrange
		caps := fakeManyAccessPaths(250)
		capChunks := sliceutils.Chunk(caps, ChunkSize)
		csvCAP := sliceutils.Map(caps, func(path domain.CourseAccessPath) *validators.CSVLineValue[domain.CourseAccessPath] {
			return &validators.CSVLineValue[domain.CourseAccessPath]{
				Value: &path,
			}
		})

		for _, chunk := range capChunks {
			locIDs := sliceutils.Map(chunk, func(c domain.CourseAccessPath) string {
				return c.LocationID
			})
			courseIDs := sliceutils.Map(chunk, func(c domain.CourseAccessPath) string {
				return c.CourseID
			})
			courses := sliceutils.Map(chunk, func(c domain.CourseAccessPath) *domain.Course {
				return &domain.Course{
					CourseID: c.CourseID,
				}
			})
			locations := sliceutils.Map(chunk, func(c domain.CourseAccessPath) *locationDomain.Location {
				return &locationDomain.Location{
					LocationID: c.LocationID,
				}
			})
			locationRepo.On("GetLocationsByLocationIDs",
				mock.Anything,
				mock.Anything,
				database.TextArray(locIDs),
				false).
				Return(locations, nil).
				Once()
			courseRepo.On("GetByIDs",
				mock.Anything,
				mock.Anything,
				courseIDs).
				Return(courses, nil).
				Once()
		}

		handler := CourseAccessPathCommandHandler{
			DB:                   db,
			CourseAccessPathRepo: courseAccessPathRepo,
			LocationRepo:         locationRepo,
			CourseRepo:           courseRepo,
		}

		// act
		_ = handler.CheckCoursesAndLocations(ctx, csvCAP)

		// assert
		for _, v := range csvCAP {
			assert.Nil(t, v.Error)
		}

		mock.AssertExpectationsForObjects(t, locationRepo)
		mock.AssertExpectationsForObjects(t, courseRepo)
	})
}

func TestFindMissing(t *testing.T) {
	t.Run("return missing elements", func(t *testing.T) {
		t.Parallel()
		//arrange
		a := []string{"apple", "banana", "kiendt"}
		b := []string{"apple", "kiendt"}
		m := []string{"banana"}

		// act
		actual := FindMissing(a, b)

		// assert
		assert.Equal(t, m, actual)
	})

	t.Run("not exist any missing elements", func(t *testing.T) {
		t.Parallel()
		//arrange
		a := []string{"apple", "banana", "kiendt"}
		b := []string{"apple", "kiendt", "banana"}
		var m []string

		// act
		actual := FindMissing(a, b)

		// assert
		assert.Equal(t, m, actual)
	})
}

func fakeManyAccessPaths(limit int) []domain.CourseAccessPath {
	caps := make([]domain.CourseAccessPath, limit)
	for i := 0; i < limit; i++ {
		counter := i + 1
		caps[i] = domain.CourseAccessPath{
			ID:         fmt.Sprintf("id_%d", counter),
			LocationID: fmt.Sprintf("location_id_%d", counter),
			CourseID:   fmt.Sprintf("course_id_%d", counter),
		}
	}
	return caps
}
