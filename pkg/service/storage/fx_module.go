package storage

import (
	"go.uber.org/fx"

	"github.com/storacha/piri/pkg/config"
	"github.com/storacha/piri/pkg/service/storage/providers"
)

// Module is the main fx module for the storage service.
// It combines all sub-modules and provides the complete storage service.
var Module = fx.Module("storage",
	// Provide configuration - this should be supplied by the application
	// fx.Supply(&Config{...}) or fx.Provide(func() *Config { ... })

	// Core modules
	providers.IdentityModule,
	providers.DatastoreModule,
	providers.StoresModule,
	providers.ServicesModule,
	fx.Provide(NewStorageService),
)

// WithConfig provides a configuration to the module
func WithConfig(config config.UCANServer) fx.Option {
	return fx.Supply(config)
}

// NewApp creates a new fx application with the storage module
// This is a convenience function for creating standalone storage applications
func NewApp(config config.UCANServer, opts ...fx.Option) *fx.App {
	baseOpts := []fx.Option{
		fx.WithLogger(NewFxLogger),
		WithConfig(config),
		Module,
	}

	// Append any additional options
	allOpts := append(baseOpts, opts...)

	return fx.New(allOpts...)
}
