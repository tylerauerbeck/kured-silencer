package server_test

import (
	"context"
	"fmt"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tylerauerbeck/kured-silencer/pkg/alertmanager"
	"github.com/tylerauerbeck/kured-silencer/pkg/server"

	"go.uber.org/zap"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/fake"
)

func TestValidateURL(t *testing.T) {
	type testCase struct {
		name           string
		url            string
		expectedErrors []error
	}

	testCases := []testCase{
		{
			name: "valid http",
			url:  "http://localhost:9093",
		},
		{
			name: "valid https",
			url:  "https://localhost:9093",
		},
		{
			name:           "invalid scheme",
			url:            "ftp://localhost:9093",
			expectedErrors: []error{server.ErrInvalidScheme},
		},
		{
			name:           "invalid url",
			url:            "http://",
			expectedErrors: []error{server.ErrMissingHost},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			url, err := url.Parse(tc.url)
			if err != nil {
				t.Fatal(err)
			}

			err = server.ValidateURL(url)

			if len(tc.expectedErrors) > 0 {
				assert.Error(t, err)
				for _, expectedError := range tc.expectedErrors {
					assert.ErrorIs(t, err, expectedError)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetKubeClient(t *testing.T) {
	srv := server.Server{
		Client: &server.Client{
			KubeClient: fake.NewSimpleClientset(),
		},
	}

	config := srv.GetKubeClient()

	assert.NotNil(t, config)
}

func TestEventHandler(t *testing.T) {
	ctx := context.Background()

	endpoint, err := AMContainer.Endpoint(ctx, "")
	if err != nil {
		t.Error(err)
	}

	u, err := url.Parse(fmt.Sprintf("http://%s", endpoint))
	if err != nil {
		t.Error(err)
	}

	amc := alertmanager.NewSilencerClient(context.TODO(), u)

	srv := server.Server{
		Client: &server.Client{
			KubeClient: fake.NewSimpleClientset(),
			AMClient:   amc,
		},
	}.WithLogger(ctx, zap.NewNop().Sugar()).WithSilenceDuration(ctx, 5)

	event := watch.Event{}

	err = srv.EventHandler(ctx, event)
	assert.NoError(t, err)

	event.Type = watch.Added
	event.Object = &v1.Node{}

	err = srv.EventHandler(ctx, event)
	assert.NoError(t, err)

	event.Type = watch.Deleted

	assert.NoError(t, err)
}
