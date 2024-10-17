package storage

import (
	"fmt"
	"net/url"
	"os"
	"path"

	leveldb "github.com/ipfs/go-ds-leveldb"
	"github.com/ipni/go-libipni/maurl"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/multiformats/go-multihash"
	"github.com/storacha/go-ucanto/principal"
	ed25519 "github.com/storacha/go-ucanto/principal/ed25519/signer"
	"github.com/storacha/go-ucanto/principal/rsa/verifier"
	"github.com/storacha/go-ucanto/ucan"
	ipnipub "github.com/storacha/ipni-publisher/pkg/publisher"
	"github.com/storacha/ipni-publisher/pkg/store"
	"github.com/storacha/storage/pkg/access"
	"github.com/storacha/storage/pkg/internal/digestutil"
	"github.com/storacha/storage/pkg/presigner"
	"github.com/storacha/storage/pkg/service/publisher"
	"github.com/storacha/storage/pkg/store/allocationstore"
	"github.com/storacha/storage/pkg/store/blobstore"
	"github.com/storacha/storage/pkg/store/claimstore"
	"github.com/storacha/storage/pkg/store/delegationstore"
)

type StorageService struct {
	id          principal.Signer
	blobs       blobstore.Blobstore
	allocations allocationstore.AllocationStore
	claims      claimstore.ClaimStore
	access      access.Access
	presigner   presigner.RequestPresigner
	publisher   publisher.Publisher
}

func (s *StorageService) Access() access.Access {
	return s.access
}

func (s *StorageService) Allocations() allocationstore.AllocationStore {
	return s.allocations
}

func (s *StorageService) Blobs() blobstore.Blobstore {
	return s.blobs
}

func (s *StorageService) Claims() claimstore.ClaimStore {
	return s.claims
}

func (s *StorageService) ID() principal.Signer {
	return s.id
}

func (s *StorageService) Presigner() presigner.RequestPresigner {
	return s.presigner
}

func (s *StorageService) Publisher() publisher.Publisher {
	return s.publisher
}

var _ Service = (*StorageService)(nil)

func New(opts ...Option) (*StorageService, error) {
	c := &config{}
	for _, opt := range opts {
		opt(c)
	}

	if c.id == nil {
		log.Warn("Generating a server identity as one has not been set!")
		id, err := ed25519.Generate()
		if err != nil {
			return nil, err
		}
		c.id = id
	}
	log.Infof("Server ID: %s", c.id.DID())

	homedir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	dataDir := c.dataDir
	if dataDir == "" {
		dataDir := path.Join(homedir, ".storage", "blob")
		log.Warnf("Data directory not set, using default: %s", dataDir)
	}

	blobs, err := blobstore.NewFsBlobstore(dataDir)
	if err != nil {
		return nil, err
	}

	allocDs := c.allocationDatastore
	if allocDs == nil {
		dir := path.Join(homedir, ".storage", "allocation")
		allocDs, err = leveldb.NewDatastore(dir, nil)
		if err != nil {
			return nil, err
		}
		log.Warnf("Allocation datastore not set, using LevelDB: %s", dir)
	}

	allocations, err := allocationstore.NewDsAllocationStore(allocDs)
	if err != nil {
		return nil, err
	}

	claimDs := c.claimDatastore
	if claimDs == nil {
		dir := path.Join(homedir, ".storage", "claim")
		claimDs, err = leveldb.NewDatastore(dir, nil)
		if err != nil {
			return nil, err
		}
		log.Warnf("Claim datastore not set, using LevelDB: %s", dir)
	}

	claims, err := delegationstore.NewDsDelegationStore(claimDs)
	if err != nil {
		return nil, err
	}

	pubURL := c.publicURL
	if pubURL == (url.URL{}) {
		u, err := url.Parse("http://localhost:3000")
		if err != nil {
			return nil, err
		}
		log.Warnf("Public URL not set, using default: %s", u)
		pubURL = *u
	}

	accessURL := pubURL
	accessURL.Path = "/blob"
	access, err := access.NewPatternAccess(fmt.Sprintf("%s/{blob}", accessURL.String()))
	if err != nil {
		return nil, err
	}

	accessKeyID := c.id.DID().String()
	idDigest, _ := multihash.Sum(c.id.Encode(), multihash.SHA2_256, -1)
	secretAccessKey := digestutil.Format(idDigest)
	presigner, err := presigner.NewS3RequestPresigner(accessKeyID, secretAccessKey, pubURL, "blob")
	if err != nil {
		return nil, err
	}

	priv, err := crypto.UnmarshalEd25519PrivateKey(c.id.Raw())
	if err != nil {
		return nil, err
	}

	publisherDs := c.publisherDatastore
	if publisherDs == nil {
		dir := path.Join(homedir, ".storage", "publisher")
		claimDs, err = leveldb.NewDatastore(dir, nil)
		if err != nil {
			return nil, err
		}
		log.Warnf("Publisher datastore not set, using LevelDB: %s", dir)
	}

	ipni, err := ipnipub.New(
		priv,
		store.FromDatastore(publisherDs),
		ipnipub.WithDirectAnnounce("https://cid.contact/announce"),
		ipnipub.WithAnnounceAddrs("/dns4/localhost/tcp/3000/https"),
	)
	if err != nil {
		return nil, err
	}

	peerid, err := toPeerID(c.id)
	if err != nil {
		return nil, err
	}
	addr, err := maurl.FromURL(&pubURL)
	if err != nil {
		return nil, err
	}
	peerInfo := peer.AddrInfo{
		ID:    peerid,
		Addrs: []multiaddr.Multiaddr{addr},
	}
	publisher, err := publisher.New(ipni, peerInfo, "claim/{claim}")
	if err != nil {
		return nil, err
	}

	return &StorageService{
		id:          c.id,
		blobs:       blobs,
		allocations: allocations,
		claims:      claims,
		access:      access,
		presigner:   presigner,
		publisher:   publisher,
	}, nil
}

func toPeerID(principal ucan.Principal) (peer.ID, error) {
	vfr, err := verifier.Decode(principal.DID().Bytes())
	if err != nil {
		return "", err
	}
	pub, err := crypto.UnmarshalEd25519PublicKey(vfr.Raw())
	if err != nil {
		return "", err
	}
	return peer.IDFromPublicKey(pub)
}
