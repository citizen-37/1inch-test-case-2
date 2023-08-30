package publisher

import (
	"context"
	"fmt"
	"time"
)

const flushInterval = 5 * time.Second

type Publisher struct {
	store Storage
	queue Queue
}

func NewPublisher(store Storage, queue Queue) *Publisher {
	return &Publisher{
		store: store,
		queue: queue,
	}
}

func (p *Publisher) Run(ctx context.Context) error {
	for {

		err := p.flush(ctx)
		if err != nil {
			return fmt.Errorf("flush failed: %w", err)
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(flushInterval):
		}
	}
}

func (p *Publisher) flush(ctx context.Context) error {
	balances, err := p.store.GetAll(ctx)
	if err != nil {
		return fmt.Errorf("get all balances failed: %w", err)
	}

	for addr, balance := range balances {
		err = p.queue.Publish(ctx, addr, balance)
		if err != nil {
			return fmt.Errorf("publish failed: %w", err)
		}
	}

	return nil
}
