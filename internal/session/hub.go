package session

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"sort"
	"sync"
	"time"
)

var (
	ErrFull     = errors.New("too many sessions")
	ErrNotFound = errors.New("session not found")
)

const defaultMaxSessions = 50

type Factory struct {
	opts Options
}

func NewFactory(opts Options) *Factory {
	return &Factory{opts: opts}
}

func (f *Factory) Start(columns, rows uint16) (*Session, error) {
	return Start(f.opts, columns, rows)
}

type Info struct {
	ID        string    `json:"id"`
	PID       int       `json:"pid"`
	Cwd       string    `json:"cwd,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
	Clients   int       `json:"clients"`
}

type Handle interface {
	ID() string
	Write(p []byte) (int, error)
	Resize(columns, rows uint16) error
	Pause()
	Resume()
	AddClient() (clientID uint64, out <-chan []byte, replay []byte, err error)
	RemoveClient(clientID uint64)
}

type Hub struct {
	mu        sync.Mutex
	sessions  map[string]*Managed
	max       int
	factory   *Factory
	idle      time.Duration
	reapEvery time.Duration
	now       func() time.Time
	stop      chan struct{}
	stopOnce  sync.Once
	wg        sync.WaitGroup
}

func NewHub(max int, factory *Factory, idle time.Duration) *Hub {
	if max <= 0 {
		max = defaultMaxSessions
	}
	h := &Hub{
		sessions: make(map[string]*Managed),
		max:      max,
		factory:  factory,
		idle:     idle,
		now:      time.Now,
		stop:     make(chan struct{}),
	}
	if idle > 0 {
		reapEvery := time.Minute
		if idle < reapEvery {
			reapEvery = idle
		}
		h.reapEvery = reapEvery
		h.wg.Add(1)
		go func() {
			defer h.wg.Done()
			h.reapLoop()
		}()
	}
	return h
}

func (h *Hub) Full() bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.sessions) >= h.max
}

func (h *Hub) Has(id string) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	_, ok := h.sessions[id]
	return ok
}

func (h *Hub) List() []Info {
	h.mu.Lock()
	defer h.mu.Unlock()
	out := make([]Info, 0, len(h.sessions))
	for _, m := range h.sessions {
		out = append(out, m.info())
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].CreatedAt.Equal(out[j].CreatedAt) {
			return out[i].ID < out[j].ID
		}
		return out[i].CreatedAt.Before(out[j].CreatedAt)
	})
	return out
}

func (h *Hub) Create(columns, rows uint16) (Handle, error) {
	h.mu.Lock()
	if len(h.sessions) >= h.max {
		h.mu.Unlock()
		return nil, ErrFull
	}
	h.mu.Unlock()

	pty, err := h.factory.Start(columns, rows)
	if err != nil {
		return nil, err
	}

	h.mu.Lock()
	if len(h.sessions) >= h.max {
		h.mu.Unlock()
		_ = pty.Close()
		return nil, ErrFull
	}
	id := newID()
	for h.sessions[id] != nil {
		id = newID()
	}
	now := h.now()
	m := &Managed{
		id:           id,
		conn:         pty,
		hub:          h,
		clients:      make(map[uint64]chan []byte),
		buf:          newRing(defaultReplayBytes),
		createdAt:    now,
		lastDetachAt: now,
	}
	h.sessions[id] = m
	h.mu.Unlock()
	go m.run()
	return m, nil
}

func (h *Hub) Get(id string) (Handle, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	m, ok := h.sessions[id]
	if !ok {
		return nil, ErrNotFound
	}
	return m, nil
}

func (h *Hub) Close(id string) error {
	h.mu.Lock()
	m, ok := h.sessions[id]
	if !ok {
		h.mu.Unlock()
		return ErrNotFound
	}
	delete(h.sessions, id)
	h.mu.Unlock()
	m.shutdown()
	return nil
}

func (h *Hub) Stop() {
	h.stopOnce.Do(func() {
		close(h.stop)
	})
	h.wg.Wait()
	h.mu.Lock()
	ids := make([]string, 0, len(h.sessions))
	for id := range h.sessions {
		ids = append(ids, id)
	}
	h.mu.Unlock()
	for _, id := range ids {
		_ = h.Close(id)
	}
}

func (h *Hub) reapLoop() {
	ticker := time.NewTicker(h.reapEvery)
	defer ticker.Stop()
	for {
		select {
		case <-h.stop:
			return
		case <-ticker.C:
			h.reapIdle()
		}
	}
}

func (h *Hub) reapIdle() {
	now := h.now()
	var stale []string
	h.mu.Lock()
	for id, m := range h.sessions {
		if m.idleSince(now, h.idle) {
			stale = append(stale, id)
		}
	}
	h.mu.Unlock()
	for _, id := range stale {
		_ = h.Close(id)
	}
}

func newID() string {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic("session id: " + err.Error())
	}
	return hex.EncodeToString(b[:])
}

type Managed struct {
	id   string
	conn Conn
	hub  *Hub

	mu           sync.Mutex
	clients      map[uint64]chan []byte
	nextID       uint64
	buf          *byteRing
	createdAt    time.Time
	lastDetachAt time.Time
	closed       bool
}

func (m *Managed) ID() string { return m.id }

func (m *Managed) Write(p []byte) (int, error) {
	return m.conn.Write(p)
}

func (m *Managed) Resize(columns, rows uint16) error {
	return m.conn.Resize(columns, rows)
}

func (m *Managed) Pause()  { m.conn.Pause() }
func (m *Managed) Resume() { m.conn.Resume() }

func (m *Managed) AddClient() (uint64, <-chan []byte, []byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return 0, nil, nil, ErrNotFound
	}
	m.nextID++
	id := m.nextID
	ch := make(chan []byte, 256)
	m.clients[id] = ch
	return id, ch, m.buf.Bytes(), nil
}

func (m *Managed) RemoveClient(id uint64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	ch, ok := m.clients[id]
	if !ok {
		return
	}
	delete(m.clients, id)
	close(ch)
	if len(m.clients) == 0 && !m.closed {
		m.lastDetachAt = m.hub.now()
	}
}

func (m *Managed) info() Info {
	m.mu.Lock()
	n := len(m.clients)
	created := m.createdAt
	m.mu.Unlock()
	pid := m.conn.PID()
	return Info{
		ID:        m.id,
		PID:       pid,
		Cwd:       cwdOf(pid),
		CreatedAt: created,
		Clients:   n,
	}
}

func (m *Managed) idleSince(now time.Time, idle time.Duration) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed || len(m.clients) > 0 {
		return false
	}
	return !m.lastDetachAt.IsZero() && now.Sub(m.lastDetachAt) >= idle
}

func (m *Managed) run() {
	buf := make([]byte, 16*1024)
	for {
		n, err := m.conn.Read(buf)
		if n > 0 {
			chunk := append([]byte(nil), buf[:n]...)
			frame := OutputFrame(chunk)
			m.mu.Lock()
			if !m.closed {
				m.buf.Append(chunk)
				for _, ch := range m.clients {
					select {
					case ch <- frame:
					default:
					}
				}
			}
			m.mu.Unlock()
		}
		if err != nil {
			_ = m.hub.Close(m.id)
			return
		}
	}
}

func (m *Managed) shutdown() {
	m.mu.Lock()
	if m.closed {
		m.mu.Unlock()
		return
	}
	m.closed = true
	for id, ch := range m.clients {
		close(ch)
		delete(m.clients, id)
	}
	m.mu.Unlock()
	_ = m.conn.Close()
}

func cwdOf(pid int) string {
	if pid <= 0 {
		return ""
	}
	p, err := os.Readlink(fmt.Sprintf("/proc/%d/cwd", pid))
	if err != nil {
		return ""
	}
	return p
}
