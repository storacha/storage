package api

import (
	"path"

	logging "github.com/ipfs/go-log/v2"
	"github.com/labstack/echo/v4"

	"github.com/storacha/storage/pkg/pdp/service"
)

var log = logging.Logger("pdp/api")

const (
	PDPRoutePath     = "/pdp"
	PRoofSetRoutPath = "/proof-sets"
)

func RegisterEchoRoutes(e *echo.Echo, p *PDP) {
	// /pdp/proof-sets
	proofSets := e.Group(path.Join(PDPRoutePath, PRoofSetRoutPath))
	proofSets.POST("", p.handleCreateProofSet)
	proofSets.GET("/created/:txHash", p.handleGetProofSetCreationStatus)

	// /pdp/proof-sets/:proofSetID
	proofSets.GET("/:proofSetID", p.handleGetProofSet)
	proofSets.DELETE("/:proofSetID", p.handleDeleteProofSet)

	// /pdp/proof-sets/:proofSetID/roots
	roots := proofSets.Group("/:proofSetID/roots")
	roots.POST("", p.handleAddRootToProofSet)
	roots.GET("/:rootID", p.handleGetProofSetRoot)
	roots.DELETE("/:rootID", p.handleDeleteRootFromProofSet)

	// /pdp/ping
	e.GET("/pdp/ping", p.handlePing)

	// /pdp/piece
	e.POST(path.Join(PDPRoutePath, "/piece"), p.handlePreparePiece)
	e.PUT(path.Join(PDPRoutePath, "/piece/upload/:uploadUUID"), p.handlePieceUpload)
	e.GET(path.Join(PDPRoutePath, "/piece"), p.handleFindPiece)
}

type PDP struct {
	Service *service.PDPService
}
