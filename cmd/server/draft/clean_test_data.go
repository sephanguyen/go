package draft

import (
	"bytes"
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/manabie-com/backend/internal/draft/configurations"
	"github.com/manabie-com/backend/internal/golibs/bootstrap"
	"github.com/manabie-com/backend/internal/golibs/configs"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	vr "github.com/manabie-com/backend/internal/golibs/variants"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type RegisterSubscriber struct {
	JSM    nats.JetStreamManagement
	Config configurations.Config
	Rsc    *bootstrap.Resources
}

func (r *RegisterSubscriber) Subscribe() error {
	opts := nats.Option{
		JetStreamOptions: []nats.JSSubOption{
			nats.AckAll(),
			nats.Bind(constants.StreamCleanDataTestEventNats, constants.DurableArchitectureCleanDataTest),
			nats.MaxDeliver(1),
			nats.DeliverSubject(constants.DeliverArchitectureCleanDataTes),
			nats.AckWait(30 * time.Second),
		},
		SpanName: "CleanData.Created",
	}
	_, err := r.JSM.QueueSubscribe(
		constants.SubjectCleanDataTestEventNats,
		constants.QueueArchitectureCleanDataTes,
		opts,
		r.Handle,
	)
	if err != nil {
		return fmt.Errorf("error subscribing to subject `%s` on `%s` queue: %w",
			constants.SubjectCleanDataTestEventNats,
			constants.QueueArchitectureCleanDataTes,
			err,
		)
	}
	return nil
}

func (r *RegisterSubscriber) Handle(ctx context.Context, msg []byte) (bool, error) {
	cleanEvent := &npb.EventDataClean{}

	if err := proto.Unmarshal(msg, cleanEvent); err != nil {
		return false, err
	}

	cleanupTable = cleanEvent.Tables
	cleanupService = cleanEvent.Service
	schoolID = cleanEvent.SchoolId
	perBatch = int(cleanEvent.PerBatch)
	dryRun = false
	cleanupBefore = cleanEvent.BeforeAt
	cleanupAfter = cleanEvent.AfterAt
	cleanupExtraCond = cleanEvent.ExtraCond
	err := cleanTestData(ctx, r.Config, r.Rsc)
	if err != nil {
		return false, err
	}
	return false, nil
}

var (
	cleanupTable     string
	cleanupService   string
	cleanupBefore    string
	schoolID         string
	dryRun           bool
	batchEnabled     bool
	perBatch         int
	cleanupAfter     string
	cleanupExtraCond []*npb.ExtraCond
	reservedTable    = map[string]struct{}{
		"organizations":     {},
		"organization_auth": {},
	}
)

type Config = configurations.CleanTableConfig

func init() {
	bootstrap.RegisterJob("clean_test_data", cleanTestData).
		StringVar(&cleanupTable, "tables", "", "list of tables to clean up").
		StringVar(&cleanupService, "service", "", "which db to run on").
		StringVar(&cleanupBefore, "before", "", "clean up from which timestamp").
		StringVar(&schoolID, "schoolID", "-2147483644", "Which school to clean for").
		BoolVar(&dryRun, "dryRun", false, "Print raw only, to run manually").
		BoolVar(&batchEnabled, "batchEnabled", false, "Execute in batch").
		IntVar(&perBatch, "perBatch", 1000, "Items per batch")

	maybeCTETemplate, err := template.New("cte").Parse(selfRefCTETemplateStr)
	if err != nil {
		panic(err)
	}
	selfRefCTETemplate = maybeCTETemplate
}

func parseTime(timeStr string) (time.Time, error) {
	reg := regexp.MustCompile(`(\d+) (day|month|days|months) ago`)
	if reg.Match([]byte(timeStr)) {
		matches := reg.FindAllStringSubmatch(timeStr, -1)[0]
		num, err := strconv.Atoi(matches[1])
		if err != nil {
			return time.Time{}, err
		}
		unit := matches[2]
		switch unit {
		case "day", "days":
			return time.Now().Add(-1 * time.Hour * 24 * time.Duration(num)), nil
		default:
			return time.Now().AddDate(0, -1*num, 0), nil
		}
	}
	return time.Parse(time.RFC3339, timeStr)
}

func cleanTestData(ctx context.Context, c configurations.Config, rsc *bootstrap.Resources) error {
	serviceConfig := c.DataPruneConfig.ServiceConfigs
	if len(cleanupExtraCond) > 0 {
		serviceConfig = convertToExtraConfig(c.DataPruneConfig.ServiceConfigs, cleanupExtraCond)
	}
	cleanupBeforeTime, err := parseTime(cleanupBefore)
	if err != nil {
		return err
	}
	var cleanupAfterTime time.Time
	if cleanupAfter != "" {
		cleanupAfterTime, err = parseTime(cleanupAfter)
		if err != nil {
			return err
		}
	}

	db, dbcancel, err := getDBForService(ctx, rsc.Logger(), c, cleanupService)
	if err != nil {
		return fmt.Errorf("failed to connect to database for service %s: %s", cleanupService, err)
	}
	defer db.Close()
	defer func() {
		if err := dbcancel(); err != nil {
			rsc.Logger().Error("dbcancel() failed", zap.Error(err))
		}
	}()

	tbls := strings.Split(cleanupTable, ",")
	for _, tbl := range tbls {
		rsc.Logger().Info("starting cleanup data", zap.String("table", tbl))

		graph, err := buildDepGraph(ctx, db, serviceConfig, tbl)
		if err != nil {
			return err
		}
		ctxzap.Extract(ctx).Debug("dependency graph", zap.Reflect("graph", graph))

		rootTableConf := serviceConfig[tbl]
		if len(rootTableConf.SelfRefFKs) > 0 {
			// must query children row and add to predicate of this table

			err = queryChildrenAndRecursiveDelete(ctx, serviceConfig, tbl, db, graph, cleanupBeforeTime, batchEnabled, perBatch, cleanupAfterTime)
			if err != nil {
				return err
			}
		} else {
			err = recursiveDelete(ctx, serviceConfig, tbl, db, graph, database.Timestamptz(cleanupBeforeTime), batchEnabled, perBatch, database.Timestamptz(cleanupAfterTime))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func convertToExtraConfig(config map[string]configurations.CleanTableConfig, externalCond []*npb.ExtraCond) map[string]configurations.CleanTableConfig {
	conf := make(map[string]configurations.CleanTableConfig, len(config))
	for key, val := range config {
		conf[key] = val
		for _, e := range externalCond {
			if e.Table == key {
				conf[key] = configurations.CleanTableConfig{
					CreatedAtColName:    val.CreatedAtColName,
					ExtraCond:           val.ExtraCond + e.Condition,
					IgnoreFks:           val.IgnoreFks,
					SelfRefFKs:          val.SelfRefFKs,
					SetNullOnCircularFk: val.SetNullOnCircularFk,
				}
			}
		}
	}
	return conf
}

func getDBForService(ctx context.Context, l *zap.Logger, c configurations.Config, svcName string) (*pgxpool.Pool, func() error, error) {
	var dbconf configs.PostgresDatabaseConfig
	switch svcName {
	case "eureka":
		dbconf = c.DataPruneConfig.PostgresLMSInstance
	case "tom", "bob", "invoicemgmt", "timesheet", "lessonmgmt":
		dbconf = c.DataPruneConfig.PostgresCommonInstance
	default:
		return nil, nil, fmt.Errorf("unknown database %q (did you add config for it?)", svcName)
	}
	dbconf.DBName = vr.DatabaseNamePrefix(vr.ToPartner(c.Common.Organization), vr.ToEnv(c.Common.Environment)) + svcName
	db, dbcancel, err := database.NewPool(ctx, l, dbconf)
	if err != nil {
		return nil, nil, err
	}
	return db, dbcancel, nil
}

type stackitem struct {
	node           GraphNode
	dependencyList []GraphNode
}

// Depth first search
func recursiveDelete(ctx context.Context,
	tableConfigs map[string]configurations.CleanTableConfig,
	higestNode string,
	db database.Ext,
	graph map[string][]GraphNode,
	deleteBefore pgtype.Timestamptz,
	isBatch bool,
	itemsPerBatch int,
	deleteAfter pgtype.Timestamptz,
	extraArgs ...interface{},
) error {
	// for example with this graph
	// t1 _ t2 _ t3
	//   |_ t4 _ t5
	// dfs will have execution order t3,t2,t5,t4,t1
	stack := []*stackitem{
		{
			node:           GraphNode{TblName: higestNode},
			dependencyList: graph[higestNode],
		},
	}
	for len(stack) > 0 {
		curNode := stack[len(stack)-1]
		if len(curNode.dependencyList) == 0 {
			// deleting from this table is safe from fk constraint because it has no dependency
			err := deleteForTable(ctx, db, tableConfigs, graph, stack, deleteBefore, isBatch, itemsPerBatch, deleteAfter, extraArgs...)
			if err != nil {
				return err
			}
			// delete(graph, curNode.node.TblName)
			stack = stack[:len(stack)-1]
		} else {
			nextNode := curNode.dependencyList[0]
			curNode.dependencyList = curNode.dependencyList[1:]
			// need to continue checking dependency of this node in the next iteration
			stack = append(stack, &stackitem{dependencyList: graph[nextNode.TblName], node: nextNode})
		}
	}
	return nil
}

var (
	// depth < 10 to avoid dead cycle some how
	// l2.referencedCol != l2.referencingCol to avoid infinite query loop (a row join with itself forever)
	selfRefCTETemplateStr = `
with recursive find_ent_family as (
select 1 as depth,l.{{.ReferencedColumn}},l.{{.ReferencingColumn}} from {{.Table}} l
	{{.Predicate}}
	union all select temp.depth + 1 as depth,l2.{{.ReferencedColumn}},l2.{{.ReferencingColumn}} from {{.Table}} l2 
	join find_ent_family temp 
	on temp.{{.ReferencedColumn}}=l2.{{.ReferencingColumn}} where depth < 10 and l2.{{.ReferencedColumn}} != l2.{{.ReferencingColumn}}
)
select array_agg(batched.{{.ReferencedColumn}}) from (
	select {{.ReferencedColumn}} from find_ent_family order by depth desc, {{.ReferencedColumn}} desc {{.Limit}}
) as batched`

	selfRefCTETemplate *template.Template
)

// use CTE to find ids of all rows (rows matching initial predicated, and its children), then use predicate where id=any($id_list),
// to execute recursive delete on those entities
func queryChildrenAndRecursiveDelete(
	ctx context.Context,
	tableConfigs map[string]configurations.CleanTableConfig,
	highestNode string,
	db database.Ext,
	graph map[string][]GraphNode,
	deleteBefore time.Time,
	isBatch bool,
	itemsPerBatch int,
	deleteAfter time.Time,
) error {
	tblConf := tableConfigs[highestNode]
	// a table has multiple self referencing fks, too complex
	if len(tblConf.SelfRefFKs) != 1 {
		panic(fmt.Sprintf("table %s has too may self reference foreign key", highestNode))
	}
	selfRefFk := tblConf.SelfRefFKs[0]
	var buf bytes.Buffer
	tsColname := "created_at"
	if tblConf.CreatedAtColName != "" {
		tsColname = tblConf.CreatedAtColName
	}

	queryCond := fmt.Sprintf("%s < $1 and %s > $2 ", tsColname, tsColname)
	if tblConf.ExtraCond != "" {
		queryCond += " " + tblConf.ExtraCond
	}
	queryCond = fmt.Sprintf("where (%s) and resource_path='%s'", queryCond, schoolID)
	args := map[string]string{
		"ReferencingColumn": selfRefFk.Referencing,
		"ReferencedColumn":  selfRefFk.Referenced,
		"Table":             highestNode,
		"Predicate":         queryCond,
		"Limit":             "",
	}
	if isBatch {
		args["Limit"] = fmt.Sprintf("limit %d", itemsPerBatch)
	}
	err := selfRefCTETemplate.Execute(&buf, args)
	if err != nil {
		return err
	}
	query := buf.String()
	fmt.Println(query)
	for {
		var ids pgtype.TextArray
		// find ids of children entities
		err = db.QueryRow(ctx, query, deleteBefore, deleteAfter).Scan(&ids)
		if err != nil {
			return err
		}
		if len(database.FromTextArray(ids)) == 0 {
			return nil
		}
		// use OR to ignore where created_at < $timestamp condition
		// we can overwrite the old extraCond, because the ids already the result from that condition
		tblConf.ExtraCond = fmt.Sprintf("and %s=any($3)", selfRefFk.Referenced)
		tableConfigs[highestNode] = tblConf
		nullts := pgtype.Timestamptz{Status: pgtype.Null}
		err = recursiveDelete(ctx, tableConfigs, highestNode, db, graph, nullts, isBatch, itemsPerBatch, nullts, ids)
		if err != nil {
			return err
		}
	}
}

func setNullOnColumnForTable(
	ctx context.Context,
	db database.Ext,
	cfg map[string]configurations.CleanTableConfig,
	setNullCol string,
	node GraphNode,
	stack []*stackitem,
	deleteBefore pgtype.Timestamptz,
	isBatch bool,
	itemsPerBatch int,
	deleteAfter pgtype.Timestamptz,
	extraArgs ...interface{},
) error {
	tsColname := "created_at"
	root := stack[0]
	specialConf := cfg[root.node.TblName]
	if specialConf.CreatedAtColName != "" {
		tsColname = specialConf.CreatedAtColName
	}

	additionalCond := specialConf.ExtraCond
	outerCond := fmt.Sprintf("(($1::timestamptz is null or %s < $1) and ($2::timestamptz is null or %s > $2))", tsColname, tsColname)
	if additionalCond != "" {
		outerCond += " " + additionalCond
	}

	outerCond = fmt.Sprintf("where (%s) and resource_path='%s'", outerCond, schoolID)
	if len(stack) > 4 {
		panic(fmt.Sprintf("dependency depth is too high: %v on table %s, column %s, abort", stack, root.node.TblName, root.node.ReferencingColName))
	}
	if len(stack) > 1 {
		for idx := range stack {
			if idx == 0 {
				continue
			}
			item := stack[idx]
			thisTable := item.node.TblName
			specialConf := cfg[thisTable]
			additionalCond = specialConf.ExtraCond
			outerCond = fmt.Sprintf("where %s=any(select %s from %s %s)",
				item.node.ReferencingColName, item.node.RemoteColName, item.node.RemoteTblName, outerCond)
			if additionalCond != "" {
				outerCond += " " + additionalCond
			}
		}
	}
	outerCond += fmt.Sprintf(" and %s is not null", setNullCol)
	if isBatch {
		outerCond += fmt.Sprintf(" limit %d", itemsPerBatch)
	}

	setNullTemplate := "update %s set %s=null where ctid in (select ctid from %s %s)"
	fullTemplate := fmt.Sprintf(setNullTemplate, node.TblName, setNullCol, node.TblName, outerCond)
	fmt.Println(fullTemplate)

	if !dryRun {
		var err error
		newArgs := append([]interface{}{deleteBefore, deleteAfter}, extraArgs...)
		if isBatch {
			err = executeInBatch(ctx, node.TblName, db, fullTemplate, newArgs...)
		} else {
			_, err = db.Exec(ctx, fullTemplate, newArgs...)
		}
		if err != nil {
			return err
		}
	}

	return nil
}

// Construct a delete script with a nested subquery predicate
// for example t1 ref t2 ref t3 will results in a predicate like `where ctid in (select ctid from t1 ... where t1.ref=(select refed from t2 where ...)`
// if t1 has a self-ref fk, we first delete the rows referencing the row of t1 matching the predicate above using the selfRefCTETemplate
// then we can safely delete t1 using the above predicate
func deleteForTable(ctx context.Context, db database.Ext,
	cfg map[string]configurations.CleanTableConfig,
	graph map[string][]GraphNode,
	stack []*stackitem,
	deleteBefore pgtype.Timestamptz,
	isBatch bool,
	itemsPerBatch int,
	deleteAfter pgtype.Timestamptz,
	extraArgs ...interface{},
) error {
	var (
		tsColname      = "created_at"
		additionalCond = ""
	)
	deletedNode := stack[len(stack)-1].node
	deletedNodeConf := cfg[deletedNode.TblName]
	outerCond := ""
	if len(stack) > 4 {
		panic(fmt.Sprintf("dependency depth is too high: %v on table %s, column %s, abort", stack, deletedNode.TblName, deletedNode.ReferencingColName))
	}
	for idx := range stack {
		curNode := stack[idx].node
		curTable := curNode.TblName
		curConf := cfg[curTable]
		if setnullCol, exist := deletedNodeConf.SetNullOnCircularFk[curTable]; exist {
			// if we don't set null, we face fk constraint violation later
			// because this table has circular fk
			err := setNullOnColumnForTable(ctx, db, cfg, setnullCol, curNode, stack[:idx+1], deleteBefore, isBatch, itemsPerBatch, deleteAfter, extraArgs...)
			if err != nil {
				return err
			}
		}
		if idx == 0 {
			if curConf.CreatedAtColName != "" {
				tsColname = curConf.CreatedAtColName
			}

			additionalCond = curConf.ExtraCond
			outerCond = fmt.Sprintf("(($1::timestamptz is null or %s < $1) and ($2::timestamptz is null or %s > $2))", tsColname, tsColname)
			if additionalCond != "" {
				outerCond += " " + additionalCond
			}
			outerCond = fmt.Sprintf("where (%s) and resource_path='%s'", outerCond, schoolID)
			continue
		}

		additionalCond = curConf.ExtraCond
		outerCond = fmt.Sprintf("where %s=any(select %s from %s %s)",
			curNode.ReferencingColName, curNode.RemoteColName, curNode.RemoteTblName, outerCond)
		if additionalCond != "" {
			outerCond += " " + additionalCond
		}
	}
	if isBatch {
		outerCond += fmt.Sprintf(" limit %d", itemsPerBatch)
	}

	deleteTemplate := "delete from %s where ctid in (select ctid from %s %s)"
	fullTemplate := fmt.Sprintf(deleteTemplate, deletedNode.TblName, deletedNode.TblName, outerCond)
	fmt.Println(fullTemplate)

	if !dryRun {
		var err error
		newArgs := append([]interface{}{deleteBefore, deleteAfter}, extraArgs...)
		if isBatch {
			err = executeInBatch(ctx, deletedNode.TblName, db, fullTemplate, newArgs...)
		} else {
			_, err = db.Exec(ctx, fullTemplate, newArgs...)
		}
		if err != nil {
			return err
		}
	}

	return nil
}

func executeInBatch(ctx context.Context, tblname string, db database.Ext, fullQuery string, args ...interface{}) error {
	var totalRemoved int
	for {
		ret, err := db.Exec(ctx, fullQuery, args...)
		if err != nil {
			return err
		}
		totalRemoved += int(ret.RowsAffected())
		fmt.Printf("Total rows removed from table %s: %d\n", tblname, totalRemoved)

		if ret.RowsAffected() == 0 {
			return nil
		}
	}
}

type GraphNode struct {
	TblName            string
	ReferencingColName string
	RemoteTblName      string
	RemoteColName      string
}

func buildDepGraph(
	ctx context.Context,
	db database.Ext,
	cfg map[string]configurations.CleanTableConfig,
	tblname string,
) (map[string][]GraphNode, error) {
	ctxzap.Extract(ctx).Debug("building dependency graph", zap.Reflect("config", cfg), zap.String("table_name", tblname))
	rows, err := db.Query(ctx, fkhierarchyv2, database.Text(tblname))
	if err != nil {
		return nil, err
	}
	depMap := map[string][]GraphNode{}

	for rows.Next() {
		var (
			depth            int64
			schema           string
			referencingTable string
			referencingCols  pgtype.TextArray
			remoteTable      string
			remoteCols       pgtype.TextArray
		)

		err := rows.Scan(&depth, &schema, &referencingTable, &referencingCols, &remoteTable, &remoteCols)
		if err != nil {
			return nil, fmt.Errorf("rows.Scan %w", err)
		}
		if _, exist := reservedTable[referencingTable]; exist {
			return nil, fmt.Errorf("table %s's data need to be cleaned up before deleting table %s in the chain, but it is reserved", referencingTable, remoteTable)
		}
		for idx, referencingCol := range database.FromTextArray(referencingCols) {
			if sliceutils.Contains(cfg[referencingTable].IgnoreFks, referencingCol) {
				continue
			}
			// one fk may have multiple graphnode, for example (user_id,lesson_id) will output 2 graph node
			n := GraphNode{
				TblName:            referencingTable,
				ReferencingColName: referencingCol,
				RemoteTblName:      remoteTable,
				RemoteColName:      database.FromTextArray(remoteCols)[idx],
			}

			// in case we have a cycle in the graph (t1->t2->t3->t1->t2...)
			if sliceutils.Contains(depMap[remoteTable], n) {
				continue
			}
			depMap[n.RemoteTblName] = append(depMap[n.RemoteTblName], n)
		}

	}
	return depMap, nil
}

// referencing_columns and remote_columns is order by the conrelid and confrelid, so that they have identical
// order regarding the component column, for example referencing cols (user_id,lesson_id) will go with remote cols of (user_id,lesson_id)
var fkhierarchyv2 = `
WITH RECURSIVE FK_HIERARCHY AS (
  SELECT 
    1 AS DEPTH, 
    PGC.CONNAME, 
    PGC.CONRELID, 
    PGC.CONFRELID, 
    PGN.NSPNAME AS TABLE_SCHEMA, 
    PUT.RELNAME AS TABLE_NAME, 
    PGC.CONKEY, 
    PUT2.RELNAME AS FOREIGN_TABLE_NAME, 
    PGC.CONFKEY 
  FROM 
    PG_CONSTRAINT PGC 
    JOIN PG_NAMESPACE PGN ON PGC.CONNAMESPACE = PGN.OID 
    JOIN PG_STATIO_USER_TABLES PUT2 ON PUT2.RELID = PGC.CONFRELID 
    JOIN PG_STATIO_USER_TABLES PUT ON PUT.RELID = PGC.CONRELID 
  WHERE 
    PGC.CONTYPE = 'f' 
    AND PUT2.RELNAME = $1 
  GROUP BY 
    (
      PGC.CONNAME, PGC.CONRELID, PGC.CONFRELID, 
      PGN.NSPNAME, PUT.RELNAME, PUT2.RELNAME, 
      PGC.CONKEY, PGC.CONFKEY, DEPTH
    ) 
  UNION ALL 
  SELECT 
    1 + TEMP.DEPTH AS DEPTH, 
    SPGC.CONNAME, 
    SPGC.CONRELID, 
    SPGC.CONFRELID, 
    SPGN.NSPNAME AS TABLE_SCHEMA, 
    SPUT.RELNAME AS TABLE_NAME, 
    SPGC.CONKEY, 
    SPUT2.RELNAME AS FOREIGN_TABLE_NAME, 
    SPGC.CONFKEY 
  FROM 
    PG_CONSTRAINT SPGC 
    JOIN PG_NAMESPACE SPGN ON SPGC.CONNAMESPACE = SPGN.OID 
    JOIN PG_STATIO_USER_TABLES SPUT2 ON SPUT2.RELID = SPGC.CONFRELID 
    JOIN PG_STATIO_USER_TABLES SPUT ON SPUT.RELID = SPGC.CONRELID 
    JOIN FK_HIERARCHY AS TEMP ON SPGN.NSPNAME = TEMP.TABLE_SCHEMA 
    AND SPUT2.RELNAME = TEMP.TABLE_NAME 
  WHERE 
    SPGC.CONTYPE = 'f' 
    AND TEMP.DEPTH < 4 
  GROUP BY 
    (
      SPGC.CONNAME, SPGC.CONRELID, SPGC.CONFRELID, 
      SPGN.NSPNAME, SPUT.RELNAME, SPUT2.RELNAME, 
      SPGC.CONKEY, SPGC.CONFKEY, DEPTH
    )
) 
SELECT 
  TREE.DEPTH, 
  TREE.TABLE_SCHEMA, 
  TREE.TABLE_NAME, 
  (
    SELECT 
      ARRAY_AGG(
        PGA.ATTNAME 
        ORDER BY 
          MULTICOL_FK_ORDER.FK_ORDER
      ) 
    FROM 
      PG_ATTRIBUTE PGA 
      JOIN UNNEST(TREE.CONKEY) WITH ORDINALITY AS MULTICOL_FK_ORDER(ID, FK_ORDER) ON PGA.ATTNUM = MULTICOL_FK_ORDER.ID 
      AND PGA.ATTRELID = TREE.CONRELID
  ) AS REFERENCING_COLUMNS, 
  TREE.FOREIGN_TABLE_NAME, 
  (
    SELECT 
      ARRAY_AGG(
        PGA.ATTNAME 
        ORDER BY 
          MULTICOL_FK_ORDER.FK_ORDER
      ) 
    FROM 
      PG_ATTRIBUTE PGA 
      JOIN UNNEST(TREE.CONFKEY) WITH ORDINALITY AS MULTICOL_FK_ORDER(ID, FK_ORDER) ON PGA.ATTNUM = MULTICOL_FK_ORDER.ID 
      AND PGA.ATTRELID = TREE.CONFRELID
  ) AS REMOTE_COLUMNS 
FROM 
  FK_HIERARCHY AS TREE
`
