package chilogger

import (
	"net/http"
	"time"

	"api/pkg/log"

	"github.com/go-chi/chi/middleware"
	"go.uber.org/zap"
)

// Logger is a middleware that logs the start and end of each request, along
// with some useful data about what was requested, what the response status was,
// and how long it took to return.
func Logger() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {

			start := time.Now()

			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			l := log.LoggerWithLambdaRqID(r.Context())

			defer func() {

				l.Info("request complete",
					zap.String("proto", r.Proto),
					zap.String("path", r.URL.Path),
					zap.String("remote", r.RemoteAddr),
					zap.Duration("took", time.Since(start)),
					zap.Int("status", ww.Status()),
					zap.Int("size", ww.BytesWritten()),
				)
			}()

			next.ServeHTTP(ww, r)
		}
		return http.HandlerFunc(fn)
	}
}
