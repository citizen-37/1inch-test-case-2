package publisher

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type (
	Storage interface {
		GetAll(ctx context.Context) (map[common.Address]*big.Int, error)
	}

	Queue interface {
		Publish(ctx context.Context, address common.Address, balance *big.Int) error
	}
)
