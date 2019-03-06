// TFTP Implementation of packet types (RFC-1350)
package shared

// Each packet includes fields including zero byte values
// so it is easier to understand what is going on as
// certain fields are chained together to create a packet
type RRQWRQPacket struct {
	Opcode   [2] byte // 01/02
	Filename string
	Zero     byte
	Mode     string // octet only for assignment
	ZeroTwo  byte
}

type DataPacket struct {
	Opcode      [2] byte // 03
	BlockNumber [2] byte
	Data        [512] byte
}

type ACKPacket struct {
	Opcode      [2] byte //04
	BlockNumber [2] byte
}

type ErrorPacket struct {
	Opcode       [2] byte //05
	ErrorCode    [2] byte
	ErrorMessage string
	Zero         byte
}

func CreateRRQWRQPacket(isRRQ bool) *RRQWRQPacket {
	var z RRQWRQPacket

	if isRRQ {
		z.Opcode = [2]byte{0, 1}
	} else {
		z.Opcode = [2]byte{0, 2}
	}

	z.Mode = "octet"

	return &z
}

func CreateDataPacket() *DataPacket {
	var d DataPacket

	d.Opcode = [2]byte{0, 3}

	return &d
}

func CreateACKPacket() *ACKPacket {
	var a ACKPacket

	a.Opcode = [2]byte{0, 4}

	return &a
}

func CreateErrorPacket() *ErrorPacket {
	var e ErrorPacket

	e.Opcode = [2]byte{0, 5}

	return &e
}
