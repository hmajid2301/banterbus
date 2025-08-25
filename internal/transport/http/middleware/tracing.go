package middleware

import (
	"fmt"
	"net/http"

	"github.com/getsentry/sentry-go"
)

func (Middleware) Tracing(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		opts := []sentry.SpanOption{
			sentry.WithOpName("http.server"),
			sentry.ContinueFromRequest(r),
			sentry.WithTransactionSource(sentry.SourceURL),
		}
		txn := sentry.StartTransaction(ctx,
			fmt.Sprintf("%s %s", r.Method, r.URL.Path),
			opts...,
		)
		defer txn.Finish()
		h.ServeHTTP(w, r.WithContext(txn.Context()))
	})
}
