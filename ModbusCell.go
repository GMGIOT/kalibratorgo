package main

import (
	"encoding/json"
	"errors"
	"fmt"
)

type ModbusCell struct {
	AbstractCell
	
	Address				uint16	
	ValueType     		string					
    Access        		string					
    Name          		string					
    Category      		string					
    Description			string							
    Variants			map[string]interface{}	
    Serializable		bool					
}

type ModbusCellMap struct {
	Address				json.RawMessage			`json:"Address"`
	ValueType     		string					`json:"ValueType"`
    Access        		string					`json:"Access"`
    Name          		string					`json:"Name"`
    Category      		string					`json:"Category"`
    Description			string					`json:"Description"`
    Variants			[]interface{}			`json:"variants"`
    VariantsAdvanced	[]interface{}			`json:"variantsAdvanced"`
    Serializable		bool					`json:"Serializable"`
}


func NewCell() (*ModbusCell, error) {
	return &ModbusCell{}, nil
}

func NewCellFromMap(m ModbusCellMap) (*ModbusCell, error) {
	result := &ModbusCell{ Access : m.Access, Name : m.Name, Category : m.Category,
		 Description : m.Description, Serializable : m.Serializable }
	
	
	
	return result, nil
}

func DecodeCells(data json.RawMessage) ([]ModbusCell, error) {
	var result []ModbusCell
	
	var cells []ModbusCellMap
	if err := json.Unmarshal(([]byte)(data), &cells); err != nil {
		return nil, err
	}
	for n, cell := range cells {
		if c, err := NewCellFromMap(cell); err != nil {
			return nil, errors.New(fmt.Sprintf("Failed to process cell %d: %s", n, err.Error()))
		} else {
			result = append(result, *c)
		}
	}
	
	return result, nil
}