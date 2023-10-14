package draft

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/draft/configurations"
	"github.com/manabie-com/backend/internal/golibs/database"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TblName            string
// ReferencingColName string
// RemoteTblName      string
// RemoteColName      string

func makeNodes(remoteTable string, tblnames []string) []GraphNode {
	ret := make([]GraphNode, 0, len(tblnames))
	for _, tblname := range tblnames {
		ret = append(ret, GraphNode{
			TblName:            tblname,
			ReferencingColName: "ref_col",
			RemoteColName:      "id",
			RemoteTblName:      remoteTable,
		})
	}
	return ret
}

func Test_recursiveDeleteBatched(t *testing.T) {
	depGraph := map[string][]GraphNode{
		"t4": makeNodes("t4", []string{"t3", "t2"}),
		"t2": makeNodes("t2", []string{"t1"}),
		"t3": makeNodes("t3", []string{"t1"}),
	}
	db := &mock_database.Ext{}
	deleteBefore := database.Timestamptz(time.Now())
	deleteAfter := database.Timestamptz(time.Now())
	extraConf := map[string]Config{
		"t4": {
			CreatedAtColName: "is_created_at",
			ExtraCond:        "and some_t4_id not in ('some_id')",
		},
		"t3": {
			CreatedAtColName: "is_created_at",
			ExtraCond:        "and some_t3_id not in ('some_id')",
		},
	}

	db.On("Exec", mock.Anything, "delete from t1 where ctid in (select ctid from t1 where ref_col=any(select id from t3 where ref_col=any(select id from t4 where ((($1::timestamptz is null or is_created_at < $1) and ($2::timestamptz is null or is_created_at > $2)) and some_t4_id not in ('some_id')) and resource_path='-2147483644') and some_t3_id not in ('some_id')) limit 100)", deleteBefore, deleteAfter).Once().Return(nil, nil)
	db.On("Exec", mock.Anything, "delete from t3 where ctid in (select ctid from t3 where ref_col=any(select id from t4 where ((($1::timestamptz is null or is_created_at < $1) and ($2::timestamptz is null or is_created_at > $2)) and some_t4_id not in ('some_id')) and resource_path='-2147483644') and some_t3_id not in ('some_id') limit 100)", deleteBefore, deleteAfter).Once().Return(nil, nil)
	db.On("Exec", mock.Anything, "delete from t1 where ctid in (select ctid from t1 where ref_col=any(select id from t2 where ref_col=any(select id from t4 where ((($1::timestamptz is null or is_created_at < $1) and ($2::timestamptz is null or is_created_at > $2)) and some_t4_id not in ('some_id')) and resource_path='-2147483644')) limit 100)", deleteBefore, deleteAfter).Once().Return(nil, nil)
	db.On("Exec", mock.Anything, "delete from t2 where ctid in (select ctid from t2 where ref_col=any(select id from t4 where ((($1::timestamptz is null or is_created_at < $1) and ($2::timestamptz is null or is_created_at > $2)) and some_t4_id not in ('some_id')) and resource_path='-2147483644') limit 100)", deleteBefore, deleteAfter).Once().Return(nil, nil)
	db.On("Exec", mock.Anything, "delete from t4 where ctid in (select ctid from t4 where ((($1::timestamptz is null or is_created_at < $1) and ($2::timestamptz is null or is_created_at > $2)) and some_t4_id not in ('some_id')) and resource_path='-2147483644' limit 100)", deleteBefore, deleteAfter).Once().Return(nil, nil)
	err := recursiveDelete(context.Background(), extraConf, "t4", db, depGraph, deleteBefore, true, 100, deleteAfter)
	assert.NoError(t, err)
	db.AssertExpectations(t)
}

func Test_setNullOnCircularFk(t *testing.T) {
	// t1 also has fk ref to t2, but ignored and not defined in the graph
	depGraph := map[string][]GraphNode{
		"t1": makeNodes("t1", []string{"t2"}),
		"t2": makeNodes("t2", nil),
	}

	db := &mock_database.Ext{}
	deleteBefore := database.Timestamptz(time.Now())
	deleteAfter := database.Timestamptz(time.Now())
	extraConf := map[string]Config{
		"t1": {
			IgnoreFks: []string{
				"t2_id",
			},
		},
		"t2": {
			SetNullOnCircularFk: map[string]string{
				"t1": "t2_id",
			},
		},
	}

	// circular fk t1->t2->t1, we setnull on t1 's referencing column, then we only have dependency from t2->t1
	db.On("Exec", mock.Anything, "update t1 set t2_id=null where ctid in (select ctid from t1 where ((($1::timestamptz is null or created_at < $1) and ($2::timestamptz is null or created_at > $2))) and resource_path='-2147483644' and t2_id is not null)", deleteBefore, deleteAfter).Once().Return(nil, nil)
	db.On("Exec", mock.Anything, "delete from t2 where ctid in (select ctid from t2 where ref_col=any(select id from t1 where ((($1::timestamptz is null or created_at < $1) and ($2::timestamptz is null or created_at > $2))) and resource_path='-2147483644'))", deleteBefore, deleteAfter).Once().Return(nil, nil)
	db.On("Exec", mock.Anything, "delete from t1 where ctid in (select ctid from t1 where ((($1::timestamptz is null or created_at < $1) and ($2::timestamptz is null or created_at > $2))) and resource_path='-2147483644')", deleteBefore, deleteAfter).Once().Return(nil, nil)

	err := recursiveDelete(context.Background(), extraConf, "t1", db, depGraph, deleteBefore, false, 0, deleteAfter)
	assert.NoError(t, err)
	db.AssertExpectations(t)
}

func Test_selfRefFkDelete(t *testing.T) {
	depGraph := map[string][]GraphNode{
		"t4": makeNodes("t4", []string{"t3"}),
		"t3": makeNodes("t3", []string{}),
	}

	mockDB := testutil.NewMockDB()
	db := mockDB.DB
	deleteBefore := database.Timestamptz(time.Now())
	deleteAfter := pgtype.Timestamptz{Status: pgtype.Null}
	extraConf := map[string]Config{
		"t4": {
			SelfRefFKs: []configurations.SelfRefFKs{
				{
					Referencing: "parent_id",
					Referenced:  "id",
				},
			},
		},
	}
	idsReturnedByCTE := database.TextArray([]string{"id1", "id2"})
	mockDB.MockRowScanFields(nil, []string{""}, []interface{}{&idsReturnedByCTE})
	// mock quering for children entity ids
	mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, deleteBefore.Time, deleteAfter.Time)
	nullts := pgtype.Timestamptz{Status: pgtype.Null}

	// db.On("Exec", mock.Anything, "delete from t1 where ctid in (select ctid from t1 where ref_col=any(select id from t3 where ref_col=any(select id from t4 where ($1::timestamptz is null or created_at < $1) and id=any($2))))", nullts, idsReturnedByCTE).Once().Return(nil, nil)
	db.On("Exec", mock.Anything, "delete from t3 where ctid in (select ctid from t3 where ref_col=any(select id from t4 where ((($1::timestamptz is null or created_at < $1) and ($2::timestamptz is null or created_at > $2)) and id=any($3)) and resource_path='-2147483644'))", nullts, nullts, idsReturnedByCTE).Once().Return(nil, nil)
	db.On("Exec", mock.Anything, "delete from t4 where ctid in (select ctid from t4 where ((($1::timestamptz is null or created_at < $1) and ($2::timestamptz is null or created_at > $2)) and id=any($3)) and resource_path='-2147483644')", nullts, nullts, idsReturnedByCTE).Once().Return(nil, nil)
	err := queryChildrenAndRecursiveDelete(context.Background(), extraConf, "t4", db, depGraph, deleteBefore.Time, false, 0, deleteAfter.Time)
	assert.NoError(t, err)
	db.AssertExpectations(t)
}

func Test_recursiveDelete(t *testing.T) {
	depGraph := map[string][]GraphNode{
		"t4": makeNodes("t4", []string{"t3", "t2"}),
		"t2": makeNodes("t2", []string{"t1"}),
		"t3": makeNodes("t3", []string{"t1"}),
	}
	db := &mock_database.Ext{}
	deleteBefore := database.Timestamptz(time.Now())
	deleteAfter := pgtype.Timestamptz{Status: pgtype.Null}
	extraConf := map[string]Config{
		"t4": {
			CreatedAtColName: "is_created_at",
			ExtraCond:        "and some_t4_id not in ('some_id')",
		},
		"t3": {
			CreatedAtColName: "is_created_at",
			ExtraCond:        "and some_t3_id not in ('some_id')",
		},
	}

	db.On("Exec", mock.Anything, "delete from t1 where ctid in (select ctid from t1 where ref_col=any(select id from t3 where ref_col=any(select id from t4 where ((($1::timestamptz is null or is_created_at < $1) and ($2::timestamptz is null or is_created_at > $2)) and some_t4_id not in ('some_id')) and resource_path='-2147483644') and some_t3_id not in ('some_id')))", deleteBefore, deleteAfter).Once().Return(nil, nil)
	db.On("Exec", mock.Anything, "delete from t3 where ctid in (select ctid from t3 where ref_col=any(select id from t4 where ((($1::timestamptz is null or is_created_at < $1) and ($2::timestamptz is null or is_created_at > $2)) and some_t4_id not in ('some_id')) and resource_path='-2147483644') and some_t3_id not in ('some_id'))", deleteBefore, deleteAfter).Once().Return(nil, nil)
	db.On("Exec", mock.Anything, "delete from t1 where ctid in (select ctid from t1 where ref_col=any(select id from t2 where ref_col=any(select id from t4 where ((($1::timestamptz is null or is_created_at < $1) and ($2::timestamptz is null or is_created_at > $2)) and some_t4_id not in ('some_id')) and resource_path='-2147483644')))", deleteBefore, deleteAfter).Once().Return(nil, nil)
	db.On("Exec", mock.Anything, "delete from t2 where ctid in (select ctid from t2 where ref_col=any(select id from t4 where ((($1::timestamptz is null or is_created_at < $1) and ($2::timestamptz is null or is_created_at > $2)) and some_t4_id not in ('some_id')) and resource_path='-2147483644'))", deleteBefore, deleteAfter).Once().Return(nil, nil)
	db.On("Exec", mock.Anything, "delete from t4 where ctid in (select ctid from t4 where ((($1::timestamptz is null or is_created_at < $1) and ($2::timestamptz is null or is_created_at > $2)) and some_t4_id not in ('some_id')) and resource_path='-2147483644')", deleteBefore, deleteAfter).Once().Return(nil, nil)
	err := recursiveDelete(context.Background(), extraConf, "t4", db, depGraph, deleteBefore, false, 0, deleteAfter)
	assert.NoError(t, err)
	db.AssertExpectations(t)
}

func Test_parseTime(t *testing.T) {
	parsed, err := parseTime("3 days ago")
	assert.NoError(t, err)
	assert.Equal(t, 3, time.Now().Day()-parsed.Day())
	parsed, err = parseTime("1 month ago")

	assert.NoError(t, err)
	assert.Equal(t, time.Month(1), time.Now().Month()-parsed.Month())
}
