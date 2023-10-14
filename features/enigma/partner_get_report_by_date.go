package enigma

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/manabie-com/backend/internal/enigma/dto"
)

type PartnerLogResponse struct {
	Data map[string]map[string]int `json:"data"`
}

func (s *suite) aRequestGetPartnerDataReport(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	layout := "2006-01-02"
	time.Now().Format(layout)
	request := &dto.PartnerLogRequestByDate{
		Timestamp: int(time.Now().Unix()),
		Payload: struct {
			FromDate string `json:"from_date"`
			ToDate   string `json:"to_date"`
		}{
			FromDate: time.Now().Format(layout),
			ToDate:   time.Now().Format(layout),
		},
	}

	s.Request = request
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theRequestGetPartnerLogReportIsPerformed(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	url := fmt.Sprintf("%s/jprep/partner-log/report", s.EnigmaSrvURL)
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

func (s *suite) aResponsePartnerLogReportMatchDB(ctx context.Context, schoolID string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	layout := "2006-01-02"
	now := time.Now().Format(layout)
	ctx = s.setResourcePathToContext(ctx, schoolID)
	row := s.BobDB.QueryRow(ctx, `SELECT count(*) FROM public.partner_sync_data_log_split p WHERE DATE(p.created_at) = $1`, now)
	var count int
	if err := row.Scan(&count); err != nil {
		return ctx, err
	}
	res := &PartnerLogResponse{}
	json.Unmarshal([]byte(stepState.BodyBytes), &res)
	if results, ok := res.Data[now]; !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expect have data of date %s", now)
	} else {
		var total int
		for _, data := range results {
			total += data
		}
		if total != count {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expect %d, but got %d logs", count, total)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}
