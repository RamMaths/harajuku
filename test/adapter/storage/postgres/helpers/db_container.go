package helpers

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type TestDBContainer struct {
	URI      string
	Teardown func()
	PORT 		 string
}

func SetupTestDB(t *testing.T) *TestDBContainer {
	t.Helper()

	ctx := context.Background()

	// Get an available host port
  hostPort, err := getAvailablePort()

	if err != nil {
		t.Fatalf("failed to get an available port: %v", err)
	}

	req := testcontainers.ContainerRequest{
		Image:        "postgres:15-alpine",
		ExposedPorts: []string{fmt.Sprintf("%s/tcp", hostPort)},
		Env: map[string]string{
			"POSTGRES_PASSWORD": "secret",
			"POSTGRES_USER":     "user",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp").WithStartupTimeout(30 * time.Second),
		HostConfigModifier: func(config *container.HostConfig) {
			config.PortBindings = nat.PortMap{
				"5432/tcp": []nat.PortBinding{
					{
						HostIP:   "0.0.0.0",
						HostPort: hostPort,
					},
				},
			}
		},
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("failed to start container: %v", err)
	}

	uri := fmt.Sprintf("postgres://user:secret@localhost:%s/testdb?sslmode=disable", hostPort)

	return &TestDBContainer{
		URI: uri,
		Teardown: func() {
			_ = container.Terminate(ctx)
		},
		PORT: hostPort,
	}
}

func getAvailablePort() (string, error) {
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		return "", err
	}
	defer l.Close()
	
	_, port, err := net.SplitHostPort(l.Addr().String())
	return port, err
}
