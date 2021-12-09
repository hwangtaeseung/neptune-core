package websock

import (
	"github.com/gorilla/websocket"
	"log"
	"net/url"
)

type WSClient struct {
	done                chan struct{}
	messageQueueToWrite chan *Message
	client              *websocket.Conn

	OnConnect    func(*websocket.Conn)
	OnDisconnect func(*websocket.Conn)

	OnReadMessage  func(message *Message)
	OnWriteMessage func(message *Message)

	OnError func(error)
}

func (w *WSClient) Connect(address string, path string) error {

	u := url.URL{
		Scheme: "ws",
		Host:   address,
		Path:   path,
	}

	// connect
	client, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		if w.OnError != nil {
			w.OnError(err)
		}
		log.Printf("connection error : %+v\n", err)
		return err
	}
	w.client = client

	// create channel
	w.done = make(chan struct{})
	w.messageQueueToWrite = make(chan *Message, 256)

	// go routine to read message
	go w.processToRead()

	// go routine to write message
	go w.processToWrite()

	// connect event
	if w.OnConnect != nil {
		w.OnConnect(client)
	}

	return nil
}

func (w *WSClient) processToRead() {

	defer func() {
		log.Printf("exit goroutine for reading message")
		close(w.done)
	}()

	for {
		msgType, message, err := w.client.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err,
				websocket.CloseGoingAway, websocket.CloseAbnormalClosure, websocket.CloseNoStatusReceived) {
				log.Printf("websocket client closed : %v\n", err)
			} else {
				log.Printf("read error : %+v\n", err)
			}

			// disconnect event
			if w.OnDisconnect != nil {
				w.OnDisconnect(w.client)
			}
			return
		}
		// call read event
		if w.OnReadMessage != nil {
			w.OnReadMessage(&Message{
				msgType: msgType,
				message: message,
			})
		}
	}
}

func (w *WSClient) processToWrite() {

	defer func() {
		log.Printf("exit goroutine for writing message")
		close(w.messageQueueToWrite)
	}()

	for {
		select {
		case <-w.done:
			return
		case message, ok := <-w.messageQueueToWrite:
			log.Printf("send message : %+v\n", message)

			if !ok {
				_ = w.client.WriteMessage(websocket.CloseMessage, []byte{})
				log.Printf("text message queue channel is closed")
				return
			}

			if err := w.client.WriteMessage(message.msgType, message.message); err != nil {
				log.Printf("write error : %+v\n", err)
				return
			}
			// call write event
			if w.OnWriteMessage != nil {
				w.OnWriteMessage(message)
			}
		}
	}
}

func (w *WSClient) Disconnect() {
	if err := w.client.WriteMessage(websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")); err != nil {
		log.Printf("disconnect error : %+v\n", err)
	}
}

func (w *WSClient) Send(msgType int, message []byte) {
	send(w.messageQueueToWrite, msgType, message)
}

func (w *WSClient) SendObject(message interface{}) {
	_ = w.client.WriteMessage(websocket.CloseMessage, []byte{})
	sendJson(w.messageQueueToWrite, message)
}
