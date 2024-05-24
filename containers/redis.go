package containers

import (
	"context"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"
)

func Redis(ctx context.Context, name string) (*redis.RedisContainer, error) {
	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Name:         "kanthorlabs-common-redis",
			Image:        "redis:7-alpine",
			ExposedPorts: []string{"6379/tcp"},
			WaitingFor:   wait.ForLog("* Ready to accept connections"),
		},
		Started: true,
		Reuse:   true,
	}
	if name != "" {
		req.ContainerRequest.Name = name
	}

	container, err := testcontainers.GenericContainer(ctx, req)
	if err != nil {
		return nil, err
	}

	return &redis.RedisContainer{Container: container}, nil
}

func RedisConnectionString(container *redis.RedisContainer) (uri string, err error) {
	for i := 0; i < 10; i++ {
		uri, err = container.ConnectionString(context.Background())
		if err == nil {
			return
		}
	}

	return
}
