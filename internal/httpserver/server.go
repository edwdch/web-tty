package httpserver

import (
	"bytes"
	"io"
	"io/fs"
	"net/http"
	"path"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/edwdch/web-tty/internal/config"
	"github.com/edwdch/web-tty/internal/handler"
	"github.com/edwdch/web-tty/web"
)

func New(cfg config.Config) *gin.Engine {
	return newEngine(cfg, web.Files())
}

func Run(cfg config.Config) error {
	return New(cfg).Run(cfg.Addr)
}

func newEngine(cfg config.Config, static fs.FS) *gin.Engine {
	if cfg.GinMode != "" {
		gin.SetMode(cfg.GinMode)
	}

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	r.GET("/api/ping", handler.Ping)
	r.NoRoute(spaFallback(static))
	return r
}

func spaFallback(static fs.FS) gin.HandlerFunc {
	return func(c *gin.Context) {
		urlPath := path.Clean("/" + c.Request.URL.Path)
		if strings.HasPrefix(urlPath, "/api/") {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}

		rel := strings.TrimPrefix(urlPath, "/")
		if rel != "" && rel != "." {
			if info, err := fs.Stat(static, rel); err == nil && !info.IsDir() {
				serveFSFile(c, static, rel)
				return
			}
		}

		serveFSFile(c, static, "index.html")
	}
}

func serveFSFile(c *gin.Context, static fs.FS, name string) {
	f, err := static.Open(name)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}

	var reader io.ReadSeeker
	if rs, ok := f.(io.ReadSeeker); ok {
		reader = rs
	} else {
		data, err := io.ReadAll(f)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}
		reader = bytes.NewReader(data)
	}

	http.ServeContent(c.Writer, c.Request, info.Name(), info.ModTime(), reader)
}
