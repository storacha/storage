package piece

import (
	"encoding/json"
	"errors"

	commcid "github.com/filecoin-project/go-fil-commcid"
	"github.com/ipfs/go-cid"
	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/datamodel"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/storacha/storage/pkg/piece/digest"
	"github.com/storacha/storage/pkg/piece/size"
)

var ErrWrongLinkType = errors.New("must be a cid link")
var ErrWrongCodec = errors.New("must be raw codec")

type PieceLink interface {
	PaddedSize() uint64
	Padding() uint64
	Height() uint8
	DataCommitment() []byte
	Link() ipld.Link
	V1Link() ipld.Link
	json.Marshaler
}

type pieceLink cidlink.Link

// DataCommitment implements PieceLink.
func (p pieceLink) DataCommitment() []byte {
	dc, _ := digest.DataCommitment(p.Cid.Hash())
	return dc
}

// Height implements PieceLink.
func (p pieceLink) Height() uint8 {
	h, _ := digest.Height(p.Cid.Hash())
	return h
}

// Link implements PieceLink.
func (p pieceLink) Link() datamodel.Link {
	return cidlink.Link(p)
}

// PaddedSize implements PieceLink.
func (p pieceLink) PaddedSize() uint64 {
	return size.HeightToPaddedSize(p.Height())
}

// Padding implements PieceLink.
func (p pieceLink) Padding() uint64 {
	pd, _ := digest.Padding(p.Cid.Hash())
	return pd
}

// V1Link implements PieceLink.
func (p pieceLink) V1Link() datamodel.Link {
	dc := p.DataCommitment()
	v1, _ := commcid.DataCommitmentV1ToCID(dc)
	return cidlink.Link{Cid: v1}
}

func FromPieceDigest(pd digest.PieceDigest) PieceLink {
	return pieceLink(cidlink.Link{Cid: cid.NewCidV1(cid.Raw, pd.Bytes())})
}

func FromLink(link ipld.Link) (PieceLink, error) {
	cl, ok := link.(cidlink.Link)
	if !ok {
		return nil, ErrWrongLinkType
	}
	if cl.Cid.Prefix().Codec != cid.Raw {
		return nil, ErrWrongCodec
	}

	pieceDigest, err := digest.NewPieceDigest(cl.Cid.Hash())
	if err != nil {
		return nil, err
	}
	return FromPieceDigest(pieceDigest), nil
}

func FromV1LinkAndSize(link datamodel.Link, unpaddedDataSize uint64) (PieceLink, error) {
	cl, ok := link.(cidlink.Link)
	if !ok {
		return nil, ErrWrongLinkType
	}
	if cl.Cid.Prefix().Codec != cid.Raw {
		return nil, ErrWrongCodec
	}
	commitment, err := commcid.CIDToDataCommitmentV1(cl.Cid)
	if err != nil {
		return nil, err
	}
	pieceDigest, err := digest.FromCommitmentAndSize(commitment, unpaddedDataSize)
	if err != nil {
		return nil, err
	}
	return FromPieceDigest(pieceDigest), nil
}
