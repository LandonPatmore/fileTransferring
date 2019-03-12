package main

import (
	"fileTransferring/shared"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strings"
	"tideWatchAPI/utils"
)

// TODO: Need to timeout users that authenticated after a certain amount of time, even if we did not get the full file

var connectedClients = make(map[string]*ConnectedClient)
var availableOptions = make(map[string]string)

type ConnectedClient struct {
	FileName            string
	IPv6                bool
	SlidingWindow       bool
	LastSeenBlockNumber [] byte // this is only for sliding window
	DropPackets         bool
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
	var client *ConnectedClient
	if checkAuthenticatedClient(addr.String()) { // know which client we are working with for options
		client = connectedClients[addr.String()]
	}
	data = data[:bytesReceived]
	utils.ErrorCheck(err)
	ack := shared.CreateACKPacket()

	switch data[1] {
	case 2:
		fmt.Println("WRQ packet has been received...")
		w, _ := shared.ReadRRQWRQPacket(data)

		var supportedOptions map[string]string
		var v6, sw, dp bool

		if w.Options != nil {
			ack.IsOACK = true
			ack.Opcode = [] byte{0, 6}
			supportedOptions, v6, sw, dp = parseOptions(w.Options)
			ack.Options = supportedOptions
		}
		if !ack.IsOACK {
			if strings.ToLower(w.Mode) != "octet" {
				sendPacketToClient(conn, addr, createErrorPacket(shared.Error0, "This server only supports octet mode, not: "+w.Mode))
				return
			}
		}
		addToAuthenticatedClients(addr.String(), w.Filename, v6, sw, dp)
		errorPacket, hasError := checkFileExists(getFileNameForAddress(addr.String()))
		if hasError {
			sendPacketToClient(conn, addr, errorPacket)
			removeAuthenticatedClient(addr.String())
			return
		} else {
			ack.BlockNumber = []byte{0, 0}
		}
	case 3:
		if client != nil {
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
		sendPacketToClient(conn, addr, createErrorPacket(shared.Error0, fmt.Sprintf("Server only supports Opcodes of 2,3, 5, and 6...not: %d", data[1])))
	}

	if client != nil {
		if client.DropPackets {
			if rand.Float64() > 0.01 {
				sendPacketToClient(conn, addr, ack.ByteArray())
			} else {
				fmt.Println("Packet dropped...")
			}
		} else {
			sendPacketToClient(conn, addr, ack.ByteArray())
		}
	} else {
		sendPacketToClient(conn, addr, ack.ByteArray())
	}
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

func addToAuthenticatedClients(addr string, fileName string, v6 bool, sw bool, dp bool) {
	hasBeenAdded := checkAuthenticatedClient(addr)
	if !hasBeenAdded {
		connectedClients[addr] = &ConnectedClient{FileName: fileName, IPv6: v6, SlidingWindow: sw, DropPackets: dp}
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
	return connectedClients[addressToGet].FileName
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

func parseOptions(oackPacketOptions map[string]string) (map[string]string, bool, bool, bool) {
	var supportedOptions = make(map[string]string)

	for k, v := range availableOptions {
		if oackPacketOptions[k] == v {
			supportedOptions[k] = v
		}
	}

	v6 := oackPacketOptions["packetMode"] == "ipv6"
	sw := oackPacketOptions["sendMode"] == "sw"
	dp := oackPacketOptions["simulation"] == "dp"

	return supportedOptions, v6, sw, dp
}
