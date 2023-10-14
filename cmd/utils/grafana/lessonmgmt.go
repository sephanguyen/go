package grafana

import (
	"github.com/spf13/cobra"
)

func genLessonMgmt(cmd *cobra.Command, args []string) error {
	return genBasicDashboard(
		destinationPath+"/backend-lessonmgmt-gen.json",
		[]string{"bob"},
		"Dashboard is generated for Lesson manager service",
		[]string{
			"bob.v1.LessonReportReaderService/RetrievePartnerDomain",
			"bob.v1.LessonReportModifierService/SubmitLessonReport",
			"bob.v1.LessonReportModifierService/SaveDraftLessonReport",
			"bob.v1.LessonManagementService/CreateLesson",
			"bob.v1.LessonManagementService/RetrieveLessons",
			"bob.v1.LessonManagementService/UpdateLesson",
			"bob.v1.LessonManagementService/DeleteLesson",
		},
	)
}

func newGenLessonMgmtCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "lessonmgmt",
		Short: "Generate grafana dashboard for lesson mgmt",
		RunE:  genLessonMgmt,
	}
}
