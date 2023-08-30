package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"1inch-test-case-2/internal/config"
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

	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("config load failed", zap.Error(err))
	}

	conn, err := rabbitmq.NewConn(
		fmt.Sprintf("amqp://%s:%s@%s", cfg.RabbitUser, cfg.RabbitPassword, cfg.RabbitHost),
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
