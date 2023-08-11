package server

import (
	"time"

	"github.com/prometheus/alertmanager/api/v2/client"
	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"
)

// Client is a struct container the kubernetes and alertmanager clients
type Client struct {
	KubeClient kubernetes.Interface
	AMClient   *client.AlertmanagerAPI
}

// Server contains settings for kured-silencer
type Server struct {
	Client *Client
	// kubeClient      *kubernetes.Interface
	// amClient        *client.AlertmanagerAPI
	logger          *zap.SugaredLogger
	removalBuffer   time.Duration
	silenceDuration time.Duration

	// silencedID string
}
