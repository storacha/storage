package api

import (
	"net/http"
	"time"

	"github.com/filecoin-project/go-state-types/abi"
	"github.com/ipfs/go-cid"
	"github.com/labstack/echo/v4"

	"github.com/storacha/piri/pkg/pdp/api/middleware"
	"github.com/storacha/piri/pkg/pdp/proof"
	"github.com/storacha/piri/pkg/pdp/service"
	"github.com/storacha/piri/pkg/pdp/service/types"
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

// handlePreparePiece -> POST /pdp/piece
func (p *PDP) handlePreparePiece(c echo.Context) error {
	ctx := c.Request().Context()
	operation := "PreparePiece"

	var req PreparePieceRequest
	if err := c.Bind(&req); err != nil {
		return middleware.NewError(operation, "Invalid request body", err, http.StatusBadRequest)
	}

	if abi.UnpaddedPieceSize(req.Check.Size) > PieceSizeLimit {
		return middleware.NewError(operation, "Piece size exceeds maximum allowed size", nil, http.StatusBadRequest).
			WithContext("allowed size", PieceSizeLimit).
			WithContext("requested size", req.Check.Size)
	}

	log.Debugw("Processing prepare piece request",
		"name", req.Check,
		"hash", req.Check.Hash,
		"size", req.Check.Size)
	start := time.Now()
	res, err := p.Service.PreparePiece(ctx, service.PiecePrepareRequest{
		Check: types.PieceHash{
			Name: req.Check.Name,
			Hash: req.Check.Hash,
			Size: req.Check.Size,
		},
		Notify: req.Notify,
	})
	if err != nil {
		return middleware.NewError(operation, "Failed to prepare piece", err, http.StatusInternalServerError)
	}

	resp := PreparePieceResponse{
		Location: res.Location,
		Created:  res.Created,
	}
	if res.PieceCID != cid.Undef {
		resp.PieceCID = res.PieceCID.String()
	}
	log.Infow("Successfully prepared piece",
		"location", resp.Location,
		"created", resp.Created,
		"duration", time.Since(start))
	if res.Created {
		c.Response().Header().Set(echo.HeaderLocation, res.Location)
		return c.JSON(http.StatusCreated, resp)
	}
	return c.JSON(http.StatusNoContent, resp)
}
