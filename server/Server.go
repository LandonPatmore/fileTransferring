package main

import (
	"fileTransferring/shared"
	"fmt"
	"net"
)

func main() {
	ServerAddr, err := net.ResolveUDPAddr("udp", ":8274")
	shared.ErrorValidation(err)
	conn, err := net.ListenUDP("udp", ServerAddr)
	shared.ErrorValidation(err)

	defer conn.Close()

	message := make([]byte, 1024)

	fmt.Println("Server started...")

	for {
		_, _, _ = conn.ReadFromUDP(message)
		r, _ := shared.ReadRRQWRQPacket(message)
		fmt.Println(r)
	}
}
