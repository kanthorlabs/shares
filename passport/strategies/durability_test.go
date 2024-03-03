package strategies

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kanthorlabs/common/cipher/password"
	"github.com/kanthorlabs/common/passport/config"
	"github.com/kanthorlabs/common/passport/entities"
	sqlx "github.com/kanthorlabs/common/persistence/sqlx/config"
	"github.com/kanthorlabs/common/testdata"
	"github.com/kanthorlabs/common/testify"
	"github.com/stretchr/testify/require"
)

func TestDurability_New(t *testing.T) {
	t.Run("OK", func(st *testing.T) {
		conf := &config.Durability{Sqlx: sqlx.Config{
			Uri: testdata.SqliteUri,
			Connection: sqlx.Connection{
				MaxLifetime:  sqlx.DefaultConnMaxLifetime,
				MaxIdletime:  sqlx.DefaultConnMaxIdletime,
				MaxIdleCount: sqlx.DefaultConnMaxIdleCount,
				MaxOpenCount: sqlx.DefaultConnMaxOpenCount,
			},
		}}
		_, err := NewDurability(conf, testify.Logger())
		require.NoError(st, err)
	})

	t.Run("KO - configuration error", func(st *testing.T) {
		_, err := NewDurability(&config.Durability{}, testify.Logger())
		require.ErrorContains(st, err, "SQLX.CONFIG")
	})

	t.Run("KO - sqlx error", func(st *testing.T) {
		conf := &config.Durability{Sqlx: sqlx.Config{}}
		_, err := NewDurability(conf, testify.Logger())
		require.ErrorContains(st, err, "SQLX.CONFIG")
	})
}

func TestDurability_Connect(t *testing.T) {
	conf := &config.Durability{Sqlx: sqlx.Config{
		Uri: testdata.SqliteUri,
		Connection: sqlx.Connection{
			MaxLifetime:  sqlx.DefaultConnMaxLifetime,
			MaxIdletime:  sqlx.DefaultConnMaxIdletime,
			MaxIdleCount: sqlx.DefaultConnMaxIdleCount,
			MaxOpenCount: sqlx.DefaultConnMaxOpenCount,
		},
	}}

	t.Run("OK", func(st *testing.T) {
		c, err := NewDurability(conf, testify.Logger())
		require.NoError(st, err)

		require.NoError(st, c.Connect(context.Background()))
	})

	t.Run("KO - already connected", func(st *testing.T) {
		c, err := NewDurability(conf, testify.Logger())
		require.NoError(st, err)

		require.NoError(st, c.Connect(context.Background()))
		require.ErrorIs(st, c.Connect(context.Background()), ErrAlreadyConnected)
	})
}

func TestDurability_Readiness(t *testing.T) {
	conf := &config.Durability{Sqlx: sqlx.Config{
		Uri: testdata.SqliteUri,
		Connection: sqlx.Connection{
			MaxLifetime:  sqlx.DefaultConnMaxLifetime,
			MaxIdletime:  sqlx.DefaultConnMaxIdletime,
			MaxIdleCount: sqlx.DefaultConnMaxIdleCount,
			MaxOpenCount: sqlx.DefaultConnMaxOpenCount,
		},
	}}

	t.Run("OK", func(st *testing.T) {
		c, err := NewDurability(conf, testify.Logger())
		require.NoError(st, err)

		require.NoError(st, c.Connect(context.Background()))
		require.NoError(st, c.Readiness())
	})

	t.Run("OK - disconnected", func(st *testing.T) {
		c, err := NewDurability(conf, testify.Logger())
		require.NoError(st, err)

		require.NoError(st, c.Connect(context.Background()))
		require.NoError(st, c.Disconnect(context.Background()))
		require.NoError(st, c.Readiness())
	})

	t.Run("KO - not connected", func(st *testing.T) {
		c, err := NewDurability(conf, testify.Logger())
		require.NoError(st, err)

		require.ErrorIs(st, c.Readiness(), ErrNotConnected)
	})
}

func TestDurability_Liveness(t *testing.T) {
	conf := &config.Durability{Sqlx: sqlx.Config{
		Uri: testdata.SqliteUri,
		Connection: sqlx.Connection{
			MaxLifetime:  sqlx.DefaultConnMaxLifetime,
			MaxIdletime:  sqlx.DefaultConnMaxIdletime,
			MaxIdleCount: sqlx.DefaultConnMaxIdleCount,
			MaxOpenCount: sqlx.DefaultConnMaxOpenCount,
		},
	}}

	t.Run("OK", func(st *testing.T) {
		c, err := NewDurability(conf, testify.Logger())
		require.NoError(st, err)

		require.NoError(st, c.Connect(context.Background()))
		require.NoError(st, c.Liveness())
	})

	t.Run("OK - disconnected", func(st *testing.T) {
		c, err := NewDurability(conf, testify.Logger())
		require.NoError(st, err)

		require.NoError(st, c.Connect(context.Background()))
		require.NoError(st, c.Disconnect(context.Background()))
		require.NoError(st, c.Liveness())
	})

	t.Run("KO - not connected", func(st *testing.T) {
		c, err := NewDurability(conf, testify.Logger())
		require.NoError(st, err)

		require.ErrorIs(st, c.Liveness(), ErrNotConnected)
	})
}

func TestDurability_Disconnect(t *testing.T) {
	conf := &config.Durability{Sqlx: sqlx.Config{
		Uri: testdata.SqliteUri,
		Connection: sqlx.Connection{
			MaxLifetime:  sqlx.DefaultConnMaxLifetime,
			MaxIdletime:  sqlx.DefaultConnMaxIdletime,
			MaxIdleCount: sqlx.DefaultConnMaxIdleCount,
			MaxOpenCount: sqlx.DefaultConnMaxOpenCount,
		},
	}}

	t.Run("OK", func(st *testing.T) {
		c, err := NewDurability(conf, testify.Logger())
		require.NoError(st, err)

		require.NoError(st, c.Connect(context.Background()))
		require.NoError(st, c.Disconnect(context.Background()))
	})

	t.Run("KO - not connected", func(st *testing.T) {
		c, err := NewDurability(conf, testify.Logger())
		require.NoError(st, err)

		require.ErrorIs(st, c.Disconnect(context.Background()), ErrNotConnected)
	})
}

func TestDurability_Login(t *testing.T) {
	accounts, passwords := setup(t)

	conf := &config.Durability{Sqlx: sqlx.Config{
		Uri: testdata.SqliteUri,
		Connection: sqlx.Connection{
			MaxLifetime:  sqlx.DefaultConnMaxLifetime,
			MaxIdletime:  sqlx.DefaultConnMaxIdletime,
			MaxIdleCount: sqlx.DefaultConnMaxIdleCount,
			MaxOpenCount: sqlx.DefaultConnMaxOpenCount,
		},
	}}
	strategy, err := NewDurability(conf, testify.Logger())
	require.NoError(t, err)

	strategy.Connect(context.Background())
	defer strategy.Disconnect(context.Background())

	orm := strategy.(*durability).orm
	tx := orm.Create(accounts)
	require.NoError(t, tx.Error)

	t.Run("OK", func(st *testing.T) {
		i := testdata.Fake.IntBetween(0, len(passwords)-1)
		credentials := &entities.Credentials{
			Username: accounts[i].Username,
			Password: passwords[i],
		}
		acc, err := strategy.Login(context.Background(), credentials)
		require.NoError(st, err)
		require.Equal(st, credentials.Username, acc.Username)
		require.Empty(st, acc.PasswordHash)
	})

	t.Run("KO - credentials error", func(st *testing.T) {
		_, err := strategy.Login(context.Background(), nil)
		require.ErrorContains(st, err, "PASSPORT.CREDENTIALS")

		_, err = strategy.Login(context.Background(), &entities.Credentials{})
		require.ErrorContains(st, err, "PASSPORT.CREDENTIALS")
	})

	t.Run("KO - user not found", func(st *testing.T) {
		credentials := &entities.Credentials{
			Username: uuid.NewString(),
			Password: testdata.Fake.Internet().Password(),
		}
		_, err := strategy.Login(context.Background(), credentials)
		require.ErrorIs(st, err, ErrLogin)
	})

	t.Run("KO - password not match", func(st *testing.T) {
		i := testdata.Fake.IntBetween(0, len(passwords)-1)
		credentials := &entities.Credentials{
			Username: accounts[i].Username,
			Password: testdata.Fake.Internet().Password(),
		}
		_, err := strategy.Login(context.Background(), credentials)
		require.ErrorIs(st, err, ErrLogin)
	})
}

func TestDurability_Verify(t *testing.T) {
	accounts, passwords := setup(t)

	conf := &config.Durability{Sqlx: sqlx.Config{
		Uri: testdata.SqliteUri,
		Connection: sqlx.Connection{
			MaxLifetime:  sqlx.DefaultConnMaxLifetime,
			MaxIdletime:  sqlx.DefaultConnMaxIdletime,
			MaxIdleCount: sqlx.DefaultConnMaxIdleCount,
			MaxOpenCount: sqlx.DefaultConnMaxOpenCount,
		},
	}}
	strategy, err := NewDurability(conf, testify.Logger())
	require.NoError(t, err)

	strategy.Connect(context.Background())
	defer strategy.Disconnect(context.Background())

	orm := strategy.(*durability).orm
	tx := orm.Create(accounts)
	require.NoError(t, tx.Error)

	t.Run("OK", func(st *testing.T) {
		i := testdata.Fake.IntBetween(0, len(passwords)-1)
		credentials := &entities.Credentials{
			Username: accounts[i].Username,
			Password: passwords[i],
		}
		acc, err := strategy.Verify(context.Background(), credentials)
		require.NoError(st, err)
		require.Equal(st, credentials.Username, acc.Username)
		require.Empty(st, acc.PasswordHash)
	})

	t.Run("KO - credentials error", func(st *testing.T) {
		_, err := strategy.Verify(context.Background(), nil)
		require.ErrorContains(st, err, "PASSPORT.CREDENTIALS")

		_, err = strategy.Verify(context.Background(), &entities.Credentials{})
		require.ErrorContains(st, err, "PASSPORT.CREDENTIALS")
	})

	t.Run("KO - user not found", func(st *testing.T) {
		credentials := &entities.Credentials{
			Username: uuid.NewString(),
			Password: testdata.Fake.Internet().Password(),
		}
		_, err := strategy.Verify(context.Background(), credentials)
		require.ErrorIs(st, err, ErrLogin)
	})

	t.Run("KO - password not match", func(st *testing.T) {
		i := testdata.Fake.IntBetween(0, len(passwords)-1)
		credentials := &entities.Credentials{
			Username: accounts[i].Username,
			Password: testdata.Fake.Internet().Password(),
		}
		_, err := strategy.Verify(context.Background(), credentials)
		require.ErrorIs(st, err, ErrLogin)
	})
}

func TestDurability_Register(t *testing.T) {
	accounts, _ := setup(t)

	conf := &config.Durability{Sqlx: sqlx.Config{
		Uri: testdata.SqliteUri,
		Connection: sqlx.Connection{
			MaxLifetime:  sqlx.DefaultConnMaxLifetime,
			MaxIdletime:  sqlx.DefaultConnMaxIdletime,
			MaxIdleCount: sqlx.DefaultConnMaxIdleCount,
			MaxOpenCount: sqlx.DefaultConnMaxOpenCount,
		},
	}}
	strategy, err := NewDurability(conf, testify.Logger())
	require.NoError(t, err)

	strategy.Connect(context.Background())
	defer strategy.Disconnect(context.Background())

	orm := strategy.(*durability).orm
	tx := orm.Create(accounts)
	require.NoError(t, tx.Error)

	t.Run("OK", func(st *testing.T) {
		pass := uuid.NewString()
		hash, err := password.HashString(pass)
		require.NoError(st, err)

		acc := &entities.Account{
			Username:     uuid.NewString(),
			PasswordHash: hash,
			Name:         testdata.Fake.Internet().User(),
			CreatedAt:    time.Now().UnixMilli(),
			UpdatedAt:    time.Now().UnixMilli(),
		}

		require.NoError(st, strategy.Register(context.Background(), acc))
	})

	t.Run("KO", func(st *testing.T) {
		i := testdata.Fake.IntBetween(0, len(accounts)-1)
		err := strategy.Register(context.Background(), &accounts[i])
		require.ErrorIs(st, err, ErrRegister)
	})
}

func TestDurability_Deactivate(t *testing.T) {
	accounts, _ := setup(t)

	conf := &config.Durability{Sqlx: sqlx.Config{
		Uri: testdata.SqliteUri,
		Connection: sqlx.Connection{
			MaxLifetime:  sqlx.DefaultConnMaxLifetime,
			MaxIdletime:  sqlx.DefaultConnMaxIdletime,
			MaxIdleCount: sqlx.DefaultConnMaxIdleCount,
			MaxOpenCount: sqlx.DefaultConnMaxOpenCount,
		},
	}}
	strategy, err := NewDurability(conf, testify.Logger())
	require.NoError(t, err)

	strategy.Connect(context.Background())
	defer strategy.Disconnect(context.Background())

	orm := strategy.(*durability).orm
	tx := orm.Create(accounts)
	require.NoError(t, tx.Error)

	t.Run("OK", func(st *testing.T) {
		i := testdata.Fake.IntBetween(0, len(accounts)-1)
		username := accounts[i].Username
		ts := time.Now().UnixMilli()

		err := strategy.Deactivate(context.Background(), username, ts)
		require.NoError(st, err)
	})

	t.Run("KO - user not found", func(st *testing.T) {
		username := uuid.NewString()
		ts := time.Now().UnixMilli()

		err := strategy.Deactivate(context.Background(), username, ts)
		require.ErrorIs(st, err, ErrAccountNotFound)
	})
}
