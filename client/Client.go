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
var v6 bool
var sw bool
var dp bool

func main() {
	v6, sw, dp = shared.InterpretCommandLineArguments(os.Args)
	fmt.Printf("IPv6 Packets: %t\nSliding Window: %t\nDrop 1%% of Packets: %t\n", v6, sw, dp)

	var serverAddress string

	fmt.Print("Server address: ")
	_, _ = fmt.Scanf("%s", &serverAddress)

	remoteAddr, err := net.ResolveUDPAddr("udp", serverAddress+shared.PORT)
	shared.ErrorValidation(err)

	conn, connError := net.DialUDP("udp", nil, remoteAddr)
	shared.ErrorValidation(connError)

	var filePath string
	fmt.Print("Enter full file path: ")
	_, _ = fmt.Scanf("%s", &filePath)

	file, fileError := os.Open(filePath)
	shared.ErrorValidation(fileError)

	fi, _ := file.Stat()
	fileSize = fi.Size()

	if dp {
		determineAmountOfPacketsToSend()
		determinePacketsToDrop()
	}

	defer file.Close()

	sendWRQPacket(conn, filepath.Base(file.Name()))

	readFile(conn, file)
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

func sendWRQPacket(conn *net.UDPConn, fileName string) {
	wPacket := shared.CreateRRQWRQPacket(false)
	wPacket.Filename = fileName

	data := shared.CreateRRQWRQPacketByteArray(wPacket)

	_, _ = conn.Write(data)
	handleTimeout(conn, data, [] byte{0, 0})
}

func sendDataPacket(conn *net.UDPConn, data *[] byte, currentPacket *int) {
	dataPacket := shared.CreateDataPacket()
	dataPacket.BlockNumber = createBlockNumber(currentPacket)
	dataPacket.Data = *data

	d := shared.CreateDataPacketByteArray(dataPacket)

	totalBytesSent += int64(len(dataPacket.Data))
	totalPacketsSent++
	displayProgressBar()

	if dp {
		var foundPacketToDrop bool
		for _, packetToDrop := range packetsToDrop {
			if totalPacketsSent == packetToDrop {
				foundPacketToDrop = true
				break
			}
		}

		if foundPacketToDrop {
			handleTimeout(conn, d, dataPacket.BlockNumber)
		} else {
			_, _ = conn.Write(d)
			handleTimeout(conn, d, dataPacket.BlockNumber)
		}
	} else {
		_, _ = conn.Write(d)
		handleTimeout(conn, d, dataPacket.BlockNumber)
	}
}

func receivePacket(data [] byte, blockNumber [] byte) error {

	t := shared.DeterminePacketType(data)

	switch t {
	case 4:
		ack, _ := shared.ReadACKPacket(data)
		if shared.CheckByteArrayEquality(ack.BlockNumber, blockNumber) {
			return nil
		}
		return errors.New("Block numbers do not match...")
	case 5:
		e, _ := shared.ReadErrorPacket(data)
		return errors.New(fmt.Sprintf("Error code: %d\nError Message: %s", e.ErrorCode[1], e.ErrorMessage))
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

func handleTimeout(conn *net.UDPConn, data []byte, blockNumber [] byte) {
	_ = conn.SetReadDeadline(time.Now().Add(time.Second * 2))

	receivedData := make([]byte, 516)
	bytesReceived, _, timedOut := conn.ReadFromUDP(receivedData)

	if timedOut != nil {
		packetsLost++
		displayProgressBar()
		_, _ = conn.Write(data)
		handleTimeout(conn, data, blockNumber)
	} else {
		err := receivePacket(receivedData[:bytesReceived], blockNumber)
		shared.ErrorValidation(err)
	}
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
