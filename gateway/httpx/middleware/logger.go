package middleware

import (
	"io"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/kanthorlabs/common/logging"
	"github.com/kanthorlabs/common/utils"
)

func Logger(logger logging.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			defer func() {
				args := []any{
					"response_status", ww.Status(),
					"request_uri", r.URL.String(),
					"request_headers", utils.Stringify(r.Header),
				}

				if ww.Status() >= http.StatusInternalServerError {
					// body is large in some case, don't parse it until we got server error
					body, _ := io.ReadAll(r.Body)
					args = append(args, "request_body", body)
					logger.Errorw("GATEWAY.REQUEST.ERROR", args...)
					return
				}

				logger.Debugw("GATEWAY.REQUEST", args...)

			}()

			next.ServeHTTP(ww, r)
		})
	}
}
