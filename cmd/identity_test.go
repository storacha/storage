package cmd_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	ed25519 "github.com/storacha/go-ucanto/principal/ed25519/signer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/storacha/storage/cmd"
	"github.com/storacha/storage/cmd/enum"
)

func TestCreateSignerKeyPairAndPrincipalSignerFromFile(t *testing.T) {
	tests := []struct {
		name   string
		format enum.KeyFormat
		ext    string
	}{
		{
			name:   "JSON round-trip",
			format: enum.KeyFormats.JSON,
			ext:    "json",
		},
		{
			name:   "PEM round-trip",
			format: enum.KeyFormats.PEM,
			ext:    "pem",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// create the signer/key bytes in the specified format
			origSigner, keyBytes, err := cmd.CreateSignerKeyPair(tt.format)
			require.NoError(t, err, "CreateSignerKeyPair should succeed")

			// write the key bytes to a temporary file
			tmpDir := t.TempDir()
			filePath := filepath.Join(tmpDir, "test."+tt.ext)
			err = os.WriteFile(filePath, keyBytes, 0600)
			require.NoError(t, err, "should be able to write key file")

			// read them back into a new signer
			loadedSigner, err := cmd.PrincipalSignerFromFile(filePath)
			require.NoError(t, err, "PrincipalSignerFromFile should succeed")

			// confirm private keys match
			require.True(t, bytes.Equal(origSigner.Raw(), loadedSigner.Raw()),
				"loaded signer's private key should match the original")

			// confirm that the DIDs match as well
			require.Equal(t, origSigner.DID().String(), loadedSigner.DID().String(),
				"DID of loaded signer should match the original")
		})
	}
}

func TestRoundTripConversion(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "identity-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Generate a signer
	signer, err := ed25519.Generate()
	require.NoError(t, err)

	// Test: JSON → PEM → JSON round trip
	t.Run("JSON to PEM to JSON", func(t *testing.T) {
		// Step 1: Create JSON key
		jsonKey, err := cmd.MarshalJSONKey(signer)
		require.NoError(t, err)

		// Write JSON key to file
		jsonPath := filepath.Join(tempDir, "key.json")
		err = os.WriteFile(jsonPath, jsonKey, 0600)
		require.NoError(t, err)

		// Step 2: Convert JSON to PEM
		pemBytes, err := cmd.JSONFileToPEM(jsonPath)
		require.NoError(t, err)

		// Write PEM key to file
		pemPath := filepath.Join(tempDir, "key.pem")
		err = os.WriteFile(pemPath, pemBytes, 0600)
		require.NoError(t, err)

		// Step 3: Convert PEM back to JSON
		jsonBytes2, err := cmd.PEMFileToJSON(pemPath)
		require.NoError(t, err)

		// Verify the contents
		var originalKey, convertedKey cmd.JsonKey
		err = json.Unmarshal(jsonKey, &originalKey)
		require.NoError(t, err)
		err = json.Unmarshal(jsonBytes2, &convertedKey)
		require.NoError(t, err)

		assert.Equal(t, originalKey.DID, convertedKey.DID, "DIDs should match after round trip")
		assert.Equal(t, originalKey.Key, convertedKey.Key, "Keys should match after round trip")
	})

	// Test: PEM → JSON → PEM round trip
	t.Run("PEM to JSON to PEM", func(t *testing.T) {
		// Step 1: Create PEM key
		pemKey, err := cmd.MarshalPEMKey(signer)
		require.NoError(t, err)

		// Write PEM key to file
		pemPath := filepath.Join(tempDir, "key2.pem")
		err = os.WriteFile(pemPath, pemKey, 0600)
		require.NoError(t, err)

		// Step 2: Convert PEM to JSON
		jsonBytes, err := cmd.PEMFileToJSON(pemPath)
		require.NoError(t, err)

		// Write JSON key to file
		jsonPath := filepath.Join(tempDir, "key2.json")
		err = os.WriteFile(jsonPath, jsonBytes, 0600)
		require.NoError(t, err)

		// Step 3: Convert JSON back to PEM
		pemBytes2, err := cmd.JSONFileToPEM(jsonPath)
		require.NoError(t, err)

		// Verify the contents - PEM content should be equivalent
		// Note: PEM encoding may have some differences but the extracted keys should be the same
		assert.True(t, isPEMEquivalent(t, pemKey, pemBytes2), "PEM keys should be equivalent after round trip")
	})

	// Test using the CreateSignerKeyPair function
	t.Run("CreateSignerKeyPair round trip", func(t *testing.T) {
		// Generate a key pair in JSON format
		_, jsonBytes, err := cmd.CreateSignerKeyPair(enum.KeyFormats.JSON)
		require.NoError(t, err)

		jsonPath := filepath.Join(tempDir, "generated.json")
		err = os.WriteFile(jsonPath, jsonBytes, 0600)
		require.NoError(t, err)

		// Convert JSON to PEM
		pemBytes, err := cmd.JSONFileToPEM(jsonPath)
		require.NoError(t, err)

		pemPath := filepath.Join(tempDir, "converted.pem")
		err = os.WriteFile(pemPath, pemBytes, 0600)
		require.NoError(t, err)

		// Convert PEM back to JSON
		jsonBytes2, err := cmd.PEMFileToJSON(pemPath)
		require.NoError(t, err)

		// Verify results
		var original, converted cmd.JsonKey
		err = json.Unmarshal(jsonBytes, &original)
		require.NoError(t, err)
		err = json.Unmarshal(jsonBytes2, &converted)
		require.NoError(t, err)

		assert.Equal(t, original.DID, converted.DID, "DID should match after round trip")
		assert.Equal(t, original.Key, converted.Key, "Key should match after round trip")
	})
}

// isPEMEquivalent determines if two PEM-encoded keys are functionally equivalent
// This helper function extracts the signers from each PEM format and compares them
func isPEMEquivalent(t *testing.T, pem1, pem2 []byte) bool {
	// Write both PEMs to temporary files
	tempDir, err := os.MkdirTemp("", "pem-compare")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	pem1Path := filepath.Join(tempDir, "key1.pem")
	pem2Path := filepath.Join(tempDir, "key2.pem")

	err = os.WriteFile(pem1Path, pem1, 0600)
	require.NoError(t, err)
	err = os.WriteFile(pem2Path, pem2, 0600)
	require.NoError(t, err)

	// Convert both to JSON and compare DIDs and Keys
	json1, err := cmd.PEMFileToJSON(pem1Path)
	require.NoError(t, err)
	json2, err := cmd.PEMFileToJSON(pem2Path)
	require.NoError(t, err)

	var key1, key2 cmd.JsonKey
	err = json.Unmarshal(json1, &key1)
	require.NoError(t, err)
	err = json.Unmarshal(json2, &key2)
	require.NoError(t, err)

	return key1.DID == key2.DID && key1.Key == key2.Key
}
