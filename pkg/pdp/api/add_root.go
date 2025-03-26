package api

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"github.com/storacha/storage/pkg/pdp/service"
)

type AddRootSubrootEntry struct {
	SubrootCid string `json:"subrootCid"`
}
type AddRootRequest struct {
	RootCid  string                `json:"rootCid"`
	Subroots []AddRootSubrootEntry `json:"subroots"`
}
type AddRootToProofSetRequest []AddRootRequest

func (p *PDP) handleAddRootToProofSet(c echo.Context) error {
	ctx := c.Request().Context()
	proofSetIDStr := c.Param("proofSetID")
	if proofSetIDStr == "" {
		return c.String(http.StatusBadRequest, "missing proofSetID")
	}

	id, err := strconv.ParseInt(proofSetIDStr, 10, 64) // ignoring error for brevity
	if err != nil {
		return c.String(http.StatusBadRequest, "invalid proofSetID")
	}

	var req AddRootToProofSetRequest
	if err := c.Bind(&req); err != nil {
		return c.String(http.StatusBadRequest, "Invalid request")
	}

	t := make([]service.AddRootRequest, 0, len(req))

	for _, r := range req {
		subroots := make([]string, 0, len(r.Subroots))
		for _, s := range r.Subroots {
			subroots = append(subroots, s.SubrootCid)
		}
		t = append(t, service.AddRootRequest{
			RootCID:     r.RootCid,
			SubrootCIDs: subroots,
		})
	}

	if _, err := p.Service.ProofSetAddRoot(ctx, id, t); err != nil {
		return c.String(http.StatusInternalServerError, "failed to add root to proofSet")
	}

	return c.NoContent(http.StatusCreated)
}
