package main

type AbstarctDevice interface {
	Protocol() string
	ConnectionID() int
	SetConnection(c Connection)
	Close(recurcive bool) error
}

