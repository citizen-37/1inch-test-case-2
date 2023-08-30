package subscriber

import (
	"context"
	"fmt"
	"math/big"

	"1inch-test-case-2/internal/contracts"
	"1inch-test-case-2/internal/service"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type PoolV2 struct {
	logger         *zap.Logger
	contract       ContractV2
	token1, token0 common.Address
}

func NewPoolV2(logger *zap.Logger, contract ContractV2) *PoolV2 {
	return &PoolV2{
		logger:   logger,
		contract: contract,
	}
}

func (p *PoolV2) Subscribe(ctx context.Context) (<-chan *service.Event, <-chan error) {
	sink := make(chan *service.Event, sinkSize)
	errs := make(chan error, 1)

	go p.subscribeAndWait(ctx, sink, errs)

	return sink, errs
}

func (p *PoolV2) subscribeAndWait(ctx context.Context, sink chan<- *service.Event, errs chan<- error) {
	defer close(errs)
	defer close(sink)
	defer func() {
		if err := recover(); err != nil {
			errs <- fmt.Errorf("pool v2 panic recovered: %s", err)
		}
	}()

	cCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	err := p.loadTokens(ctx)
	if err != nil {
		errs <- fmt.Errorf("load tokens failed: %w", err)
		return
	}

	swaps := make(chan *contracts.PairSwap)

	sub, err := p.contract.WatchSwap(&bind.WatchOpts{
		Context: cCtx,
	}, swaps, nil, nil)
	if err != nil {
		errs <- fmt.Errorf("watch swap failed: %w", err)
		return
	}
	defer sub.Unsubscribe()

	for {
		select {
		case data := <-swaps:
			p.logger.Info("swaps received", zap.Any("data", data))
			for _, event := range p.swapsToEvents(data) {
				p.logger.Info("event transformed", zap.Any("event", event))
				sink <- event
			}
		case err := <-sub.Err():
			errs <- err
			return
		case <-ctx.Done():
			errs <- ctx.Err()
			return
		}
	}
}

func (p *PoolV2) loadTokens(ctx context.Context) error {
	wg, gCtx := errgroup.WithContext(ctx)

	wg.Go(func() error {
		t0, err := p.contract.Token0(&bind.CallOpts{
			Context: gCtx,
		})
		if err != nil {
			return fmt.Errorf("load token 0 failed: %w", err)
		}

		p.token0 = t0
		return nil
	})

	wg.Go(func() error {
		t1, err := p.contract.Token1(&bind.CallOpts{
			Context: gCtx,
		})
		if err != nil {
			return fmt.Errorf("load token 1 failed: %w", err)
		}

		p.token1 = t1
		return nil
	})

	err := wg.Wait()
	if err != nil {
		return fmt.Errorf("errgroup failed: %w", err)
	}

	return nil
}

func (p *PoolV2) swapsToEvents(swap *contracts.PairSwap) []*service.Event {
	if swap.Amount0In != nil && swap.Amount0In.Cmp(big.NewInt(0)) != 0 {
		// token 0 spent
		return []*service.Event{
			{
				Token:  p.token0,
				Amount: swap.Amount0In.Neg(swap.Amount0In), // spent
			}, {
				Token:  p.token1,
				Amount: swap.Amount1Out, // gained
			},
		}
	}

	// swap.Amount1In != nil
	// token 1 spent
	return []*service.Event{
		{
			Token:  p.token0,
			Amount: swap.Amount0Out, // gained
		}, {
			Token:  p.token1,
			Amount: swap.Amount1In.Neg(swap.Amount1In), // spent
		},
	}
}
