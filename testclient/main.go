package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"time"
)

func main() {

	if len(os.Args) < 2 {
		fmt.Println("Please provide host:port to connect to")
		os.Exit(1)
	}

	// Resolve the string address to a UDP address
	udpAddr, err := net.ResolveUDPAddr("udp", os.Args[1])

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Dial to the address with UDP
	conn, err := net.DialUDP("udp", nil, udpAddr)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// // Send a message to the server
	// _, err = conn.Write([]byte(os.Args[2]))
	// fmt.Println("send...")
	// if err != nil {
	// 	fmt.Println(err)
	// 	os.Exit(1)
	// }

	// // Read from the connection untill a new line is send
	// data, err := bufio.NewReader(conn).ReadString('\n')
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// // Print the data read from the connection to the terminal
	// fmt.Print("> ", string(data))

	// Read from the connection untill a new line is send
	go func() {
		for {
			data, err := bufio.NewReader(conn).ReadString('\n')
			if err != nil {
				fmt.Println(err)
				return
			}
			// Print the data read from the connection to the terminal
			fmt.Print("> ", string(data))
		}
	}()

	// Send a message to the server
	// go func() {
	for c := range os.Args[2] {
		_, err = conn.Write([]byte(string(os.Args[2][c])))
		if err != nil {
			fmt.Println(err)
			return
		}
		time.Sleep(1 * time.Second)
	}
	// }()

}
