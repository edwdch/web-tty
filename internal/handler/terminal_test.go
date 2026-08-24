package handler_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"github.com/edwdch/web-tty/internal/handler"
	"github.com/edwdch/web-tty/internal/session"
)

func TestCheckOrigin(t *testing.T) {
	f := handler.CheckOrigin([]string{"http://127.0.0.1:5173"})

	req := httptest.NewRequest(http.MethodGet, "http://example.com/ws", nil)
	if !f(req) {
		t.Fatal("empty origin should pass")
	}

	req.Host = "example.com"
	req.Header.Set("Origin", "http://example.com")
	if !f(req) {
		t.Fatal("same origin should pass")
	}

	req.Header.Set("Origin", "http://127.0.0.1:5173")
	if !f(req) {
		t.Fatal("allow list should pass")
	}

	req.Header.Set("Origin", "https://evil.example")
	if f(req) {
		t.Fatal("foreign origin should fail")
	}
}

type fullHub struct{}

func (fullHub) TryAcquire() bool { return false }
func (fullHub) Release()         {}
func (fullHub) Start(uint16, uint16) (session.Conn, error) {
	panic("Start should not run")
}

func TestTerminalTooManySessions(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/ws", handler.Terminal(fullHub{}, true, handler.CheckOrigin(nil)))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ws", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d", w.Code)
	}
	var body struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.Error != "too many sessions" {
		t.Fatalf("error = %q", body.Error)
	}
}

func TestTerminalEchoAndIndependentSessions(t *testing.T) {
	gin.SetMode(gin.TestMode)
	hub := session.NewHub(4, session.NewFactory(session.Options{Shell: "/bin/cat"}))
	r := gin.New()
	r.GET("/ws", handler.Terminal(hub, true, handler.CheckOrigin(nil)))
	srv := httptest.NewServer(r)
	t.Cleanup(srv.Close)

	a := dialWS(t, srv)
	defer a.Close()
	b := dialWS(t, srv)
	defer b.Close()

	hello := []byte(`{"columns":80,"rows":24,"cmd":"/bin/false"}`)
	writeBin(t, a, hello)
	writeBin(t, b, hello)

	writeBin(t, a, append([]byte{'0'}, []byte("alpha\n")...))
	writeBin(t, b, append([]byte{'0'}, []byte("bravo\n")...))

	if got := readOutput(t, a); !bytes.Contains(got, []byte("alpha")) {
		t.Fatalf("session a = %q", got)
	}
	if got := readOutput(t, b); !bytes.Contains(got, []byte("bravo")) {
		t.Fatalf("session b = %q", got)
	}
}

func TestTerminalInvalidHelloCloses(t *testing.T) {
	gin.SetMode(gin.TestMode)
	hub := session.NewHub(2, session.NewFactory(session.Options{Shell: "/bin/cat"}))
	r := gin.New()
	r.GET("/ws", handler.Terminal(hub, true, handler.CheckOrigin(nil)))
	srv := httptest.NewServer(r)
	t.Cleanup(srv.Close)

	conn := dialWS(t, srv)
	defer conn.Close()
	writeBin(t, conn, []byte(`{{"columns":80,"rows":24}`))

	_ = conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	_, _, err := conn.ReadMessage()
	if err == nil {
		t.Fatal("expected close after invalid hello")
	}
}

func TestTerminalKeepAlive(t *testing.T) {
	t.Run("websocket pong keeps idle session", func(t *testing.T) {
		t.Cleanup(handler.OverrideKeepAlive(300*time.Millisecond, 80*time.Millisecond))
		conn := startCatSession(t)
		client := pumpWS(conn)

		time.Sleep(time.Second)
		writeBin(t, conn, append([]byte{'0'}, []byte("alive\n")...))
		waitOutputContains(t, client, []byte("alive"))
	})

	t.Run("application ping keeps session without pong", func(t *testing.T) {
		t.Cleanup(handler.OverrideKeepAlive(300*time.Millisecond, 80*time.Millisecond))
		conn := startCatSession(t)
		conn.SetPingHandler(func(string) error { return nil })
		client := pumpWS(conn)

		deadline := time.Now().Add(time.Second)
		for time.Now().Before(deadline) {
			writeBin(t, conn, []byte{session.CmdPing})
			time.Sleep(80 * time.Millisecond)
		}
		writeBin(t, conn, append([]byte{'0'}, []byte("alive\n")...))
		got := waitOutputContains(t, client, []byte("alive"))
		if bytes.Contains(got, []byte{session.CmdPing}) {
			t.Fatalf("ping was written to pty: %q", got)
		}
	})

	t.Run("idle session dies without pong or client ping", func(t *testing.T) {
		t.Cleanup(handler.OverrideKeepAlive(250*time.Millisecond, 80*time.Millisecond))
		conn := startCatSession(t)
		conn.SetPingHandler(func(string) error { return nil })
		client := pumpWS(conn)

		select {
		case err := <-client.err:
			if err == nil {
				t.Fatal("expected close error")
			}
		case <-time.After(2 * time.Second):
			t.Fatal("expected server to close idle connection")
		}
	})
}

func TestTerminalReadonlyIgnoresInput(t *testing.T) {
	gin.SetMode(gin.TestMode)
	hub := session.NewHub(2, session.NewFactory(session.Options{Shell: "/bin/cat"}))
	r := gin.New()
	r.GET("/ws", handler.Terminal(hub, false, handler.CheckOrigin(nil)))
	srv := httptest.NewServer(r)
	t.Cleanup(srv.Close)

	conn := dialWS(t, srv)
	defer conn.Close()
	writeBin(t, conn, []byte(`{"columns":80,"rows":24}`))
	writeBin(t, conn, append([]byte{'0'}, []byte("secret\n")...))

	_ = conn.SetReadDeadline(time.Now().Add(400 * time.Millisecond))
	_, msg, err := conn.ReadMessage()
	if err == nil && bytes.Contains(msg, []byte("secret")) {
		t.Fatalf("readonly wrote through: %q", msg)
	}
}

type wsPump struct {
	out chan []byte
	err chan error
}

func startCatSession(t *testing.T) *websocket.Conn {
	t.Helper()
	gin.SetMode(gin.TestMode)
	hub := session.NewHub(4, session.NewFactory(session.Options{Shell: "/bin/cat"}))
	r := gin.New()
	r.GET("/ws", handler.Terminal(hub, true, handler.CheckOrigin(nil)))
	srv := httptest.NewServer(r)
	t.Cleanup(srv.Close)

	conn := dialWS(t, srv)
	t.Cleanup(func() { _ = conn.Close() })
	writeBin(t, conn, []byte(`{"columns":80,"rows":24}`))
	return conn
}

func pumpWS(conn *websocket.Conn) *wsPump {
	p := &wsPump{
		out: make(chan []byte, 32),
		err: make(chan error, 1),
	}
	go func() {
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				p.err <- err
				return
			}
			if len(msg) > 0 && msg[0] == session.MsgOutput {
				chunk := append([]byte(nil), msg[1:]...)
				p.out <- chunk
			}
		}
	}()
	return p
}

func waitOutputContains(t *testing.T, p *wsPump, needle []byte) []byte {
	t.Helper()
	deadline := time.After(5 * time.Second)
	var buf []byte
	for {
		select {
		case err := <-p.err:
			t.Fatalf("ws closed: %v (got %q)", err, buf)
		case <-deadline:
			t.Fatalf("timeout, got %q want %q", buf, needle)
		case chunk := <-p.out:
			buf = append(buf, chunk...)
			if bytes.Contains(buf, needle) {
				return buf
			}
		}
	}
}

func dialWS(t *testing.T, srv *httptest.Server) *websocket.Conn {
	t.Helper()
	u := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws"
	conn, resp, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		if resp != nil {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("dial: %v (%d %s)", err, resp.StatusCode, body)
		}
		t.Fatalf("dial: %v", err)
	}
	return conn
}

func writeBin(t *testing.T, conn *websocket.Conn, msg []byte) {
	t.Helper()
	if err := conn.WriteMessage(websocket.BinaryMessage, msg); err != nil {
		t.Fatalf("write: %v", err)
	}
}

func readOutput(t *testing.T, conn *websocket.Conn) []byte {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	var out []byte
	for time.Now().Before(deadline) {
		_ = conn.SetReadDeadline(time.Now().Add(time.Second))
		_, msg, err := conn.ReadMessage()
		if err != nil {
			if len(out) > 0 {
				return out
			}
			t.Fatalf("read: %v", err)
		}
		if len(msg) > 0 && msg[0] == '0' {
			out = append(out, msg[1:]...)
			if bytes.Contains(out, []byte("\n")) {
				return out
			}
		}
	}
	t.Fatalf("timeout, got %q", out)
	return out
}
