package entities

import (
	"time"
)

type History struct {
	ID               int64     `json:"id"`
	BranchName       string    `json:"branch_name"`
	Coverage         float32   `json:"coverage"`
	Time             time.Time `json:"time"`
	Status           string    `json:"status"`
	Repository       string    `json:"repository"`
	TargetBranchName string    `json:"target_branch_name"`
	Integration      bool      `json:"integration"`
}

func (e *History) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"id", "branch_name", "coverage", "time", "status", "repository", "target_branch_name", "integration"}
	values = []interface{}{&e.ID, &e.BranchName, &e.Coverage, &e.Time, &e.Status, &e.Repository, &e.TargetBranchName, &e.Integration}
	return
}

func (e *History) TableName() string {
	return "history"
}

type TargetCoverage struct {
	ID          int64     `json:"id"`
	BranchName  string    `json:"branch_name"`
	Coverage    float32   `json:"coverage"`
	UpdateAt    time.Time `json:"update_at"`
	Repository  string    `json:"repository"`
	Key         string    `json:"key"`
	Integration bool      `json:"integration"`
}

func (e *TargetCoverage) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"id", "branch_name", "coverage", "update_at", "repository", "secret_key", "integration"}
	values = []interface{}{&e.ID, &e.BranchName, &e.Coverage, &e.UpdateAt, &e.Repository, &e.Key, &e.Integration}
	return
}

func (e *TargetCoverage) TableName() string {
	return "target_coverage"
}
