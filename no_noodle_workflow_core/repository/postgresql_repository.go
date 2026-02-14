package repository

import "database/sql"

type PostgreSQLNoNoodleWorkflow struct {
	db *sql.DB
}

func NewPostgreSQLNoNoodleWorkflow(db *sql.DB) *PostgreSQLNoNoodleWorkflow {
	return &PostgreSQLNoNoodleWorkflow{db: db}
}

func (p *PostgreSQLNoNoodleWorkflow) GetDB() *sql.DB {
	return p.db
}
