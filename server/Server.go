package main

import (
	"fileTransferring/shared"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"
	"tideWatchAPI/utils"
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

func readPacket(conn *net.UDPConn) {
	data := make([]byte, 516)

	amountOfBytes, addr, err := conn.ReadFromUDP(data)
	utils.ErrorCheck(err)
	data = data[:amountOfBytes]
	ack := shared.CreateACKPacket()

	switch data[1] { // check the opcode of the packet
	case 2:
		fmt.Println("WRQ packet has been received...")
		w, _ := shared.ReadRRQWRQPacket(data)

		filename = w.Filename

		if !ack.IsOACK {
			if strings.ToLower(w.Mode) != "octet" {
				sendPacketToClient(conn, addr, createErrorPacket(shared.Error0, "This server only supports octet mode, not: "+w.Mode))
				return
			}
		}

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

func sendPacketToClient(conn *net.UDPConn, addr *net.UDPAddr, data [] byte) {
	_, _ = conn.WriteToUDP(data, addr)
}

func checkFileExists(fileName string) (ePacket [] byte, hasError bool) {
	_, err := os.Stat(fileName)

	if os.IsNotExist(err) {
		return nil, false
	}

	return createErrorPacket(shared.Error6, shared.Error6Message), true
}

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

func checkEndOfTransfer(data [] byte) {
	if len(data) < 512 { // although the packet is 516, 512 is the max for the data portion...anything smaller and we know it is the end of the file
		fmt.Println("File has been fully transferred")
	}
}

func createErrorPacket(errorCode [] byte, errorMessage string) [] byte {
	ePacket := shared.CreateErrorPacket(errorCode, errorMessage)
	return ePacket.ByteArray()
}

func displayExternalIP() {
	resp, err := http.Get("http://myexternalip.com/raw")

	defer resp.Body.Close()

	utils.ErrorCheck(err)
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	utils.ErrorCheck(err)
	bodyString := string(bodyBytes)
	fmt.Println("External IP: " + bodyString)
}
