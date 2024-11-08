package pg

import (
	"bytes"
	"context"
	"log/slog"

	"github.com/uptrace/bun"
)

type logHook struct {
}

func newLogHook() *logHook {
	return &logHook{}
}

func (h *logHook) BeforeQuery(ctx context.Context, event *bun.QueryEvent) context.Context {
	slog.Info("sql debug", "query", string(bytes.Trim([]byte(event.Query), "\r\n\t")))
	return ctx
}

func (h *logHook) AfterQuery(_ context.Context, _ *bun.QueryEvent) {
}
