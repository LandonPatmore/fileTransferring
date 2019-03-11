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
