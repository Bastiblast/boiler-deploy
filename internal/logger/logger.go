package logger

import (
"io"
"os"
"strings"
"time"

"github.com/rs/zerolog"
"github.com/rs/zerolog/log"
)

type Level string

const (
LevelDebug Level = "debug"
LevelInfo  Level = "info"
LevelWarn  Level = "warn"
LevelError Level = "error"
)

type Config struct {
Level      Level
JSONFormat bool
NoColor    bool
Output     io.Writer
}

var (
globalLogger zerolog.Logger
initialized  bool
)

func Init(cfg Config) {
if initialized {
return
}

level := parseLevel(cfg.Level)
zerolog.SetGlobalLevel(level)

output := cfg.Output
if output == nil {
output = os.Stdout
}

if cfg.JSONFormat {
globalLogger = zerolog.New(output).With().Timestamp().Logger()
} else {
consoleWriter := zerolog.ConsoleWriter{
Out:        output,
TimeFormat: time.RFC3339,
NoColor:    cfg.NoColor,
}
globalLogger = zerolog.New(consoleWriter).With().Timestamp().Logger()
}

log.Logger = globalLogger
initialized = true
}

func InitDefault() {
Init(Config{
Level:      LevelInfo,
JSONFormat: false,
NoColor:    false,
Output:     os.Stdout,
})
}

func parseLevel(level Level) zerolog.Level {
switch strings.ToLower(string(level)) {
case "debug":
return zerolog.DebugLevel
case "info":
return zerolog.InfoLevel
case "warn", "warning":
return zerolog.WarnLevel
case "error":
return zerolog.ErrorLevel
default:
return zerolog.InfoLevel
}
}

func Get(module string) zerolog.Logger {
if !initialized {
InitDefault()
}
return globalLogger.With().Str("module", module).Logger()
}
