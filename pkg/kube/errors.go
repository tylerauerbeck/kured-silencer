package kube

import "errors"

var (
	// errInvalidKubeClient is returned when the kube client is invalid.
	errInvalidKubeConfig = errors.New("invalid kube client")

	// errInvalidKubeConfig is returned when the kube config is invalid.
	errMissingKubeConfig = errors.New("running outside of cluster and kubeconfig flag is not set")

	// errInvalidKubeClient is returned when the kube client is invalid.
	errInvalidKubeClient = errors.New("invalid kube client")
)
