package dbutil

import (
	"context"
	"database/sql"
	"time"

	"golang.org/x/sync/errgroup"
)

func (d DbConn) iterateReplicas(eg *errgroup.Group, f func(replica *sql.DB) error) {
	for _, replica := range d.replicas {
		replica := replica // make sure concurrent safe [ref](https://go.dev/blog/loopvar-preview)
		eg.Go(func() error {
			return f(replica)
		})
	}
}

func (d DbConn) Begin() (*sql.Tx, error) {
	return d.BeginTx(context.Background(), nil)
}

func (d DbConn) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return d.Primary().BeginTx(ctx, opts)
}

func (d DbConn) Close() error {
	eg := errgroup.Group{}
	eg.Go(func() error {
		return d.primary.Close()
	})

	d.iterateReplicas(&eg, func(replica *sql.DB) error {
		return replica.Close()
	})

	return eg.Wait()
}

func (d DbConn) PingContext(ctx context.Context) error {
	eg := errgroup.Group{}
	eg.Go(func() error {
		return d.primary.PingContext(ctx)
	})

	d.iterateReplicas(&eg, func(replica *sql.DB) error {
		return replica.PingContext(ctx)
	})

	return eg.Wait()
}

func (d DbConn) SetConnMaxLifetime(duration time.Duration) {
	d.primary.SetConnMaxLifetime(duration)
	for _, replica := range d.replicas {
		replica.SetConnMaxLifetime(duration)
	}
}

func (d DbConn) SetMaxIdleConns(n int) {
	d.primary.SetMaxIdleConns(n)
	for _, replica := range d.replicas {
		replica.SetMaxIdleConns(n)
	}
}

func (d DbConn) SetMaxOpenConns(n int) {
	d.primary.SetMaxOpenConns(n)
	for _, replica := range d.replicas {
		replica.SetMaxOpenConns(n)
	}
}

func (d DbConn) Stats() sql.DBStats {
	return d.primary.Stats()
}
