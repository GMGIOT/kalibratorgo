package main

const (
	Connection_unknown = 0
	Connection_available = 1
	Connection_used = 2
	Connection_blocked = 3
	
	str_Serial = "Serial"
)

type Connection interface {
	ID() int
	Type() string
	Device() string
	Options() map[string]interface{}
	Protocol() string
	Status() int
	UsedBy() int
	Close(recursive bool) error
}
