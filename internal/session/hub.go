package session

type Factory struct {
	opts Options
}

func NewFactory(opts Options) *Factory {
	return &Factory{opts: opts}
}

func (f *Factory) Start(columns, rows uint16) (*Session, error) {
	return Start(f.opts, columns, rows)
}

type Hub struct {
	slots   chan struct{}
	factory *Factory
}

func NewHub(max int, factory *Factory) *Hub {
	if max <= 0 {
		max = 8
	}
	return &Hub{
		slots:   make(chan struct{}, max),
		factory: factory,
	}
}

func (h *Hub) TryAcquire() bool {
	select {
	case h.slots <- struct{}{}:
		return true
	default:
		return false
	}
}

func (h *Hub) Release() {
	<-h.slots
}

func (h *Hub) Start(columns, rows uint16) (Conn, error) {
	return h.factory.Start(columns, rows)
}
