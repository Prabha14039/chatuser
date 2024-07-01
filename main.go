package main

import (
	"fmt"
	"log"
	"net"
	"time"
)

const port = "6969"
const MessageRate = 0.5
const limit = 5
const BanLimit = 10 * 50

type MessageType int

const (
	ClientConnected MessageType = iota + 1
	ClientDisconnected
	NewMessage
)

type Message struct {
	Type MessageType
	Conn net.Conn
	Text string
}

type Client struct {
	Conn        net.Conn
	LastMessage time.Time
	StrikeCount int
}

func server(messages chan Message) {
	clients := map[string]*Client{}
	bannedMF := map[string]time.Time{}
	for {
		msg := <-messages
		switch msg.Type {
		case ClientDisconnected:
			delete(clients, msg.Conn.RemoteAddr().String())
		case ClientConnected:
			addr := msg.Conn.RemoteAddr().(*net.TCPAddr)
			BannedAT, banned := bannedMF[addr.IP.String()]

			if banned {
				if time.Now().Sub(BannedAT).Seconds() >= BanLimit {
					delete(bannedMF, addr.IP.String())
					banned = false

				}
			}

			if !banned {
				log.Printf("Client %s connected", msg.Conn.RemoteAddr())
				clients[msg.Conn.RemoteAddr().String()] = &Client{
					Conn:        msg.Conn,
					LastMessage: time.Now(),
				}
			}

		case NewMessage:
			now := time.Now()
			addr := msg.Conn.RemoteAddr().String()
			author := clients[addr]
			if now.Sub(author.LastMessage).Seconds() > MessageRate && author.StrikeCount < limit {
				author.LastMessage = now
				log.Printf("Client %s sent Message %s", msg.Conn.RemoteAddr(), msg.Text)
				for _, client := range clients {
					if client.Conn.RemoteAddr().String() != msg.Conn.RemoteAddr().String() {
						_, err := client.Conn.Write([]byte(msg.Text))
						if err != nil {
							fmt.Println("Could not send data to %s: %s", client.Conn.RemoteAddr(), err)
						}
					}
				}
			} else {
				if author.StrikeCount < limit {
					author.StrikeCount = +1
				} else {
					author
				}

			}
		}
	}
}

func clients(conn net.Conn, messages chan Message) {
	buffer := make([]byte, 64)
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			log.Printf("Could Not read from client %s", conn.RemoteAddr())
			conn.Close()
			messages <- Message{
				Type: ClientDisconnected,
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
