package subscriber

import (
	"1inch-test-case-2/internal/service"
	"context"
	"fmt"
	"golang.org/x/sync/errgroup"
)

type Subscriber struct {
	pools []Pool
}

func NewSubscriber(pools []Pool) *Subscriber {
	return &Subscriber{
		pools: pools,
	}
}

func (s *Subscriber) Subscribe(ctx context.Context) (<-chan *service.Event, <-chan error) {
	sink := make(chan *service.Event, sinkSize)
	errs := make(chan error, 1)

	go s.subscribeAndWait(ctx, sink, errs)

	return sink, errs
}

func (s *Subscriber) subscribeAndWait(ctx context.Context, sink chan<- *service.Event, errs chan<- error) {
	defer close(errs)
	defer close(sink)
	defer func() {
		if err := recover(); err != nil {
			errs <- fmt.Errorf("composite subscriber panic recovered: %s", err)
		}
	}()

	wg, gCtx := errgroup.WithContext(ctx)

	for _, pool := range s.pools {
		pool := pool
		wg.Go(func() error {
			poolSink, poolErrs := pool.Subscribe(gCtx)
			for {
				select {
				case event := <-poolSink:
					sink <- event
				case err := <-poolErrs:
					errs <- err
				}
			}
		})
	}

	err := wg.Wait()
	if err != nil {
		errs <- err
	}
}
