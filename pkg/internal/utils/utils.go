package utils

import "github.com/go-openapi/strfmt"

// NewBool returns a pointer to a bool
func NewBool(b bool) *bool {
	return &b
}

// NewString returns a pointer to a string
func NewString(s string) *string {
	return &s
}

// NewDateTime returns a pointer to a DateTime
func NewDateTime(s strfmt.DateTime) *strfmt.DateTime {
	return &s
}
