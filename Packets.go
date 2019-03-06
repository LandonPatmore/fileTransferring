package fileTransferring

// TFTP Implementation of packet types
// RFC-1350
type RRQWRQPacket struct {
	Opcode [2] byte
	Filename string
	Zero byte
	Mode string
	ZeroTwo byte
}

type DataPacket struct {
	Opcode [2] byte
	BlockNumber [2] byte
	Data [512] byte
}

type ACKPacket struct {
	Opcode [2] byte
	BlockNumber [2] byte
}

type ErrorPacket struct {
	Opcode [2] byte
	ErrorCode [2] byte
	ErrorMessage string
	Zero byte
}
