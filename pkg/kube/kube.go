package kube

import (
	"context"
	"errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	timeoutSeconds = int64(600)
)

// NewKubeClient returns a new kubernetes clientset
func NewKubeClient(_ context.Context, path string) (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		if path != "" {
			config, err = clientcmd.BuildConfigFromFlags("", path)
			if err != nil {
				return nil, errors.Join(err, ErrInvalidKubeConfig)
			}
		} else {
			return nil, errors.Join(err, ErrMissingKubeConfig)
		}
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, errors.Join(err, ErrInvalidKubeClient)
	}

	return client, nil
}

// NewNodeWatcher returns a new node watcher for nodes with the specified label
func NewNodeWatcher(ctx context.Context, cli kubernetes.Interface, label string) (watch.Interface, error) {
	// TODO: add bookmarker here so that it's not picking up old labels
	// This is particularly important if it just keeps failing and sees
	// an existing label and then just adds another new silencer
	watcher, err := cli.CoreV1().Nodes().Watch(ctx, metav1.ListOptions{LabelSelector: label, TimeoutSeconds: &timeoutSeconds})
	if err != nil {
		// TODO: errInvalidWatcher
		return nil, err
	}

	return watcher, nil
}
