package main

import (
	"fmt"
	"log"
	"net"
)

const port = "6969"

type MessageType int

const (
	ClientConnected MessageType = iota + 1
	DeleteClient
	NewMessage
)

type Message struct {
	Type MessageType
	Conn net.Conn
	Text string
}

func server(messages chan Message) {
	conns := map[string]net.Conn{}
	for {
		msg := <-messages
		switch msg.Type {
		case DeleteClient:
			delete(conns, msg.Conn.RemoteAddr().String())
		case ClientConnected:
			conns[msg.Conn.RemoteAddr().String()] = msg.Conn
		case NewMessage:
			for _, conn := range conns {
				_, err := conn.Write([]byte(msg.Text))
				if err != nil {
					fmt.Println("Could nor send data to %s: %s", conn.RemoteAddr(), err)
				}
			}
		}
	}
}

func clients(conn net.Conn, messages chan Message) {
	buffer := make([]byte, 512)
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			conn.Close()
			messages <- Message{
				Type: DeleteClient,
				Conn: conn,
			}
			return
		}

		messages <- Message{
			Type: NewMessage,
			Text: string(buffer[0:n]),
			Conn: conn,
		}
	}
}
func main() {
	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Error : could not listen to the port :%s\n", port)
	}
	log.Printf("Listning to Tcp connection on port :%s .....", port)

	messages := make(chan Message)
	go server(messages)

	for {
		conn, err := ln.Accept()
		if err != nil {
			// handle error:
			log.Printf("Errro: could not accept a connection: %s\n", err)
		}
		log.Printf("Accepter connection from %s", conn.RemoteAddr())
		messages <- Message{
			Type: ClientConnected,
			Conn: conn,
		}
		go clients(conn, messages)
	}

}
