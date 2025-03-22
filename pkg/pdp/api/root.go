package api

import (
	logging "github.com/ipfs/go-log/v2"
	"github.com/labstack/echo/v4"

	"github.com/storacha/storage/pkg/pdp/service"
)

var log = logging.Logger("pdp/api")

const ()

func RegisterEchoRoutes(e *echo.Echo, p *PDP) {
	// /pdp/proof-sets
	proofSets := e.Group("/pdp/proof-sets")
	proofSets.POST("", p.handleCreateProofSet)
	proofSets.GET("/created/:txHash", p.handleGetProofSetCreationStatus)

	// /pdp/proof-sets/:proofSetID
	proofSets.GET("/:proofSetID", p.handleGetProofSet)
	proofSets.DELETE("/:proofSetID", p.handleDeleteProofSet)

	// /pdp/proof-sets/:proofSetID/roots
	roots := proofSets.Group("/:proofSetID/roots")
	roots.POST("", p.handleAddRootToProofSet)
	roots.GET("/:rootID", p.handleGetProofSetRoot)
	roots.DELETE("/:rootID", p.handleDeleteProofSetRoot)

	// /pdp/ping
	e.GET("/pdp/ping", p.handlePing)

	// /pdp/piece
	e.POST("/pdp/piece", p.handlePiecePost)
	e.GET("/pdp/piece/", p.handleFindPiece)
	e.PUT("/pdp/piece/upload/:uploadUUID", p.handlePieceUpload)
}

type PDP struct {
	Service service.PDPService
}
