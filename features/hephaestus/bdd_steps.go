package hephaestus

import (
	"regexp"
	"sync"

	"github.com/manabie-com/backend/features/helper"

	"github.com/cucumber/godog"
)

var (
	buildRegexpMapOnce sync.Once
	regexpMap          map[string]*regexp.Regexp
)

func initSteps(ctx *godog.ScenarioContext, s *suite) {
	steps := map[string]interface{}{
		`^delete debezium source connector$`:                                 s.deleteDebeziumSourceConnector,
		`^the data insert before will not be synced$`:                        s.theDataInsertBeforeWillNotBeSynced,
		`^create debezium source connector for that table "([^"]*)" in bob$`: s.createDebeziumSourceConnectorForThatTableInBob,
		`^create sink connector for that table "([^"]*)" in fatima$`:         s.createSinkConnectorForThatTableInFatima,
		`^run job upsert kafka connector$`:                                   s.runJobUpsertKafkaConnector,
		`^the data will be synced$`:                                          s.theDataWillBeSynced,
		`^table "([^"]*)" in database bob and fatima$`:                       s.tableInDatabaseBobAndFatima,
		`^source connector file for table "([^"]*)" in bob$`:                 s.sourceConnectorFileForTableInBob,
		`^sink connector file for table "([^"]*)" in fatima$`:                s.sinkConnectorFileForTableInFatima,
		`^insert several records to table "([^"]*)"$`:                        s.insertSeveralRecordsToTable,
		`^records is synced in source and sink table "([^"]*)"$`:             s.recordsIsSyncedInSourceAndSinkTable,
		`^add table "([^"]*)" to captured table list in source connector$`:   s.addTableToCapturedTableListInSourceConnector,
	}
	buildRegexpMapOnce.Do(func() { regexpMap = helper.BuildRegexpMapV2(steps) })
	for k, v := range steps {
		ctx.Step(regexpMap[k], v)
	}
}
