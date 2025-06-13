package serve

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	ucanserver "github.com/storacha/go-ucanto/server"
	"go.uber.org/fx"

	"github.com/storacha/piri/cmd/cliutil"
	"github.com/storacha/piri/pkg/config"
	"github.com/storacha/piri/pkg/principalresolver"
	"github.com/storacha/piri/pkg/server"
	"github.com/storacha/piri/pkg/service/storage"
)

var (
	UCANCmd = &cobra.Command{
		Use:   "ucan",
		Short: "Start the UCAN server.",
		Args:  cobra.NoArgs,
		RunE:  startServer,
	}
)

func init() {
	UCANCmd.Flags().String(
		"host",
		config.DefaultUCANServer.Host,
		"Host to listen on")
	cobra.CheckErr(viper.BindPFlag("host", UCANCmd.Flags().Lookup("host")))

	UCANCmd.Flags().Uint(
		"port",
		config.DefaultUCANServer.Port,
		"Port to listen on",
	)
	cobra.CheckErr(viper.BindPFlag("port", UCANCmd.Flags().Lookup("port")))

	UCANCmd.Flags().String(
		"public-url",
		config.DefaultUCANServer.PublicURL,
		"URL the node is publicly accessible at and exposed to other storacha services",
	)
	cobra.CheckErr(viper.BindPFlag("public_url", UCANCmd.Flags().Lookup("public-url")))

	UCANCmd.Flags().String(
		"pdp-server-url",
		"",
		"URL used to connect to pdp server",
	)
	cobra.CheckErr(viper.BindPFlag("pdp_server_url", UCANCmd.Flags().Lookup("pdp-server-url")))

	UCANCmd.Flags().Uint64(
		"proof-set",
		config.DefaultUCANServer.ProofSet,
		"Proofset to use with PDP",
	)
	cobra.CheckErr(viper.BindPFlag("proof_set", UCANCmd.Flags().Lookup("proof-set")))
	UCANCmd.MarkFlagsRequiredTogether("pdp-server-url", "proof-set")

	UCANCmd.Flags().String(
		"indexing-service-proof",
		config.DefaultUCANServer.IndexingServiceProof,
		"A delegation that allows the node to cache claims with the indexing service",
	)
	cobra.CheckErr(viper.BindPFlag("indexing_service_proof", UCANCmd.Flags().Lookup("indexing-service-proof")))

	UCANCmd.Flags().String(
		"indexing-service-did",
		config.DefaultUCANServer.IndexingServiceDID,
		"DID of the indexing service",
	)
	cobra.CheckErr(UCANCmd.Flags().MarkHidden("indexing-service-did"))
	cobra.CheckErr(viper.BindPFlag("indexing_service_did", UCANCmd.Flags().Lookup("indexing-service-did")))

	UCANCmd.Flags().String(
		"indexing-service-url",
		config.DefaultUCANServer.IndexingServiceURL,
		"URL of the indexing service",
	)
	cobra.CheckErr(UCANCmd.Flags().MarkHidden("indexing-service-url"))
	cobra.CheckErr(viper.BindPFlag("indexing_service_url", UCANCmd.Flags().Lookup("indexing-service-url")))

	UCANCmd.Flags().String(
		"upload-service-did",
		config.DefaultUCANServer.UploadServiceDID,
		"DID of the upload service",
	)
	cobra.CheckErr(UCANCmd.Flags().MarkHidden("upload-service-did"))
	cobra.CheckErr(viper.BindPFlag("upload_service_did", UCANCmd.Flags().Lookup("upload-service-did")))

	UCANCmd.Flags().String(
		"upload-service-url",
		config.DefaultUCANServer.UploadServiceURL,
		"URL of the upload service",
	)
	cobra.CheckErr(UCANCmd.Flags().MarkHidden("upload-service-url"))
	cobra.CheckErr(viper.BindPFlag("upload_service_url", UCANCmd.Flags().Lookup("upload-service-url")))

	UCANCmd.Flags().StringSlice(
		"ipni-announce-urls",
		config.DefaultUCANServer.IPNIAnnounceURLs,
		"A list of IPNI announce URLs")
	cobra.CheckErr(UCANCmd.Flags().MarkHidden("ipni-announce-urls"))
	cobra.CheckErr(viper.BindPFlag("ipni_announce_urls", UCANCmd.Flags().Lookup("ipni-announce-urls")))

	UCANCmd.Flags().StringToString(
		"service-principal-mapping",
		config.DefaultUCANServer.ServicePrincipalMapping,
		"Mapping of service DIDs to principal DIDs",
	)
	cobra.CheckErr(UCANCmd.Flags().MarkHidden("service-principal-mapping"))
	cobra.CheckErr(viper.BindPFlag("service_principal_mapping", UCANCmd.Flags().Lookup("service-principal-mapping")))

}

func startServer(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()

	cfg, err := config.Load[config.UCANServer]()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	presolv, err := principalresolver.New(cfg.ServicePrincipalMapping)
	if err != nil {
		return fmt.Errorf("creating principal resolver: %w", err)
	}

	app := storage.NewApp(cfg,
		fx.Invoke(func(lc fx.Lifecycle, svc storage.Service) {
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					// Start server in a goroutine since it blocks
					go func() {
						err := server.ListenAndServe(
							fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
							svc,
							ucanserver.WithPrincipalResolver(presolv.ResolveDIDKey),
						)
						if err != nil {
							log.Errorf("failed to start server: %w", err)
						}
					}()
					cliutil.PrintHero(svc.ID().DID())
					cmd.Println("Listening on", cfg.Host+":"+strconv.Itoa(int(cfg.Port)))
					return nil
				},
				OnStop: func(ctx context.Context) error {
					// Implement graceful shutdown if server.ListenAndServe supports it
					return nil
				},
			})
		}),
	)

	if err := app.Start(ctx); err != nil {
		return fmt.Errorf("starting app: %w", err)
	}
	<-ctx.Done()
	stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return app.Stop(stopCtx)
}
