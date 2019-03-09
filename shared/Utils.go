package shared

import (
	"fmt"
	"log"
	"os"
)

// Checks if there are any errors panics if there are
func ErrorValidation(err error) {
	if err != nil {
		log.Fatal(err)
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

// Interprets command line arguments for the program
func InterpretCommandLineArguments(args [] string) (v6 bool, sw bool, dp bool) {
	var ipv6 bool
	var slidingWindow bool
	var dropPackets bool

	for _, arg := range args[1:] {
		switch arg {
		case "--ipv6":
			ipv6 = true
			break
		case "--sw":
			slidingWindow = true
			break
		case "--dp":
			dropPackets = true
			break
		case "-h":
			showHelp()
			os.Exit(0)
		default:
			fmt.Println(`The argument: "` + arg + `" is not valid`)
			os.Exit(1)
		}
	}

	return ipv6, slidingWindow, dropPackets
}

func showHelp() {
	fmt.Println("usage: ./fileTransferring [<options>]")
	fmt.Println()
	fmt.Printf("\t--ipv6\t\t %s\n", "Specify if packets are IPv6 UDP datagrams instead of IPv4 packets")
	fmt.Printf("\t--sw\t\t %s\n", "Specify use of TCP-style sliding windows rather than the sequential acks used in TFTP")
	fmt.Printf("\t--dp\t\t %s\n\n", "Pretend to drop 1% of packets")
}
