package build

import (
	"github.com/filecoin-project/go-address"
)

func SetAddressNetwork(n address.Network) {
	address.CurrentNetwork = n
}
