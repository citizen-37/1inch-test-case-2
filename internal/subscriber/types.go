package subscriber

import (
	"context"

	"1inch-test-case-2/internal/contracts"
	"1inch-test-case-2/internal/service"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/event"
)

const sinkSize = 100

type (
	ContractV2 interface {
		Token0(opts *bind.CallOpts) (common.Address, error)
		Token1(opts *bind.CallOpts) (common.Address, error)
		WatchSwap(opts *bind.WatchOpts, sink chan<- *contracts.PairSwap, sender []common.Address, to []common.Address) (event.Subscription, error)
		// TODO: watch whatever else
	}

	Pool interface {
		Subscribe(ctx context.Context) (<-chan *service.Event, <-chan error)
	}
)
