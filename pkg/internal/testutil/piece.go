package testutil

import (
	"fmt"
	"io"
	"math/rand"
	"testing"
	"time"

	commp "github.com/filecoin-project/go-fil-commp-hashhash"
	"github.com/storacha/go-libstoracha/piece/digest"
	"github.com/storacha/go-libstoracha/piece/piece"
	"github.com/stretchr/testify/require"
)

// CreatePiece is a helper that produces a piece with the given unpadded size.
func CreatePiece(t testing.TB, unpaddedSize int64) piece.PieceLink {
	t.Helper()

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	dataReader := io.LimitReader(r, unpaddedSize)

	calc := &commp.Calc{}
	n, err := io.Copy(calc, dataReader)
	require.NoError(t, err, "failed copying data into commp.Calc")
	require.Equal(t, unpaddedSize, n)

	commP, paddedSize, err := calc.Digest()
	require.NoError(t, err, "failed to compute commP")

	pieceDigest, err := digest.FromCommitmentAndSize(commP, uint64(unpaddedSize))
	require.NoError(t, err, "failed building piece digest from commP")

	p := piece.FromPieceDigest(pieceDigest)
	// Ensure our piece’s PaddedSize matches the commp library’s reported paddedSize.
	require.Equal(t, paddedSize, p.PaddedSize())

	t.Logf("Created test piece: %s from unpadded size: %d", pieceLinkString(p), unpaddedSize)
	return p
}

// pieceLinkString is a helper to display piece metadata in logs.
func pieceLinkString(p piece.PieceLink) string {
	return fmt.Sprintf("Piece: %s, Height: %d, Padding: %d, PaddedSize: %d",
		p.Link(), p.Height(), p.Padding(), p.PaddedSize())
}
