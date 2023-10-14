package mock

import (
	"path/filepath"

	"github.com/manabie-com/backend/internal/golibs/clients"
	"github.com/manabie-com/backend/internal/golibs/tools"
	live_room_repo "github.com/manabie-com/backend/internal/virtualclassroom/modules/liveroom/infrastructure/repo"
	logger_svc_repo "github.com/manabie-com/backend/internal/virtualclassroom/modules/logger/infrastructure/repo"
	virtual_lesson_repo "github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/infrastructure/repo"

	"github.com/spf13/cobra"
)

func genMockVirtualClassroomRepo(cmd *cobra.Command, args []string) error {
	virtualClassroomModuleRepos := map[string]interface{}{
		"virtual_lesson_repo":                    &virtual_lesson_repo.VirtualLessonRepo{},
		"user_repo":                              &virtual_lesson_repo.UserRepo{},
		"lesson_group_repo":                      &virtual_lesson_repo.LessonGroupRepo{},
		"lesson_member_repo":                     &virtual_lesson_repo.LessonMemberRepo{},
		"lesson_teacher_repo":                    &virtual_lesson_repo.LessonTeacherRepo{},
		"logger_svc_repo":                        &logger_svc_repo.VirtualClassroomLogRepo{},
		"lesson_room_state":                      &virtual_lesson_repo.LessonRoomStateRepo{},
		"media_repo":                             &virtual_lesson_repo.MediaRepo{},
		"recorded_video":                         &virtual_lesson_repo.RecordedVideoRepo{},
		"organization_repo":                      &virtual_lesson_repo.OrganizationRepo{},
		"course_repo":                            &virtual_lesson_repo.CourseRepo{},
		"activity_log_repo":                      &virtual_lesson_repo.ActivityLogRepo{},
		"old_class_repo":                         &virtual_lesson_repo.OldClassRepo{},
		"course_class_repo":                      &virtual_lesson_repo.CourseClassRepo{},
		"student_enrollment_status_history_repo": &virtual_lesson_repo.StudentEnrollmentStatusHistoryRepo{},
		"live_lesson_sent_notification_repo":     &virtual_lesson_repo.LiveLessonSentNotificationRepo{},
		"config_repo":                            &virtual_lesson_repo.ConfigRepo{},
		"live_lesson_conversation_repo":          &virtual_lesson_repo.LiveLessonConversationRepo{},
		"user_basic_info_repo":                   &virtual_lesson_repo.UserBasicInfoRepo{},
		"students_repo":                          &virtual_lesson_repo.StudentsRepo{},
		"student_parent_repo":                    &virtual_lesson_repo.StudentParentRepo{},
	}
	tools.MockRepository("mock_repositories", filepath.Join(args[0], "virtualclassroom/repositories"), "virtualclassroom", virtualClassroomModuleRepos)

	virtualClassroomModuleClients := map[string]interface{}{
		"conversation_client": &clients.ConversationClient{},
	}
	tools.MockRepository("mock_clients", filepath.Join(args[0], "virtualclassroom/clients"), "virtualclassroom", virtualClassroomModuleClients)

	liveRoomModuleRepos := map[string]interface{}{
		"live_room_repo":                 &live_room_repo.LiveRoomRepo{},
		"live_room_logger_repo":          &logger_svc_repo.LiveRoomLogRepo{},
		"live_room_state_repo":           &live_room_repo.LiveRoomStateRepo{},
		"live_room_member_state_repo":    &live_room_repo.LiveRoomMemberStateRepo{},
		"live_room_poll_repo":            &live_room_repo.LiveRoomPollRepo{},
		"live_room_recorded_videos_repo": &live_room_repo.LiveRoomRecordedVideosRepo{},
		"live_room_activity_log_repo":    &live_room_repo.LiveRoomActivityLogRepo{},
	}
	tools.MockRepository("mock_repositories", filepath.Join(args[0], "liveroom/repositories"), "liveroom", liveRoomModuleRepos)

	return nil
}

func newGenVirtualClassroomCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "virtualclassroom [../../mock/virtualclassroom]",
		Short: "generate virtualclassroom repository type",
		Args:  cobra.ExactArgs(1),
		RunE:  genMockVirtualClassroomRepo,
	}
}
