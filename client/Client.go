package main

import (
	"bufio"
	"fileTransferring/shared"
	"fmt"
	"math"
	"net"
	"os"
)

func main() {
	//var serverAddress string
	//
	//fmt.Print("Server address: ")
	//_, _ = fmt.Scanf("%s", &serverAddress)
	//
	//conn, connError := net.Dial("udp", serverAddress+":8274")
	//shared.ErrorValidation(connError)

	//var filePath = "/Users/landon/Desktop/WarThunder-Helper/index.html"
	////fmt.Print("Enter full file path: ")
	////_, _ = fmt.Scanf("%s", &filePath)
	//
	//file, fileError := os.Open(filePath)
	//shared.ErrorValidation(fileError)
	//
	//readFile(nil, file)
	sendWRQPacket(nil, "Test.txt")
}

func readFile(conn net.Conn, file *os.File) {
	scanner := bufio.NewScanner(file)

	var message = make([]byte, 0, 512)
	var currentPacket int
	for scanner.Scan() {
		for _, character := range scanner.Bytes() {
			addToArray(conn, &message, character, &currentPacket)
		}
		addToArray(conn, &message, '\n', &currentPacket)
	}
	sendDataPacket(conn, &message, &currentPacket)
}

func addToArray(conn net.Conn, message *[] byte, nextByteToAppend byte, currentPacket *int) {
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

func sendWRQPacket(conn net.Conn, fileName string) {
	wPacket := shared.CreateRRQWRQPacket(false)
	wPacket.Filename = fileName

	packet := shared.CreateRRQWRQPacketByteArray(wPacket)
	
	fmt.Println(packet)
	fmt.Println()
}

func sendDataPacket(conn net.Conn, data *[] byte, currentPacket *int) {
	dataPacket := shared.CreateDataPacket()
	dataPacket.BlockNumber = createBlockNumber(currentPacket)
	dataPacket.Data = *data

	packet := shared.CreateDataPacketByteArray(dataPacket)

	fmt.Println(packet)
	fmt.Println()
}

func createBlockNumber(currentPacketNumber *int) [] byte {
	*currentPacketNumber++
	if *currentPacketNumber < 256 {
		return [] byte{0, byte(*currentPacketNumber)}
	}
	leadingByte := math.Floor(float64(*currentPacketNumber / 256))
	return [] byte{byte(leadingByte), byte(*currentPacketNumber % 256)}

}
