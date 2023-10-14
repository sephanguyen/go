package mock

import (
	"github.com/spf13/cobra"
)

// RootCmd for mock command
var RootCmd = &cobra.Command{
	Use:   "mock [command]",
	Short: "regenerate mock for project",
}

func init() {
	RootCmd.AddCommand(
		newGenTomCmd(),
		newGenYasuoCmd(),
		newGenBobCmd(),
		newGenFatimaCmd(),
		newGenEurekaCmd(),
		newGenEnigmaCmd(),
		newGenDraftCmd(),
		newGenZeusCmd(),
		newGenGolibCmd(),
		newGenUsermgmtCmd(),
		newGenPaymentCmd(),
		newGenLessonmgmtCmd(),
		newGenMasterMgmtCmd(),
		newGenEntryexitmgmtCmd(),
		newGenNotificationCmd(),
		newGenTimesheetCmd(),
		newGenInvoicemgmtCmd(),
		newGenCalendarCmd(),
		newGenVirtualClassroomCmd(),
		newGenSpikeCmd(),
		newGenDiscountCmd(),
		newGenEurekaV2Cmd(),
		newGenConversationmgmtCmd(),
		newGenAuthCmd(),
	)
}
