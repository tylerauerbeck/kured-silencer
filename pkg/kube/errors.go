package kube

import "errors"

var (
	// ErrInvalidKubeConfig is returned when the kube client is invalid.
	ErrInvalidKubeConfig = errors.New("invalid kube config")

	// ErrMissingKubeConfig is returned when the kube config is invalid.
	ErrMissingKubeConfig = errors.New("running outside of cluster and kubeconfig flag is not set")

	// ErrInvalidKubeClient is returned when the kube client is invalid.
	ErrInvalidKubeClient = errors.New("invalid kube client")
)
