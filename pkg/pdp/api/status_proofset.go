package api

import (
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/labstack/echo/v4"
)

type GetProofSetCreationStatusResponse struct {
	CreateMessageHash string `json:"createMessageHash"`
	ProofsetCreated   bool   `json:"proofsetCreated"`
	Service           string `json:"service"`
	TxStatus          string `json:"txStatus"`
	OK                bool   `json:"ok"`
	ProofSetId        int64  `json:"proofSetId,omitempty"`
}

// echoHandleGetProofSetCreationStatus -> GET /pdp/proof-sets/created/:txHash
func (p *PDP) handleGetProofSetCreationStatus(c echo.Context) error {
	ctx := c.Request().Context()
	txHash := c.Param("txHash")

	// Clean txHash (ensure it starts with '0x' and is lowercase)
	if !strings.HasPrefix(txHash, "0x") {
		txHash = "0x" + txHash
	}
	txHash = strings.ToLower(txHash)

	// Validate txHash is a valid hash
	if len(txHash) != 66 { // '0x' + 64 hex chars
		return c.String(http.StatusBadRequest, "Invalid txHash length")
	}
	if _, err := hex.DecodeString(txHash[2:]); err != nil {
		return c.String(http.StatusBadRequest, "Invalid txHash format")
	}
	txh := common.HexToHash(txHash)

	status, err := p.Service.ProofSetStatus(ctx, txh)
	if err != nil {
		log.Errorw("failed to get status proof set creation", "error", err)
		return c.String(http.StatusInternalServerError, "Failed to get proof set status")
	}

	resp := GetProofSetCreationStatusResponse{
		CreateMessageHash: status.CreateMessageHash,
		ProofsetCreated:   status.ProofsetCreated,
		Service:           status.Service,
		TxStatus:          status.TxStatus,
		OK:                status.OK,
		ProofSetId:        status.ProofSetId,
	}
	return c.JSON(http.StatusOK, resp)

}
