package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

func setupLogging(ctx context.Context) context.Context {
	logFile := os.Stdout

	var output io.Writer

	switch config.LogFormat {
	case "json":
		output = logFile
	case "text":
		prefixList := []string{}
		info, ok := debug.ReadBuildInfo()
		if ok {
			prefixList = append(prefixList, info.Path+"/")
		}

		basedir := ""
		_, sourceFile, _, ok := runtime.Caller(0)
		if ok {
			basedir = filepath.Dir(sourceFile)
		}

		if basedir != "" && strings.HasPrefix(basedir, "/") {
			prefixList = append(prefixList, basedir+"/")
			head, _ := filepath.Split(basedir)
			for head != "/" {
				prefixList = append(prefixList, head)
				head, _ = filepath.Split(strings.TrimSuffix(head, "/"))
			}
		}

		output = zerolog.ConsoleWriter{
			Out:        logFile,
			NoColor:    true,
			TimeFormat: time.RFC3339,
			PartsOrder: []string{
				zerolog.LevelFieldName,
				zerolog.TimestampFieldName,
				zerolog.CallerFieldName,
				zerolog.MessageFieldName,
			},
			FormatFieldName:  func(i interface{}) string { return fmt.Sprintf("%s:", i) },
			FormatFieldValue: func(i interface{}) string { return fmt.Sprintf("%s", i) },
			FormatCaller: func(i interface{}) string {
				s := i.(string)
				for _, p := range prefixList {
					s = strings.TrimPrefix(s, p)
				}
				return s
			},
		}
	default:
		log.Fatalf("Invalid log format specified: %q", config.LogFormat)
	}

	logger := zerolog.New(output).Level(zerolog.Level(config.LogLevel)).With().Caller().Timestamp().Logger()

	ctx = logger.WithContext(ctx)

	zerolog.DefaultContextLogger = &logger
	log.SetOutput(logger)

	return ctx
}
