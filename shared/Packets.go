// TFTP Implementation of packet types (RFC-1350)
package shared

// Each packet includes fields including zero byte values
// so it is easier to understand what is going on as
// certain fields are chained together to create a packet
type RRQWRQPacket struct {
	opcode   [] byte // 01/02
	Filename string
	zero     byte
	mode     string // octet only for assignment
	zeroTwo  byte
}

type DataPacket struct {
	opcode      [] byte // 03
	BlockNumber [] byte
	Data        [] byte
}

type ACKPacket struct {
	opcode      [] byte //04
	BlockNumber [] byte
}

type ErrorPacket struct {
	opcode       [] byte //05
	ErrorCode    [] byte
	ErrorMessage string
	zero         byte
}

func CreateRRQWRQPacket(isRRQ bool) *RRQWRQPacket {
	var z RRQWRQPacket

	if isRRQ {
		z.opcode = []byte{0, 1}
	} else {
		z.opcode = []byte{0, 2}
	}

	z.mode = "octet"

	return &z
}

func CreateDataPacket() *DataPacket {
	var d DataPacket

	d.opcode = []byte{0, 3}
	d.Data = make([]byte, 0, 512)

	return &d
}

func CreateACKPacket() *ACKPacket {
	var a ACKPacket

	a.opcode = []byte{0, 4}

	return &a
}

func CreateErrorPacket() *ErrorPacket {
	var e ErrorPacket

	e.opcode = []byte{0, 5}

	return &e
}

func CreateRRQWRQPacketByteArray(z *RRQWRQPacket) [] byte {
	var byteArray []byte

	byteArray = append(byteArray, z.opcode...)
	byteArray = append(byteArray, z.Filename...)
	byteArray = append(byteArray, z.zero)
	byteArray = append(byteArray, z.mode...)
	byteArray = append(byteArray, z.zeroTwo)

	return byteArray
}

func CreateDataPacketByteArray(d *DataPacket) [] byte {
	var byteArray []byte

	byteArray = append(byteArray, d.opcode...)
	byteArray = append(byteArray, d.BlockNumber...)
	byteArray = append(byteArray, d.Data...)

	return byteArray
}

func CreateAckPacketByteArray(a *ACKPacket) [] byte {
	var byteArray []byte

	byteArray = append(byteArray, a.opcode...)
	byteArray = append(byteArray, a.BlockNumber...)

	return byteArray
}

func CreateErrorPacketByteArray(e *ErrorPacket) [] byte {
	var byteArray []byte

	byteArray = append(byteArray, e.opcode...)
	byteArray = append(byteArray, e.ErrorCode...)
	byteArray = append(byteArray, e.ErrorMessage...)
	byteArray = append(byteArray, e.zero)

	return byteArray
}
