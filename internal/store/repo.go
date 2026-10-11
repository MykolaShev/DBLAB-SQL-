package store

import (
	"context"
	"database/sql"
)

type ProjectRepository interface {
	GetActiveProjects(ctx context.Context) ([]ProjectView, error)
	UpdateStatus(ctx context.Context, projectID int, status string, userID int) error
}

type ProposalRepository interface {
	CreateProposal(ctx context.Context, projectID, freelancerID int, coverLetter string, bid float64) error
	AcceptProposal(ctx context.Context, proposalID, userID int, startDate, endDate string) error
}

type projectRepo struct {
	tx *sql.Tx
}

func (r *projectRepo) GetActiveProjects(ctx context.Context) ([]ProjectView, error) {
	query := `SELECT project_id, title, budget, category_name, client, created_at FROM vw_active_projects;`
	rows, err := r.tx.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []ProjectView
	for rows.Next() {
		var p ProjectView
		if err := rows.Scan(&p.ProjectID, &p.Title, &p.Budget, &p.CategoryName, &p.Client, &p.CreatedAt); err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}
	return projects, nil
}

func (r *projectRepo) UpdateStatus(ctx context.Context, projectID int, status string, userID int) error {
	query := `CALL sp_update_project_status($1, $2, $3);`
	_, err := r.tx.ExecContext(ctx, query, projectID, status, userID)
	return err
}

type proposalRepo struct {
	tx *sql.Tx
}

func (r *proposalRepo) CreateProposal(ctx context.Context, projectID, freelancerID int, coverLetter string, bid float64) error {
	query := `CALL sp_create_proposal($1, $2, $3, $4);`
	_, err := r.tx.ExecContext(ctx, query, projectID, freelancerID, coverLetter, bid)
	return err
}

func (r *proposalRepo) AcceptProposal(ctx context.Context, proposalID, userID int, startDate, endDate string) error {
	query := `CALL sp_accept_proposal($1, $2, $3, $4);`
	_, err := r.tx.ExecContext(ctx, query, proposalID, userID, startDate, endDate)
	return err
}