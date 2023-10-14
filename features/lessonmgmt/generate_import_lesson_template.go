package lessonmgmt

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/features/helper"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"
)

func (s *Suite) downloadLessonTemplate(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := &lpb.GenerateLessonCSVTemplateRequest{}
	stepState.Response, stepState.ResponseErr = lpb.NewLessonExecutorServiceClient(s.LessonMgmtConn).
		GenerateLessonCSVTemplate(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) returnsLessonCSVTemplate(ctx context.Context, expectColsStr string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return ctx, fmt.Errorf("can not generate lesson csv template: %s", stepState.ResponseErr.Error())
	}
	resp := stepState.Response.(*lpb.GenerateLessonCSVTemplateResponse)

	headers := strings.Split(expectColsStr, ",")
	lessonStr := [][]string{headers}
	lineStr := []string{}

	for _, h := range headers {
		switch h {
		case "partner_internal_id":
			lineStr = append(lineStr, "sample_center_id")
		case "start_date_time":
			lineStr = append(lineStr, time.Now().Format("2006-01-02 15:04:05"))
		case "end_date_time":
			lineStr = append(lineStr, time.Now().Add(2*time.Hour).Format("2006-01-02 15:04:05"))
		case "teaching_method":
			lineStr = append(lineStr, domain.MapKeyLessonTeachingMethod[domain.LessonTeachingMethodIndividual])
		case "teaching_medium":
			lineStr = append(lineStr, domain.MapKeyLessonTeachingMedium[domain.LessonTeachingMediumOffline])
		case "teacher_ids":
			lineStr = append(lineStr, "teacherID1_teacherID2_teacherID3")
		case "student_course_ids":
			lineStr = append(lineStr, "studentID1/courseID1_studentID2/courseID2")
		}
	}
	lessonStr = append(lessonStr, lineStr)

	sb := strings.Builder{}
	for _, line := range lessonStr {
		row := sliceutils.Map(line, getEscapedStr)
		sb.WriteString(fmt.Sprintf("%s\n", strings.Join(row, ",")))
	}
	expectedData := sb.String()
	if expectedData != string(resp.GetData()) {
		return ctx, fmt.Errorf("template csv is not valid:\ngot:\n%s \nexpected: \n%s", resp.Data, expectedData)
	}
	return StepStateToContext(ctx, stepState), nil
}

// Get escaped tricky character like double quote
func getEscapedStr(s string) string {
	mustQuote := strings.ContainsAny(s, `"`)
	if mustQuote {
		s = strings.ReplaceAll(s, `"`, `""`)
	}

	return fmt.Sprintf(`"%s"`, s)
}
