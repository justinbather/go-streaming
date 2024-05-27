package main

import (
	"bufio"
	"crypto/rand"
	"fmt"
	amf "github.com/speps/go-amf"
	"io"
	"log"
	"net"
)

const (
	rtmpPort      = ":1935"
	rtmpVersion   = 3
	handshakeSize = 1536
)

func main() {

	listener, err := net.Listen("tcp", rtmpPort)
	if err != nil {
		log.Fatalf("Error connecting to socket: %s", err)
	}

	log.Printf("RTMP Server listening on %s", rtmpPort)

	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalf("Error accepting connection: %s", err)
			continue
		}

		go handleConnection(conn)
	}

}

func handleConnection(conn net.Conn) {
	log.Printf("Recieved a new connection from %s", conn.LocalAddr())
	err := performHandshake(conn)
	if err != nil {
		log.Fatalf("Error performing handshake: %s", err)
	}

	log.Println("Handshake Successful")

	reader := bufio.NewReader(conn)

	for {

		data, err := reader.ReadBytes('\n')
		if err != nil {
			log.Fatalf("error reading data after handshake: %s", err)
		}

		decoded, err := amf.DecodeAMF0(data)
		if err != nil {
			log.Fatalf("Error decoding AMF: %s", err)
		}

		fmt.Println(decoded)

		/*	log.Println(data)
			log.Println(data[0] == 0x02)
			parsedStr := ""
			if data[0] == 0x02 {
				for _, d := range data[1:] {
					parsedStr += string(d)
				}
				log.Println(parsedStr)

			}
		*/
	}

}

func performHandshake(conn net.Conn) error {
	// inital 2 packets c0 and c1 are sent from client
	// c0 contains 0x03 (1 byte) which is our rtmpVersion
	// c1 contains 1536 random bytes
	c0c1 := make([]byte, 1+handshakeSize)

	// read full reads exacly the amount of bytes from the buffer c0c1
	// returns number of bytes copied and err
	//
	// Client first sends 2 packets, 1 with 0x03 (3 our rtmpVersion number (c0)), and 1536 random bytes (c1)
	_, err := io.ReadFull(conn, c0c1)
	if err != nil {
		return err
	}

	if c0c1[0] != rtmpVersion {
		return fmt.Errorf("Incompatible RTMP Version. Expected %d but got %d", rtmpVersion, c0c1[0])
	}

	// Need to save c1 to send back to client
	c1 := c0c1[1:]

	// Write our rtmpVersion number and 1536 random bytes to client
	s0s1 := make([]byte, 1+handshakeSize)
	s0s1[0] = rtmpVersion

	_, err = rand.Read(s0s1[1:])
	if err != nil {
		return fmt.Errorf("Error writing random bytes to handshake s1: %s", err)
	}

	_, err = conn.Write(s0s1)
	if err != nil {
		return fmt.Errorf("Error writing to client during handshake: %s", err)
	}

	// Send back the clients initial 1536 random bytes (c1)
	_, err = conn.Write(c1)
	if err != nil {
		return fmt.Errorf("Error writing c1 to client: %s", err)
	}

	// Client resends our random 1536 bytes
	c2 := make([]byte, handshakeSize)
	_, err = io.ReadFull(conn, c2)
	if err != nil {
		return fmt.Errorf("Didn't get expected return from client c2 packet: %s", err)
	}

	// We could check here to make sure its the same as our s1

	return nil
}
