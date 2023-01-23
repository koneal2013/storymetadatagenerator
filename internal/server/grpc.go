package server

import (
	"context"
	"encoding/json"
	"time"

	grpcmiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpcauth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	otel_codes "go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	peer2 "google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"

	metadata_api_v1 "github.com/koneal2013/storymetadatagenerator/api/v1"
	grpc_api "github.com/koneal2013/storymetadatagenerator/api/v1/grpc"
)

const (
	objectWildCard         = "*"
	getStoryMetadataAction = "GetStoryMetadata"
)

type Authorizer interface {
	Authorize(subject, object, action string) error
}

type GrpcConfig struct {
	Authorizer
}

func NewGRPCServer(config *GrpcConfig, opts ...grpc.ServerOption) (*grpc.Server, error) {
	logger := zap.L().Named("grpc_server")
	zapOpts := []grpc_zap.Option{
		grpc_zap.WithDurationField(
			func(duration time.Duration) zapcore.Field {
				return zap.Int64("grpc.time_ns", duration.Nanoseconds())
			}),
	}
	opts = append(opts, grpc.StreamInterceptor(
		grpcmiddleware.ChainStreamServer(
			grpc_ctxtags.StreamServerInterceptor(),
			grpc_zap.StreamServerInterceptor(logger, zapOpts...),
			grpcauth.StreamServerInterceptor(authenticate),
			otelgrpc.StreamServerInterceptor(),
		)), grpc.UnaryInterceptor(grpcmiddleware.ChainUnaryServer(
		grpc_zap.UnaryServerInterceptor(logger, zapOpts...),
		grpcauth.UnaryServerInterceptor(authenticate),
		otelgrpc.UnaryServerInterceptor(),
	)))
	gsrv := grpc.NewServer(opts...)
	if srv, err := newGrpcServer(config); err != nil {
		return nil, err
	} else {
		grpc_api.RegisterStorymetadataServer(gsrv, srv)
		return gsrv, nil
	}
}

type grpcServer struct {
	grpc_api.UnimplementedStorymetadataServer
	*GrpcConfig
}

func (g grpcServer) GetMetadata(ctx context.Context, request *grpc_api.GetStoryMetadataRequest) (*grpc_api.GetStoryMetadataResponse, error) {
	_, span := otel.GetTracerProvider().Tracer("GrpcTracer").Start(ctx, "GetMetadata")
	if err := g.Authorizer.Authorize(subject(ctx), objectWildCard, getStoryMetadataAction); err != nil {
		span.RecordError(err)
		span.SetStatus(otel_codes.Error, err.Error())
		return nil, err
	}
	res := metadata_api_v1.New(int(request.NumberOfPages))
	res.LoadStories()
	storiesBytes, err := json.Marshal(res.Stories)
	if err != nil {
		return nil, err
	}
	errsBytes, err := json.Marshal(res.Errs)
	if err != nil {
		return nil, err
	}
	return &grpc_api.GetStoryMetadataResponse{
		Stories: storiesBytes,
		Errs:    errsBytes,
	}, nil
}

func newGrpcServer(config *GrpcConfig) (srv *grpcServer, err error) {
	srv = &grpcServer{GrpcConfig: config}
	return srv, nil
}

func authenticate(ctx context.Context) (context.Context, error) {
	if peer, ok := peer2.FromContext(ctx); !ok {
		return ctx, status.New(codes.Unknown, "couldn't find peer info").Err()
	} else if peer.AuthInfo == nil {
		return context.WithValue(ctx, subjectContextKey{}, ""), nil
	} else {
		tlsInfo := peer.AuthInfo.(credentials.TLSInfo)
		subject := tlsInfo.State.VerifiedChains[0][0].Subject.CommonName
		ctx = context.WithValue(ctx, subjectContextKey{}, subject)
		return ctx, nil
	}
}

type subjectContextKey struct{}

func subject(ctx context.Context) string {
	return ctx.Value(subjectContextKey{}).(string)
}
