package main

import (
	"io"
)

type AbstractCell interface {
	io.Reader
	io.Writer
	Type() string
	Address() uint16
	Name() string
	Category() string
	Description() string
	Variants() interface{}
	variantsAdvanced() map[string]interface{}
}

