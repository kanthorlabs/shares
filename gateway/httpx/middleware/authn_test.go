package middleware

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/kanthorlabs/common/cipher/password"
	"github.com/kanthorlabs/common/gateway/httpx/writer"
	"github.com/kanthorlabs/common/mocks/passport"
	"github.com/kanthorlabs/common/mocks/passport/strategies"
	ppentities "github.com/kanthorlabs/common/passport/entities"
	"github.com/kanthorlabs/common/safe"
	"github.com/kanthorlabs/common/testdata"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var (
	user        = uuid.NewString()
	pass        = uuid.NewString()
	hash, _     = password.HashString(pass)
	basic       = base64.StdEncoding.EncodeToString([]byte(user + ":" + pass))
	credentials = &ppentities.Credentials{Username: user, Password: pass}
	account     = &ppentities.Account{
		Username:     user,
		PasswordHash: hash,
		Metadata:     &safe.Metadata{}}
)

func TestAuthn(t *testing.T) {

	t.Run("OK - fallback", func(st *testing.T) {
		name := uuid.NewString()
		s, path, authn, strategy := authentication(st, name)

		req, err := http.NewRequest(http.MethodGet, path, nil)
		require.NoError(st, err)

		req.Header.Set(HeaderAuthnCredentials, "basic "+basic)

		authn.EXPECT().
			Strategy(name).
			Return(strategy, nil).
			Once()
		strategy.EXPECT().
			ParseCredentials(mock.Anything, "basic "+basic).
			Return(credentials, nil).
			Once()
		strategy.EXPECT().
			Verify(mock.Anything, credentials).
			Return(account, nil).
			Once()

		res := httptest.NewRecorder()
		s.ServeHTTP(res, req)

		require.Equal(st, http.StatusOK, res.Code)
	})

	t.Run("OK - set via header", func(st *testing.T) {
		name := uuid.NewString()
		s, path, authn, strategy := authentication(st, name)

		req, err := http.NewRequest(http.MethodGet, path, nil)
		require.NoError(st, err)

		setname := uuid.NewString()
		req.Header.Set(HeaderAuthnStrategy, setname)
		req.Header.Set(HeaderAuthnCredentials, "basic "+basic)

		authn.EXPECT().
			Strategy(setname).
			Return(strategy, nil).
			Once()
		strategy.EXPECT().
			ParseCredentials(mock.Anything, "basic "+basic).
			Return(credentials, nil).
			Once()
		strategy.EXPECT().
			Verify(mock.Anything, credentials).
			Return(account, nil).
			Once()

		res := httptest.NewRecorder()
		s.ServeHTTP(res, req)

		require.Equal(st, http.StatusOK, res.Code)
	})

	t.Run("KO - unknown strategy", func(st *testing.T) {
		name := uuid.NewString()
		s, path, authn, _ := authentication(st, name)

		req, err := http.NewRequest(http.MethodGet, path, nil)
		require.NoError(st, err)

		setname := uuid.NewString()
		req.Header.Set(HeaderAuthnStrategy, setname)
		req.Header.Set(HeaderAuthnCredentials, "basic "+basic)

		authn.EXPECT().
			Strategy(setname).
			Return(nil, testdata.ErrGeneric).
			Once()

		res := httptest.NewRecorder()
		s.ServeHTTP(res, req)

		require.Equal(st, http.StatusUnauthorized, res.Code)

		var body writer.E
		err = json.Unmarshal(res.Body.Bytes(), &body)
		require.NoError(st, err)

		require.Contains(st, body.Error, testdata.ErrGeneric.Error())
	})

	t.Run("KO - parse credentials error", func(st *testing.T) {
		name := uuid.NewString()
		s, path, authn, strategy := authentication(st, name)

		req, err := http.NewRequest(http.MethodGet, path, nil)
		require.NoError(st, err)

		req.Header.Set(HeaderAuthnCredentials, "basic "+basic)

		authn.EXPECT().
			Strategy(name).
			Return(strategy, nil).
			Once()
		strategy.EXPECT().
			ParseCredentials(mock.Anything, "basic "+basic).
			Return(nil, testdata.ErrGeneric).
			Once()

		res := httptest.NewRecorder()
		s.ServeHTTP(res, req)

		require.Equal(st, http.StatusUnauthorized, res.Code)

		var body writer.E
		err = json.Unmarshal(res.Body.Bytes(), &body)
		require.NoError(st, err)

		require.Contains(st, body.Error, testdata.ErrGeneric.Error())
	})

	t.Run("KO - verify error", func(st *testing.T) {
		name := uuid.NewString()
		s, path, authn, strategy := authentication(st, name)

		req, err := http.NewRequest(http.MethodGet, path, nil)
		require.NoError(st, err)

		req.Header.Set(HeaderAuthnCredentials, "basic "+basic)

		authn.EXPECT().
			Strategy(name).
			Return(strategy, nil).
			Once()
		strategy.EXPECT().
			ParseCredentials(mock.Anything, "basic "+basic).
			Return(credentials, nil).
			Once()
		strategy.EXPECT().
			Verify(mock.Anything, credentials).
			Return(nil, testdata.ErrGeneric).
			Once()

		res := httptest.NewRecorder()
		s.ServeHTTP(res, req)

		require.Equal(st, http.StatusUnauthorized, res.Code)

		var body writer.E
		err = json.Unmarshal(res.Body.Bytes(), &body)
		require.NoError(st, err)

		require.Contains(st, body.Error, testdata.ErrGeneric.Error())
	})
}

func authentication(t *testing.T, name string) (http.Handler, string, *passport.Passport, *strategies.Strategy) {
	s := chi.NewRouter()
	authn := passport.NewPassport(t)
	strategy := strategies.NewStrategy(t)

	s.Use(Authn(authn, name))

	path := "/protected"
	s.Get(path, func(w http.ResponseWriter, r *http.Request) {
		writer.Ok(w, writer.M{})
	})

	return s, path, authn, strategy
}
