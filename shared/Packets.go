// TFTP Implementation of packet types (RFC-1350)
package shared

// Each packet includes fields including zero byte values
// so it is easier to understand what is going on as
// certain fields are chained together to create a packet
type RRQWRQPacket struct {
	Opcode   [] byte // 01/02
	Filename string
	Zero     byte
	Mode     string // octet only for assignment
	ZeroTwo  byte
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
	ErrorCode    [] byte
	ErrorMessage string
	Zero         byte
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

func CreateErrorPacket() *ErrorPacket {
	var e ErrorPacket

	e.Opcode = []byte{0, 5}

	return &e
}
