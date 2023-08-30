package rabbit

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/wagslane/go-rabbitmq"
)

const (
	ExchangeName = "pool"
	RoutingKey   = "balance_updates"
)

type Publisher struct {
	cli *rabbitmq.Publisher
}

func NewPublisher(cli *rabbitmq.Publisher) *Publisher {
	return &Publisher{cli: cli}
}

func (p *Publisher) Publish(ctx context.Context, address common.Address, amount *big.Int) error {
	event := &Event{
		Address: address.String(),
		Balance: amount.String(),
	}

	rawMsg, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal failed: %w", err)
	}

	err = p.cli.PublishWithContext(
		ctx,
		rawMsg,
		[]string{RoutingKey},
		rabbitmq.WithPublishOptionsContentType("application/json"),
		rabbitmq.WithPublishOptionsExchange(ExchangeName),
	)
	if err != nil {
		return fmt.Errorf("publish failed: %w", err)
	}

	return nil
}
