package handler

import (
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
	pongWait         = 60 * time.Second
	pingPeriod       = (pongWait * 9) / 10
	readBufSize      = 16 * 1024
)

type TerminalHub interface {
	TryAcquire() bool
	Release()
	Start(columns, rows uint16) (session.Conn, error)
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
		if !hub.TryAcquire() {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "too many sessions"})
			return
		}
		defer hub.Release()

		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		serveTerminal(conn, hub, writable)
	}
}

func serveTerminal(conn *websocket.Conn, hub TerminalHub, writable bool) {
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

	sess, err := hub.Start(size.Columns, size.Rows)
	if err != nil {
		return
	}
	defer sess.Close()

	_ = conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		_ = conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	var writeMu sync.Mutex
	writeFrame := func(payload []byte) error {
		writeMu.Lock()
		defer writeMu.Unlock()
		_ = conn.SetWriteDeadline(time.Now().Add(writeWait))
		return conn.WriteMessage(websocket.BinaryMessage, payload)
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		buf := make([]byte, readBufSize)
		for {
			n, rerr := sess.Read(buf)
			if n > 0 {
				if werr := writeFrame(session.OutputFrame(buf[:n])); werr != nil {
					return
				}
			}
			if rerr != nil {
				return
			}
		}
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
				_, _ = sess.Write(msg[1:])
			}
		case session.CmdResize:
			if sz, err := session.ParseResize(msg); err == nil {
				_ = sess.Resize(sz.Columns, sz.Rows)
			}
		case session.CmdPause:
			sess.Pause()
		case session.CmdResume:
			sess.Resume()
		}
	}
	_ = conn.Close()
	_ = sess.Close()
	<-done
}
