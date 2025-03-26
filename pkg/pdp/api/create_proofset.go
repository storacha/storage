package api

import (
	"net/http"
	"path"

	"github.com/ethereum/go-ethereum/common"
	"github.com/labstack/echo/v4"
)

type CreateProofSetRequest struct {
	RecordKeeper string `json:"recordKeeper"`
}

// CreateProofSetResponse is the JSON output for the CreateProofSet endpoint.
type CreateProofSetResponse struct {
	TxHash   string `json:"txHash"`
	Location string `json:"location"`
}

// echoHandleCreateProofSet -> POST /pdp/proof-sets
func (p *PDP) handleCreateProofSet(c echo.Context) error {
	ctx := c.Request().Context()

	var req CreateProofSetRequest
	if err := c.Bind(&req); err != nil {
		return c.String(http.StatusBadRequest, "Invalid JSON: "+err.Error())
	}
	if req.RecordKeeper == "" {
		return c.String(http.StatusBadRequest, "recordKeeper address is required")
	}
	recordKeeperAddr := common.HexToAddress(req.RecordKeeper)
	if recordKeeperAddr == (common.Address{}) {
		return c.String(http.StatusBadRequest, "Invalid recordKeeper address")
	}

	txHash, err := p.Service.ProofSetCreate(ctx, recordKeeperAddr)
	if err != nil {
		log.Errorw("failed to create proof set", "error", err)
		return c.String(http.StatusInternalServerError, "Failed to create proof set")
	}

	location := path.Join("/pdp/proof-sets/created", txHash.Hex())
	c.Response().Header().Set("Location", location)

	resp := CreateProofSetResponse{
		TxHash:   txHash.Hex(),
		Location: location,
	}
	return c.JSON(http.StatusCreated, resp)
}
