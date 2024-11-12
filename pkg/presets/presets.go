package presets

import (
	"net/url"

	"github.com/storacha/go-ucanto/did"
)

var (
	AnnounceURL, _        = url.Parse("https://cid.contact/announce")
	IndexingServiceDID, _ = did.Parse("did:web:indexer.storacha.network")
	IndexingServiceURL, _ = url.Parse("https://indexer.storacha.network")
	PrincipalMapping      = map[string]string{
		"did:web:staging.upload.storacha.network": "did:key:z6MkqVThfb3PVdgT5yxumxjFFjoQ2vWd26VUQKByPuSB9N91",
		"did:web:upload.storacha.network":         "did:key:z6MkmbbLigYdv5EuU9tJMDXXUudbySwVNeHNqhQGJs7ALUsF",
	}
)
