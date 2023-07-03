package alertmanager

import (
	"context"
	"net/url"
	"path"
	"time"

	runtimeclient "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"

	"github.com/tylerauerbeck/kured-silencer/pkg/internal/utils"

	"github.com/prometheus/alertmanager/api/v2/client"
	"github.com/prometheus/alertmanager/api/v2/client/silence"
	"github.com/prometheus/alertmanager/api/v2/models"
)

// NewSilencerClient returns a new alertmanager client pointed at the specified url
func NewSilencerClient(_ context.Context, u *url.URL) *client.AlertmanagerAPI {
	return client.New(
		runtimeclient.New(u.Host, path.Join(u.Path, "/api/v2"), []string{u.Scheme}),
		strfmt.Default,
	)
}

// PostSilence creates a new silence for all warning and critical alerts
func PostSilence(ctx context.Context, cli *client.AlertmanagerAPI, duration time.Duration) (string, error) {
	// TODO: allow for defining matchers?
	ms := []*models.Matcher{
		{
			IsRegex: utils.NewBool(true),
			Name:    utils.NewString("severity"),
			Value:   utils.NewString("warning"),
		},
		{
			IsRegex: utils.NewBool(true),
			Name:    utils.NewString("severity"),
			Value:   utils.NewString("critical"),
		},
	}

	params := silence.NewPostSilencesParamsWithContext(ctx).
		WithSilence(&models.PostableSilence{
			Silence: models.Silence{
				StartsAt:  utils.NewDateTime(strfmt.DateTime(time.Now())),
				EndsAt:    utils.NewDateTime(strfmt.DateTime(time.Now().Add(duration * time.Minute))),
				Comment:   utils.NewString("silenced for kured reboot"),
				CreatedBy: utils.NewString("kured-silencer"),
				Matchers:  ms,
			},
		})

	// TODO: maybe break up the generation of params and actually calling the API?

	// TODO: lets not add another silencer if there is already one in place

	id, err := cli.Silence.PostSilences(params)
	if err != nil {
		return "", err
	}

	return id.Payload.SilenceID, nil
}

// DeleteSilence deletes the silence with the specified id
func DeleteSilence(ctx context.Context, cli *client.AlertmanagerAPI, id string) error {
	params := silence.NewDeleteSilenceParamsWithContext(ctx).
		WithSilenceID(strfmt.UUID(id))

	if _, err := cli.Silence.DeleteSilence(params); err != nil {
		return err
	}

	return nil
}
