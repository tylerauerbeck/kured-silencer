package utils

import "github.com/go-openapi/strfmt"

func NewBool(b bool) *bool {
	return &b
}

func NewString(s string) *string {
	return &s
}

func NewDateTime(s strfmt.DateTime) *strfmt.DateTime {
	return &s
}
