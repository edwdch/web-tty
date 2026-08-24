package handler_test

import (
	"encoding/json"
	"errors"
	"io"
	"net"
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

func TestListSessionsEmpty(t *testing.T) {
	srv, _ := testSessionServer(t, 4, session.Options{Shell: "/bin/cat"})
	list := getSessions(t, srv)
	if len(list) != 0 {
		t.Fatalf("list = %+v", list)
	}
}

func TestDeleteMissingSession(t *testing.T) {
	srv, _ := testSessionServer(t, 4, session.Options{Shell: "/bin/cat"})
	res, err := http.DefaultClient.Do(deleteReq(t, srv, "missing"))
	if err != nil {
		t.Fatalf("delete: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d", res.StatusCode)
	}
}

func TestSessionPersistsAfterWSCloseAndReplay(t *testing.T) {
	srv, hub := testSessionServer(t, 4, session.Options{Shell: "/bin/cat"})

	conn := dialWS(t, srv)
	writeBin(t, conn, []byte(`{"columns":80,"rows":24}`))
	id := readInfoID(t, conn)
	writeBin(t, conn, append([]byte{'0'}, []byte("keepme\n")...))
	if got := readOutput(t, conn); !strings.Contains(string(got), "keepme") {
		t.Fatalf("first output = %q", got)
	}
	_ = conn.Close()

	deadline := time.Now().Add(2 * time.Second)
	for {
		list := hub.List()
		if len(list) == 1 && list[0].Clients == 0 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("detach not reflected: %+v", list)
		}
		time.Sleep(20 * time.Millisecond)
	}

	conn2 := dialWSPath(t, srv, "/ws?id="+id)
	writeBin(t, conn2, []byte(`{"columns":80,"rows":24}`))
	if got := readInfoID(t, conn2); got != id {
		t.Fatalf("id = %s want %s", got, id)
	}
	got := readOutput(t, conn2)
	if !strings.Contains(string(got), "keepme") {
		t.Fatalf("replay = %q", got)
	}
	_ = conn2.Close()
}

func TestSessionSharedByTwoClients(t *testing.T) {
	srv, _ := testSessionServer(t, 4, session.Options{Shell: "/bin/cat"})

	a := dialWS(t, srv)
	writeBin(t, a, []byte(`{"columns":80,"rows":24}`))
	id := readInfoID(t, a)

	b := dialWSPath(t, srv, "/ws?id="+id)
	writeBin(t, b, []byte(`{"columns":80,"rows":24}`))
	if got := readInfoID(t, b); got != id {
		t.Fatalf("id = %s", got)
	}

	pumpA := pumpWS(a)
	pumpB := pumpWS(b)
	writeBin(t, a, append([]byte{'0'}, []byte("shared\n")...))
	waitOutputContains(t, pumpA, []byte("shared"))
	waitOutputContains(t, pumpB, []byte("shared"))
	_ = a.Close()
	_ = b.Close()
}

func TestSessionOutputWhileDetached(t *testing.T) {
	srv, _ := testSessionServer(t, 2, session.Options{
		Shell:     "/bin/sh",
		ShellArgs: []string{"-c", "echo one; sleep 0.4; echo two; sleep 30"},
	})

	conn := dialWS(t, srv)
	writeBin(t, conn, []byte(`{"columns":80,"rows":24}`))
	id := readInfoID(t, conn)
	pump := pumpWS(conn)
	waitOutputContains(t, pump, []byte("one"))
	_ = conn.Close()

	time.Sleep(600 * time.Millisecond)

	conn2 := dialWSPath(t, srv, "/ws?id="+id)
	writeBin(t, conn2, []byte(`{"columns":80,"rows":24}`))
	_ = readInfoID(t, conn2)
	got := readOutput(t, conn2)
	if !strings.Contains(string(got), "two") {
		t.Fatalf("replay while detached = %q", got)
	}
	_ = conn2.Close()
}

func TestDeleteSessionKillsPTY(t *testing.T) {
	srv, _ := testSessionServer(t, 2, session.Options{Shell: "/bin/cat"})

	conn := dialWS(t, srv)
	writeBin(t, conn, []byte(`{"columns":80,"rows":24}`))
	id := readInfoID(t, conn)

	res, err := http.DefaultClient.Do(deleteReq(t, srv, id))
	if err != nil {
		t.Fatalf("delete: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusNoContent {
		t.Fatalf("status = %d", res.StatusCode)
	}

	_ = conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	for {
		_, _, err := conn.ReadMessage()
		if err == nil {
			continue
		}
		var ne net.Error
		if errors.As(err, &ne) && ne.Timeout() {
			t.Fatal("expected ws close after delete")
		}
		break
	}

	if list := getSessions(t, srv); len(list) != 0 {
		t.Fatalf("list = %+v", list)
	}

	u := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws?id=" + id
	_, resp, err := websocket.DefaultDialer.Dial(u, nil)
	if err == nil {
		t.Fatal("expected attach 404")
	}
	if resp == nil || resp.StatusCode != http.StatusNotFound {
		t.Fatalf("attach after delete: err=%v resp=%v", err, resp)
	}
	if resp != nil {
		_, _ = io.ReadAll(resp.Body)
		_ = resp.Body.Close()
	}
}

func TestAttachUnknownSession(t *testing.T) {
	srv, _ := testSessionServer(t, 2, session.Options{Shell: "/bin/cat"})
	u := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws?id=missing"
	_, resp, err := websocket.DefaultDialer.Dial(u, nil)
	if err == nil {
		t.Fatal("expected 404")
	}
	if resp == nil || resp.StatusCode != http.StatusNotFound {
		t.Fatalf("err=%v resp=%v", err, resp)
	}
	if resp != nil {
		_, _ = io.ReadAll(resp.Body)
		_ = resp.Body.Close()
	}
}

func TestCreateRejectedWhenFullAfterPersist(t *testing.T) {
	srv, _ := testSessionServer(t, 1, session.Options{Shell: "/bin/cat"})
	conn := dialWS(t, srv)
	writeBin(t, conn, []byte(`{"columns":80,"rows":24}`))
	_ = readInfoID(t, conn)
	_ = conn.Close()

	u := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws"
	_, resp, err := websocket.DefaultDialer.Dial(u, nil)
	if err == nil {
		t.Fatal("expected 503")
	}
	if resp == nil || resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("err=%v resp=%v", err, resp)
	}
	if resp != nil {
		_, _ = io.ReadAll(resp.Body)
		_ = resp.Body.Close()
	}
}

func testSessionServer(t *testing.T, max int, opts session.Options) (*httptest.Server, *session.Hub) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	hub := session.NewHub(max, session.NewFactory(opts), 0)
	t.Cleanup(hub.Stop)
	r := gin.New()
	r.GET("/api/sessions", handler.ListSessions(hub))
	r.DELETE("/api/sessions/:id", handler.DeleteSession(hub))
	r.GET("/ws", handler.Terminal(hub, true, handler.CheckOrigin(nil)))
	srv := httptest.NewServer(r)
	t.Cleanup(srv.Close)
	return srv, hub
}

func getSessions(t *testing.T, srv *httptest.Server) []session.Info {
	t.Helper()
	res, err := http.Get(srv.URL + "/api/sessions")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", res.StatusCode)
	}
	var body struct {
		Sessions []session.Info `json:"sessions"`
	}
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.Sessions == nil {
		return []session.Info{}
	}
	return body.Sessions
}

func deleteReq(t *testing.T, srv *httptest.Server, id string) *http.Request {
	t.Helper()
	req, err := http.NewRequest(http.MethodDelete, srv.URL+"/api/sessions/"+id, nil)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	return req
}
