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

func NewKubeClient(path string) (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		if path != "" {
			config, err = clientcmd.BuildConfigFromFlags("", path)
			if err != nil {
				return nil, errors.Join(err, errInvalidKubeConfig)
			}
		} else {
			return nil, errors.Join(err, errMissingKubeConfig)
		}
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, errors.Join(err, errInvalidKubeClient)
	}

	return client, nil
}

func NewNodeWatcher(ctx context.Context, cli *kubernetes.Clientset, label string) (watch.Interface, error) {
	// TODO: add bookmarker here so that it's not picking up old labels
	// This is particularly important if it just keeps failing and sees
	// an existing label and then just adds another new silencer
	watcher, err := cli.CoreV1().Nodes().Watch(ctx, metav1.ListOptions{LabelSelector: label})
	if err != nil {
		return nil, err
	}

	return watcher, nil
}
