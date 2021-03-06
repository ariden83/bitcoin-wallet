package middleware

import (
	"context"
	"errors"
	"github.com/ariden83/bitcoin-wallet/zap-graylog/logger"
	"github.com/satori/go.uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const (
	RequestIDHeaderKey = "X-Request-ID"
	RequestIDKey       = "RequestID"
)

type WrappedLogger struct {
	*zap.Logger
}

func NewWrappedLogger(l *zap.Logger) *WrappedLogger {
	return &WrappedLogger{l}
}

func (l *WrappedLogger) PanicInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		defer func() {
			if err := recover(); err != nil {
				l.Panic("panic from handler", zap.Error(err.(error)))
			}
		}()

		resp, err = handler(ctx, req)
		return
	}
}

func (l *WrappedLogger) LoggerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, errors.New("no metadata found in context")
		}
		requestID := ""
		values := md.Get(RequestIDHeaderKey)
		if len(values) == 0 || values[0] == "" {
			requestID = uuid.NewV4().String()
		} else {
			requestID = values[0]
		}

		ctx = logger.ToContext(ctx, l.Logger.With(zap.String(RequestIDKey, requestID)))
		ctx = context.WithValue(ctx, RequestIDKey, requestID)

		resp, err = handler(ctx, req)
		return
	}
}
