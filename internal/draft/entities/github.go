package entities

import (
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

// TODO: use pgtype, because some of them may be null, which will panic calling row.Scan
type GithubPRData struct {
	ID                  int64       `json:"id"`
	Number              int         `json:"number"`
	BranchName          string      `json:"branch_name"`
	Create              time.Time   `json:"create_at"`
	Close               time.Time   `json:"close_at"`
	NumOfComments       int16       `json:"number_comments"`
	TotalToFirstComment float32     `json:"time_to_first_comment"`
	TotalTimeConsuming  float32     `json:"total_time_consuming"`
	IsMerged            bool        `json:"is_merged"`
	MergeStatus         pgtype.Text `json:"merge_status"`
}

func (g *GithubPRData) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"id", "branch_name", "number", "create_at", "close_at", "number_comments", "time_to_first_comment", "total_time_consuming", "is_merged", "merge_status"}
	values = []interface{}{&g.ID, &g.BranchName, &g.Number, &g.Create, &g.Close, &g.NumOfComments, &g.TotalToFirstComment, &g.TotalTimeConsuming, &g.IsMerged, &g.MergeStatus}
	return
}

func (g *GithubPRData) TableName() string {
	return "github_pr_statistic"
}

////Thi-code

// this is for implement Entities interface
type GithubPRDatas []*GithubPRData

// Add append new QuizSet
func (u *GithubPRDatas) Add() database.Entity {
	e := &GithubPRData{}
	*u = append(*u, e)

	return e
}

type GithubPRStatus struct {
	ID           int64  `json:"id"`
	Repo         string `json:"repo"`
	PRNumber     int    `json:"prnumber"`
	BaseBranch   string `json:"basebr"`
	TargetBranch string `json:"targetbr"`
	Status       string `json:"status"`
	IsMerged     bool   `json:"ismerged"`
}

func (g *GithubPRStatus) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"id", "repo", "prnumber", "basebr", "targetbr", "status", "ismerged"}
	values = []interface{}{&g.ID, &g.Repo, &g.PRNumber, &g.BaseBranch, &g.TargetBranch, &g.Status, &g.IsMerged}
	return
}

func (g *GithubPRStatus) TableName() string {
	return "github_pr"
}

// this is for implement Entities interface
type GithubPRStatuses []*GithubPRStatus

// Add append new QuizSet
func (u *GithubPRStatuses) Add() database.Entity {
	e := &GithubPRStatus{}
	*u = append(*u, e)

	return e
}
