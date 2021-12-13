package websock

import (
	"reflect"
)

type WSMessageHandler func(session *WSSession, message []byte)

type WSContext struct {
	MessageHandler func(client *WSSession, message interface{})
	MessageType    reflect.Type
}

type WSHandler struct {
	handlers             map[string]*WSContext
	binaryMessageHandler WSMessageHandler
	textMessageHandler   WSMessageHandler
}

// NewWSHandler NewHandler create web socket handler
func NewWSHandler(textMessageHandler WSMessageHandler, binaryMessageHandler WSMessageHandler) *WSHandler {
	return &WSHandler{
		textMessageHandler:   textMessageHandler,
		binaryMessageHandler: binaryMessageHandler,
	}
}

func (m *WSHandler) AddReceiveMessageHandlers(handlers map[string]*WSContext) map[string]*WSContext {
	if m.handlers == nil {
		m.handlers = handlers
		return m.handlers
	}
	for protocolId, wsContext := range handlers {
		m.handlers[protocolId] = wsContext
	}
	return m.handlers
}

func (m *WSHandler) AddReceiveMessageHandler(protocolId string, wsContext *WSContext) map[string]*WSContext {
	if m.handlers == nil {
		m.handlers = map[string]*WSContext{
			protocolId: wsContext,
		}
		return m.handlers
	}
	m.handlers[protocolId] = wsContext
	return m.handlers
}
