package distributedlockmanager

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/kanthorlabs/common/distributedlockmanager/config"
	"github.com/kanthorlabs/common/testdata"
	"github.com/stretchr/testify/require"
)

func TestMemory_New(t *testing.T) {
	t.Run("OK", func(st *testing.T) {
		_, err := NewMemory(testconf)
		require.NoError(st, err)
	})

	t.Run("KO - configuration error", func(st *testing.T) {
		conf := &config.Config{}
		_, err := NewMemory(conf)
		require.ErrorContains(st, err, "DISTRIBUTED_LOCK_MANAGER.CONFIG.")
	})
}

func TestMemory_Lock(t *testing.T) {
	dlm, err := NewMemory(testconf)
	require.NoError(t, err)

	t.Run("OK", func(st *testing.T) {
		key := uuid.NewString()
		locker := dlm(key, config.TimeToLive(testdata.Fake.UInt64Between(10000, 100000)))

		err = locker.Lock(context.Background())
		require.NoError(st, err)
	})

	t.Run("KO - key empty error", func(st *testing.T) {
		locker := dlm("")

		err = locker.Lock(context.Background())
		require.ErrorIs(st, err, ErrKeyEmpty)
	})

	t.Run("KO - key already locked error", func(st *testing.T) {
		key := uuid.NewString()
		locker := dlm(key, config.TimeToLive(testdata.Fake.UInt64Between(10000, 100000)))

		err = locker.Lock(context.Background())
		require.NoError(st, err)

		err = locker.Lock(context.Background())
		require.ErrorContains(st, err, "DISTRIBUTED_LOCK_MANAGER.LOCK.ERROR")
	})
}

func TestMemory_Unlock(t *testing.T) {
	dlm, err := NewMemory(testconf)
	require.NoError(t, err)

	t.Run("OK", func(st *testing.T) {
		key := uuid.NewString()
		locker := dlm(key, config.TimeToLive(testdata.Fake.UInt64Between(10000, 100000)))

		err = locker.Lock(context.Background())
		require.NoError(st, err)

		err = locker.Unlock(context.Background())
		require.NoError(st, err)
	})

	t.Run("KO - key empty error", func(st *testing.T) {
		locker := dlm("")

		err = locker.Unlock(context.Background())
		require.ErrorIs(st, err, ErrKeyEmpty)
	})

	t.Run("KO - key not locked error", func(st *testing.T) {
		key := uuid.NewString()
		locker := dlm(key, config.TimeToLive(testdata.Fake.UInt64Between(10000, 100000)))

		err = locker.Unlock(context.Background())
		require.ErrorContains(st, err, "DISTRIBUTED_LOCK_MANAGER.UNLOCK.ERROR")
	})
}
