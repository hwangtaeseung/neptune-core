package websock

import (
	"bytes"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"time"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 1024 * 10
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upGrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type WSSession struct {
	// network
	Server *WSServer

	// The web socket connection.
	Conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan *Message

	// user context
	UserContext interface{}
}

func (w *WSSession) processToRead() {

	defer func() {
		w.Server.unregister <- w
		//_ = w.Conn.Close()
		log.Printf("client read go routine stop..")
	}()

	w.Conn.SetReadLimit(maxMessageSize)
	_ = w.Conn.SetReadDeadline(time.Now().Add(pongWait))
	w.Conn.SetPongHandler(func(string) error {
		_ = w.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		messageType, message, err := w.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err,
				websocket.CloseGoingAway, websocket.CloseAbnormalClosure, websocket.CloseNoStatusReceived) {
				log.Printf("websocket client close ==> %+v\n", err)
			} else {
				log.Printf("websock client read error ==> %+v", err)
			}
			return
		}

		// remove '\n'
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))

		// callback
		w.Server.MsgHandler(w, messageType, message)
	}
}

func (w *WSSession) processToWrite() {

	ticker := time.NewTicker(pingPeriod)

	defer func() {
		ticker.Stop()
		_ = w.Conn.Close()
		log.Printf("client write go routine stop..")
	}()

	for {
		select {
		case buffer, ok := <-w.send:
			_ = w.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				err := w.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				if err != nil {
					log.Printf("websocket write error : %v\n", err)
				}
				log.Println("websocket client send channel closed")
				return
			}

			err := w.Conn.WriteMessage(buffer.MsgType, buffer.Message)
			if err != nil {
				log.Printf("write close error : %v\n", err)
				return
			}

		case <-ticker.C:
			_ = w.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := w.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("tick write error : %v\n", err)
				return
			}
		}
	}
}

func (w *WSSession) SendString(message string)  {
	send(w.send, websocket.TextMessage, []byte(message))
}

func (w *WSSession) SendMessage(message *Message)  {
	send(w.send, message.MsgType, message.Message)
}

func (w *WSSession) Send(msgType int, message []byte) {
	send(w.send, msgType, message)
}

func (w *WSSession) SendJson(message interface{}) {
	sendJson(w.send, message)
}

func (w *WSSession) Close() {
	w.Server.unregister <- w
}

func runWSSession(server *WSServer, responseWriter http.ResponseWriter, request *http.Request) {

	connection, err := upGrader.Upgrade(responseWriter, request, nil)
	if err != nil {
		log.Printf("upgrade error : %v\n", err)
		return
	}

	// create client
	client := &WSSession{
		Server: server,
		Conn:   connection,
		send:   make(chan *Message, 256),
	}

	// register client
	server.register <- client

	// call connect handler
	if server.OnDisconnect != nil {
		server.OnConnect(client)
	}

	// run read/write go routine
	go client.processToWrite()
	go client.processToRead()
}
