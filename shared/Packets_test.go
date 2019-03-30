package shared

import "testing"

func TestCreateRRQWRQPacket1(t *testing.T) {
	const testFile = "test.txt"
	const mode = "octet"
	var opCode = [] byte{0, 1}

	packet := CreateRRQWRQPacket(true, testFile, nil)

	if !BlockNumberChecker(packet.Opcode, opCode) {
		t.Errorf("Got = %v; want %v", packet.Opcode, opCode)
	}

	if packet.Filename != testFile {
		t.Errorf("Got = %s; want %s", packet.Filename, testFile)
	}

	if packet.Mode != mode {
		t.Errorf("Got = %s; want %s", packet.Mode, mode)
	}

	if packet.Options != nil {
		t.Errorf("Got = %v; want nil %v", packet.Options, nil)
	}
}

func TestCreateRRQWRQPacket2(t *testing.T) {
	var opCode = [] byte{0, 2}

	packet := CreateRRQWRQPacket(false, "test.txt", nil)

	if !BlockNumberChecker(packet.Opcode, opCode) {
		t.Errorf("Got = %v; want %v", packet.Opcode, opCode)
	}
}

func TestCreateRRQWRQPacket3(t *testing.T) {
	var options = make(map[string]string)
	const option = "test"
	options["option"] = option

	packet := CreateRRQWRQPacket(false, "test.txt", options)

	if packet.Options["option"] != "test" {
		t.Errorf("Got = %v; want %s", packet.Options["test"], option)
	}
}

func TestCreateDataPacket(t *testing.T) {
	packet := CreateDataPacket([]byte{}, []byte{})

	if !BlockNumberChecker(packet.Opcode, []byte{0, 3}) {
		t.Errorf("Got = %v; want [0 3]", packet.Opcode)
	}

	if packet.Data == nil {
		t.Error("Got nil data array")
	}
}

func TestCreateACKPacket(t *testing.T) {
	packet := CreateACKPacket()

	if !BlockNumberChecker(packet.Opcode, []byte{0, 4}) {
		t.Errorf("Got = %v; want [0 4]", packet.Opcode)
	}
}
