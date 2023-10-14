package hephaestus

import "github.com/manabie-com/backend/internal/golibs/bootstrap"

func init() {
	bootstrap.RegisterJob("hephaestus_migrate_ksql", MigrateKsql)
	bootstrap.RegisterJob("hephaestus_upsert_kafka_connect", RunUpsertKafkaConnect).
		BoolVar(&DeployCustomSinkConnector, "deployCustomSinkConnector", false, "used to deploy custom sink connector which not defined in generated connector").
		BoolVar(&SendIncrementalSnapshot, "sendIncrementalSnapshot", false, "used to send incremental snapshot - debezium do snapshot new table")
	bootstrap.RegisterJob("hephaestus_migrate_datalake", MigrateDataLake).
		StringVar(&dataLakeMigratePath, "dataLakeMigratePath", "", "migrate folder for datalake - alloydb").
		StringVar(&dataLakeName, "dataLakeName", "", "name of datalake - we're currently use alloydb")
	bootstrap.RegisterJob("hephaestus_migrate_datawarehouse", MigrateDataWarehouse).
		StringVar(&dataWarehouseMigratePath, "dataWarehouseMigratePath", "", "migrate folder for data warehouse").
		StringVar(&dataWarehouseName, "dataWarehouseName", "", "name of dataware - following partner name")

	bootstrap.RegisterJob("hephaestus_dwh_accuracy", RunAccuracyDWH).
		StringVar(&DWHResourcePath, "dwhResourcePath", "", "resource_path of data warehouse")
}
