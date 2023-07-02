package utils_test

import (
	"testing"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/stretchr/testify/assert"
	"github.com/tylerauerbeck/kured-silencer/pkg/internal/utils"
)

func TestNewBool(t *testing.T) {
	b := utils.NewBool(true)
	assert.True(t, *b)
	assert.IsType(t, new(bool), b)

	b = utils.NewBool(false)
	assert.False(t, *b)
	assert.IsType(t, new(bool), b)
}

func TestNewString(t *testing.T) {
	s := utils.NewString("test")
	assert.Equal(t, "test", *s)
	assert.IsType(t, new(string), s)
}

func TestNewDateTime(t *testing.T) {
	d := utils.NewDateTime(strfmt.DateTime(time.Now()))
	assert.Equal(t, strfmt.DateTime(time.Now()).String(), d.String())
	assert.IsType(t, new(strfmt.DateTime), d)
}
