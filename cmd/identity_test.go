package cmd_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/storacha/storage/cmd"
	"github.com/storacha/storage/cmd/enum"
	"github.com/stretchr/testify/require"
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
