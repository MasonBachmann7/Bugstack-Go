package middleware

import (
	"fmt"

	bugstack "github.com/MasonBachmann7/bugstack-go"
)

// GinContext is a minimal interface matching gin.Context to avoid
// importing gin as a dependency. Users pass their *gin.Context directly.
type GinContext interface {
	FullPath() string
	Request() any
	Errors() any
}

// Gin returns a Gin middleware that captures panics.
//
// Usage:
//
//	r := gin.Default()
//	r.Use(middleware.GinMiddleware())
//
// This function returns a func(c *gin.Context) compatible with Gin.
// Due to Go's type system and to avoid importing gin as a dependency,
// the function signature uses any. Cast to gin.HandlerFunc at the call site:
//
//	r.Use(gin.HandlerFunc(middleware.GinRecovery()))
//
// Or use the provided helper that accepts the gin context interface.
func GinRecovery() func(c any) {
	return func(c any) {
		type ginCtx interface {
			Next()
			FullPath() string
			Request() any
			AbortWithStatus(int)
		}

		gc, ok := c.(ginCtx)
		if !ok {
			return
		}

		defer func() {
			if rec := recover(); rec != nil {
				client := bugstack.GetClient()
				if client != nil {
					msg := fmt.Sprintf("panic: %v", rec)
					client.CaptureMessage(msg,
						bugstack.WithRequest(&bugstack.RequestContext{
							Route: gc.FullPath(),
						}),
						bugstack.WithMetadata(map[string]any{
							"framework": "gin",
						}),
					)
				}
				gc.AbortWithStatus(500)
				panic(rec)
			}
		}()

		gc.Next()
	}
}
