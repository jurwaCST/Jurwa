package main

import (
	"context"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/jurwaCST/Jurwa/services/identity/adapters/out/memory"
	"github.com/jurwaCST/Jurwa/services/identity/adapters/out/postgres"
	"github.com/jurwaCST/Jurwa/services/identity/ports"
)

// openRepo returns the PostgreSQL repository when url is set (creating the
// tables if needed), otherwise the in-memory one. A bad url is an error, so
// the service never falls back to memory by mistake.
func openRepo(url string) (ports.UserRepository, func(), error) {
	if url == "" {
		slog.Info("storage: in memory (DATABASE_URL not set)")
		return memory.NewUsers(), func() {}, nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, nil, err
	}
	repo := postgres.New(pool)
	if err := repo.Migrate(ctx); err != nil {
		pool.Close()
		return nil, nil, err
	}
	slog.Info("storage: postgres")
	return repo, pool.Close, nil
}
