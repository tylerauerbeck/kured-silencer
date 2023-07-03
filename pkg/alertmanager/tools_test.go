package alertmanager_test

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	AMContainer testcontainers.Container
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	ephemeralAlertManager(ctx)

	code := m.Run()

	os.Exit(code)
}

func ephemeralAlertManager(ctx context.Context) {
	req := testcontainers.ContainerRequest{
		Image:        "prom/alertmanager:latest",
		ExposedPorts: []string{"9093/tcp"},
		WaitingFor:   wait.ForLog("Completed loading of configuration file"),
	}

	amC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		errPanic("unable to create container", err)
	}

	AMContainer = amC
}

func errPanic(msg string, err error) {
	if err != nil {
		log.Panicf("%s err: %s", msg, err.Error())
	}
}
