package main

type Oneshot struct {
	Name  string
	Needs []string
	Func  func() error
}

func (s Oneshot) GetName() string {
	return s.Name
}

func (s Oneshot) CanStart() bool {
	return services.HasFlag(s.Needs...)
}

func (s Oneshot) Run(_ func()) error {
	return s.Func()
}

func (s Oneshot) Stop() {
	// no-op
}
