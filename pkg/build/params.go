package build

import (
	"github.com/filecoin-project/go-address"
)

var BuildType int

const (
	BuildMainnet  = 1
	BuildCalibnet = 2
)

func SetAddressNetwork(n address.Network) {
	address.CurrentNetwork = n
}
