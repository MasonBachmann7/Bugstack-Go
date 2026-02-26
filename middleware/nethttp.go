// Package middleware provides HTTP middleware for capturing errors with BugStack.
package middleware

import (
	"fmt"
	"net/http"
	"net/url"

	bugstack "github.com/MasonBachmann7/bugstack-go"
)

// NetHTTP returns a middleware that captures panics from net/http handlers.
//
// Usage:
//
//	mux := http.NewServeMux()
//	mux.HandleFunc("/", handler)
//	http.ListenAndServe(":8080", middleware.NetHTTP(mux))
func NetHTTP(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				client := bugstack.GetClient()
				if client != nil {
					msg := fmt.Sprintf("panic: %v", rec)
					client.CaptureMessage(msg,
						bugstack.WithRequest(&bugstack.RequestContext{
							Route:       r.URL.Path,
							Method:      r.Method,
							QueryParams: flattenQuery(r.URL.Query()),
						}),
						bugstack.WithMetadata(map[string]any{
							"framework": "net/http",
						}),
					)
				}
				// Re-panic so the server handles it normally
				panic(rec)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// flattenQuery converts url.Values (multi-valued) to a single-valued map.
func flattenQuery(values url.Values) map[string]string {
	if len(values) == 0 {
		return nil
	}
	flat := make(map[string]string, len(values))
	for k, v := range values {
		if len(v) > 0 {
			flat[k] = v[0]
		}
	}
	return flat
}

// NetHTTPFunc wraps a handler function with panic recovery.
func NetHTTPFunc(next http.HandlerFunc) http.HandlerFunc {
	return NetHTTP(next).(http.HandlerFunc)
}
