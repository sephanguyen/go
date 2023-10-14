package kafkaconnect

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/go-kafka/connect"
)

type tableInfoList []tableInfo

func (ts tableInfoList) GetTableList() []string {
	res := make([]string, 0, len(ts))
	for _, t := range ts {
		res = append(res, t.GetTableName())
	}
	return res
}

type tableInfo struct {
	db     string
	schema string
	table  string
}

func (t *tableInfo) GetTableName() string {
	return fmt.Sprintf("%s.%s", t.schema, t.table)
}

func (t *tableInfo) Equal(other tableInfo) bool {
	return t.db == other.db &&
		t.schema == other.schema &&
		t.table == other.table
}

func extractSourceIDFromConnector(c *connect.Connector) (string, error) {
	sourceName, ok := c.Config["database.server.name"]
	if !ok {
		return "", fmt.Errorf("cannot extract field database.server.name in config connector %s", c.Name)
	}

	return sourceName, nil
}

func extractReplicationSlotNameFromConnector(c *connect.Connector) (string, error) {
	repName, ok := c.Config["slot.name"]
	if !ok {
		return "", fmt.Errorf("cannot extract field slot.name in config connector %s", c.Name)
	}

	return repName, nil
}

func extractNewTableFromConnector(newConfig, oldConfig *connect.Connector) tableInfoList {
	res := make([]tableInfo, 0)

	getTableFromField := func(config map[string]string) []tableInfo {
		tablesStr, ok := config["table.include.list"]
		if !ok {
			return nil
		}
		dbName, ok := config["database.server.name"]
		if !ok {
			return nil
		}

		tables := strings.Split(tablesStr, ",")
		res := make([]tableInfo, 0, len(tables))
		for i := range tables {
			tables[i] = strings.TrimSpace(tables[i])
			res = append(res, tableInfo{
				db:     dbName,
				schema: strings.Split(tables[i], ".")[0],
				table:  strings.Split(tables[i], ".")[1],
			})
		}

		return res
	}

	newTable := getTableFromField(newConfig.Config)
	oldTable := getTableFromField(oldConfig.Config)

	mp := make(map[tableInfo]bool)
	for _, t := range oldTable {
		mp[t] = true
	}
	for _, t := range newTable {
		if _, ok := mp[t]; !ok {
			res = append(res, t)
		}
	}
	return res
}

func extractTableFromTopic(tableTopics []string) tableInfoList {
	tables := make([]tableInfo, 0, len(tableTopics))
	// stag.manabie.bob.public.users
	rgx := regexp.MustCompile(`([^.]+).([^.]+).([^.]+).([^.]+).([^.]+)`)
	for _, t := range tableTopics {
		res := rgx.FindStringSubmatch(t)
		if len(res) >= 6 {
			tables = append(tables, tableInfo{
				db:     res[3],
				schema: res[4],
				table:  res[5],
			})
		}
	}

	return tables
}

func isChangeFieldFunc(newConf, oldConf map[string]string) bool {
	for k := range newConf {
		if ignorableField(k) {
			continue
		}
		n := newConf[k]
		o, ok := oldConf[k]
		if !ok {
			continue
		}
		if n != o {
			return true
		}
	}
	return false
}

func isAddFieldFunc(newConf, oldConf map[string]string) bool {
	for k := range newConf {
		if ignorableField(k) {
			continue
		}
		_, ok := oldConf[k]
		if !ok {
			return true
		}
	}
	return false
}

func isRemoveFieldFunc(newConf, oldConf map[string]string) bool {
	for k := range oldConf {
		if ignorableField(k) {
			continue
		}
		_, ok := newConf[k]
		if !ok {
			return true
		}
	}
	return false
}

func ignorableField(f string) bool {
	whiteList := map[string]bool{
		"name":              true,
		"table.name.format": true,
		"fields.whitelist":  true,
		"pk.fields":         true,
	}

	_, ok := whiteList[f]

	return ok
}

func isChangedInColumnFields(oldListRaw, newListRaw string) bool {
	oldList := strings.Split(oldListRaw, ",")
	newList := strings.Split(newListRaw, ",")

	if len(oldList) != len(newList) {
		return true
	}

	n := len(newList)
	sort.Strings(oldList)
	sort.Strings(newList)
	for i := 0; i < n; i++ {
		if strings.TrimSpace(oldList[i]) != strings.TrimSpace(newList[i]) {
			return true
		}
	}
	return false
}

func anyChangeInConfig(oldConnector, newConnector *connect.Connector) bool {
	oldConf := oldConnector.Config
	newConf := newConnector.Config

	return isChangeFieldFunc(newConf, oldConf) ||
		isAddFieldFunc(newConf, oldConf) ||
		isRemoveFieldFunc(newConf, oldConf) ||
		isChangedInColumnFields(oldConnector.Config["fields.whitelist"], newConnector.Config["fields.whitelist"]) ||
		isChangedInColumnFields(oldConnector.Config["pk.fields"], newConnector.Config["pk.fields"])
}
