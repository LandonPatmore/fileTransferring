package main

import (
	"fileTransferring/shared"
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"math"
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

func main() {
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

	fmt.Println("Buffering file...")
	fileBytes, err := ioutil.ReadFile(filePath) // b has type []byte
	shared.ErrorValidation(err)
	fmt.Println("File Buffered!")

	file, fileError := os.Open(filePath)
	shared.ErrorValidation(fileError)

	fi, err := file.Stat()
	shared.ErrorValidation(err)
	fileSize = fi.Size()

	totalPacketsToSend = determineAmountOfPacketsToSend(fileSize)

	defer file.Close()

	sendWRQPacket(conn, filepath.Base(file.Name()), nil) // TODO: Change this to work with new way for options

	sendFile(conn, fileBytes)
}

// Reads a file and sends it to the server
func sendFile(conn *net.UDPConn, fileBytes [] byte) {
	var currentPacket int

	if len(fileBytes) >= 512 {
		sendDataPacket(conn, fileBytes[:512], &currentPacket)
		sendFile(conn, fileBytes[512:])
	} else { // at end of file
		sendDataPacket(conn, fileBytes, &currentPacket)
	}
	fmt.Println("\nDone reading and sending file...")
}

// Creates and sends a WRQ packet
func sendWRQPacket(conn *net.UDPConn, fileName string, options map[string]string) {
	wPacket := shared.CreateRRQWRQPacket(false, fileName, options)
	send(conn, wPacket.ByteArray(), []byte{0, 0})
}

// Creates and sends a data packet
func sendDataPacket(conn *net.UDPConn, data [] byte, currentPacket *int) {
	dataPacket := shared.CreateDataPacket(createBlockNumber(currentPacket), data)
	send(conn, dataPacket.ByteArray(), dataPacket.BlockNumber)

	totalBytesSent += int64(len(dataPacket.Data))
	totalPacketsSent++
	displayProgress()
}

// Receives a packet and does something with it based on the opcode
func receivePacket(data [] byte, blockNumber [] byte) error {
	opcode := data[1]

	switch opcode {
	case 4: // ack packet
		ack, _ := shared.ReadACKPacket(data)
		if shared.BlockNumberChecker(ack.BlockNumber, blockNumber) {
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
		return errors.New(fmt.Sprintf("Client can only read Opcodes of 4, 5, and 6...not: %d", opcode))
	}
}

// Generates a block number based on the current packet number
func createBlockNumber(currentPacketNumber *int) [] byte {
	*currentPacketNumber++
	if *currentPacketNumber < 256 {
		return [] byte{0, byte(*currentPacketNumber)}
	}
	leadingByte := math.Floor(float64(*currentPacketNumber / 256))
	return [] byte{byte(leadingByte), byte(*currentPacketNumber % 256)}
}

// Sends the packet to the server
func send(conn *net.UDPConn, data []byte, blockNumber [] byte) {
	for i := 0; i < 10; i++ {
		_, _ = conn.Write(data)
		receivedData, err := handleReadTimeout(conn)
		if err == nil {
			err := receivePacket(receivedData, blockNumber)
			shared.ErrorValidation(err)
			return
		} else {
			packetsLost++
			displayProgress()
			time.Sleep(time.Millisecond * 500)
		}
	}

	shared.ErrorValidation(errors.New("Total retries exhausted...exiting"))
}

// Handles the read timeout of the server sending back an ACK
func handleReadTimeout(conn *net.UDPConn) ([] byte, error) {
	_ = conn.SetReadDeadline(time.Now().Add(time.Millisecond * 500))

	receivedData := make([]byte, 516)
	bytesReceived, _, timedOut := conn.ReadFromUDP(receivedData)

	return receivedData[:bytesReceived], timedOut
}

// Displays a progress bar that updates with the total data and total packets sent
func displayProgress() {
	var totalDataSent = math.Ceil(float64(totalBytesSent) / float64(fileSize) * 100)

	fmt.Print("\r")
	fmt.Printf("Progress: (%d%% | Packets Lost: %d | %d/%d packets sent) ", int(totalDataSent), packetsLost, totalPacketsSent, totalPacketsToSend)
}

// Returns amount of packets that must be sent for file to be transferred
// completely
func determineAmountOfPacketsToSend(fileSize int64) int { // yes the last packet will be smaller, but we don't care
	return int(math.Ceil(float64(fileSize) / 512))
}
