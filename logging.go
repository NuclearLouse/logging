package logging

import (
	"fmt"
	"io"
	"os"
	"path"
	"runtime"
	"strings"

	format "github.com/antonfisher/nested-logrus-formatter"
	logrus "github.com/sirupsen/logrus"
	rotate "gopkg.in/natefinch/lumberjack.v2"
)

type Config struct {
	Level     string     `json:"loglevel" yaml:"loglevel" ini:"loglevel" cfg:"loglevel"`
	Logfile   string     `json:"logfile" yaml:"logfile" ini:"logfile" cfg:"logfile"`
	Rotation  *Rotation  `json:"logrotation" yaml:"logrotation" ini:"logrotation" cfg:"logrotation"`
	Formatter *Formatter `json:"logformatter" yaml:"logformatter" ini:"logformatter" cfg:"logformatter"`
}

type Rotation struct {
	MaxSize         int  `json:"maxsize" yaml:"maxsize" ini:"maxsize" cfg:"maxsize"`
	MaxBackups      int  `json:"maxbackups" yaml:"maxbackups" ini:"maxbackups" cfg:"maxbackups"`
	MaxAge          int  `json:"maxage" yaml:"maxage" ini:"maxage" cfg:"maxage"`
	Localtime       bool `json:"localtime" yaml:"localtime" ini:"localtime" cfg:"localtime"`
	Compress        bool `json:"compress" yaml:"compress" ini:"compress" cfg:"compress"`
	RotateAtStartup bool `json:"rotate_at_startup" yaml:"rotate_at_startup" ini:"rotate_at_startup" cfg:"rotate_at_startup"`
}

type Formatter struct {
	TimestampFormat string `json:"timestamp_format" yaml:"timestamp_format" ini:"timestamp_format" cfg:"timestamp_format"`
	HideKeys        bool   `json:"hide_keys" yaml:"hide_keys" ini:"hide_keys" cfg:"hide_keys"`
	ShowFullLevel   bool   `json:"show_full_level" yaml:"show_full_level" ini:"show_full_level" cfg:"show_full_level"`
	TraceCaller     bool   `json:"trace_caller" yaml:"trace_caller" ini:"trace_caller" cfg:"trace_caller"`
	CallerFirst     bool   `json:"caller_first" yaml:"caller_first" ini:"caller_first" cfg:"caller_first"`
	FullPathCaller  bool   `json:"full_path_caller" yaml:"full_path_caller" ini:"full_path_caller" cfg:"full_path_caller"`
}

func DefaultConfig(filename ...string) *Config {
	logfile := ""
	if filename != nil {
		logfile = filename[0]
	}
	return &Config{
		Level:   "trace",
		Logfile: logfile,
		Rotation: &Rotation{
			MaxSize:         10,
			MaxBackups:      30,
			MaxAge:          30,
			RotateAtStartup: true,
			Localtime:       true,
		},
		Formatter: &Formatter{
			HideKeys:       true,
			TraceCaller:    true,
			CallerFirst:    true,
			FullPathCaller: true,
			// Timestamp: "2006-01-02 15:04:05.000",
		},
	}
}

type Logger struct {
	*logrus.Logger
}

func New(cfg *Config) *Logger {
	log := logrus.New()

	level, err := logrus.ParseLevel(cfg.Level)
	if err != nil {
		level = logrus.TraceLevel
	}
	log.SetLevel(level)
	log.SetReportCaller(cfg.Formatter.TraceCaller)

	formatter := &format.Formatter{
		TimestampFormat: cfg.Formatter.TimestampFormat,
		ShowFullLevel:   cfg.Formatter.ShowFullLevel,
		HideKeys:        cfg.Formatter.HideKeys,
		CallerFirst:     cfg.Formatter.CallerFirst,
	}
	if !cfg.Formatter.FullPathCaller {
		formatter.CustomCallerFormatter = func(f *runtime.Frame) string {
			s := strings.Split(f.Function, ".")
			funcName := s[len(s)-1]
			return fmt.Sprintf(" (%s:%d %s) ->", path.Base(f.File), f.Line, funcName)
		}
	}

	writers := []io.Writer{os.Stderr}

	if cfg.Logfile != "" {
		formatter.NoColors = true
		rotator := &rotate.Logger{
			Filename:   cfg.Logfile,
			MaxSize:    cfg.Rotation.MaxSize,
			MaxBackups: cfg.Rotation.MaxBackups,
			MaxAge:     cfg.Rotation.MaxAge,
			Compress:   cfg.Rotation.Compress,
			LocalTime:  cfg.Rotation.Localtime,
		}

		if cfg.Rotation.RotateAtStartup {
			rotator.Rotate()
		}
		writers = append(writers, rotator)

	}

	log.SetFormatter(formatter)
	log.SetOutput(io.MultiWriter(writers...))

	return &Logger{log}
}
