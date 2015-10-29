package main

/*
#include <modbus.h>
#include <errno.h>

int getErrno() {
	return errno;
}
*/
import "C"

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"runtime"
	"log"
	"time"
	"math"
	"unsafe"
)

const (
	str_ModbusRTU = "ModbusRTU"
	RTU_CONST_FIELDS_LEN = 4
	
	RTU_01_LEN = 4
	RTU_02_LEN = RTU_01_LEN
	RTU_03_LEN = RTU_01_LEN
	RTU_05_LEN = 4
	RTU_06_LEN = RTU_05_LEN
	RTU_0F_LEN = 5
	RTU_10_LEN = RTU_0F_LEN
)

var parser = regexp.MustCompile(`(\d)([A-Z])(\d)`) 

var index = 0

type ioCtlHook interface {
	OnStartTransmitting()
	OnEndTransmitting()
}

type ModbusRTUConnection struct {
	Connection
	id 					int
	deviceName			string
	speed				int
	serialMode			string
	attachedDevicese	[](*ModbusRTUDevice)
	hook				ioCtlHook
	
	ctx					*C.modbus_t
	
	rwCtlChan			chan int
}

var ModbusRTUConnections [](*ModbusRTUConnection)

func NewModbusRTUConnection(DeviceName string,
	Speed int, SerialMode string, Hook ioCtlHook) (*ModbusRTUConnection, error) {
	// is allready opened?
	for _, connection := range ModbusRTUConnections {
		if connection.Device() == DeviceName {
			if connection.serialMode == SerialMode {
				return connection, nil	
			} else {
				return nil, errors.New(
					fmt.Sprintf("Device '%s' used by connection %d, mode '%s'",
					DeviceName, connection.ID(), connection.serialMode))
			}
		}
	}
	
	result := &ModbusRTUConnection{ deviceName : DeviceName, speed : Speed, 
		serialMode : SerialMode, hook : Hook }
	
	if Hook != nil {
		result.rwCtlChan = make(chan int)
		
		go func(oneByteTime time.Duration) {
			for v := range result.rwCtlChan {
				if v < 0 {
					break
				}
				result.hook.OnStartTransmitting()
				time.Sleep(oneByteTime * time.Duration(v))
				result.hook.OnEndTransmitting()
			}
		}(time.Duration(math.Ceil(float64(time.Second * (1 + 8 + 0 + 1)) / float64(result.speed))))
	}
	
	// try open
	match := parser.FindStringSubmatch(result.serialMode)
	if len(match) != (1 + 3) {
		result.ctx = C.modbus_new_rtu(C.CString(result.deviceName), C.int(result.speed),
			C.char('N'), C.int(8), C.int(1))
	} else {
		databits, _ := strconv.ParseInt(match[1 + 0], 10, 32)
		stopbits, _ := strconv.ParseInt(match[1 + 2], 10, 32)
		result.ctx = C.modbus_new_rtu(C.CString(result.deviceName), C.int(result.speed),
			C.char(match[1 + 1][0]),
			C.int(databits),
			C.int(stopbits))
	}
	if result.ctx == nil {
		return nil, errors.New(fmt.Sprintf("Unable to create the libmodbus context\n%s",
				C.GoString(C.modbus_strerror(C.getErrno()))))
	}
	runtime.SetFinalizer(result, finaliserModbusRTUConnection)
	
	result.id = index
	index++
	ModbusRTUConnections = append(ModbusRTUConnections, result)
	
	return result, nil
}

func (this *ModbusRTUConnection) ID() int {
	return this.id
}

func (this *ModbusRTUConnection) Type() string {
	return str_Serial
}

func (this *ModbusRTUConnection) Device() string {
	return this.deviceName
}

func (this *ModbusRTUConnection) Options() map[string]interface{} {
	return map[string]interface{}{"speed" : this.speed, "mode" : this.serialMode } 
}

func (this *ModbusRTUConnection) Protocol() string {
	return str_ModbusRTU
}

func (this *ModbusRTUConnection) Status() int {
	return Connection_unknown
}

func (this *ModbusRTUConnection) UsedBy() int {
	return len(this.attachedDevicese)
}

func finaliserModbusRTUConnection(obj *ModbusRTUConnection) {
	log.Print("Cleaning libmodbus structure")
	C.modbus_close(obj.ctx)
	C.modbus_free(obj.ctx)
	if obj.hook != nil {
		obj.rwCtlChan <- (-1)
	}
}

func (this *ModbusRTUConnection) Timeout() time.Duration {
	var sec, usec C.uint32_t
	C.modbus_get_response_timeout(this.ctx, &sec, &usec)
	return time.Duration(sec) * time.Second + time.Duration(usec) * time.Microsecond
}
	
func (this *ModbusRTUConnection) SetTimeout(timeout time.Duration) error {
	var sec C.uint32_t = C.uint32_t(timeout / time.Second)
	var usec C.uint32_t = C.uint32_t((timeout - time.Duration(sec)) / time.Microsecond)
	if C.modbus_set_response_timeout(this.ctx, sec, usec) == -1 {
		return errors.New("Invalid timeout value")
	}
	return nil
}

func (this *ModbusRTUConnection) SetDebug(flag bool) {
	var f C.int = 0
	if flag {
		f = 1
	}
	C.modbus_set_debug(this.ctx, f)
}

func (this *ModbusRTUConnection) ProcessHook(messageLen int) {
	messageLen += RTU_CONST_FIELDS_LEN
	this.rwCtlChan <- messageLen
}

func (this *ModbusRTUConnection) ReadCoils(slave int8, startAddr int, nb int) ([]bool, error) {
	if C.modbus_set_slave(this.ctx, C.int(slave)) != 0 {
		return nil, errors.New(fmt.Sprintf("Invalid slave id %d", slave))
	}
	
	var space [65535]C.uint8_t
	 
	this.ProcessHook(RTU_01_LEN)
	if ressived := C.modbus_read_bits(this.ctx, C.int(startAddr), C.int(nb),
		&(space[0])); int(ressived) != nb {
		if ressived == -1 {
			errno := C.getErrno()
			if errno == C.EMBMDATA {
				return nil, errors.New(C.GoString(C.modbus_strerror(errno)))	
			}
			return nil, errors.New(fmt.Sprintf("Unknown modbus error errno=%d", errno))
		} else if ressived == 0 {
			return nil, errors.New("No ansver ressived")
		}
	} else {
		result := make([]bool, ressived)
		for i := 0; i < int(ressived); i++ {
			if space[i] != C.FALSE {
				result[i] = true
			}
		}
		return result, nil
	}
	
	return nil, nil
}

func (this *ModbusRTUConnection) ReadInputBits(slave int8, startAddr int, nb int) ([]bool, error) {
	if C.modbus_set_slave(this.ctx, C.int(slave)) != 0 {
		return nil, errors.New(fmt.Sprintf("Invalid slave id %d", slave))
	}
	
	var space [65535]C.uint8_t
	 
	this.ProcessHook(RTU_02_LEN)
	if ressived := C.modbus_read_input_bits(this.ctx, C.int(startAddr), C.int(nb),
		&(space[0])); int(ressived) != nb {
		if ressived == -1 {
			errno := C.getErrno()
			if errno == C.EMBMDATA {
				return nil, errors.New(C.GoString(C.modbus_strerror(errno)))
			}
			return nil, errors.New(fmt.Sprintf("Unknown modbus error errno=%d", int(errno)))
		} else if ressived == 0 {
			return nil, errors.New("No ansver ressived")
		}
	} else {
		result := make([]bool, ressived)
		for i := 0; i < int(ressived); i++ {
			if space[i] != C.FALSE {
				result[i] = true
			}
		}
		return result, nil
	}
	
	return nil, nil
}

func (this *ModbusRTUConnection) ReadHoldings(slave int8, startAddr int, nb int) ([]uint16, error) {
	if C.modbus_set_slave(this.ctx, C.int(slave)) != 0 {
		return nil, errors.New(fmt.Sprintf("Invalid slave id %d", slave))
	}
	
	var space [65535]C.uint16_t
	 
	this.ProcessHook(RTU_03_LEN)
	if ressived := C.modbus_read_registers(this.ctx, C.int(startAddr),
		C.int(nb), &(space[0])); int(ressived) != nb {
		if ressived == -1 {
			errno := C.getErrno()
			if errno == C.EMBMDATA {
				return nil, errors.New(C.GoString(C.modbus_strerror(errno)))
			}
			return nil, errors.New(fmt.Sprintf("Unknown modbus error errno=%d", errno))
		} else if ressived == 0 {
			return nil, errors.New("No ansver ressived")
		}
	} else {
		result := make([]uint16, ressived)
		for i := 0; i < int(ressived); i++ {
			result[i] = uint16(space[i])
		}
		return result, nil
	}
	
	return nil, nil
}

func (this *ModbusRTUConnection) ReadInputRegisters(slave int8, startAddr int, nb int) ([]uint16, error) {
	if C.modbus_set_slave(this.ctx, C.int(slave)) != 0 {
		return nil, errors.New(fmt.Sprintf("Invalid slave id %d", slave))
	}
	
	var space [65535]C.uint16_t
	 
	this.ProcessHook(RTU_03_LEN)
	if ressived := C. modbus_read_input_registers(this.ctx, C.int(startAddr),
		C.int(nb), &(space[0])); int(ressived) != nb {
		if ressived == -1 {
			errno := C.getErrno()
			if errno == C.EMBMDATA {
				return nil, errors.New(C.GoString(C.modbus_strerror(errno)))
			}
			return nil, errors.New(fmt.Sprintf("Unknown modbus error errno=%d", errno))
		} else if ressived == 0 {
			return nil, errors.New("No ansver ressived")
		}
	} else {
		result := make([]uint16, ressived)
		for i := 0; i < int(ressived); i++ {
			result[i] = uint16(space[i])
		}
		return result, nil
	}
	
	return nil, nil
}

func (this *ModbusRTUConnection) WriteSingleCoil(slave int8, addr int, value bool) error {
	if C.modbus_set_slave(this.ctx, C.int(slave)) != 0 {
		return errors.New(fmt.Sprintf("Invalid slave id %d", slave))
	}
	
	var v C.int
	if value {
		v = C.TRUE
	}
	
	this.ProcessHook(RTU_05_LEN)
	if C.modbus_write_bit(this.ctx, C.int(addr), v) != 1 {
		return errors.New(C.GoString(C.modbus_strerror(C.getErrno())))
	}
	
	return nil
}

func (this *ModbusRTUConnection) WriteHolding(slave int8, addr int, value uint16) error {
	if C.modbus_set_slave(this.ctx, C.int(slave)) != 0 {
		return errors.New(fmt.Sprintf("Invalid slave id %d", slave))
	}
	
	this.ProcessHook(RTU_06_LEN)
	if C.modbus_write_register(this.ctx, C.int(addr), C.int(value)) != 1 {
		return errors.New(C.GoString(C.modbus_strerror(C.getErrno())))
	}
	
	return nil
}

func (this *ModbusRTUConnection) WriteCoils(slave int8, startAddr int, values []bool) error {
	if C.modbus_set_slave(this.ctx, C.int(slave)) != 0 {
		return errors.New(fmt.Sprintf("Invalid slave id %d", slave))
	}
	
	vals := make([]C.uint8_t, len(values))
	for i, v := range values {
		if v {
			vals[i] = C.TRUE
		}
	}
	
	this.ProcessHook(RTU_0F_LEN + int(math.Ceil(float64(len(values)) / 8.0)))
	if C.modbus_write_bits(this.ctx, C.int(startAddr), C.int(len(values)), &vals[0]) < 0 {
		return errors.New(C.GoString(C.modbus_strerror(C.getErrno())))		
	}
	return nil
}

func (this *ModbusRTUConnection) WriteHoldings(slave int8, startAddr int, values []uint16) error {
	if C.modbus_set_slave(this.ctx, C.int(slave)) != 0 {
		return errors.New(fmt.Sprintf("Invalid slave id %d", slave))
	}
	
	vals := make([]C.uint16_t, len(values))
	for i, v := range values {
		vals[i] = C.uint16_t(v)
	}
	
	this.ProcessHook(RTU_10_LEN + len(values) * int(unsafe.Sizeof(vals[0])))
	if C.modbus_write_registers(this.ctx, C.int(startAddr), C.int(len(values)), &vals[0]) < 0 {
		return errors.New(C.GoString(C.modbus_strerror(C.getErrno())))		
	}
	return nil
}

func (this *ModbusRTUConnection) AttachedDevice(dev AbstarctDevice) error {
	if d, ok := dev.(ModbusRTUDevice); ok {
		if d.ConnectionID() != this.ID() {
			this.attachedDevicese = append(this.attachedDevicese, &d)
			d.SetConnection(this)
		} else {
			return errors.New(fmt.Sprintf("Device allready attached to connection %d", this.ID()))
		}
	}
	return errors.New(fmt.Sprintf("Device requiered incompotable protocol %s", dev.Protocol()))
}

func (this *ModbusRTUConnection) Close(recursive bool) error {
	if recursive {
		for _, dev := range this.attachedDevicese {
			dev.Close(true)
		}
		return nil
	} else {
		if this.UsedBy() > 0 {
			return errors.New("Connection still used")
		}
		var n int
		var item *ModbusRTUConnection
		for n, item = range ModbusRTUConnections {
			if item == this {
				break
			}
		}
		ModbusRTUConnections, ModbusRTUConnections[len(ModbusRTUConnections)-1] =
			append(ModbusRTUConnections[:n], ModbusRTUConnections[n+1:]...), nil
		return nil
	}
}