package main

import (
	"context"
	"fmt"

	"github.com/jurwaCST/Jurwa/services/identity/application"
)

// seedFromEnv creates the demo accounts when SEED_DEMO_USERS=true, with the
// password from SEED_DEMO_PASSWORD. There is no default password. For the
// demo only: never set SEED_DEMO_USERS in production.
func seedFromEnv(ctx context.Context, app *application.Service, getenv func(string) string) (int, error) {
	if getenv("SEED_DEMO_USERS") != "true" {
		return 0, nil
	}
	created, err := app.SeedDemoUsers(ctx, getenv("SEED_DEMO_PASSWORD"))
	if err != nil {
		// The error names the variable, never its value.
		return created, fmt.Errorf("SEED_DEMO_PASSWORD: %w", err)
	}
	return created, nil
}
