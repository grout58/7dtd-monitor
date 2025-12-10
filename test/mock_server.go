package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

func main() {
	listener, err := net.Listen("tcp", "localhost:8081")
	if err != nil {
		fmt.Println("Error starting mock server:", err)
		return
	}
	defer listener.Close()
	fmt.Println("Mock 7DTD Telnet Server listening on localhost:8081")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	writer := bufio.NewWriter(conn)
	scanner := bufio.NewScanner(conn)

	// Initial handshake
	writer.WriteString("Please enter password:\r\n")
	writer.Flush()

	// Expect password (we'll accept anything for mock)
	if scanner.Scan() {
		_ = scanner.Text() // consume password
		writer.WriteString("Logon successful.\r\n\r\n")
		writer.Flush()
	}

	for {
		if !scanner.Scan() {
			return
		}
		cmd := strings.TrimSpace(scanner.Text())
		response := ""

		switch cmd {
		case "lpi":
			response = `Total of 3 in the game
1. id=171, Survivor (PL), pos=(-1050.5, 65.0, 890.3), rot=(0.0, -135.0, 0.0), remote=True, health=150, deaths=2, zombies=0, players=0, score=15, level=13, steamid=76561198012345678, ip=127.0.0.1, ping=24
2. id=172, ZombieSlayer, pos=(-1040.1, 65.0, 895.1), rot=(0.0, 45.0, 0.0), remote=True, health=80, deaths=5, zombies=12, players=1, score=55, level=24, steamid=76561198087654321, ip=192.168.0.5, ping=45
3. id=175, Newbie, pos=(-1060.0, 64.0, 880.0), rot=(0.0, 0.0, 0.0), remote=True, health=100, deaths=0, zombies=0, players=0, score=0, level=1, steamid=76561199000000001, ip=10.0.0.99, ping=120
`
		case "mem":
			// Output simulates 7DTD memory output often containing FPS
			response = `Heap: 2500.5 MB, Max: 3500.0 MB, Objects: 150000
GC: 120.5 MB
FPS: 58.4
`
		case "le":
			response = `Total of 5 entities in the game
1. id=171, name=Survivor (PL), ... type=Player
2. id=801, name=Zombie Boe, ... type=Zombie
3. id=802, name=Zombie Joe, ... type=Zombie
4. id=803, name=Stag, ... type=Animal
5. id=804, name=Bear, ... type=Animal
`
		case "gettime":
			response = `Day 7, 21:45
`
		case "exit", "quit":
			writer.WriteString("Goodbye.\r\n")
			writer.Flush()
			return
		default:
			response = fmt.Sprintf("*** Unknown command: %s\r\n", cmd)
		}

		writer.WriteString(response)
		writer.WriteString("\r\n") // Prompt spacing
		writer.Flush()
	}
}
