package main

import (
	"encoding/json"
	"testing"
	"reflect"
	"os"
	"fmt"
)

func TestReflection(t *testing.T) {
	fi, err := os.Open("memmaps/DB00.json")
    if err != nil {
        t.Error(err)
        return
    }
    defer func() {
        if err := fi.Close(); err != nil {
            t.Error(err)
        }
    }()
    
    stat, err := fi.Stat()
    if err != nil {
    	t.Error(err)
        return
    }
    data := make([]byte, stat.Size())
    if n, err := fi.Read(data); err != nil || n != cap(data) {
    	t.Error("Read file error")
    	return
    }
    var jsonData map[string]interface{}
    if err = json.Unmarshal(data, &jsonData); err != nil {
    	t.Error(err)
    	return
    } 
    
    s := reflect.ValueOf(jsonData)
    typeOfT := s.Type()
	for i := 0; i < s.NumField(); i++ {
	    f := s.Field(i)
	    fmt.Printf("%d: %s %s = %v\n", i,
	        typeOfT.Field(i).Name, f.Type(), f.Interface())
    }
}


