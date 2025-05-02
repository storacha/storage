package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/go-multierror"
	"github.com/spf13/viper"
	"github.com/urfave/cli/v2"

	"github.com/storacha/storage/pkg/presets"
)

const DefaultServicePort = 3000

// CoreConfig contains the core settings for the storage node
type CoreConfig struct {
	KeyFilePath string `toml:"key_file" json:"key_file" mapstructure:"key_file"`
	ServerPort  int    `toml:"port" json:"port" mapstructure:"port"`
	PublicURL   string `toml:"public_url" json:"public_url" mapstructure:"public_url"`
}

// DirectoriesConfig contains file system paths for the storage node
type DirectoriesConfig struct {
	DataDir string `toml:"data_dir" json:"data_dir" mapstructure:"data_dir"`
	TempDir string `toml:"temp_dir" json:"temp_dir" mapstructure:"temp_dir"`
}

// IndexingConfig contains settings for the indexing service
type IndexingConfig struct {
	StorageProof string `toml:"storage_proof" json:"storage_proof" mapstructure:"storage_proof"`
	ServiceDID   string `toml:"service_did" json:"service_did" mapstructure:"service_did"`
	ServiceURL   string `toml:"service_url" json:"service_url" mapstructure:"service_url"`
	AnnounceURL  string `toml:"announce_url" json:"announce_url" mapstructure:"announce_url"`
}

// UploadConfig contains settings for the upload service
type UploadConfig struct {
	ServiceDID string `toml:"service_did" json:"service_did" mapstructure:"service_did"`
	ServiceURL string `toml:"service_url" json:"service_url" mapstructure:"service_url"`
}

// PDPConfig holds configuration for the Proof of Data Possession subsystem
type PDPConfig struct {
	ServerURL string `toml:"server_url" json:"server_url" mapstructure:"server_url"`
	ProofSet  uint64 `toml:"proof_set" json:"proof_set" mapstructure:"proof_set"`
}

// Node represents the full configuration for a storage node
type Node struct {
	// Core settings
	Core CoreConfig `toml:"core" json:"core" mapstructure:"core"`

	// Data storage locations
	Directories DirectoriesConfig `toml:"directories" json:"directories" mapstructure:"directories"`

	// Indexing service configuration
	Indexing IndexingConfig `toml:"indexing" json:"indexing" mapstructure:"indexing"`

	// Upload service configuration
	Upload UploadConfig `toml:"upload" json:"upload" mapstructure:"upload"`

	// PDP (Proof of Data Possession) configuration
	PDP *PDPConfig `toml:"pdp" json:"pdp" mapstructure:"pdp"`

	// Storage principal mapping
	Principals map[string]string `toml:"principals" json:"principals" mapstructure:"principals"`
}

// LoadConfig is a comprehensive method that handles the entire configuration loading process
// flags > environment variables > config file > defaults
// It takes care of:
// 1. Loading defaults specified in code
// 2. Loading config from file if provided via --config
// 3. Setting up default directories if they are not provided
// 4. Applying CLI flag overrides to config state
// 5. Validating the final configuration
func LoadConfig(cCtx *cli.Context) (*Node, error) {
	// Start with defaults
	cfg := newDefault()

	// load from config file if specified
	configPath := cCtx.String("config")
	if configPath != "" {
		loadedCfg, err := load(configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load configuration from file: %w", err)
		}
		cfg = loadedCfg
	}

	// Set up default directories (creates them if they don't exist)
	if err := setupDefaultDirectories(cfg); err != nil {
		return nil, fmt.Errorf("failed to set up default directories: %w", err)
	}

	// Apply CLI flags and environment variable overrides
	fromCLI(cCtx, cfg)

	// Validate the final configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return cfg, nil
}

// Validate performs validation on the configuration values and returns any errors.
// This can be called before using the configuration to ensure all required values are set.
func (cfg *Node) Validate() error {
	var errs error
	// Validate Core settings
	if cfg.Core.KeyFilePath == "" {
		errs = multierror.Append(errs, fmt.Errorf("key file path is required"))
	}

	if cfg.Core.ServerPort <= 0 || cfg.Core.ServerPort > 65535 {
		errs = multierror.Append(errs, fmt.Errorf("invalid server port: %d, must be between 1 and 65535", cfg.Core.ServerPort))
	}

	// Validate directory paths
	if cfg.Directories.DataDir == "" {
		errs = multierror.Append(errs, fmt.Errorf("data directory path is required"))
	}

	if cfg.Directories.TempDir == "" {
		errs = multierror.Append(errs, fmt.Errorf("temporary directory path is required"))
	}

	// Validate PDP configuration if provided
	if cfg.PDP != nil && cfg.PDP.ServerURL != "" {
		if cfg.PDP.ProofSet == 0 {
			errs = multierror.Append(errs, fmt.Errorf("PDP proof set must be specified when PDP server URL is provided"))
		}
	}

	return errs
}

// load reads the configuration from the given path and returns a Node.
// It binds and loads values from environment variables.
// It preserves default values for fields not specified in the config file, flags or env vars
func load(path string) (*Node, error) {
	if len(path) == 0 {
		return nil, fmt.Errorf("config file path cannot be empty")
	}
	if stat, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("config file path does not exist: %s", path)
		}
		return nil, fmt.Errorf("failed to read config file at path %s: %w", path, err)
	} else if stat.IsDir() {
		return nil, fmt.Errorf("config file path points to a directory: %s", path)
	}

	// Initialize viper with defaults
	v, err := setupViperWithDefaults()
	if err != nil {
		return nil, err
	}

	// Read the configuration file
	v.SetConfigFile(path)
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Unmarshal into our config struct
	cfg := new(Node)
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config file: %w", err)
	}

	// Initialize empty maps if nil to avoid nil pointer panics
	if cfg.Principals == nil {
		cfg.Principals = make(map[string]string)
	}

	return cfg, nil
}

// newDefault creates a new configuration with pure default values.
// This only sets defaults that are not dependent on platform-specific logic.
// Application code should handle platform-specific defaults like file paths.
func newDefault() *Node {
	return &Node{
		Core: CoreConfig{
			PublicURL:  fmt.Sprintf("http://localhost:%d", DefaultServicePort),
			ServerPort: DefaultServicePort, // Default HTTP port
		},
		Directories: DirectoriesConfig{
			// No defaults for paths - platform-specific logic should set these
		},
		Indexing: IndexingConfig{
			AnnounceURL: presets.AnnounceURL.String(),
			ServiceDID:  presets.IndexingServiceDID.String(),
			ServiceURL:  presets.IndexingServiceURL.String(),
		},
		Upload: UploadConfig{
			ServiceDID: presets.UploadServiceDID.String(),
			ServiceURL: presets.UploadServiceURL.String(),
		},
		Principals: make(map[string]string),
	}
}

// fromCLI loads configuration values from CLI flags
func fromCLI(ctx *cli.Context, cfg *Node) {
	// Core settings
	if ctx.IsSet("key-file") {
		cfg.Core.KeyFilePath = ctx.String("key-file")
	}
	if ctx.IsSet("port") {
		cfg.Core.ServerPort = ctx.Int("port")
	}
	if ctx.IsSet("public-url") {
		cfg.Core.PublicURL = ctx.String("public-url")
	}

	// Directory settings
	if ctx.IsSet("data-dir") {
		cfg.Directories.DataDir = ctx.String("data-dir")
	}
	if ctx.IsSet("tmp-dir") {
		cfg.Directories.TempDir = ctx.String("tmp-dir")
	}

	// Indexing settings
	if ctx.IsSet("indexing-service-proof") {
		cfg.Indexing.StorageProof = ctx.String("indexing-service-proof")
	}
	if ctx.IsSet("indexing-service-url") {
		cfg.Indexing.ServiceURL = ctx.String("indexing-service-url")
	}
	if ctx.IsSet("indexing-service-did") {
		cfg.Indexing.ServiceDID = ctx.String("indexing-service-did")
	}
	if ctx.IsSet("announce-url") {
		cfg.Indexing.AnnounceURL = ctx.String("announce-url")
	}

	// Upload settings
	if ctx.IsSet("upload-service-url") {
		cfg.Upload.ServiceURL = ctx.String("upload-service-url")
	}
	if ctx.IsSet("upload-service-did") {
		cfg.Upload.ServiceDID = ctx.String("upload-service-did")
	}

	// PDP settings
	if ctx.IsSet("curio-url") {
		if cfg.PDP == nil {
			cfg.PDP = &PDPConfig{}
		}
		cfg.PDP.ServerURL = ctx.String("curio-url")
	}
	if ctx.IsSet("pdp-proofset") {
		if cfg.PDP == nil {
			cfg.PDP = &PDPConfig{}
		}
		cfg.PDP.ProofSet = ctx.Uint64("pdp-proofset")
	}
}

// setupViperWithDefaults creates a new Viper instance with default values and environment bindings
func setupViperWithDefaults() (*viper.Viper, error) {
	v := viper.New()

	// Set up environment variable binding
	v.SetEnvPrefix("STORAGE")
	v.AutomaticEnv()

	// Define specific environment variable mappings
	envMappings := map[string]string{
		// Core
		"core.key_file":   "PRIVATE_KEY",
		"core.port":       "PORT",
		"core.public_url": "PUBLIC_URL",

		// Directories
		"directories.data_dir": "DATA_DIR",
		"directories.temp_dir": "TMP_DIR",

		// Indexing
		"indexing.storage_proof": "INDEXING_SERVICE_PROOF",
		"indexing.service_did":   "INDEXING_SERVICE_DID",
		"indexing.service_url":   "INDEXING_SERVICE_URL",
		"indexing.announce_url":  "ANNOUNCE_URL",

		// Upload
		"upload.service_did": "UPLOAD_SERVICE_DID",
		"upload.service_url": "UPLOAD_SERVICE_URL",

		// PDP
		"pdp.server_url": "CURIO_URL",
		"pdp.proof_set":  "PDP_PROOFSET",
	}

	// Create the aliases for environment variables
	for key, envVar := range envMappings {
		if err := v.BindEnv(key, "STORAGE_"+envVar); err != nil {
			return nil, fmt.Errorf("failed to bind environment variable %s: %w", key, err)
		}
	}

	// Start with default values
	defaultCfg := newDefault()

	// Set default values in Viper
	v.SetDefault("core.port", defaultCfg.Core.ServerPort)
	v.SetDefault("core.public_url", defaultCfg.Core.PublicURL)
	v.SetDefault("indexing.announce_url", defaultCfg.Indexing.AnnounceURL)
	v.SetDefault("indexing.service_did", defaultCfg.Indexing.ServiceDID)
	v.SetDefault("indexing.service_url", defaultCfg.Indexing.ServiceURL)
	v.SetDefault("upload.service_did", defaultCfg.Upload.ServiceDID)
	v.SetDefault("upload.service_url", defaultCfg.Upload.ServiceURL)

	return v, nil
}

// setupDefaultDirectories configures default directories if they are not already set
func setupDefaultDirectories(cfg *Node) error {
	if cfg.Directories.DataDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("getting user home directory: %w", err)
		}

		dataDir := filepath.Join(homeDir, ".storacha")
		if err := os.MkdirAll(dataDir, 0755); err != nil {
			return fmt.Errorf("creating default data directory %s: %w", dataDir, err)
		}
		cfg.Directories.DataDir = dataDir
	}

	if cfg.Directories.TempDir == "" {
		tempDir := filepath.Join(os.TempDir(), "storage")
		if err := os.MkdirAll(tempDir, 0755); err != nil {
			return fmt.Errorf("creating default temp directory %s: %w", tempDir, err)
		}
		cfg.Directories.TempDir = tempDir
	}

	return nil
}
