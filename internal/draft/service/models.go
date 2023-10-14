package service

import "time"

type User struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Username string `json:"username"`
}

type EventPayload struct {
	Action     string `json:"action"`
	Ref        string `json:"ref"`
	Before     string `json:"before"`
	After      string `json:"after"`
	Repository struct {
		ID            int    `json:"id"`
		Name          string `json:"name"`
		FullName      string `json:"full_name"`
		DefaultBranch string `json:"default_branch"`
		Private       bool   `json:"private"`
		Owner         struct {
			Name  string `json:"name"`
			Email string `json:"email"`
			Login string `json:"login"`
			ID    int    `json:"id"`
		} `json:"owner"`
	} `json:"repository"`
	Pusher struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	} `json:"pusher"`
	Commits []struct {
		ID        string        `json:"id"`
		TreeID    string        `json:"tree_id"`
		Message   string        `json:"message"`
		Timestamp time.Time     `json:"timestamp"`
		Author    User          `json:"author"`
		Committer User          `json:"committer"`
		Added     []string      `json:"added"`
		Removed   []interface{} `json:"removed"`
		Modified  []string      `json:"modified"`
	} `json:"commits"`
	HeadCommit struct {
		ID        string        `json:"id"`
		TreeID    string        `json:"tree_id"`
		Message   string        `json:"message"`
		Timestamp time.Time     `json:"timestamp"`
		Author    User          `json:"author"`
		Committer User          `json:"committer"`
		Added     []string      `json:"added"`
		Removed   []interface{} `json:"removed"`
		Modified  []string      `json:"modified"`
	} `json:"head_commit"`

	PullRequest struct {
		Number    int    `json:"number"`
		State     string `json:"state"`
		Title     string `json:"title"`
		CreatedAt string `json:"created_at"`
		Body      string `json:"body"`
		Merged    bool   `json:"merged"`
		Head      struct {
			Ref string `json:"ref"`
			Sha string `json:"sha"`
		} `json:"head"`
		Base struct {
			Ref string `json:"ref"`
		} `json:"base"`
	} `json:"pull_request"`

	Review struct {
		ID     int    `json:"id"`
		NodeID string `json:"node_id"`
		User   struct {
			Login string `json:"login"`
		} `json:"user"`
		State    string `json:"state"`
		CommitID string `json:"commit_id"`
	} `json:"review"`

	Issue struct {
		Number int `json:"number"`
	} `json:"issue"`
	Comment struct {
		Body string `json:"body"`
	} `json:"comment"`
	Label struct {
		Name string `json:"name"`
	} `json:"label"`
}

type SetStatusPayload struct {
	Repo       string `json:"repo"`
	Owner      string `json:"owner"`
	BlockMerge bool   `json:"block_merge"`
}

type ExtraCond struct {
	Table     string `json:"table"`
	Condition string `json:"condition"`
}

type CleanDataPayload struct {
	Tables    string      `json:"tables"`
	Service   string      `json:"service"`
	SchoolID  string      `json:"school_id"`
	PerBatch  int         `json:"per_batch"`
	BeforeAt  string      `json:"before_at"`
	AfterAt   string      `json:"after_at"`
	ExtraCond []ExtraCond `json:"extra_cond"`
}
