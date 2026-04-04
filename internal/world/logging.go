package world

import "log/slog"

func logger() *slog.Logger {
	return slog.Default().WithGroup("world")
}
