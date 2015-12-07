package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

const (
	MB_CELL_ACCESS_INVALID = 0
	MB_CELL_ACCESS_READ    = 1 << 0
	MB_CELL_ACCESS_WRITE   = 1 << 1
	MB_CELL_ACCESS_RW      = MB_CELL_ACCESS_READ | MB_CELL_ACCESS_WRITE
)

var typeMap = map[string]reflect.Kind{
	"uint16_t": reflect.Uint16,
	"int16_t":  reflect.Int16,
	"uint32_t": reflect.Uint32,
	"int32_t":  reflect.Int32,
	"float":    reflect.Float32,
	"double":   reflect.Float64,
	"bool":     reflect.Bool,
}

var accessMap = map[string]int{
	"ReadOnly":  MB_CELL_ACCESS_READ,
	"WriteOnly": MB_CELL_ACCESS_WRITE,
	"ReadWrite": MB_CELL_ACCESS_RW,
}

type ModbusCell struct {
	AbstractCell

	pAddress     uint16
	ValueType    reflect.Kind
	Access       int
	pName        string
	pCategory    string
	pDescription string
	pVariants    map[string]interface{}
	Serializable bool
	hint         CellHint
}

type ModbusCellMap struct {
	Address          json.RawMessage `json:"Address"`
	ValueType        string          `json:"ValueType"`
	Access           string          `json:"Access"`
	Name             string          `json:"Name"`
	Category         string          `json:"Category"`
	Description      string          `json:"Description"`
	Variants         []interface{}   `json:"variants"`
	VariantsAdvanced []interface{}   `json:"variantsAdvanced"`
	Serializable     json.RawMessage `json:"Serializable"`
}

type AdvancedVariant struct {
	Display string      `json:"Display"`
	Value   interface{} `json:"value"`
}

func NewCell(hint CellHint) *ModbusCell {
	cell := &ModbusCell{Serializable: true}

	cell.hint = hint
	if hint != nil {
		hint.SetDefaults(cell)
	}

	return cell
}

func NewCellFromJSONMap(c *ModbusCell, m ModbusCellMap) error {
	var t int64
	var err error
	if strings.Contains(string(m.Address), "0x") {
		t, err = DecodeHex(string(m.Address))
	} else {
		t, err = strconv.ParseInt(string(m.Address), 0, 64)
	}
	if err != nil {
		return err
	} else {
		c.pAddress = uint16(t)
	}

	if m.ValueType != "" {
		c.ValueType = typeMap[m.ValueType]
		if c.Type == nil {
			return errors.New(fmt.Sprintf("Failed to parce type of cell \"%s\"", m.ValueType))
		}
	}

	if m.Access != "" {
		c.Access = accessMap[m.Access]
		if c.Access == 0 {
			return errors.New(fmt.Sprintf("Access info incorrect for cell \"%s\"", m.Access))
		}
	}

	c.pCategory = m.Category
	c.pName = m.Name
	c.pDescription = m.Description

	if string(m.Serializable) != "" {
		if err := json.Unmarshal(m.Serializable, &c.Serializable); err != nil {
			return err
		}
	}

	for _, v := range m.Variants {
		if s, ok := v.(string); ok {
			if c.ValueType != reflect.TypeOf(v).Kind() {
				return errors.New(fmt.Sprintf("Can't convert %v to cell type (%v)", v, c.ValueType))
			} else {
				c.pVariants[s] = v
			}
		} else {
			return errors.New(fmt.Sprint("%v inconvertable to sring", v))
		}
	}

	for _, v := range m.VariantsAdvanced {
		if v, ok := v.([]byte); ok {
			var varianStruct AdvancedVariant
			if err := json.Unmarshal(v, &varianStruct); err != nil {
				return err
			}
			if c.ValueType != reflect.TypeOf(varianStruct.Value).Kind() {
				var sv string
				if sv, ok = (varianStruct.Value).(string); !ok {
					panic("can't convert value to string")
				}
				if strings.Contains(sv, "0x") {
					if t, err = DecodeHex(sv); err == nil {
						c.pVariants[varianStruct.Display] = t
						goto OK
					}
				}
				return errors.New(fmt.Sprintf("Can't convert %v to cell type (%v)", varianStruct.Value, c.ValueType))
			} else {
				c.pVariants[varianStruct.Display] = varianStruct.Value
			}
		} else {
			return errors.New(fmt.Sprintf("can't parce variant \"%v\"", v))
		}
	}
OK:
	return c.hint.Validete(c)
}

func DecodeCellsJSON(data json.RawMessage, hint CellHint) ([]ModbusCell, error) {
	var result []ModbusCell

	var cells []ModbusCellMap
	if err := json.Unmarshal(([]byte)(data), &cells); err != nil {
		return nil, err
	}
	for n, cell := range cells {
		c := NewCell(hint)
		if err := NewCellFromJSONMap(c, cell); err != nil {
			return nil, errors.New(fmt.Sprintf("Failed to process cell %d: %s", n, err.Error()))
		} else {
			result = append(result, *c)
		}
	}

	return result, nil
}

func (this ModbusCell) ID() string {
	return fmt.Sprintf("%s_%s%d", this.Parent().ID(), this.Type(), this.Address)
}

func (this ModbusCell) Type() string {
	switch this.hint.(type) {
	case HoldingCellHint:
		return "H"
	case InputCellHint:
		return "I"
	case CoilCellHint:
		return "C"
	case DiscreteInputCellHint:
		return "D"
	default:
		return ""
	}
}

func (this ModbusCell) Address() uint16 {
	return this.pAddress
}

func (this ModbusCell) Name() string {
	return this.pName
}

func (this ModbusCell) Category() string {
	return this.pCategory
}

func (this ModbusCell) Description() string {
	return this.pDescription
}

func (this ModbusCell) Variants() map[string]interface{} {
	return this.pVariants
}

func (this ModbusCell) Close() {
	
}