package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"1inch-test-case-2/internal/rabbit"
	"github.com/wagslane/go-rabbitmq"
	"go.uber.org/zap"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	cancelChan := make(chan os.Signal, 1)

	signal.Notify(cancelChan, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-cancelChan
		cancel()
	}()

	logger, _ := zap.NewDevelopment()

	conn, err := rabbitmq.NewConn(
		"amqp://user:password@localhost:5672",
		rabbitmq.WithConnectionOptionsLogging,
	)
	if err != nil {
		logger.Fatal("rabbitmq connect failed", zap.Error(err))
	}
	defer conn.Close()

	consumer, err := rabbitmq.NewConsumer(conn, func(d rabbitmq.Delivery) (action rabbitmq.Action) {
		logger.Info("message received", zap.ByteString("data", d.Body))

		return rabbitmq.Ack
	}, "whatever queue",
		rabbitmq.WithConsumerOptionsRoutingKey(rabbit.RoutingKey),
		rabbitmq.WithConsumerOptionsExchangeName(rabbit.ExchangeName),
	)
	defer consumer.Close()

	<-ctx.Done()
}
