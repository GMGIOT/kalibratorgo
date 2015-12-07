package main

import (
	"encoding/json"
	"testing"
	"os"
	"fmt"
	"errors"
)

func ReadJsonFile(filename string) ([]byte, error) {
	fi, err := os.Open(filename)
    if err != nil {
        return nil, err
    }
    defer fi.Close()
    
    stat, err := fi.Stat()
    if err != nil {
    	return nil, err
    }
    data := make([]byte, stat.Size())
    if n, err := fi.Read(data); err != nil || n != cap(data) {
    	return nil, errors.New("Read file error")
    }
    return data, nil
}

func TestParce(t *testing.T) {
	data, err := ReadJsonFile("memmaps/DB00.json")
	if err != nil {
		t.Error(err)
    	return
	}
	
    var jsonData ModbusRTUDeviceJSONMap
    if err = json.Unmarshal(data, &jsonData); err != nil {
    	t.Error(err)
    	return
    }
    
    if v, err := DecodeHex(string(jsonData.DevID)); err != nil {
    	t.Error(err)
    } else {
    	t.Log(fmt.Sprintf("Found ID = %d", v))
    }
    // decodong cells
    var cells []ModbusCellMap
    if err = json.Unmarshal(([]byte)(jsonData.HoldingRegisters), &cells); err != nil {
    	t.Error(err)
    	return
    }
    t.Logf("Found %d cells", len(cells))
}

