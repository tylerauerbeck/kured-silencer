package kube_test

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tylerauerbeck/kured-silencer/pkg/kube"
	"github.com/tylerauerbeck/kured-silencer/pkg/server"

	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/fake"
)

func TestNewNodeWatcher(t *testing.T) {
	s := server.Server{
		Client: &server.Client{
			KubeClient: fake.NewSimpleClientset(),
		},
	}

	watcher, err := kube.NewNodeWatcher(context.TODO(), s.Client.KubeClient, "hello=world")
	assert.NoError(t, err)
	assert.IsType(t, &watch.RaceFreeFakeWatcher{}, watcher)

	s.Client.KubeClient = &fake.Clientset{}

	watcher, err = kube.NewNodeWatcher(context.TODO(), s.Client.KubeClient, "hello=world")
	assert.Error(t, err)
	assert.Nil(t, watcher)
}

func TestNewKubeClient(t *testing.T) {
	type testCase struct {
		name           string
		path           string
		expectedErrors []error
	}

	pwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	testCases := []testCase{
		{
			name: "valid path - valid config",
			path: pwd + "/../../hack/ci/testdata/kubeconfig-valid",
		},
		{
			name:           "valid path - no config",
			path:           pwd + "/../../hack/ci/testdata/kubeconfig-empty",
			expectedErrors: []error{kube.ErrInvalidKubeConfig},
		},
		{
			name:           "invalid path",
			path:           "",
			expectedErrors: []error{kube.ErrMissingKubeConfig},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := kube.NewKubeClient(context.TODO(), tc.path)

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
