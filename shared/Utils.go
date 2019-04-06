package shared

import (
	"fmt"
	"log"
	"os"
	"strings"
)

// Checks if there are any errors panics if there are
func ErrorValidation(err error) {
	if err != nil {
		log.Panic(err)
	}
}

func BlockNumberChecker(block1 [] byte, block2 [] byte) bool {
	if len(block1) == len(block2) {
		for index, v := range block1 {
			if v != block2[index] {
				return false
			}
		}

		return true
	}

	return false
}

// Interprets command line arguments for the program
func GetCMDArgs(args [] string, isClient bool) (bool, bool, bool) {
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
				case "--dp":
					builder.WriteString(" Drop Packets Simulation |")
					dropPackets = true
					break
				case "-h":
					fallthrough
				case "--help":
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
				case "-h":
					fallthrough
				case "--help":
					showHelp(isClient)
					os.Exit(0)
				default:
					fmt.Println(`The argument: "` + arg + `" is not valid`)
					os.Exit(1)
				}
			}
		}

		strippedVersion := builder.String()[:len(builder.String())-1]
		fmt.Println(strippedVersion)
	} else {
		if isClient {
			fmt.Println("Default Options: IPv4 Mode | Sequential Acks Mode | No Drop Packets Simulation")
		} else {
			fmt.Println("Default Option: IPv4 Mode")
		}
	}

	return ipv6, slidingWindow, dropPackets
}

func showHelp(isClient bool) {
	if isClient {
		fmt.Println("usage: ./Client [<options>]")
		fmt.Println()
		fmt.Printf("\t--ipv6\t\t %s\n", "Specify to start client in IPv4 or IPv6 mode")
		fmt.Printf("\t--sw\t\t %s\n", "Specify use of TCP-style sliding windows rather than the sequential acks used in TFTP")
		fmt.Printf("\t--dp\t\t %s\n\n", "Pretend to drop 1% of packets")
	} else {
		fmt.Println("usage: ./Server [<options>]")
		fmt.Println()
		fmt.Printf("\t--ipv6\t\t %s\n", "Specify to start server in IPv4 or IPv6 mode")
	}
}
