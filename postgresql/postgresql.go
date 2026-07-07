package chi_postgresql

import (
	"context"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

type PostgreSQLClientConfig struct {
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
	PingTimeout     time.Duration
}

type PostgreSQL struct {
	client *sqlx.DB
}

func NewPostgreSQL(cfg PostgreSQLClientConfig) (*PostgreSQL, error) {
	client, err := sqlx.Connect("pgx", cfg.DSN)
	if err != nil {
		return nil, err
	}

	client.SetMaxOpenConns(cfg.MaxOpenConns)
	client.SetMaxIdleConns(cfg.MaxIdleConns)
	client.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	client.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	return &PostgreSQL{
		client: client,
	}, nil
}

func (p *PostgreSQL) Check(ctx context.Context) error {
	return p.client.PingContext(ctx)
}

func (p *PostgreSQL) Cleanup() {
	p.client.Close()
}

func (p *PostgreSQL) GetClient() any { return p.client }
