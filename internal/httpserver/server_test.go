package httpserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/gin-gonic/gin"

	"github.com/edwdch/web-tty/internal/config"
)

func TestSPAFallback(t *testing.T) {
	gin.SetMode(gin.TestMode)

	static := fstest.MapFS{
		"index.html":    {Data: []byte("<!doctype html><title>app</title>")},
		"assets/app.js": {Data: []byte("console.log(1)")},
	}
	r := newEngine(config.Config{GinMode: gin.TestMode}, static)

	t.Run("root serves index", func(t *testing.T) {
		assertBody(t, r, "/", http.StatusOK, "<!doctype html><title>app</title>")
	})

	t.Run("asset is served", func(t *testing.T) {
		assertBody(t, r, "/assets/app.js", http.StatusOK, "console.log(1)")
	})

	t.Run("unknown path falls back to index", func(t *testing.T) {
		assertBody(t, r, "/about", http.StatusOK, "<!doctype html><title>app</title>")
	})

	t.Run("directory falls back to index", func(t *testing.T) {
		assertBody(t, r, "/assets", http.StatusOK, "<!doctype html><title>app</title>")
	})

	t.Run("api prefix is json 404", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/missing", nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusNotFound)
		}
		var body struct {
			Error string `json:"error"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if body.Error != "not found" {
			t.Fatalf("error = %q, want %q", body.Error, "not found")
		}
	})
}

func TestSPAFallbackMissingIndex(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.NoRoute(spaFallback(fstest.MapFS{}))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func assertBody(t *testing.T, h http.Handler, path string, wantStatus int, wantBody string) {
	t.Helper()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	h.ServeHTTP(w, req)
	if w.Code != wantStatus {
		t.Fatalf("status = %d, want %d", w.Code, wantStatus)
	}
	got := w.Body.String()
	if got != wantBody {
		t.Fatalf("body = %q, want %q", got, wantBody)
	}
}
