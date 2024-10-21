package testutil

import (
	"testing"

	"github.com/storacha/go-ucanto/core/delegation"
	"github.com/stretchr/testify/require"
)

// RequireEqualDelegation compares two delegations to verify their equality
func RequireEqualDelegation(t *testing.T, expected delegation.Delegation, actual delegation.Delegation) {
	if expected == nil {
		require.Nil(t, actual)
		return
	}
	require.Equal(t, expected.Issuer(), actual.Issuer())
	require.Equal(t, expected.Audience(), actual.Audience())
	require.Equal(t, expected.Capabilities(), actual.Capabilities())
	require.Equal(t, expected.Expiration(), actual.Expiration())
	require.Equal(t, expected.Signature(), actual.Signature())
	require.Equal(t, expected.Version(), actual.Version())
	require.Equal(t, expected.Facts(), actual.Facts())
	require.Equal(t, expected.Nonce(), actual.Nonce())
	require.Equal(t, expected.NotBefore(), actual.NotBefore())
	require.Equal(t, expected.Proofs(), actual.Proofs())
}
