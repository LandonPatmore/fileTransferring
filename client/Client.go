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
	//
	//var filePath = "/Users/landon/Desktop/WarThunder-Helper/index.html"
	////fmt.Print("Enter full file path: ")
	////_, _ = fmt.Scanf("%s", &filePath)
	//
	//file, fileError := os.Open(filePath)
	//shared.ErrorValidation(fileError)
	//
	//readFile(conn, file)

	var currentPacket int
	for i := 0; i < 512; i++ {
		fmt.Println(createBlockNumber(&currentPacket))
	}
}

func readFile(conn net.Conn, file *os.File) {
	scanner := bufio.NewScanner(file)

	var message = make([]byte, 0, 512)
	var currentPacket int
	for scanner.Scan() {
		for _, character := range scanner.Bytes() {
			addToArray(conn, &message, character, createBlockNumber(&currentPacket))
		}
		addToArray(conn, &message, '\n', createBlockNumber(&currentPacket))
	}
	sendDataPacket(conn, &message, createBlockNumber(&currentPacket))
}

func addToArray(conn net.Conn, message *[] byte, nextByteToAppend byte, blockNumber [] byte) {
	if checkMessageLength(message) {
		sendDataPacket(conn, message, blockNumber)
	}
	*message = append(*message, nextByteToAppend)
}

func checkMessageLength(message *[] byte) bool {
	if len(*message) == 512 {
		fmt.Println("Packet can be sent!")
		*message = make([]byte, 0, 512)
		return true
	}

	return false
}

func sendDataPacket(conn net.Conn, data *[] byte, blockNumber [] byte) {
	var dataPacket = shared.CreateDataPacket()
	dataPacket.BlockNumber = blockNumber
	dataPacket.Data = *data
	// Send data
}

func createBlockNumber(currentPacketNumber *int) [] byte {
	*currentPacketNumber++
	if *currentPacketNumber < 256 {
		return [] byte{0, byte(*currentPacketNumber)}
	}
	leadingByte := math.Floor(float64(*currentPacketNumber / 256))
	return [] byte{byte(leadingByte), byte(*currentPacketNumber % 256)}

}
