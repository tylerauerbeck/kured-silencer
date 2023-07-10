package alertmanager_test

import (
	"context"
	"fmt"
	"net/url"
	"testing"

	"github.com/go-openapi/strfmt"
	"github.com/stretchr/testify/assert"

	"github.com/prometheus/alertmanager/api/v2/client"
	"github.com/prometheus/alertmanager/api/v2/client/silence"

	"github.com/tylerauerbeck/kured-silencer/pkg/alertmanager"
)

func TestSilenceWorkflow(t *testing.T) {
	endpoint, err := AMContainer.Endpoint(context.Background(), "")
	if err != nil {
		t.Error(err)
	}

	u, err := url.Parse(fmt.Sprintf("http://%s", endpoint))
	assert.NoError(t, err)

	c := alertmanager.NewSilencerClient(context.TODO(), u)
	assert.NotNil(t, c)

	ctx := context.Background()
	id, err := alertmanager.PostSilence(ctx, c, 5)
	assert.NoError(t, err)
	assert.NotEmpty(t, id)

	s, err := getSilence(ctx, c, id)
	assert.NoError(t, err)
	assert.NotNil(t, s)
	assert.Equal(t, len(s.Payload.Matchers), 2)

	err = alertmanager.DeleteSilence(ctx, c, id)
	assert.NoError(t, err)

	s, err = getSilence(ctx, c, id)
	assert.Equal(t, *s.Payload.Status.State, "expired")
	assert.NoError(t, err)
}

func getSilence(ctx context.Context, cli *client.AlertmanagerAPI, id string) (*silence.GetSilenceOK, error) {
	params := silence.NewGetSilenceParamsWithContext(ctx).
		WithSilenceID(strfmt.UUID(id))

	s, err := cli.Silence.GetSilence(params)
	if err != nil {
		return nil, err
	}

	return s, nil
}
