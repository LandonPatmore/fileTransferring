package main

import (
	"fileTransferring/shared"
	"fmt"
	"net"
	"os"
	"strings"
	"tideWatchAPI/utils"
)

// TODO: Need to timeout users that authenticated after a certain amount of time, even if we did not get the full file

var connectedClients = make(map[string]string)
var availableOptions = make(map[string]string)

type ConnectedClient struct {
	FileName               string
	LastTimePacketReceived int64
}

func main() {
	initializeOptions()

	ServerAddr, err := net.ResolveUDPAddr("udp", shared.PORT)
	shared.ErrorValidation(err)
	conn, err := net.ListenUDP("udp", ServerAddr)
	shared.ErrorValidation(err)

	defer conn.Close()

	fmt.Println("Server started...")

	for {
		readPacket(conn)
	}
}

func readPacket(conn *net.UDPConn) {
	data := make([]byte, 516)
	bytesReceived, addr, err := conn.ReadFromUDP(data)
	data = data[:bytesReceived]
	utils.ErrorCheck(err)
	t := data[1]

	ack := shared.CreateACKPacket()

	switch t {
	case 2:
		fmt.Println("WRQ packet has been received...")
		w, _ := shared.ReadRRQWRQPacket(data)
		if w.Options != nil {
			ack.IsOACK = true
			ack.Opcode = [] byte{0, 6}
			supportedOptions := parseOptions(w.Options)
			ack.Options = supportedOptions
			addToAuthenticatedClients(addr.String(), w.Filename)
			break
		}
		if strings.ToLower(w.Mode) != "octet" {
			sendPacketToClient(conn, addr, createErrorPacket(shared.Error0, "This server only supports octet mode, not: "+w.Mode))
			return
		} else {
			addToAuthenticatedClients(addr.String(), w.Filename)
			errorPacket, hasError := checkFileExists(getFileNameForAddress(addr.String()))
			if hasError {
				sendPacketToClient(conn, addr, errorPacket)
				fmt.Println("WRQ place removed")
				removeAuthenticatedClient(addr.String())
				return
			} else {
				ack.BlockNumber = []byte{0, 0}
			}
		}
	case 3:
		if checkAuthenticatedClient(addr.String()) {
			d, _ := shared.ReadDataPacket(data)
			errorPacket, hasError := writeToFile(getFileNameForAddress(addr.String()), d.Data)
			if hasError {
				sendPacketToClient(conn, addr, errorPacket)
				return
			} else {
				checkEndOfTransfer(d.Data, addr.String())
				ack.BlockNumber = d.BlockNumber
			}
		} else {
			sendPacketToClient(conn, addr, createErrorPacket(shared.Error0, fmt.Sprintf("Client has not sent a WRQ Packet, permission denied")))
		}
	default:
		sendPacketToClient(conn, addr, createErrorPacket(shared.Error0, fmt.Sprintf("Server only supports Opcodes of 2,3, and 5...not: %d", t)))
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

func addToAuthenticatedClients(addr string, fileName string) {
	hasBeenAdded := checkAuthenticatedClient(addr)
	if !hasBeenAdded {
		connectedClients[addr] = fileName
	}
}

func checkAuthenticatedClient(toAdd string) bool {
	_, isAuthenticated := connectedClients[toAdd]
	return isAuthenticated
}

func removeAuthenticatedClient(toRemove string) {
	delete(connectedClients, toRemove)
}

func getFileNameForAddress(addressToGet string) string {
	return connectedClients[addressToGet]
}

func checkEndOfTransfer(data [] byte, addressToRemove string) {
	if len(data) < 512 {
		fmt.Println("File has been fully transferred")
		removeAuthenticatedClient(addressToRemove)
	}
}

func createErrorPacket(errorCode [] byte, errorMessage string) [] byte {
	ePacket := shared.CreateErrorPacket(errorCode, errorMessage)
	return ePacket.ByteArray()
}

func initializeOptions() {
	availableOptions["packetMode"] = "ipv6"
	availableOptions["sendMode"] = "sw"
	availableOptions["simulation"] = "dp"
}

func parseOptions(oackPacketOptions map[string]string) map[string]string {
	var supportedOptions = make(map[string]string)

	for k, v := range availableOptions {
		if oackPacketOptions[k] == v {
			supportedOptions[k] = v
		}
	}

	return supportedOptions
}
