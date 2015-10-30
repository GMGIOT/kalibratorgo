package main

import (
	//"reflect"
)


type ModbusRTUDevice struct {
	AbstarctDevice
	
	cells []AbstractCell
}

func NewModbusRTUDevice(DeviceMap []byte) (*ModbusRTUDevice, error) {
	/*
	re := reflect.ValueOf(DeviceMap)
	for i := 0; i < re.NumField(); i++ {
		//f := re.Field(i)
		
	}*/
	
	return nil, nil
}