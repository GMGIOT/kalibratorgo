package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type ModbusRTUDevice struct {
	AbstarctDevice

	parent                      Connection
	cells                       []ModbusCell
	DevID                       uint16
	Class                       string
	Description                 string
	MaxSimulateneousCellsToRead int
}

type ModbusRTUDeviceJSONMap struct {
	DevID                       json.RawMessage `json:"DevID"`
	Class                       string          `json:"Class"`
	Description                 string          `json:"Description"`
	MaxSimulateneousCellsToRead int             `json:"maxSimulateneousCellsToRead"`
	HoldingRegisters            json.RawMessage `json:"HoldingRegisters"`
	InputRegisters              json.RawMessage `json:"InputRegisters"`
	Coils                       json.RawMessage `json:"Coils"`
	DiscreteInputs              json.RawMessage `json:"DiscreteInputs"`
}

type CellMapPair struct {
	value *json.RawMessage
	hint  CellHint
}

func NewModbusRTUDeviceFromJSON(DeviceMap []byte) (*ModbusRTUDevice, error) {
	var devmap ModbusRTUDeviceJSONMap
	var err error
	var t int64

	if err = json.Unmarshal(DeviceMap, &devmap); err != nil {
		return nil, err
	}

	pairs := [...]CellMapPair{{&devmap.HoldingRegisters, &HoldingCellHint{}},
		{&devmap.InputRegisters, &InputCellHint{}},
		{&devmap.Coils, &CoilCellHint{}},
		{&devmap.DiscreteInputs, &DiscreteInputCellHint{}}}

	result := &ModbusRTUDevice{
		Class: devmap.Class, Description: devmap.Description,
		MaxSimulateneousCellsToRead: devmap.MaxSimulateneousCellsToRead}

	if strings.Contains(string(devmap.DevID), "0x") {
		t, err = DecodeHex(string(devmap.DevID))
	} else {
		t, err = strconv.ParseInt(string(devmap.DevID), 10, -1)
	}
	if err != nil {
		return nil, err
	} else {
		result.DevID = uint16(t)
	}

	for _, f := range pairs {
		if cells, err := DecodeCellsJSON(*f.value, f.hint); err == nil {
			result.cells = append(result.cells, cells...)
		} else {
			return nil, err
		}
	}

	return result, nil
}

func (this ModbusRTUDevice) ID() string {
	return fmt.Sprintf("%s_%d", this.Connection().ID, this.DevID)
}

func (this ModbusRTUDevice) Protocol() string {
	return str_ModbusRTU
}

func (this ModbusRTUDevice) Connection() Connection {
	return this.parent
}

func (this ModbusRTUDevice) BindTo(c Connection) error {
	if this.parent != nil {
		if p, ok := (this.parent).(ModbusRTUConnection); !ok {
			panic("parent !mast! be instance of ModbusRTUConnection or nil")
		} else {
			p.RemoveDevice(&this)
		}
	}
	
	if mbc, ok := c.(ModbusRTUConnection); ok {
		mbc.AttachedDevice(&this)
		return nil
	} else {
		return errors.New(fmt.Sprintf("Connection id %d has inapplicatiable protocol %s", c.ID(),
			c.Protocol()))
	}
}

func (this ModbusRTUDevice) Close(recurcive bool) error {
	if recurcive {
		for _, c := range this.cells {
			c.Close()
		}
	}

	return nil
}
