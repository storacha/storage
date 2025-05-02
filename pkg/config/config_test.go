package config

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"

	"github.com/storacha/storage/pkg/presets"
)

// TestDefaultConfig verifies that the default configuration is set correctly
func TestDefaultConfig(t *testing.T) {
	// Get a default config
	cfg := newDefault()

	// Verify default values
	assert.Equal(t, DefaultServicePort, cfg.Core.ServerPort)
	assert.Equal(t, fmt.Sprintf("http://localhost:%d", DefaultServicePort), cfg.Core.PublicURL)
	assert.Empty(t, cfg.Core.KeyFilePath)
	assert.Empty(t, cfg.Directories.DataDir)
	assert.Empty(t, cfg.Directories.TempDir)

	// Ensure defaults are set for services
	assert.Equal(t, presets.AnnounceURL.String(), cfg.Indexing.AnnounceURL)
	assert.Equal(t, presets.IndexingServiceDID.String(), cfg.Indexing.ServiceDID)
	assert.Equal(t, presets.IndexingServiceURL.String(), cfg.Indexing.ServiceURL)
	assert.Equal(t, presets.UploadServiceDID.String(), cfg.Upload.ServiceDID)
	assert.Equal(t, presets.UploadServiceURL.String(), cfg.Upload.ServiceURL)

	// Ensure empty but initialized maps
	assert.NotNil(t, cfg.Principals)
	assert.Empty(t, cfg.Principals)

	// Ensure PDP is not set by default
	assert.Nil(t, cfg.PDP)
}

// TestDirectorySetup tests that default directories are created properly
func TestDirectorySetup(t *testing.T) {
	// Create a temporary directory for testing
	t.Run("default, no dir paths provided", func(t *testing.T) {
		testDir := t.TempDir()

		// Temporarily replace the os.UserHomeDir function to return our test directory
		t.Setenv("HOME", testDir)

		// Test with empty directories
		cfg := newDefault()

		// Setup directories should create the defaults
		err := setupDefaultDirectories(cfg)
		require.NoError(t, err)

		// Verify the directories were set to the defaults
		assert.Equal(t, filepath.Join(testDir, ".storacha"), cfg.Directories.DataDir)
		assert.Equal(t, filepath.Join(os.TempDir(), "storage"), cfg.Directories.TempDir)

		// Verify the directories were created
		_, err = os.Stat(cfg.Directories.DataDir)
		assert.NoError(t, err)
		_, err = os.Stat(cfg.Directories.TempDir)
		assert.NoError(t, err)

	})

	t.Run("dirs set explicitly", func(t *testing.T) {
		testDir := t.TempDir()

		// Test with predefined directories
		cfg := newDefault()
		customDataDir := filepath.Join(testDir, "custom-data")
		customTempDir := filepath.Join(testDir, "custom-temp")
		require.NoError(t, os.Mkdir(customDataDir, os.ModePerm))
		require.NoError(t, os.Mkdir(customTempDir, os.ModePerm))
		cfg.Directories.DataDir = customDataDir
		cfg.Directories.TempDir = customTempDir

		// Setup should not override the existing directories
		err := setupDefaultDirectories(cfg)
		require.NoError(t, err)

		// Verify the directories remain unchanged
		assert.Equal(t, customDataDir, cfg.Directories.DataDir)
		assert.Equal(t, customTempDir, cfg.Directories.TempDir)

		// Verify the directories were created
		_, err = os.Stat(customDataDir)
		assert.NoError(t, err)
		_, err = os.Stat(customTempDir)
		assert.NoError(t, err)

	})

}

// TestValidate tests the configuration validation function
func TestValidate(t *testing.T) {
	// Create a valid configuration
	cfg := newDefault()
	cfg.Core.KeyFilePath = "/path/to/key.pem"
	cfg.Directories.DataDir = "/data/dir"
	cfg.Directories.TempDir = "/tmp/dir"

	// Valid configuration should not return an error
	err := cfg.Validate()
	assert.NoError(t, err)

	// Test invalid port
	invalidCfg := *cfg
	invalidCfg.Core.ServerPort = 0
	err = invalidCfg.Validate()
	assert.Error(t, err)

	// Test missing key file
	invalidCfg = *cfg
	invalidCfg.Core.KeyFilePath = ""
	err = invalidCfg.Validate()
	assert.Error(t, err)

	// Test missing data directory
	invalidCfg = *cfg
	invalidCfg.Directories.DataDir = ""
	err = invalidCfg.Validate()
	assert.Error(t, err)

	// Test missing temp directory
	invalidCfg = *cfg
	invalidCfg.Directories.TempDir = ""
	err = invalidCfg.Validate()
	assert.Error(t, err)

	// Test PDP config with server URL but no proof set
	invalidCfg = *cfg
	invalidCfg.PDP = &PDPConfig{
		ServerURL: "http://pdp.example.com",
		ProofSet:  0,
	}
	err = invalidCfg.Validate()
	assert.Error(t, err)

	// Test valid PDP config
	validCfg := *cfg
	validCfg.PDP = &PDPConfig{
		ServerURL: "http://pdp.example.com",
		ProofSet:  123,
	}
	err = validCfg.Validate()
	assert.NoError(t, err)
}

// mockCLIContext creates a CLI context for testing with the given flag values
func mockCLIContext(t *testing.T, args ...string) *cli.Context {
	app := &cli.App{
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "config"},
			&cli.StringFlag{Name: "key-file"},
			&cli.IntFlag{Name: "port"},
			&cli.StringFlag{Name: "public-url"},
			&cli.StringFlag{Name: "data-dir"},
			&cli.StringFlag{Name: "tmp-dir"},
			&cli.StringFlag{Name: "indexing-service-proof"},
			&cli.StringFlag{Name: "indexing-service-url"},
			&cli.StringFlag{Name: "indexing-service-did"},
			&cli.StringFlag{Name: "announce-url"},
			&cli.StringFlag{Name: "upload-service-url"},
			&cli.StringFlag{Name: "upload-service-did"},
			&cli.StringFlag{Name: "curio-url"},
			&cli.Uint64Flag{Name: "pdp-proofset"},
		},
		Action: func(ctx *cli.Context) error {
			return nil
		},
	}

	// Reset flag.CommandLine to avoid conflicts between tests
	flag.CommandLine = flag.NewFlagSet("test", flag.ExitOnError)

	// Create a CLI context with the provided args
	return cli.NewContext(app, flag.NewFlagSet("test", flag.ContinueOnError), nil)
}

// TestFromCLI tests that CLI flags are correctly applied to the configuration
func TestFromCLI(t *testing.T) {
	// Create a default configuration
	cfg := newDefault()

	// Create a CLI context with no flags set
	ctx := mockCLIContext(t)

	// Apply CLI flags (none should be set)
	fromCLI(ctx, cfg)

	// Configuration should remain unchanged
	defaultCfg := newDefault()
	assert.Equal(t, defaultCfg.Core.ServerPort, cfg.Core.ServerPort)
	assert.Equal(t, defaultCfg.Core.PublicURL, cfg.Core.PublicURL)

	// Create a CLI context with various flags set
	ctx = mockCLIContext(t)

	// Set flags manually
	flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
	flagSet.String("key-file", "/path/to/key.pem", "")
	flagSet.Int("port", 8080, "")
	flagSet.String("public-url", "https://example.com", "")
	flagSet.String("data-dir", "/data/dir", "")
	flagSet.String("tmp-dir", "/tmp/dir", "")
	flagSet.String("indexing-service-proof", "test-proof", "")
	flagSet.String("indexing-service-url", "https://indexing.example.com", "")
	flagSet.String("indexing-service-did", "did:key:test-indexing", "")
	flagSet.String("announce-url", "https://announce.example.com", "")
	flagSet.String("upload-service-url", "https://upload.example.com", "")
	flagSet.String("upload-service-did", "did:key:test-upload", "")
	flagSet.String("curio-url", "https://curio.example.com", "")
	flagSet.Uint64("pdp-proofset", 123, "")

	// Parse the flags
	err := flagSet.Parse([]string{
		"--key-file", "/path/to/key.pem",
		"--port", "8080",
		"--public-url", "https://example.com",
		"--data-dir", "/data/dir",
		"--tmp-dir", "/tmp/dir",
		"--indexing-service-proof", "test-proof",
		"--indexing-service-url", "https://indexing.example.com",
		"--indexing-service-did", "did:key:test-indexing",
		"--announce-url", "https://announce.example.com",
		"--upload-service-url", "https://upload.example.com",
		"--upload-service-did", "did:key:test-upload",
		"--curio-url", "https://curio.example.com",
		"--pdp-proofset", "123",
	})
	require.NoError(t, err)

	// Create a new CLI context with the parsed flags
	ctxWithFlags := cli.NewContext(nil, flagSet, nil)

	// Apply CLI flags to the configuration
	cfg = newDefault()
	fromCLI(ctxWithFlags, cfg)

	// Verify the configuration was updated properly
	assert.Equal(t, "/path/to/key.pem", cfg.Core.KeyFilePath)
	assert.Equal(t, 8080, cfg.Core.ServerPort)
	assert.Equal(t, "https://example.com", cfg.Core.PublicURL)
	assert.Equal(t, "/data/dir", cfg.Directories.DataDir)
	assert.Equal(t, "/tmp/dir", cfg.Directories.TempDir)
	assert.Equal(t, "test-proof", cfg.Indexing.StorageProof)
	assert.Equal(t, "https://indexing.example.com", cfg.Indexing.ServiceURL)
	assert.Equal(t, "did:key:test-indexing", cfg.Indexing.ServiceDID)
	assert.Equal(t, "https://announce.example.com", cfg.Indexing.AnnounceURL)
	assert.Equal(t, "https://upload.example.com", cfg.Upload.ServiceURL)
	assert.Equal(t, "did:key:test-upload", cfg.Upload.ServiceDID)
	assert.NotNil(t, cfg.PDP)
	assert.Equal(t, "https://curio.example.com", cfg.PDP.ServerURL)
	assert.Equal(t, uint64(123), cfg.PDP.ProofSet)
}

// Creating a temporary file for testing config loading
func createTempConfigFile(t *testing.T, content string) string {
	tempFile, err := os.CreateTemp("", "config-*.toml")
	require.NoError(t, err)
	defer tempFile.Close()

	_, err = tempFile.WriteString(content)
	require.NoError(t, err)

	return tempFile.Name()
}

// TestLoadConfig tests loading configuration from a file
func TestLoadConfig(t *testing.T) {
	// Create a temporary config file
	configContent := `
[core]
key_file = "/path/to/key.toml"
port = 9090
public_url = "https://config-file.example.com"

[directories]
data_dir = "/config/data"
temp_dir = "/config/temp"

[indexing]
storage_proof = "config-proof"
service_did = "did:key:config-indexing"
service_url = "https://config-indexing.example.com"
announce_url = "https://config-announce.example.com"

[upload]
service_did = "did:key:config-upload"
service_url = "https://config-upload.example.com"

[pdp]
server_url = "https://config-curio.example.com"
proof_set = 456

[principals]
key1 = "value1"
key2 = "value2"
`
	configPath := createTempConfigFile(t, configContent)
	defer os.Remove(configPath)

	// Load the config from the file
	cfg, err := load(configPath)
	require.NoError(t, err)

	// Verify the config was loaded properly
	assert.Equal(t, "/path/to/key.toml", cfg.Core.KeyFilePath)
	assert.Equal(t, 9090, cfg.Core.ServerPort)
	assert.Equal(t, "https://config-file.example.com", cfg.Core.PublicURL)
	assert.Equal(t, "/config/data", cfg.Directories.DataDir)
	assert.Equal(t, "/config/temp", cfg.Directories.TempDir)
	assert.Equal(t, "config-proof", cfg.Indexing.StorageProof)
	assert.Equal(t, "did:key:config-indexing", cfg.Indexing.ServiceDID)
	assert.Equal(t, "https://config-indexing.example.com", cfg.Indexing.ServiceURL)
	assert.Equal(t, "https://config-announce.example.com", cfg.Indexing.AnnounceURL)
	assert.Equal(t, "did:key:config-upload", cfg.Upload.ServiceDID)
	assert.Equal(t, "https://config-upload.example.com", cfg.Upload.ServiceURL)
	assert.NotNil(t, cfg.PDP)
	assert.Equal(t, "https://config-curio.example.com", cfg.PDP.ServerURL)
	assert.Equal(t, uint64(456), cfg.PDP.ProofSet)
	assert.Equal(t, map[string]string{"key1": "value1", "key2": "value2"}, cfg.Principals)

	// Test error cases
	_, err = load("")
	assert.Error(t, err, "Empty path should return an error")

	_, err = load("/non/existent/path")
	assert.Error(t, err, "Non-existent path should return an error")

	// Create a temporary directory to test directory path error
	tempDir, err := os.MkdirTemp("", "config-test-dir-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	_, err = load(tempDir)
	assert.Error(t, err, "Directory path should return an error")
}

// TestConfigPrecedence tests that configuration values are loaded with the correct precedence:
// CLI flags > Environment variables > Config file > Defaults
func TestConfigPrecedence(t *testing.T) {
	// Create a temporary directory for testing
	testDir := t.TempDir()

	// Temporarily replace the home dir
	t.Setenv("HOME", testDir)

	// 1. Create a config file with some values
	configContent := `
[core]
port = 9090
public_url = "https://config-file.example.com"

[directories]
data_dir = "/config/data"
temp_dir = "/config/temp"

[indexing]
service_url = "https://config-indexing.example.com"
`
	configPath := createTempConfigFile(t, configContent)
	defer os.Remove(configPath)

	// 2. Set environment variables with different values
	t.Setenv("STORAGE_PORT", "8080")
	t.Setenv("STORAGE_PUBLIC_URL", "https://env-var.example.com")
	t.Setenv("STORAGE_INDEXING_SERVICE_URL", "https://env-indexing.example.com")

	// 3. Set CLI flags with yet different values
	flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
	flagSet.String("config", configPath, "")
	flagSet.String("key-file", "/path/to/key.pem", "")
	flagSet.Int("port", 7070, "")
	flagSet.String("indexing-service-url", "https://cli-indexing.example.com", "")

	err := flagSet.Parse([]string{
		"--config", configPath,
		"--key-file", "/path/to/key.pem",
		"--port", "7070",
		"--indexing-service-url", "https://cli-indexing.example.com",
	})
	require.NoError(t, err)

	ctxWithFlags := cli.NewContext(nil, flagSet, nil)

	// 4. Load the configuration with all sources
	cfg, err := LoadConfig(ctxWithFlags)
	require.NoError(t, err)

	// 5. Verify the precedence: CLI flags > Environment > Config file > Defaults

	// CLI flags should have highest precedence
	assert.Equal(t, "/path/to/key.pem", cfg.Core.KeyFilePath)                    // CLI flag
	assert.Equal(t, 7070, cfg.Core.ServerPort)                                   // CLI flag overrides env and config
	assert.Equal(t, "https://cli-indexing.example.com", cfg.Indexing.ServiceURL) // CLI flag overrides env and config

	// Environment vars should override config file but not CLI flags
	assert.Equal(t, "https://env-var.example.com", cfg.Core.PublicURL) // Env var overrides config

	// Config file should override defaults but not env vars or CLI flags
	assert.Equal(t, "/config/data", cfg.Directories.DataDir) // From config file
	assert.Equal(t, "/config/temp", cfg.Directories.TempDir) // From config file

	// Default values should be used when not specified elsewhere
	defaultCfg := newDefault()
	assert.Equal(t, defaultCfg.Upload.ServiceDID, cfg.Upload.ServiceDID) // Default value
	assert.Equal(t, defaultCfg.Upload.ServiceURL, cfg.Upload.ServiceURL) // Default value
}
