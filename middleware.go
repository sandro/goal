package goal

import (
	"context"
	"net/http"
	"time"
)

type GoalHttpWriter struct {
	http.ResponseWriter
	Request     *http.Request
	startAt     time.Time
	RecordExtra RecordExtra
}

func (o *GoalHttpWriter) WriteHeader(code int) {
	o.ResponseWriter.WriteHeader(code)
}

func (o *GoalHttpWriter) Write(p []byte) (int, error) {
	n, err := o.ResponseWriter.Write(p)
	o.RecordExtra.ResponseTime = time.Since(o.startAt).Milliseconds()
	Record(o.Request, o.RecordExtra)
	return n, err
}

type GoalMiddlewareConfig struct {
	Next      func(r *http.Request) bool
	VisitorID func(r *http.Request) string
}

var GoalMiddlewareConfigDefault = GoalMiddlewareConfig{
	Next:      nil,
	VisitorID: nil,
}

func goalMiddlewareConfigDefault(config ...GoalMiddlewareConfig) GoalMiddlewareConfig {
	if len(config) < 1 {
		return GoalMiddlewareConfigDefault
	}
	return config[0]
}

func GoalMiddleware(config ...GoalMiddlewareConfig) func(http.Handler) http.Handler {
	cfg := goalMiddlewareConfigDefault(config...)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if cfg.Next != nil && cfg.Next(r) {
				next.ServeHTTP(w, r)
				return
			}
			extra := RecordExtra{
				HitID: GenHitID(),
			}
			if cfg.VisitorID != nil {
				extra.VisitorID = cfg.VisitorID(r)
			}
			ww := GoalHttpWriter{ResponseWriter: w, Request: r, startAt: time.Now(), RecordExtra: extra}
			ctx := context.WithValue(r.Context(), Config.GoalIDKey, ww.RecordExtra.HitID)
			next.ServeHTTP(&ww, r.WithContext(ctx))
		})
	}
}
