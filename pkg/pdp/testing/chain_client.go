package testing

import (
	"context"
	crand "crypto/rand"
	"sync"
	"testing"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/chain/store"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/ipfs/go-cid"
	mh "github.com/multiformats/go-multihash"
	"github.com/stretchr/testify/require"
)

type FakeChainClient struct {
	currentMu     sync.Mutex
	currentHeight abi.ChainEpoch
	currentTipSet *types.TipSet
	notifyChans   []chan []*api.HeadChange

	miner address.Address
	t     testing.TB
}

func RandomBytes(t testing.TB, size int) []byte {
	bytes := make([]byte, size)
	_, err := crand.Read(bytes)
	require.NoError(t, err)
	return bytes
}

func RandomCID(t testing.TB) cid.Cid {
	bytes := RandomBytes(t, 10)
	c, err := cid.Prefix{
		Version:  1,
		Codec:    cid.Raw,
		MhType:   mh.SHA2_256,
		MhLength: -1,
	}.Sum(bytes)
	require.NoError(t, err)
	return c
}

func NewFakeChainClient(t testing.TB) *FakeChainClient {
	// Create a fake tipset at height 1
	miner, err := address.NewIDAddress(1)
	require.NoError(t, err)

	ts, err := types.NewTipSet([]*types.BlockHeader{
		{
			Height:                1,
			Miner:                 miner,
			Parents:               []cid.Cid{RandomCID(t)},
			ParentStateRoot:       RandomCID(t),
			ParentMessageReceipts: RandomCID(t),
			Messages:              RandomCID(t),
		},
	})
	require.NoError(t, err)

	return &FakeChainClient{
		currentHeight: 0,
		currentTipSet: ts,
		notifyChans:   make([]chan []*api.HeadChange, 0),
		miner:         miner,
		t:             t,
	}
}

func (c *FakeChainClient) CurrentHeight() abi.ChainEpoch {
	c.currentMu.Lock()
	defer c.currentMu.Unlock()
	return c.currentHeight
}

func (c *FakeChainClient) ChainHead(ctx context.Context) (*types.TipSet, error) {
	c.currentMu.Lock()
	defer c.currentMu.Unlock()

	return c.currentTipSet, nil
}

func (c *FakeChainClient) ChainNotify(ctx context.Context) (<-chan []*api.HeadChange, error) {
	c.currentMu.Lock()
	defer c.currentMu.Unlock()

	// Create notification channel
	ch := make(chan []*api.HeadChange, 16)
	c.notifyChans = append(c.notifyChans, ch)

	// Send current head as the first notification
	// HCCurrent is always the first notification in the "real" implementation
	// Gripe: how the F^%! Filecoin existed for so long with a testing harness that solves this is beyond me.
	ch <- []*api.HeadChange{
		{
			Type: store.HCCurrent,
			Val:  c.currentTipSet,
		},
	}

	return ch, nil
}

func (c *FakeChainClient) StateGetRandomnessDigestFromBeacon(ctx context.Context, randEpoch abi.ChainEpoch, tsk types.TipSetKey) (abi.Randomness, error) {
	randBytes := make([]byte, 32)

	// Use epoch value to influence the randomness
	epoch := uint64(randEpoch)
	for i := 0; i < 8 && i < len(randBytes); i++ {
		randBytes[i] = byte((epoch >> (i * 8)) & 0xff)
	}

	return randBytes, nil
}

func (c *FakeChainClient) AdvanceChain() abi.ChainEpoch {
	return c.AdvanceByHeight(1)
}

// AdvanceHeight advances the chain by the specified number of epochs
func (c *FakeChainClient) AdvanceByHeight(epochs int64) abi.ChainEpoch {
	c.currentMu.Lock()
	defer c.currentMu.Unlock()

	newHeight := c.currentHeight + abi.ChainEpoch(epochs)

	// Create a new tipset at the new height
	newTs, err := types.NewTipSet([]*types.BlockHeader{
		{
			Height: newHeight,
			Miner:  c.miner,
			// Add necessary parent info
			Parents:               c.currentTipSet.Key().Cids(),
			ParentStateRoot:       RandomCID(c.t),
			ParentMessageReceipts: RandomCID(c.t),
			Messages:              RandomCID(c.t),
		},
	})
	require.NoError(c.t, err)

	// Update the current height and tipset
	c.currentHeight = newHeight
	c.currentTipSet = newTs

	// Notify all registered channels
	change := []*api.HeadChange{
		{
			Type: store.HCApply,
			Val:  newTs,
		},
	}

	for _, ch := range c.notifyChans {
		select {
		case ch <- change:
			// Successfully sent notification
		default:
			// Channel is full, continue without blocking
		}
	}
	return newHeight
}
