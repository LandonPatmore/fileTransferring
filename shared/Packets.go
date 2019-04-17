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

type SlidingWindowPacket struct {
	Opcode [] byte // 07
}

// Helper so there isn't a need for a 2D array (even though it would probably be more efficient)
type ArrayBytesHelper struct {
	Bytes [] byte
}

// Creates a RRQ or WRQ Packet
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

// Creates a Data Packet
func CreateDataPacket(blockNumber [] byte, data [] byte) *DataPacket {
	var d DataPacket

	d.Opcode = []byte{0, 3}
	d.BlockNumber = blockNumber
	d.Data = data

	return &d
}

// Creates an ACK Packet
func CreateACKPacket() *ACKPacket {
	var a ACKPacket

	a.Opcode = []byte{0, 4}

	return &a
}

// Creates an Error Packet
func CreateErrorPacket(errorCode [] byte, errorMessage string) *ErrorPacket {
	var e ErrorPacket

	e.Opcode = []byte{0, 5}
	e.ErrorCode = errorCode
	e.ErrorMessage = errorMessage

	return &e
}

func CreateSlidingWindowPacket() *SlidingWindowPacket {
	var sw SlidingWindowPacket

	sw.Opcode = []byte{0, 7}

	return &sw
}

// Returns a byte array of a RRQ or WRQ Packet
func (z *RRQWRQPacket) ByteArray() [] byte {
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

// Returns a byte array of a Data Packet
func (d *DataPacket) ByteArray() [] byte {
	var byteArray []byte

	byteArray = append(byteArray, d.Opcode...)
	byteArray = append(byteArray, d.BlockNumber...)
	byteArray = append(byteArray, d.Data...)

	return byteArray
}

// Returns a byte array of an ACK Packet
func (a *ACKPacket) ByteArray() [] byte {
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

// Returns a byte array of an Error Packet
func (e *ErrorPacket) ByteArray() [] byte {
	var byteArray []byte

	byteArray = append(byteArray, e.Opcode...)
	byteArray = append(byteArray, e.ErrorCode...)
	byteArray = append(byteArray, e.ErrorMessage...)
	byteArray = append(byteArray, e.zero)

	return byteArray
}

func (sw *SlidingWindowPacket) ByteArray() [] byte {
	var byteArray []byte

	byteArray = append(byteArray, sw.Opcode...)

	return byteArray
}

// Reads a data array and returns an RRQ or WRQ packet with a possible error as well
func ReadRRQWRQPacket(data []byte) (p *RRQWRQPacket, err error) {
	packet := RRQWRQPacket{}

	packet.Opcode = data[:2]

	var lastZeroSeen = 2
	var packetBytes [] ArrayBytesHelper
	for index, b := range data[2:] {
		if b == 0 {
			dataBytes := data[lastZeroSeen : index+2]
			lastZeroSeen = index + 3
			packetBytes = append(packetBytes, ArrayBytesHelper{dataBytes})
		}
	}

	if packetBytes != nil {
		packet.Filename = string(packetBytes[0].Bytes)
		packet.Mode = string(packetBytes[1].Bytes)

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

		return &packet, nil
	}

	return nil, errors.New("Error parsing the packet...")
}

// Reads a data array and returns a Data packet with a possible error as well
func ReadDataPacket(data []byte) (d *DataPacket, err error) {
	packet := DataPacket{}

	packet.Opcode = data[:2]
	packet.BlockNumber = data[2:4]
	packet.Data = data[4:]

	return &packet, nil
}

// Reads a data array and returns an ACK packet with a possible error as well
func ReadACKPacket(data []byte) (a *ACKPacket, err error) {
	packet := ACKPacket{}

	packet.Opcode = data[:2]
	packet.BlockNumber = data[2:4]

	return &packet, nil
}

// Reads a data array and returns an OACK/ACK packet with a possible error as well
func ReadOACKPacket(data []byte) (a *ACKPacket, err error) {
	packet := ACKPacket{}

	packet.Opcode = data[:2]
	packet.IsOACK = true

	var lastZeroSeen = 2
	var packetBytes [] ArrayBytesHelper
	for index, b := range data[2:] {
		if b == 0 {
			bytes := data[lastZeroSeen : index+2]
			lastZeroSeen = index + 3
			packetBytes = append(packetBytes, ArrayBytesHelper{bytes})
		}
	}

	if len(packetBytes) >= 2 {
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

// Reads a data array and returns an Error packet with a possible error as well
func ReadErrorPacket(data []byte) (e *ErrorPacket, err error) {
	packet := ErrorPacket{}

	packet.Opcode = data[:2]
	packet.ErrorCode = data[2:4]
	packet.ErrorMessage = string(data[4 : len(data)-1])

	return &packet, nil
}
