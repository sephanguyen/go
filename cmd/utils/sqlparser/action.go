package sqlparser

import "strings"

type ActionType int

const (
	UnknownAction ActionType = 0
	CreateAction  ActionType = 1
	UpdateAction  ActionType = 2
	DeleteAction  ActionType = 3
	AlterAction   ActionType = 4
	DropAction    ActionType = 5
)

func ParseAction(val string) ActionType {
	switch strings.ToLower(val) {
	case "create":
		return CreateAction
	case "update":
		return UpdateAction
	case "delete":
		return DeleteAction
	case "alter":
		return AlterAction
	case "drop":
		return DropAction
	}
	return UnknownAction
}

func (p ActionType) String() string {
	switch p {
	case CreateAction:
		return "create"
	case UpdateAction:
		return "update"
	case DeleteAction:
		return "delete"
	case AlterAction:
		return "alter"
	case DropAction:
		return "drop"
	}
	return "unknown"
}

type EntityType int

const (
	UnknownEntity   EntityType = 0
	DatabaseEntity  EntityType = 1
	TableEntity     EntityType = 2
	ProcedureEntity EntityType = 3
	FunctionEntity  EntityType = 4
	IndexEntity     EntityType = 5
)

func ParseEntity(val string) EntityType {
	switch strings.ToLower(val) {
	case "database":
	case "schema":
		return DatabaseEntity
	case "table":
		return TableEntity
	case "procedure":
		return ProcedureEntity
	case "function":
		return FunctionEntity
	case "index":
		return IndexEntity
	}
	return UnknownEntity
}

func (p EntityType) String() string {
	switch p {
	case DatabaseEntity:
		return "database"
	case TableEntity:
		return "table"
	case ProcedureEntity:
		return "procedure"
	case FunctionEntity:
		return "function"
	case IndexEntity:
		return "index"
	}
	return "unknown"
}
