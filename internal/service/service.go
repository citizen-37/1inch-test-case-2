package service

import (
	"context"
	"fmt"
	"go.uber.org/zap"
)

const eventBuffSize = 100 // рандомное значение, возможно буффер не нужен совсем - зависит от характера нагрузки

type Service struct {
	logger     *zap.Logger
	subscriber Subscriber
	storage    Storage
}

func NewService(logger *zap.Logger, subscriber Subscriber, storage Storage) *Service {
	return &Service{
		logger:     logger,
		subscriber: subscriber,
		storage:    storage,
	}
}

func (s *Service) Run(ctx context.Context) error {
	sink, errs := s.subscriber.Subscribe(ctx)

	// TODO: initialize state, keep it consistent

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-errs:
			return fmt.Errorf("subscriber error; %w", err)
		case event := <-sink:
			s.logger.Info("swap", zap.Any("data", event))
			err := s.storage.Save(ctx, event)
			if err != nil {
				return fmt.Errorf("save event failed: %w", err)
			}
		}
	}
}
