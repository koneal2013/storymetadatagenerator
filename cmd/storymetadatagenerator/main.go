package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/koneal2013/storymetadatagenerator/internal/agent"
	"github.com/koneal2013/storymetadatagenerator/internal/config"
	"github.com/koneal2013/storymetadatagenerator/internal/middleware"
)

type cli struct {
	cfg cfg
}

type cfg struct {
	agent.Config
	ServerTLSConfig config.TLSConfig
	PeerTLSConfig   config.TLSConfig
}

func (c *cli) setupConfig(cmd *cobra.Command, args []string) error {
	if configFile, err := cmd.Flags().GetString("config-file"); err != nil {
		return err
	} else {
		viper.SetConfigFile(configFile)
		if err = viper.ReadInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
				return err
			}
		}
		c.cfg.NodeName = viper.GetString("node-name")
		c.cfg.HttpPort = viper.GetInt("http-port")
		c.cfg.GrpcPort = viper.GetInt("grpc-port")
		c.cfg.IsDevelopment = viper.GetBool("is-development")
		c.cfg.ACLModelFile = viper.GetString("acl-model-file")
		c.cfg.ACLPolicyFile = viper.GetString("acl-policy-file")
		c.cfg.ServerTLSConfig.CertFile = viper.GetString("server-tls-cert-file")
		c.cfg.ServerTLSConfig.KeyFile = viper.GetString("server-tls-key-file")
		c.cfg.ServerTLSConfig.CAFile = viper.GetString("server-tls-ca-file")
		c.cfg.PeerTLSConfig.CertFile = viper.GetString("peer-tls-cert-file")
		c.cfg.PeerTLSConfig.KeyFile = viper.GetString("peer-tls-key-file")
		c.cfg.PeerTLSConfig.CAFile = viper.GetString("peer-tls-ca-file")
		c.cfg.OTPLCollectorInsecure = viper.GetBool("otpl-collector-insecure")
		c.cfg.OTPLCollectorURL = viper.GetString("optl-collector-endpoint")
		if viper.GetBool("enable-logging-middleware") {
			// log each request with the global zap logger (initialized in server.NewHTTPServer)
			c.cfg.MiddlewareFuncs = append(c.cfg.MiddlewareFuncs, middleware.LogRequest)
		}
		c.cfg.MiddlewareFuncs = append(c.cfg.MiddlewareFuncs, middleware.Recovery)
	}
	return nil
}

func (c *cli) run(cmd *cobra.Command, args []string) error {
	if agent, err := agent.New(c.cfg.Config); err != nil {
		return err
	} else {
		// graceful shutdown procedure
		sigc := make(chan os.Signal, 1)
		signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)
		<-sigc
		return agent.Shutdown()
	}
}

func setupFlags(cmd *cobra.Command) error {
	if hostname, err := os.Hostname(); err != nil {
		log.Fatal(err)
	} else {
		otelCollectorEndpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
		if otelCollectorEndpoint == "" {
			otelCollectorEndpoint = "localhost:4317"
		}
		cmd.Flags().String("node-name", hostname, "Unique server ID.")
		cmd.Flags().String("config-file", "", "Path to config file.")
		cmd.Flags().Bool("is-development", true, "Flag to set log level.")
		cmd.Flags().Int("http-port", 8080, "Port to serve Http requests on.")
		cmd.Flags().Int("grpc-port", 8081, "Port to serve Grpc requests on.")
		cmd.Flags().Bool("enable-logging-middleware", true, "Enable logging of each request")
		cmd.Flags().String("acl-model-file", "", "Path to ACL model.")
		cmd.Flags().String("acl-policy-file", "", "Path to ACL policy.")
		cmd.Flags().String("server-tls-cert-file", "", "Path to server tls cert.")
		cmd.Flags().String("server-tls-key-file", "", "Path to server tls key.")
		cmd.Flags().String("server-tls-ca-file", "", "Path to server certificate authority.")
		cmd.Flags().String("peer-tls-cert-file", "", "Path to peer tls cert.")
		cmd.Flags().String("peer-tls-key-file", "", "Path to peer tls key.")
		cmd.Flags().String("peer-tls-ca-file", "", "Path to peer certificate authority.")
		cmd.Flags().String("optl-collector-endpoint", otelCollectorEndpoint, "Endpoint for OTPL tracing collector.")
		cmd.Flags().Bool("otpl-collector-insecure", true, "Flag to enable insecure mode for OTPL Collector.")
		return viper.BindPFlags(cmd.Flags())
	}
	return nil
}

func main() {
	cli := &cli{}

	cmd := &cobra.Command{
		Use:     "storymetadatagenerator",
		PreRunE: cli.setupConfig,
		RunE:    cli.run,
	}
	if err := setupFlags(cmd); err != nil {
		log.Fatal(err)
	}
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
