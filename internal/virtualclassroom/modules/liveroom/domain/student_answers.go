package domain

import "time"

type StudentAnswers struct {
	UserID    string    `json:"user_id"`
	Answers   []string  `json:"answers"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type StudentAnswersList []*StudentAnswers
