package server_test

import (
	"context"
	"net"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"

	api "github.com/koneal2013/storymetadatagenerator/api/v1/grpc"
	"github.com/koneal2013/storymetadatagenerator/internal/auth"
	"github.com/koneal2013/storymetadatagenerator/internal/config"
	"github.com/koneal2013/storymetadatagenerator/internal/observability"
	"github.com/koneal2013/storymetadatagenerator/internal/server"
)

func TestGrpcServer(t *testing.T) {
	for scenario, fn := range map[string]func(
		t *testing.T,
		rootClient, nobodyClient api.StorymetadataClient,
	){
		"get story metadata for 'N' number of pages succeeds": testGetStoryMetadata,
		"unauthorized fails": testUnauthorized,
	} {
		t.Run(scenario, func(t *testing.T) {
			tp, err := observability.NewTrace("test.grpc.storymetadatagenerator",
				"localhost:4317", true)
			require.NoError(t, err)
			rootClient, nobodyClient, teardown := setupTest(t, tp)
			defer teardown()
			fn(t, rootClient, nobodyClient)
		})
	}
}

func setupTest(t *testing.T, tp *sdktrace.TracerProvider) (rootClient, nobodyClient api.StorymetadataClient,
	teardown func()) {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	newClient := func(crtPath, keyPath string) (*grpc.ClientConn, api.StorymetadataClient, []grpc.DialOption) {
		tlsConfig, err := config.SetupTLSConfig(config.TLSConfig{
			CertFile: crtPath,
			KeyFile:  keyPath,
			CAFile:   config.CAFile,
			Server:   false,
		})
		require.NoError(t, err)
		tlsCreds := credentials.NewTLS(tlsConfig)
		opts := []grpc.DialOption{grpc.WithTransportCredentials(tlsCreds),
			grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor(otelgrpc.WithTracerProvider(tp))),
			grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor(otelgrpc.WithTracerProvider(tp)))}
		conn, err := grpc.Dial(l.Addr().String(), opts...)
		require.NoError(t, err)
		client := api.NewStorymetadataClient(conn)
		return conn, client, opts
	}

	rootConn, rootClient, _ := newClient(config.RootClientCertFile, config.RootClientKeyFile)
	nobodyConn, nobodyClient, _ := newClient(config.NobodyClientCertFile, config.NobodyClientKeyFile)

	serverTLSConfig, err := config.SetupTLSConfig(config.TLSConfig{
		CertFile:      config.ServerCertFile,
		KeyFile:       config.ServerKeyFile,
		CAFile:        config.CAFile,
		ServerAddress: l.Addr().String(),
		Server:        true,
	})
	require.NoError(t, err)
	serverCreds := credentials.NewTLS(serverTLSConfig)

	authorizer, err := auth.New(config.ACLModelFile, config.ACLPolicyFile)
	require.NoError(t, err)

	cfg := &server.GrpcConfig{
		Authorizer: authorizer,
	}

	server, err := server.NewGRPCServer(cfg, grpc.Creds(serverCreds))
	require.NoError(t, err)

	go func() {
		server.Serve(l)
	}()

	return rootClient, nobodyClient, func() {
		server.Stop()
		rootConn.Close()
		nobodyConn.Close()
		l.Close()
		tp.Shutdown(context.Background())
	}
}

func testGetStoryMetadata(t *testing.T, client api.StorymetadataClient, _ api.StorymetadataClient) {
	ctx := context.Background()
	metadata, err := client.GetMetadata(ctx, &api.GetStoryMetadataRequest{NumberOfPages: 1})
	require.NoError(t, err, "err should be nil")
	require.NotNil(t, metadata, "metadata should not be nil")
}

func testUnauthorized(t *testing.T, _, client api.StorymetadataClient) {
	ctx := context.Background()
	if metadata, err := client.GetMetadata(ctx, &api.GetStoryMetadataRequest{NumberOfPages: 1}); metadata != nil {
		t.Fatalf("metadata response should be nil")
	} else {
		gotCode, wantCode := status.Code(err), codes.PermissionDenied
		if gotCode != wantCode {
			t.Fatalf("got code: %d, want: %d", gotCode, wantCode)
		}
	}
}
