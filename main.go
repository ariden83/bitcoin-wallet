package main

import (
	"fmt"
	l "log"

	"context"
	"github.com/ariden83/bitcoin-wallet/config"
	"github.com/ariden83/bitcoin-wallet/wallet"
	"github.com/ariden83/bitcoin-wallet/zap-graylog/logger"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"time"
)

func main() {
	conf := config.New()

	log, err := logger.NewLogger(
		fmt.Sprintf("%s:%d", conf.Logger.Host, conf.Logger.Port),
		logger.Level(logger.LevelsMap[conf.Logger.Level]),
		logger.Level(logger.LevelsMap[conf.Logger.CLILevel]))
	if err != nil {
		l.Fatal(fmt.Sprintf("cannot setup logger %s:%d", conf.Logger.Host, conf.Logger.Port))
	}

	log = log.With(zap.String("facility", conf.Name), zap.String("version", conf.Version), zap.String("env", conf.Env))
	defer log.Sync()

	stop := make(chan error, 1)

	w := wallet.New(conf, log)

	server := &Server{
		log:    log,
		conf:   conf,
		wallet: w,
	}

	server.startHealthzRoutes(stop)
	server.startGRPCServer(stop)

	/**
	 * And wait for shutdown via signal or error.
	 */
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt)
		<-sig
		stop <- fmt.Errorf("received Interrupt signal")
	}()

	err = <-stop
	log.Error("Shutting down services", zap.Error(err))
	stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	server.Shutdown(stopCtx)
	log.Debug("Services shutted down")
}
