package middleware

import (
	"fmt"

	bugstack "github.com/MasonBachmann7/bugstack-go"
)

// EchoMiddleware returns an Echo middleware function that captures panics.
//
// Usage with Echo:
//
//	e := echo.New()
//	e.Use(echo.MiddlewareFunc(middleware.EchoRecover()))
//
// Since we don't import echo, the returned function works with
// Echo's middleware signature: func(next echo.HandlerFunc) echo.HandlerFunc
// Users should adapt at the call site.
func EchoRecover() func(next func(c any) error) func(c any) error {
	return func(next func(c any) error) func(c any) error {
		return func(c any) error {
			type echoCtx interface {
				Path() string
				Request() any
			}

			defer func() {
				if rec := recover(); rec != nil {
					client := bugstack.GetClient()
					if client != nil {
						msg := fmt.Sprintf("panic: %v", rec)
						opts := []bugstack.CaptureOption{
							bugstack.WithMetadata(map[string]any{
								"framework": "echo",
							}),
						}

						if ec, ok := c.(echoCtx); ok {
							opts = append(opts, bugstack.WithRequest(&bugstack.RequestContext{
								Route: ec.Path(),
							}))
						}

						client.CaptureMessage(msg, opts...)
					}
					panic(rec)
				}
			}()

			return next(c)
		}
	}
}
