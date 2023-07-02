package cmd

import (
	"context"
	"errors"
	"net/url"
	"os"
	"os/signal"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"k8s.io/apimachinery/pkg/watch"

	v1 "k8s.io/api/core/v1"

	"github.com/tylerauerbeck/kured-silencer/pkg/alertmanager"
	"github.com/tylerauerbeck/kured-silencer/pkg/kube"
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

	serveCmd.Flags().Duration("silence-duration", 15, "silence duration in minutes")
	viperBindFlag("silence-duration", serveCmd.Flags().Lookup("silence-duration"))
}

func serve(ctx context.Context) {
	var silencedID string

	ctx, cancel := context.WithCancel(ctx)

	logger.Infow("starting kured-silencer", "alertmanager", viper.GetString("alertmanager-endpoint"), "label", viper.GetString("kured-label"))

	kcli, err := kube.NewKubeClient(viper.GetString("kubeconfig-path"))
	if err != nil {
		logger.Fatalw("error creating kube client", "error", err)
	}

	url, err := url.Parse(viper.GetString("alertmanager-endpoint"))
	if err != nil {
		logger.Fatalw("error parsing alertmanager endpoint", "error", err)
	}

	if err = validateURL(url); err != nil {
		logger.Fatalw("error validating alertmanager endpoint", "error", err)
	}

	amcli := alertmanager.NewSilencerClient(url)

	watcher, err := kube.NewNodeWatcher(ctx, kcli, viper.GetString("kured-label"))
	if err != nil {
		logger.Fatalw("error creating node watcher", "error", err)
	}

	for event := range watcher.ResultChan() {
		switch event.Type {
		case watch.Added:
			logger.Infow("label added", "node", event.Object.(*v1.Node).Name)
			silencedID, err = alertmanager.PostSilence(ctx, amcli, viper.GetDuration("silence-duration"))
			if err != nil {
				logger.Fatalw("error posting silence", "error", err)

				// TODO: have something that triggers a check to see if the node is still labeled
				// after silencer expiration
			}
		case watch.Deleted:
			logger.Infow("label removed", "node", event.Object.(*v1.Node).Name)
			if err = alertmanager.DeleteSilence(ctx, amcli, silencedID); err != nil {
				logger.Fatalw("error deleting silence", "error", err)
			}
		}
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)

	recvSig := <-sigCh
	signal.Stop(sigCh)
	cancel()
	logger.Infof("exiting. Performing necessary cleanup", recvSig)

}

func validateURL(u *url.URL) error {
	if u.Scheme == "" {
		return errors.New("invalid scheme")
	}

	if u.Host == "" {
		return errors.New("missing host")
	}

	return nil
}
