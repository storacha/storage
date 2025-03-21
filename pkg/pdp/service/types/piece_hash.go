package types

import (
	"encoding/hex"
	"fmt"

	"github.com/filecoin-project/go-state-types/abi"
	"github.com/ipfs/go-cid"
	"github.com/multiformats/go-multihash"
)

// PieceSizeLimit in bytes
var PieceSizeLimit = abi.PaddedPieceSize(256 << 20).Unpadded()

type PieceHash struct {
	// Name of the hash function used
	// sha2-256-trunc254-padded - CommP
	// sha2-256 - Blob sha256
	Name string `json:"name"`

	// hex encoded hash
	Hash string `json:"hash"`

	// Size of the piece in bytes
	Size int64 `json:"size"`
}

func (ph *PieceHash) Set() bool {
	return ph.Name != "" && ph.Hash != "" && ph.Size > 0
}

func (ph *PieceHash) Multihash() (multihash.Multihash, error) {
	_, ok := multihash.Names[ph.Name]
	if !ok {
		return nil, fmt.Errorf("hash function name not recognized: %s", ph.Name)
	}

	hashBytes, err := hex.DecodeString(ph.Hash)
	if err != nil {
		return nil, fmt.Errorf("failed to decode hash: %w", err)
	}

	return multihash.EncodeName(hashBytes, ph.Name)
}

func (ph *PieceHash) CommP() (cid.Cid, bool, error) {
	// commp, known, error
	mh, err := ph.Multihash()
	if err != nil {
		return cid.Undef, false, fmt.Errorf("failed to decode hash: %w", err)
	}

	if ph.Name == multihash.Codes[multihash.SHA2_256_TRUNC254_PADDED] {
		return cid.NewCidV1(cid.FilCommitmentUnsealed, mh), true, nil
	}

	return cid.Undef, false, fmt.Errorf("hash function not recognized: %s", ph.Name)

	/*
		var commpStr string
		err = db.QueryRow(ctx, `
			SELECT commp FROM pdp_piece_mh_to_commp WHERE mhash = $1 AND size = $2
		`, mh, ph.Size).Scan(&commpStr)
		if err != nil {
			if err == pgx.ErrNoRows {
				return cid.Undef, false, nil
			}
			return cid.Undef, false, fmt.Errorf("failed to query pdp_piece_mh_to_commp: %w", err)
		}

		commpCid, err := cid.Parse(commpStr)
		if err != nil {
			return cid.Undef, false, fmt.Errorf("failed to parse commp CID: %w", err)
		}

		return commpCid, true, nil

	*/
}

func (ph *PieceHash) maybeStaticCommp() (cid.Cid, bool) {
	mh, err := ph.Multihash()
	if err != nil {
		return cid.Undef, false
	}

	if ph.Name == multihash.Codes[multihash.SHA2_256_TRUNC254_PADDED] {
		return cid.NewCidV1(cid.FilCommitmentUnsealed, mh), true
	}

	return cid.Undef, false
}
