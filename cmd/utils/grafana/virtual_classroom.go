package grafana

import (
	"github.com/spf13/cobra"
)

func genVirtualClassroom(cmd *cobra.Command, args []string) error {
	return genCustomDashboard(
		"cmd/utils/grafana/virtual_classroom.jsonnet",
		map[string]string{
			"virtual_classroom_grid_pos.jsonnet": "cmd/utils/grafana/virtual_classroom_grid_pos.jsonnet",
		},
		destinationPath+"/backend-virtualclassroom-gen.json",
		"Dashboard is generated for Virtual Classroom service",
		[]string{"virtualclassroom", "bob"},
		[]string{
			"bob.v1.LessonReaderService/GetLiveLessonState",
			"bob.v1.LessonModifierService/ModifyLiveLessonState",
			"manabie.bob.Course/RetrieveLiveLesson",
			"bob.v1.CourseReaderService/ListCourses",
			"bob.v1.ClassModifierService/JoinLesson",
			"bob.v1.ClassReaderService/ListStudentsByLesson",
			"bob.v1.CourseReaderService/ListLessonMedias",
			"bob.v1.ClassModifierService/LeaveLesson",
			"bob.v1.ClassModifierService/EndLiveLesson",
			"manabie.bob.Student/GetStudentProfile",
			"bob.v1.LessonModifierService/PreparePublish",
			"bob.v1.LessonModifierService/Unpublish",
			"virtualclassroom.v1.VirtualClassroomReaderService/RetrieveWhiteboardToken",
			"virtualclassroom.v1.VirtualClassroomReaderService/GetLiveLessonState",
			"virtualclassroom.v1.VirtualClassroomModifierService/JoinLiveLesson",
			"virtualclassroom.v1.VirtualClassroomModifierService/LeaveLiveLesson",
			"virtualclassroom.v1.VirtualClassroomModifierService/EndLiveLesson",
			"virtualclassroom.v1.VirtualClassroomModifierService/ModifyVirtualClassroomState",
			"virtualclassroom.v1.VirtualClassroomModifierService/PreparePublish",
			"virtualclassroom.v1.VirtualClassroomModifierService/Unpublish",
			"virtualclassroom.v1.VirtualLessonReaderService/GetLiveLessonsByLocations",
			"virtualclassroom.v1.VirtualLessonReaderService/GetLearnersByLessonID",
			"virtualclassroom.v1.VirtualLessonReaderService/GetLearnersByLessonIDs",
			"virtualclassroom.v1.VirtualLessonReaderService/GetLessons",
			"virtualclassroom.v1.LessonRecordingService/StartRecording",
			"virtualclassroom.v1.LessonRecordingService/GetRecordingByLessonID",
			"virtualclassroom.v1.LessonRecordingService/GetRecordingDownloadLinkByID",
			"virtualclassroom.v1.LessonRecordingService/StopRecording",
			"virtualclassroom.v1.LiveRoomModifierService/JoinLiveRoom",
			"virtualclassroom.v1.LiveRoomModifierService/LeaveLiveRoom",
			"virtualclassroom.v1.LiveRoomModifierService/EndLiveRoom",
			"virtualclassroom.v1.LiveRoomModifierService/ModifyLiveRoomState",
			"virtualclassroom.v1.LiveRoomModifierService/PreparePublishLiveRoom",
			"virtualclassroom.v1.LiveRoomModifierService/UnpublishLiveRoom",
			"virtualclassroom.v1.LiveRoomReaderService/GetLiveRoomState",
			"virtualclassroom.v1.LiveRoomReaderService/GetWhiteboardToken",
			"virtualclassroom.v1.ZegoCloudService/GetAuthenticationToken",
		},
	)
}

func newGenVirtualClassroomCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "virtualclassroom",
		Short: "Generate grafana dashboard for virtual classroom",
		RunE:  genVirtualClassroom,
	}
}
