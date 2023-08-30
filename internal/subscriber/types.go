package subscriber

import (
	"1inch-test-case-2/internal/contracts"
	"1inch-test-case-2/internal/service"
	"context"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/event"
)

const sinkSize = 100 // выбрал не думая; нужен какой-то буффер, не тратил время на поиски обоснования

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
