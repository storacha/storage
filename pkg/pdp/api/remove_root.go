package api

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

func (p *PDP) handleDeleteRootFromProofSet(c echo.Context) error {
	ctx := c.Request().Context()
	// Step 2: Extract parameters from the URL
	proofSetIdStr := c.Param("proofSetID")
	if proofSetIdStr == "" {
		return c.String(http.StatusBadRequest, "missing proofSetID")
	}
	rootIdStr := c.Param("rootID")
	if rootIdStr == "" {
		return c.String(http.StatusBadRequest, "missing rootID")
	}

	proofSetID, err := strconv.ParseUint(proofSetIdStr, 10, 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "invalid proofSetID")
	}
	rootID, err := strconv.ParseUint(rootIdStr, 10, 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "invalid rootID")
	}

	// check if the proofset belongs to the service in pdp_proof_sets
	if err := p.Service.RemoveRoot(ctx, proofSetID, rootID); err != nil {
		return c.String(http.StatusInternalServerError, "failed to remove root")
	}
	return c.NoContent(http.StatusNoContent)
}
