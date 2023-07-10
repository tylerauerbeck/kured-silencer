package server

import (
	"context"
	"net/url"
	"time"

	"github.com/spf13/viper"

	"github.com/tylerauerbeck/kured-silencer/pkg/alertmanager"
	"github.com/tylerauerbeck/kured-silencer/pkg/kube"

	"go.uber.org/zap"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
)

var (
	silenceIDs = make(map[string]string)
)

// NewServer creates a new server
func NewServer(ctx context.Context, logger *zap.SugaredLogger) (*Server, error) {
	kcli, err := kube.NewKubeClient(ctx, viper.GetString("kubeconfig-path"))
	if err != nil {
		return nil, err
	}

	url, err := url.Parse(viper.GetString("alertmanager-endpoint"))
	if err != nil {
		return nil, err
	}

	if err = ValidateURL(url); err != nil {
		return nil, err
	}

	amcli := alertmanager.NewSilencerClient(context.TODO(), url)

	srv := &Server{
		Client: &Client{
			KubeClient: kcli,
			AMClient:   amcli,
		},
		logger:          logger,
		silenceDuration: viper.GetDuration("silence-duration"),
	}

	return srv, nil
}

// func (srv Server) setSilencedID(s string) Server {
// 	srv.silencedID = s //nolint:staticcheck
// 	return srv
// }

// WithLogger sets the logger for the server
func (srv Server) WithLogger(_ context.Context, logger *zap.SugaredLogger) *Server {
	srv.logger = logger
	return &srv
}

// WithSilenceDuration sets the silence duration for the server
func (srv Server) WithSilenceDuration(_ context.Context, d time.Duration) *Server {
	srv.silenceDuration = d
	return &srv
}

// GetKubeClient returns the kubernetes client from the running server
func (srv Server) GetKubeClient() kubernetes.Interface {
	return srv.Client.KubeClient
}

// EventHandler provides logic for handling node label event types
func (srv Server) EventHandler(ctx context.Context, event watch.Event) error {
	switch event.Type {
	case watch.Added:
		silencedID, err := alertmanager.PostSilence(ctx, srv.Client.AMClient, srv.silenceDuration)
		if err != nil {
			// TODO: have something that triggers a check to see if the node is still labeled
			// after silencer expiration
			return err
		}

		silenceIDs[event.Object.(*v1.Node).Name] = silencedID

		srv.logger.Infow("label added", "node", event.Object.(*v1.Node).Name)

		return nil
	case watch.Deleted:
		id, ok := silenceIDs[event.Object.(*v1.Node).Name]
		if ok {
			if err := alertmanager.DeleteSilence(ctx, srv.Client.AMClient, id); err != nil {
				return err
			}

			delete(silenceIDs, event.Object.(*v1.Node).Name)

			srv.logger.Infow("label removed", "node", event.Object.(*v1.Node).Name)

			return nil
		}

		return ErrMissingNode
	default:
		return nil
	}
}

// Run starts the server
func (srv *Server) Run(ctx context.Context, watcher watch.Interface) error {
	for event := range watcher.ResultChan() {
		if err := srv.EventHandler(ctx, event); err != nil {
			return err
		}
	}

	return nil
}

// ValidateURL ensures that a valid url with both scheme and host is provided
func ValidateURL(u *url.URL) error {
	if u.Scheme != "http" && u.Scheme != "https" {
		return ErrInvalidScheme
	}

	if u.Host == "" {
		return ErrMissingHost
	}

	return nil
}
