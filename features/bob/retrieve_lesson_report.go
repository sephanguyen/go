package bob

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/manabie-com/backend/features/helper"
	entities_bob "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/timeutil"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"

	"go.uber.org/multierr"
)

const (
	NIL_VALUE string = "nil"
)

func (s *suite) aListOfLessonReportsOfSchoolAreExistedInDB(ctx context.Context, strFromID, strToID string, schoolID int32) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if len(stepState.Random) == 0 {
		stepState.Random = helper.RandomString(6)
	}
	stepState.SchoolIDs = append(stepState.SchoolIDs, schoolID)

	fromID, err := strconv.Atoi((strings.Split(strFromID, "_"))[1])
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("cannot init lesson with outline init param")
	}
	toID, err := strconv.Atoi((strings.Split(strToID, "_"))[1])
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("cannot init lesson with outline init param")
	}
	for i := fromID; i < toID; i++ {
		lessonReport := &entities_bob.LessonReport{}
		database.AllNullEntity(lessonReport)
		lessonReportID := fmt.Sprintf("id_%d_%s", i, stepState.Random)
		err = multierr.Combine(
			lessonReport.LessonReportID.Set(lessonReportID),
			lessonReport.ReportSubmittingStatus.Set("submmited"),
			lessonReport.CreatedAt.Set(timeutil.Now()),
			lessonReport.UpdatedAt.Set(timeutil.Now()),
		)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		cmdTag, err := database.Insert(ctx, lessonReport, s.DB.Exec)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		if cmdTag.RowsAffected() != 1 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("cannot insert lesson")
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) cleanTestData(ctx context.Context) {
	// TODO: fix this test
	stepState := StepStateFromContext(ctx)
	sql := `UPDATE lesson_reports
	SET deleted_at = $1
	WHERE school_id = ANY($2)`
	s.DB.Exec(ctx, sql, time.Now(), database.Int4Array(stepState.SchoolIDs))
}

func (s *suite) userGetPartnerDomain(ctx context.Context, domainType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	domainTypes := map[string]bpb.DomainType{
		"Bo":      bpb.DomainType_DOMAIN_TYPE_BO,
		"Teacher": bpb.DomainType_DOMAIN_TYPE_TEACHER,
		"Learner": bpb.DomainType_DOMAIN_TYPE_LEARNER,
	}
	if _, ok := domainTypes[domainType]; !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("domain type %s is not valid", domainType)
	}

	req := &bpb.GetPartnerDomainRequest{
		Type: domainTypes[domainType],
	}

	stepState.Request = req
	stepState.Response, stepState.ResponseErr = bpb.NewLessonReportReaderServiceClient(s.Conn).RetrievePartnerDomain(contextWithToken(s, ctx), req)

	resp := stepState.Response.(*bpb.GetPartnerDomainResponse)
	val, err := url.Parse(resp.Domain)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("cannot parse domain %s", resp.Domain)
	}
	if val.Scheme != "https" {
		return StepStateToContext(ctx, stepState), fmt.Errorf("domain %s need ssl", resp.Domain)
	}
	return StepStateToContext(ctx, stepState), nil
}
