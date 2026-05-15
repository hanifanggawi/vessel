package daemon

type EventKind int

const (
	EventTick EventKind = iota
	EventReload
	EventShutdown
)

type Event struct {
	Kind EventKind
}
