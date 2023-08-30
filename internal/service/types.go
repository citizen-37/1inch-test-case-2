package service

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

type (
	Event struct {
		Token  common.Address
		Amount *big.Int
	}

	Subscriber interface {
		Subscribe(ctx context.Context) (<-chan *Event, <-chan error)
	}

	Storage interface {
		Save(ctx context.Context, event *Event) error
	}
)
