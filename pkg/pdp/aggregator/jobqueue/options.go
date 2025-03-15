package jobqueue

import (
	"time"
)

// Config holds all parameters needed to initialize a JobQueue.
type Config struct {
	Log           StandardLogger
	JobCountLimit int
	PollInterval  time.Duration
	Extend        time.Duration
}

// Option modifies a Config before creating the JobQueue.
type Option func(*Config)

func WithLog(l StandardLogger) Option {
	return func(cfg *Config) {
		cfg.Log = l
	}
}

func WithLimit(limit int) Option {
	return func(cfg *Config) {
		cfg.JobCountLimit = limit
	}
}

func WithPollInterval(interval time.Duration) Option {
	return func(cfg *Config) {
		cfg.PollInterval = interval
	}
}

func WithExtend(d time.Duration) Option {
	return func(cfg *Config) {
		cfg.Extend = d
	}
}

// subset from ipfs go-log v2
type StandardLogger interface {
	Debug(args ...interface{})
	Debugw(msg string, keysAndValues ...interface{})
	Debugf(format string, args ...interface{})

	Info(args ...interface{})
	Infof(format string, args ...interface{})
	Infow(msg string, keysAndValues ...interface{})

	Warn(args ...interface{})
	Warnf(format string, args ...interface{})
	Warnw(msg string, keysAndValues ...interface{})

	Error(args ...interface{})
	Errorf(format string, args ...interface{})
	Errorw(msg string, keysAndValues ...interface{})
}

type discardLogger struct{}

var _ StandardLogger = (*discardLogger)(nil)

func (d *discardLogger) Debug(args ...interface{})                       {}
func (d *discardLogger) Debugf(format string, args ...interface{})       {}
func (d *discardLogger) Debugw(msg string, keysAndValues ...interface{}) {}
func (d *discardLogger) Error(args ...interface{})                       {}
func (d *discardLogger) Errorf(format string, args ...interface{})       {}
func (d *discardLogger) Errorw(msg string, keysAndValues ...interface{}) {}
func (d *discardLogger) Infof(format string, args ...interface{})        {}
func (d *discardLogger) Info(args ...interface{})                        {}
func (d *discardLogger) Infow(msg string, keysAndValues ...interface{})  {}
func (d *discardLogger) Warn(args ...interface{})                        {}
func (d *discardLogger) Warnf(format string, args ...interface{})        {}
func (d *discardLogger) Warnw(msg string, keysAndValues ...interface{})  {}
