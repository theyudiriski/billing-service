package billing

import (
	"flag"
	"log/slog"
	"os"
	"time"

	"github.com/lmittmann/tint"
)

type Logger = *slog.Logger

func NewLogger() Logger {
	handler := getLocalHandler()

	return slog.New(handler)
}

func getLocalHandler() slog.Handler {
	return tint.NewHandler(os.Stdout, &tint.Options{
		Level:      slog.LevelDebug,
		TimeFormat: time.TimeOnly,
		NoColor:    flag.Lookup("test.v") != nil,
	})
}
