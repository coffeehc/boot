package manage

import (
	"net/http/pprof"

	"github.com/gin-gonic/gin"
)

func registerPprof(route gin.IRouter) {
	route = route.Group("/debug/pprof")
	route.GET("/", indexHandler())
	route.GET("/heap", heapHandler())
	route.GET("/goroutine", goroutineHandler())
	route.GET("/block", blockHandler())
	route.GET("/threadcreate", threadCreateHandler())
	route.GET("/cmdline", cmdlineHandler())
	route.GET("/profile", profileHandler())
	route.GET("/symbol", symbolHandler())
	route.POST("/symbol", symbolHandler())
	route.GET("/trace", traceHandler())
	route.GET("/mutex", mutexHandler())
	route.GET("/allocs", allocsHandler())

}

// IndexHandler will pass the call from /debug/pprof to pprof
func indexHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		pprof.Index(ctx.Writer, ctx.Request)
	}
}

// HeapHandler will pass the call from /debug/pprof/heap to pprof
func heapHandler() gin.HandlerFunc {
	return gin.WrapH(pprof.Handler("heap"))
}

// GoroutineHandler will pass the call from /debug/pprof/goroutine to pprof
func allocsHandler() gin.HandlerFunc {
	return gin.WrapH(pprof.Handler("allocs"))
}

// GoroutineHandler will pass the call from /debug/pprof/goroutine to pprof
func goroutineHandler() gin.HandlerFunc {
	return gin.WrapH(pprof.Handler("goroutine"))
}

// BlockHandler will pass the call from /debug/pprof/block to pprof
func blockHandler() gin.HandlerFunc {
	return gin.WrapH(pprof.Handler("block"))
}

// ThreadCreateHandler will pass the call from /debug/pprof/threadcreate to pprof
func threadCreateHandler() gin.HandlerFunc {
	return gin.WrapH(pprof.Handler("threadcreate"))
}

// CmdlineHandler will pass the call from /debug/pprof/cmdline to pprof
func cmdlineHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		pprof.Cmdline(ctx.Writer, ctx.Request)
	}
}

// ProfileHandler will pass the call from /debug/pprof/profile to pprof
func profileHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		pprof.Profile(ctx.Writer, ctx.Request)
	}
}

// SymbolHandler will pass the call from /debug/pprof/symbol to pprof
func symbolHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		pprof.Symbol(ctx.Writer, ctx.Request)
	}
}

// TraceHandler will pass the call from /debug/pprof/trace to pprof
func traceHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		pprof.Trace(ctx.Writer, ctx.Request)
	}
}

// MutexHandler will pass the call from /debug/pprof/mutex to pprof
func mutexHandler() gin.HandlerFunc {
	return gin.WrapH(pprof.Handler("mutex"))
}
