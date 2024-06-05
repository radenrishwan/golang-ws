package gows

import (
	"log/slog"
	"os"
)

var Logger = slog.New(slog.NewJSONHandler(os.Stderr, nil))
