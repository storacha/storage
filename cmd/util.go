package cmd

import (
	"fmt"
	"os"
	"path"

	logging "github.com/ipfs/go-log/v2"
	"github.com/storacha/go-ucanto/did"
)

var log = logging.Logger("cmd")

func PrintHero(id did.DID) {
	fmt.Printf(`
 00000000                                                   00                  
00      00    00                                            00                  
 000        000000   00000000   00000  0000000    0000000   00000000    0000000 
    00000     00    00     000  00           00  00     0   00    00         00 
        000   00    00      00  00     00000000  00         00    00    0000000 
000     000   00    00     000  00    000    00  000    00  00    00   00    00 
 000000000    0000   0000000    00     000000000   000000   00    00   000000000

🔥 Storage Node %s
🆔 %s
🚀 Ready!
`, "v0.0.0", id.String())
}

func mkdirp(dirpath ...string) (string, error) {
	dir := path.Join(dirpath...)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return "", fmt.Errorf("creating directory: %s: %w", dir, err)
	}
	return dir, nil
}
