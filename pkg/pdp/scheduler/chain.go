package scheduler

import (
	"context"
	"time"

	"golang.org/x/xerrors"

	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/chain/store"
	"github.com/filecoin-project/lotus/chain/types"
)

type NodeAPI interface {
	ChainHead(context.Context) (*types.TipSet, error)
	ChainNotify(context.Context) (<-chan []*api.HeadChange, error)
}

type Chain struct {
	api NodeAPI

	callbacks []UpdateFunc
	started   bool
}

func NewChain(api NodeAPI) *Chain {
	return &Chain{
		api: api,
	}
}

type UpdateFunc func(ctx context.Context, revert, apply *types.TipSet) error

func (s *Chain) AddHandler(ch UpdateFunc) error {
	if s.started {
		return xerrors.Errorf("cannot add handler after start")
	}

	s.callbacks = append(s.callbacks, ch)
	return nil
}

func (s *Chain) Run(ctx context.Context) {
	s.started = true

	var (
		notifs <-chan []*api.HeadChange
		err    error
		gotCur bool
	)

	// not fine to panic after this point
	for {
		if notifs == nil {
			notifs, err = s.api.ChainNotify(ctx)
			if err != nil {
				log.Errorw("ChainNotify failed to get chain state, retrying...", "error", err)

				// TODO use a mockable clock, like the one Raul wrote
				time.Sleep(3 * time.Second)
				continue
			}

			gotCur = false
			log.Debug("restarting chain scheduler")
		}

		select {
		case changes, ok := <-notifs:
			if !ok {
				log.Warn("chain notifs channel closed")
				notifs = nil
				continue
			}

			if !gotCur {
				if len(changes) != 1 {
					log.Errorf("expected first notif to have len = 1")
					continue
				}
				chg := changes[0]
				if chg.Type != store.HCCurrent {
					log.Errorf("expected first notif to tell current ts")
					continue
				}

				s.update(ctx, nil, chg.Val)

				gotCur = true
				continue
			}

			var lowest, highest *types.TipSet = nil, nil

			for _, change := range changes {
				if change.Val == nil {
					log.Errorf("change.Val was nil")
				}
				switch change.Type {
				case store.HCRevert:
					lowest = change.Val
				case store.HCApply:
					highest = change.Val
				}
			}

			s.update(ctx, lowest, highest)

		case <-ctx.Done():
			return
		}
	}
}

func (s *Chain) update(ctx context.Context, revert, apply *types.TipSet) {
	if apply == nil {
		log.Error("no new tipset in Chain.update")
		return
	}

	for _, ch := range s.callbacks {
		if err := ch(ctx, revert, apply); err != nil {
			log.Errorf("handling head updates in curio chain sched: %+v", err)
		}
	}
}
