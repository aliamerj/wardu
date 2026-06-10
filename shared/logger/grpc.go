package logger

import (
	"context"
	"time"

	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// UnaryServerInterceptor logs each unary RPC with duration and request correlation data.
// UnaryClientInterceptor logs each unary outbound RPC with duration and request correlation data.
func UnaryClientInterceptor(log zerolog.Logger) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		start := time.Now()
		err := invoker(ctx, method, req, reply, cc, opts...)

		event := log.Info().
			Str("rpc_method", method).
			Dur("latency", time.Since(start))

		if requestID := requestIDFromContext(ctx); requestID != "" {
			event = event.Str("request_id", requestID)
		}

		if err != nil {
			event = log.Error().
				Err(err).
				Str("rpc_method", method).
				Dur("latency", time.Since(start))
			if requestID := requestIDFromContext(ctx); requestID != "" {
				event = event.Str("request_id", requestID)
			}
			event.Msg("grpc request failed")
			return err
		}

		event.Msg("grpc request")
		return nil
	}
}

func UnaryServerInterceptor(log zerolog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		start := time.Now()
		resp, err := handler(ctx, req)

		event := log.Info().
			Str("rpc_method", info.FullMethod).
			Dur("latency", time.Since(start))

		if requestID := requestIDFromContext(ctx); requestID != "" {
			event = event.Str("request_id", requestID)
		}

		if err != nil {
			event = log.Error().
				Err(err).
				Str("rpc_method", info.FullMethod).
				Dur("latency", time.Since(start))
			if requestID := requestIDFromContext(ctx); requestID != "" {
				event = event.Str("request_id", requestID)
			}
			event.Msg("grpc request failed")
			return resp, err
		}

		event.Msg("grpc request")
		return resp, nil
	}
}

func requestIDFromContext(ctx context.Context) string {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if requestID := requestIDFromMetadata(md); requestID != "" {
			return requestID
		}
	}

	if md, ok := metadata.FromOutgoingContext(ctx); ok {
		if requestID := requestIDFromMetadata(md); requestID != "" {
			return requestID
		}
	}

	return ""
}

func requestIDFromMetadata(md metadata.MD) string {
	if values := md.Get("x-request-id"); len(values) > 0 {
		return values[0]
	}

	return ""
}
