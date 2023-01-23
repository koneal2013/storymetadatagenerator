package agent

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/koneal2013/storymetadatagenerator/internal/auth"
	"github.com/koneal2013/storymetadatagenerator/internal/observability"
	"github.com/koneal2013/storymetadatagenerator/internal/server"
)

type Config struct {
	ServerTLSConfig       *tls.Config
	PeerTLSConfig         *tls.Config
	NodeName              string
	ACLModelFile          string
	ACLPolicyFile         string
	OTPLCollectorURL      string
	OTPLCollectorInsecure bool
	IsDevelopment         bool
	HttpPort              int
	GrpcPort              int
	MiddlewareFuncs       []mux.MiddlewareFunc
}
type Agent struct {
	Config
	traceProvider *trace.TracerProvider
	serverGrpc    *grpc.Server
	serverHttp    *http.Server

	shutdown     bool
	shutdowns    chan struct{}
	shutdownLock sync.Mutex
}

func (a *Agent) setupLogger() error {
	if logger, err := observability.NewLogger(a.Config.IsDevelopment, "storymetadatagenerator"); err != nil {
		return err
	} else {
		zap.ReplaceGlobals(logger.Named(a.Config.NodeName))
		if a.Config.IsDevelopment {
			if _, err := zap.RedirectStdLogAt(zap.L(), zap.DebugLevel); err != nil {
				return err
			} else {
				if _, err := zap.RedirectStdLogAt(zap.L(), zap.WarnLevel); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (a *Agent) setupServers() error {
	if authorizer, err := auth.New(a.Config.ACLModelFile, a.Config.ACLPolicyFile); err != nil {
		return err
	} else if tp, err := observability.NewTrace(fmt.Sprintf("storymetadatagenerator.%s", a.Config.NodeName), a.Config.OTPLCollectorURL, a.Config.OTPLCollectorInsecure); err != nil {
		return err
	} else {
		a.traceProvider = tp
		grpcServerConfig := &server.GrpcConfig{Authorizer: authorizer}
		httpServerConfig := &server.HttpConfig{
			Port:            a.HttpPort,
			MiddlewareFuncs: a.MiddlewareFuncs,
		}
		var opts []grpc.ServerOption
		if a.Config.ServerTLSConfig != nil {
			creds := credentials.NewTLS(a.Config.ServerTLSConfig)
			opts = append(opts, grpc.Creds(creds))
		}
		var err error
		a.serverGrpc, err = server.NewGRPCServer(grpcServerConfig, opts...)
		if err != nil {
			return err
		}
		a.serverHttp = server.NewHTTPServer(httpServerConfig)
	}
	return nil
}

func (a *Agent) Shutdown() error {
	a.shutdownLock.Lock()
	defer a.shutdownLock.Unlock()
	if a.shutdown {
		return nil
	}
	a.shutdown = true
	close(a.shutdowns)

	shutdown := []func(ctx context.Context) error{
		func(ctx context.Context) error {
			a.serverGrpc.GracefulStop()
			return nil
		},
		a.serverHttp.Shutdown,
		a.traceProvider.Shutdown,
	}
	for _, fn := range shutdown {
		if err := fn(context.Background()); err != nil {
			return err
		}
	}
	return nil
}

func (a *Agent) serveHttp() error {
	if err := a.serverHttp.ListenAndServe(); err != nil {
		_ = a.Shutdown()
		return err
	}
	return nil
}

func (a *Agent) serveGrpc() error {
	if ln, err := net.Listen("tcp", fmt.Sprintf(":%d", a.GrpcPort)); err != nil {
		_ = a.Shutdown()
		return err
	} else if err := a.serverGrpc.Serve(ln); err != nil {
		_ = a.Shutdown()
		return err
	}
	return nil
}

func New(config Config) (*Agent, error) {
	a := &Agent{
		Config:    config,
		shutdowns: make(chan struct{}),
	}
	setup := []func() error{
		a.setupLogger,
		a.setupServers,
	}
	for _, fn := range setup {
		if err := fn(); err != nil {
			return nil, err
		}
	}
	logger := zap.L().Named("agent")
	// goroutine for http server
	go func() {
		logger.Sugar().Infof("starting http server on port %d", a.HttpPort)
		err := a.serveHttp()
		if err != nil {
			logger.Sugar().Error("error starting http server", err)
		}
	}()
	// goroutine for grpc server
	go func() {
		logger.Sugar().Infof("starting grpc server on port %d", a.GrpcPort)
		err := a.serveGrpc()
		if err != nil {
			logger.Sugar().Error("error starting Grpc server", err)
		}
	}()
	return a, nil
}
