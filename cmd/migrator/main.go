package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/ssimpl/simple-storage/migrations"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/migrate"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg, err := newConfig()
	if err != nil {
		return err
	}

	sqlDB := sql.OpenDB(pgdriver.NewConnector(
		pgdriver.WithAddr(cfg.PG.Addr),
		pgdriver.WithDatabase(cfg.PG.Database),
		pgdriver.WithUser(cfg.PG.User),
		pgdriver.WithPassword(cfg.PG.Password),
		pgdriver.WithInsecure(true),
	))

	db := bun.NewDB(sqlDB, pgdialect.New(), bun.WithDiscardUnknownColumns())

	m := migrate.NewMigrations()
	if err := m.Discover(migrations.SQLMigrations); err != nil {
		return err
	}

	ctx := context.Background()

	migrator := migrate.NewMigrator(db, m)
	if err := migrator.Init(ctx); err != nil {
		return err
	}

	if err := migrator.Lock(ctx); err != nil {
		return err
	}
	defer func() {
		_ = migrator.Unlock(ctx)
	}()

	group, err := migrator.Migrate(ctx)
	if err != nil {
		return err
	}
	if group.IsZero() {
		fmt.Printf("there are no new migrations to run (database is up to date)\n")
		return nil
	}
	fmt.Printf("migrated to %s\n", group)

	return nil
}
