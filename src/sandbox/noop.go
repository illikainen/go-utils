package sandbox

type Noop struct {
}

func NewNoop() (*Noop, error) {
	return &Noop{}, nil
}

func (n *Noop) AddReadOnlyPath(...string) error {
	return nil
}

func (n *Noop) AddReadWritePath(...string) error {
	return nil
}

func (n *Noop) AddDevPath(...string) error {
	return nil
}

func (n *Noop) Confine() error {
	return nil
}
