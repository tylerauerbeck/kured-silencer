package server

import "errors"

var (
	// ErrMissingHost is returned when the URL is missing a scheme
	ErrMissingHost = errors.New("missing host")

	// ErrInvalidScheme is returned when the URL has an invalid scheme
	ErrInvalidScheme = errors.New("invalid scheme")

	// ErrMissingHost is returned when the URL is missing a host
	// ErrMissingScheme = errors.New("missing scheme")

	// ErrMissingNode is returned when the node a label is deleted from is not found
	ErrMissingNode = errors.New("missing node")
)
