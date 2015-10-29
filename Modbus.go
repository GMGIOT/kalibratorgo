package main

import (
	"time"
)

type Modbus interface {
	Timeout() time.Duration
	SetTimeout(time.Duration) error
	
	SetDebug(bool)
	
	ReadCoils(slave int8, startAddr int, nb int) ([]bool, error)
	ReadInputBits(slave int8, startAddr int, nb int) ([]bool, error)
	ReadHoldings(slave int8, startAddr int, nb int) ([]uint16, error)
	ReadInputRegisters(slave int8, startAddr int, nb int) ([]uint16, error)
	
	WriteSingleCoil(slave int8, addr int, value bool) error
	WriteHolding(slave int8, addr int, value uint16) error
	WriteCoils(slave int8, startAddr int, values []bool) error
	WriteHoldings(slave int8, startAddr int, values []uint16) error
}
