package blobs

import logging "github.com/ipfs/go-log/v2"

type options struct{}

type Option func(*options) error

// WithLogLevel changes the log level for the claims subsystem.
func WithLogLevel(level string) Option {
	return func(c *options) error {
		logging.SetLogLevel("blobs", level)
		return nil
	}
}
