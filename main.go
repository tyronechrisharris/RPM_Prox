package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"
	"time"
)

var (
	clients = make(map[net.Conn]bool)
	mu      sync.Mutex
)

func handleClient(conn net.Conn) {
	log.Printf("CAS connected: %s", conn.RemoteAddr())
	mu.Lock()
	clients[conn] = true
	mu.Unlock()

	defer func() {
		mu.Lock()
		delete(clients, conn)
		mu.Unlock()
		conn.Close()
		log.Printf("CAS disconnected: %s", conn.RemoteAddr())
	}()

	buf := make([]byte, 1024)
	for {
		if _, err := conn.Read(buf); err != nil {
			return
		}
	}
}

func connectToRPM(rpmAddr string) {
	for {
		log.Printf("Attempting connection to RPM at %s...", rpmAddr)
		rpmConn, err := net.DialTimeout("tcp", rpmAddr, 5*time.Second)
		if err != nil {
			log.Printf("RPM connection failed: %v. Retrying in 5s...", err)
			time.Sleep(5 * time.Second)
			continue
		}
		log.Println("Successfully connected to RPM stream.")

		buf := make([]byte, 4096)
		for {
			n, err := rpmConn.Read(buf)
			if err != nil {
				if err != io.EOF {
					log.Printf("RPM read error: %v", err)
				}
				rpmConn.Close()
				break 
			}

			data := buf[:n]
			mu.Lock()
			for client := range clients {
				if _, err := client.Write(data); err != nil {
					client.Close() 
				}
			}
			mu.Unlock()
		}
	}
}

func main() {
	if len(os.Args) != 4 {
		fmt.Println("Usage: multiplexer <listen_port> <rpm_ip> <rpm_port>")
		os.Exit(1)
	}

	listenPort := os.Args[1]
	rpmIP := os.Args[2]
	rpmPort := os.Args[3]

	go connectToRPM(fmt.Sprintf("%s:%s", rpmIP, rpmPort))

	listener, err := net.Listen("tcp", ":"+listenPort)
	if err != nil {
		log.Fatalf("Failed to bind to port %s: %v", listenPort, err)
	}
	defer listener.Close()

	log.Printf("Listening for CAS connections on port %s", listenPort)
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}
		go handleClient(conn)
	}
}
