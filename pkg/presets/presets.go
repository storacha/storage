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
		"did:web:staging.up.storacha.network": "did:key:z6MkhcbEpJpEvNVDd3n5RurquVdqs5dPU16JDU5VZTDtFgnn",
		"did:web:up.storacha.network":         "did:key:z6MkmbbLigYdv5EuU9tJMDXXUudbySwVNeHNqhQGJs7ALUsF",
		"did:web:staging.web3.storage":        "did:key:z6MkhcbEpJpEvNVDd3n5RurquVdqs5dPU16JDU5VZTDtFgnn",
		"did:web:web3.storage":                "did:key:z6MkmbbLigYdv5EuU9tJMDXXUudbySwVNeHNqhQGJs7ALUsF",
	}
)
