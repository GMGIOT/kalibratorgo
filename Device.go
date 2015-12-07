package main

type AbstarctDevice interface {
	ID() string
	Protocol() string
	Connection() Connection
	BindTo(c Connection) error
	Close(recurcive bool) error
}

