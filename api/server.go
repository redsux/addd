package api

import (
	"net/http"
	"time"
	
    "github.com/gin-gonic/gin"
	"github.com/gin-contrib/cors"
	
	"github.com/redsux/addd/core"
)

var (
	engine *gin.Engine
	secret string = "secret"
)

func init() {
	gin.SetMode(gin.ReleaseMode)

	engine = gin.New()
	engine.RedirectTrailingSlash = false
	engine.Use( logger() )
	engine.Use( gin.Recovery() )
	engine.Use( cors.Default() )
	engine.Use( authRequired() )
}

func Serve(listen, auth string, debug ...bool) {
	if len(debug) > 0 {
		if debug[0] {
			gin.SetMode(gin.DebugMode)
		}
	}
	registerRoutes(engine)
	
	if err := engine.Run( listen ); err != nil {
		addd.Log.Error("Failed to run the rest api server.")
		panic(err.Error())
	}
	
}

func authRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		if secret != "" {
			xauth := c.Request.Header.Get("X-AUTH-TOKEN")
			if xauth == "" || xauth != secret {
				c.AbortWithStatus(http.StatusUnauthorized)
			}
		}
		c.Next()
	}
}

func logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Stop timer
		end := time.Now()
		latency := end.Sub(start)

		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()

		if raw != "" {
			path = path + "?" + raw
		}

		addd.Log.NoticeF("[API] %3d | %13v | %15s | %-7s %s",
			statusCode,
			latency,
			clientIP,
			method,
			path,
		)
	}
}