package main

import (
	"io"
)

type AbstractCell interface {
	io.Reader
	io.Writer
	ID() string
	Type() string
	Address() uint16
	Name() string
	Category() string
	Description() string
	Variants() map[string]interface{}
}

