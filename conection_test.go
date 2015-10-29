package main

import (
	"testing"
	"runtime"
	"time"
	"github.com/ololoshka2871/serialdeviceenumeratorgo"
)

func TestOpen(t *testing.T) {
	const (
		speed = 57600
		mode = "8N1"
	)
	
	ports, err := serialdeviceenumeratorgo.Enumerate()
	if err != nil {
		t.Error(err)
		return
	}
	
	for _, p := range ports {
		NewModbusRTUConnection(p.Name, speed, mode, nil)	
	}
	ModbusRTUConnections[0].Close(false)
	time.Sleep(time.Second)
	runtime.GC()
}
