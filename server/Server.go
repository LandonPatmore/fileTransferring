package main

import (
	"fileTransferring/shared"
	"fmt"
	"net"
	"os"
	"strings"
	"tideWatchAPI/utils"
)

// TODO: Need to fix issue with files that are smaller than 512 bytes

var connectedClients = make(map[*net.UDPAddr]string)

func main() {
	v6, sw, dp := shared.InterpretCommandLineArguments(os.Args)
	fmt.Printf("IPv6 Packets: %t\nSliding Window: %t\nDrop 1%% of Packets: %t\n", v6, sw, dp)

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
	t := shared.DeterminePacketType(data)

	ack := shared.CreateACKPacket()

	switch t {
	case 2:
		fmt.Println("WRQ packet has been received...")
		r, _ := shared.ReadRRQWRQPacket(data)
		if strings.ToLower(r.Mode) != "octet" {
			sendPacketToClient(conn, addr, createErrorPacket(shared.ERROR_0, "This server only supports octet mode, not: "+r.Mode))
			return
		} else {
			addToAuthenticatedClients(addr, r.Filename)
			errorPacket, hasError := createFile(getFileNameForAddress(addr))
			if hasError {
				sendPacketToClient(conn, addr, errorPacket)
				removeAuthenticatedClient(addr)
				return
			} else {
				ack.BlockNumber = []byte{0, 0}
			}
		}
	case 3:
		isAuthenticated := checkAuthenticatedClient(addr)

		if isAuthenticated {
			d, _ := shared.ReadDataPacket(data)
			errorPacket, hasError := writeToFile(getFileNameForAddress(addr), d.Data)
			if hasError {
				sendPacketToClient(conn, addr, errorPacket)
				return
			} else {
				checkEndOfTransfer(d.Data, addr)
				ack.BlockNumber = d.BlockNumber
			}
		} else {
			sendPacketToClient(conn, addr, createErrorPacket(shared.ERROR_0, fmt.Sprintf("Client has not sent a WRQ Packet, permission denied")))
		}
	case 5:
		// TODO: Do something with an error packet from the client...
	default:
		sendPacketToClient(conn, addr, createErrorPacket(shared.ERROR_0, fmt.Sprintf("Server only supports Opcodes of 2,3, and 5...not: %d", t)))
	}

	sendPacketToClient(conn, addr, shared.CreateAckPacketByteArray(ack))
}

func sendPacketToClient(conn *net.UDPConn, addr *net.UDPAddr, data [] byte) {
	_, _ = conn.WriteToUDP(data, addr)
}

func createFile(fileName string) (ePacket [] byte, hasError bool) {
	_, err := os.Stat(fileName)
	if e, hasError := checkFileError(err); hasError {
		return e, true
	}

	_, err = os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if e, hasError := checkFileError(err); hasError {
		return e, true
	}

	return nil, false
}

func writeToFile(fileName string, data []byte) (eData [] byte, hasError bool) {
	f, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if e, hasError := checkFileError(err); hasError {
		return e, true
	}
	if err != nil {
		if e, hasError := checkFileError(err); hasError {
			return e, true
		}
	}
	if _, err := f.Write(data); err != nil {
		if e, hasError := checkFileError(err); hasError {
			return e, true
		}
	}
	if err := f.Close(); err != nil {
		if e, hasError := checkFileError(err); hasError {
			return e, true
		}
	}

	return nil, false
}

func addToAuthenticatedClients(addr *net.UDPAddr, fileName string) {
	hasBeenAdded := checkAuthenticatedClient(addr)
	if !hasBeenAdded {
		connectedClients[addr] = fileName
	}
}

func checkAuthenticatedClient(toAdd *net.UDPAddr) bool {
	for addr := range connectedClients {
		if addr.IP.Equal(toAdd.IP) {
			return true
		}
	}

	return false
}

func removeAuthenticatedClient(toRemove *net.UDPAddr) {
	for addr := range connectedClients {
		if addr.IP.Equal(toRemove.IP) {
			delete(connectedClients, addr)
			return
		}
	}
}

func getFileNameForAddress(addressToGet *net.UDPAddr) string {
	// TODO: Need to handle non-found addresses if they get through
	for addr := range connectedClients {
		if addr.IP.Equal(addressToGet.IP) {
			return connectedClients[addr]
		}
	}

	return ""
}

func checkEndOfTransfer(data [] byte, addressToRemove *net.UDPAddr) {
	if len(data) < 512 {
		fmt.Println("File has been fully transferred")
		removeAuthenticatedClient(addressToRemove)
	}
}

func checkFileError(err error) (ePacket [] byte, hasError bool) {
	if os.IsExist(err) {
		return createErrorPacket(shared.ERROR_6, shared.ERROR_6_MESSAGE), true
	} else if os.IsPermission(err) {
		return createErrorPacket(shared.ERROR_2, shared.ERROR_2_MESSAGE), true
	} else if os.IsNotExist(err) {
		return nil, false
	} else if err != nil {
		return createErrorPacket(shared.ERROR_0, err.Error()), true
	}

	return nil, false
}

func createErrorPacket(errorCode [] byte, errorMessage string) [] byte {
	ePacket := shared.CreateErrorPacket(errorCode, errorMessage)
	return shared.CreateErrorPacketByteArray(ePacket)
}
