package commands

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/scanner"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/infrastructure"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/support"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"
)

type CourseTeachingTimeCommandHandler struct {
	WrapperConnection *support.WrapperDBConnection
	CourseRepo        infrastructure.CourseRepo
}

func (c *CourseTeachingTimeCommandHandler) ImportCourseTeachingTime(ctx context.Context, req *lpb.ImportCourseTeachingTimeRequest) (domain.Courses, []*lpb.ImportError, error) {
	courses, CSVErrors := c.buildImportCourseTeachingTimeArgs(ctx, req.Payload)
	if len(CSVErrors) > 0 {
		return courses, ConvertErrToImportCSVErr(CSVErrors), nil
	}
	conn, err := c.WrapperConnection.GetDB(golibs.ResourcePathFromCtx(ctx))
	if err != nil {
		return nil, nil, err
	}
	err = c.CourseRepo.RegisterCourseTeachingTime(ctx, conn, courses)
	if err != nil {
		return courses, nil, err
	}
	return courses, nil, nil
}

func (c *CourseTeachingTimeCommandHandler) buildImportCourseTeachingTimeArgs(ctx context.Context, data []byte) (domain.Courses, map[int]error) {
	sc1 := scanner.NewCSVScanner(bytes.NewReader(data))
	columnsIndex := map[string]int{
		"course_id":        0,
		"course_name":      1,
		"preparation_time": 2,
		"break_time":       3,
		"action":           4,
	}

	CSVErrors := ValidateImportFileHeader(sc1, columnsIndex)
	if len(CSVErrors) > 0 {
		return nil, CSVErrors
	}

	courseIDs := make([]string, 0, len(sc1.GetRow()))
	courses := make([]*domain.Course, len(courseIDs))
	for sc1.Scan() {
		curRow := sc1.GetCurRow()
		courseID, preparationTime, breakTime, action := sc1.Text("course_id"), sc1.Text("preparation_time"), sc1.Text("break_time"), sc1.Text("action")

		if courseID == "" {
			CSVErrors[curRow] = fmt.Errorf("course_id should not be null value")
			continue
		}

		if strings.ToLower(action) != "upsert" && strings.ToLower(action) != "delete" {
			CSVErrors[curRow] = fmt.Errorf("invalid action")
		}
		course := domain.Course{
			CourseID:        courseID,
			PreparationTime: 0,
			BreakTime:       0,
		}

		if preparationTime != "" {
			ptime, err := strconv.ParseInt(preparationTime, 10, 32)
			if err != nil || ptime < 0 {
				CSVErrors[curRow] = fmt.Errorf("preparation time must be numeric and should be greater than or equal to 0")
			}
			course.PreparationTime = int32(ptime)
		}

		if breakTime != "" {
			btime, err := strconv.ParseInt(breakTime, 10, 32)
			if err != nil || btime < 0 {
				CSVErrors[curRow] = fmt.Errorf("break time must be numeric and should be greater than or equal to 0")
			}
			course.BreakTime = int32(btime)
		}
		if strings.ToLower(action) == "delete" {
			course.DeletedAt = time.Now()
		}

		courses = append(courses, &course)
		courseIDs = append(courseIDs, courseID)
	}
	if len(courseIDs) > 0 {
		conn, err := c.WrapperConnection.GetDB(golibs.ResourcePathFromCtx(ctx))
		if err != nil {
			CSVErrors[0] = err
			return courses, CSVErrors
		}
		err = c.CourseRepo.CheckCourseIDs(ctx, conn, courseIDs)
		if err != nil {
			CSVErrors[0] = err
		}
	}
	return courses, CSVErrors
}
