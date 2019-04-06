package main

import (
	"fileTransferring/shared"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"
)

var filename string // this server will only handle one connection at a time, so we just set this variable each time a new WRQ packet comes int

func main() {
	ServerAddr, err := net.ResolveUDPAddr("udp", shared.PORT)
	shared.ErrorValidation(err)
	conn, err := net.ListenUDP("udp", ServerAddr)
	shared.ErrorValidation(err)

	defer conn.Close()

	fmt.Println("Server started...")

	displayExternalIP()

	for {
		readPacket(conn)
	}
}

// Reads the incoming packet and performs operations based on the packet received
func readPacket(conn *net.UDPConn) {
	data := make([]byte, 516)

	amountOfBytes, addr, err := conn.ReadFromUDP(data)
	shared.ErrorValidation(err)
	data = data[:amountOfBytes]
	ack := shared.CreateACKPacket()

	switch data[1] { // check the opcode of the packet
	case 2:
		fmt.Println("WRQ packet has been received...")
		w, _ := shared.ReadRRQWRQPacket(data)

		filename = w.Filename

		//if !ack.IsOACK { // TODO: Need to be able to determine if the WRQ has options attached
			if strings.ToLower(w.Mode) != "octet" {
				sendPacketToClient(conn, addr, createErrorPacket(shared.Error0, "This server only supports octet mode, not: "+w.Mode))
				return
			}
		//}

		errorPacket, hasError := checkFileExists(filename)

		if hasError {
			sendPacketToClient(conn, addr, errorPacket)
			return
		} else {
			ack.BlockNumber = []byte{0, 0}
		}
	case 3:
		d, _ := shared.ReadDataPacket(data)
		errorPacket, hasError := writeToFile(filename, d.Data)
		if hasError {
			sendPacketToClient(conn, addr, errorPacket)
			return
		} else {
			checkEndOfTransfer(d.Data)
			ack.BlockNumber = d.BlockNumber
		}
	default:
		sendPacketToClient(conn, addr, createErrorPacket(shared.Error0, fmt.Sprintf("Server only supports Opcodes of 2,3, 5, and 6...not: %d", data[1])))
	}

	sendPacketToClient(conn, addr, ack.ByteArray())
}

// Sends the packet to the client
func sendPacketToClient(conn *net.UDPConn, addr *net.UDPAddr, data [] byte) {
	_, _ = conn.WriteToUDP(data, addr)
}

// Checks if a file exists and returns an error if so
func checkFileExists(fileName string) (ePacket [] byte, hasError bool) {
	_, err := os.Stat(fileName)

	if os.IsNotExist(err) {
		return nil, false
	}

	return createErrorPacket(shared.Error6, shared.Error6Message), true
}

// Writes to a file and returns an error if it cannot write to that specific file
func writeToFile(fileName string, data []byte) (eData [] byte, hasError bool) {
	f, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		return createErrorPacket(shared.Error0, err.Error()), true
	}
	if _, err := f.Write(data); err != nil {
		return createErrorPacket(shared.Error0, err.Error()), true
	}
	if err := f.Close(); err != nil {
		return createErrorPacket(shared.Error0, err.Error()), true
	}

	return nil, false
}

// Checks the end of the file transfer based on the data portion of the packet
func checkEndOfTransfer(data [] byte) {
	if len(data) < 512 { // although the packet is 516, 512 is the max for the data portion...anything smaller and we know it is the end of the file
		fmt.Println("File has been fully transferred")
	}
}

// Helper to create an error packet
func createErrorPacket(errorCode [] byte, errorMessage string) [] byte {
	ePacket := shared.CreateErrorPacket(errorCode, errorMessage)
	return ePacket.ByteArray()
}

// Displays the external IP of the running server
func displayExternalIP() {
	resp, err := http.Get("http://myexternalip.com/raw")

	defer resp.Body.Close()

	shared.ErrorValidation(err)
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	shared.ErrorValidation(err)
	bodyString := string(bodyBytes)
	fmt.Println("External IP: " + bodyString)
}
