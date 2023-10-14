package tom

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/auth"
	"github.com/manabie-com/backend/internal/golibs/bootstrap"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/elastic"
	"github.com/manabie-com/backend/internal/golibs/logger"
	"github.com/manabie-com/backend/internal/tom/app/support"
	"github.com/manabie-com/backend/internal/tom/configurations"
	tomcons "github.com/manabie-com/backend/internal/tom/constants"
	"github.com/manabie-com/backend/internal/tom/repositories"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
	tpb "github.com/manabie-com/backend/pkg/manabuf/tom/v1"

	"github.com/jackc/pgtype"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var (
	schoolName string
	schoolID   string
)

func init() {
	bootstrap.RegisterJob("tom_sync_conversations_document", RunSyncConversationDocument).
		StringVar(&schoolID, "schoolID", "", "sync for specific school").
		StringVar(&schoolName, "schoolName", "", "should match with school name in secret config, for sanity check")
}

// Tom has its own org table now
func checkSchoolIDMatchSchoolName(ctx context.Context, db database.QueryExecer, schoolID, schoolName string) error {
	var count pgtype.Int8
	if err := db.QueryRow(ctx, "select count(*) from organizations where organization_id=$1 and name=$2", schoolID, schoolName).Scan(&count); err != nil {
		return err
	}
	if count.Int != 1 {
		return fmt.Errorf("school name %s with schoolID %s has %d count in db", schoolName, schoolID, count.Int)
	}
	return nil
}

// RunSyncConversationsDocument sync conversation document.
func SyncConversationDocument(
	ctx context.Context,
	c configurations.Config,
	dbTrace *database.DBTrace,
	searchClient *elastic.SearchFactoryImpl,
	schoolID, schoolName string,
	eurekaConn *grpc.ClientConn,
) {
	var isFailed bool
	zapLogger := logger.NewZapLogger("debug", c.Common.Environment == "local")

	err := checkSchoolIDMatchSchoolName(ctx, dbTrace, schoolID, schoolName)
	if err != nil {
		zapLogger.Fatal("checkSchoolIDMatchSchoolName", zap.Error(err))
	}

	// support chat
	supportChatReader := &support.ChatReader{
		SearchClient:            searchClient,
		DB:                      dbTrace,
		Logger:                  zapLogger,
		ConversationMemberRepo:  &repositories.ConversationMemberRepo{},
		ConversationStudentRepo: &repositories.ConversationStudentRepo{},
		ConversationRepo:        &repositories.ConversationRepo{},
		MessageRepo:             &repositories.MessageRepo{},
		ConversationSearchRepo:  &repositories.SearchRepo{},
		LocationRepo:            &repositories.LocationRepo{},
	}

	if err != nil {
		zapLogger.Fatal("grpc.Dial (tom)", zap.Error(err))
	}
	eurekaCourseReaderClient := epb.NewCourseReaderServiceClient(eurekaConn)

	// check the index existed or not
	isIndexExist, err := searchClient.CheckIndexExists(constants.ESConversationIndexName)
	if err != nil {
		zapLogger.Fatal("unable to check the index existed", zap.Error(err))
	}
	// create new index if not existed
	if !isIndexExist {
		zapLogger.Info(`release "conversations" index does not exist. Creating it now`)
		idxMap := strings.NewReader(tomcons.ElasticsearchConversationMapping)
		response, err := searchClient.CreateIndex(constants.ESConversationIndexName, idxMap)
		if err != nil {
			zapLogger.Error("unable to create conversations index", zap.Error(err))
		}
		defer response.Body.Close()
		if response.StatusCode != http.StatusOK {
			zapLogger.Error("unable to create conversations index", zap.String("error", response.String()))
		}
		zapLogger.Info(`"conversations" index created!`)
	}
	searchRepo := &repositories.SearchRepo{}
	searchRepo.V2()

	searchIndexer := &support.SearchIndexer{
		SearchFactory:             searchClient,
		EurekaCourseReaderService: eurekaCourseReaderClient,
		ChatReader:                supportChatReader,
		SearchRepo:                searchRepo,
		DB:                        dbTrace,
		ConversationLocationRepo:  &repositories.ConversationLocationRepo{},
	}
	var (
		nextPage      *cpb.Paging = &cpb.Paging{Limit: 200}
		nothingToSync bool
		counter       int
	)
	zapLogger.Info("Syncing conversations documents...")
	for {
		conversationIdsResponse, err := supportChatReader.ListConversationIDs(ctx, &tpb.ListConversationIDsRequest{Paging: nextPage})
		if err != nil {
			isFailed = true
			zapLogger.Fatal("ConversationReaderServiceClient(tomConn).ListConversationIDs", zap.Error(err))
		}
		nextPage = conversationIdsResponse.GetNextPage()
		if nextPage == nil || len(conversationIdsResponse.ConversationIds) == 0 {
			if counter == 0 {
				nothingToSync = true
			}
			break
		}
		_, err = searchIndexer.BuildConversationDocument(ctx, &tpb.BuildConversationDocumentRequest{
			ConversationIds: conversationIdsResponse.ConversationIds,
		})
		if err != nil {
			zapLogger.Error("unable to BuildConversationDocument", zap.Error(err))
			break
		}
		counter++
	}
	if nothingToSync {
		zapLogger.Info("Nothing to sync conversation document to elasticsearch!")
	} else {
		if isFailed {
			zapLogger.Info("Sync conversations document to elasticsearch failed!")
		} else {
			zapLogger.Info("Sync conversations document to elasticsearch completed!")
		}
	}
}

// TODO: call master data to validate schoolID and schoolName
func RunSyncConversationDocument(ctx context.Context, c configurations.Config, rsc *bootstrap.Resources) error {
	// for db RLS query
	ctx = auth.InjectFakeJwtToken(ctx, schoolID)

	SyncConversationDocument(ctx, c, rsc.DB(), rsc.Elastic(), schoolID, schoolName, rsc.GRPCDial("eureka"))
	return nil
}
