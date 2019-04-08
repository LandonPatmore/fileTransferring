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

// Checks block numbers for length and value equality
func BlockNumberChecker(receivedBlock [] byte, expectedBlock [] byte) bool {
	if len(receivedBlock) == len(expectedBlock) {
		for index, v := range receivedBlock {
			if v != expectedBlock[index] {
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
					builder.WriteString(" Drop Packets |")
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

		strippedVersion := builder.String()[:len(builder.String())-1] // just to cleans up output, removing last |
		fmt.Println(strippedVersion)
	} else {
		if isClient {
			fmt.Println("Default Options: IPv4 | Sequential Mode | Normal Operation")
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
		fmt.Printf("\t--ipv6\t\t %s\n", "IPv6 mode")
		fmt.Printf("\t--sw\t\t %s\n", "TCP-style sliding window sending algorithm")
		fmt.Printf("\t--dp\t\t %s\n\n", "Drop 1% of packets")
	} else {
		fmt.Println("usage: ./Server [<options>]")
		fmt.Println()
		fmt.Printf("\t--ipv6\t\t %s\n", "IPv6 mode")
	}
}
