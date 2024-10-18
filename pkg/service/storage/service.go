package storage

import (
	"fmt"
	"io"
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
	"github.com/storacha/go-ucanto/principal/ed25519/verifier"
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
	closeFuncs  []func() error
	io.Closer
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

func (s *StorageService) Close() error {
	var err error
	for _, close := range s.closeFuncs {
		err = close()
	}
	s.closeFuncs = []func() error{}
	return err
}

var _ Service = (*StorageService)(nil)

func New(opts ...Option) (*StorageService, error) {
	c := &config{}
	for _, opt := range opts {
		err := opt(c)
		if err != nil {
			return nil, err
		}
	}

	if c.id == nil {
		log.Warn("Generating a server identity as one has not been configured!")
		id, err := ed25519.Generate()
		if err != nil {
			return nil, err
		}
		c.id = id
	}
	log.Infof("Server ID: %s", c.id.DID())

	peerid, err := toPeerID(c.id)
	if err != nil {
		return nil, err
	}
	log.Infof("Peer ID: %s", peerid.String())

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("getting user home directory: %w", err)
	}

	var closeFuncs []func() error

	dataDir := c.dataDir
	if dataDir == "" {
		dir, err := mkdirp(homeDir, ".storacha")
		if err != nil {
			return nil, err
		}
		log.Warnf("Data directory not configured, using default: %s", dir)
		dataDir = dir
	}

	blobs, err := blobstore.NewFsBlobstore(path.Join(dataDir, "blobs"))
	if err != nil {
		return nil, err
	}

	allocDs := c.allocationDatastore
	if allocDs == nil {
		dir, err := mkdirp(dataDir, "allocation")
		if err != nil {
			return nil, err
		}
		allocDs, err = leveldb.NewDatastore(dir, nil)
		if err != nil {
			return nil, err
		}
		log.Warnf("Allocation datastore not configured, using LevelDB: %s", dir)
	}
	closeFuncs = append(closeFuncs, allocDs.Close)

	allocations, err := allocationstore.NewDsAllocationStore(allocDs)
	if err != nil {
		return nil, err
	}

	claimDs := c.claimDatastore
	if claimDs == nil {
		dir, err := mkdirp(dataDir, "claim")
		if err != nil {
			return nil, err
		}
		claimDs, err = leveldb.NewDatastore(dir, nil)
		if err != nil {
			return nil, err
		}
		log.Warnf("Claim datastore not configured, using LevelDB: %s", dir)
	}
	closeFuncs = append(closeFuncs, claimDs.Close)

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
		log.Warnf("Public URL not configured, using default: %s", u)
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
		dir, err := mkdirp(dataDir, "publisher")
		if err != nil {
			return nil, err
		}
		publisherDs, err = leveldb.NewDatastore(dir, nil)
		if err != nil {
			return nil, err
		}
		log.Warnf("Publisher datastore not configured, using LevelDB: %s", dir)
	}
	closeFuncs = append(closeFuncs, publisherDs.Close)

	addr, err := maurl.FromURL(&pubURL)
	if err != nil {
		return nil, err
	}

	ipni, err := ipnipub.New(
		priv,
		store.FromDatastore(publisherDs),
		ipnipub.WithDirectAnnounce("https://cid.contact/announce"),
		ipnipub.WithAnnounceAddrs(addr.String()),
	)
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
		closeFuncs:  closeFuncs,
	}, nil
}

func mkdirp(dirpath ...string) (string, error) {
	dir := path.Join(dirpath...)
	err := os.MkdirAll(dir, 0777)
	if err != nil {
		return "", fmt.Errorf("creating directory: %s: %w", dir, err)
	}
	return dir, nil
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
