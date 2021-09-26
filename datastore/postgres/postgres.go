package postgres

import (
	"context"
	"fmt"
	"github.com/jackc/pgconn"
	"strings"
	"time"

	app "github.com/danvixent/aboki-africa-assessment"
	"github.com/danvixent/aboki-africa-assessment/config"
	"github.com/jackc/pgx/v4"
	pool "github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type Client struct {
	pool *pool.Pool
}

func (c *Client) BeginTx() (pgx.Tx, error) {
	tx, err := c.pool.BeginTx(context.Background(), defaultOptions)
	if err != nil {
		return nil, errors.Wrap(err, "failed to begin new transaction")
	}
	return tx, nil
}

// GetTx extracts a pgx.Tx from ctx
func (c *Client) GetTx(ctx context.Context) (Tx, error) {
	tx := ctx.Value(app.TxContextKey)
	if tx != nil {
		return tx.(Tx), nil
	}
	return c, nil
}

// New Returns a new database initialized with credentials from config
func New(ctx context.Context, config *config.PostgresConfig) *Client {
	const format = "postgres://%s:%s@%s:%s/%s?sslmode=disable&pool_max_conns=%d"
	uri := fmt.Sprintf(format, config.Username, config.Password, config.Host, config.Port, config.Database, config.MaxConn)

	cfg, err := pool.ParseConfig(uri)
	if err != nil {
		log.Panicf("failed to parse pgx config: %v", err)
	}

	cfg.ConnConfig.ConnectTimeout = time.Minute

	pool, err := pool.ConnectConfig(ctx, cfg)
	if err != nil {
		log.Panicf("pgx pool failed to connect: %v", err)
	}

	return &Client{pool: pool}
}

// Query executues a query that typically returns more than one row
func (c *Client) Query(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error) {
	rs, err := c.pool.Query(ctx, query, args...)
	return rs, err
}

// QueryRow excutes a query that typically returns one row
func (c *Client) QueryRow(ctx context.Context, query string, args ...interface{}) pgx.Row {
	row := c.pool.QueryRow(ctx, query, args...)
	return row
}

// Exec executes a query that doesn't return rows
func (c *Client) Exec(ctx context.Context, query string, args ...interface{}) (pgconn.CommandTag, error) {
	_, err := c.pool.Exec(ctx, query, args...)
	return nil, err
}

func (c *Client) Commit(ctx context.Context) error {
	return nil
}

func (c *Client) Rollback(ctx context.Context) error {
	return nil
}

func IsDuplicateError(err error) bool {
	return strings.Contains(err.Error(), "duplicate key value violates unique constraint")
}

// Tx represents a database transaction
type Tx interface {
	// Query executes a query that typically returns more than one row
	Query(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error)

	// QueryRow executes a query that typically returns one row
	QueryRow(ctx context.Context, query string, args ...interface{}) pgx.Row

	// Exec executes a query that doesn't return rows
	Exec(ctx context.Context, query string, args ...interface{}) (pgconn.CommandTag, error)

	// Commit commits the transaction
	Commit(ctx context.Context) error

	// Rollback rolls back the transaction
	Rollback(ctx context.Context) error
}
