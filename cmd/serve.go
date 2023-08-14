package cmd

import (
	"context"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/tylerauerbeck/kured-silencer/pkg/server"
)

var (
	defaultDuration = 15
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start listening for node label events",
	Run: func(cmd *cobra.Command, args []string) {
		serve(cmd.Context())
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)

	serveCmd.Flags().String("kubeconfig-path", "", "Path to kubeconfig file if not running in cluster")
	viperBindFlag("kubeconfig-path", serveCmd.Flags().Lookup("kubeconfig-path"))

	serveCmd.Flags().String("kured-label", "", "Label to watch for on nodes")
	viperBindFlag("kured-label", serveCmd.Flags().Lookup("kured-label"))

	serveCmd.Flags().String("alertmanager-endpoint", "", "Alertmanager endpoint to send silences to")
	viperBindFlag("alertmanager-endpoint", serveCmd.Flags().Lookup("alertmanager-endpoint"))

	serveCmd.Flags().Duration("removal-buffer", time.Duration(1*time.Minute), "buffer time before removing a silence from a node")
	viperBindFlag("removal-buffer", serveCmd.Flags().Lookup("removal-buffer"))

	serveCmd.Flags().Duration("silence-duration", time.Duration(defaultDuration), "silence duration in minutes")
	viperBindFlag("silence-duration", serveCmd.Flags().Lookup("silence-duration"))
}

func serve(ctx context.Context) {
	logger.Infow("starting kured-silencer", "alertmanager", viper.GetString("alertmanager-endpoint"), "label", viper.GetString("kured-label"))

	srv, err := server.NewServer(ctx, logger)
	if err != nil {
		logger.Fatalw("error creating server", "error", err)
	}

	srv.Run(ctx)
}
