package handler

import (
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"github.com/edwdch/web-tty/internal/session"
)

const (
	handshakeTimeout = 10 * time.Second
	writeWait        = 10 * time.Second
	readBufSize      = 16 * 1024
	tcpKeepAlive     = 30 * time.Second
)

// pingPeriod stays below common proxy/NAT idle timeouts (~30–60s). Client
// CmdPing also refreshes the read deadline if intermediaries drop control frames.
var (
	pongWait   = 60 * time.Second
	pingPeriod = 15 * time.Second
)

type TerminalHub interface {
	Full() bool
	Has(id string) bool
	Create(columns, rows uint16) (session.Handle, error)
	Get(id string) (session.Handle, error)
}

func CheckOrigin(allow []string) func(*http.Request) bool {
	allowed := make(map[string]struct{}, len(allow))
	for _, o := range allow {
		o = strings.TrimSpace(o)
		if o != "" {
			allowed[o] = struct{}{}
		}
	}
	return func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		if origin == "" {
			return true
		}
		u, err := url.Parse(origin)
		if err != nil || u.Host == "" {
			return false
		}
		if strings.EqualFold(u.Host, r.Host) {
			return true
		}
		_, ok := allowed[origin]
		return ok
	}
}

func Terminal(hub TerminalHub, writable bool, checkOrigin func(*http.Request) bool) gin.HandlerFunc {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  readBufSize,
		WriteBufferSize: readBufSize,
		CheckOrigin:     checkOrigin,
	}
	return func(c *gin.Context) {
		attachID := strings.TrimSpace(c.Query("id"))
		if attachID != "" {
			if !hub.Has(attachID) {
				c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
				return
			}
		} else if hub.Full() {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "too many sessions"})
			return
		}

		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		enableTCPKeepAlive(conn)

		serveTerminal(conn, hub, writable, attachID)
	}
}

func serveTerminal(conn *websocket.Conn, hub TerminalHub, writable bool, attachID string) {
	conn.SetReadLimit(1 << 20)
	_ = conn.SetReadDeadline(time.Now().Add(handshakeTimeout))

	_, data, err := conn.ReadMessage()
	if err != nil {
		return
	}
	size, err := session.ParseHello(data)
	if err != nil {
		return
	}

	var handle session.Handle
	if attachID == "" {
		handle, err = hub.Create(size.Columns, size.Rows)
	} else {
		handle, err = hub.Get(attachID)
		if err == nil {
			_ = handle.Resize(size.Columns, size.Rows)
		}
	}
	if err != nil {
		return
	}

	var writeMu sync.Mutex
	writeFrame := func(payload []byte) error {
		writeMu.Lock()
		defer writeMu.Unlock()
		_ = conn.SetWriteDeadline(time.Now().Add(writeWait))
		return conn.WriteMessage(websocket.BinaryMessage, payload)
	}

	if err := writeFrame(session.InfoFrame(handle.ID())); err != nil {
		return
	}
	if title := handle.Title(); title != "" {
		if err := writeFrame(session.TitleFrame(title)); err != nil {
			return
		}
	}

	clientID, out, replay, err := handle.AddClient()
	if err != nil {
		return
	}
	defer handle.RemoveClient(clientID)
	if len(replay) > 0 {
		if err := writeFrame(session.OutputFrame(replay)); err != nil {
			return
		}
	}

	_ = conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		_ = conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	done := make(chan struct{})
	go func() {
		defer close(done)
		for frame := range out {
			if werr := writeFrame(frame); werr != nil {
				_ = conn.Close()
				return
			}
		}
		_ = conn.Close()
	}()

	go func() {
		ticker := time.NewTicker(pingPeriod)
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				writeMu.Lock()
				_ = conn.SetWriteDeadline(time.Now().Add(writeWait))
				err := conn.WriteMessage(websocket.PingMessage, nil)
				writeMu.Unlock()
				if err != nil {
					_ = conn.Close()
					return
				}
			}
		}
	}()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}
		_ = conn.SetReadDeadline(time.Now().Add(pongWait))
		if len(msg) == 0 {
			continue
		}
		switch msg[0] {
		case session.CmdInput:
			if writable && len(msg) > 1 {
				_, _ = handle.Write(msg[1:])
			}
		case session.CmdResize:
			if sz, err := session.ParseResize(msg); err == nil {
				_ = handle.Resize(sz.Columns, sz.Rows)
			}
		case session.CmdPause:
			handle.Pause()
		case session.CmdResume:
			handle.Resume()
		case session.CmdPing:
			// keep-alive; read deadline already refreshed
		}
	}
	_ = conn.Close()
	handle.RemoveClient(clientID)
	<-done
}

func enableTCPKeepAlive(conn *websocket.Conn) {
	nc := conn.NetConn()
	if tc, ok := nc.(*net.TCPConn); ok {
		_ = tc.SetKeepAlive(true)
		_ = tc.SetKeepAlivePeriod(tcpKeepAlive)
	}
}
