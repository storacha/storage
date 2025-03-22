package api

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type PieceUploadRequest struct {
	// Add fields for the upload
	Data []byte `json:"data"`
}
type PieceUploadResponse struct {
	UploadUUID string `json:"uploadUUID"`
	Status     string `json:"status"`
}

func (p *PDP) handlePieceUpload(c echo.Context) error {
	ctx := c.Request().Context()
	uploadUUID := c.Param("uploadUUID")

	if uploadUUID == "" {
		return c.String(http.StatusBadRequest, "uploadUUID is required")
	}

	uploadID, err := uuid.Parse(uploadUUID)
	if err != nil {
		return c.String(http.StatusBadRequest, "uploadUUID is invalid")
	}

	var req PieceUploadRequest
	if err := c.Bind(&req); err != nil {
		return c.String(http.StatusBadRequest, "Invalid request")
	}

	if _, err := p.Service.UploadPiece(ctx, uploadID, req.Data); err != nil {
		return c.String(http.StatusBadRequest, "Failed to upload piece")
	}

	return c.NoContent(http.StatusNoContent)
}
