package infrastructure

import (
	"context"
	"net/url"
	"time"

	"github.com/manabie-com/backend/internal/bob/services/filestore"
	"github.com/manabie-com/backend/internal/golibs/database"
	lesson_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/application/queries/payloads"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/infrastructure/repo"
	vl_payloads "github.com/manabie-com/backend/internal/virtualclassroom/modules/virtuallesson/application/queries/payloads"

	"github.com/jackc/pgtype"
)

type VirtualLessonRepo interface {
	GetVirtualLessons(ctx context.Context, db database.QueryExecer, params *vl_payloads.GetVirtualLessonsArgs) ([]*domain.VirtualLesson, int32, error)
	GetVirtualLessonByID(ctx context.Context, db database.QueryExecer, id string) (*domain.VirtualLesson, error)
	GetVirtualLessonOnlyByID(ctx context.Context, db database.QueryExecer, id string) (*domain.VirtualLesson, error)
	GetVirtualLessonsByLessonIDs(ctx context.Context, db database.QueryExecer, lessonIDs []string) ([]*domain.VirtualLesson, error)
	UpdateLessonRoomState(ctx context.Context, db database.QueryExecer, lessonID string, state *domain.OldLessonRoomState) error
	GrantRecordingPermission(ctx context.Context, db database.QueryExecer, lessonID string, recordingState []byte) error
	StopRecording(ctx context.Context, db database.QueryExecer, lessonID string, creator string, recordingState []byte) error
	GetLearnerIDsOfLesson(ctx context.Context, db database.QueryExecer, lessonID string) ([]string, error)
	GetTeacherIDsOfLesson(ctx context.Context, db database.QueryExecer, lessonID string) ([]string, error)
	GetVirtualLessonByLessonIDsAndCourseIDs(ctx context.Context, db database.QueryExecer, lessonIDs, courseIDs []string) ([]*domain.VirtualLesson, error)
	UpdateRoomID(ctx context.Context, db database.QueryExecer, lessonID, roomID string) error
	EndLiveLesson(ctx context.Context, db database.QueryExecer, lessonID string, endTime time.Time) error
	GetStreamingLearners(ctx context.Context, db database.QueryExecer, lessonID string, lockForUpdate bool) ([]string, error)
	IncreaseNumberOfStreaming(ctx context.Context, db database.QueryExecer, lessonID, learnerID string, maximumLearnerStreamings int) error
	DecreaseNumberOfStreaming(ctx context.Context, db database.QueryExecer, lessonID, learnerID string) error
	GetLessons(ctx context.Context, db database.QueryExecer, payload vl_payloads.GetLessonsArgs) (lessons []domain.VirtualLesson, total uint32, offsetID string, preTotal uint32, err error)
}

type LessonTeacherRepo interface {
	GetTeacherIDsOnlyByLessonIDs(ctx context.Context, db database.QueryExecer, lessonIDs []string) (map[string][]string, error)
	GetTeacherIDsByLessonID(ctx context.Context, db database.QueryExecer, lessonID string) ([]string, error)
	GetTeachersByLessonIDs(ctx context.Context, db database.QueryExecer, lessonIDs []string) (map[string]domain.LessonTeachers, error)
}

type UserRepo interface {
	UserGroup(context.Context, database.QueryExecer, string) (string, error)
	GetUsersByIDs(ctx context.Context, db database.QueryExecer, userIDs []string) (map[string]*domain.User, error)
}

type UserBasicInfoRepo interface {
	GetUserInfosByIDs(ctx context.Context, db database.QueryExecer, userIDs []string) ([]domain.UserBasicInfo, error)
}

type LessonGroupRepo interface {
	GetByIDAndCourseID(ctx context.Context, db database.QueryExecer, lessonGroupID, courseID string) (*repo.LessonGroupDTO, error)
	Insert(ctx context.Context, db database.QueryExecer, e *repo.LessonGroupDTO) error
}

type LessonMemberRepo interface {
	GetLessonMembersInLesson(ctx context.Context, db database.QueryExecer, lessonID string) (repo.LessonMemberDTOs, error)
	GetLessonMemberUsersByLessonID(ctx context.Context, db database.QueryExecer, params *vl_payloads.GetLessonMemberUsersByLessonIDArgs) ([]*domain.User, error)
	UpsertLessonMemberState(ctx context.Context, db database.QueryExecer, state *repo.LessonMemberStateDTO) error
	GetLessonMemberStatesWithParams(ctx context.Context, db database.QueryExecer, filter *repo.MemberStatesFilter) (repo.LessonMemberStateDTOs, error)
	GetLessonMemberStatesByLessonID(ctx context.Context, db database.QueryExecer, lessonID string) (domain.LessonMemberStates, error)
	UpsertAllLessonMemberStateByStateType(ctx context.Context, db database.QueryExecer, lessonID string, stateType domain.LearnerStateType, state *repo.StateValueDTO) error
	UpsertMultiLessonMemberStateByState(ctx context.Context, db database.QueryExecer, lessonID string, stateType domain.LearnerStateType, userIds []string, state *repo.StateValueDTO) error
	GetCourseAccessible(ctx context.Context, db database.QueryExecer, userID string) ([]string, error)
	GetLearnerIDsByLessonID(ctx context.Context, db database.QueryExecer, lessonID string) ([]string, error)
	GetLearnersByLessonIDWithPaging(ctx context.Context, db database.QueryExecer, params *vl_payloads.GetLearnersByLessonIDArgs) ([]domain.LessonMember, error)
	InsertMissingLessonMemberStateByState(ctx context.Context, db database.QueryExecer, lessonID string, stateType domain.LearnerStateType, state *repo.StateValueDTO) error
	InsertLessonMemberState(ctx context.Context, db database.QueryExecer, state *repo.LessonMemberStateDTO) error
	GetLessonLearnersByLessonIDs(ctx context.Context, db database.QueryExecer, lessonIDs []string) (map[string]domain.LessonLearners, error)
}

type VirtualLessonPollingRepo interface {
	Create(ctx context.Context, db database.Ext, poll *repo.VirtualLessonPolling) (*repo.VirtualLessonPolling, error)
}

type MediaRepo interface {
	InsertMedia(ctx context.Context, db database.QueryExecer, media *domain.Media) (*domain.Media, error)
	ListByIDs(ctx context.Context, db database.QueryExecer, mediaIDs []string) (domain.Medias, error)
	DeleteByIDs(ctx context.Context, db database.QueryExecer, mediaIDs []string) error
}

type RecordedVideoRepo interface {
	InsertRecordedVideos(ctx context.Context, db database.QueryExecer, videos []*domain.RecordedVideo) error
	ListRecordingByLessonIDWithPaging(ctx context.Context, db database.QueryExecer, payload *payloads.RetrieveRecordedVideosByLessonIDPayload) (domain.RecordedVideos, uint32, string, uint32, error)
	GetRecordingByID(ctx context.Context, db database.QueryExecer, args *payloads.GetRecordingByIDPayload) (*domain.RecordedVideo, error)
	DeleteRecording(ctx context.Context, db database.QueryExecer, recordIDs []string) error
	ListRecordingByLessonIDs(ctx context.Context, db database.QueryExecer, lessonIDs []string) (domain.RecordedVideos, error)
}

type FileStore interface {
	GetObjectInfo(ctx context.Context, bucketName, objectName string) (*filestore.StorageObject, error)
	GenerateGetObjectURL(ctx context.Context, objectName string, fileName string, expiry time.Duration) (*url.URL, error)
	DeleteObject(ctx context.Context, bucketName, objectName string) error
	GetObjectsWithPrefix(ctx context.Context, bucketName, prefix, delim string) ([]*filestore.StorageObject, error)
}

type OrganizationRepo interface {
	GetIDs(ctx context.Context, db database.QueryExecer) ([]string, error)
}

type LessonRoomStateRepo interface {
	UpsertCurrentPollingState(ctx context.Context, db database.QueryExecer, lessonID string, polling *domain.CurrentPolling) error
	UpsertWhiteboardZoomState(ctx context.Context, db database.QueryExecer, lessonID string, wbZoomState *domain.WhiteboardZoomState) error
	UpsertSpotlightState(ctx context.Context, db database.QueryExecer, lessonID string, spotlightedUser string) error
	UnSpotlight(ctx context.Context, db database.QueryExecer, lessonID string) error
	UpsertCurrentMaterialState(ctx context.Context, db database.QueryExecer, lessonID string, currentMaterial *domain.CurrentMaterial) error
	UpsertLiveLessonSessionTime(ctx context.Context, db database.QueryExecer, lessonID string, sessionTime time.Time) error
	GetLessonRoomStateByLessonID(ctx context.Context, db database.QueryExecer, lessonID string) (*domain.LessonRoomState, error)
	UpsertRecordingState(ctx context.Context, db database.QueryExecer, lessonID string, recording *domain.CompositeRecordingState) error
}

type LessonMGMTLessonRoomStateRepo interface {
	GetLessonRoomStateByLessonID(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text) (*lesson_domain.LessonRoomState, error)
}

type CourseRepo interface {
	GetValidCoursesByCourseIDsAndStatus(ctx context.Context, db database.QueryExecer, courseIDs []string, status domain.CourseStatus) (domain.Courses, error)
}

type ActivityLogRepo interface {
	Create(ctx context.Context, db database.Ext, userID, actionType string, payload map[string]interface{}) error
}

type OldClassRepo interface {
	FindJoined(ctx context.Context, db database.QueryExecer, userID string) (domain.OldClasses, error)
}

type CourseClassRepo interface {
	FindActiveCourseClassByID(ctx context.Context, db database.QueryExecer, classIDs []int32) (map[int32][]string, error)
}

type StudentEnrollmentStatusHistoryRepo interface {
	GetStatusHistoryByStudentIDsAndLocationID(ctx context.Context, db database.QueryExecer, studentIDs []string, locationID string) (domain.StudentEnrollmentStatusHistories, error)
}

type LiveLessonSentNotificationRepo interface {
	GetLiveLessonSentNotificationCount(ctx context.Context, db database.QueryExecer, lessonID, interval string) (int32, error)
	CreateLiveLessonSentNotificationRecord(ctx context.Context, db database.QueryExecer, lessonID, interval string, sentAt time.Time) error
	SoftDeleteLiveLessonSentNotificationRecord(ctx context.Context, db database.QueryExecer, lessonID string) error
}

type StudentParentRepo interface {
	GetStudentParents(ctx context.Context, db database.QueryExecer, studentIDs []string) ([]domain.StudentParent, error)
}

type ConfigRepo interface {
	GetConfigWithResourcePath(ctx context.Context, db database.QueryExecer, country domain.Country, group string, keys []string, resourcePath string) ([]*domain.Config, error)
}

type LiveLessonConversationRepo interface {
	GetConversationByLessonIDAndConvType(ctx context.Context, db database.QueryExecer, lessonID, convType string) (domain.LiveLessonConversation, error)
	GetConversationIDByExactInfo(ctx context.Context, db database.QueryExecer, lessonID string, participants []string, convType string) (string, error)
	UpsertConversation(ctx context.Context, db database.QueryExecer, conversation domain.LiveLessonConversation) error
}

type StudentsRepo interface {
	IsUserIDAStudent(ctx context.Context, db database.QueryExecer, userID string) (bool, error)
	GetStudentByStudentID(ctx context.Context, db database.QueryExecer, studentID string) (domain.Student, error)
}
