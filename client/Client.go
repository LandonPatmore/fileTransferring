package main

import (
	"fileTransferring/shared"
	"fmt"
	"github.com/mholt/archiver"
	"github.com/pkg/errors"
	"io/ioutil"
	"math"
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var tempZipName string

var fileSize int64
var totalBytesSent int64
var totalPacketsSent int
var totalPacketsToSend int
var packetsLost int

const MaxWindowSize = shared.MaxWindowSize

var ipv6, sw, dp = shared.GetCMDArgs(os.Args, true)

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	var serverAddress = "127.0.0.1"

	//fmt.Print("Server address: ")
	//_, _ = fmt.Scanf("%s", &serverAddress)

	var conn *net.UDPConn
	var connError error

	if ipv6 {
		remoteAddr, err := net.ResolveUDPAddr("udp6", serverAddress+shared.PORT)
		shared.ErrorValidation(err)
		conn, connError = net.DialUDP("udp6", nil, remoteAddr)
		shared.ErrorValidation(connError)
	} else {
		remoteAddr, err := net.ResolveUDPAddr("udp4", serverAddress+shared.PORT)
		shared.ErrorValidation(err)
		conn, connError = net.DialUDP("udp4", nil, remoteAddr)
		shared.ErrorValidation(connError)
	}

	var filePath = "/Users/landon/Desktop/sendable.jpg"
	//fmt.Print("Enter full file path: ")
	//_, _ = fmt.Scanf("%s", &filePath)

	fmt.Println("Buffering file...")
	zipError := zipFiles(filePath)
	shared.ErrorValidation(zipError)

	fileBytes, err := ioutil.ReadFile(tempZipName)
	shared.ErrorValidation(err)
	fmt.Println("File Buffered!")

	file, fileError := os.Open(tempZipName)
	shared.ErrorValidation(fileError)

	defer file.Close()

	fi, err := file.Stat()
	shared.ErrorValidation(err)
	fileSize = fi.Size()

	totalPacketsToSend = determineAmountOfPacketsToSend(fileSize)

	sendWRQPacket(conn, strings.Split(filepath.Base(filePath), ".")[0]+".zip")

	var startTime = time.Now().UnixNano() / 1e6 // get it in milliseconds
	sendFile(conn, fileBytes)
	var endTime = time.Now().UnixNano() / 1e6                                                     // get it in milliseconds
	fmt.Printf("Throughput: %f megabits/sec", (float64(fileSize)/float64(endTime-startTime))/125) // from bytes/millsecond to megabits/sec
}

// Sends a file to the server
func sendFile(conn *net.UDPConn, fileBytes [] byte) {
	if sw {
		err := slidingWindowSend(conn, fileBytes)
		if err != nil {
			shared.ErrorValidation(err)
			return
		}
		err = os.Remove(tempZipName)
		shared.ErrorValidation(err)
	} else {
		var currentPacket int
		var bytesToSend = fileBytes

		for {
			if len(bytesToSend) >= 512 {
				sendDataPacket(conn, bytesToSend[:512], &currentPacket)
				bytesToSend = bytesToSend[512:]
			} else {
				sendDataPacket(conn, bytesToSend, &currentPacket)
				err := os.Remove(tempZipName)
				shared.ErrorValidation(err)
				break
			}
		}
	}
}

func slidingWindowSend(conn *net.UDPConn, data []byte) error {
	_ = conn.SetReadDeadline(time.Time{}) // no timeout on the client side

	var bytesToSend = data
	var dataPackets [] shared.DataPacket
	var currentPacketNumber int

	for {
		if len(bytesToSend) >= 512 {
			dataPacket := shared.CreateDataPacket(createBlockNumber(&currentPacketNumber), bytesToSend[:512])
			dataPackets = append(dataPackets, *dataPacket)
			bytesToSend = bytesToSend[512:]
		} else {
			dataPacket := shared.CreateDataPacket(createBlockNumber(&currentPacketNumber), bytesToSend[:])
			dataPackets = append(dataPackets, *dataPacket)
			break
		}
	}

	var windowSize = 1
	var windowStart = 0                      // window
	var windowEnd = windowStart + windowSize // window
	var currentPacketsInFlight = 0

	for {
		var blockNumberToReceiveToAdvanceWindow [] byte
		if windowStart < len(dataPackets) {
			blockNumberToReceiveToAdvanceWindow = dataPackets[windowStart].BlockNumber
		} else {
			blockNumberToReceiveToAdvanceWindow = dataPackets[len(dataPackets)-1].BlockNumber
		}

		for i := windowStart; i < windowEnd; i++ {
			if currentPacketsInFlight != windowSize {
				if !shouldDropPacket() {
					if (windowStart + currentPacketsInFlight) < len(dataPackets) {
						var dataPacketToSend = dataPackets[windowStart+currentPacketsInFlight]
						_, _ = conn.Write(dataPacketToSend.ByteArray())
						currentPacketsInFlight++
					}
				} else {
					currentPacketsInFlight++
					fmt.Println("Dropped packet...")
				}
			} else {
				break
			}
		}

		endOfTransfer, sendWasSuccessful, _, err := receiveSlidingWindowPacket(conn, blockNumberToReceiveToAdvanceWindow)
		shared.ErrorValidation(err)

		if endOfTransfer {
			fmt.Println("File transferred fully")
			return nil
		}

		if sendWasSuccessful {
			windowStart += 1 // first move where the window starts
			currentPacketsInFlight--

			windowSize++
			if windowSize > MaxWindowSize {
				windowSize = MaxWindowSize
			}

			fmt.Printf("Window size increased to...%d\n", windowSize)
		} else {
			windowSize /= 2
			if windowSize <= 0 {
				windowSize = 1
			}

			fmt.Printf("Window size decreased to...%d\n", windowSize)

			currentPacketsInFlight = 0
		}

		windowEnd = windowStart + windowSize
	}

}

func readSlidingWindowPacket(data [] byte, blockNumber [] byte) (endOfTransfer bool, sendSuccessful bool, blockNumberReceived [] byte, err error) {
	opcode := data[1]

	switch opcode {
	case 4: // ack packet
		ack, _ := shared.ReadACKPacket(data)
		if shared.BlockNumberChecker(ack.BlockNumber, blockNumber) {
			return false, true, nil, nil
		}
		return false, false, ack.BlockNumber, nil
	case 5:
		e, _ := shared.ReadErrorPacket(data)
		return false, false, nil, errors.New(fmt.Sprintf("Error code: %d\nError Message: %s", e.ErrorCode[1], e.ErrorMessage))
	case 7:
		return true, false, nil, nil
	default:
		return false, false, nil, errors.New(fmt.Sprintf("Client can only read Opcodes of 4,5, and 7 in Sliding Window Mode...not: %d", opcode))
	}
}

func receiveSlidingWindowPacket(conn *net.UDPConn, blockNumber [] byte) (bool, bool, [] byte, error) {
	receivedData := make([]byte, 516)

	bytesReceived, _, err := conn.ReadFromUDP(receivedData)
	shared.ErrorValidation(err)

	return readSlidingWindowPacket(receivedData[:bytesReceived], blockNumber)
}

// Creates and sends a WRQ packet
func sendWRQPacket(conn *net.UDPConn, fileName string) {
	var wPacket *shared.RRQWRQPacket

	if sw {
		options := map[string]string{
			"sendingMode": "slidingWindow",
		}
		wPacket = shared.CreateRRQWRQPacket(false, fileName, options)
	} else {
		wPacket = shared.CreateRRQWRQPacket(false, fileName, nil)
	}
	send(conn, wPacket.ByteArray(), []byte{0, 0})
}

// Creates and sends a data packet
func sendDataPacket(conn *net.UDPConn, data [] byte, currentPacket *int) {
	dataPacket := shared.CreateDataPacket(createBlockNumber(currentPacket), data)
	send(conn, dataPacket.ByteArray(), dataPacket.BlockNumber)

	totalBytesSent += int64(len(dataPacket.Data))
	totalPacketsSent++
	displayProgress()
}

// Receives a packet and does something with it based on the opcode
func readSequentialPacket(data [] byte, blockNumber [] byte) error {
	opcode := data[1]

	switch opcode {
	case 4: // ack packet
		ack, _ := shared.ReadACKPacket(data)
		if shared.BlockNumberChecker(ack.BlockNumber, blockNumber) {
			return nil
		}
		return errors.New("Block numbers do not match...")
	case 5:
		e, _ := shared.ReadErrorPacket(data)
		return errors.New(fmt.Sprintf("Error code: %d\nError Message: %s", e.ErrorCode[1], e.ErrorMessage))
	case 6:
		oack, _ := shared.ReadOACKPacket(data)
		if oack.Options["sendingMode"] == "slidingWindow" { // just simulating if there were other options, to set the client to what the server wants...there is only one option for this specific project
			sw = true
		} else {
			sw = false
		}
		return nil
	default:
		return errors.New(fmt.Sprintf("Client can only read Opcodes of 4, 5, and 6...not: %d", opcode))
	}
}

// Generates a block number based on the current packet number
func createBlockNumber(currentPacketNumber *int) [] byte {
	*currentPacketNumber++
	leadingByte := math.Floor(float64(*currentPacketNumber / 256))
	return [] byte{byte(leadingByte), byte(*currentPacketNumber % 256)}
}

// Sends the packet to the server
func send(conn *net.UDPConn, data []byte, blockNumber [] byte) {
	for i := 0; i < 10; i++ {
		if !shouldDropPacket() {
		_, _ = conn.Write(data)
		}
		receivedData, err := receiveSequentialPacket(conn)
		if err == nil {
			err := readSequentialPacket(receivedData, blockNumber)
			shared.ErrorValidation(err)
			return
		} else {
			packetsLost++
			displayProgress()
			time.Sleep(time.Millisecond * 100)
		}
	}

	shared.ErrorValidation(errors.New("Total retries exhausted...exiting"))
}

// Handles the read timeout of the server sending back an ACK
func receiveSequentialPacket(conn *net.UDPConn) ([] byte, error) {
	_ = conn.SetReadDeadline(time.Now().Add(time.Millisecond * 500))

	receivedData := make([]byte, 516)
	bytesReceived, _, timedOut := conn.ReadFromUDP(receivedData)

	return receivedData[:bytesReceived], timedOut
}

// Displays a progress bar that updates with the total data and total packets sent
func displayProgress() {
	var totalDataSent = math.Ceil(float64(totalBytesSent) / float64(fileSize) * 100)

	fmt.Print("\r")
	fmt.Printf("Progress: (%d%% | Packets Lost: %d | %d/%d packets sent) ", int(totalDataSent), packetsLost, totalPacketsSent, totalPacketsToSend)
}

// Returns amount of packets that must be sent for file to be transferred
// completely
func determineAmountOfPacketsToSend(fileSize int64) int {
	return int(math.Ceil(float64(fileSize) / 512))
}

// Figures out whether or not to drop the current packet
func shouldDropPacket() bool {
	if dp {
		return rand.Float64() <= .01
	}

	return false
}

// Takes in a path and recursively goes down the directory tree and creates a zip to send to the server
func zipFiles(path string) error {
	generateTempZipName()

	var filesToZip [] string

	err := filepath.Walk(path,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			fi, err := os.Stat(path)
			if err != nil {
				return err
			}
			if fi.Mode().IsRegular() {
				filesToZip = append(filesToZip, path)
			}
			return nil
		})

	if err != nil {
		return err
	}

	return archiver.Archive(filesToZip, tempZipName)
}

// Generates a random name for temporary zip file
func generateTempZipName() {
	bytes := make([]byte, 10)
	for i := 0; i < 10; i++ {
		bytes[i] = byte(65 + rand.Intn(25)) //A=65 and Z = 65+25
	}

	tempZipName = string(bytes) + ".zip"
}
