package constant

import (
	"time"

	"github.com/manabie-com/backend/internal/golibs/timeutil"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"
)

const (
	KTimesheetRemarkLimit            = 500
	KOtherWorkingHoursRemarksLimit   = 100
	KListOtherWorkingHoursLimit      = 5
	SeparatorComma                   = ","
	FeatureToggleAutoCreate          = "BACKEND_Timesheet_TimesheetManagement_AutoCreate"
	FeatureToggleActionLog           = "BACKEND_Timesheet_TimesheetManagement_ActionLog"
	KTransportExpensesRemarksLimit   = 100
	KTransportExpensesFromToLimit    = 100
	KListTransportExpensesLimit      = 10
	KListStaffTransportExpensesLimit = 10
	KPartnerAutoCreateDefaultValue   = false
)

var (
	KTimesheetMinDate = time.Date(2022, 1, 1, 0, 0, 0, 0, timeutil.Timezone(pb.COUNTRY_JP)) // constant value, do not change
)

type ConfigSettingStatus int8

const (
	On ConfigSettingStatus = iota
	Off
)

var configStatusToString = map[ConfigSettingStatus]string{
	On:  "on",
	Off: "off",
}

func (c ConfigSettingStatus) String() string {
	s, ok := configStatusToString[c]
	if !ok {
		return ""
	}
	return s
}

type ConfigSettingType int8

const (
	ConfigSettingTypeString ConfigSettingType = iota
)

var configTypeToString = map[ConfigSettingType]string{
	ConfigSettingTypeString: "string",
}

func (c ConfigSettingType) String() string {
	s, ok := configTypeToString[c]
	if !ok {
		return ""
	}
	return s
}
