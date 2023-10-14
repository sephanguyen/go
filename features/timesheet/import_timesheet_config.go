package timesheet

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/timesheet/domain/entity"
	pt "github.com/manabie-com/backend/pkg/manabuf/timesheet/v1"

	"github.com/pkg/errors"
)

func (s *Suite) insertSomeTimesheetConfigs(ctx context.Context) error {
	for i := 0; i < 5; i++ {
		config_id := idutil.ULIDNow()
		configType := pt.TimesheetConfigType_OTHER_WORKING_HOURS.String()
		configValue := database.Text("value " + idutil.ULIDNow())
		isArchived := database.Bool(rand.Int()%2 == 0)
		stmt := `INSERT INTO timesheet_config
		(timesheet_config_id, config_type, config_value, is_archived, created_at, updated_at)
		VALUES ($1, $2, $3, $4, now(), now())`
		_, err := s.TimesheetDBTrace.Exec(ctx, stmt, config_id, configType, configValue, isArchived)
		if err != nil {
			return fmt.Errorf("cannot insert timesheet config, err: %s", err)
		}
	}
	return nil
}

func (s *Suite) selectAllTimesheetConfigs(ctx context.Context) ([]*entity.TimesheetConfig, error) {
	allEntities := []*entity.TimesheetConfig{}
	stmt :=
		`
		SELECT
			timesheet_config_id,
			config_type,
			config_value,
			is_archived
		FROM
			timesheet_config
		ORDER BY
			created_at ASC
        `
	rows, err := s.TimesheetDBTrace.Query(
		ctx,
		stmt,
	)
	if err != nil {
		return nil, errors.Wrap(err, "query timesheet_config")
	}

	defer rows.Close()
	for rows.Next() {
		e := &entity.TimesheetConfig{}
		err := rows.Scan(
			&e.ID,
			&e.ConfigType,
			&e.ConfigValue,
			&e.IsArchived,
		)
		if err != nil {
			return nil, errors.WithMessage(err, "rows.Scan timesheet config")
		}
		allEntities = append(allEntities, e)
	}
	return allEntities, nil
}

func (s *Suite) importingTimesheetConfig(ctx context.Context, userGroup string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.SignedAsAccount(ctx, userGroup)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.RequestSentAt = time.Now()
	stepState.Response, stepState.ResponseErr = pt.NewImportMasterDataServiceClient(s.TimesheetConn).
		ImportTimesheetConfig(contextWithToken(ctx), stepState.Request.(*pt.ImportTimesheetConfigRequest))

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) aTimesheetConfigValidRequestPayload(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	err := s.insertSomeTimesheetConfigs(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	validRow1 := fmt.Sprintf(",%d,value %s,0", pt.TimesheetConfigType_OTHER_WORKING_HOURS, idutil.ULIDNow())
	validRow2 := fmt.Sprintf(",%d,value %s,1", pt.TimesheetConfigType_OTHER_WORKING_HOURS, idutil.ULIDNow())

	stepState.Request = &pt.ImportTimesheetConfigRequest{
		Payload: []byte(fmt.Sprintf(`timesheet_config_id,config_type,config_value,is_archived
			%s
			%s`, validRow1, validRow2)),
	}
	stepState.ValidCsvRows = []string{validRow1, validRow2}

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) aTimesheetConfigValidRequestPayloadWithIncorrectData(ctx context.Context, rowCondition string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	err := s.insertSomeTimesheetConfigs(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	existingTimesheetConfigs, err := s.selectAllTimesheetConfigs(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	validRow1 := fmt.Sprintf(",%d,value %s,1", pt.TimesheetConfigType_OTHER_WORKING_HOURS, idutil.ULIDNow())
	validRow2 := fmt.Sprintf(",%d,value %s,0", pt.TimesheetConfigType_OTHER_WORKING_HOURS, idutil.ULIDNow())
	validRow3 := fmt.Sprintf("%s,%d,value %s,0", existingTimesheetConfigs[0].ID.String, pt.TimesheetConfigType_OTHER_WORKING_HOURS, idutil.ULIDNow())
	invalidEmptyRow1 := fmt.Sprintf(",%d,value %s,", pt.TimesheetConfigType_OTHER_WORKING_HOURS, idutil.ULIDNow())
	invalidEmptyRow2 := fmt.Sprintf("%s,%d,value %s,", existingTimesheetConfigs[1].ID.String, pt.TimesheetConfigType_OTHER_WORKING_HOURS, idutil.ULIDNow())
	invalidValueRow1 := fmt.Sprintf(",%d,value %s,Archived", pt.TimesheetConfigType_OTHER_WORKING_HOURS, idutil.ULIDNow())
	invalidValueRow2 := fmt.Sprintf("%s,%d,value %s,Archived", existingTimesheetConfigs[2].ID.String, pt.TimesheetConfigType_OTHER_WORKING_HOURS, idutil.ULIDNow())

	stepState.ValidCsvRows = []string{}
	stepState.InvalidCsvRows = []string{}
	switch rowCondition {
	case "empty value row":
		stepState.Request = &pt.ImportTimesheetConfigRequest{
			Payload: []byte(fmt.Sprintf(`timesheet_config_id,config_type,config_value,is_archived
			%s
			%s`, invalidEmptyRow1, invalidEmptyRow2)),
		}
		stepState.InvalidCsvRows = []string{invalidEmptyRow1, invalidEmptyRow2}
	case "invalid value row":
		stepState.Request = &pt.ImportTimesheetConfigRequest{
			Payload: []byte(fmt.Sprintf(`timesheet_config_id,config_type,config_value,is_archived
			%s
			%s`, invalidValueRow1, invalidValueRow2)),
		}
		stepState.InvalidCsvRows = []string{invalidValueRow1, invalidValueRow2}
	case "valid and invalid rows":
		stepState.Request = &pt.ImportTimesheetConfigRequest{
			Payload: []byte(fmt.Sprintf(`timesheet_config_id,config_type,config_value,is_archived
			%s
			%s
			%s
			%s
			%s
			%s
			%s`, validRow1, validRow2, validRow3, invalidEmptyRow1, invalidEmptyRow2, invalidValueRow1, invalidValueRow2)),
		}
		stepState.ValidCsvRows = []string{validRow1, validRow2, validRow3}
		stepState.InvalidCsvRows = []string{invalidEmptyRow1, invalidEmptyRow2, invalidValueRow1, invalidValueRow2}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) theValidTimesheetConfigLinesAreImportedSuccessfully(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.ResponseErr != nil {
		return ctx, stepState.ResponseErr
	}

	allTimesheetConfigs, err := s.selectAllTimesheetConfigs(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	// we should use map for allAccoutingCategories but it leads to some more code and not many items in
	// stepState.ValidCsvRows and allAccoutingCategories, so we can do like below to make it simple
	for _, row := range stepState.ValidCsvRows {
		rowSplit := strings.Split(row, ",")
		configTypeInt, _ := strconv.Atoi(rowSplit[1])
		configType := pt.TimesheetConfigType(configTypeInt).String()
		configValue := rowSplit[2]
		isArchived, err := strconv.ParseBool(rowSplit[3])
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		found := false
		for _, e := range allTimesheetConfigs {
			if e.ConfigType.String == configType && e.ConfigValue.String == configValue && e.IsArchived.Bool == isArchived {
				found = true
				break
			}
		}
		if !found {
			return StepStateToContext(ctx, stepState), fmt.Errorf("failed to import valid csv row")
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) theImportTimesheetConfigTransactionIsRolledBack(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	allTimesheetConfigs, err := s.selectAllTimesheetConfigs(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if len(stepState.ValidCsvRows) == 0 {
		return StepStateToContext(ctx, stepState), nil
	}

	for _, row := range stepState.ValidCsvRows {
		found := false

		rowSplit := strings.Split(row, ",")
		configTypeInt, _ := strconv.Atoi(rowSplit[1])
		configType := pt.TimesheetConfigType(configTypeInt).String()
		configValue := rowSplit[2]
		isArchived, err := strconv.ParseBool(rowSplit[3])
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		for _, e := range allTimesheetConfigs {
			if e.ConfigType.String == configType && e.ConfigValue.String == configValue && e.IsArchived.Bool == isArchived {
				found = true
				break
			}
		}

		if found {
			return StepStateToContext(ctx, stepState), fmt.Errorf("failed to rollback valid csv row")
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) theInvalidTimesheetConfigLinesAreReturnedWithError(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := stepState.Request.(*pt.ImportTimesheetConfigRequest)
	reqSplit := strings.Split(string(req.Payload), "\n")
	resp := stepState.Response.(*pt.ImportTimesheetConfigResponse)
	for _, row := range stepState.InvalidCsvRows {
		found := false
		for _, e := range resp.Errors {
			if strings.TrimSpace(reqSplit[e.RowNumber-1]) == row {
				found = true
				break
			}
		}
		if !found {
			return StepStateToContext(ctx, stepState), fmt.Errorf("invalid line is not returned in response")
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) aTimesheetConfigInvalidRequestPayload(ctx context.Context, invalidFormat string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	switch invalidFormat {
	case "no data":
		stepState.Request = &pt.ImportTimesheetConfigRequest{}
	case "header only":
		stepState.Request = &pt.ImportTimesheetConfigRequest{
			Payload: []byte(`timesheet_config_id,config_type,config_value,is_archived`),
		}
	case "number of column is not equal 4":
		stepState.Request = &pt.ImportTimesheetConfigRequest{
			Payload: []byte(`timesheet_config_id,config_type,config_value
			1,0,value 1`),
		}
	case "mismatched number of fields in header and content":
		stepState.Request = &pt.ImportTimesheetConfigRequest{
			Payload: []byte(`timesheet_config_id,config_type,config_value,is_archived
			1,0,value 1
			2,0,value 2
			3,0,value 3`),
		}
	case "wrong timesheet_config_id column name in header":
		stepState.Request = &pt.ImportTimesheetConfigRequest{
			Payload: []byte(`Number,config_type,config_value,is_archived
			1,0,value 1,0
			2,0,value 2,0
			3,0,value 3,0`),
		}
	case "wrong config_type column name in header":
		stepState.Request = &pt.ImportTimesheetConfigRequest{
			Payload: []byte(`timesheet_config_id,Naming,config_value,is_archived
			1,0,value 1,0
			2,0,value 2,0
			3,0,value 3,0`),
		}
	case "wrong config_value column name in header":
		stepState.Request = &pt.ImportTimesheetConfigRequest{
			Payload: []byte(`timesheet_config_id,config_type,Description,is_archived
			1,0,value 1,0
			2,0,value 2,0
			3,0,value 3,0`),
		}
	case "wrong is_archived column name in header":
		stepState.Request = &pt.ImportTimesheetConfigRequest{
			Payload: []byte(`timesheet_config_id,config_type,config_value,IsArchived
			1,0,value 1,0
			2,0,value 2,0
			3,0,value 3,0`),
		}
	}
	return StepStateToContext(ctx, stepState), nil
}
