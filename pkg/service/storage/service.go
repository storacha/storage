package storage

import (
	logging "github.com/ipfs/go-log/v2"
	"github.com/storacha/go-ucanto/client"
	"github.com/storacha/go-ucanto/principal"

	"github.com/storacha/piri/pkg/pdp"
	"github.com/storacha/piri/pkg/service/blobs"
	"github.com/storacha/piri/pkg/service/claims"
	"github.com/storacha/piri/pkg/service/replicator"
	"github.com/storacha/piri/pkg/service/storage/providers"
	"github.com/storacha/piri/pkg/store/receiptstore"
)

var log = logging.Logger("service/storage")

var _ Service = (*StorageService)(nil)

func NewStorageService(params providers.StorageServiceParams) Service {
	return &StorageService{
		params.Identity,
		params.Blobs,
		params.Claims,
		params.PDP,
		params.ReceiptStore,
		params.Replicator,
		params.UploadService,
	}
}

type StorageService struct {
	id            principal.Signer
	blobs         blobs.Blobs
	claims        claims.Claims
	pdp           pdp.PDP
	receiptStore  receiptstore.ReceiptStore
	replicator    replicator.Replicator
	uploadService client.Connection
}

func (s *StorageService) Replicator() replicator.Replicator {
	return s.replicator
}

func (s *StorageService) UploadConnection() client.Connection {
	return s.uploadService
}

func (s *StorageService) Blobs() blobs.Blobs {
	return s.blobs
}

func (s *StorageService) Claims() claims.Claims {
	return s.claims
}

func (s *StorageService) ID() principal.Signer {
	return s.id
}

func (s *StorageService) PDP() pdp.PDP {
	return s.pdp
}

func (s *StorageService) Receipts() receiptstore.ReceiptStore {
	return s.receiptStore
}
