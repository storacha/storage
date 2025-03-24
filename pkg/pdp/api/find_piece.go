package api

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

type FindPieceResponse struct {
	PieceCID string `json:"piece_cid"`
}

func (p *PDP) handleFindPiece(c echo.Context) error {
	ctx := c.Request().Context()

	sizeStr := c.QueryParam("size")
	if sizeStr == "" {
		return c.String(http.StatusBadRequest, "size is required")
	}
	name := c.QueryParam("name")
	if name == "" {
		return c.String(http.StatusBadRequest, "name is required")
	}
	hash := c.QueryParam("hash")
	if hash == "" {
		return c.String(http.StatusBadRequest, "hash is required")
	}

	size, err := strconv.ParseInt(sizeStr, 10, 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "size is invalid")
	}

	// Verify that a 'parked_pieces' entry exists for the given 'piece_cid'
	pieceCID, has, err := p.Service.FindPiece(ctx, name, hash, size)
	if err != nil {
		return c.String(http.StatusInternalServerError, fmt.Sprint("failed to find piece in database"))
	}
	if !has {
		return c.String(http.StatusNotFound, "piece not found")
	}

	resp := FindPieceResponse{
		PieceCID: pieceCID.String(),
	}

	return c.JSON(http.StatusOK, resp)
}
