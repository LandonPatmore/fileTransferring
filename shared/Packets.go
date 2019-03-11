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
	Options  map[string]string
}

type DataPacket struct {
	Opcode      [] byte // 03
	BlockNumber [] byte
	Data        [] byte
}

type ACKPacket struct {
	Opcode      [] byte // 04/06
	BlockNumber [] byte
	IsOACK      bool
	Options     map[string]string
}

type ErrorPacket struct {
	Opcode       [] byte //05
	ErrorCode    [] byte //00 - 08
	ErrorMessage string
	zero         byte
}

type ArrayBytesHelper struct {
	// Helper so there isn't a need for a 2D array (even though it would probably be more efficient)
	Bytes [] byte
}

func CreateRRQWRQPacket(isRRQ bool, fileName string, options map[string]string) *RRQWRQPacket {
	var z RRQWRQPacket

	if isRRQ {
		z.Opcode = []byte{0, 1}
	} else {
		z.Opcode = []byte{0, 2}
	}

	z.Filename = fileName
	z.Mode = "octet"
	z.Options = options

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
	for k := range z.Options {
		byteArray = append(byteArray, []byte(k)...)
		byteArray = append(byteArray, 0)
		byteArray = append(byteArray, []byte(z.Options[k])...)
		byteArray = append(byteArray, 0)
	}

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
	if !a.IsOACK {
		byteArray = append(byteArray, a.BlockNumber...)
	} else {
		for k := range a.Options {
			byteArray = append(byteArray, []byte(k)...)
			byteArray = append(byteArray, 0)
			byteArray = append(byteArray, []byte(a.Options[k])...)
			byteArray = append(byteArray, 0)
		}
	}

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
	packet := RRQWRQPacket{}

	packet.Opcode = data[:2]

	var lastZeroSeen = 2
	var packetBytes [] ArrayBytesHelper
	for index, b := range data[2:] {
		if b == 0 {
			bytes := data[lastZeroSeen : index+2]
			lastZeroSeen = index + 3
			packetBytes = append(packetBytes, ArrayBytesHelper{bytes})
		}
	}

	if len(packetBytes) > 2 { // we now know its an option packet
		var options = packetBytes[2:]
		var optionsMapping = make(map[string]string)
		if len(options)%2 == 0 {
			for i := 0; i <= len(options)-2; i += 2 { // loop over and map the keys to values of the options
				optionsMapping[string(options[i].Bytes)] = string(options[i+1].Bytes)
			}

			packet.Options = optionsMapping

			return &packet, nil
		}
		return nil, errors.New("Options are missing values")
	}

	packet.Filename = string(packetBytes[0].Bytes)
	packet.Mode = string(packetBytes[1].Bytes)

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

func ReadOACKPacket(data []byte) (a *ACKPacket, err error) {
	packet := ACKPacket{}

	packet.Opcode = data[:2]

	var lastZeroSeen = 2
	var packetBytes [] ArrayBytesHelper
	for index, b := range data[2:] {
		if b == 0 {
			bytes := data[lastZeroSeen : index+2]
			lastZeroSeen = index + 3
			packetBytes = append(packetBytes, ArrayBytesHelper{bytes})
		}
	}

	if len(packetBytes) > 2 { // we now know its an option packet
		var options = packetBytes
		var optionsMapping = make(map[string]string)
		if len(options)%2 == 0 {
			for i := 0; i <= len(options)-2; i += 2 { // loop over and map the keys to values of the options
				optionsMapping[string(options[i].Bytes)] = string(options[i+1].Bytes)
			}

			packet.Options = optionsMapping

			return &packet, nil
		}
		return nil, errors.New("Options are missing values")
	}

	return &packet, nil
}

func ReadErrorPacket(data []byte) (e *ErrorPacket, err error) {
	packet := ErrorPacket{}

	packet.Opcode = data[:2]
	packet.ErrorCode = data[2:4]
	packet.ErrorMessage = string(data[4 : len(data)-1])

	return &packet, nil
}
