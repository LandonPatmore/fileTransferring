package main

import (
	"bufio"
	"fileTransferring/shared"
	"fmt"
	"github.com/pkg/errors"
	"math"
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"time"
)

var fileSize int64
var totalBytesSent int64
var totalPacketsSent int
var totalPacketsToSend int
var packetsLost int
var packetsToDrop [] int

func main() {
	options := InterpretCommandLineArguments(os.Args)

	var serverAddress string = "127.0.0.1"

	//fmt.Print("Server address: ")
	//_, _ = fmt.Scanf("%s", &serverAddress)

	remoteAddr, err := net.ResolveUDPAddr("udp", serverAddress+shared.PORT)
	shared.ErrorValidation(err)

	conn, connError := net.DialUDP("udp", nil, remoteAddr)
	shared.ErrorValidation(connError)

	var filePath string = "/Users/landon/Desktop/bigfile.txt"
	//fmt.Print("Enter full file path: ")
	//_, _ = fmt.Scanf("%s", &filePath)

	file, fileError := os.Open(filePath)
	shared.ErrorValidation(fileError)

	fi, _ := file.Stat()
	fileSize = fi.Size()

	//determineAmountOfPacketsToSend()
	//
	//if dp {
	//	determinePacketsToDrop()
	//}

	defer file.Close()

	sendWRQPacket(conn, filepath.Base(file.Name()), options)

	//readFile(conn, file)
}

func readFile(conn *net.UDPConn, file *os.File) {
	scanner := bufio.NewScanner(file)

	var message = make([]byte, 0, 512)
	var currentPacket int
	for scanner.Scan() {
		for _, character := range scanner.Bytes() {
			addToArray(conn, &message, character, &currentPacket)
		}
		addToArray(conn, &message, '\n', &currentPacket)
	}
	message = message[:len(message)-1] // to remove the last \n
	sendDataPacket(conn, &message, &currentPacket)
	fmt.Println("\nDone reading and sending file...")
}

func addToArray(conn *net.UDPConn, message *[] byte, nextByteToAppend byte, currentPacket *int) {
	if checkMessageLength(message) {
		sendDataPacket(conn, message, currentPacket)
		*message = make([]byte, 0, 512)
	}
	*message = append(*message, nextByteToAppend)
}

func checkMessageLength(message *[] byte) bool {
	if len(*message) == 512 {
		return true
	}
	return false
}

func sendWRQPacket(conn *net.UDPConn, fileName string, options map[string]string) {
	wPacket := shared.CreateRRQWRQPacket(false, fileName, options)

	data := shared.CreateRRQWRQPacketByteArray(wPacket)

	//sendPacket(conn, data, []byte{0, 0})
	conn.Write(data)
	receivedData := make([]byte, 516)
	bytesReceived, _, _ := conn.ReadFromUDP(receivedData)
	_ = receivePacket(receivedData[:bytesReceived], [] byte{0, 0})
}

func sendDataPacket(conn *net.UDPConn, data *[] byte, currentPacket *int) {
	dataPacket := shared.CreateDataPacket()
	dataPacket.BlockNumber = createBlockNumber(currentPacket)
	dataPacket.Data = *data

	d := shared.CreateDataPacketByteArray(dataPacket)

	totalBytesSent += int64(len(dataPacket.Data))
	totalPacketsSent++
	displayProgressBar()

	sendPacket(conn, d, dataPacket.BlockNumber)
}

func receivePacket(data [] byte, blockNumber [] byte) error {

	t := data[1]

	switch t {
	case 4:
		ack, _ := shared.ReadACKPacket(data)
		if ack.Options != nil {
			// TODO: Do something with the options returned by the server
			return nil
		} else if shared.CheckByteArrayEquality(ack.BlockNumber, blockNumber) {
			return nil
		}
		return errors.New("Block numbers do not match...")
	case 5:
		e, _ := shared.ReadErrorPacket(data)
		return errors.New(fmt.Sprintf("Error code: %d\nError Message: %s", e.ErrorCode[1], e.ErrorMessage))
	case 6:
		oack, _ := shared.ReadOACKPacket(data)
		fmt.Println(oack)
		return nil
	default:
		return errors.New(fmt.Sprintf("Client can only read Opcodes of 4 and 5...not: %d", t))
	}
}

func createBlockNumber(currentPacketNumber *int) [] byte {
	*currentPacketNumber++
	if *currentPacketNumber < 256 {
		return [] byte{0, byte(*currentPacketNumber)}
	}
	leadingByte := math.Floor(float64(*currentPacketNumber / 256))
	return [] byte{byte(leadingByte), byte(*currentPacketNumber % 256)}
}

func sendPacket(conn *net.UDPConn, data []byte, blockNumber [] byte) {
	for i := 0; i < 10; i++ {
		//if !shouldDropPacket() {
		_, _ = conn.Write(data)
		//}
		receivedData, err := handleReadTimeout(conn)
		if err == nil {
			err := receivePacket(receivedData, blockNumber)
			shared.ErrorValidation(err)
			return
		} else {
			packetsLost++
			displayProgressBar()
		}
	}

	shared.ErrorValidation(errors.New("Total retries exhausted...exiting"))
}

func handleReadTimeout(conn *net.UDPConn) ([] byte, error) {
	_ = conn.SetReadDeadline(time.Now().Add(time.Second * 2))

	receivedData := make([]byte, 516)
	bytesReceived, _, timedOut := conn.ReadFromUDP(receivedData)

	return receivedData[:bytesReceived], timedOut
}

func displayProgressBar() {
	var totalDataSent = math.Floor(float64(totalBytesSent) / float64(fileSize) * 100)

	fmt.Print("\r")
	fmt.Printf("Progress: (%d%% | Packets Lost: %d | %d/%d packets sent) ", int(totalDataSent), packetsLost, totalPacketsSent, totalPacketsToSend)
}

func determineAmountOfPacketsToSend() { // yes the last packet will be smaller, but we don't care
	totalPacketsToSend = int(math.Ceil(float64(fileSize) / 512))
}

func determinePacketsToDrop() {
	var onePercentOfPacketsToDrop = math.Ceil(float64(totalPacketsToSend) * 0.01)

	// randomly choose which packets to drop
	for {
		if len(packetsToDrop) != int(onePercentOfPacketsToDrop) {
			packetsToDrop = append(packetsToDrop, rand.Intn(totalPacketsToSend))
		} else {
			break
		}
	}
}

//func shouldDropPacket() bool {
//	if dp {
//		for _, packetToDrop := range packetsToDrop {
//			if totalPacketsSent == packetToDrop {
//				return true
//			}
//		}
//	}
//
//	return false
//}

// Interprets command line arguments for the program
func InterpretCommandLineArguments(args [] string) map[string]string {
	options := make(map[string]string)

	if len(args[1:]) > 0 {
		fmt.Print("Options Specified: ")

		for _, arg := range args[1:] {
			switch arg {
			case "--ipv6":
				options["packetMode"] = "ipv6"
				fmt.Print(" IPv6 |")
				break
			case "--sw":
				options["sendMode"] = "sw"
				fmt.Print(" Sliding Window Mode |")
				break
			case "--dp":
				options["simulation"] = "dp"
				fmt.Print(" Drop Packets Simulation |")
				break
			case "-h":
				showHelp()
				os.Exit(0)
			default:
				fmt.Print(" " + arg + " (not supported) |")
			}
		}

		fmt.Println()
		return options
	} else {
		fmt.Println("Default Options: IPv4 | Sequential Acks Mode | No Simulation")
		return nil
	}
}

func showHelp() {
	fmt.Println("usage: ./fileTransferring [<options>]")
	fmt.Println()
	fmt.Printf("\t--ipv6\t\t %s\n", "Specify if packets are IPv6 UDP datagrams instead of IPv4 packets")
	fmt.Printf("\t--sw\t\t %s\n", "Specify use of TCP-style sliding windows rather than the sequential acks used in TFTP")
	fmt.Printf("\t--dp\t\t %s\n\n", "Pretend to drop 1% of packets")
}
