package helpers

import (
	"math/rand"

	"github.com/manabie-com/backend/features/common"
	"github.com/manabie-com/backend/features/eibanam/communication/util"
	"github.com/manabie-com/backend/internal/golibs/configs"
	"github.com/manabie-com/backend/internal/golibs/kafka"
	"github.com/manabie-com/backend/internal/golibs/nats"

	"github.com/jackc/pgx/v4/pgxpool"
	"google.golang.org/grpc"
)

type ConversationMgmtHelper struct {
	BobDBConn                      *pgxpool.Pool
	AuthDBConn                     *pgxpool.Pool
	BobPostgresDBConn              *pgxpool.Pool
	FatimaDBConn                   *pgxpool.Pool
	EurekaDBConn                   *pgxpool.Pool
	NotificationMgmtDBConn         *pgxpool.Pool
	NotificationMgmtPostgresDBConn *pgxpool.Pool
	BobGRPCConn                    *grpc.ClientConn
	YasuoGRPCConn                  *grpc.ClientConn
	TomGRPCConn                    *grpc.ClientConn
	UserMgmtGRPCConn               *grpc.ClientConn
	NotificationMgmtGRPCConn       *grpc.ClientConn
	SpikeGRPCConn                  *grpc.ClientConn
	ShamirGRPCConn                 *grpc.ClientConn
	MasterMgmtGRPCConn             *grpc.ClientConn
	ConversationMgmtGRPCConn       *grpc.ClientConn
	FirebaseAddress                string
	JSM                            nats.JetStreamManagement
	Kafka                          kafka.KafkaManagement
	Storage                        configs.StorageConfig
	ApplicantID                    string
	exampleNames                   []string

	// ConversationMgmt svc use Tom DB
	TomDBConn         *pgxpool.Pool
	TomPostgresDBConn *pgxpool.Pool
}

func NewConversationMgmtHelper(
	firebaseAddress string,
	applicantID string,
	connections *common.Connections,
	cfg *common.Config,
) *ConversationMgmtHelper {
	exampleName, _ := util.LoadExampleName()

	return &ConversationMgmtHelper{
		BobDBConn:                      connections.BobDB,
		AuthDBConn:                     connections.AuthPostgresDB,
		BobPostgresDBConn:              connections.BobPostgresDB,
		FatimaDBConn:                   connections.FatimaDB,
		EurekaDBConn:                   connections.EurekaDB,
		NotificationMgmtDBConn:         connections.NotificationMgmtDB,
		NotificationMgmtPostgresDBConn: connections.NotificationMgmtPostgresDB,
		BobGRPCConn:                    connections.BobConn,
		YasuoGRPCConn:                  connections.YasuoConn,
		TomGRPCConn:                    connections.TomConn,
		UserMgmtGRPCConn:               connections.UserMgmtConn,
		NotificationMgmtGRPCConn:       connections.NotificationMgmtConn,
		SpikeGRPCConn:                  connections.SpikeConn,
		ShamirGRPCConn:                 connections.ShamirConn,
		MasterMgmtGRPCConn:             connections.MasterMgmtConn,
		ConversationMgmtGRPCConn:       connections.ConversationMgmtConn,
		FirebaseAddress:                firebaseAddress,
		ApplicantID:                    applicantID,
		JSM:                            connections.JSM,
		Kafka:                          connections.Kafka,
		Storage:                        cfg.Storage,
		exampleNames:                   exampleName,
		TomDBConn:                      connections.TomDB,
		TomPostgresDBConn:              connections.TomPostgresDB,
	}
}

func (helper *ConversationMgmtHelper) PickName() string {
	l := len(helper.exampleNames)
	if l < 1 {
		return ""
	}
	// nolint
	return helper.exampleNames[rand.Intn(l)]
}
