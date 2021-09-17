package main

import (
	"context"
	"fmt"

	pb "github.com/ariden83/bitcoin-wallet/proto/btchdwallet"
	"github.com/ariden83/bitcoin-wallet/wallet"

	"github.com/ariden83/bitcoin-wallet/config"
	"go.uber.org/zap"

	"google.golang.org/grpc"
	"net"

	"github.com/juju/errors"
	"google.golang.org/grpc/reflection"

	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-prometheus"

	"encoding/json"
	logGrpc "github.com/ariden83/bitcoin-wallet/middleware"
	promhttp "github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"time"
)

type Server struct {
	pb.UnimplementedWalletServer
	log           *zap.Logger
	conf          *config.Config
	grpcServer    *grpc.Server
	wallet        *wallet.Wallet
	healthzServer *http.Server
}

func (s *Server) CreateWallet(ctx context.Context, in *pb.Request) (*pb.Response, error) {
	fmt.Println("\nCreating new wallet")

	w := s.wallet.CreateWallet()

	return w, nil
}

func (s *Server) GetWallet(ctx context.Context, in *pb.Request) (*pb.Response, error) {
	fmt.Println("\nGetting wallet data")

	w := s.wallet.DecodeWallet(in.Mnemonic)

	return w, nil
}

func (s *Server) GetBalance(ctx context.Context, in *pb.Request) (*pb.Response, error) {
	fmt.Println("\nGetting Balance data")

	balance := s.wallet.GetBalance(in.Address)

	return balance, nil
}

// Start Set classic routes
func (s *Server) startGRPCServer(stop chan error) {
	go func() {
		lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.conf.Host, s.conf.Port))
		if err != nil {
			s.log.Fatal("failed to listen", zap.Error(err))
		}

		wrap := logGrpc.NewWrappedLogger(s.log)

		s.grpcServer = grpc.NewServer(
			grpc.UnaryInterceptor(
				grpc_middleware.ChainUnaryServer(
					grpc_prometheus.UnaryServerInterceptor,
					wrap.PanicInterceptor(),
					wrap.LoggerInterceptor(),
				)))

		pb.RegisterWalletServer(s.grpcServer, s)

		reflection.Register(s.grpcServer)

		grpc_prometheus.EnableHandlingTimeHistogram()
		grpc_prometheus.Register(s.grpcServer)

		s.log.Info(fmt.Sprintf("Service running at port: %d", s.conf.Port))

		if err := s.grpcServer.Serve(lis); err != nil {
			stop <- errors.Annotate(err, "cannot start server gRPC")
		}
	}()
}

func (s *Server) Shutdown(ctx context.Context) {
	if s.grpcServer != nil {
		s.grpcServer.GracefulStop()
	}
	s.healthzServer.Shutdown(ctx)
}

// Healthz structure
type Healthz struct {
	Result   bool     `json:"result"`
	Messages []string `json:"messages"`
	Version  string   `json:"version"`
}

func (s *Server) startHealthzRoutes(stop chan error) {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		message := "The service " + s.conf.Name + " responds correctly"
		res := Healthz{Result: true, Messages: []string{message}, Version: s.conf.Version}
		js, err := json.Marshal(res)
		if err != nil {
			s.log.Fatal("Fail to jsonify", zap.Error(err))
		}
		if _, err := w.Write(js); err != nil {
			s.log.Fatal("Fail to Write response in http.ResponseWriter", zap.Error(err))
			return
		}
	})

	mux.HandleFunc("/readiness", func(w http.ResponseWriter, r *http.Request) {
		result := true
		message := "The service " + s.conf.Name + " responds correctly"

		res := Healthz{Result: result, Messages: []string{message}, Version: s.conf.Version}
		js, err := json.Marshal(res)
		if err != nil {
			s.log.Fatal("Fail to jsonify", zap.Error(err))
		}
		if _, err := w.Write(js); err != nil {
			s.log.Fatal("Fail to Write response in http.ResponseWriter", zap.Error(err))
			return
		}
	})

	mux.Handle("/metrics", promhttp.Handler())

	addr := fmt.Sprintf("%s:%d", s.conf.Metrics.Host, s.conf.Metrics.Port)
	s.healthzServer = &http.Server{
		Addr:           addr,
		Handler:        mux,
		ReadTimeout:    time.Duration(s.conf.Healthz.ReadTimeout) * time.Second,
		WriteTimeout:   time.Duration(s.conf.Healthz.WriteTimeout) * time.Second,
		IdleTimeout:    120 * time.Second,
		MaxHeaderBytes: 1 << 12,
	}
	go func() {
		s.log.Info("Listening HTTP for healthz route", zap.String("address", addr))
		if err := s.healthzServer.ListenAndServe(); err != nil {
			stop <- errors.Annotate(err, "cannot start healthz server")
		}
	}()
}
