package shared

import (
	"fmt"
	"log"
	"os"
)

// Checks if there are any errors panics if there are
func ErrorValidation(err error) {
	if err != nil {
		log.Panic(err)
	}
}

func InterpretCommandLineArguments(args [] string) (bool, bool, bool) {
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

func showHelp()  {
	fmt.Println("usage: ./fileTransferring [<options>]")
	fmt.Println()
	fmt.Printf("\t--ipv6\t\t %s\n", "Specify if packets are IPv6 UDP datagrams instead of IPv4 packets")
	fmt.Printf("\t--sw\t\t %s\n", "Specify use of TCP-style sliding windows rather than the sequential acks used in TFTP")
	fmt.Printf("\t--dp\t\t %s\n\n" ,"Pretend to drop 1% of packets")
}