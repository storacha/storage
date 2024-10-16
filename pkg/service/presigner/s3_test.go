package presigner

import (
	"context"
	"fmt"
	"net/url"
	"testing"

	ed25519 "github.com/storacha/go-ucanto/principal/ed25519/signer"
	"github.com/storacha/storage/pkg/internal/testutil"
	"github.com/stretchr/testify/require"
)

func TestS3Signer(t *testing.T) {
	endpoint, err := url.Parse("http://localhost:3000")
	require.NoError(t, err)

	signer := testutil.RandomSigner()

	accessKeyID := signer.DID().String()
	secretAccessKey := testutil.Must(ed25519.Format(signer))(t)

	t.Run("sign and verify", func(t *testing.T) {
		reqSigner, err := NewS3RequestPresigner(accessKeyID, secretAccessKey, *endpoint, "data")
		require.NoError(t, err)

		url, headers, err := reqSigner.SignUploadURL(context.Background(), testutil.RandomMultihash(), 138, 900)
		require.NoError(t, err)

		fmt.Println(url.String())
		fmt.Printf("%+v\n", headers)

		_, _, err = reqSigner.VerifyUploadURL(context.Background(), url, headers)
		require.NoError(t, err)
	})

	t.Run("invalid URL", func(t *testing.T) {
		reqSigner, err := NewS3RequestPresigner(accessKeyID, secretAccessKey, *endpoint, "")
		require.NoError(t, err)

		url, headers, err := reqSigner.SignUploadURL(context.Background(), testutil.RandomMultihash(), 138, 900)
		require.NoError(t, err)

		// mess with the url
		url.Path += "/index.html"

		_, _, err = reqSigner.VerifyUploadURL(context.Background(), url, headers)
		require.Error(t, err)

		require.Equal(t, err.Error(), "signature verification failed")
	})

	t.Run("invalid header", func(t *testing.T) {
		reqSigner, err := NewS3RequestPresigner(accessKeyID, secretAccessKey, *endpoint, "")
		require.NoError(t, err)

		url, headers, err := reqSigner.SignUploadURL(context.Background(), testutil.RandomMultihash(), 138, 900)
		require.NoError(t, err)

		// mess with the headers
		headers.Set("Content-Length", "10000")

		_, _, err = reqSigner.VerifyUploadURL(context.Background(), url, headers)
		require.Error(t, err)

		require.Equal(t, err.Error(), "signature verification failed")
	})
}