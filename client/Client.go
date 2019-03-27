package main

import (
	"bufio"
	"fileTransferring/shared"
	"fmt"
	"github.com/pkg/errors"
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

var v6 bool
var sw bool

func main() {
	v6, sw, _ = shared.InterpretCommandLineArguments(os.Args, true)
	options := createOptions(sw)

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

	totalPacketsToSend = determineAmountOfPacketsToSend(fileSize)

	defer file.Close()

	sendWRQPacket(conn, filepath.Base(file.Name()), options) // TODO: Change this to work with new way for options

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

func sendWRQPacket(conn *net.UDPConn, fileName string, options map[string]string) {
	wPacket := shared.CreateRRQWRQPacket(false, fileName, options)

	data := wPacket.ByteArray()

	sendPacket(conn, data, []byte{0, 0})
}

func sendDataPacket(conn *net.UDPConn, data *[] byte, currentPacket *int) {
	dataPacket := shared.CreateDataPacket()
	dataPacket.BlockNumber = createBlockNumber(currentPacket)
	dataPacket.Data = *data

	d := dataPacket.ByteArray()

	totalBytesSent += int64(len(dataPacket.Data))
	totalPacketsSent++
	displayProgress()

	sendPacket(conn, d, dataPacket.BlockNumber)
}

func receivePacket(data [] byte, blockNumber [] byte) error {

	t := data[1]

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
	case 6:
		oack, _ := shared.ReadOACKPacket(data)
		fmt.Println(oack)
		return nil
	default:
		return errors.New(fmt.Sprintf("Client can only read Opcodes of 4, 5, and 6...not: %d", t))
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
		_, _ = conn.Write(data)
		receivedData, err := handleReadTimeout(conn)
		if err == nil {
			err := receivePacket(receivedData, blockNumber)
			shared.ErrorValidation(err)
			return
		} else {
			packetsLost++
			displayProgress()
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

func displayProgress() {
	var totalDataSent = math.Floor(float64(totalBytesSent) / float64(fileSize) * 100)

	fmt.Print("\r")
	fmt.Printf("Progress: (%d%% | Packets Lost: %d | %d/%d packets sent) ", int(totalDataSent), packetsLost, totalPacketsSent, totalPacketsToSend)
}

func determineAmountOfPacketsToSend(fileSize int64) int { // yes the last packet will be smaller, but we don't care
	return int(math.Ceil(float64(fileSize) / 512))
}

func createOptions(isSW bool) map[string]string {
	options := make(map[string]string)
	if isSW {
		options["packetMode"] = "sw"
	}

	return options
}
