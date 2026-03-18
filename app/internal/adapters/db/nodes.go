package db

import (
	"context"
	"fmt"

	"github.com/andrii-reshko/3c-informational-technical-support-lab/app/internal/domain"
	"github.com/jmoiron/sqlx"
)

type nodeRepo struct {
	db    *sqlx.DB
	table string
}

func NewNodeRepository(db *sqlx.DB) domain.NodeRepository {
	return &nodeRepo{
		db:    db,
		table: "nodes",
	}
}

func (r *nodeRepo) UpsertNode(ctx context.Context, n domain.Node) error {
	query := `INSERT INTO %s (id, name, cpu_cores, ram_gb, updated_at)
            VALUES (:id, :name, :cpu_cores, :ram_gb, CURRENT_TIMESTAMP)
            ON CONFLICT(id) DO UPDATE SET 
            cpu_cores=excluded.cpu_cores, ram_gb=excluded.ram_gb, updated_at=CURRENT_TIMESTAMP`
	query = fmt.Sprintf(query, r.table)

	_, err := r.db.NamedExecContext(ctx, query, n)
	return err
}

func (r *nodeRepo) GetAllNodes(ctx context.Context) ([]domain.Node, error) {
	var nodes []domain.Node
	query := `SELECT id, name, cpu_cores, ram_gb, updated_at FROM %s`
	query = fmt.Sprintf(query, r.table)

	err := r.db.SelectContext(ctx, &nodes, query)
	return nodes, err
}

func (r *nodeRepo) GetNodeByID(ctx context.Context, id string) (*domain.Node, error) {
	var node domain.Node
	query := `SELECT id, name, cpu_cores, ram_gb, updated_at FROM %s WHERE id = $1`
	query = fmt.Sprintf(query, r.table)

	err := r.db.GetContext(ctx, &node, query, id)
	if err != nil {
		return nil, err
	}
	return &node, nil
}
