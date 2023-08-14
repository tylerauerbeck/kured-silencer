package server

import (
	"context"
	"net/url"
	"os"
	"time"

	"github.com/spf13/viper"

	"github.com/tylerauerbeck/kured-silencer/pkg/alertmanager"
	"github.com/tylerauerbeck/kured-silencer/pkg/kube"

	"go.uber.org/zap"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
)

var (
	silenceIDs           = make(map[string][]string)
	defaultLeaseDuration = 15 * time.Second
	defaultRenewDeadline = 10 * time.Second
	defaultRetryPeriod   = 2 * time.Second
	leaseLockName        = "kured-silencer"
	leaseLockNamespace   = os.Getenv("POD_NAMESPACE")
	podName              = os.Getenv("POD_NAME")
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
		removalBuffer:   viper.GetDuration("removal-buffer"),
	}

	return srv, nil
}

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
		silencedIDs, err := alertmanager.PostSilence(ctx, srv.Client.AMClient, srv.silenceDuration)
		if err != nil {
			if len(silencedIDs) > 0 {
				for _, id := range silencedIDs {
					if err := alertmanager.DeleteSilence(ctx, srv.Client.AMClient, id); err != nil {
						return err
					}
				}
			}

			return err
		}

		silenceIDs[event.Object.(*v1.Node).Name] = silencedIDs

		srv.logger.Infow("label added", "node", event.Object.(*v1.Node).Name)

		return nil
	case watch.Deleted:
		// TODO: probably a better way to do this, but we're finding that we get alerted once
		// the silence is removed because there are alerts that haven't cleared. This is a
		// configurable period of time, but it would be better to have a smarter way to handle
		// this

		time.Sleep(srv.removalBuffer)

		ids, ok := silenceIDs[event.Object.(*v1.Node).Name]
		if ok {
			for _, id := range ids {
				if err := alertmanager.DeleteSilence(ctx, srv.Client.AMClient, id); err != nil {
					// TODO: emit metric that we failed to delete a set of silences
					return err
				}
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
func (srv *Server) Run(ctx context.Context) {
	if viper.GetString("kubeconfig-path") == "" {
		client := srv.GetKubeClient().(*kubernetes.Clientset)
		lock := getNewLock(client, leaseLockName, podName, leaseLockNamespace)
		srv.runLeaderElection(ctx, lock, os.Getenv("POD_NAME"))
	} else {
		for {
			if err := srv.watcherRun(ctx); err != nil {
				srv.logger.Infow("restarting watcher...", "error", err.Error())
			}
		}
	}
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

func getNewLock(client *kubernetes.Clientset, lockname, podname, namespace string) *resourcelock.LeaseLock {
	return &resourcelock.LeaseLock{
		LeaseMeta: metav1.ObjectMeta{
			Name:      lockname,
			Namespace: namespace,
		},
		Client: client.CoordinationV1(),
		LockConfig: resourcelock.ResourceLockConfig{
			Identity: podname,
		},
	}
}

func (srv *Server) runLeaderElection(ctx context.Context, lock *resourcelock.LeaseLock, id string) {
	leaderelection.RunOrDie(ctx, leaderelection.LeaderElectionConfig{
		Lock:            lock,
		ReleaseOnCancel: true,
		LeaseDuration:   defaultLeaseDuration,
		RenewDeadline:   defaultRenewDeadline,
		RetryPeriod:     defaultRetryPeriod,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: func(c context.Context) {
				for {
					if err := srv.watcherRun(ctx); err != nil {
						srv.logger.Info("Watcher closed", "error", err.Error())
					}
				}
			},
			OnStoppedLeading: func() {
				srv.logger.Info("new leader elected, stepping down...")
			},
			OnNewLeader: func(current_id string) {
				if current_id == id {
					srv.logger.Debug("re-elected as leader, continuing...")
					return
				}
				srv.logger.Infow("new leader elected", "leader", current_id)
			},
		},
	})
}

func (srv *Server) watcherRun(ctx context.Context) error {

	watcher, err := kube.NewNodeWatcher(ctx, srv.GetKubeClient(), viper.GetString("kured-label"))
	if err != nil {
		return err
	}

	for {
		select {
		case event, ok := <-watcher.ResultChan():
			if !ok {
				srv.logger.Info("refreshing watcher...")
				watcher.Stop()
				if watcher, err = kube.NewNodeWatcher(ctx, srv.GetKubeClient(), viper.GetString("kured-label")); err != nil {
					return err
				}

				continue
			}

			srv.EventHandler(ctx, event)
		}
	}
}
