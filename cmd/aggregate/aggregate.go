package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/ipfs/go-cid"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/storacha/go-piece/pkg/piece"
	"github.com/storacha/storage/pkg/pdp/aggregator/aggregate"
)

func main() {
	argsWithoutProg := os.Args[1:]
	pieceLinks := make([]piece.PieceLink, 0, len(argsWithoutProg))
	for _, arg := range argsWithoutProg {
		c, err := cid.Decode(arg)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(-1)
		}
		pl, err := piece.FromLink(cidlink.Link{Cid: c})
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(-1)
		}
		pieceLinks = append(pieceLinks, pl)
	}
	aggregate, err := aggregate.NewAggregate(pieceLinks)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(-1)
	}
	asJson, err := json.MarshalIndent(aggregate, "", "  ")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(-1)
	}
	fmt.Printf(string(asJson))
}
