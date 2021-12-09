package websock

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"log"
)

type Message struct {
	msgType int
	message []byte
}

func send(buffer chan *Message, msgType int, message []byte) {
	buffer <- &Message{
		msgType: msgType,
		message: message,
	}
}

func sendJson(buffer chan *Message, message interface{}) {
	if jsonMessage, err := json.Marshal(&message); err != nil {
		log.Printf("invalid message in send object : %+v\n", message)
	} else {
		send(buffer, websocket.TextMessage, jsonMessage)
	}
}
