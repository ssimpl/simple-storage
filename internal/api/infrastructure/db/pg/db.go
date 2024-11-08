package pg

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/ssimpl/simple-storage/internal/api/infrastructure/db/pg/entity"
	"github.com/ssimpl/simple-storage/internal/api/model"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

const pingTimeout = 5 * time.Second

type DB struct {
	*bun.DB
}

func NewDB(cfg Config) (*DB, error) {
	cfg.SetDefaults()

	sqlDB := sql.OpenDB(pgdriver.NewConnector(
		pgdriver.WithAddr(cfg.Addr),
		pgdriver.WithDatabase(cfg.Database),
		pgdriver.WithUser(cfg.User),
		pgdriver.WithPassword(cfg.Password),
		pgdriver.WithApplicationName(cfg.AppName),
		pgdriver.WithTimeout(cfg.Timeout),
		pgdriver.WithInsecure(true),
	))

	db := bun.NewDB(sqlDB, pgdialect.New(), bun.WithDiscardUnknownColumns())
	if cfg.SQLDebug {
		db.AddQueryHook(newLogHook())
	}

	ctx, cancel := context.WithTimeout(context.Background(), pingTimeout)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("unable to connect to PG database: %w", err)
	}

	slog.Info(
		"connect to PG database is successful",
		"pg_addr", cfg.Addr,
		"pg_user", cfg.User,
		"pg_database", cfg.Database,
	)

	return &DB{DB: db}, nil
}

func (db *DB) SaveObjectMeta(ctx context.Context, meta model.ObjectMeta) error {
	e, err := entity.ObjectMetaFromModel(meta)
	if err != nil {
		return fmt.Errorf("convert object meta to db entity: %w", err)
	}

	_, err = db.NewInsert().
		Model(&e).
		On("CONFLICT (name) DO UPDATE").
		Set("fragments = EXCLUDED.fragments").
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("insert object metadata: %w: %w", err, model.ErrDBMalfunctioning)
	}

	return nil
}

func (db *DB) GetObjectMeta(ctx context.Context, objectName string) (model.ObjectMeta, error) {
	var e entity.ObjectMeta

	err := db.NewSelect().
		Model(&e).
		Where("name = ?", objectName).
		Scan(ctx)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.ObjectMeta{}, model.ErrObjectNotFound
		}
		return model.ObjectMeta{}, fmt.Errorf(
			"select object metadata: %w: %w", err, model.ErrDBMalfunctioning,
		)
	}

	meta, err := e.ToModel()
	if err != nil {
		return model.ObjectMeta{}, fmt.Errorf("convert object meta to model: %w", err)
	}

	return meta, nil
}

func (db *DB) GetServers(ctx context.Context) ([]model.Server, error) {
	var entities []entity.Server

	err := db.NewSelect().
		Model(&entities).
		Scan(ctx)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []model.Server{}, nil
		}
		return nil, fmt.Errorf(
			"select all servers: %w: %w", err, model.ErrDBMalfunctioning,
		)
	}

	servers := make([]model.Server, 0, len(entities))
	for _, e := range entities {
		servers = append(servers, e.ToModel())
	}

	return servers, nil
}
