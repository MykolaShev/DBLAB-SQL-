package store

import (
	"context"
	"database/sql"
)

// UnitOfWork manages atomic operations across multiple repositories.
type UnitOfWork interface {
	Commit() error
	Rollback() error
	Projects() ProjectRepository
	Proposals() ProposalRepository
}

type uow struct {
	tx        *sql.Tx
	projects  ProjectRepository
	proposals ProposalRepository
}

func NewUnitOfWork(ctx context.Context, db *sql.DB) (UnitOfWork, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	return &uow{
		tx:        tx,
		projects:  &projectRepo{tx: tx},
		proposals: &proposalRepo{tx: tx},
	}, nil
}

func (u *uow) Commit() error {
	return u.tx.Commit()
}

func (u *uow) Rollback() error {
	return u.tx.Rollback()
}

func (u *uow) Projects() ProjectRepository {
	return u.projects
}

func (u *uow) Proposals() ProposalRepository {
	return u.proposals
}