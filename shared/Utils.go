package shared

import (
	"fmt"
	"os"
	"strings"
)

// Checks if there are any errors panics if there are
func ErrorValidation(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func CheckByteArrayEquality(byte1 [] byte, byte2 [] byte) bool {
	return len(byte1) == len(byte2)
}

// Interprets command line arguments for the program
func InterpretCommandLineArguments(args [] string, isClient bool) (bool, bool, bool) {
	var ipv6 bool
	var slidingWindow bool
	var dropPackets bool
	var builder = strings.Builder{}

	if len(args[1:]) > 0 {
		builder.WriteString("Options specified: ")

		for _, arg := range args[1:] {
			if isClient {
				switch arg {
				case "--ipv6":
					builder.WriteString(" IPv6 Mode |")
					ipv6 = true
					break
				case "--sw":
					builder.WriteString(" Sliding Window Mode |")
					slidingWindow = true
					break
				case "-h":
					showHelp(isClient)
					os.Exit(0)
				default:
					fmt.Println(`The argument: "` + arg + `" is not valid`)
					os.Exit(1)
				}
			} else {
				switch arg {
				case "--ipv6":
					builder.WriteString(" IPv6 Mode |")
					ipv6 = true
					break
				case "--dp":
					builder.WriteString(" Drop Packets Simulation |")
					dropPackets = true
					break
				case "-h":
					showHelp(isClient)
					os.Exit(0)
				default:
					fmt.Println(`The argument: "` + arg + `" is not valid`)
					os.Exit(1)
				}
			}
		}

		fmt.Println(builder.String())
	} else {
		if isClient {
			fmt.Println("Default Options: IPv4 Mode | Sequential Acks Mode")
		} else {
			fmt.Println("Default Options: IPv4 Mode | No Drop Packets Simulation")
		}
	}

	return ipv6, slidingWindow, dropPackets
}

func showHelp(isClient bool) {
	if isClient {
		fmt.Println("usage: ./client [<options>]")
		fmt.Println()
		fmt.Printf("\t--ipv6\t\t %s\n", "Specify to start client in IPv4 or IPv6 mode")
		fmt.Printf("\t--sw\t\t %s\n", "Specify use of TCP-style sliding windows rather than the sequential acks used in TFTP")
		fmt.Printf("\t--dp\t\t %s\n\n", "Pretend to drop 1% of packets")
	} else {
		fmt.Println("usage: ./server [<options>]")
		fmt.Println()
		fmt.Printf("\t--ipv6\t\t %s\n", "Specify to start server in IPv4 or IPv6 mode")
	}
}
