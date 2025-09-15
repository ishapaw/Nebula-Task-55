package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gin-gonic/gin"
)

func ReverseProxy(target string) gin.HandlerFunc {
    return func(c *gin.Context) {
        remote, err := url.Parse(target)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid target"})
            return
        }

        proxy := httputil.NewSingleHostReverseProxy(remote)

        // Preserve original path for upstream
        c.Request.URL.Scheme = remote.Scheme
        c.Request.URL.Host = remote.Host

        c.Request.Host = remote.Host
        proxy.ServeHTTP(c.Writer, c.Request)
    }
}
