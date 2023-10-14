package commands

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/course/domain"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/course/infrastructure"
	locationDomain "github.com/manabie-com/backend/internal/mastermgmt/modules/location/domain"
	locationInfras "github.com/manabie-com/backend/internal/mastermgmt/modules/location/infrastructure"
	"github.com/manabie-com/backend/internal/mastermgmt/shared/dto"
	"github.com/manabie-com/backend/internal/mastermgmt/shared/validators"

	"github.com/jackc/pgx/v4"
)

var ChunkSize = 200

type CourseAccessPathCommandHandler struct {
	DB database.Ext

	CourseAccessPathRepo infrastructure.CourseAccessPathRepo
	LocationRepo         locationInfras.LocationRepo
	CourseRepo           infrastructure.CourseRepo
}

func (c *CourseAccessPathCommandHandler) UpsertCourseAccessPaths(ctx context.Context, payload UpsertCourseAccessPathsCommand) error {
	if len(payload.CourseAccessPaths) == 0 {
		return fmt.Errorf("%s", "course acess paths are empty")
	}

	capChunks := sliceutils.Chunk(payload.CourseAccessPaths, ChunkSize)
	return database.ExecInTx(ctx, c.DB, func(ctx context.Context, tx pgx.Tx) error {
		for i, chunk := range capChunks {
			err := c.CourseAccessPathRepo.Upsert(ctx, c.DB, chunk)
			if err != nil {
				return fmt.Errorf("CourseAccessPathRepo.Upsert chunk %d: %w", i, err)
			}
		}
		return nil
	})
}

// CheckCoursesAndLocations
// Check if course_ids and locations_id are existed in DB
func (c *CourseAccessPathCommandHandler) CheckCoursesAndLocations(ctx context.Context, courseAPs []*validators.CSVLineValue[domain.CourseAccessPath]) error {
	locationIDs := make([]string, len(courseAPs))
	courseIDs := make([]string, len(courseAPs))
	locationIDPos := make(map[string]int)
	courseIDPos := make(map[string]int)

	for i, courseAP := range courseAPs {
		locID := courseAP.Value.LocationID
		courseID := courseAP.Value.CourseID
		locationIDs[i] = locID
		courseIDs[i] = courseID

		locationIDPos[locID] = i
		courseIDPos[courseID] = i
	}

	locIDChunks := sliceutils.Chunk(locationIDs, ChunkSize)
	courseIDChunks := sliceutils.Chunk(courseIDs, ChunkSize)

	for _, locIDs := range locIDChunks {
		exLocs, err := c.LocationRepo.GetLocationsByLocationIDs(ctx, c.DB, database.TextArray(locIDs), false)
		if err != nil {
			return fmt.Errorf("CheckCoursesAndLocations.LocationRepo.GetLocationsByLocationIDs: %w", err)
		}

		exLocIDs := sliceutils.Map(exLocs, func(l *locationDomain.Location) string {
			return l.LocationID
		})
		if len(exLocIDs) < len(locIDs) {
			missing := FindMissing(locIDs, exLocIDs)
			for _, v := range missing {
				if cIndex, ok := locationIDPos[v]; ok {
					courseAP := courseAPs[cIndex]
					courseAP.Error = &dto.UpsertError{
						RowNumber: int32(cIndex + 2), // row of the csv
						Error:     fmt.Sprintf("location id %s is not exist", v),
					}
				}
			}
		}
	}
	for _, chunkCourseIDs := range courseIDChunks {
		exCourses, err := c.CourseRepo.GetByIDs(ctx, c.DB, chunkCourseIDs)
		if err != nil {
			return fmt.Errorf("CheckCoursesAndLocations.CourseRepo.GetByIDs: %w", err)
		}

		exCourseIDs := sliceutils.Map(exCourses, func(l *domain.Course) string {
			return l.CourseID
		})

		if len(exCourseIDs) < len(chunkCourseIDs) {
			missing := FindMissing(chunkCourseIDs, exCourseIDs)
			for _, v := range missing {
				if cIndex, ok := courseIDPos[v]; ok {
					courseAP := courseAPs[cIndex]
					if courseAP.Error == nil {
						courseAP.Error = &dto.UpsertError{
							RowNumber: int32(cIndex + 2),
							Error:     fmt.Sprintf("course id %s is not exist", v),
						}
					}
				}
			}
		}
	}

	return nil
}

// FindMissing
// find the missing element from slice b compared with slice a.
// "b" is the subset of "a"
func FindMissing(a, b []string) []string {
	smaller := make(map[string]bool, len(b))
	for _, v := range b {
		smaller[v] = true
	}

	var missing []string
	for _, v := range a {
		if _, ok := smaller[v]; !ok {
			missing = append(missing, v)
		}
	}
	return missing
}
