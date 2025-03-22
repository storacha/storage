package service

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/filecoin-project/go-state-types/abi"
	"github.com/google/uuid"
	"github.com/ipfs/go-cid"
	"github.com/snadrus/must"
	"gorm.io/gorm"

	"github.com/storacha/storage/pkg/pdp/proof"
	"github.com/storacha/storage/pkg/pdp/service/models"
	"github.com/storacha/storage/pkg/pdp/service/types"
)

var PieceSizeLimit = abi.PaddedPieceSize(proof.MaxMemtreeSize).Unpadded()

type PieceHash struct {
	// Name of the hash function used
	// sha2-256-trunc254-padded - CommP
	// sha2-256 - Blob sha256
	Name string

	// hex encoded hash
	Hash string

	// Size of the piece in bytes
	Size int64
}

type PiecePrepareRequest struct {
	Check  types.PieceHash
	Notify string
}

type PiecePrepareResponse struct {
	Location string
	PieceCID cid.Cid
	Created  bool
}

func (p *PDPService) PreparePiece(ctx context.Context, req PiecePrepareRequest) (*PiecePrepareResponse, error) {
	if abi.UnpaddedPieceSize(req.Check.Size) > PieceSizeLimit {
		return nil, fmt.Errorf("piece size exceeds the maximum allowed size")
	}

	pieceCid, havePieceCid, err := req.Check.CommP()
	if err != nil {
		return nil, err
	}

	// Variables to hold information outside the transaction
	var uploadUUID uuid.UUID
	var uploadURL string
	var created bool

	if err := p.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if havePieceCid {
			// Check if a 'parked_pieces' entry exists for the given 'piece_cid'
			// Look up existing parked piece with the given pieceCid, long_term = true, complete = true
			var parkedPiece models.ParkedPiece
			err := tx.Where("piece_cid = ? AND long_term = ? AND complete = ?", pieceCid, true, true).
				First(&parkedPiece).Error

			// If it's neither "record not found" nor nil, it's some other error
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("failed to query parked_pieces: %w", err)
			}

			if err == nil {
				// Create a new parked_piece_refs entry referencing the existing piece
				parkedRef := &models.ParkedPieceRef{
					PieceID:  parkedPiece.ID,
					LongTerm: true,
				}
				if createErr := tx.Create(&parkedRef).Error; createErr != nil {
					return fmt.Errorf("failed to insert into parked_piece_refs: %w", createErr)
				}

				// Create the pdp_piece_uploads record pointing to the parked_piece_refs entry
				uploadUUID = uuid.New()
				upload := &models.PDPPieceUpload{
					ID:             uploadUUID.String(),
					Service:        "storacha",
					PieceCID:       models.Ptr(pieceCid.String()),
					NotifyURL:      req.Notify,
					PieceRef:       &parkedRef.RefID,
					CheckHashCodec: req.Check.Name,
					CheckHash:      must.One(hex.DecodeString(req.Check.Hash)),
					CheckSize:      req.Check.Size,
				}
				if createErr := tx.Create(&upload).Error; createErr != nil {
					return fmt.Errorf("failed to insert into pdp_piece_uploads: %w", createErr)
				}

				// ends transaction
				return nil
			}
		} // else

		// Piece does not exist, proceed to create a new upload request
		uploadUUID = uuid.New()

		// Store the upload request in the database
		var pieceCidStr *string
		if p, ok := req.Check.MaybeStaticCommp(); ok {
			ps := p.String()
			pieceCidStr = &ps
		}

		newUpload := &models.PDPPieceUpload{
			ID:             uploadUUID.String(),
			Service:        "storacha",
			PieceCID:       pieceCidStr, // might be empty if no static commP
			NotifyURL:      req.Notify,
			CheckHashCodec: req.Check.Name,
			CheckHash:      must.One(hex.DecodeString(req.Check.Hash)),
			CheckSize:      req.Check.Size,
		}
		if createErr := tx.Create(&newUpload).Error; createErr != nil {
			return fmt.Errorf("failed to store upload request in database: %w", createErr)
		}

		created = true
		return nil // Commit the transaction

	}); err != nil {
		return nil, err
	}

	if created {
		return &PiecePrepareResponse{
			Location: uploadURL,
			Created:  created,
		}, nil
	}

	return &PiecePrepareResponse{
		PieceCID: pieceCid,
		Created:  false,
	}, nil
}
