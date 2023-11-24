package log

import (
	"log/slog"
	"os"
)

var logger = slog.New(
	slog.NewTextHandler(os.Stdout,
		&slog.HandlerOptions{Level: slog.LevelDebug},
	),
)

func Logger() *slog.Logger {
	return logger
}
