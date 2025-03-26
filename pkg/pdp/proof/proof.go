package proof

import (
	"encoding/hex"
	"fmt"
	"io"

	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/lotus/storage/sealer/fr32"
	logger "github.com/ipfs/go-log/v2"
	pool "github.com/libp2p/go-buffer-pool"
	"github.com/minio/sha256-simd"
)

const NODE_SIZE = 32

var log = logger.Logger("proof")

type RawMerkleProof struct {
	Leaf  [32]byte
	Proof [][32]byte
	Root  [32]byte
}

// MemtreeProof generates a Merkle proof for the given leaf index from the memtree.
// The memtree is a byte slice containing all the nodes of the Merkle tree, including leaves and internal nodes.
func MemtreeProof(memtree []byte, leafIndex int64) (*RawMerkleProof, error) {
	// Currently, the implementation supports only binary trees (arity == 2)
	const arity = 2

	// Calculate the total number of nodes in the memtree
	totalNodes := int64(len(memtree)) / NODE_SIZE

	// Reconstruct level sizes from the total number of nodes
	// Starting from the number of leaves, compute the number of nodes at each level
	nLeaves := (totalNodes + 1) / 2

	currLevelCount := nLeaves
	levelSizes := []int64{}
	totalNodesCheck := int64(0)

	for {
		levelSizes = append(levelSizes, currLevelCount)
		totalNodesCheck += currLevelCount

		if currLevelCount == 1 {
			break
		}
		// Compute the number of nodes in the next level
		currLevelCount = (currLevelCount + int64(arity) - 1) / int64(arity)
	}

	// Verify that the reconstructed total nodes match the actual total nodes
	if totalNodesCheck != totalNodes {
		return nil, fmt.Errorf("invalid memtree size; reconstructed total nodes do not match")
	}

	// Compute the starting byte offset for each level in memtree
	levelStarts := make([]int64, len(levelSizes))
	var offset int64 = 0
	for i, size := range levelSizes {
		levelStarts[i] = offset
		offset += size * NODE_SIZE
	}

	// Validate the leaf index
	if leafIndex < 0 || leafIndex >= levelSizes[0] {
		return nil, fmt.Errorf("invalid leaf index %d for %d leaves", leafIndex, levelSizes[0])
	}

	// Initialize the proof structure
	proof := &RawMerkleProof{
		Proof: make([][NODE_SIZE]byte, 0, len(levelSizes)-1),
	}

	// Extract the leaf hash from the memtree
	leafOffset := levelStarts[0] + leafIndex*NODE_SIZE
	copy(proof.Leaf[:], memtree[leafOffset:leafOffset+NODE_SIZE])

	// Build the proof by collecting sibling hashes at each level
	index := leafIndex
	for level := 0; level < len(levelSizes)-1; level++ {
		siblingIndex := index ^ 1 // Toggle the last bit to get the sibling index

		siblingOffset := levelStarts[level] + siblingIndex*NODE_SIZE
		nodeOffset := levelStarts[level] + index*NODE_SIZE
		var siblingHash [NODE_SIZE]byte
		var node [NODE_SIZE]byte
		copy(node[:], memtree[nodeOffset:nodeOffset+NODE_SIZE])
		copy(siblingHash[:], memtree[siblingOffset:siblingOffset+NODE_SIZE])
		proof.Proof = append(proof.Proof, siblingHash)

		if index < siblingIndex { // left
			log.Debugw("Proof", "position", index, "left-c", hex.EncodeToString(node[:]), "right-s", hex.EncodeToString(siblingHash[:]), "ouh", hex.EncodeToString(shabytes(append(node[:], siblingHash[:]...))[:]))
		} else { // right
			log.Debugw("Proof", "position", index, "left-s", hex.EncodeToString(siblingHash[:]), "right-c", hex.EncodeToString(node[:]), "ouh", hex.EncodeToString(shabytes(append(siblingHash[:], node[:]...))[:]))
		}
		// Move up to the parent index
		index /= int64(arity)
	}

	// Extract the root hash from the memtree
	rootOffset := levelStarts[len(levelSizes)-1]
	copy(proof.Root[:], memtree[rootOffset:rootOffset+NODE_SIZE])

	return proof, nil
}

func shabytes(in []byte) []byte {
	out := sha256.Sum256(in)
	return out[:]
}

const MaxMemtreeSize = 256 << 20

// BuildSha254Memtree builds a sha256 memtree from the input data
// Returned slice should be released to the pool after use
func BuildSha254Memtree(rawIn io.Reader, size abi.UnpaddedPieceSize) ([]byte, error) {
	if size.Padded() > MaxMemtreeSize {
		return nil, fmt.Errorf("piece too large for memtree: %d", size)
	}

	unpadBuf := pool.Get(int(size))
	// read into unpadBuf
	_, err := io.ReadFull(rawIn, unpadBuf)
	if err != nil {
		pool.Put(unpadBuf)
		return nil, fmt.Errorf("failed to read into unpadBuf: %w", err)
	}

	nLeaves := int64(size.Padded()) / NODE_SIZE
	totalNodes, levelSizes := computeTotalNodes(nLeaves, 2)
	memtreeBuf := pool.Get(int(totalNodes * NODE_SIZE))

	fr32.Pad(unpadBuf, memtreeBuf[:size.Padded()])
	pool.Put(unpadBuf)

	d := sha256.New()

	levelStarts := make([]int64, len(levelSizes))
	levelStarts[0] = 0
	for i := 1; i < len(levelSizes); i++ {
		levelStarts[i] = levelStarts[i-1] + levelSizes[i-1]*NODE_SIZE
	}

	for level := 1; level < len(levelSizes); level++ {
		levelNodes := levelSizes[level]
		prevLevelStart := levelStarts[level-1]
		currLevelStart := levelStarts[level]

		for i := int64(0); i < levelNodes; i++ {
			leftOffset := prevLevelStart + (2*i)*NODE_SIZE

			d.Reset()
			d.Write(memtreeBuf[leftOffset : leftOffset+(NODE_SIZE*2)])

			outOffset := currLevelStart + i*NODE_SIZE
			// sum calls append, so we give it a zero len slice at the correct offset
			d.Sum(memtreeBuf[outOffset:outOffset])

			// set top bits to 00
			memtreeBuf[outOffset+NODE_SIZE-1] &= 0x3F
		}
	}

	return memtreeBuf, nil
}

func ComputeBinShaParent(left, right [NODE_SIZE]byte) [NODE_SIZE]byte {
	out := sha256.Sum256(append(left[:], right[:]...))
	out[NODE_SIZE-1] &= 0x3F
	return out
}

func computeTotalNodes(nLeaves, arity int64) (int64, []int64) {
	totalNodes := int64(0)
	levelCounts := []int64{}
	currLevelCount := nLeaves
	for currLevelCount > 0 {
		levelCounts = append(levelCounts, currLevelCount)
		totalNodes += currLevelCount
		if currLevelCount == 1 {
			break
		}
		currLevelCount = (currLevelCount + arity - 1) / arity
	}
	return totalNodes, levelCounts
}

func NodeLevel(leaves, arity int64) int {
	if leaves == 0 {
		return 0
	}
	level := 0
	for leaves > 1 {
		leaves = (leaves + arity - 1) / arity
		level++
	}
	return level + 1
}
