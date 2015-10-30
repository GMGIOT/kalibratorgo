package main

type AbstarctDevice interface {
	Protocol() string
	Connection() Connection
	BindTo(c Connection)
	Close(recurcive bool) error
}

