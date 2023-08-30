package state

import (
	"1inch-test-case-2/internal/service"
	"context"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
	"sync"
)

type Storage struct {
	mx   *sync.RWMutex
	data map[common.Address]*big.Int
}

func NewStorage() *Storage {
	return &Storage{
		mx:   &sync.RWMutex{},
		data: make(map[common.Address]*big.Int),
	}
}

func (s *Storage) Save(_ context.Context, event *service.Event) error {
	s.mx.Lock()
	defer s.mx.Unlock()

	if _, ok := s.data[event.Token]; !ok {
		s.data[event.Token] = &big.Int{}
	}

	s.data[event.Token].Add(s.data[event.Token], event.Amount)

	return nil
}

func (s *Storage) GetAll(_ context.Context) (map[common.Address]*big.Int, error) {
	s.mx.RLock()
	defer s.mx.RUnlock()

	result := make(map[common.Address]*big.Int, len(s.data))

	for k, v := range s.data {
		result[k] = &big.Int{}
		*result[k] = *v
	}

	return result, nil
}
