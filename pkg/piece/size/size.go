package size

import (
	"fmt"
	"math/bits"
)

// Fr32PaddedSizeToV1TreeHeight calculates the height of the piece tree given data that's been FR32 padded. Because
// pieces are only defined on binary trees if the size is not a power of 2 it will be rounded up to the next one under
// the assumption that the rest of the tree will be padded out (e.g. with zeros)
func Fr32PaddedSizeToV1TreeHeight(size uint64) uint8 {
	if size <= 32 {
		return 0
	}

	// Calculate the floor of log2(size)
	b := 63 - bits.LeadingZeros64(size)
	// Leaf size is 32 == 2^5
	b -= 5

	// Check if the size is a power of 2 and if not then add one since the tree will need to be padded out
	if 32<<b < size {
		b++
	}
	return uint8(b)
}

// UnpaddedSizeToV1TreeHeight calculates the height of the piece tree given the data that's meant to be encoded in the
// tree before any FR32 padding is applied. Because pieces are only defined on binary trees of FR32 encoded data if the
// size is not a power of 2 after the FR32 padding is applied it will be rounded up to the next one under the assumption
// that the rest of the tree will be padded out (e.g. with zeros)
func UnpaddedSizeToV1TreeHeight(size uint64) (uint8, error) {
	if size*128 < size {
		return 0, fmt.Errorf("unsupported size: too big")
	}

	paddedSize := size * 128 / 127
	if paddedSize*127 != size*128 {
		paddedSize++
	}

	return Fr32PaddedSizeToV1TreeHeight(paddedSize), nil
}

// UnpaddedSizeToV1TreeHeightAndPadding calculates the height of the piece tree given the data that's meant to be
// encoded in the tree before any FR32 padding is applied. Because pieces are only defined on binary trees of FR32
// encoded data if the size is not a power of 2 after the FR32 padding is applied it will be rounded up to the next one
// under the assumption that the rest of the tree will be padded out (e.g. with zeros). The amount of data padding that
// is needed to be applied is returned alongside the tree height.
func UnpaddedSizeToV1TreeHeightAndPadding(dataSize uint64) (uint8, uint64, error) {
	if dataSize*128 < dataSize {
		return 0, 0, fmt.Errorf("unsupported size: too big")
	}

	if dataSize < 127 {
		return 0, 0, fmt.Errorf("unsupported size: too small")
	}

	fr32DataSize := dataSize * 128 / 127
	// If the FR32 padding doesn't fill an exact number of bytes add up to 1 more byte of zeros to round it out
	if fr32DataSize*127 != dataSize*128 {
		fr32DataSize++
	}

	treeHeight := Fr32PaddedSizeToV1TreeHeight(fr32DataSize)
	paddedFr32DataSize := HeightToPaddedSize(treeHeight)
	paddedDataSize := paddedFr32DataSize / 128 * 127
	padding := paddedDataSize - dataSize

	return treeHeight, padding, nil
}

func HeightToPaddedSize(height uint8) uint64 {
	return uint64(32) << height
}
