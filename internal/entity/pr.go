package entity

import "time"

type PRStatus string

const (
	StatusOpen   PRStatus = "OPEN"
	StatusMerged PRStatus = "MERGED"
)

type PullRequest struct {
	ID               string     `json:"pull_request_id"`
	Name             string     `json:"pull_request_name"`
	AuthorID         string     `json:"author_id"`
	Status           PRStatus   `json:"status"`
	AssignedReviewers []string  `json:"assigned_reviewers"`
	CreatedAt        time.Time  `json:"createdAt"`
	MergedAt         *time.Time `json:"mergedAt,omitempty"`
}

// Short version for user reviews list
type PullRequestShort struct {
	ID       string   `json:"pull_request_id"`
	Name     string   `json:"pull_request_name"`
	AuthorID string   `json:"author_id"`
	Status   PRStatus `json:"status"`
}