package importmasterdata

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/timesheet/domain/entity"
	"github.com/manabie-com/backend/internal/timesheet/service/import_master_data/utils"
	pt "github.com/manabie-com/backend/pkg/manabuf/timesheet/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ImportTimesheetConfigService struct {
	DB database.Ext

	TimesheetConfigRepo interface {
		Create(ctx context.Context, db database.QueryExecer, config *entity.TimesheetConfig) error
		Update(ctx context.Context, db database.QueryExecer, config *entity.TimesheetConfig) error
	}
}

const (
	_offsetLineIndexInCSV = 2 // i = 0 <=> line number 2 in csv file
)

func _timesheetConfigCSVHeader() []string {
	return []string{
		"timesheet_config_id",
		"config_type",
		"config_value",
		"is_archived",
	}
}

func (s *ImportTimesheetConfigService) ImportTimesheetConfig(ctx context.Context, payload []byte) ([]*pt.ImportTimesheetConfigError, error) {

	configErrors := []*pt.ImportTimesheetConfigError{}

	lines, err := readCSV(payload)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	err = utils.ValidateCsvHeader(
		len(_timesheetConfigCSVHeader()),
		lines[0],
		_timesheetConfigCSVHeader(),
	)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("csv file invalid format - %s", err.Error()))
	}

	err = database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) (err error) {
		// first line is header
		for i, line := range lines[1:] {
			timesheetConfig, err := newTimesheetConfigFromCSV(line, _timesheetConfigCSVHeader())
			if err != nil {
				configErrors = append(configErrors, &pt.ImportTimesheetConfigError{
					RowNumber: int32(i) + _offsetLineIndexInCSV,
					Error:     fmt.Sprintf("unable to parse timesheet config item: %s", err),
				})
				continue
			}

			if timesheetConfig.ID.String == "" {
				err := timesheetConfig.ID.Set(idutil.ULIDNow())
				if err == nil {
					err = s.TimesheetConfigRepo.Create(ctx, tx, timesheetConfig)
				}

				if err != nil {
					configErrors = append(configErrors, &pt.ImportTimesheetConfigError{
						RowNumber: int32(i) + _offsetLineIndexInCSV,
						Error:     fmt.Sprintf("unable to create new timesheet config item: %s", err),
					})
				}

			} else {
				err := s.TimesheetConfigRepo.Update(ctx, tx, timesheetConfig)
				if err != nil {
					configErrors = append(configErrors, &pt.ImportTimesheetConfigError{
						RowNumber: int32(i) + _offsetLineIndexInCSV,
						Error:     fmt.Sprintf("unable to update timesheet config item: %s", err),
					})
				}
			}
		}
		if len(configErrors) > 0 {
			return fmt.Errorf(configErrors[0].Error)
		}
		return nil
	})
	if err != nil {
		log.Printf("Error when importing service timesheet config: %s", err.Error())
	}

	return configErrors, nil
}

func readCSV(payload []byte) ([][]string, error) {
	r := csv.NewReader(bytes.NewReader(payload))
	lines, err := r.ReadAll()
	if err != nil {
		return nil, err
	}

	if len(lines) < _offsetLineIndexInCSV {
		return nil, errors.New("no data in csv file")
	}

	return lines, nil
}

func newTimesheetConfigFromCSV(line []string, columnNames []string) (*entity.TimesheetConfig, error) {
	const (
		TimesheetConfigID = iota
		ConfigType
		ConfigValue
		IsArchived
	)

	mandatory := []int{ConfigType, ConfigValue, IsArchived}

	areMandatoryDataPresent, colPosition := checkMandatoryColumnAndGetIndex(line, mandatory)
	if !areMandatoryDataPresent {
		return nil, fmt.Errorf("missing mandatory data: %v", columnNames[colPosition])
	}

	timesheetConfig := &entity.TimesheetConfig{}

	var configTypeInt pgtype.Int2

	if err := multierr.Combine(
		utils.StringToFormatString("timesheet_config_id", line[TimesheetConfigID], true /*nullable*/, timesheetConfig.ID.Set),
		utils.StringToInt("config_type", line[ConfigType], false /*nullable*/, configTypeInt.Set),
		utils.StringToFormatString("config_value", line[ConfigValue], false /*nullable*/, timesheetConfig.ConfigValue.Set),
		utils.StringToBool("is_archived", line[IsArchived], false /*nullable*/, timesheetConfig.IsArchived.Set),
	); err != nil {
		return nil, err
	}

	if value, found := pt.TimesheetConfigType_name[int32(configTypeInt.Int)]; found {
		err := timesheetConfig.ConfigType.Set(value)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("invalid config_type")
	}

	return timesheetConfig, nil
}

func checkMandatoryColumnAndGetIndex(column []string, positions []int) (bool, int) {
	for _, position := range positions {
		if strings.TrimSpace(column[position]) == "" {
			return false, position
		}
	}
	return true, 0
}
