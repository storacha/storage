package api

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

type GetProofSetResponse struct {
	ID                 int64       `json:"id"`
	NextChallengeEpoch int64       `json:"nextChallengeEpoch"`
	Roots              []RootEntry `json:"roots"`
}

type RootEntry struct {
	RootID        int64  `json:"rootId"`
	RootCID       string `json:"rootCid"`
	SubrootCID    string `json:"subrootCid"`
	SubrootOffset int64  `json:"subrootOffset"`
}

// handleGetProofSet -> GET /pdp/proof-sets/:proofSetID
func (p *PDP) handleGetProofSet(c echo.Context) error {
	ctx := c.Request().Context()
	proofSetIDStr := c.Param("proofSetID")

	if proofSetIDStr == "" {
		return c.String(http.StatusBadRequest, "missing proofSetID")
	}

	id, err := strconv.ParseInt(proofSetIDStr, 10, 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "invalid proofSetID")
	}

	ps, err := p.Service.ProofSet(ctx, id)
	if err != nil {
		return c.String(http.StatusInternalServerError, "failed to fetch proofSet")
	}

	resp := GetProofSetResponse{
		ID:                 ps.ID,
		NextChallengeEpoch: ps.NextChallengeEpoch,
	}
	for _, root := range ps.Roots {
		resp.Roots = append(resp.Roots, RootEntry{
			RootID:        int64(root.RootID),
			RootCID:       root.RootCID,
			SubrootCID:    root.SubrootCID,
			SubrootOffset: root.SubrootOffset,
		})
	}
	return c.JSON(http.StatusOK, resp)
}
