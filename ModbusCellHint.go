package main

import (
	"reflect"
	"errors"
	"fmt"
)

type CellHint interface {
	SetDefaults(cell *ModbusCell)
	Validete(cell *ModbusCell) error
}

type HoldingCellHint struct{}
type InputCellHint struct{}
type CoilCellHint struct{}
type DiscreteInputCellHint struct{}

func validateComon(cell *ModbusCell) error {
	if cell.Name() == "" {
		return  errors.New(fmt.Sprintf("The cell id=%v has empty name", cell.ID()))
	}
	return nil
}

func (this HoldingCellHint) SetDefaults(cell *ModbusCell) {
	cell.Access = MB_CELL_ACCESS_RW
	cell.ValueType = reflect.Uint16
}

func (this HoldingCellHint) Validete(cell *ModbusCell) error {
	if err := validateComon(cell); err != nil {
		return err
	}
	return nil
}

func (this InputCellHint) SetDefaults(cell *ModbusCell) {
	cell.Access = MB_CELL_ACCESS_READ
	cell.ValueType = reflect.Uint16
}

func (this InputCellHint) Validete(cell *ModbusCell) error {
	if err := validateComon(cell); err != nil {
		return err
	}
	return nil
}

func (this CoilCellHint) SetDefaults(cell *ModbusCell) {
	cell.Access = MB_CELL_ACCESS_RW
	cell.ValueType = reflect.Bool
}

func (this CoilCellHint) Validete(cell *ModbusCell) error {
	if err := validateComon(cell); err != nil {
		return err
	}
	return nil
}

func (this DiscreteInputCellHint) SetDefaults(cell *ModbusCell) {
	cell.Access = MB_CELL_ACCESS_READ
	cell.ValueType = reflect.Bool
}

func (this DiscreteInputCellHint) Validete(cell *ModbusCell) error {
	if err := validateComon(cell); err != nil {
		return err
	}
	return nil
}
