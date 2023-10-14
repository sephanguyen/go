package enigma

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/manabie-com/backend/internal/enigma/dto"
	"github.com/manabie-com/backend/internal/enigma/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"

	"go.uber.org/multierr"
)

func (s *suite) someDataLogSplitOfKindWithStatusAndTry(ctx context.Context, kind, status string, retryTimes int, date string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	parentID, childID := s.newID(), s.newID()
	s.CurrentUserID = idutil.ULIDNow()
	layout := "2006-01-02"
	createdDate, err := time.Parse(layout, date)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	// create parent
	partnerSyncDataLog := &entities.PartnerSyncDataLog{}
	database.AllNullEntity(partnerSyncDataLog)
	ctx = s.setResourcePathToContext(ctx, "-2147483647")
	err = multierr.Combine(
		partnerSyncDataLog.PartnerSyncDataLogID.Set(parentID),
		partnerSyncDataLog.CreatedAt.Set(createdDate),
		partnerSyncDataLog.Signature.Set(""),
		partnerSyncDataLog.Payload.Set(database.JSONB(fmt.Sprintf(`[{"last_name": "Last name %s", "given_name": "Given name %s", "student_id": "1", "action_kind": 1}]`, s.CurrentUserID, s.CurrentUserID))),
		partnerSyncDataLog.UpdatedAt.Set(createdDate),
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	_, err = database.Insert(ctx, partnerSyncDataLog, s.BobDBTrace.Exec)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("can't create partnerSyncDataLog %+v", err)
	}
	partnerSyncDataLogSplit := &entities.PartnerSyncDataLogSplit{}
	database.AllNullEntity(partnerSyncDataLogSplit)
	var payload string

	switch kind {
	case string(entities.KindStudent):
		payload = fmt.Sprintf(`[{"last_name": "Last name %s", "given_name": "Given name %s", "student_id": %s, "action_kind": 1}]`, s.CurrentUserID, s.CurrentUserID, stepState.Random)
	case string(entities.KindStaff):
		payload = fmt.Sprintf(`[{"name": "Name %s", "staff_id": %s, "action_kind": 1}]`, s.CurrentUserID, stepState.Random)
	case string(entities.KindLesson):
		payload = fmt.Sprintf(`[{"end_date": {"seconds": 1650861800}, "course_id": "JPREP_COURSE_%s", "lesson_id": "JPREP_LESSON_%s", "class_name": "class name %s", "start_date": {"seconds": 1650861800}, "action_kind": 1, "lesson_type": 1, "lesson_group": "1"}]`, stepState.Random, stepState.Random, s.CurrentUserID)
	case string(entities.KindCourse):
		payload = fmt.Sprintf(`[{"status": 4, "course_id": "JPREP_COURSE_%s", "action_kind": 1, "course_name": "course-name %s"}]`, stepState.Random, s.CurrentUserID)
	case string(entities.KindClass):
		payload = fmt.Sprintf(`[{"class_id": 1, "end_date": {"seconds": 1650898799}, "course_id": "JPREP_COURSE_%s", "class_name": "class name %s", "start_date": {"seconds": 1650812400}, "action_kind": 1, "academic_year_id": "JPREP_ACADEMIC_YEAR_%s"}]`, stepState.Random, s.CurrentUserID, stepState.Random)
	case string(entities.KindAcademicYear):
		payload = fmt.Sprintf(`[{"name": "name %s", "action_kind": 1, "end_year_date": {"seconds": 1}, "start_year_date": {"seconds": 1}, "academic_year_id": "JPREP_ACADEMIC_YEAR_%s"}]`, s.CurrentUserID, stepState.Random)
	case string(entities.KindStudentLessons):
		payload = fmt.Sprintf(`[{"lesson_ids": ["JPREP_LESSON_%s"], "student_id": "501", "action_kind": 1}]`, stepState.Random)
	default:
		return StepStateToContext(ctx, stepState), fmt.Errorf("Kind %s invalid", kind)
	}
	err = multierr.Combine(
		partnerSyncDataLogSplit.PartnerSyncDataLogID.Set(parentID),
		partnerSyncDataLogSplit.PartnerSyncDataLogSplitID.Set(childID),
		partnerSyncDataLogSplit.CreatedAt.Set(createdDate),
		partnerSyncDataLogSplit.UpdatedAt.Set(createdDate),
		partnerSyncDataLogSplit.Status.Set(status),
		partnerSyncDataLogSplit.Payload.Set(database.JSONB(payload)),
		partnerSyncDataLogSplit.Kind.Set(kind),
		partnerSyncDataLogSplit.RetryTimes.Set(retryTimes),
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	_, err = database.Insert(ctx, partnerSyncDataLogSplit, s.BobDBTrace.Exec)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("can't create partnerSyncDataLog %+v", err)
	}
	stepState.PartnerSyncDataLogSplitID = childID
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) requestWithRecoverDataSync(ctx context.Context, date string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	request := &dto.PartnerLogRequestByDate{
		Timestamp: int(time.Now().Unix()),
		Payload: struct {
			FromDate string `json:"from_date"`
			ToDate   string `json:"to_date"`
		}{
			FromDate: date,
			ToDate:   date,
		},
	}

	s.Request = request
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theRequestRecoverDataSyncIsPerformed(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	url := fmt.Sprintf("%s/jprep/partner-log/recover", s.EnigmaSrvURL)
	bodyBytes, err := s.makeHTTPRequest(http.MethodPost, url)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if bodyBytes == nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("body is nil")
	}
	stepState.BodyBytes = string(bodyBytes)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aLogSplitMatchWithStatusAndRetryTimes(ctx context.Context, schoolID string, expectRetryTimes int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx = s.setResourcePathToContext(ctx, schoolID)
	row := s.BobDB.QueryRow(ctx, `SELECT retry_times FROM public.partner_sync_data_log_split p WHERE p.partner_sync_data_log_split_id = $1 `, stepState.PartnerSyncDataLogSplitID)
	var retryTimes int
	if err := row.Scan(&retryTimes); err != nil {
		return ctx, err
	}
	if retryTimes != expectRetryTimes {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expect retry_times %d, but got %d", expectRetryTimes, retryTimes)
	}

	return StepStateToContext(ctx, stepState), nil
}
