package shared

import (
	"fmt"
	"os"
)

// Checks if there are any errors panics if there are
func ErrorValidation(err error) {
	if err != nil {
		fmt.Println("\n")
		fmt.Println(err)
		os.Exit(-1)
	}
}

func DeterminePacketType(data [] byte) int {
	switch data[1] {
	case 1:
		return 1
	case 2:
		return 2
	case 3:
		return 3
	case 4:
		return 4
	case 5:
		return 5
	default:
		return 0
	}
}

func CheckByteArrayEquality(byte1 [] byte, byte2 [] byte) bool {
	//fmt.Println("ACK: ", byte1, "Block: ", byte2)
	if len(byte1) != len(byte2) {
		return false
	}

	for index, value := range byte1 {
		if value != byte2[index] {
			return false
		}
	}

	return true
}
