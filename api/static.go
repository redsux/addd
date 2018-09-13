package api

import (
	"net/http"
	"os"
	"path"

	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
)

func registerStatic(uipath string, uigroup *gin.RouterGroup) {
	if uipath != "" {
		if fs, err := os.Stat(uipath); err == nil && fs.IsDir() {
			spaServe := func(urlPrefix string, fs static.ServeFileSystem) gin.HandlerFunc {
				fileserver := http.FileServer(fs)
				if urlPrefix != "" {
					fileserver = http.StripPrefix(urlPrefix, fileserver)
				}
				return func(c *gin.Context) {
					if fs.Exists(urlPrefix, c.Request.URL.Path) {
						fileserver.ServeHTTP(c.Writer, c.Request)
						c.Abort()
					} else if indexPath := path.Join(uipath, "/index.html"); fs.Exists(urlPrefix, indexPath) {
						http.ServeFile(c.Writer, c.Request, indexPath)
						c.Abort()
					}
				}
			}
			spaHandler := spaServe(uigroup.BasePath(), static.LocalFile(uipath, false))
			uigroup.GET("/*filepath", spaHandler)
			uigroup.HEAD("/*filepath", spaHandler)
		}
	}

}
