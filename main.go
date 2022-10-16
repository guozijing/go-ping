package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"time"
)

var (
	timeout int64
	size    int
	count   int
)

type ICMP struct {
	Type        uint8
	Code        uint8
	CheckSum    uint16
	ID          uint16
	SequenceNum uint16
}

func main() {
	GetCommandArgs()
	desIP := os.Args[len(os.Args)-1]
	// fmt.Println(timeout, size, count, desIP)

	conn, err := net.DialTimeout("ip:icmp", desIP, time.Duration(timeout)*time.Millisecond)
	if err != nil {
		log.Fatalln(err)
	}
	defer conn.Close()
	remoteAddr := conn.RemoteAddr()
	fmt.Println("Address of remote: ", remoteAddr)

	for i := 0; i < count; i++ {
		startTime := time.Now()
		icmp := &ICMP{
			Type:        uint8(8),
			Code:        uint8(0),
			CheckSum:    uint16(0),
			ID:          uint16(i),
			SequenceNum: uint16(i),
		}

		var buf bytes.Buffer
		binary.Write(&buf, binary.BigEndian, icmp)
		data := make([]byte, size)
		buf.Write(data)
		data = buf.Bytes()
		checkSum := checkSum(data)
		data[2] = byte(checkSum >> 8)
		data[3] = byte(checkSum)

		conn.SetDeadline(time.Now().Add(time.Duration(timeout) * time.Millisecond))
		n, err := conn.Write(data)
		if err != nil {
			log.Fatalln("Write err: ", err)
		}

		bufR := make([]byte, 1024)
		n, err = conn.Read(bufR)
		if err != nil {
			log.Fatalln("Read err: ", err)
		}
		fmt.Printf("%d.%d.%d.%d: bytes=%d, time=%dus, TTL=%d \n", bufR[12],
			bufR[13], bufR[14], bufR[15], n-28, time.Since(startTime).Microseconds(), bufR[8])
		time.Sleep(time.Second)
	}
}

func GetCommandArgs() {
	flag.Int64Var(&timeout, "w", 1000, "Time of request timeout")
	flag.IntVar(&size, "l", 64, "Size of package")
	flag.IntVar(&count, "n", 10, "Count of request times")
	flag.Parse()
}

func checkSum(data []byte) uint16 {
	length := len(data)
	index := 0
	var sum uint32
	for length > 1 {
		sum += uint32(data[index])<<8 + uint32(data[index+1])
		length -= 2
		index += 2
	}
	if length == 1 {
		sum += uint32(data[index])
	}
	hi := sum >> 16
	for hi != 0 {
		sum = hi + uint32(uint16(sum))
		hi = sum >> 16
	}
	return uint16(^sum)
}
