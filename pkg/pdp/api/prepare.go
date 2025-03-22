package api

import (
	"net/http"

	"github.com/filecoin-project/go-state-types/abi"
	"github.com/ipfs/go-cid"
	"github.com/labstack/echo/v4"

	"github.com/storacha/storage/pkg/pdp/proof"
	"github.com/storacha/storage/pkg/pdp/service"
	"github.com/storacha/storage/pkg/pdp/service/types"
)

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

type PreparePieceRequest struct {
	Check  PieceHash `json:"check"`
	Notify string    `json:"notify,omitempty"`
}
type PreparePieceResponse struct {
	Location string `json:"location,omitempty"`
	PieceCID string `json:"piece_cid,omitempty"`
	Created  bool   `json:"created,omitempty"`
}

var PieceSizeLimit = abi.PaddedPieceSize(proof.MaxMemtreeSize).Unpadded()

// handlePiecePost -> POST /pdp/piece
func (p *PDP) handlePiecePost(c echo.Context) error {
	ctx := c.Request().Context()
	var req PreparePieceRequest
	if err := c.Bind(&req); err != nil {
		return c.String(http.StatusBadRequest, "Invalid request")
	}

	if abi.UnpaddedPieceSize(req.Check.Size) > PieceSizeLimit {
		return c.String(http.StatusBadRequest, "Piece size exceeds the maximum allowed size")
	}

	res, err := p.Service.PreparePiece(ctx, service.PiecePrepareRequest{
		Check: types.PieceHash{
			Name: req.Check.Name,
			Hash: req.Check.Hash,
			Size: req.Check.Size,
		},
		Notify: req.Notify,
	})
	if err != nil {
		return c.String(http.StatusInternalServerError, "failed to prepare piece")
	}

	resp := PreparePieceResponse{
		Location: res.Location,
		Created:  res.Created,
	}
	if res.PieceCID != cid.Undef {
		resp.PieceCID = res.PieceCID.String()
	}
	return c.JSON(http.StatusOK, resp)
}
