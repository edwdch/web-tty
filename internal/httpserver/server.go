package httpserver

import (
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"

	"simple-app/internal/config"
	"simple-app/internal/handler"
)

func New(cfg config.Config) *gin.Engine {
	if cfg.GinMode != "" {
		gin.SetMode(cfg.GinMode)
	}

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	r.GET("/api/ping", handler.Ping)
	r.NoRoute(spaFallback(cfg.DistDir))
	return r
}

func Run(cfg config.Config) error {
	return New(cfg).Run(cfg.Addr)
}

func spaFallback(distDir string) gin.HandlerFunc {
	return func(c *gin.Context) {
		urlPath := path.Clean("/" + c.Request.URL.Path)
		if strings.HasPrefix(urlPath, "/api/") {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}

		if urlPath != "/" {
			candidate := filepath.Join(distDir, filepath.FromSlash(urlPath))
			if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
				c.File(candidate)
				return
			}
		}

		c.File(filepath.Join(distDir, "index.html"))
	}
}
