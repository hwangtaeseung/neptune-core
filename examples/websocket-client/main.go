package main

import (
	"github.com/gorilla/websocket"
	"log"
	"neptune-core/pkg/network/websock"
	"time"
)

func main() {

	client := &websock.WSClient{

		// connect handler
		OnConnect: func(conn *websocket.Conn) {
			log.Printf("connected sucessfully... (%v)\n", conn.RemoteAddr().String())
		},

		OnDisconnect: func(conn *websocket.Conn) {
			log.Printf("disconnected sucessfully...(%v)\n", conn.RemoteAddr().String())
		},

		// receive handler
		OnReadMessage: func(message *websock.Message) {
			log.Printf("received message : %+v\n", message)
		},

		OnWriteMessage: func(message *websock.Message) {
			log.Printf("sent message : %+v\n", message)
		},

		// error
		OnError: func(err error) {
			log.Printf("error : %v\n", err)
		},
	}

	// connect to server
	if err := client.Connect("0.0.0.0:7777", "/ws"); err != nil {
		log.Printf("app connection erorr : %v\n", err)
		return
	}

	// send text message
	client.Send(websocket.TextMessage, []byte("반가워요~!!! ^^"))

	// send binary message
	client.Send(websocket.BinaryMessage, []byte{ 0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07 })

	// wait for 2 seconds
	time.Sleep(2 * time.Second)

	// disconnect server
	client.Disconnect()

	// wait for 2 seconds
	time.Sleep(2 * time.Second)

	log.Printf("web socket client has been finished")
}
