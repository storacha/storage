package providers

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ipfs/go-datastore"
	leveldb "github.com/ipfs/go-ds-leveldb"
	"go.uber.org/fx"

	"github.com/storacha/piri/pkg/config"
)

// DatastoreParams contains dependencies for creating datastores
type DatastoreParams struct {
	fx.In
	Config    config.UCANServer
	Lifecycle fx.Lifecycle
}

// DatastoreResult provides all datastores with proper naming
type DatastoreResult struct {
	fx.Out
	Allocation datastore.Datastore `name:"allocation"`
	Claim      datastore.Datastore `name:"claim"`
	Publisher  datastore.Datastore `name:"publisher"`
	Receipt    datastore.Datastore `name:"receipt"`
}

// NewDatastores creates all required datastores based on configuration
func NewDatastores(params DatastoreParams) (DatastoreResult, error) {
	var result DatastoreResult
	var err error

	useMemory := false
	// Create allocation datastore
	result.Allocation, err = createDatastore("allocation", params.Config.DataDir, false)
	if err != nil {
		return result, fmt.Errorf("creating allocation datastore: %w", err)
	}

	// Create claim datastore
	result.Claim, err = createDatastore("claim", params.Config.DataDir, useMemory)
	if err != nil {
		return result, fmt.Errorf("creating claim datastore: %w", err)
	}

	// Create publisher datastore
	result.Publisher, err = createDatastore("publisher", params.Config.DataDir, useMemory)
	if err != nil {
		return result, fmt.Errorf("creating publisher datastore: %w", err)
	}

	// Create receipt datastore
	result.Receipt, err = createDatastore("receipt", params.Config.DataDir, useMemory)
	if err != nil {
		return result, fmt.Errorf("creating receipt datastore: %w", err)
	}

	// Register cleanup
	params.Lifecycle.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			var errs []error
			if err := result.Allocation.Close(); err != nil {
				errs = append(errs, fmt.Errorf("closing allocation datastore: %w", err))
			}
			if err := result.Claim.Close(); err != nil {
				errs = append(errs, fmt.Errorf("closing claim datastore: %w", err))
			}
			if err := result.Publisher.Close(); err != nil {
				errs = append(errs, fmt.Errorf("closing publisher datastore: %w", err))
			}
			if err := result.Receipt.Close(); err != nil {
				errs = append(errs, fmt.Errorf("closing receipt datastore: %w", err))
			}

			if len(errs) > 0 {
				return fmt.Errorf("errors closing datastores: %v", errs)
			}
			return nil
		},
	})

	return result, nil
}

// PDPDatastoreParams for PDP-specific datastore
type PDPDatastoreParams struct {
	fx.In
	Config    config.UCANServer
	Lifecycle fx.Lifecycle
}

// NewPDPDatastore creates the PDP datastore if PDP is enabled
func NewPDPDatastore(params PDPDatastoreParams) (datastore.Datastore, error) {
	if params.Config.PDPServerURL == "" {
		return nil, nil
	}

	ds, err := createDatastore("pdp", params.Config.DataDir, false)
	if err != nil {
		return nil, fmt.Errorf("creating PDP datastore: %w", err)
	}

	params.Lifecycle.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return ds.Close()
		},
	})

	return ds, nil
}

// createDatastore creates either a persistent or in-memory datastore
func createDatastore(name, path string, useMemory bool) (datastore.Datastore, error) {
	if useMemory || path == "" {
		log.Warnf("%s datastore not configured, using in-memory datastore", name)
		return datastore.NewMapDatastore(), nil
	}

	// Ensure the directory exists
	dsPath := filepath.Join(path, name)
	if err := os.MkdirAll(dsPath, 0755); err != nil {
		return nil, fmt.Errorf("creating directory %s: %w", dsPath, err)
	}

	ds, err := leveldb.NewDatastore(dsPath, nil)
	if err != nil {
		return nil, fmt.Errorf("creating badger datastore at %s: %w", path, err)
	}

	log.Infof("Created datastore at %s", dsPath)
	return ds, nil
}

// DatastoreModule provides all datastore dependencies
var DatastoreModule = fx.Module("datastores",
	fx.Provide(
		NewDatastores,
		fx.Annotate(
			NewPDPDatastore,
			fx.ResultTags(`name:"pdp"`),
		),
	),
)
