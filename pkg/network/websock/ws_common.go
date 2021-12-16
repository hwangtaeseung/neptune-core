package websock

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"log"
)

type Message struct {
	MsgType int
	Message []byte
}

func send(buffer chan *Message, msgType int, message []byte) {
	buffer <- &Message{
		MsgType: msgType,
		Message: message,
	}
}

func sendJson(buffer chan *Message, message interface{}) {
	if jsonMessage, err := json.Marshal(&message); err != nil {
		log.Printf("invalid message in send object : %+v\n", message)
	} else {
		log.Printf("message sent to client : %v\n", string(jsonMessage))
		send(buffer, websocket.TextMessage, jsonMessage)
	}
}
