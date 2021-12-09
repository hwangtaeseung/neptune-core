package main

import (
	"github.com/gorilla/websocket"
	"log"
	"neptune-core/pkg/common"
	"neptune-core/pkg/network/websock"
)

const defaultAddress = ":7777"

func main() {

	// create ws server
	server := websock.NewWSServer(
		// default address
		defaultAddress,
		// web socket handler
		websock.NewWSHandler(
			// text receive handler
			func(client *websock.WSSession, message string) {
				log.Printf("[TEXT] received message : %v\n", message)
				client.Send(websocket.TextMessage, []byte(message))
			},
			// binary receive handler
			func(client *websock.WSSession, message []byte) {
				log.Printf("[BINARY] received message : %v\n", message)
				client.Send(websocket.BinaryMessage, message)
			},
		),
		nil)

	// run server
	server.Run()

	// shutdown handler
	common.WaitForShutdown(func() {

		log.Printf("server is terminating. (interrupt)")

		// stop network
		server.Stop()
	})
}
