package storage

import (
	logging "github.com/ipfs/go-log/v2"
	"github.com/storacha/go-ucanto/server"
	"github.com/storacha/go-ucanto/ucan"
	"github.com/storacha/indexing-service/pkg/capability/assert"
)

var log = logging.Logger("storage")

func NewService() map[ucan.Ability]server.ServiceMethod[assert.Unit] {
	return map[ucan.Ability]server.ServiceMethod[assert.Unit]{}
}
