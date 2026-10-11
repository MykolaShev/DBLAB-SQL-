package store

import "time"

// ProjectView maps to the vw_active_projects database view.
type ProjectView struct {
	ProjectID    int
	Title        string
	Budget       float64
	CategoryName string
	Client       string
	CreatedAt    time.Time
}

// ProposalView maps to the vw_project_proposals database view.
type ProposalView struct {
	ProposalID     int
	ProjectTitle   string
	BidAmount      float64
	Status         string
	FreelancerName string
}