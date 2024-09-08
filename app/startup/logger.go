package startup

import (
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"strings"
)

func BootLogger() {
	output := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: "2006-01-02 15:04:05",
		PartsOrder: []string{
			"time",
			"level",
			"caller",
			"message",
		},
		FormatLevel: func(i interface{}) string {
			var levelColor int
			switch i {
			case zerolog.LevelTraceValue:
				levelColor = 36 // Cyan
			case zerolog.LevelDebugValue:
				levelColor = 90 // Dark grey
			case zerolog.LevelInfoValue:
				levelColor = 32 // Green
			case zerolog.LevelWarnValue:
				levelColor = 33 // Yellow
			case zerolog.LevelErrorValue:
				levelColor = 31 // Red
			case zerolog.LevelFatalValue:
				levelColor = 35 // Magenta
			case zerolog.LevelPanicValue:
				levelColor = 95 // Bright magenta
			default:
				levelColor = 37 // White
			}
			return fmt.Sprintf("\x1b[%dm%s\x1b[0m", levelColor, strings.ToUpper(fmt.Sprintf("%-6s", i)))
		},
		FormatMessage: func(i interface{}) string {
			return fmt.Sprintf("%s", i)
		},
		FormatFieldName: func(i interface{}) string {
			return fmt.Sprintf("\x1b[34m%s:\x1b[0m", i) // Field names in blue
		},
		FormatFieldValue: func(i interface{}) string {
			return fmt.Sprintf("%s", i)
		},
		FormatCaller: func(i interface{}) string {
			if i == nil {
				return ""
			}
			// Extract the function and line number only
			parts := strings.Split(fmt.Sprintf("%s", i), "/")
			return parts[len(parts)-1]
		},
	}
	// Set global logger with above configuration
	log.Logger = zerolog.New(output).With().Timestamp().Caller().Logger()
	debugMode := true
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if debugMode {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}
}
