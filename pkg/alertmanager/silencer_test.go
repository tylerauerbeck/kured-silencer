package alertmanager_test

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tylerauerbeck/kured-silencer/pkg/alertmanager"
)

func TestNewSilencerClient(t *testing.T) {
	u, err := url.Parse("http://localhost:9093")
	assert.NoError(t, err)

	c := alertmanager.NewSilencerClient(u)
	assert.NotNil(t, c)
}
