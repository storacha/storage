package cmd

import (
	"fmt"
	"os"
	"path"

	logging "github.com/ipfs/go-log/v2"
	"github.com/storacha/go-ucanto/did"
	"github.com/storacha/piri/pkg/build"
)

var log = logging.Logger("cmd")

func PrintHero(id did.DID) {
	fmt.Printf(`
▗▄▄▖ ▄  ▄▄▄ ▄ 
▐▌ ▐▌▄ █    ▄ 
▐▛▀▘ █ █    █ 
▐▌   █      █

🔥 %s
🆔 %s
🚀 Ready!
`, build.Version, id.String())
}

func mkdirp(dirpath ...string) (string, error) {
	dir := path.Join(dirpath...)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return "", fmt.Errorf("creating directory: %s: %w", dir, err)
	}
	return dir, nil
}
