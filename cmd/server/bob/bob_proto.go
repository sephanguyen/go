package bob

import (
	"github.com/manabie-com/backend/internal/bob/services"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/nats"
	newNotiSvcs "github.com/manabie-com/backend/internal/notification/services"
	notigrpctrans "github.com/manabie-com/backend/internal/notification/transports/grpc"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	"google.golang.org/grpc"
)

// init services using new proto APIs
//
//nolint:interfacer
func initV1BOB(s *grpc.Server,
	eurekaDBTrace *database.DBTrace,
	dbTrace database.Ext,
	env string,
	jsm nats.JetStreamManagement,
	studyPlanReaderSvc epb.StudyPlanReaderServiceClient,
	eurekaCourseModifierSvc epb.CourseModifierServiceClient,
	eurekaAssignmentModifierSvc epb.AssignmentModifierServiceClient,
	eurekaBookReadeSvc epb.BookReaderServiceClient,
	eurekaChapterReaderSvc epb.ChapterReaderServiceClient,
	flashcardReaderSvc epb.FlashCardReaderServiceClient,
	eurekaStudyPlanReaderSvc epb.StudyPlanReaderServiceClient,
	eurekaQuizReaderSvc epb.QuizReaderServiceClient,
	eurekaQuizModifierSvc epb.QuizModifierServiceClient,
	eurekaLearningObjectiveModifierSvc epb.LearningObjectiveModifierServiceClient,
) {
	bpb.RegisterCourseModifierServiceServer(s, services.NewCourseModifierService(eurekaDBTrace, dbTrace, jsm, eurekaCourseModifierSvc, eurekaQuizReaderSvc, eurekaQuizModifierSvc, eurekaLearningObjectiveModifierSvc, env))
	bpb.RegisterCourseReaderServiceServer(s, services.NewCourseReaderService(eurekaDBTrace, dbTrace, studyPlanReaderSvc, flashcardReaderSvc, eurekaStudyPlanReaderSvc, env))

	bobNotiReaderSvc := notigrpctrans.NewBobLegacyNotificationReaderService(dbTrace, env)
	bpb.RegisterNotificationReaderServiceServer(s, bobNotiReaderSvc)

	// nolint
	bobNotiModifierSvc := notigrpctrans.NewSimpleBobLegacyNotificationModifierService(dbTrace)
	// nolint
	bpb.RegisterNotificationModifierServiceServer(s, bobNotiModifierSvc)

	// nolint
	newNotiSvc := newNotiSvcs.NewSimpleNotificationModifierService(dbTrace)
	npb.RegisterNotificationModifierServiceServer(s, newNotiSvc)
}
