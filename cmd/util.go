package cmd

import (
	"context"
	"fmt"
	"os"
	"path"
	"time"

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

// Poll calls fn every interval until fn returns done=true, fn returns a non‑nil error,
// or the context expires. It returns the first error encountered, or ctx.Err() if the deadline is reached.
func Poll(ctx context.Context, interval time.Duration, fn func() (done bool, err error)) error {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		// First, check immediately (in case the condition is already met)
		done, err := fn()
		if err != nil {
			return err
		}
		if done {
			return nil
		}

		// Wait for next tick or context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			// loop and call fn again
		}
	}
}
