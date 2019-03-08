package main

import (
	"fileTransferring/shared"
	"fmt"
	"log"
	"net"
	"os"
	"tideWatchAPI/utils"
)

var connectedClients = make(map[*net.UDPAddr]string)

func main() {
	ServerAddr, err := net.ResolveUDPAddr("udp", ":8274")
	shared.ErrorValidation(err)
	conn, err := net.ListenUDP("udp", ServerAddr)
	shared.ErrorValidation(err)

	defer conn.Close()

	fmt.Println("Server started...")

	for {
		receivePacket(conn)
	}
}

func receivePacket(conn *net.UDPConn) {
	// TODO: When an error occurs here, send an error packet back (possibly if it is the client)

	data := make([]byte, 1024)
	_, addr, err := conn.ReadFromUDP(data)
	data = shared.StripOffExtraneousBytes(data)
	utils.ErrorCheck(err)
	t := shared.DeterminePacketType(data)

	ack := shared.CreateACKPacket()

	switch t {
	case 2:
		// TODO: Send error packet if error
		r, _ := shared.ReadRRQWRQPacket(data)
		addToAuthenticatedClients(addr, r.Filename)
		err := createFile(r.Filename)
		shared.ErrorValidation(err)
		ack.BlockNumber = []byte{0, 0}
	case 3:
		isAuthenticated := checkAuthenticatedClient(addr)

		if isAuthenticated {
			d, _ := shared.ReadDataPacket(data)
			writeToFile(getFileNameForAddress(addr), d.Data)
		} else {
			// TODO: Error packet to be sent back
			log.Fatal("This client has not been authenticated")
		}

	default:
		log.Fatal("Server can only read Opcodes of 2 and 3...not: ", t)
	}

	_, _ = conn.WriteToUDP(shared.CreateAckPacketByteArray(ack), addr)
}

func createFile(fileName string) error {
	_, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	return err
}

func writeToFile(fileName string, data []byte) {
	// TODO: Error packets need to be sent here
	fmt.Println("Writing to file...")
	f, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	if _, err := f.Write(data); err != nil {
		log.Fatal(err)
	}
	if err := f.Close(); err != nil {
		log.Fatal(err)
	}
}

func addToAuthenticatedClients(addr *net.UDPAddr, fileName string) {
	hasBeenAdded := checkAuthenticatedClient(addr)
	if !hasBeenAdded {
		connectedClients[addr] = "_test_" + fileName
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

func getFileNameForAddress(addressToGet *net.UDPAddr) string {
	// TODO: Need to handle non-found addresses if they get through
	for addr := range connectedClients {
		if addr.IP.Equal(addressToGet.IP) {
			return connectedClients[addr]
		}
	}

	return ""
}
