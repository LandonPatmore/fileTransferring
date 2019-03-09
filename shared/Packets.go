// TFTP Implementation of packet types (RFC-1350)
package shared

import (
	"github.com/pkg/errors"
)

// Each packet includes fields including zero byte values
// so it is easier to understand what is going on as
// certain fields are chained together to create a packet
type RRQWRQPacket struct {
	Opcode   [] byte // 01/02
	Filename string
	zero     byte
	Mode     string // octet only for assignment
	zeroTwo  byte
}

type DataPacket struct {
	Opcode      [] byte // 03
	BlockNumber [] byte
	Data        [] byte
}

type ACKPacket struct {
	Opcode      [] byte //04
	BlockNumber [] byte
}

type ErrorPacket struct {
	Opcode       [] byte //05
	ErrorCode    [] byte //00 - 07
	ErrorMessage string
	zero         byte
}

func CreateRRQWRQPacket(isRRQ bool) *RRQWRQPacket {
	var z RRQWRQPacket

	if isRRQ {
		z.Opcode = []byte{0, 1}
	} else {
		z.Opcode = []byte{0, 2}
	}

	z.Mode = "octet"

	return &z
}

func CreateDataPacket() *DataPacket {
	var d DataPacket

	d.Opcode = []byte{0, 3}
	d.Data = make([]byte, 0, 512)

	return &d
}

func CreateACKPacket() *ACKPacket {
	var a ACKPacket

	a.Opcode = []byte{0, 4}

	return &a
}

func CreateErrorPacket(errorCode [] byte, errorMessage string) *ErrorPacket {
	var e ErrorPacket

	e.Opcode = []byte{0, 5}
	e.ErrorCode = errorCode
	e.ErrorMessage = errorMessage

	return &e
}

func CreateRRQWRQPacketByteArray(z *RRQWRQPacket) [] byte {
	var byteArray []byte

	byteArray = append(byteArray, z.Opcode...)
	byteArray = append(byteArray, z.Filename...)
	byteArray = append(byteArray, z.zero)
	byteArray = append(byteArray, z.Mode...)
	byteArray = append(byteArray, z.zeroTwo)

	return byteArray
}

func CreateDataPacketByteArray(d *DataPacket) [] byte {
	var byteArray []byte

	byteArray = append(byteArray, d.Opcode...)
	byteArray = append(byteArray, d.BlockNumber...)
	byteArray = append(byteArray, d.Data...)

	return byteArray
}

func CreateAckPacketByteArray(a *ACKPacket) [] byte {
	var byteArray []byte

	byteArray = append(byteArray, a.Opcode...)
	byteArray = append(byteArray, a.BlockNumber...)

	return byteArray
}

func CreateErrorPacketByteArray(e *ErrorPacket) [] byte {
	var byteArray []byte

	byteArray = append(byteArray, e.Opcode...)
	byteArray = append(byteArray, e.ErrorCode...)
	byteArray = append(byteArray, e.ErrorMessage...)
	byteArray = append(byteArray, e.zero)

	return byteArray
}

func ReadRRQWRQPacket(data []byte) (p *RRQWRQPacket, err error) {
	// TODO: Figure out where to throw error packet here instead of error
	packet := RRQWRQPacket{}

	packet.Opcode = data[:2]

	var firstZeroFound bool
	var modeStart int

	for index, b := range data[2:] {
		if !firstZeroFound {
			if b == 0 { // found the first one
				packet.Filename = string(data[2 : index+2])
				modeStart = index + 3
				firstZeroFound = true
			}
		} else {
			if b == 0 { // now found the second one
				packet.Mode = string(data[modeStart : index+2])
				return &packet, nil
			}
		}
	}

	return nil, errors.New("There was an error parsing the packet")
}

func ReadDataPacket(data []byte) (d *DataPacket, err error) {
	// TODO: Figure out where to throw error packet here
	packet := DataPacket{}

	packet.Opcode = data[:2]
	packet.BlockNumber = data[2:4]
	packet.Data = data[4:]

	return &packet, nil
}

func ReadACKPacket(data []byte) (a *ACKPacket, err error) {
	packet := ACKPacket{}

	packet.Opcode = data[:2]
	packet.BlockNumber = data[2:4]

	return &packet, nil
}

func ReadErrorPacket(data []byte) (e *ErrorPacket, err error) {
	packet := ErrorPacket{}

	packet.Opcode = data[:2]
	packet.ErrorCode = data[2:4]
	packet.ErrorMessage = string(data[4 : len(data)-1])

	return &packet, nil
}
