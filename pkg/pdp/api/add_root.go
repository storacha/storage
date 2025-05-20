package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/storacha/storage/pkg/pdp/api/middleware"
	"github.com/storacha/storage/pkg/pdp/service"
)

type AddRootSubrootEntry struct {
	SubrootCid string `json:"subrootCid"`
}
type AddRootRequest struct {
	RootCid  string                `json:"rootCid"`
	Subroots []AddRootSubrootEntry `json:"subroots"`
}

type AddRootsPayload struct {
	Roots     []AddRootRequest `json:"roots"`
	ExtraData *string          `json:"extraData,omitempty"`
}
type AddRootToProofSetRequest []AddRootRequest

func (p *PDP) handleAddRootToProofSet(c echo.Context) error {
	ctx := c.Request().Context()
	operation := "AddRootToProofSet"

	proofSetIDStr := c.Param("proofSetID")
	if proofSetIDStr == "" {
		return middleware.NewError(operation, "missing proofSetID", nil, http.StatusBadRequest)
	}

	id, err := strconv.ParseInt(proofSetIDStr, 10, 64)
	if err != nil {
		return middleware.NewError(operation, "invalid proofSetID format", err, http.StatusBadRequest).
			WithContext("proofSetID", proofSetIDStr)
	}

	var req AddRootsPayload
	if err := c.Bind(&req); err != nil {
		return middleware.NewError(operation, "failed to parse request body", err, http.StatusBadRequest).
			WithContext("proofSetID", id)
	}

	if len(req.Roots) == 0 {
		return middleware.NewError(operation, "no roots provided", nil, http.StatusBadRequest).
			WithContext("proofSetID", id)
	}

	t := make([]service.AddRootRequest, 0, len(req.Roots))

	for _, r := range req.Roots {
		subroots := make([]string, 0, len(r.Subroots))
		for _, s := range r.Subroots {
			subroots = append(subroots, s.SubrootCid)
		}
		t = append(t, service.AddRootRequest{
			RootCID:     r.RootCid,
			SubrootCIDs: subroots,
		})
	}

	log.Debugw("Processing add root request",
		"proofSetID", id,
		"rootCount", len(req.Roots))

	start := time.Now()
	if _, err := p.Service.ProofSetAddRoot(ctx, id, t); err != nil {
		return middleware.NewError(operation, "failed to add root to proofSet", err, http.StatusInternalServerError).
			WithContext("proofSetID", id).
			WithContext("rootCount", len(req.Roots)).
			WithContext("clientIP", c.RealIP())
	}

	log.Infow("Successfully added roots to proofSet",
		"proofSetID", id,
		"rootCount", len(req.Roots),
		"duration", time.Since(start))
	return c.NoContent(http.StatusCreated)
}
