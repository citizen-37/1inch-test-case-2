package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"1inch-test-case-2/internal/config"
	"1inch-test-case-2/internal/contracts"
	"1inch-test-case-2/internal/publisher"
	"1inch-test-case-2/internal/rabbit"
	"1inch-test-case-2/internal/service"
	"1inch-test-case-2/internal/state"
	"1inch-test-case-2/internal/subscriber"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
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

	client, err := ethclient.DialContext(ctx, fmt.Sprintf("wss://mainnet.infura.io/ws/v3/%s", cfg.InfuraKey))
	if err != nil {
		logger.Fatal("dial infura failed", zap.Error(err))
	}

	conn, err := rabbitmq.NewConn(
		fmt.Sprintf("amqp://%s:%s@%s", cfg.RabbitUser, cfg.RabbitPassword, cfg.RabbitHost),
		rabbitmq.WithConnectionOptionsLogging,
	)
	if err != nil {
		logger.Fatal("rabbitmq connect failed", zap.Error(err))
	}
	defer conn.Close()

	rabbitPub, err := rabbitmq.NewPublisher(
		conn,
		rabbitmq.WithPublisherOptionsExchangeDeclare,
		rabbitmq.WithPublisherOptionsExchangeName(rabbit.ExchangeName),
	)
	if err != nil {
		logger.Fatal("rabbitmq new publisher failed", zap.Error(err))
	}

	store := state.NewStorage()
	queue := rabbit.NewPublisher(rabbitPub)

	compositeSubscriber, err := makeCompositeSubscriber(logger, cfg.PoolV2Addresses, cfg.PoolV3Addresses, client)
	if err != nil {
		logger.Fatal("make composite subscriber failed", zap.Error(err))
	}

	srv := service.NewService(logger, compositeSubscriber, store)
	pub := publisher.NewPublisher(store, queue)

	wg := &sync.WaitGroup{}

	wg.Add(1)
	go func() {
		err := srv.Run(ctx)
		if err != nil && !errors.Is(err, context.Canceled) {
			logger.Fatal("service run failed", zap.Error(err))
		}
	}()

	wg.Add(1)
	go func() {
		err := pub.Run(ctx)
		if err != nil && !errors.Is(err, context.Canceled) {
			logger.Fatal("publisher run failed", zap.Error(err))
		}
	}()

	wg.Wait()
}

func makeCompositeSubscriber(logger *zap.Logger, addressesV2, addressesV3 []string, client *ethclient.Client) (*subscriber.Subscriber, error) {
	var pools []subscriber.Pool

	for _, addr := range addressesV2 {
		cont, err := makeContractV2(addr, client)
		if err != nil {
			return nil, fmt.Errorf("make contract v2 failed: %s: %w", addr, err)
		}

		pools = append(pools, makePoolV2(logger, cont))
	}

	// TODO: build and append poolV3
	// for _, addr := range addressesV3 {
	//	cont, err := makeContractV3(addr, client)
	//	if err != nil {
	//		return nil, fmt.Errorf("make contract v3 failed: %s: %w", addr, err)
	//	}
	//
	//	pools = append(pools, makePoolV3(logger, cont))
	// }

	return subscriber.NewSubscriber(pools), nil
}

func makePoolV2(logger *zap.Logger, cont *contracts.Pair) *subscriber.PoolV2 {
	return subscriber.NewPoolV2(logger, cont)
}

func makeContractV2(address string, client *ethclient.Client) (*contracts.Pair, error) {
	cont, err := contracts.NewPair(common.HexToAddress(address), client)
	if err != nil {
		return nil, fmt.Errorf("new contract v2 failed: %w", err)
	}

	return cont, nil
}
