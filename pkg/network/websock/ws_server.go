package websock

import (
	"context"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"reflect"
	"time"
)

// WSServer WebSocket Server
type WSServer struct {

	// http server
	server *http.Server

	// sessions
	sessions map[*WSSession]bool

	// Inbound messages from the sessions.
	broadcast chan *Message

	// Register requests from the sessions.
	register chan *WSSession

	// Unregister requests from sessions.
	unregister chan *WSSession

	// ws handler
	wsHandler *WSHandler

	// on connect
	OnConnect func(*WSSession)

	// on disconnect
	OnDisconnect func(*WSSession)
}

type StaticFileHandler struct {
	PathPrefix  string
	Folder      string
	CertsFolder string
}

type HttpHandler struct {
	Path           string
	Method         string
	MessageHandler func(http.ResponseWriter, *http.Request)
}

func NewWSServer(addr string, wsHandler *WSHandler,
	staticFile *StaticFileHandler, httpHandlers ...*HttpHandler) *WSServer {

	// create websocket network
	wsServer := &WSServer{
		broadcast:  make(chan *Message),
		register:   make(chan *WSSession),
		unregister: make(chan *WSSession),
		sessions:   make(map[*WSSession]bool),
		wsHandler:  wsHandler,
	}

	// router for http server
	router := mux.NewRouter()

	// set up http handler
	if httpHandlers != nil {
		for _, httpHandler := range httpHandlers {
			router.HandleFunc(httpHandler.Path, httpHandler.MessageHandler).Methods(httpHandler.Method)
		}
	}

	// default websocket handler
	router.HandleFunc("/ws", func(writer http.ResponseWriter, request *http.Request) {
		runWSSession(wsServer, writer, request)
	})

	// set static files
	if staticFile != nil {
		fs := http.FileServer(http.Dir(staticFile.Folder))
		router.PathPrefix(staticFile.PathPrefix).Handler(http.StripPrefix(staticFile.PathPrefix, fs))
	}

	// set properties
	wsServer.server = &http.Server{
		Addr:    addr,
		Handler: router,
	}

	return wsServer
}

func (s *WSServer) MsgHandler(session *WSSession, messageType int, message []byte) {
	switch messageType {
	case websocket.TextMessage:
		if s.wsHandler.textMessageHandler != nil {
			s.wsHandler.textMessageHandler(session, string(message))
		}
	case websocket.BinaryMessage:
		if s.wsHandler.binaryMessageHandler != nil {
			s.wsHandler.binaryMessageHandler(session, message)
		}
	}
}

func (s *WSServer) RunWithTLS(certFile, KeyFile string) {
	s.run(func() error {
		return s.server.ListenAndServeTLS(certFile, KeyFile)
	})
}

func (s *WSServer) Run() {
	s.run(func() error {
		return s.server.ListenAndServe()
	})
}

func (s *WSServer) run(callback func() error) {

	// run to process websocket client
	go s.processSession()

	// listen & serve
	go func() {
		if err := callback(); err != nil {
			if reflect.TypeOf(err) != reflect.TypeOf(http.ErrServerClosed) {
				log.Printf("network finish : %v\n", err)
				panic(err)
			}
		}
	}()

	log.Printf("network hans been started... (listen=%v)\n", s.server.Addr)
}

func (s *WSServer) Stop() *WSServer {

	defer func() {
		close(s.broadcast)
		close(s.unregister)
		close(s.register)
	}()

	// clear sessions map
	for client := range s.sessions {
		s.unregister <- client
		log.Printf("unregister client : %v\n", client)
	}

	// wait for...
	sessionCount := len(s.sessions)
	for sessionCount != 0 {
		log.Printf("network is terminating... (session count : %v)\n", sessionCount)
		time.Sleep(time.Second)
		sessionCount = len(s.sessions)
	}

	// shutdown http network
	_ = s.server.Shutdown(context.Background())

	log.Printf("network shutdown gracefully")

	return s
}

func (s *WSServer) processSession() {

	for {
		select {
		case client, ok := <-s.register:
			if !ok {
				log.Printf("register channel closed")
				return
			}
			s.sessions[client] = true
			log.Printf("session has been created. count of session : %v\n", len(s.sessions))

		case session, ok := <-s.unregister:
			if !ok {
				log.Printf("unregister channel closed")
				return
			}

			// remove session object from map
			if _, ok := s.sessions[session]; ok {
				// call disconnect handler
				if s.OnDisconnect != nil {
					s.OnDisconnect(session)
				}
				delete(s.sessions, session)
				close(session.send)
			}
			log.Printf("session has been destroyed. count of session : %v\n", len(s.sessions))

		case message, ok := <-s.broadcast:
			if !ok {
				log.Printf("broadcast channel closed")
				return
			}
			for session := range s.sessions {
				select {
				case session.send <- message:
				default:
					close(session.send)
					delete(s.sessions, session)
				}
			}
		}
	}
}
