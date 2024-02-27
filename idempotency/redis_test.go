package idempotency

import (
	"context"
	"github.com/google/uuid"
	"github.com/kanthorlabs/common/idempotency/config"
	"github.com/kanthorlabs/common/testdata"
	"github.com/kanthorlabs/common/testify"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestRedis(t *testing.T) {
	testconf := &config.Config{
		Uri:        os.Getenv("REDIS_URI"),
		TimeToLive: testdata.Fake.IntBetween(10000, 100000),
	}
	if testconf.Uri == "" {
		testconf.Uri = "redis://localhost:6379/0"
	}

	t.Run("New", func(st *testing.T) {
		st.Run("KO - configuration error", func(sst *testing.T) {
			conf := &config.Config{}
			_, err := NewRedis(conf, testify.Logger())
			require.ErrorContains(st, err, "IDEMPOTENCY.CONFIG.")
		})
	})

	t.Run(".Connect/.Readiness/.Liveness/.Disconnect", func(st *testing.T) {
		c, err := NewRedis(testconf, testify.Logger())
		require.Nil(st, err)

		require.ErrorIs(st, c.Readiness(), ErrNotConnected)
		require.ErrorIs(st, c.Liveness(), ErrNotConnected)

		require.Nil(st, c.Connect(context.Background()))

		require.ErrorIs(st, c.Connect(context.Background()), ErrAlreadyConnected)

		require.Nil(st, c.Readiness())
		require.Nil(st, c.Liveness())

		require.Nil(st, c.Disconnect(context.Background()))

		require.Nil(st, c.Readiness())
		require.Nil(st, c.Liveness())

		require.ErrorIs(st, c.Disconnect(context.Background()), ErrNotConnected)
	})

	t.Run(".Validate", func(st *testing.T) {
		c, err := NewRedis(testconf, testify.Logger())
		require.Nil(st, err)
		c.Connect(context.Background())
		defer c.Disconnect(context.Background())

		st.Run("OK", func(sst *testing.T) {
			key := uuid.NewString()
			err := c.Validate(context.Background(), key)
			require.Nil(st, err)
		})

		st.Run("KO", func(sst *testing.T) {
			key := uuid.NewString()
			require.Nil(st, c.Validate(context.Background(), key))
			require.ErrorIs(st, c.Validate(context.Background(), key), ErrConflict)
		})
	})
}
