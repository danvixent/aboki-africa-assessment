package postgres

import (
	"context"
	"fmt"
	"github.com/jackc/pgx"
	"github.com/pkg/errors"
	"net"
	"time"

	"github.com/danvixent/buycoin-challenge2/config"
	pool "github.com/jackc/pgx/v4/pgxpool"
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

// NewFromCfgMap Returns a new database initialized with credentials from config
func NewFromCfgMap(ctx context.Context, config *config.PostgresConfig) (*Client, error) {
	const format = "postgres://%s:%s@%s:%d/%s?sslmode=disable&pool_max_conns=%d"
	uri := fmt.Sprintf(format, config.Username, config.Password, config.Host, config.Port, config.Database, config.MaxConn)

	cfg, err := pool.ParseConfig(uri)
	if err != nil {
		return nil, err
	}

	cfg.ConnConfig.DialFunc = func(ctx context.Context, host string, addr string) (net.Conn, error) {
		return net.Dial(host, addr)
	}

	cfg.ConnConfig.PreferSimpleProtocol = true
	cfg.ConnConfig.ConnectTimeout = time.Minute

	pool, err := pool.ConnectConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}

	w := &Client{pool: pool}
	return w, nil
}
